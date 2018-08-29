---
title: Ruby, and Challenges of Scale
published_at: 2017-04-18T14:23:28Z
location: San Francisco
hook: The challenges of scaling and operating a big Ruby
  codebase (that are not related to performance).
---

Ruby is a beautiful language. Speaking from experience,
it's easy to get attached to everything from its pleasantly
readable syntax, to the all encompassing range of useful
libraries found in Rubygems, to the friendly demeanor of
its creator Matz, who gives off the air of a trustworthy
benevolent dictator if there ever was one.

But Ruby is also a very frustrating. Its easy syntax and
generously loose constraints let you build small programs
with incredible speed, but the longer lived that a project
is, the more those features become liabilities. I've worked
on projects on the scale of hundreds of thousands of lines
for my last two jobs now, and at that size, the language's
problems are very obviously visible in shocking relief.

Ruby's core team works industriously, but major changes
tend to be aimed at solving problems of _computing_;
building a better garbage compiler, improving the way
floating point numbers are handled, or accelerating
performance. These are useful additions, but they don't
address Ruby's major weaknesses which are problems of
_engineering_ that make working in large codebases
difficult. The discrepancy in priorities isn't surprising;
these problems take frequent work in a large Ruby codebase
to become visible, and the majority of Ruby's core and
community are working primarily in C (Ruby's language of
implementation), or working on small or medium-sized
programs.

## I've already heard it's slow (#slow)

Ruby _is_ slow and resource intensive, but although
considerable, it's less of a problem than you'd think.
Production apps spend an inordinate amount of time waiting
on database calls and other I/O, so improvements to the
program's structure and efficiency will generally yield
results that are good enough to stave off an expensive
rewrite in a more performant language. Organizations can
also address capacity by throwing more hardware at the
problem; it's more expensive, but not unreasonable when
compared to the cost of engineering time.

What I want to focus on are problems of organizational
scale; how the language itself starts to break down once
you have a service that's taking a lot of traffic, or one
that's being modified by a lot of engineers.

## Beyond the shell script (#beyond-the-shell-script)

### Zero information pre-runtime (#runtime)

Aside from the most egregious syntax problems, Ruby won't
catch anything until code runs. Bad variable references,
invoking non-existent methods, uninitialized variables, and
type mismatches are all fair game. This is generally fine
for a disposable shell script that's just going to be run
once or twice, but presents a much bigger problem for a
production service.

The language's overpermissiveness forces operators to come
up with all kinds of creative hacks to get slightly better
safety:

* CI suites that try to test code to an absolute exhaustive
  extent, sometimes even aiming for 100% line coverage.
  Every path needs regular testing or a change made at some
  point will cause a regression.
* Heavy reliance on exception tracking services like Sentry
  and Rollbar to tell us what's happening. Errors in
  production aren't an "if"; they're a "when".
* Canary deploys to help detect problems early, and contain
  the fallout in the inevitable case of a bad deploy.

If Ruby code hasn't been run, it doesn't work. Even once
you've fixed it and it does, it's bound to break again at
some point in the future without line coverage that's
complete enough to ensure that every line changed is valid.

!fig src="/assets/ruby-scale/knot.jpg" caption="Without constraints, code becomes a tight knot as modules bleed into each other."

### Boundary bleeding (#boundary-bleeding)

Symbols loaded into a Ruby runtime all end up in one big
pot so that anything that's been loaded at any point is
available to be run from any module. Loading order needs to
be preserved as code is being initially parsed, so this
_doesn't_ work:

``` ruby
class App
  # Error: Util is not yet available.
  include Util
end

module Util
end
```

But once everything is loaded in, cross dependencies within
method bodies are perfectly kosher:

``` ruby
module Util
  def self.hello
    App.hello
  end
end

class App
  include Util

  def self.hello
    puts "hello"
  end
end

# Prints "hello".
Util.hello
```

Most developers are savvy enough not to introduce
pathologically illogical circular dependencies between
modules where it's obvious that one should be the lower
substrate (e.g. `Util`), but given enough code, enough
modules, and enough developers, things will get hazy. A few
encapsulation violations will start to appear and the
interpreter won't complain.

!fig src="/assets/ruby-scale/modularity.svg" caption="Boundary violations grow linearly with the number of lines of Ruby produced."

Soon the violations are everywhere, and module hierarchy
(if there ever was one) becomes indistinct. It's no longer
possible to consider just one module in isolation because
with the exception of the most basic dependencies, almost
every module is tightly intertwined with every other.

### The edit-compile-debug cycle (#edit-compile-debug)

Zeus

Without Zeus, large Ruby codebases would quite simply not
be tenable. Many Ruby advocates don't know this because
you'll never notice the problem with a few hundred lines of
code and a couple dependencies.

### Tooling (#tooling)

Ruby metaprogramming constructs are well known, and though
they may lead to pleasantly readable code, they also make
infamously difficult to figure out what's actually going to
be run. It's not uncommon to be trying to find a method
that's callable from a mixin included by a base class
that's defined in a gem opaquely required by Bundler. Even
once you've located that source package, the definition may
yet be another two gem indirections and six DSL/mixin
layers deep.

The invention of Pry has made this more manageable in that
all of this can be determined at runtime fairly easily, but
the difficulty in statically analyzing Ruby continues to
make it difficult to implement editor "jump to",
auto-completion, and other functions that are invaluable
for developer productivity.

There are options available, but they're entirely based on
heuristics. They might provide some gain in working speed,
but are not accurate or reliable.

!fig src="/assets/ruby-scale/tooling.jpg" caption="Good tooling is sadly lacking/impossible."

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

Ruby does have a few advantages. Pry is the best REPL in
the world, and it helps immensely while debugging or trying
to any kind online introspection.

## Summary (#summary)

I decided to write this not so much in disdain for the
language, but more because I wish other people wrote about
experiences in the real world more often. It's easy to get
the impression from meetups and conferences that languages
and technologies are flawless, but that's almost never the
case.
