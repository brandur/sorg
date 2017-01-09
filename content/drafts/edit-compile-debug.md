---
hook: How a fast edit-compile-debug cycle is the most
  important in a high quality developer experience and in
  maintaining high levels of productivity.
location: Calgary
published_at: 2017-01-05T17:18:18Z
title: Edit-Compile-Debug
---

You lose the ability to be productive when making changes
to your own product.

At some point you cross a threshold where instead of just
waiting for your code to build or your tests to run, you go
and find something else to do while that's happening.
That's the danger threshold here productivity is reduced by
10x. You've now context switched away from what you were
working on and will have to eventually context switch back
to it.

Test suites in CI. Make sure they're runnable locally.

## Ruby (#ruby)

This is the biggest disconnect I notice between people who
talk about writing Ruby (i.e. at conferences and the like)
and people who actually write Ruby. The former don't know
that for a large enough codebase, the time it takes just
for the code to be interpreted and start running is so
extreme that you _need_ a program like Zeus to keep even a
nominal level of productivity. As a member of the latter
camp, I feel that Zeus (or something like it) should be a
part of the core language.

You never hit this with smaller projects, which can
generally start up and run their test suites in a few
seconds or less. You can even largely avoid it for larger
(and popular) OSS projects, that have more time and
contributors to get things exactly right compared to a
private company building products.

Ruby/Zeus.

Slow test suites.

[rust]: https://github.com/aturon/rfcs/blob/roadmap-2017/text/0000-roadmap-2017.md
