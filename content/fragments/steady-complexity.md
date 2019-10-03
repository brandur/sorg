+++
hook = "TODO"
published_at = 2019-10-03T14:37:32Z
title = "The steady march of complexity"
+++

It's not a new idea that complexity is the Achilles’ heel
of large software projects, but it’s such an important
point that I’m going to give it my semi-annual drumbeat 

It’s not intuitive just how bad this problem gets as
projects trend bigger, especially in corporate environments
where there's a low bar on quality (short-term shipping is
favored over long-term sustainability). I’d go so far as to
say that after some point the majority of every engineer’s
time is being attritted to complexity — working around it
is what most people are doing most of the time.

As a recent example from my own life: We have a data
deletion facility that allows a user to delete the entirety
of the test data in their account. It works by rotating
through a series of model types, querying a user’s account
for each one, and deleting all the objects that are found.
I was making a change yesterday to add a couple new model
types to the process after a recent change had made them
user visible.

It went a little off the beaten path because the new model
types were of an ephemeral sort; they’re still stored the
same as any other, but historically weren’t deleted by
convention. It should still have been straightforward, but
upon digging in, I found that the team that owned their
base type had created a series of save hooks that didn’t
support the entirety of the save interface. It was a
simplification that was strictly incorrect, but one that
was possible because the models were being deployed in a
limited sense.

This still should have been okay because I was just running
a deletion instead of a save, but elsewhere in the codebase
a different team had installed data redaction save/delete
hooks that had the (probably unintended) side effect of
converting all delete operations into save operations with
a special `op:` `:``delete` directive to flag it as a
deletion to the underlying machinery. This of course is
part of the extended slightly-less-common save API which
the limited API of the first save hooks didn’t support.

It was still tractable, but required open-heart surgery
deep in the plumbing. A 15 minute project turned into four
hours. Four hours of lost time and productivity during
which no progress is made on macro projects. Theoretically,
I shouldn't even be working on this sort of thing, but
generally these sorts of minor product bugs don't get fixed
otherwise.

And it's not an outlier — accidental difficulty is the norm
for most things anyone tries to do, and over time only
becomes more normal. The teams putting in the problematic
components above were nominally doing the right thing at
the time, but every new feature that breaks well outside
the bounds of its area of responsibility becomes deadweight
on every future change. These features are almost always
strictly additive, and introduction of new ones accelerates
as engineering teams grow.

Taming complexity is a hugely important, _hugely_ unsolved
problem. Again, corporate software tends to be particularly
bad, but even well-designed projects still have the
problem, albeit to a lesser extent [1]. The solutions
aren’t new or novel — more modularity, more encapsulation,
smaller APIs between modules — but as a profession we need
to develop better instincts in these areas, and better
frameworks to force the issue.

[1] e.g. Postgres: try adding a new feature to the B-tree
implementation, and you may be amazed by the vast amount of
context you need to ingest before even being able to get
started.
