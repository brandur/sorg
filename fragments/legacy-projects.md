---
title: Legacy Projects & the Maintainer
published_at: 2015-08-14T05:27:58Z
---

Code has an incredible ability to survive far beyond the expectations of its
original authors. There are a multitude of reasons for this, but foremost is
simply that at some point software becomes so feature-rich that trying to
reimplement it would be a costly endeavor, and would provide no guarantee that
the new software would be superior quality to its predecessor. There are many
reasons that your bank still programs modules in COBOL, but the language's
technical merits aren't one of them.

By extension, almost all of us will at some point (as we start a new job or
take over maintenance of an open-source project) have to work in a codebase
that we would consider to be "legacy", a term that we apply because in our
minds it's not up to the same exacting standards that we would demand were we
to build it again today. Simultaneously though, most of us understand that this
is a falsehood; poor quality software rarely got the way it is due to malicious
intent or sheer incompetence by its original authors, but rather from slow
degradation over time as it continued to accumulate new capabilities that
didn't fit well into the original framework.

Developers understand this on some level, but the natural instinct of the
majority of us is still to treat any less-than-optimal projects that we come
across as a dumping ground for low quality patches. This causes these projects
to succumb to the tragedy of the commons and accelerate their free fall in
quality. This is detrimental to any company or organization -- if a project is
worthy of attention and contribution, it's probably also relevant to the bottom
line.

The obvious advice here is to encourage everyone to treat it like it's their
own and to try and make sure that every new patch leaves the code in a better
state than it'd previously been. This is a good general guideline, but also
simultaneously a platitude. Although unquestionably an admirable aim, a little
like the case of a popular summertime park that's left with excess trash at the
end of the day, it's unlikely that every developer in your organization will
take it upon themselves equally to help improve things.

A more pragmatic step would be elect a maintainer for the project. This might
also seem like an obvious step, but keep this in mind: you're not just looking
for somebody to merge pull requests, you're looking for an overseer who cares
about the project on a deep level, and who strive continuously to help build a
better future for it. Once other contributors realize that patches are now
being rejected on the basis of quality, it doesn't take long for things to start
improving.

Finding this person is the first step; empowering them to do their job
effectively is the second. This involves a few things:

1. Give them the slack time necessary to build improvements.
2. Give them the political capital necessary to apply backpressure even in the
   face of overwhelming force.

It's the second point here which is challenging. Especially in the private
world where hard deadlines are around every corner, it's tempting to justify
inferior quality with a line like, "We just don't have time to do this the
right way right now. We'll improve it later." As most engineers with a few
years of experience under their belts can tell you, that "later" will never
come. The best way to stop that debt from piling up is to stop it in the
present, not to offload it to an imaginary future of relative idleness. A
maintainer must be able to call out problems and apply resistance even in the
face of important company directives.

The key to improving that legacy codebase probably isn't a rewrite. It's a
reversal of its downward trend in quality to one of slow but steady
improvement. If you keep it going long enough, one day you can even drop the
"legacy" moniker entirely and be left with something that's pleasurable to work
with.
