---
hook: Notes on the language after spending a few weeks building a large project in
  it.
location: San Francisco
published_at: 2016-03-31T20:37:11Z
title: Notes on Go
---

Despite having worked for so many years with colleagues who were major
proponents (or even contributors) to Go/Golang, I'd somehow gone this long
without having ever written a production grade service in the language, having
only dabbled so far with toy projects, small scripts, and minor contributions
to existing software. That all changed in the last few weeks, where I had the
opportunity to assemble a pretty serious project in the language from scratch.

I took notes throughout the process so as not to lose my (relatively) fresh
outlook on the language.

## The Important Things (#important)

### Simple, but Verbose (#verbose)

Building new programs requires typing **a lot**. The language is incredibly
verbose and has few shortcuts. The upside is that once you have typed out
that initial code, it's eminently readable and relatively easy to maintain
compared to many other languages.

All too often, the bar to understanding projects written in Ruby, Lisp,
Haskell, Rust, C#, C++, or whatever, isn't _just_ figuring out the code, it's
also deciphering the localized (and often overcomplicated) abstractions that
every developer ends up baking into their code to reduce line count, and in
many cases that's significant cognitive overhead. In Go that problem doesn't
exist to anywhere near the same extent.

### Concurrency (#concurrency)

After spending some time with them, I'm firmly convinced that green threads
(Goroutines) and channels are the only way that concurrency should be exposed
to developers.

When working in languages like Ruby (to pick an example of one that I'm very
familiar with), even with experience, doing any work with concurrency is
incredibly frustrating in just the sheer number of problems that you're likely
to run into. It's tempting to think that this is because concurrency is
inherently difficult, but it's more to do with dull primitives that are error
prone by default.

By comparison, when working in Go, it's amazing how many programs you can write
where your concurrent code will work _perfectly_ the first time. I also find
that even in cases where it doesn't, it's far more often due to a conceptual
mistake that I've made than it is to a poorly designed language feature.

I also appreciate just how opinionated the Go team has been on this front.
Other languages with strong concurrency stories like Haskell and Rust have
opted to give users access to every type of primitive under the sun, and in the
long run that tyranny of choice leads to an ecosystem of mixed ideas and no
clear way forward.

### Speed (#speed)

Speed is absolutely critical, and not just for the runtime, but _for the
tooling_. Being able to compile and run your entire test suite in under a
second [1] changes the entire development experience in such a fundamental way
that it's hard to adequately describe. After working with a Go project for a
while, going back to 10+ second iteration loops in languages like Ruby feels
like trying to run a marathon while waist deep in a bog.

This is still a key place where Go stands apart even from other modern
languages which tend to focus on runtime speed or sophisticated features while
ignoring the development cycle [2].

But Go is also fast at runtime too. It's nice to be able to write code in a
high-level language and be able to trust that it will run quickly.

### Deployment (#deployment)

If every language was as easy to deploy as Go, Docker would never have been
invented (maybe a slight exaggeration, but the need for it wouldn't have been
anywhere near as dire). Build a binary. Copy it to a target server. Stop the
old service and bring the new one up. That's it. No weird environment problems.
No dependency headaches. No Bundler.

I now write all my throwaway scripts in Go for the same reason. If I ever need
to run one with Cron, I know that I'm never going to have to deal with issues
with `$PATH` or rbenv or anything else. Copy the executable to
`/usr/local/bin`, inject it straight into my Crontab, and you're done.
`killall` even works; incredible.

## Other Notes (#other)

### The Good (#good)

There's a lot to like about Go:

* **Defer:** I love this abstraction. Although not quite as safe as something
  like a C# `using` block (in that you might accidentally remove the line and
  not notice), it's far less cluttering.
