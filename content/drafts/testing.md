---
title: Building a Great Test Suite
published_at: 2018-12-22T15:17:40Z
location: Calgary
hook: TODO
---

I've always had opinions on how to write good test tests
for software, but I never expected to care about it quite
as much as I've come to over the last couple years. A
poorly conceived test suite can be the one difference
between making software development the most rewarding
profession in the world versus a hellish grind where
engineers are chronically busy, but not productive.

Anyone can pile on a heap of poorly targeted tests that are
more or less test the feature set of a program, but are
extraordinarily painful because they run slowly, fail
badly, or intermittently produce wrong results. When
building your test suite, you should aim for *nothing less
than great*. Investing a little more in the beginning to do
so will return that effort 1000 times over the coming years
in increased productivity and job satisfaction.

Building a great test suite over a bad one isn't even that
much more work. It largely requires a little additional
upfront awareness, care, and thought to make sure
conventions are established that will scale out gracefully
as a codebase and team of engineers working on it increase
in size by orders of magnitude.

A great test suite has these properties:

* It's **fast**.
* It's **transparent** when it fails.
* It's **reliable**.

They're invariant regardless of the programming language or
technology stack in use. They were true for writing in C
programs in 1979, Ruby programs in 2009, or will be true
for writing Rust programs in 2019.

## Speed as a service (#fast)

### Fewer tests (#fewer-tests)

Interpreted languages.

## Failing transparently (#transparent)

I see an error like this one a hundred times a week:

    FAILURE!
    Expected status: 200
    Actual status: 400

For all intents and purposes, this is the same message as a
simple, "error: test failed". It gives me no information on
what happened, no idea of where the problem might actually
be, and no leads on what to do next.

Unit tests again. Get closer to the error.

## Exhaustive reliability (#reliable)

## Recap (#recap)
