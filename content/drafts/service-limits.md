---
hook: If you don't put a limit on it, expect abuse.
location: San Francisco
published_at: 2016-11-06T17:26:49Z
title: Service Limits
---

While I was working at Heroku, we stored a JSON blob in the database for each
release of an app whose values were used to inject configuration variables
into the app's environment. There was initially no size limit on the blob --
not so much by design, but more because it was something that no one had
really thought about all that much.

Months later, we noticed with some surprise that some customers seemed to be
using their environments like a database and taking advantage of the field to
store multi-megabyte behemoths. This had the effect of bloating our database
and slowed start times as the relevant bytes had to be moved over the wire and
into place for the app to run.

It doesn't take a lot of digging to find similar stories -- at Stripe we had a
situation where we weren't limiting the maximum length of text fields that
could be passed into the API. It wasn't long before some users were assembling
huge JSON payloads, encoding them to a string, and then handing them off to us
for cold storage.

An unfortunate truth about the nature of the Internet is that if you don't put
a limit on a resource, you can fully expect it to be eventually abused -- even
if that resource is only modestly valuable. The definition of "resource" here
is purposely vague; it can be anything: an app at Heroku, a free TLS
certificate issued from Let's Encrypt, a request to Spotify's song search API
endpoint, or anything in between.

Today our core services are rate limiting [1] on 20+ different dimensions. For
example:

* The number of inbound requests from a single user.
* The origin IP address of a request.
* The number of requests that a single user is allowed to have in-flight at any
  time. This prevents the saturation of a significant portion of the server by
  issuing a more moderate number of very expensive requests.
* A fleet-wide limit on non-critical requests so that a set of users running a
  barrage of analytical queries can't interfere with a user trying to create a
  charge.
* (And many more.)

Most of these limits were added reactively rather than being originally
designed. This has occasionally been because of an oversight on our part, but
more often it's because we identified a hole in our perimeter after a user was
able to push the system past it limits with a previously unseen technique.
Running large production systems tends to lead to this sort of game of cat and
mouse.

## Moderation and Transparency (#moderation-and-transparency)

Although limits are required to prevent abuse, it's important to remember that
they'll have an effect on legitimate users as they run into them accidentally.
This can be a very frustrating experience for them, and it's worth following a
few best practices for their sake:

* Keep limits _moderate_. Don't be excessively restrictive unless there's a
  very good reason to be.
* Be _transparent_ about limits by publishing them publicly. This will help
  users who encounter them compensate accordingly.

I ran into this recently while I was building a [small app to automatically
create playlists on Spotify][death-guild]. Once a week, it would try to retrieve
Spotify IDs for a list of 30 to 40 songs and turn them into a playlist. I'd
originally implemented a process that would do up to 5 fetches in parallel and
sleep 0 to 1 seconds between each one. This was too aggressive for Spotify's
fairly meager rate limits, and I eventually had to back it off to a only
sequential fetches (no parallelism) and with sleeps of 1 to 2 seconds between
each. Although far from the end of the world, it's probably not a good thing
that such a small integration with a fairly modest task was rate limited so
quickly [2].

Amazon is the gold standard here. As a light-to-moderate user of their
services you're unlikely to brush up limits of any kind, but even once you do,
they have a [service limits page][aws-service-limits] that goes into
meticulous details on limits for everything from the maximum number of
DynamoDB tables that are allowed to how many VPCs a single account can have in
a region.

Spotify will serve as our bad example. Along with stringent limits in the
first place, their [page describing rate limiting][spotify-limits] is devoid
of any useful information with no concrete numbers in sight.

## Rate Limiting Algorithms (#algorithms)

Rate limiting is generally performed by a ["token bucket"][token-bucket]
algorithm. I've written about it and a specific (and quite good)
implementation called GCRA (generic cell rate algorithm) [in more detail
previously](/rate-limits).

I've recently started distributing a GCRA implementation as a Redis module as
part of the [redis-throttle][redis-throttle] project. It aims to provide a
robust and correct implementation in a language agnostic way by abstracting
rate limiting away into a single Redis command:

```
TH.THROTTLE <key> <max_burst> <count per period> <period> [<quantity>]
```

[1] I use _rate limiting_ to refer to limits measured against a time component
as opposed to more static limits like a ceiling on the number of allowed users
or the maximum size of a database row.

[2] Ideally, Spotify's overall limits should be raised, but 

[aws-service-limits]: http://docs.aws.amazon.com/general/latest/gr/aws_service_limits.html
[death-guild]: https://github.com/brandur/deathguild
[redis-throttle]: https://github.com/brandur/redis-throttle
[spotify-limits]: https://developer.spotify.com/web-api/user-guide/
[token-bucket]: https://en.wikipedia.org/wiki/Token_bucket
