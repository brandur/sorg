+++
hn_link = "https://news.ycombinator.com/item?id=16833568"
hook = "Let's build more resilient services by paying attention to the edges."
published_at = 2018-04-13T14:37:39Z
title = "The 500 test"
+++

I’ve been thinking about software resilience lately, and
especially where systems that lack resiliency went wrong.
It's a problem that lands squarely in the bucket of "things
that are tractable early on, but nearly impossible to fix
later", which in software, is a big one.

In a typical online service, most of us expect
upwards of 99.9% of the calls it processes to be fulfilled
successfully. However, there will always be some number off
in the tail lost to an accidental bug, a bad deploy, an
external service that fails to respond, or a database
failure or timeout. These are often surfaced to users as an
HTTP `5xx` status code.

It's those edges where engineering time is lost. For
services where every call is critical and must resolve, but
which don't bake in strong rules around consistency, state
must be laboriously repaired by hand -- at incredible
expense.

We should endeavor to not be losing that time. To that end,
I propose *the 500 test* as a guiding operating principle
for building online software:

> In the event of any internal fault (`500`), your service
> should be left in a state that’s consistent, or one that
> will be eventually consistent, as long as the latter will
> occur in a fully automated way.

The consistency I'm talking about here is the same one as
in [ACID][acid]. All state, be it in the local service or
in foreign ones should be valid, even if the `500`
interrupted a request midway between object mutations.

Readonly requests like `GET`s usually pass the 500 test
automatically -- no state is modified, so it's just as
consistent after the failed request as before. Mutations
are more difficult. To ensure safety, services need to be
built with [strong transactional
consistency](/http-transactions). Where in-band invocations
of foreign services are necessary, more sophisticated (and
more complex) techniques like [transactional state
machines](/idempotency-keys) are appropriate.

Why the emphasis on "automated"? The human time spent on
operation and recovery is expensive, and a top factor
leeching engineering productivity, second only to
widespread intrinsic technical debt. Building systems that
are safe at rest (or at least dramatically safer) is the
ultimate way of producing engineering leverage to make new
things.

Think about the edges, and bake in resiliency from day one.
Avoid technologies that make this difficult-to-impossible
(e.g., transaction-less data stores). Failure shouldn’t be
routine, but it should be *handled* routinely.

[acid]: https://en.wikipedia.org/wiki/ACID
