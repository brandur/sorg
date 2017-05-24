---
title: Architectural Honesty
published_at: 2017-05-24T16:33:53Z
hook: Talking about bad technology and bad architecture
  more often.
---

In our industry, we have a bad tendency to hype technology
up, but never knock it down.

Over the last few decades we've seen big pushes for tools
like XML, PHP, Erlang, Node, Rails, Mongo, Riak, Go, and
countless others. More recently, the industry's seen enough
saturation that fatigue has lead to a bit more skepticism,
but we can still see a lot of ongoing pushes, even if
somewhat more moderated (e.g. Rust, Elixir, Crystal, ...).

Some of these have withstood the test of time and proven
themselves to have enough merit that they're still fine
choices for a new technology stack. On the other hand, a
good number of them are _not_ fine choices for a new
technology stack. Sometimes we make mistakes, or technology
improves, and ideas that we originally thought to be good
hit the end of their useful lifetime.

Edit: Dangerous lack of runtime safety, few constraints to help code scale, or a decaying ecosystem are all good reasons to avoid certain technology.

It's a very human thing to do to withhold criticism. Real
people aren't only emotionally invested in technology they
use, but in many cases their livelihoods depend on it; even
honest criticism could hurt them. We have to consider the
other side though -- by being disingenuous or withholding
information on bad technology, we're cheating people and
companies who aren't using them yet, but may yet adopt
them. By speaking out, you have the potential to save
_millions_ of hours of lost future productivity.

Disastrous pitfalls, vampiric operational overhead, and
chronic underdesign are never in the documentation, and
often not obvious until you're already waist deep. By being
on the inside of these things, you have access to special
insight that other people can't get without fully investing
themselves at great expense.

This isn't to say that we should unduly sling mud, but
pieces that are honest, detail-oriented, thoroughly
researched, but also critical, might be the best way that
you can help your fellow builders.
