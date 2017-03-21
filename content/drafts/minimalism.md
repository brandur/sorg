---
title: Practicing Minimalism In Production
published_at: 2017-03-21T15:20:48Z
location: San Francisco
hook: TODO
---

While developing the U-2 and SR-71 Blackbird spy planes at
the Lockheed Skunk Works, Kelly Johnson was reported to
have coined the [KISS][kiss] principle ("keep it simple,
stupid") which suggests that systems should be designed to
be simple instead of complicated. While the latter is
tempting in the pursuit of certain features, simplicity
wins out in the long run by producing a product that's more
maintainable, more reliable, and more flexible. In the case
of jet fighters, that might mean a plane that can be
repaired in the field with few tools and under the
stressful conditions of combat.

More aspirationally, Antoine de Saint Exupéry, a French
poet and pioneering aviator, said this about product
minimalism:

> It seems that perfection is reached not when there is
> nothing left to add, but when there is nothing left to
> take away.

<img src="/assets/minimalism/sea.jpg" data-rjs="2" class="overflowing">

## Minimalism In Technology (#in-technology)

Technology is cool, and as engineers, we tend to want to
use new and interesting things that appear on the scene. In
fact, it's easy to start favoring new technologies because
they're the ones getting the most press. Use Mongo! No,
Node! Wait, Go! Kafka!

Over time, technologies are added, but are rarely removed,
leading to an accumulation effect. Left unchecked, stacks that
have been around enough tend to combine just about every
technology under the sun.

This is an instinct that needs to be suppressed to build a
stable and maintainable production stack. More technology
can intuitively seem better, but it's often worse:

* Nothing operates flawlessly once it hits production.
  Every component in the stack is a candidate for failure,
  and with sufficient scale, _something_ will be failing all
  the time.

* More parts means more cognitive complexity. If a system
  becomes too difficult to understand then the risk of bugs
  increases as developers make changes without
  understanding all the possible ramifications.

* With more technologies engineers will tend to be come
  jacks of all trades, but masters of none. If a
  particularly nefarious problem comes along, it may be
  harder to diagnose and fix because there are few
  specialists around.

Even knowing this, the instinct to use something new will
be hard to suppress. As engineers we're used to justifying
our design choices all the time, and we're often so good at
it that we can even justify some misguided ones. "But we
_need_ horizontal scalability. We _need_ a streaming
replicated log. We _need_ an highly available key/value
store. We _need_ DNS-based service discovery."

## Practicing Minimalism (#practicing)

Practicing minimalism in a production stack is pretty
approachable by following generally good engineering
guidelines:

* Is a new technology being introduced? Look for
  opportunities to retire an old one that's roughly
  equivalent. If you're about to put Kafka in, maybe you
  can get away with retiring NSQ.

* Build common technology paths. Standardize on one
  database, one language/runtime, one job queue, one web
  server, one reverse proxy, etc. If not one, then
  standardize on _as few as possible_.

* 

Introduce a new technology? Retire an old one.

Build common paths.

Use versatile systems that have a history of reliability.

* When adding new technology, migrate old products over to
  it where possible, and try to retire 

## Examples (#examples)

Here are some favorite examples from my time at Heroku:

* The core Heroku database used to be its own special
  snowflake hosted on an AWS instance. It was eventually
  folded into Heroku Postgres, and became just one more
  node to be managed along with every customer DB.

* Entire products were retired where possible. Dedicated
  IP endpoints (i.e. `ssl:ip`) and shared databases (which
  ran as a component separate to Heroku Postgres) were
  end-of-lifed completely, leading to less ongoing
  maintenance burden.

* All non-ephemeral data was moved out of Redis so that the
  only persistent data store connected to most internal
  apps was Postgres. This was also nice because stacks
  could be programmed to tolerate a downed Redis.

* After an admittedly misguided foray into polyglotism, the
  last component written in Scala was retired. Fewer
  programming languages in use meant that the entire system
  became easier to operate.

* The component that handled Heroku orgs was originally run
  as its own microservice. It eventually became obvious
  that improper disciple in microservicification had been
  applied, so it was folded back into the core API.

Retired products were symbolically fed to the flame at a
[Heroku burn party](/fragments/burn-parties) to make sure
that there was adequate recognition of the effort that went
into tearing down products and technology, and not just for
adding them.

TODO: photo of Fire.

## Do More With Less (#more-with-less)

A "big idea" architectural principle at Heroku was to
eventually have Heroku run _on Heroku_. This was obviously
difficult for a variety of reasons, but would be the
ultimate manifestation of production minimalism.

A favorite practice that I learned while at Heroku was the
idea of production minimalism -- curbing a system's
complexity by reducing the total number of moving parts
that it contains. Ideally it should be distilled down to
its simplist possible form, 

Do more with less by embracing production minimalism. I'd
of course recommend that production stacks should be made
up of good programming language, Postgres for persistent
data, and Redis for everything else; all baked up top of a
deployment stack that someone else maintains, but it's up
to individual implementers to choose their own common
paths. Brand new technology should generally be avoided at
all costs until its had a few years to bake in someone
else's stack, and even then, should only be brought if
there's dire need or if it will allow a system to be
further simplified.

[kiss]: https://en.wikipedia.org/wiki/KISS_principle
