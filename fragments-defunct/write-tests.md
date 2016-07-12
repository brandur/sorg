---
title: Write Tests
published_at: 2015-11-11T16:52:49Z
---

If you walked into a room full of software developers and asserted that writing
tests is a good thing, it might be one of the few things that you could say
which everyone in the room could agree on. More dogmatic methods like TDD may
be contentious, but even proponents of languages that are type-safe to a fault
would agree that maintaining a basic test suite is good practice.

But when the time comes, those same developers will largely fail to do so.
They'll be good about it for those core repositories with a lot of eyes on
them; these projects are "important" and there's a good chance that a reviewer
would call them out without tests besides. But when it comes to bootstrapping a
new project or working on smaller supporting components, all of those best
practices will be wholly disregarded. This seems fine because the surface area
of those projects is still small, and their original author has an absolute
understanding of things inside their head.

Over time those new and small projects become more integral to the whole.
People start using them. They become key pieces to supporting the larger
infrastructure. Given their newfound status, contributions to them might have
included tests if a suite already existed, but will rarely start one from
scratch if it doesn't.

New developers lacking historical context join the company. The original
authors leave. What's left are critical components that are still nominally
functional, but which are difficult to improve because it's hard to tell that
changes don't break anything. Eventually a component _has_ to be changed, and it
causes an outage once it is.

You don't write tests for yourself. You write them for your colleagues. You
write them for the generations of engineers that haven't joined the company
yet. You write them for the "you" of the future who doesn't know what the "you"
of today was thinking. Help prevent time's erosion. Write tests.
