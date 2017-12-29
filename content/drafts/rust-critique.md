---
title: Constructive Criticism for Rust
published_at: 2017-12-29T22:17:27Z
location: Calgary
hook: TODO
---

Writing Rust code is like writing a poem. Writing a poem is
difficult: every word is agony in that you have to find the
perfect one. But once the labor is done, it's _done_. Every
word fits and the product as a whole is a work of art. You
can take it and hang it on a wall knowing that you'll be
happy with it for years to come.

Most code isn't like this. You write it once and send it
through the interpreter a few dozen times to get it roughly
right. You write some tests and eke out the problems you
didn't find in the interpreter. You send it to production
and redeploy a dozen times to fix all the new edge cases
that you find there. When you introduce threading, you
start the process all over again.

## Leftpad-ification of the ecosystem (#leftpadification)

Three days into a simple project, I had accumulated X
dependencies.

Dependencies have a cost. Everyone one of those risks
falling out of maintenance and into delapidation.

It also introduces dependency hell problems. I had trouble
getting a basic Iron/Juniper setup going because there'd
been enough API changes that you had to lock to specific
versions to get a program compiling.

## Haskell-ification of the language (#haskellification)

To programming language academics, the more the better.
Higher order types. More powerful macro systems.

A language that's powerful enough ephemeralizes the code
you have to write until there's nothing left -- put a few
symbols down and a chain of macros long enough to wrap the
Earth a few times expands it into a fully runnable program.
Perfection.

Or is it? Complex language mechanics have a cost, and that
cost can roughly be described as the bar that needs to be
vaulted in order to get productive in it. If a type system
is so elaborate and so complicated that it's hard to get a
program to compile, people give up. Future programmers die
before they reach adolescence in a hostile environment of
lifetime parameters, move semantics, and ownership
structures.

I call this "Haskellification" because while to some
Haskell is the most beautiful language ever created, to
most it's an arcane, unapproachable monster. Like it or
not, it's not how you should ever want your new language to
turn out because a barrier to entry higher than Everest is
an absolute guarantee that history will remember it as
intellectual curiosity used by no more than a few
eccentrics instead of a mainstream practical tool.

## Haskell-ification of the documentation (#docs)

The types are so sophisticated that you don't need any
documentation beyond a function signature, right? Wrong.
Absolutely, definitively, wrong.

Rust has introduced some nice features like runnable code
snippets in documentation and a built-in `examples/`
directory, but the documentation for most of the standard
library and most packages is still terse and unfeeling.

[Staring at a page full of complex type glyphs][badtypes]
is enough to make anyone run the other way in fear.

I didn't use even one package that didn't require that I
clone the source code and start searching through its tests
to try and figure out how to use it. This goes for even
headliner packages like Diesel, Juniper, and Hyper.

## Concurrency (#concurrency)

In my opinion Rust's concurrency story is the number one
thing that makes me not want to use the language and fear
for it's future. Bar none. It's not even close.

Hyper example:

``` rust
```

Async/await example:

``` rust
```

Red-green colored function link.

Compared this to Go and it's a night and day difference.
Even if it's possible to make a Rust package that's
theoretically faster than a Go package, it doesn't matter.
The ergonomics of Go's approach are better for productivity
and comprehensive by a thousand-fold.

[badtypes]: https://todo
