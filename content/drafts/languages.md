---
title: Your Programming Language Does Matter
location: San Francisco
published_at: 2018-01-28T18:52:56Z
hook: TODO
---

Years ago, I transitioned from a job of three years where
I'd worked on C# to one where I'd be working with PHP full
time. Although PHP was the first language that I was ever
really productive in, I'd since become disillusioned of it.
But nevertheless, the new job was exciting, and I
rationalized it with an attitude that it's the poor
craftsman who blames his shoddy tools. It didn't matter if
the new stack was PHP, I'd still be productive.

It was a catastrophe. The codebase had all the classic
signs of mismanagement: no cohesive organization, nothing
beyond trivial testing, terrible readability and
modularity, and homegrown frameworks for everything.

All of that was attributable to poor development practices,
but there was an entirely different class of problems that
could only be attributed to PHP itself. I learnt that it
was possible to have large swaths of code whose behavior
_nobody_ understood. One variable `$id` in one part of the
code might usually an integer, but could also sometimes
it's a string. Hopefully both branches were exercised by
tests, but you couldn't depend on it. Without type
annotations in variable declarations or type signatures,
even the most isolated sections of code were difficult to
reason about. Production mistakes (and reverts) were a
daily occurrence. Non-trivial refactoring was impossible.
The contrast between this and C#, where I used to have my
IDE change thousands of lines for me and have the result be
problem-free every time, was sharp.

I don't need to convince most people that PHP isn't a good
language, but only because it occupies a rare place of
widely held abhorrence; many others are its spiritual
siblings, but nowhere near as broadly disliked. There's a
common trope in the industry that programming languages are
all more-or-less equal. By moving from one to another, you
may ditch problems associated with the former, but you'll
end up with an equal number of new problems created by the
new language. In the end, it's a wash.

It's not true. A basic counterargument is one of
probability -- consider how unlikely it is that given 50
years of language development, all of them ended up with
about the same number of mistakes and same number of
strengths. Newer languages didn't learn anything from their
predecessors?

_Of course_ programming languages aren't all a set of equal
tradeoffs. There may be no "best" languages -- strengths
and weaknesses are subjective so the error bars when trying
to compare them are huge, but when all things are
considered, some excel over others by wide margins.

## Language creators learn (#creators-learn)

Every new language learned considerably from everything
that came before it.

As an early and common example, many languages realized
from C that it might be a good idea to include
higher-level, safer memory management primitives after it
was obvious that every project was implementing its own
memory allocation abstractions, and the use of `strcpy`
(vs. `strncopy` or the like) on unsafe buffers was starting
to have some serious ramifications where security was
concerned. Today, _every_ new language provides high-level
abstractions for memory management, even if those
abstractions can vary widely.

### A packaging lesson learned over decades (#packaging)

Packing systems make a more nuanced example. Ruby [1] had a
fairly long prehistory that started somewhere around the
publication of Ruby 0.95 in 1995. Language maintainers
realized it would be useful if code could be packaged up,
distributed, and reused, and so RubyGems was born somewhere
around Ruby 1.8 in 2003 [2].

But _just_ packages were still not enough. Projects needed
to have a way to define a precise and stable set of
dependencies for repeatable installation and deployment,
and so Bundler was born. But even that wasn't the end --
serious developers would maintaining gems cross-compatible
between different versions of Ruby or upgrading apps from
one version to another needed to able to work across whole
Ruby versions. RVM and later Rbenv were created to solve
this problem by allowing multiple versions of Ruby to be
installed and managed.

This progression was perfectly natural -- computing was
still rapidly evolving in the 90s and no one foresaw the
need for this much packaging infrastructure. It all still
exists and is still useful, but with the exception of
RubyGems, all lives as separate tooling outside of Ruby
Core. Any developer new to the language has to learn Ruby,
then go to each of these separate projects, read their
documentation, and figure out how to use them. Many
comparable languages (e.g., JavaScript and Python) went
through a similar evolutionary process and are left in a
similar state.

