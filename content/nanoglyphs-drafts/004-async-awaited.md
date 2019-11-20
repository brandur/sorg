+++
published_at = 2019-11-19T19:53:29Z
title = "Async, Awaited"
+++

![Seagrape](/assets/images/nanoglyphs/004-async-awaited/seagrape@2x.jpg)

The folding keyboard [from last week’s](/nanoglyphs/003-12-factors) been scrapped, and I’m now using a similar mobile setup involving Apple’s Magic Keyboard. Not designed at all to be carried around, but surprisingly mobile. So far it's been a rare holdout in a company with the single overarching value of thinner products at any cost, even if that means eviscerating usability. Its battery life is good. No butterfly keys. The only downside is that I have to keep it in a Ziplock bag when it goes into my backpack. Otherwise, its sheen of the purest white becomes a muddied black faster than you say "reality distortion".

Welcome to Nanoglyph, a software weekly, issue no. four. One point of interest is that I migrated the system that handles this newsletter’s testing and deployment from Travis to GitHub Actions, which was a fun mini-project. See a few of my notes on that at the end of this newsletter.

Today’s photo is from a recent dive trip to Honduras. More on that in soon-to-come edition of [_Passages & Glass_](/newsletter) (my other, less software oriented, much less frequent newsletter) if you’re interested in hearing more.

---

## The beginning at long last (#beginning)

Last week, async-await stabilized in Rust. It was a long road to get there. There were some early forays into alternative concurrency models like green threads that were eventually deliberated out. Then came a terrible misadventure in building async-await as user-space abstractions on top of macros. The result was [pure agony](/fragments/rust-brick-walls). Then, perhaps the longest phase: today’s idea was pitched, implemented, and travelled through a lengthy vetting process before getting to where it is today. It’s been almost two years since the [original RFC](https://github.com/rust-lang/rfcs/pull/2394).

Plausible concurrency is such an important aspect for a modern programming language that in a big way, this is really Rust’s beginning. From here, the higher level constructs can come in on top of the async-await primitives, starting with a release of Tokio this month, and culminating with refined (and very fast and very low resource) web frameworks. There are some exciting times for this language ahead. 

## The real cost of dependencies (#dependencies)


Russ Cox writes [about software dependencies](https://queue.acm.org/detail.cfm?id=3344149), a topic that’s infamously relevant to the Go community at large, and even more so him personally as one of the main drivers of the new-ish [Go Module](https://blog.golang.org/using-go-modules) system.

The article’s premise is that although the steady advances in the quality and ease-of-use of package managers is a good thing, their use is now so streamlined that modern developers rarely even hesitate before adding a new one, maybe best exemplified by Node’s `escape-string-regexp` package, an 8-line package depended on by 1,000+ other packages and untold numbers of apps. 

Dependencies are useful for quick features, but can have less obvious long term ramifications in maintenance burden, bugs, and security. A direct example was the injection of malicious code to access the Bitcoin wallets of Copay users into Node’s `event-stream` package. An even larger sensation was the exfiltration of data of 146M Americans from Equifax, which turned out to be due to a vulnerability in Apache Struts.

Cox suggests vetting potential dependencies before using them by closely examining their design, code quality, state of maintenance, transitive dependencies, and security history. Depending on the results, you may want to engage the dependency, engage it more carefully through isolation like a sandbox, or avoiding it. He notes that Apache Struts disclosed major remote code execution vulnerabilities in 2016, 2017, and 2018 — it might be best to avoid the use of projects with such patchy histories altogether.

One point of interest is that he mentioned that Go is experimenting with including dependency version manifests into compiled binaries, the purpose of which would be to allow deployment and security tools to scan them for potential liabilities that should spur an upgrade. This strikes me as a very good idea given that the opaqueness of an already-compiled binary is one of their few downsides compared to alternative approaches.

## Static trees (#static-trees)

A view tree is re-rendered as state changes, but the types within that tree are invariant.

In similar architectures like React or Elm, the tree model is common, but there’s a step during a re-render where the old and new trees need to be compared to check if any new elements were added or old ones removed.

SwiftUI’s static types guarantee that the elements between renders are the same, and it’s just the state that may need to be modified.

There are more complex cases like where an `if` conditional is introduced in view rendering logic:

``` swift
if myState {
	Text(“Hello”)
}
```

SwiftUI can handle this case as well by replacing a previously required `Text` field with an optional `Text?` in the final type. It also has scheme for handling conditional branches like with `if ... else ...` and in cases where the logic becomes too complex to encode in types, can fall back to a type-erased `AnyView` to represent anything.

## GitHub Actions (#github-actions)
