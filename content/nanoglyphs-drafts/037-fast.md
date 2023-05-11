+++
image_alt = "La Cheminée (a tall multi-colored art project) dans La Défense"
image_url = "/photographs/nanoglyphs/037-fast/la-cheminee@2x.jpg"
published_at = 2023-05-10T08:59:02+01:00
title = "Fast as a Service"
+++

The middle days of Heroku were distinguished by two ideological groups, cooperating to build the platform, but also at odds with each other. I've named them retroactively "the Stables" and "the Philosophers".

As the name would suggest, the Stables were the more conservative of the two, having existing services to run, existing users to think about, and large, haphazard codebases acting to constrain possibilities around making fast progress. Pragmatic day-to-day engineers that cared about the product, but who were also incentivized to make changes that'd limit the number of middle-of-the-night pages sent their way.

The Philosophers, inspired by the folk history of early innovation groups like Bell Labs and Xerox PARC, specialized in forward thinking in isolation from the lofty heights of an ivory tower, and with limited ability to make real changes to existing products, relied on their ideas being so impactful that the trickle down effect would inspire Stables to take them up and turn them into reality.

Heroku's long term failure to meet the ambitious goals it'd set for itself was partially result of the failure of both these groups. As technical debt and operational burden mounted over the years, the Stables became quagmired, and unable to make the macro-level product changes that were needed for the company to make a real market impact to follow up its initial innovations like `git push heroku master` and builpacks. The Philosophers had some good ideas, but no ability to effect real change. The trickle down effect didn't work, and none of their prototypes made it to realized product form.

---

## Speed as UX (#speed-ux)

One of those early Philosopher prototypes was a toy program to demonstrate the concept of **"fast as a service"**. Heroku's API at the time was not fast. Written and Ruby and leveraging high-magic modules like ActiveRecord, it had all the usual issues like N+1s resulting from hard-to-spot lazy loading problems, and committed other performance sins like in-band remote service calls, largely the result of an accelerated development schedule during a time when service engineering as a field was still emergent and best practices weren't well established. API latency manifested in the real world by making the Heroku CLI less responsive. Commands like `heroku apps` took long enough that there was a noticeable delay before results were delivered.

The fast as a service prototype demonstrated a single API endpoint rewritten in Go and was built around the premise that improved API responsiveness was a feature that'd flow into increased user satisfaction. A popular conference slide making the rounds at the time made the case that responsiveness of < 100ms felt instant, 100ms – 1s felt fast, and 1s+ felt slow, territory where users would lose interest, hitting the back button or starting to multi-task. Making every API operation fast, it'd make the Heroku CLI feel more like a local CLI tool, and unlock new heights of productivity. Long chains of commands (e.g. `heroku apps`, `heroku info -a api`, `heroku ps -a api`, `heroku ps:scale web=2 -a api`) would be as fluid as traversing your own local file system.

The Go API prototype showed an impressive ~10x speed improvement over the Ruby implementation, but it was never a realistic path forward. It implemented only one API endpoint, and only the easy parts, ignoring all the hard, inconvenient stuff that'd be required to make a practical alternative that was production ready and backwards compatible. A Go rewrite might've been possible, but it'd take on the order of 100x more effort to wrangle into reality, and far more tenacity than a Philosopher would be able or willing to apply.

---

## Greenfield (#greenfield)

Still, fast as a service a good idea. Responsive web services are objectively a good thing.

While changing existing products with a large base of users is hard, starting greenfield, or from a project of smaller size, is easier. Fast as a service is a concept that I've kept in mind while building Bridge, and we've managed to mostly achieve it by focusing on it as a goal from the beginning. We're not doing anything extraordinarily novel, but we've made sure that we're building on solid foundations:

* The stack is written in Go, a fast language. It's still possible to make a Go program accidentally slow, but a well-written Go program is faster than most peers automatically (and _much_ faster than many).

* Go is incredibly verbose. This isn't usually a good thing, but does have some advantages. Producing something like ActiveRecord's lazy loading magic would be practically impossible, making it much harder to accidentally write an N+1 query. You can still write one by querying inside of loops, but those instances are easier to spot.

* Since data loading in Go is very manual anyway, we might as well load in batches where possible. If an API resource that's going to be rendered in a loop needs a subresource to render, we load all those subresources first, put them in a map, and making that available to the loop, minimize expensive database operations.

* Every call to a remote service that can be done from a background worker _is_ done from a background worker. So for example, something like sending an email via Mailgun would never be done synchronous. We'd insert a database record representing the intent to send an email, and a worker would make a pass to do the slower heavy lifting.

* Go's `net/http` package does a good job of pooling connections to remote services. For services that are used often, we need to build a new connection relatively infrequently, thereby saving more time on network calls.