Compare that state of affairs to a newer language like Rust
[3] where equivalent tooling existed well before the 1.0
release. The distinction between RubyGems and Bundler turns
out to be not useful in practice, so in Rust it's all just
in one command called `cargo`. Toolchain management in the
vein of RVM/Rbenv is provided by `rustup` (which also does
much more). All of these tools are officially supported,
come with the default distributions, and are described in
the [Rust book][rustbook]. None of this was devised in a
vacuum -- it was inspired from the hard lessons of
predecessors like Ruby.

## The melancholy of interpreted languages (#interpreted-languages)

So that brings us interpreted languages, which are still
huge, and show no signs of slowing. I'm going to
generalize by putting a lot of languages like PHP, Python,
Ruby, Perl, and JavaScript in one bucket. They're
different, but they share important properties like dynamic
typing, poor enforcement of modularity, poor parallelism
(at least historically), and few checks prior to runtime.

When it comes to whether interpreted languages are
desirable, there's still no wide consensus. Although many
people feel that they're unsafe, lots of people still like
them because they're easier to learn, and development with
them feels faster (at the beginning of a project anyway),
which leaves many of them as plausible options for industry
newcomers.

A big reason for the contention may be that interpreted
languages really are better for small projects. A developer
proficient with both Ruby and Go can write a small program
in Ruby about 10x faster because there's less boilerplate
and syntax. As long as you can fit the entire program's
context in your head, refactoring is faster too.

### Scaling code (#scaling-code)

Interpreted languages don't have a problem with
productivity or syntax, they have a problem with *scaled
engineering*. As the lines of code you have increase, you
eventually hit a point where your progress is steeply
diminished because you can't be sure of anything anymore.
Code paths become increasingly poorly understood, and
unforeseen side effects become more common as modules are
ever more deeply intertwined.

Here are a few properties of interpreted languages that
make them tend to scale poorly:

* **Poor safety:** With no compiler to provide guarantees,
  checks are reduced, and those that are left are pushed to
  runtime. An exhaustive test suite is needed just to make
  sure that code will run, and although the tests will be
  useful for vetting other behavior, they'll never provide
  the same level of safety as a compiler.
* **Poor modularity:** The tools available for enforcing
  strong boundaries between modules are generally
  non-existent, which make them difficult to scale in a
  large organization where a lot of developers are working
  with the same large body of code, many of whom are
  looking the easiest way to do things.

    Ruby provides `private` and `protected` keywords that
    are trivially bypassed. Python doesn't even provide
    those. Neither provide a way of fully encapsulating a
    type like an internal class.
* **Poor readability:** Skipping type annotations on
  variable declarations and function signatures makes code
  faster to write and allows it to be more dynamic, but it
  makes it much more difficult to reason about later. In
  any mature software project you'll be spending more time
  reading code than writing it, and it's much better to
  optimize for the former. Type annotations make it easier
  to understand code more quickly and make changes safer.

### The jury's still out (#jury)

I once upgraded a gem to a new minor version and caused a
production incident in a different project that shared the
codebase. It turned out that a different gem they'd been
using worked by re-opening the first gem and rewriting some
of its private implementation. The upgrade to the first gem
didn't change its API (thus the minor version), but it did
change its internals, and the result was that the second
gem no longer did what it was supposed to, and in a
particularly nefarious way that was difficult to detect in
tests.

At first, I blamed the silly gem that rewrote another gem's
private implementation. What kind of irresponsible idiot
could write something like that? And indeed, its authors do
deserve some of the blame.

Next I blamed us. We should have had more comprehensive
end-to-end testing that could catch even subtle problems
like this one. We should have vetted incoming gems more
carefully by reading their source code and weeding out any
that were implemented too dangerously. We should have a
safer deployment strategy. And indeed, we deserve some of
the blame.

But lastly, I realized that the lion's share of the blame
lies with neither the gem's authors nor with us, but with
the Ruby language itself. Why is one gem allowed to reopen
the implementation of another and overwrite it? This is a
tradeoff of a little additional convenience for an
unspeakably large loss in long term reliability. By
providing better and safer guide rails, programming
languages are in a place to provide unparalleled additional
resilience to programs implemented in them. Interpreted
languages have made an explicit decision not to, and even
to move in the opposite direction.

A lot of experienced developers do eventually come to the
realization that they'd rather avoid interpreted languages
if possible, and especially recently, have started to opt
for languages with compilers and types. And yet, the images
of many interpreted languages continues to stay strong, and
they're still used for many (even most) new projects. The
jury's still out, even if it shouldn't be.

