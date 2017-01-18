---
title: Production Minimalism
published_at: 2016-11-08T04:56:03Z
location: San Francisco
hook: TODO
---

A favorite practice that I learned while at Heroku was the
idea of production minimalism -- curbing a system's
complexity by reducing the total number of moving parts
that it contains. Ideally it should be distilled down to
its simplist possible form, 

The idea stems at least in part from an old US Navy design
principle called [KISS][kiss] ("keep it simple, stupid")
which warns that unnecessary complexity should be avoided.
More aspirationally, Antoine de Saint Exupéry, a French
poet and pioneering aviator, said this about about
minimalist:

> It seems that perfection is reached not when there is
> nothing left to add, but when there is nothing left to
> take away

And similarly again, a quote widely attribute to Albert
Einstein says:

> Make everything as simple as possible, but not simpler.

## Minimalism In Technology (#in-technology)

Technology is cool, and as engineers, we tend to want to
use new and interesting things that appear on the scene.
Over time, components tend to get added, but are rarely
removed, so left unaddressed the default path is to
eventually run a stack that combines just about every
technology under the sun.

This is an instinct that needs to be suppressed to build a
stable and maintainable production stack. More technology
isn't better, it's worse:

* Nothing operates flawlessly once it hits production.
  Every component in the stack is a candidate for failure.

* More parts means more cognitivie complexity. If a system
  becomes too difficult to understand then the risk of bugs
  increases astronomically as developers make changes for
  which they can't grasp all the possible repercussions.

* People get good at operating common paths. More shared
  technology means that when there's a particularly sticky
  problem, there's a better chance that someone will be
  able to fix it.

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
