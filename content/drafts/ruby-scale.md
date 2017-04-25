---
title: Ruby, and Problems of Scale
published_at: 2017-04-18T14:23:28Z
location: San Francisco
hook: The challenges of scaling and operating a big Ruby
  codebase (that are not related to performance).
hook_image: true
---

Ruby is a beautiful language. Speaking from experience,
it's easy to get attached to everything from its pleasantly
readable syntax to the friendly demeanor of its creator
Matz, who gives off the air of a trustworthy benevolent
dictator if there ever was one.

TODO

I have a love/hate relationship with Ruby. I love how
expressive the language is and how it lets me write
programs (especially small ones) fast, but I'm also
consistently frustrated by how poor its development
experience is in certain areas.

I know a fair number of Ruby evangelists from my time at
Heroku who in most cases consider Ruby to be a nearly
perfect language. The feeling isn't so much the result of
willful ignorance, but more because Ruby's cracks don't
appear until you're managing a sizable codebase. If the
largest Ruby programs you write are only a few thousand
lines, then you might never run into anything that I talk
about here.

My last two jobs have had me work in huge Ruby codebases on
the order of hundreds of thousands of lines and dozens (if
not hundreds) of gem dependencies. It's at this scale that
all the language's problems come into sharp relief.

## I've already heard it's slow (#slow)

Ruby _is_ slow and resource intensive, but this is less of
a problem than you'd think. Many production apps spend an
inordinate amount of time waiting on database calls and
other I/O, so improvements to the program's structure and
efficiency will yield better result than rewriting in a new
language. Capacity can also often be addressed by throwing
more hardware at the problem; it's more expensive, but not
unreasonable when compared to the cost of engineering time.

What I want to focus on are problems of organizational
scale; how the language itself starts to break down once
you have a service that's taking a lot of traffic, or one
that's being modified by a lot of engineers.

## Beyond the shell script (#beyond-the-shell-script)

### Zero information pre-runtime (#runtime)

Aside from the most egregious syntax problems, Ruby won't
catch anything until code actually runs. Bad variable
references, invoking non-existent methods, uninitialized
variables, and type mismatches are all fair game. This is
generally fine for a disposable shell script that's just
going to be run once or twice, but presents a much bigger
problem for a production service.

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

If Ruby code hasn't been run, it almost certainly doesn't
work. Even if it's been run and everything works today,
without total test coverage something is bound to break
tomorrow.

### Boundary bleeding (#boundary-bleeding)

Symbols loaded into a Ruby runtime all end up in one big
pot so that anything that's been loaded at any point is
available to be run from any module. Loading order needs to
be preserved as code is being initially parsed, so this
_doesn't_ work:

``` ruby
class Foo
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
    Foo.hello
  end
end

class Foo
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

Eventually the violations are everywhere, and module
hierarchy (if there ever was one) becomes indistinct. It's
no longer possible to consider just one module in isolation
because with the exception of the most primitive
dependencies, almost every module is tightly intertwined
with every other.

!fig src="/assets/ruby-scale/knot.jpg" caption="Without constraints, code becomes a tight knot as modules bleed into each other."

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

The invention of Pry has made this more managable in that
all of this can be determined at runtime fairly easily, but
the difficulty in statically analyzing Ruby continues to
make it difficult to implement editor "jump to",
auto-completion, and other functions that are invaluable
for developer productivity.

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