* Go makes real parallelization quite easy via goroutines, and although it's usually not necessary to dip into it explicitly (most parallelization is along the axis of parallel HTTP requests), we'll use constructs like [Go's `errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup) where it makes sense to do so.

---

## Canonical lines in Postgres (#canonical-postgres)

Well, at least I _think_ our API is fast, but it'd be nice to quantify that.

Something I've been struggling with over the past couple years is finding a good way to perform ad-hoc queries to get insight into how our stack is running. At my last couple jobs I've had access to Splunk, which is one of the better tools on the market, but with a pricing model is so absolutely outrageous that it's hard to justify introducing it (you're given the privilege of paying upwards of millions of dollars a year _and_ get to run it all yourself). We have an alternative log provider that our logs drain into, but it can't perform aggregates of any kind, making it a glorified log viewer that's borderline useless for operational work.

So a few days ago I did something that I never thought I'd do again, and started putting some of our more critical logs in Postgres.

Never, _ever_ do this, _unless_:

* You're draining only [canonical log lines](https://stripe.com/blog/canonical-log-lines) which are a summarized digest of everything that happened during a single API call, rather than a flood of voluminous and low quality logging data.

* You're putting them in [an ephemeral database](/fragments/ephemeral-db), so they can be shedded as necessary and won't interfere with higher fidelity data like in the case where recovery from backup is necessary.

* You're using [a partitioned table](/fragments/postgres-partitioning-2022) which makes dropping old data easy and fast.

And even then, it's a technique that's probably going to have trouble scaling to enormous sizes.

We have a middleware that generates per-request canonical digests and hands them off to a waiting goroutine which upserts them in batches for every 100 that accumulate (or every 15 seconds, whichever occurs first). A well-defined schema keeps us honest, and provides tab completion when querying from Postgres:

``` sql
CREATE TABLE canonical_api_line (
    id                               uuid NOT NULL DEFAULT gen_ulid(),
    account_email                    varchar(200),
    account_id                       uuid,
    api_error_code                   varchar(200),
    api_error_internal_code          varchar(200),
    api_error_cause                  varchar(2000),
    api_error_message                varchar(2000),
    auth_internal_name               varchar(200),
    auth_type                        varchar(200),
    content_type                     varchar(200),
    created_at                       timestamptz NOT NULL DEFAULT current_timestamp,
    duration                         interval NOT NULL,
    http_method                      varchar(20) NOT NULL,
    http_path                        varchar(200) NOT NULL,
    http_path_original               varchar(200),
    http_route                       varchar(200), -- may be NULL in case of no route match
    idempotency_key                  uuid,
    idempotency_replay               boolean NOT NULL DEFAULT false,
    ip                               inet NOT NULL,
    query_string                     varchar(2000),
    request_id                       uuid NOT NULL,
    request_payload                  jsonb,
    request_payload_capture_disabled boolean NOT NULL DEFAULT false,
    sentry_event_id                  varchar(200),
    sso_provider                     varchar(200),
    statistics                       jsonb,
    status                           int NOT NULL,
    sudo                             boolean NOT NULL,
    sudo_account_email               varchar(200),
    sudo_account_id                  uuid,
    updated_at                       timestamptz NOT NULL DEFAULT current_timestamp,
    user_agent                       varchar(200),
    x_crunchy_headers                varchar(200)[],
    PRIMARY KEY (created_at, id)
) PARTITION BY RANGE (created_at);
```

## Are we fast as a service? (#are-we)

So with Postgres-based canonical lines in place, we can start to produce some numbers. Pure retrievals tend to be faster than mutations, so let's look at `GET` actions first, taken at 50th, 95th, and 99th percentiles:

``` sql
SELECT
    http_method,
    http_route,
    count(*),
    (percentile_cont(0.50) WITHIN GROUP (ORDER BY duration)) AS p50,
    (percentile_cont(0.95) WITHIN GROUP (ORDER BY duration)) AS p95,
    (percentile_cont(0.99) WITHIN GROUP (ORDER BY duration)) AS p99
FROM canonical_api_line
WHERE http_method = 'GET'
GROUP BY http_method, http_route
ORDER BY 3 DESC
LIMIT 10;
```

