---
title: Practicing Minimalism In Production
published_at: 2017-03-21T15:20:48Z
location: San Francisco
hook: A few guidelines for practicing minimalism in
  production to produce tech stacks that are stable and
  operable.
---

While working at Lockheed during the cold war, Kelly
Johnson was reported to have coined the [KISS][kiss]
principle ("keep it simple, stupid") which suggests that
systems should be designed to be simple instead of
complicated. While the latter is tempting in the pursuit of
certain features, simplicity wins out in the long run by
producing a product that's more maintainable, more
reliable, and more flexible. In the case of jet fighters,
that might mean a plane that can be repaired in the field
with few tools and under the stressful conditions of
combat.

During his tenure, Lockheed's Skunk Works would produce
planes like the U-2 and SR-71, so notable for their
engineering excellence, that they've left a legacy that we
reflect on even today.

## Minimalism In Technology (#in-technology)

Technology is cool, and as engineers, we tend to want to
use new and interesting things that appear. That sort of
intellectual curiosity is the reason we're in this industry
in the first place! Many of us have had a colleague say
that the company's bound to lose its competitive edge
unless it gets an HA key/value store written in Go or a
real-time event stream into production ASAP. Some of us
have even been that person.

Our news sources, meetups, and even conversations also bias
towards newer technologies that are either under active
development or being promoted. Older technology that sits
quietly and does its job well disappears into the
background.

Over time, technologies have a tendency to be added, but
are rarely removed. Left unchecked, stacks that have been
around long enough tend to be a sprawling patchwork
combining just about every technology under the sun. This
effect can be dangerous:

* More parts means more cognitive complexity. If a system
  becomes too difficult to understand then the risk of bugs
  or operational mishaps increases as developers make
  changes without understanding all the intertwined
  concerns.

* Nothing operates flawlessly once it hits production.
  Every component in the stack is a candidate for failure,
  and with sufficient scale, _something_ will be failing all
  the time.

* With more technologies engineers will tend to be come
  jacks of all trades, but masters of none. If a
  particularly nefarious problem comes along, it may be
  harder to diagnose and repair because there are few
  specialists around who are able to dig deeply.

Even knowing this, the instinct to expand our tools might
be hard to suppress. As engineers we're used to justifying
our technical decisions all the time, and we're so good at
it that we can justify even the bad ones.

## Some Stories From the Inside (#stories)

Here are some favorite examples of production minimalism in
practice from my time operating the platform at Heroku:

* The core database that tracked all apps, users, releases,
  configuration, etc. used to be its own special snowflake
  hosted on a custom-built AWS instance. It was eventually
  folded into Heroku Postgres, and became just one more
  node to be managed along with every customer DB.

* Entire products were retired where possible. For example,
  the `ssl:ip` add-on, which used to be provisioned and run
  on its own dedicated servers, was end-of-lifed completely
  when a better (and cheaper) option for terminating SSL
  was available through Amazon. With SNI support now
  widespread, hopefully `ssl:endpoint` will soon follow
  suit.

* All non-ephemeral data was moved out of Redis so that the
  only data store handling persistent data for internal
  apps was Postgres. This had the added advantage of stacks
  being able to tolerate a downed Redis and stay online.

* After an admittedly misguided foray into polyglotism, the
  last component written in Scala was retired. Fewer
  programming languages in use meant that the entire system
  became easier to operate, and by more engineers.

* The component that handled Heroku orgs was originally run
  as its own microservice. It eventually became obvious
  that there had been a time when our microservice
  expansion had been a little overzealous, so to simplify
  operation, we folded a few services back into the hub.

To recognize the effort that went into tearing down or
replacing old technology, we symbolically fed dead
components to a flame at a [Heroku burn
party](/fragments/burn-parties). The time and energy spent
on some of these projects would in some cases be as great
as it would for shipping a new product.

TODO: photo of Fire.

## Minimalist Guidelines (#guidelines)

Practicing minimalism in production is mostly about
recognizing that the problem exists. After that,
mitigations that can address it effectively are pretty
straightforward:

* ***Retire old technology.*** Is something new being
  introduced? Look for opportunities to retire older
  technology that's roughly equivalent. If you're about to
  put Kafka in, maybe you can get away with retiring Rabbit
  or NSQ.

* ***Build common paths.*** Standardize on one database, one
  language/runtime, one job queue, one web server, one
  reverse proxy, etc. If not one, then standardize on _as
  few as possible_.

* ***Favor simplicity.*** Try to keep the total number of
  moving parts small to keep the system easy to understand
  and easy to operate. In some cases this will be a
  compromise because a technology that's slight less
  suited to a job may have to be re-used instead of a
  new one that's a little better introduced.

* ***Don't use new technology the day, or even the year, that
  it's initially released.*** Save yourself time and energy by
  letting others vet it, find bugs, and help stabilize it.

* ***Discuss new additions broadly.*** Be cognizant that some
  FUD against new ideas will be unreasonable, but try to
  have a cohesive long term technology strategy across the
  entire engineering organization.

## Nothing Left to Add, Nothing Left to Take Away (#nothing-left-to-add-or-take-away)

Antoine de Saint Exupéry, a French poet and pioneering
aviator, had this to say about perfection:

> It seems that perfection is reached not when there is
> nothing left to add, but when there is nothing left to
> take away.

<img src="/assets/minimalism/sea.jpg" data-rjs="2" class="overflowing">

A "big idea" architectural principle at Heroku was to
eventually have Heroku run _on Heroku_. This was obviously
difficult for a variety of reasons, but would be the
ultimate manifestation of production minimalism in that our
"kernel" (the infrastructure supporting the user platform)
would be broken away piece by piece until it was reduced to
the smallest size necessary to keep everything running.

This concept obviously won't apply to everyone, but most of
us can benefit from a stack that's a little simpler, a
little more cautious, and a little more directed. Only by
concertedly building a minimal stack that's stable and
nearly perfectly operable can we maximize our ability to
push forward with new products and ideas.

[kiss]: https://en.wikipedia.org/wiki/KISS_principle
