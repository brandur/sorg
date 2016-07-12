---
title: Go Relax
published_at: 2014-06-28T20:16:22Z
---

I've been writing a lot of Go code over the last week, and I'm finally coming
around. I've been previously kept at arm's length due to its lack of generics,
primitive dependency management system, and questionable enforced conventions
(like workspaces), but a deep dive into the language has gone a long way to
convince me.

Go's may great features have already been covered to exhausation elsewhere:
speed, conurrency, minimal syntax, fast compilation, pragmatic OO approach,
nice standard library abstractions, and so on. I'll skip talking about those
for now.

What struck me most about Go was how relaxing it was to use. Relaxing in the
sense that I don't have a constant fear that I'm making the wrong decision
about design or implementation. I don't think about whether I should be
forking, using OS threads, a callback system like EventMachine, fibers, or a
higher level actor framework like Celluloid. I use Goroutines and channels, and
it's fine. I don't worry about which of six different HTTP clients have the
most technical merit. I use the standard library, and it's fine. I don't worry
about which style of loop will fit best relative to local conventions or
whether I should be writing in a more functional style. I use `for` and
imperative style, and it's fine. I don't think about exception control flow or
how granular my error subclasses should be. I use the `error` convention, and
it's fine. I don't think about whether I should be running it on the reference
implementation, or the JVM for better parallelism. I pull down the official
runtime, and it's fine.

In so many cases, basic decisions that were made during a language's design
phase balloon out into big unexpected side effects. Those might look like
anything from added layers to wrap the standard library to alternate runtime
implementations that work around defects in the original. Go didn't do anything
special here, but as a more recent language, did base design based on pitfalls
observed elsewhere over time.

Go's real winning decision was the convention. As a developer, I don't want to
spend my time agonizing over which concurrency strategy I should use. I want to
see one obvious choice which is well-designed and effective. Go has gone all in
on this strategy.

It's far from perfect of course, but it doesn't matter because Go is good
enough --- unlike so many other languages, the wheel doesn't need to be
re-invented just to make it palatable. This has the added side effect of making
almost any Go program easily readable.

I'll continue to remain a steadfast proponent of other great modern languages
like Rust and Swift, but Go has a bright future as a balanced approach to
engineering for big and small problems alike.
