---
hook: Why and how to put limits and rate limits on your web services.
  Introducing redis-cell.
location: San Francisco
published_at: 2016-11-08T17:35:34Z
title: Service Limits
---

While I was working at Heroku, we stored a JSON blob in the database for each
release of an app, the values of which were used to inject configuration
variables into the app's environment. There was initially no size limit on the
blob -- not so much by design, but more because it was something that no one
had really thought about.

Months later, we noticed with some surprise that some customers seemed to be
using their environments like a database and taking advantage of the field to
store multi-megabyte behemoths. This had the effect of bloating our database
and slowed start times as the relevant bytes had to be moved over the wire and
into place for the app to run.

It doesn't take a lot of digging to find similar stories. For example, for a
long time at Stripe we weren't limiting the maximum length of text fields that
could be passed into the API. It wasn't long before some users were assembling
huge JSON payloads, encoding them to a string, and then handing them off to us
for cold storage.

The nature of the Internet is such that if you don't put a limit on a resource,
you can fully expect it to be eventually abused. A "resource" can be anything:
an app at Heroku, a free TLS certificate issued from Let's Encrypt, a request
to Spotify's song search API endpoint, or anything in between.

Today our core services at Stripe are rate limiting [1] on 20+ different
dimensions. For example:

* The number of inbound requests from a single user.
* The origin IP address of a request.
* The number of requests that a single user is allowed to have in-flight at any
  time. This prevents the saturation of a significant portion of the server by
  issuing a more moderate number of very expensive requests.
* A fleet-wide limit on non-critical requests so that a set of users running a
  barrage of analytical queries can't interfere with a user trying to create a
  charge.
* (And many more.)

Most of these were added reactively over time as holes in our perimeter were
revealed. Operating large production systems tends to be this sort of game of
cat and mouse -- a constant cycle of reinforcing the fort as new cracks appear.

## Moderation and transparency (#moderation-and-transparency)

Although limits are required to prevent abuse, it's worth remembering that
they'll have an effect on legitimate users too. It can be quite a frustrating
experience for them as they hit rate limits accidentally, and I'd recommend a
few guidelines for their sake:

* Keep limits _moderate_. Don't be excessively restrictive unless there's a
  very good reason to be.
* Be _transparent_ about limits by publishing them publicly. This will help
  users who encounter them compensate accordingly.

I ran into this recently while I was building a [small app to automatically
create playlists on Spotify][death-guild]. Once a week, it would try to retrieve
Spotify IDs for a list of 30 to 40 songs and turn them into a playlist. I'd
originally implemented a process that would do up to 5 fetches in parallel and
sleep 0 to 1 seconds between each one. This was too aggressive for Spotify's
fairly meager rate limits though, and I eventually had to back it off to a only
sequential fetches (no parallelism) and with sleeps of 1 to 2 seconds between
each. Although far from the end of the world, it's probably not a good thing
that such a small integration with a fairly modest task was rate limited so
quickly [2].

On the other side things, Amazon is the gold standard. As a light-to-moderate
user of their services you're unlikely to brush up limits of any kind, but even
once you do, they have a [service limits page][aws-service-limits] that goes
into meticulous details on limits for everything from the maximum number of
DynamoDB tables that are allowed to how many VPCs a single account can have in
a region.

Spotify will serve as our bad example. Along with stringent limits in the first
place, their [page describing rate limiting][spotify-limits] is very long, but
mostly devoid of useful information, with no concrete numbers in sight.

## Rate limiting algorithms (#algorithms)

Rate limiting is generally performed by a ["token bucket"][token-bucket]
algorithm. I've written about it and a specific (and quite good)
implementation called GCRA (generic cell rate algorithm) [in more detail
previously](/rate-limiting).

I've recently started distributing a GCRA implementation as a Redis module as
part of the [redis-cell][redis-cell] project. It aims to provide a robust and
correct implementation in a language agnostic way by abstracting rate limiting
away into a single Redis command:

```
CL.THROTTLE <key> <max_burst> <count per period> <period> [<quantity>]
```

It'll respond with an array of integers indicating whether the action should be
limited and provides some other metadata that can be added to response headers
(i.e. `X-RateLimit-Remaining` and the like):

```
127.0.0.1:6379> CL.THROTTLE user123 15 30 60
1) (integer) 0   # 0 is allowed, 1 is limited
2) (integer) 16  # X-RateLimit-Limit
3) (integer) 15  # X-RateLimit-Remaining
4) (integer) -1  # Retry-After (seconds)
5) (integer) 2   # X-RateLimit-Reset (seconds)
```

redis-cell gets some nice speed advantages by being written as a Redis module.
I [ran some informal benchmarks for it][benchmarks] and found it to be a little
under twice as slow as invoking Redis `SET` command (very roughly 0.1 ms per
execution, or ~10,000 operations per second, as seen from a Redis client). It's
also written in Rust and uses the language's FFI module to interact with Redis;
making contribution easier, and the implementation more likely to be free of
bugs and the memory issues that tend to plague C programs. Alternatively,
there's also [throttled], another implementation of GCRA in pure Go.

[1] I use _rate limiting_ to refer to limits measured against a time component
as opposed to more static limits like a ceiling on the number of allowed users
or the maximum size of a database row.

[2] Ideally, Spotify's overall limits should be raised, but 

[benchmarks]: https://gist.github.com/brandur/90698498bd543598d00df46e32be3268
[aws-service-limits]: http://docs.aws.amazon.com/general/latest/gr/aws_service_limits.html
[death-guild]: https://github.com/brandur/deathguild
[redis-cell]: https://github.com/brandur/redis-cell
[spotify-limits]: https://developer.spotify.com/web-api/user-guide/
[throttled]: https://github.com/throttled/throttled
[token-bucket]: https://en.wikipedia.org/wiki/Token_bucket
