---
title: Microservices and the Monolith
published_at: 2017-01-05T16:41:25Z
hook: Microservices may be out of vogue, but we should be
  wary of overcompensation.
location: San Francisco
---

About three years ago, the idea of a service-oriented
architecture was re-popularized by a successful blog post
that rebranded the idea as _microservices_. A golden age
followed, with many supporting articles, talks, and books
published in support the subject (I'm [guilty of this
myself](/microservices)).

Since then, there's been a reaction in the other direction
as people who've worked with these architectures have
had time to see their blemishes in a stronger light.
Complaints on Twitter are not unusual, and material in
support of the monolith is being published more frequently.
Edit: Love thy monolith -- who wrote that?

## Fighting the sprawl (#sprawl)

Indeed building a product based off of microservices
presents quite a few challenges that were ignored in the
initial puff pieces:

* Major refactors across service boundaries require so much
  coordination that they're nigh impossible. Architecture
  will default to a frozen state without major effort.

* Each service becomes its own [Galapagos][galapagos] of
  local conventions which makes cross-service contribution
  difficult. The worst case is that you end up with a true
  polyglot environment where so many languages are in use
  that few people can contribute to all of them.

* Operational visibility becomes a huge concern in that
  it's difficult to reason about what's happening during
  the execution of a task as it crosses service boundaries.

* Tooling, alerting, visibility, and other supporting
  infrastructure need to be installed and managed for every
  service installation. Updates need to be distributed to
  many different places. It's hard to gain company-wide
  leverage by sharing work.

!fig src="/assets/microservices-and-the-monolith/monolith.jpg" caption="Something to remind you of a monolith."

## The monolith's cracks (#cracks)

I've now had the opportunity to see both sides of the coin:
an internal architecture that was largely service-oriented,
and one that depends heavily on a very large monolith; both
at medium-sized 50+ engineer companies.

It's been tempting lately to glorify the monolith after
being fed aspirational platitudes from the microservices
camp for so long, but we should be wary of
overcompensation. Monolithic architectures have some
serious problems once they hit scale:

* Their test suite gets big and the edit-compile-debug gets
  very long. This has the effect of _crushing_ developer
  productivity. It can't be exaggerated how much agility
  gets lost when developers lose the ability to run a full
  test suite locally and have to rely on hugely long
  iteration loops in CI and clicking through glacially slow
  web interfaces to see results.

* Code quality falls through the floor. Despite mostly good
  intentions, a tragedy of the commons situation develops
  very quickly. Knowing that the onus of maintenance will
  fall on the many rather than be an individual's
  responsibility implicitly encourages sloppiness.

* After a certain size of either code or organization, it's
  impossible for any one person to direct architecture due
  to the number of incoming changes and the overwhelming
  amount of existing code. As we've seen with many of the
  best OSS projects (e.g. the kernel, Postgres, Sequel),
  having a small team of core architects (or even a team of
  one) that sign off on all changes is hugely beneficial in
  ensuring uniformly high code quality. A big monolith
  makes this model untenable.

* Despite being a single codebase, common convention is
  still a problem. You can try to fight deviation with
  linters, but there's always going to be many ways of
  writing anything that aren't related to how much
  whitespace there is around a curly brace.

## And history repeats (#history-repeats)

There's a familiar concept here: software doesn't have a
silver bullet. Certain technologies, methodologies, and
architectures can make things a little better or a little
worse, but the only way to keep things great is to practice
a rigorous level of discipline in their application.

I don't even know where I stand on the subject anymore; the
only right answer is "it depends". Microservices seems like
an obvious mistake at smaller team sizes, but I'm still
largely in favor of them for larger products and
organizations. While many of problems with a large monolith
are difficult to overcome, many of the weaknesses of
microservices seem to me to have plausible mitigations
(even if their implementation might be difficult):

* Prevent stagnant architecture by being very conservative
  when breaking out new services. Only do it across
  boundaries that have a very high likelihood of staying
  stable and being long-lived.

* Prevent language/framework/technology proliferation by
  having a technical practices team assemble and mandate
  certain recommendations.

* Gain operational insight keeping service models simple
  and well-documented, and using technology agnostic
  techniques for visibility like [canonical log
  lines](/canonical-log-lines).

* Have internal tooling teams build and maintain frameworks
  that can plug into a new service and give it turn key CI,
  metrics, monitoring, deployment tooling, etc. Try to
  provide this as its own service so that updates and
  improvements are easy and frequent.

Choosing an architectural model is easy. It's building an
engine to consistently guarantee a high level of
organizational rigor to support it that's hard.

[galapagos]: https://en.wikipedia.org/wiki/Galápagos_syndrome
