---
hook: Useful distinction or new buzzword? Comments on 200-500 line services.
location: San Francisco
published_at: 2014-04-01T12:37:17Z
title: Microservices
---

A recent article by James Lewis and Martin Fowler has kicked off a discussion about "microservices" and SOA in general; one of the few topics in the API community besides Hypermedia that's pretty much guaranteed to generate some healthy debate. The [original body of work](http://martinfowler.com/articles/microservices.html) defines the characteristics of a microservice:

* Services are out-of-process components that can be deployed and operated independently.
* Organizes teams around _business capability_ so that each service is run by a cross-functional group.
* Services communicate via lightweight HTTP interfaces. A heavy [ESB](http://en.wikipedia.org/wiki/Enterprise_service_bus) is no longer needed to choreograph data.
* Services are language agnostic and can be built according to the technologies that a team selects.
* Services don't share databases.

At least some portion of the community expresses some incredulity that the talking points are enough to justify a distinct concept from the more classic idea of SOA that was first put forward [somewhere in the neighborhood of 2005](http://books.google.com/books/about/Service_Oriented_Architecture.html?id=qLrLngEACAAJ). Many companies have architectures that would check all the boxes above without ever having heard the word "microservice".

Even among its originators, the term doesn't seem to have a perfect definition with wide consensus. A talk by [Fred George on the subject in Barcelona](https://www.youtube.com/watch?v=2rKEveL55TY) adds the "micro-" back in microservices and puts forward some more radical assertions regarding their nature:

* Services should be 200-500 LOC.
* Self-monitoring services replace unit tests, and business monitoring replaces acceptance tests.
* The system is long-lived while its services are short-lived. Services are disposed as refinements come along to re-work the architecture.

## Autonomy of Small Services (#autonomy)

Although microservices might be SOA with a modern veneer of HTTP, JSON, and polygot, the concept of a "micro microservice" (that 200-500 LOC sweet spot) is worth considering in a bit more depth. In my own experience, not all services can fit into this size, but services that do are remarkably more stable than their counterparts --- and for anyone who's an operator as well as a developer, building a completely autonomous system is an idyllic vision well worthy of pursuit.

These tiny services have some major advantages over their heavier counterparts:

* They have an inherently smaller surface area, and their development can be continually iterated until all bugs are squashed.
* Due to their small area of responsibility, they're rarely in the same state of constant change that's a matter of course many larger codebases. Less change is natural protection against new bugs or regressions.
* In many cases their resource use will be smaller, which could help avoid a class of bugs stemming from overutilization like GC pauses, out-of-memory errors, swapping, etc.
* May be able to use only a very reliable data store like S3 or even be made completely stateless, which can avoid a single point of failure like a relational database.

I ran some inventory of our own production services, and came up with a few that do in fact make the 500 LOC microservice cut:

* **Addons SSO:** 171 LOC. This tiny service authenticates a user via OAuth, then asks the API to sign a request on their behalf before redirecting. Powers `heroku addons:open`.
* **Anvil:** 337 LOC. [A platform-powered build system](https://github.com/ddollar/heroku-anvil) that compiled slugs and released them directly rather than using the more traditional `git push heroku master` route.
* **Cloner:** 305 LOC. A tiny app that authenticates via OAuth and makes an API call. Powers [java.heroku.com](https://java.heroku.com).
* **Zendesk SSO:** 348 LOC. Creates Zendesk accounts for new Heroku users so that they can open support tickets.

A few others didn't quite make weight, but are still small:

* **Deployhooks:** 1240 LOC. A small service that powers the [Heroku Deployhooks add-on](https://devcenter.heroku.com/articles/deploy-hooks).
* **Scheduler:** 630 LOC. Powers the web frontend for the [Heroku Scheduler add-on](https://devcenter.heroku.com/articles/scheduler).
* **Vixie:** 805 LOC. Powers the backend of Heroku's Scheduler add-on, and receives instructions from the scheduler above.

_(Jamie Hodge points out that many of these services lack full databases and therefore might better be described as coordinators. The more limited state which they maintain simplifies their operational characteristics to a large degree.)_

One common trait of all the services listed above is that their autonomy is remarkable. We have some basic alarms on them in case they go down, but they go off rarely. Being deployed on the Heroku platform is certainly a big help here, but also that their concerns are so narrow and unchanging that there isn't a lot of room for bugs to hide.

I suspect that 500 LOC isn't enough to box in all concerns of many components, even if they are quite narrow in scope -- anecdotally, most of our more important services easily blow past this limit. I'm also not quite at the point where I'd replace my unit or acceptance tests with self- or business monitoring (also one of the assertions made above).

## SOA isn't a Silver Bullet (#downsides)

SOA bestows a huge number of architectural advantages, but we probably want to be careful to keep its expectations in check. Boiled down to a fundamental level, SOA is about introducing very hard isolation between components that can result in big gains in organizational and operational effectiveness, but by extension leads to component boundaries that are more difficult to evolve.

The less desirable features of a SOA-like system include the following:

* Any change to the contract between two services will require coordinated development and deployment on both sides. This can be especially slow if those services are managed by different groups of people.
* Data beccomes much more widely distributed and more difficult to inspect and query. This can be solved with something like a data warehouse, but that's another component to build and manage.
* There is some overhead to building the platform that will enable the deployment of new services. Tools like Docker (or Heroku) will help with this, but new services still need metrics dashboards, alarms, deployment and operation tools, etc.
* More integration testing has to be moved out of individual components and up into a working system to be effective, this is inevitably slower and more opaque.

Although not meant to sound too bearish on microservice type architectures, the short version is that it's not just about building the system --- it's also about building the tools, processes, and infrastructure to operate it. Larger companies will almost inevitably have to move to something that looks like SOA to keep forward progress possible, but it might not make sense for smaller shops to rush into it headlong, even if the technology is really cool.

[Wikipedia](http://en.wikipedia.org/wiki/Service-oriented_architecture) nails this idea in a single line:

> Significant vendor hype surrounds SOA, which can create exaggerated expectations.
