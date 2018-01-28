---
title: Programming Language Does Matter
location: San Francisco
published_at: 2018-01-28T18:52:56Z
hook: TODO
---

Years ago, I transitioned from a job of three years where
I'd worked on C# to one where I'd be working with PHP full
time. I'd worked with PHP before and wasn't a fan, but the
opportunity was good, and so I adopted a "it's the poor
carpenter who blames his shoddy tools" attitude to
rationalize it. It didn't matter if the new stack was PHP
-- I'd still be productive.

The catastrophe's proportions were greater than I could've
ever imagined. Some of that was from prevalent bad coding
practices, but a healthy part of it was PHP itself. For the
first time I learnt that it was possible to have large
swaths of code whose behavior _nobody_ understood. `$id` is
usually an integer, but maybe sometimes it's a string.
Hopefully both paths are exercised by a test, but don't
depend on it. Despite attempted hardening, production
mistakes were a daily occurrence. Non-trivial refactoring
was impossible. The contrast between this state of affairs
and C#, where I used to have my IDE change thousands of
lines for me and have the operation be problem-free _every_
time, was sharp. I left after six months.

To this day there's a common trope in the community that
all programming languages are roughly equal. By moving from
one to another, you may ditch problems associated with the
former, but you'll end up with an equal number of new
problems created by the new language. In the end, it's a
wash.

Regardless of how you feel about that, try to take a step
back and consider for a moment just how unlikely it is that
this is true. For the last 50 years of language
development, all of them ended up with the same number of
mistakes? Newer languages didn't learn anything from their
predecessors?

_Of course_ not all programming languages are equal. It's
fair to say that there is no "best" language -- strengths
and weaknesses are subjective so the error bars when trying
to compare them are huge, but in at least some cases when
all the pros and cons are tallied up, some languages excel
over others by wide margins.

## Language creators learn (#creators-learn)

In fact, every new language has had considerable learning
from everything that came before it. As a basic example,
many languages realized from C that it might be a good idea
to include higher-level, safer memory management primitives
after it was obvious that every project was implementing
its own memory allocation abstractions, and the use of
`strcpy` (vs. `strncopy` or the like) on unsafe buffers was
starting to have some serious ramifications where security
was concerned.

### A packaging story (#packaging)

An example I like to use because it's divorced from big
opinions around types and syntax is packaging. Ruby [1] had
a fairly long prehistory that started somewhere around the
publication of Ruby 0.95 in 1995. Language maintainers
realized it would be useful if code could be packaged up,
distributed, and reused, and so RubyGems was born somewhere
around Ruby 1.8 in 2003 [2]. From there, it became obvious
that packages in themselves were not enough. Any serious
project needed to be able to define its dependencies and
the versions of those dependencies so that an installation
could be reproduced, and so Bundler was born. But even that
wasn't the end -- serious developers would maintaining gems
cross-compatible between different versions of Ruby or
upgrading apps from one version to another needed to able
to work across Ruby versions, so we got RVM and later
rbenv.

This progression was perfectly natural -- computing was
still rapidly evolving in the 90s and no one foresaw the
need for this much packaging infrastructure. It all still
exists and is still useful, but with the exception of
RubyGems, all lives as separate tooling outside of Ruby
Core. Any developer new to the language has to learn Ruby,
then go to each of these projects, read their local
documentation, and figure out how to use them as well.
Comparable languages like JavaScript and Python went
through a similar evolutionary process and are left in a
similar state.

Compare that state of affairs to a newer language like
Rust. All of this tooling existed well before the 1.0
release. The distinction between RubyGems and Bundler turns
out to be not all that useful in practice, so in Rust it's
all just in one command called `cargo`. Toolchain
management in the vein of RVM/rbenv is provided by `rustup`
(and it also does so much more). Everything is officially
supported and described in the [Rust book][rustbook]. None
of this was devised in a vacuum -- it was inspired from the
hard lessons of predecessors.

## Consensus on interpreted languages (#interpreted-languages)

## Technology ebbs and flows (#ebbs-and-flows)

[1] I'm going to disproportionately pick on Ruby throughout
this piece -- not because it's that much worse, but because
I use it that much more.
[2] RubyGems would officially become part of Ruby in 1.9.

[rustbook]: https://doc.rust-lang.org/1.5.0/book/installing-rust.html
