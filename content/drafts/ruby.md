---
title: The Challenges of Building With Ruby at Scale
published_at: 2017-04-07T14:49:57Z
location: San Francisco
hook: TODO
---

I decided to write this not so much in disdain for the
language, but more because I wish other people wrote about
experiences in the real world more often. It's easy to get
the impression from meetups and conferences that languages
and technologies are flawless, but that's almost never the
case.

I know a fair number of Ruby evangelists from my time at
Heroku, and 

## Right, I've Already Heard It's Slow (#slow)

Resource intensive

## You Don't Know Anything Until Runtime (#runtime)

This forces us to come up with all kinds of creative hacks:

* CI suites that try to test code to an absolute exhaustive
  extent, sometimes even aiming for 100% line coverage.
* Heavy reliance on except tracking services like Sentry
  and Rollbar to tell us what's happening. Errors in
  production aren't an "if", they're an absolute certainty.
* Canary deploys to help detect problems early, and contain
  the fallout the inevitable case of a bad deploy.

## 

The edit-compile-debug cycle

Zeus

Without Zeus, large Ruby codebases would quite simply not
be tenable. Many Ruby advocates don't know this because
you'll never notice the problem with a few hundred lines of
code and a couple dependencies.

## The Gordian Knot

## Other Languages

But doesn't this apply to JavaScript/PHP/Python as well?
Yes, it absolutely does.

## Towards a Brighter Future

Ruby does have a few advantages. Pry is probably the best
REPL in the world, and it helps immensely while debugging
or trying to any kind online introspection.

That said though, I think to a very large degree, the age
of the 90s era dynamic languages was an interesting
learning experience for the world, but one that we should
be aiming to move beyond.

New languages are still rolling out at an pretty good clip,
Go, Rust, Swift, Elixir, and Crystal for example all look
like interesting possibilities for future stacks, and
although the future isn't clear, its worthwhile noting that
this next generation of languages share many
characteristics that the old dynamic guard don't:
compilers, strong typing, and interest in fast
edit-compile-debug loops.
