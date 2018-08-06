---
title: In Pursuit of Production Minimalism
published_at: 2017-05-10T13:35:02Z
location: San Francisco
hook: Practicing minimalism with the lofty goal of total
  ephemeralization to build coherent, stable, and operable
  stacks.
attributions: Photographs by <strong><a href="https://www.flickr.com/photos/i-am-mclovin/14601998033/">Ben Harrington</a></strong> (SR-71), <strong><a href="https://www.flickr.com/photos/learnscope/5032942270/">Robyn Jay</a></strong> (embers of a burning fire), and <strong><a href="https://www.flickr.com/photos/alamin_bd/22969073683/">Md. Al Amin</a></strong> (boat and sky). Licensed under Creative Commons BY-NC-ND 2.0, BY-SA 2.0, and CC BY 2.0 respectively.
hn_link: https://news.ycombinator.com/item?id=17675243
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

!fig src="/assets/minimalism/sr71.jpg" caption="The famous SR-71, one of the flag ships of Lockheed's Skunk Works. Very fast even if not particularly simple."

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
long enough become sprawling patchworks combining
everything under the sun. This effect is dangerous:

* More parts means more cognitive complexity. If a system
  becomes too difficult to understand then the risk of bugs
  or operational mishaps increases as developers make
  changes without understanding all the intertwined
  concerns.

* Nothing operates flawlessly once it hits production.
  Every component in the stack is a candidate for failure,
  and with sufficient scale, _something_ will be failing all
  the time.

* With more technologies engineers will tend to become
  jacks of all trades, but masters of none. If a
  particularly nefarious problem comes along, it may be
  harder to diagnose and repair because there are few
  specialists around who are able to dig deeply.

Even knowing this, the instinct to expand our tools is hard
to suppress. Oftentimes persuasion is a core competency of
our jobs, and we can use that same power to convince
ourselves and our peers that it's critical to get new
technologies into our stack _right now_. That Go-based HA
key/value store will take our uptime and fault resilience
to new highs. That real-time event stream will enable an
immutable ledger that will become foundational keystone for
the entire platform. That sexy new container orchestration
system that will take ease of deployment and scaling to new
levels. In many cases, a step back and a moment of
dispassionate thought would reveal that their use could be
withheld until a time when they're known to be well vetted,
and it's well understood how they'll fit into the current
architecture (and what they'll replace).

## Through ephemeralization (#ephemeralization)

In his book _Nine Chains to the Moon_ (published 1938),
inventor R. Buckminster Fuller described the idea of
***ephemeralization***:

> Do more and more with less and less until eventually you
> can do everything with nothing.

It suggests improving increasing productive output by
continually improving the efficiency of a system even while
keeping input the same. I project this onto technology to
mean building a stack that scales to more users and more
activity while the people and infrastructure supporting it
stay fixed. This is accomplished by building systems that
are more robust, more automatic, and less prone to problems
because the tendency to grow in complexity that's inherent
to them has been understood, harnessed, and reversed.

For a long time we had a very big and very aspirational
goal of ephemeralization at Heroku. The normal app platform
that we all know was referred to as "user space" while the
internal infrastructure that supported it was called
"kernel space". We want to break up the kernel in the
kernel and move it piece by piece to run inside the user
space that it supported, in effect rebuilding Heroku so
that it itself ran _on Heroku_. In the ultimate
manifestation of ephemeralization, the kernel would
diminish in size until it vanished completely. The
specialized components that it contained would be retired,
and we'd be left a single perfectly uniform stack.

Realistic? Probably not. Useful? Yes. Even falling short of
an incredibly ambitious goal tends to leave you somewhere
good.

## In practice (#in-practice)

Here are a few examples of minimalism and ephemeralization
in practice from Heroku's history:

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
  widespread, `ssl:endpoint` will eventually follow suit.

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

## In ideas (#in-ideas)

Practicing minimalism in production is mostly about
recognizing that the problem exists. After achieving that,
mitigations are straightforward:

* ***Retire old technology.*** Is something new being
  introduced? Look for opportunities to retire older
  technology that's roughly equivalent. If you're about to
  put Kafka in, maybe you can get away with retiring Rabbit
  or NSQ.

* ***Build common service conventions.*** Standardize on
  one database, one language/runtime, one job queue, one
  web server, one reverse proxy, etc. If not one, then
  standardize on _as few as possible_.

* ***Favor simplicity and reduce moving parts.*** Try to
  keep the total number of things in a system small so that
  it stays easy to understand and easy to operate. In some
  cases this will be a compromise because a technology
  that's slightly less suited to a job may have to be 
  re-used even if there's a new one that would technically 
  be a better fit.

* ***Don't use new technology the day, or even the year,
  that it's initially released.*** Save yourself time and
  energy by letting others vet it, find bugs, and do the
  work to stabilize it. Avoid it permanently if it doesn't
  pick up a significant community that will help support it
  well into the future.

* ***Avoid custom technology.*** Software that you write is
  software that you have to maintain. Forever. Don't
  succumb to NIH when there's a well supported public
  solution that fits just as well (or even almost as well).

* ***Use services.*** Software that you install is software
  that you have to operate. From the moment it's activated,
  someone will be taking regular time out of their schedule
  to perform maintenance, troubleshoot problems, and
  install upgrades. Don't succumb to NHH (not hosted here)
  when there's a public service available that will do the
  job better.

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

!fig src="/assets/minimalism/sea.jpg" caption="Nothing left to add. Nothing left to take away."

Most of us can benefit from architecture that's a little
simpler, a little more conservative, and a little more
directed. Only by concertedly building a minimal stack
that's stable and nearly perfectly operable can we maximize
our ability to push forward with new products and ideas.

[kiss]: https://en.wikipedia.org/wiki/KISS_principle
