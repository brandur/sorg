---
title: "Time and Isolation: Divergent Ecologies in Software"
published_at: 2017-07-31T17:47:39Z
hook: TODO
location: San Francisco
---

This isn't my usual technical piece. It's more of a short
parable and a few opinions that I've developed after having
observed for a while how software is built at large companies.

---

Custom test conventions -> slow tests

Slow tests -> custom parallel CI

* Once you have CI in place, slow tests -> slower tests

Parallelized CI is regarded as a magnificent advancement
and the engineers involved take a moment to bask in the
light of their accomplishment. But before long, a strange
effect starts to emerge: the test suite is still getting
slower, but far _faster_ than before. How could this be? It
becomes clear that now that developers are entirely reliant
on parallel cloud infrastructure to run their tests,
there's no longer much in the way of negative feedback when
they write slow tests -- it's all going up into the cloud
anyway. Slow tests being to proliferate at a rate that
would've been previously unimaginable. Every new one that's
added makes the possibility of a future fix less likely.

Custom parallel -> frequent master breakages

* Two checkboxes in Travis.

TODO: Screenshot.

Frequent master breakages -> custom merge flow

The Jenkins build becomes so complicated that only a
specialized few can understand and fix it. It's as fragile
as a porcelain cup, and even minor changes to its
underlying infrastructure, like AWS applying kernel
mitigations for Meltdown/Spectre, are enough to take it
offline for days.

Another side effect is that build output is hundreds of
pages long and difficult to parse. A new layer of custom
software is written so to help developers locate the
relevant lines for test failures.

Customizing an editor that lets a developer be reasonably
productive becomes a difficult enough task (even basic
tasks like listing files in the project or pushing a Git
branch are multi-second operations) that the organization
provisions one on new computers that's preconfigured.
Eventually, it's deemed to be lacking certain features, so
the editor is forked and further customized. More time
passes, and it's deemed that the organization's
requirements are so specialized that other editors are
unsuitable. Support for them is dropped, and the
organization's custom fork becomes the only allowed path.

---

Developers who join the organization are like Darwin
stepping off the Beagle onto the Galapagos islands, and
find an exotic ecosystem of tooling they've never seen
before. They'll spend weeks learning the basics to get
up-to-speed, and if they spend a long time there, there'll
always be large dark corners that they'll never fully
understand.

Similarly, those will leave will find that the level of
customization will mean that little of what they learned is
transferable. Those unlucky enough for it to have been
their first job out of college might find themselves a
little like the Eloi -- they're familiar with how to use
their local tools, but have a poor grasp of the underlying
fundamentals because they've been so far abstracted away
from them for so long.

---

Lots of Ruby code -> slow startup

Slow startup -> Zeus

Zeus -> custom bootloader

Lots of Ruby code -> poor modularity

Poor modularity -> Static analysis

Static analysis -> custom require/build system

Custom require/build system -> custom bootloader scripts

## Macro-scale effects are difficult to unwind (#macro)

It might be possible to reverse the effects of climate
change through macro-level manipulation in the future, but
it would've been easier to address the problem sooner by
building more sustainable cities, encouraging better
industry/transport structures, and implementing economic
policies that incentivize the right practices.

The same goes for great pacific garbage patch, or for ocean
acidification. Maybe you can reverse those effects with
future technology and resources, but it would've been
better just not to have brought them on in the first place.

## The moderate path (#moderate)

Use Travis.
Have a fast build.
Use languages that provide type annotations.
