---
title: API Gateways
published_at: 2015-07-27T00:13:33Z
---

With the recent release Amazon's [API Gateway][api-gateway], we're once again
seeing the discussion around this type of service pick up in industry and on
various topical mailing lists. This type of service is far from a new idea,
with some API-centric companies having offered something like it for years
already and with several open-source alternatives also in play for some time:

* [API Umbrella][api-umbrella]
* [Kong][kong]
* [Tyk][tyk]

## Features

The feature sets between these offerings vary, but some of the most common ones
are:

* Analytics/metrics
* Authentication
* Caching
* Logging
* Rate limiting
* Transformations

I didn't include "routing and acting as a reverse proxy", but this is obviously
also a key concern for all these products.

In my mind, the trouble with these feature sets is that none of them are
obviously part of the category of "so invaluable that I'll gladly add another
thing to my stack to get it". Rather the opposite is true in that they all fall
into the buckets of either "trivial", "difficult to generalize", or "somewhat
dubious".

Analytics, caching, logging, and rate limiting fall into the first with each
being relatively easily to get up and running on its own. Rate limiting may be
the least obvious of the three, but given a Redis instance (a dependency that
many apps already have already these days) and the addition of a library like
[throttled][throttled] to your stack, you can still have pretty sophisticated
rate limiting up and running in minutes. Analytics and logging are also
problematic in that whatever the gateway happens to be doing there, there's a
fair chance that it's not going to fit into your internal conventions and set
of tools, necessitating the addition of some kind of custom module anyway.

Along the same lines, authentication is one of those features that's easy to
put on a list, but looks harder the closer you get to it. Anyone can read an
`Authorization` header, but how easy is this going to be to plug into your
existing user management system? After all the configuration involved in such a
project, you might have to ask yourself what you really gained.

Transformations is a pretty questionable feature. If the only way to produce a
consistent set of API responses is to patch in another layer of indirection at
the top of your stack to tweak the bytes in a response, that suggests a
concerning inability to change elsewhere in the organization.

Of course it's never a good idea to develop too much of a NIH attitude, so not
re-inventing the wheel is an admirable goal where possible. The value of any of
these features may not even be in question if adding a gateway was free, but
it's not. Any new component is something else that needs to be explored,
configured, and operated (although the last isn't applicable to Amazon's
gateway), and these costs should be accounted for in any engineering decision.
Some like Kong even require a Cassandra installation be procured for it to work
with, which is a pretty tall order for any organization not already using one.
They also confine you to the bounds of the particular product; if any future
addition requires more freedom than it allows, it might take some creative
plugging to get it to play ball.

The most concisely that I can put it is that it feels like these products are
solving the easy parts of operating a service, while still leaving you shackled
to the hard ones.

## Too Early or Too Late

[api-gateway]: https://aws.amazon.com/api-gateway/
[api-umbrella]: http://apiumbrella.io/
[kong]: http://getkong.org/
[throttled]: 
[tyk]: https://tyk.io/
