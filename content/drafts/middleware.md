---
title: Middleware Boundary Bleeding and Enforced Modularity in Rust
location: San Francisco
published_at: 2018-03-01T18:26:25Z
hook: TODO
---

Lately I've been playing with the idea with building a web
application in Rust. I don't know whether it's a good idea
-- the language is still new, still pretty impractical due
to a steep learning curve and complex type system, and
major components like a concurrency story are still under
active development and seeing rapid change.

I'm doing it because I have a question above software that
I want answered: can we build more reliable systems with
programming languages that provide better checks and
stronger constraints? I've been working on large projects
in interpreted languages for years and the neverending game
of whack-a-mole is tiring. You fix one problem only to have
a new one pop up somewhere else. Any new development comes,
no matter how careful, produces a stream of bugs that will
be shored up over the course of years.

Rust is interesting because its sophisticated type system
unlocks the potential for very strong compile-time checks
-- for example to check the correctness of an SQL query
with [Diesel][diesel] or an HTML view with
[Horrorshow][horrorshow]. A pedantic compiler isn't a new
idea, but I'm also attracted to Rust because of its modern
toolchains, strong conventions, pleasant syntax and
explicitness, and the attention to detail put into its
upfront design (compared to say a Haskell, which in this
author's opinion, has lapsed on most of these fronts).

One are that I've looked into lately is web middleware.
Middleware is a great idea -- perfectly modular components
that can be written once and shared between projects, but
like many good ideas, in practice they still often violate
encapsulation and lead to bugs. I found some of the ideas
for middleware in Rust's web frameworks to be pretty good
at counteracting this effect. Let's take a closer look.

## The promiscuity of unchecked keys (#unchecked-keys)

TODO

I work on a project that has ~50 middlewares and reordering
them is perilous.

## Middleware modules and types in Rust (#rust)

## Safety-by-convention is not enough (#safety)

[diesel]: https://github.com/diesel-rs/diesel
[horrorshow]: https://github.com/Stebalien/horrorshow-rs