* **Import:** I'm firmly convinced now that importing packages with a short
  canonical identifier (e.g. `fmt` or `http` from "net/http") and then having
  only have one option for referencing that package in code is the One True
  Way. No more symbols with unknown and dubious origin (Haskell) or artisanal
  blends of qualified and non-qualified names (C#/Java/other).
* **Select:** Although decisions like using `default:` to make a `select`
  blocking or non-blocking are a little obtuse, overall this construct is
  incredibly powerful.
* **Pipelines:** By combining a few of the built-in language features, Go
  enables the construction of composable, scalable [pipelines][pipelines]. This
  approach to parallelism is truly elegant and encourages developers to write
  programs that are more performant and which make better use of modern system
  resources (namely, more cores).
* **Labels:** Incredibly useful for breaking out of an outer loop without
  introducing boilerplate. When used carefully, `goto` is also tremendously
  powerful (and comes with the perfect number of restrictions to prevent its
  abuse).
* **No metaprogramming and minimal OO:** Sometimes the costs of what seem like
  good features on the surface outweigh their benefits. I'll gladly write a
  little more code if it means that someone else will be able to understand it.
* **Static linking:** Go didn't invent this, but they did make it default.
  Static linking introduces some headaches in a few cases, but vastly improves
  the lives of the other 99% of users.
* **Standard library in Go:** It's an amazing feature to be able to check the
  implementation of core packages in the standard libraries. This isn't all
  that unusual for newer languages these days, but it's becoming increasingly
  harder to make the argument that languages like Ruby and Python that insist
  that having a standard library written in C is just fine.
* **Nice documentation features:** Go has some neat innovations in
  documentation that solve real problems that are observable in almost every
  other language and framework. e.g. A locally runnable documentation server
  (great for flights), or [testable examples][testable-examples], which mean
  that examples in documentation get run with the test suite so that they don't
  fall out of date.

### The Surprisingly Good (#surprisingly-good)

There were a lot of facets of Go that I read or heard about before trying and
which I was pretty sure that I wouldn't like. However, after using the language
a while I quickly started warming up to them:

* **Dependency management:** It took me a while to warm up to Go's design
  around dependency management, but not having to run and manage everything
  through a slow and complex system like Bundler hugely improves the
  development experience. It also makes it very easy to jump into foreign
  libraries and examine their implementations when necessary.
* **Gofmt:** Having a single convention for the language makes collaboration
  easier, and makes my own coding faster (in that I can rely on gofmt to
  correct certain things).
* **Errors on unused variables:** These can be very annoying, but I can't deny
  that these error messages have saved me from what would otherwise have been a
  bug multiple times now.
* **No generics:** Having types only on special data structures like slices and
  maps gets you surprisingly far. Although not having generics does make using
  the language for certain things difficult, I was amazed after having built a
  multi-thousand LOC program to realize that I hadn't wanted for them once.

### The Bad (#bad)

I really did make an effort, but even so, some parts of the language are hard
to love:

* **Error handling:** I like that generally my programs don't crash, but
  dealing with errors requires an incredible level of micro-management. Worse
  yet, the encouraged patterns of passing errors around through returns can
  occasionally make it very difficult to identify the original site of a
  problem.
* **The commmunity:** Reading the mailings lists can be still be pretty
  depressing. Every critique of the language or suggestion for improvement, no
  matter how valid, is met with a barrage of either "you're doing it wrongs",
  or "only the Go core team that can have thoughts that are worth
  consideration" [3]. Previously this level of zealotry had been reserved for
  holy crusades and text editors.
* **Debugging:** gdb and pprof both work with Go, but with experiences that are
  rough enough around the edges that you'll find yourself often resorting to
  print-debugging just to avoid the hassle.
* **Noisy diffs:** The downside of gofmt is the possibility of noisy diffs. If
  someone adds a new field with a long name to a large struct, all the spacing
  changes and you end up with a huge block of red and a slow review [4].
* **Quirky syntax and semantics:** Go is littered with quirky language syntax
  and semantics that are fine once you know them, but are unnecessarily opaque.
  Some examples:
    1. The distinction between `new`, `make`, and initialization with composite
       literals.
    2. Interfaces are always references.
    3. Symbols that start with capital letters are exported from packages.
    4. Channels created without a size are blocking.
    5. `select` blocks with a `default` case become non-blocking.
    6. You check if a key exists in a map by using a rare second return value
       of a normal lookup with square brackets.
    7. Named return values.
    8. Closing a channel causes any Goroutine that was listening on it to fall
       through having received a zero value of the channel's type.
    9. Comparing interfaces to `nil` is allowed by the compiler, but apparently
       not a good idea, and can lead to some strange bugs.
* **JSON:** Is [as slow as reported][slow-json] due to its extensive use of
  reflection. This wouldn't seem like it should be a problem, but can lead to
  surprising bottlenecks in otherwise fast programs.

### The Ugly (#ugly)

There are very few parts of the language that are unapologetically bad, but
that said:

* **Assertions:** Although mostly palatable, the omission of a collection of
  meaningful assert functions (and the corresponding expectation that a
  custom-tailored message should be written every time you want to check that
  an error is nil) isn't great to say the least.
  
  The real problem though is that the unnecessary verbosity of tests acts as a
  natural deterrent to writing them, and projects with
  non-existent/poor/incomplete test suites are a dime a dozen. I've been using
  the [testify require package][testify] to ease this problem, but there should
  be answer in the standard library.

## Summary (#summary)

Overall, I never quite reached the level of feverish passion for Go that many
others have, but I consider it a thoroughly solid language that's pleasant to
work with. Furthermore, it may have reached the best compromise that we've seen
so far in the landscape of contemporary languages in that it imposes
constraints that are strong enough to detect large classes of problems during
compilation, but is still fluid enough to work with that it's easy to learn,
and fast to develop.

[1] Without tricks like Zeus that come with considerable gotchas and side
    effects.

[2] e.g. Rust, or, and it hurts me to say this, Haskell.

[3] The best single example of this that I've found so far is [a request for a
    non-zero exit code in Golint][golint]. The community articulates the
    problem and shows an obvious demand and willingness to help. Meanwhile the
    member of Go core can't manage to build even a single cohesive
    counterargument, but even so, the issue along with all its ideas and
    suggestions are summarily rejected.

**Follow-up (2016/06/24):** Russ Cox re-opened the issue (presumably after
seeing complaints on Twitter given that the original maintainer had locked
discussion on it) and it was subsequently resolved.

[4] `?w=1` on GitHub to hide whitespace changes helps mitigate this problem,
    but isn't the default, and doesn't allow comments to be added.

[golint]: https://github.com/golang/lint/issues/65
[pipelines]: https://blog.golang.org/pipelines
[slow-json]: https://github.com/golang/go/issues/5683
[testable-examples]: https://blog.golang.org/examples
[testify]: https://github.com/stretchr/testify#require-package
