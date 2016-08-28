---
title: OSX and Entropy
published_at: 2016-07-31T17:39:44Z
hook: UNWRITTEN. This should not appear on the front page.
---

Bertrand Serlet stopped by Stripe yesterday to do a fireside-chat style
interview with Patrick. Bertrand's a well-traveled former SVP of Apple, veteran
of PARC and NeXT, and quite a lively and entertaining speaker.

Among the ideas that he expressed, one of my favorites was that software needs
a certain amount of constant "anti-entropy maintenance". Left alone, it'll tend
to get slower over time as new features and inefficient bug fixes are
introduced, and this builds toward death by a thousand cuts. For every big OSX
release, Apple committed a baseline level of engineers to fight the effect,
along with the normal pool of engineers to do feature work.

This couldn't have rung more true with me; all software starts out good, and
all software ends up bad [1]. The best we can do is maintain our projects to
the highest possible standard in the interim to stave off that fall into
darkness as long as possible. Some organizations never acknowledge the problem,
some implement periodic clean-up phases that attempt to improve code quality
retroactively, but the best possible model is the one that Bertrand suggest,
where a permanent staffing of engineers to apply constant backpressures to keep
the overwhelming entropic forces at bay.

[1] That's "all" with a few exceptions; most notably, OSS projects that can
    afford to commit a disproportionate amount of time in the interest of
    archieving the highest quality standards. But this level of expensive
    attention to detail is vanishingly rare within a resource-constrained
    corporate environment.
