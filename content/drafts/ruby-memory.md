---
title: How Ruby Allocates Memory, and the Limits of
  Copy-on-write
published_at: 2017-08-24T13:39:04Z
hook: TODO
---

Anyone who's run Unicorn (or Puma, or Einhorn) will have
noticed a curious phenomena. Worker processes that have
been forked from a master start with low memory usage, but
before too long will bloat to a similar size as their
parent.

This is in spite of the ***copy-on-write*** (COW) features
provided by the virtual memory systems in modern operating
systems. Startup and runtime overhead is reduced by having
forked children share the same memory space as their
parent. Only when a child modifies some piece of shared
memory does the OS intercept the call and copy the page for
exclusive use by the child.

TODO: UNICORN DIAGRAM.

So what's going on here? Most programs have a sizeable
collection of static objects that are initialized once, and
sit in memory largely unmodified throughout its entire
lifetime. Child processes should have no problem sharing
that collection with their parent. To get to the heart of
the problem, we'll ahve to dig into how Ruby allocates
memory.
