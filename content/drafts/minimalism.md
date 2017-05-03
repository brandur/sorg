---
title: In Pursuit of Production Minimalism
published_at: 2017-04-11T06:37:34Z
location: San Francisco
hook: A few guidelines for practicing minimalism in
  production to produce tech stacks that are stable and
  operable.
hook_image: true
---

While working at Lockheed during the cold war, Kelly
Johnson was reported to have coined [KISS][kiss] ("keep it
simple, stupid"); a principle that suggests glibly that
systems should be designed to be as simple as possible.

While complexity is never a conscious design goal of any
project, it arises inherently as new features are pursued
or new components are introduced. KISS encourages designers
to actively counteract this force by making simplicity an
objective in itself, and thus produce products that are
more maintainable, more reliable, and more flexible. In the
case of jet fighters, that might mean a plane that can be
repaired in the field with few tools and under the
stressful conditions of combat.

During his tenure, Lockheed's Skunk Works would produce
planes like the U-2 and SR-71; so notable for their
engineering excellence that they've left a legacy that we
reflect on even today.

!fig src="/assets/minimalism/sr71.jpg" caption="The famous SR-71, one of the flag ships of Lockheed's Skunk Works. Very fast even if not simple."

## Minimalism in technology (#in-technology)

Many of us pursue work in the engineering field because
we're intellectually curious. Technology is cool, and new
technology is even better. We want to be using what
everyone's talking about.

Our news sources, meetups, conferences, and even
conversations bias towards shiny new tech that's either
under active development or being energetically promoted.
Older components that sit quietly and do their job well
disappear into the background.

Over time, technologies are added, but are rarely removed.
Left unchecked, production stacks that have been around
long enough become sprawling patchworks combining just
about every technology under the sun. The effect can be
dangerous:

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

Even knowing this, the instinct to expand our tools is hard
to suppress. Oftentimes persuasion is a core competency of
our jobs, and we can use that same power to convince
ourselves and our peers that it's _critical_ to get new
technologies into our stack. That Go-based HA key/value
store will take our uptime and fault resilience to new
highs. That real-time event stream will allow us to create
an immutable ledger that will be foundational for the
entire architecture. In many cases, if we'd taken a step
back and considered things a little more dispassionately,
we might have realized that we could do without them.

## Minimalism in practice (#in-practice)

During my time at Heroku, we endeavored to follow these
ideas where we could. Here are a few examples of the types
of things we did in pursuit of a stack that was more easily
operable and comprehensible:

* The core database that tracked all apps, users, releases,
  configuration, etc. used to be its own special snowflake
  hosted on a custom-built AWS instance. It was eventually
  folded into Heroku Postgres, and became just one more
  node to be managed along with every other customer DB.

* Entire products were retired where possible. For example,
  the `ssl:ip` add-on (providing SSL/TLS terminate for an
  app), which used to be provisioned and run on its own
  dedicated servers, was end-of-lifed completely when a
  better (and cheaper) option for terminating SSL was
  available through Amazon. With SNI support now
  widespread, `ssl:endpoint` will soon follow suit.

* All non-ephemeral data was moved out of Redis so that the
  only data store handling persistent data for internal
  apps was Postgres. This had the added advantage of stacks
  being able to tolerate a downed Redis and stay online.

* After a misguided foray into production polyglotism, the
  last component written in Scala was retired. Fewer
  programming languages in use meant that the entire system
  became easier to operate, and by more engineers.

* The component that handled Heroku orgs was originally run
  as its own microservice. It eventually became obvious
  that there had been a time when our microservice
  expansion had been a little overzealous, so to simplify
  operation, we folded a few services back into the hub.

To recognize the effort that went into tearing down or
replacing old technology, we created a ritual where we
symbolically fed dead components to a flame called a [burn
party](/fragments/burn-parties). The time and energy spent
on some of these projects would in some cases be as great,
or even greater, as it would for shipping a new product.

!fig src="/assets/minimalism/fire.jpg" caption="At Heroku, we'd hold regular \"burn parties\" to recognize the effort that went into deprecating old products and technology."

## Minimalism in simple ideas (#in-simple-ideas)

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

* ***Favor simplicity and reduce moving parts.*** Try to
  keep the total number of things in a system small so that
  it stays easy to understand and easy to operate. In some
  cases this will be a compromise because a technology
  that's slight less suited to a job may have to be re-used
  even if there's a new one that would technically be a
  better fit.

* ***Don't use new technology the day, or even the year,
  that it's initially released.*** Save yourself time and
  energy by letting others vet it, find bugs, and do the
  work to stabilize it.

* ***Discuss new additions broadly.*** Be cognizant that some
  FUD against new ideas will be unreasonable, but try to
  have a cohesive long term technology strategy across the
  entire engineering organization.

It's not that new technology should _never_ be introduced,
but it should be done with rational defensiveness, and with
a critical eye in how it'll fit into an evolving (and
hopefully ever-improving) architecture.

## Nothing left to add, nothing left to take away (#nothing-left-to-add-or-take-away)

Antoine de Saint Exupéry, a French poet and pioneering
aviator, had this to say on the subject:

> It seems that perfection is reached not when there is
> nothing left to add, but when there is nothing left to
> take away.

!fig src="/assets/minimalism/sea.jpg" caption="Nothing left to take away."

For a long time we had a very big and very aspirational
goal at Heroku: by breaking up the platform's "kernel" (the
infrastructure powering user applications) and moving it
piece by piece into the user space that it supported, we
could have Heroku run _on Heroku_. In the ultimate
manifestation of production minimalism, the kernel would
continue to diminish in size until eventually vanishing
completely. The specialized components that it contained
would be retired, and we'd be left a single perfectly
uniform stack.

This concept obviously won't apply to everyone, but most of
us can benefit from architecture that's a little simpler, a
little more conservative, and a little more directed. Only
by concertedly building a minimal stack that's stable and
nearly perfectly operable can we maximize our ability to
push forward with new products and ideas.

[kiss]: https://en.wikipedia.org/wiki/KISS_principle
