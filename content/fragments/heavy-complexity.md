---
title: Heavy Complexity
published_at: 2016-01-25T06:12:05Z
---

The article [_the Sad State of Web Development_][article] may hyperbolic to the
point of excess, but it's more right than it is wrong. It's all too often these
days that I realize I need twenty gems to get a Ruby web app up and running, or
that installing a project's frameworks takes ten times as long as the bug fix
that I'd set out to accomplish, or that I can finish an app without a framework
in a quarter of the time it would've taken me to read its "quickstart" docs.

The author gets to the heart of the problem here:

> Javascript has zero standard library, so as soon as you npm init a new Node
> app you better install at least 15 terabytes of modules to get 1/16th the
> standard lib of something like Ruby or Go.

The wider programming community hasn't yet fully realized the collateral damage
of an overly minimal or poorly designed standard library. While it sounds like
a noble principle that the community should be empowered to create its own
solutions, this leads to an inevitable fracturing of techniques that will never
be resolved. The generation of popular dynamic languages like Ruby, JavaScript,
and Python thought that they were solving the mistakes of their predecessors by
being largely "batteries included", but they didn't go far enough. They also
got their interfaces wrong; best demonstrated by the HTTP libraries of both
Python and Ruby.

Finding that perfect balance between functionality and simplicity is incredibly
difficult, and nobody's gotten it right so far [1]. As developers, all that we
can do is endeavor to build tools that are powerful, composable, and
practically effective, while rejecting overdesign and complexity as best as we
can.

[1] Go's probably gotten the closest so far, but even there they managed to
botch very important things like their HTTP handler signature despite having a
lot of years up on the competition to think about it (i.e. in that it's missing
any kind of request context). Everything's a work in progress.

[article]: https://medium.com/@wob/the-sad-state-of-web-development-1603a861d29f#.rdgs64qz9