```
 http_method |                http_route                | count  |       p50       |       p95       |       p99
-------------+------------------------------------------+--------+-----------------+-----------------+-----------------
 GET         | /metric-views/{name}                     | 135521 | 00:00:00.012136 | 00:00:00.021853 | 00:00:00.039698
 GET         | /clusters                                |  23793 | 00:00:00.011221 | 00:00:00.024589 | 00:00:00.046851
 GET         | /teams                                   |  20250 | 00:00:00.009701 | 00:00:00.023705 | 00:00:00.041172
 GET         | /account                                 |  17017 | 00:00:00.007547 | 00:00:00.021594 | 00:00:00.03704
 GET         | /health-checks/{name}                    |  12553 | 00:00:00.002203 | 00:00:00.004137 | 00:00:00.007808
 GET         | /clusters/{cluster_id}                   |   2336 | 00:00:00.037166 | 00:00:00.07439  | 00:00:00.128384
 GET         | /clusters/{cluster_id}/status            |   1211 | 00:00:00.046427 | 00:00:00.923967 | 00:00:02.620303
 GET         | /clusters/{cluster_id}/upgrade           |   1192 | 00:00:00.023269 | 00:00:00.049101 | 00:00:00.077546
 GET         | /clusters/{cluster_id}/databases         |   1116 | 00:00:00.350668 | 00:00:07.739117 | 00:00:26.266837
 GET         | /clusters/{cluster_id}/roles/{role_name} |    500 | 00:00:00.018712 | 00:00:00.037584 | 00:00:00.046902

```

Everything is under 100 ms (and usually well under), with the only outlier being the `/databases` endpoint which is slower because it's SSHing out to a Postgres server to see what Postgres databases it has. `/status` has some bad outliers because it's also remoting down a layer to our backend service.

Now bucket up all the mutating verbs of `DELETE`, `PATCH`, `POST`, and `PUT`:

``` sql
SELECT
    http_method,
    http_route,
    count(*),
    (percentile_cont(0.50) WITHIN GROUP (ORDER BY duration)) AS p50,
    (percentile_cont(0.95) WITHIN GROUP (ORDER BY duration)) AS p95,
    (percentile_cont(0.99) WITHIN GROUP (ORDER BY duration)) AS p99
FROM canonical_api_line
WHERE http_method IN ('DELETE', 'PATCH', 'POST', 'PUT')
GROUP BY http_method, http_route
ORDER BY 3 DESC
LIMIT 10;
```

```
 http_method |              http_route              | count  |       p50       |       p95       |       p99
-------------+--------------------------------------+--------+-----------------+-----------------+-----------------
 POST        | /metrics                             | 943142 | 00:00:00.009432 | 00:00:00.014876 | 00:00:00.024603
 POST        | /vendor/owl/metrics                  | 572721 | 00:00:00.006064 | 00:00:00.010637 | 00:00:00.016132
 POST        | /vendor/owl/pellet                   |  14332 | 00:00:00.021899 | 00:00:00.04165  | 00:00:00.058119
 POST        | /access-tokens                       |  13051 | 00:00:00        | 00:00:00.001168 | 00:00:00.008782
 POST        | /queries                             |    825 | 00:00:00.282502 | 00:00:04.67332  | 00:00:25.809127
 POST        | /sessions                            |     94 | 00:00:00.033717 | 00:00:00.07389  | 00:00:00.213992
 POST        | /clusters/{cluster_id}/backup-tokens |     56 | 00:00:00.055898 | 00:00:00.300625 | 00:00:00.672795
 DELETE      | /clusters/{cluster_id}               |     29 | 00:00:00.052374 | 00:00:00.075728 | 00:00:00.080546
 POST        | /clusters                            |     20 | 00:00:00.14218  | 00:00:00.251888 | 00:00:00.267392
 POST        | /clusters/{cluster_id}/upgrade       |     18 | 00:00:00.180902 | 00:00:00.430255 | 00:00:00.528606
```

Like `/databases` above, `POST /queries` has to SSH to individual Postgres clusters so it's going to take longer and have some bad tail latency. `POST /clusters` and `POST /clusters/{cluster_id}/upgrade` which create and upgrade clusters respectively both reach down a layer and do a fair bit of heavy lifting. Still, those would be good candidates for looking into seeing if there's anything they do which would be moved out-of-band to a background worker.

Overall though, mostly good, and we're keeping responsiveness on most endpoints close to or under that 100ms target. This is of course duration from the perspective of inside the stack, and the numbers for clients making remote calls to us won't be as good. There's always areas for improvement, but especially relatively speaking, I'd say that yes, we are fast as a service.

---

## La Défense (#la-defense)

<img src="/photographs/nanoglyphs/037-fast/la-defense@2x.jpg" alt="La Défense" class="wide" loading="lazy">

The photo at the top is [La Cheminée by Raymond Moretti](https://parisladefense.com/en/discover/artwork/le-moretti), located in Paris' La Défense district. Cities around the world have all trended towards the same car-centric pavement hellscapes and it's rare that I'm enamored by one, but La Défense in Paris is genuinely unique. You emerge from the closest metro station onto a pedestrian promenade kilometers in length, surrounded on all sides by art projects like this one and modern glass buildings, and bookkended by [La Grande Arche](/sequences/052), a massive cube that's as distinctive to Paris' skyline as the Eiffel Tower.

Until next week.

<img src="/photographs/nanoglyphs/037-fast/la-grande-arche@2x.jpg" alt="La Défense" class="wide" loading="lazy">