#### The evasiveness of consensus (#consensus)

Here are a few of my favorite explanations for why that is:

1. Many of the problems that interpreted languages cause
   don't appear until late in a project's lifecycle. Many
   developers never work with projects large enough to get
   there.
2. You need some experience with other languages to know
   that the problems that exist in interpreted languages
   aren't fundamental to programming. Many people never
   investigate very widely outside of what they're
   currently working on.
3. There's a widespread reluctance in attributing problems
   that occur to being flaws in the parent programming
   language. In my example of the gem upgrade above, many
   would stop at blaming themselves or the project owner,
   but would never think to trace the problem back to Ruby.
   We need to make sure we're being fully honest when
   examining root causes.

## Technology ebbs and flows (#ebbs-and-flows)

Because computing as a discipline is older than most of us,
it's easy to forget that it's still very new. We're still
very much in its prehistory -- there's still a lot to
learn, and nothing is set in stone.

I like to look at the history of programming languages up
until now in a few major phases (please note this is far,
far from exhaustive):

1. In the beginning there was a long period of early
   development that result in a Cambrian explosion of
   ideas. Many languages were developed during this period,
   but most of them haven't survived as viable alternatives
   to the present day. Some well-known examples are COBOL
   (1959) and BASIC (1964).
2. Those ideas were refined into a wave of older languages
   that are still in use today: C (1972), SQL (1978), C++
   (1980), Common Lisp (1984), Erlang (1986), and Perl and
   (1987).
3. After Perl, there was a major bloom in new interpreted
   languages. It was followed by Python (1991), JavaScript
   (1995), PHP (1995), Ruby (1995), and a host of others.
4. Alongside the explosive popularity of interpreted
   languages, other branches of advancement continued. An
   important one was the paradigm that followed in the
   footsteps of C++ with an object-oriented design and
   strong types and led to Java (1995), and C# (2000).
5. Many of the ideas from the last few decades were remixed
   (and some rejected) into what I'd class as a "modern"
   wave of languages: Go (2009), Ruby (2010), and Swift
   (2014).

### Themes for the future (#themes)

While there's still a huge amount of diversity between
implementations, I'm quite comforted that many of these
more recent innovations share some common themes.

1. **Types, static typing, and a compiler:** Types tend to be
   strong, static (non-dynamic), and caught at compile
   time.
2. **Runtime speed:** It turns out that we don't need to
   trade development productivity for runtime speed as long
   as we build and use better tools. The LLVM was a huge
   advancement here in that it allowed a powerful VM-like
   abstraction layer, but continued to provide incredible
   performance.
3. **Concurrency:** Concurrency/parallelism models that work,
   and which are baked into the core of the language.
   For example, Goroutines in Go, or the
   async/await/futures model found in Rust.
4. **The good parts of object-orientation:** Structs with
   member functions are a good idea. Don't provide too much
   power like C++ or Java did so as to avoid inheritance
   hell and fragility.
5. **Tooling:** A "batteries included" standard library was
   a good idea, but don't stop there. Also provide tooling
   for use with the language like build tools, dependency
   managers, documentation generators, version managers,
   etc.
6. **Explicitness:** Type interpolation is fine, but limit
   it. Functions have type signatures to keep them explicit
   and easy to reason about at the minor cost of a few
   extra keystrokes.
7. **Modularity:** Module/packaging systems provide strong
   encapsulation, and can protect the entirety of their
   private APIs including classes and submodules.

We've made some missteps in the past, but we should think
of them as learning experiences towards a better future. A
programming language that can provide the right combination
of powerful primitives, abstractions, and tooling can save
millions of hours for its users as they avoid inventing
them themselves, and working for years towards establishing
conventions that are informal but recommended.

[1] I'm going to disproportionately pick on Ruby throughout
this piece -- not because it's that much worse, but because
I use it that much more.

[2] RubyGems would officially become part of Ruby in 1.9.

[3] Rust isn't the only language that does packaging well,
but they're a good example of one which I'll use for the
purpose of this article.

[rustbook]: https://doc.rust-lang.org/1.5.0/book/installing-rust.html
