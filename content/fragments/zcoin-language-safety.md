---
title: Learning About Language Safety From Zcoin
published_at: 2017-03-07T18:11:03Z
hook: Is it irresponsible to start projects in C/C++?
  Probably, yes.
---

A [recent bug in Zcoin][bug] allowed an attacker to mint
~550,000 new coins, the theoretical equivalent to
USD$750,000 (and about 25% of the entire Zcoin supply). As
is common with a C++ codebase, it turned out to be an
easy-to-make and hard-to-spot typing problem that the
permissive compiler had no qualms about letting happen
(namely, a mistake of one extra character and use of a `==`
equality operator instead of an `=` assignment operator).

One user on the HN comment thread asked the question that
as an industry we should all be asking ourselves: "Given
the well known pitfalls, is it irresponsible to start
projects in C/C++?"

As is pretty normal when C++ is criticized, its proponents
started coming out of the woodwork to pronounce all the
things that the author had done wrong. He should have
written more modular code! He should have been using
`-Wall`! No, he should have been using `-Wall -Wextra
-Werrors`! He should have been running Valgrind in CI!

They're missing the point entirely. If it's possible to
introduce bugs of this severity just because you haven't
divined the use of the right tool or perscribed to a
particular development doctrine, then the language is _not
safe_. No language's users will ever be perfectly educated,
and as a result this kind of problem is bound to repeat
itself until the end of time wherever it's possible for it
to happen.

Consensus on values around language design is hard to come
by, but the features that newer languages all agree on
seems to suggest some convergence. Compilers with strict
checks are good. Types are good. Compile-time and runtime
performance is good. Concurrency primitives are good.
Memory safety is good. We should be aiming to write
programs that are more likely to be bug-free by moving
towards languages that provide these basic utilities.

[bug]: https://news.ycombinator.com/item?id=13807693
