+++
hook = "Examining the nature of a 429 (too many requests), and the purpose it serves."
published_at = 2023-07-06T12:39:10-07:00
title = "Rate limiting, DDOS, and hyperbole"
+++

Last week in the whacky real world sitcom that is now normal life in the 2020s, we had the brief episode of rate-limit-exceeded-gate, in which Elon added a daily limit to viewable tweets for Twitter's users, and one that was so miserly that just about every active user hit it in the first hour, seeing their requests fail with a "rate limit exceeded" (429) error. The ostensible reason for the change was to discourage scraping, which was also the rationale for adding a login wall a few days before.

Naturally, it drove the usual Atlantic-reading Maddow-fearing partisan crowd into a feeding frenzy, and we were treated to multiple days of wild hot takes and blustering. In one Mastodon post that made the rounds, the writer claimed with gleeful rage how Twitter was DDOSing itself, showing a web console wherein Twitter's web client, treating a rate limited request the same as any other request error, retried it many times in a row, thus the DDOS.

Look, this whole escapade was stupid, and I'll be the last one to defend Elon's choices, but it's stupid enough that there's also no need to exaggerate it beyond what it is. The user called it a DDOS, but noticeably absent from their post was the suggestion that it was a _successful_ DDOS. It caused many more requests than necessary, but didn't actually cause the service to buckle, which is what any good, god-fearing DDOS is supposed to do.

## The 429's raison d'être (#429-raison-d-etre)

An important thing to understand about 429s is that they're not so much a punishment as they are a _mitigation_. Under normal use (not what happened here), they keep a system healthy even in the presence of a DDOS by short-circuiting traffic that would otherwise need to be fulfilled by resource-limited systems. A well-designed rate limiter:

* Makes as few database calls as possible.
* Stores state in something highly scalable, usually a key/value store like Redis Cluster, or even in memory [1].
* Runs as close to the edge of the stack as possible, maybe even at [_the_ edge](https://en.wikipedia.org/wiki/Edge_computing).

These techniques make it possible to fulfill millions of 429s per second at minimal cost, which is exactly what you want when your service is under attack. So even fulfilling ten 429s that otherwise might've been one successful call, you're usually (again, depends heavily on implementation) still coming out ahead because each 429 uses 1/10th, 1/100th, or 1/1000th the resources the 200 would've otherwise.

## 429s IRL (#429s-irl)

To demonstrate this (on a real world service, albeit not at Twitter's scale), I put our own rate limiting layer through its paces.

First, I'll arbitrary pick our saved queries endpoint and query our [canonical log lines](/canonical-log-lines) to get a rough idea as to its performance and the number of database calls it makes:

```
=> select http_route, status, duration, statistics ->> 'db_num_queries' as db_num_queries from canonical_api_line where http_method = 'GET' and http_route = '/saved-queries/{saved_query_id_or_secret}' and status = 200 order by id desc limit 10;
           http_route            | status |    duration     | db_num_queries
-------------------------------------------+--------+-----------------+----------------
 /saved-queries/{saved_query_id} |    200 | 00:00:00.028939 | 13
 /saved-queries/{saved_query_id} |    200 | 00:00:00.032773 | 13
 /saved-queries/{saved_query_id} |    200 | 00:00:00.031184 | 13
 /saved-queries/{saved_query_id} |    200 | 00:00:00.039129 | 13
 /saved-queries/{saved_query_id} |    200 | 00:00:00.030355 | 13
 /saved-queries/{saved_query_id} |    200 | 00:00:00.031275 | 13
 /saved-queries/{saved_query_id} |    200 | 00:00:00.026365 | 13
 /saved-queries/{saved_query_id} |    200 | 00:00:00.027473 | 13
 /saved-queries/{saved_query_id} |    200 | 00:00:00.046657 | 13
 /saved-queries/{saved_query_id} |    200 | 00:00:00.031122 | 13
```

So a typical API request on this endpoint takes **~30ms** and makes **13** queries to the database.

Next, I induce a lot of artificial traffic to get myself rate limited. Here are some more requests to the same endpoint, but rate limited with 429s this time:

```
=> select http_route, status, duration, statistics ->> 'db_num_queries' as db_num_queries from canonical_api_line where http_method = 'GET' and http_route = '/saved-queries/{saved_query_id_or_secret}' and status = 429 order by id desc limit 10;
           http_route            | status |    duration     | db_num_queries
-------------------------------------------+--------+-----------------+----------------
 /saved-queries/{saved_query_id} |    429 | 00:00:00:000135 | 0
 /saved-queries/{saved_query_id} |    429 | 00:00:00:000108 | 0
 /saved-queries/{saved_query_id} |    429 | 00:00:00:000156 | 0
 /saved-queries/{saved_query_id} |    429 | 00:00:00:000197 | 0
 /saved-queries/{saved_query_id} |    429 | 00:00:00:000093 | 0
 /saved-queries/{saved_query_id} |    429 | 00:00:00:000137 | 0
 /saved-queries/{saved_query_id} |    429 | 00:00:00:000132 | 0
 /saved-queries/{saved_query_id} |    429 | 00:00:00:000128 | 0
 /saved-queries/{saved_query_id} |    429 | 00:00:00:000109 | 0
 /saved-queries/{saved_query_id} |    429 | 00:00:00:000168 | 0
 ```

A typical rate limited request takes **~100µs** and makes zero queries to the database. That's **300x** faster and infinitely less load on the database.

I'm cheating a little bit in that this is the per-IP rate limiter kicking in which requires zero database operations. There's also a per-account limiter that needs two database queries to authenticate a user before rate limiting. It takes longer, but is still **~10x** faster than a normal request:

```
=> select http_route, status, duration, statistics ->> 'db_num_queries' as db_num_queries from canonical_api_line where (statistics ->> 'db_num_queries')::integer > 0 and status = 429 order by id desc limit 10;
            http_route            | status |    duration     | db_num_queries
----------------------------------+--------+-----------------+----------------
 /metric-views/{name}             |    429 | 00:00:00.005425 | 2
 /metric-views/{name}             |    429 | 00:00:00.003625 | 2
 /account                         |    429 | 00:00:00.00916  | 2
 /teams                           |    429 | 00:00:00.006519 | 2
 /account                         |    429 | 00:00:00.007714 | 2
 /account                         |    429 | 00:00:00.009666 | 2
 /teams/{team_id}                 |    429 | 00:00:00.005238 | 2
 /clusters/{cluster_id}           |    429 | 00:00:00.010141 | 2
 /clusters/{cluster_id}/databases |    429 | 00:00:00.008411 | 2
 /teams                           |    429 | 00:00:00.006754 | 2
```

Lastly, keep in mind that we're not using an ORM, and are careful to a fault about using query patterns that minimize database interactions. 13 database operations for an API request is a small number. A similar endpoint at Stripe would've performed hundreds of Mongo operations, and might even reach thousands. In other words, our stack's 429s are 10x cheaper than successful API request, but I'd expect that for a lot of real world services they're more like 100x cheaper to fulfill.

A long of way of saying that serving 429s is _fast_ and they shouldn't be equated with standard API requests.

[1] Made possible if a service is deployed so that traffic gets roughly equally distributed between nodes. Limiting decisions won't be perfect, but can be close enough for viability.