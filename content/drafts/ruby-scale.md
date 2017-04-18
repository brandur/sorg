---
title: Ruby, and its Problems of Scale
published_at: 2017-04-18T14:23:28Z
location: San Francisco
hook: TODO
hook_image: true
---

I know a fair number of Ruby evangelists from my time at
Heroku, and 

## Right, I've already heard it's slow (#slow)

Right, Ruby _is_ slow and resource intensive.

## Beyond the shell script (#beyond-the-shell-script)

### Zero information pre-runtime (#runtime)

This forces us to come up with all kinds of creative hacks:

* CI suites that try to test code to an absolute exhaustive
  extent, sometimes even aiming for 100% line coverage.
* Heavy reliance on except tracking services like Sentry
  and Rollbar to tell us what's happening. Errors in
  production aren't an "if"; they're an absolute certainty.
* Canary deploys to help detect problems early, and contain
  the fallout in the inevitable case of a bad deploy.

### The edit-compile-debug cycle (#edit-compile-debug)

Zeus

Without Zeus, large Ruby codebases would quite simply not
be tenable. Many Ruby advocates don't know this because
you'll never notice the problem with a few hundred lines of
code and a couple dependencies.

### Boundary bleeding (#boundary-bleeding)

!fig src="/assets/ruby-scale/knot.jpg" caption="Without constraints, boundaries tend to dissipate, especially within larger engineering teams."

## Other languages (#other-languages)

But doesn't this apply to JavaScript/PHP/Python as well?
Yes, it absolutely does. Some of those are trying to make
some incremental progress like Hack or Python's TODO.

To a large degree, the age of the 90s era of dynamic
languages was an academically interesting learning
experience for the industry, but one that we should be
aiming to move beyond.

The killer "future stack" has a lot of good candidates
right now with languages like Go, Rust, Swift, Elixir, and
Crystal all very exciting in their own ways. Many of these
share some characteristics that the old dynamic guard
don't: compilers, strong typing, performance (with
compilers and standard libraries that can be implemented in
the language itself instead of C), and focus on fast
edit-compile-debug loops.

## Towards a brighter future (#brighter-future)

Ruby does have a few advantages. Pry is probably the best
REPL in the world, and it helps immensely while debugging
or trying to any kind online introspection.

## Summary (#summary)

I decided to write this not so much in disdain for the
language, but more because I wish other people wrote about
experiences in the real world more often. It's easy to get
the impression from meetups and conferences that languages
and technologies are flawless, but that's almost never the
case.
