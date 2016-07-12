---
title: Life Without Mediators
published_at: 2016-04-27T16:11:28Z
---

In a small program, the permissiveness that makes Ruby and other dynamic
languages famous acts as a catalyst that hugely accelerates its development,
but as the program grows, this same permissiveness becomes a downside as every
module is allowed to call every other module, and without careful curation a
sane model for organization won't emerge naturally. In the worst case, every
module will be calling every other module, and the program's flow becomes a
rat's nest of spaghetti code that can only be made sense of by stepping through
it with a debugger.

I've previously suggested the use of [mediators](/mediator) to overcome this
problem, a pattern that encourages all major program flows to be encapsulated
in PORO classes that read like lists of instructions, and control the
interaction between primitive types like database models.

It's an interesting experience now developing in the largest Ruby codebase I've
ever seen (~1M lines of code), and one where the use of a pattern like
mediators never developed. Domain logic can be, and therefore is, implemented
in any number of different places:

* In models.
* In API endpoints.
* In base classes (on either models, endpoints, or other).
* In mix-ins (on either models, endpoints, or other).
* In various static utility classes.
* In internal gems added as dependencies.

But the larger problem is that it's rarely just one of these; instead, logic
gets strewn amongst all of them in dozens or hundreds of small complex
interactions across countless small modules.

Another side effect is that because so much domain logic is locked up in API
endpoints, it's common to rely heavily on functional tests that use rack-test
to make mock HTTP calls to adequately test it. It's not the end of the world,
but has the effect of making debugging problems inordinately difficult (because
exceptions come back as status 500 responses), and makes the development cycle
_very_ slow due to the extra overhead of the HTTP stack in every test case.

Possibly the most important thing you can do when starting a new Ruby project
is to get a handle on the mediator pattern (or one like it) right away because
it's very work intensive to get backported. It's also a somewhat unnatural
direction given that the idea is not present in major frameworks like Rails,
and more generally, is also missing from MVC.
