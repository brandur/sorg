---
title: Constructive Criticism for the Rust Language
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
start the process all over again. Code's never done -- it's
just (mostly) working for today.

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
you have to write until there's nothing left. A true type
disciple's dream is to be able to write out one glyph and
have it expand into a fully runnable version of Quake II
through a long chain of macros long enough to wrap around
the world a few times.

Or is it? Complex language mechanics have a cost, and that
cost can roughly be described as the bar that needs to be
vaulted in order to get productive in it. If a type system
is so elaborate and so complicated that it's hard to get a
program to compile, people give up. Potential developers die
before they reach adolescence in a hostile environment of
type parameters nested eight levels deep and inscrutable
errors generated from bad macro invocations.

Let's take a look at an example. [Diesel][diesel] is an ORM
written for Rust. It has a multitude of great features and
it does certain things better than any comparable system
before it (including ActiveRecord, Sequel, etc.). One neat
feature is that you can compose queries with its typed DSL,
and it will catch malformed queries at compile time. 

```
error[E0271]: type mismatch resolving `<schema::podcast_feed_locations::table as diesel::query_source::AppearsInFromClause<schema::podcast_feed_locations::table>>::Count == diesel::query_source::Never`
  --> src/mediators/podcast_reingester.rs:38:18
   |
38 |                 .load::<Vec<(i64, String, String)>>(
   |                  ^^^^ expected struct `diesel::query_source::Once`, found struct `diesel::query_source::Never`
   |
   = note: expected type `diesel::query_source::Once`
              found type `diesel::query_source::Never`
```

I call this "Haskellification" because while to some
Haskell is the most beautiful language ever created, to
most it's an arcane, unapproachable monster. Like it or
not, it's not how you should ever want your new language to
turn out because a barrier to entry higher than Everest is
an absolute guarantee that history will remember it as
intellectual curiosity used by no more than a few
eccentrics instead of a mainstream practical tool.

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

[diesel]: 
