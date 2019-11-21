+++
published_at = 2019-11-19T19:53:29Z
title = "Async, Awaited"
+++

![Seagrape](/assets/images/nanoglyphs/004-async-awaited/seagrape@2x.jpg)

The folding keyboard [from last week](/nanoglyphs/003-12-factors) been scrapped, but I'm still experimenting with writing on mobile platforms, so I’m now using a similar setup with one of Apple’s Magic Keyboards. These weren't designed with mobility in mind, but are surprisingly good at it: compact, good battery life, no butterfly keys. The downside is that they need to be stored in a Ziplock bag lest their veneer of pure white become a muddied black faster than you say One Infinite Loop.

Welcome to Nanoglyph, a software weekly, issue four. It's still an in-progress experiment in disciplined writing -- to follow along, you can [subscribe here](https://nanoglyph-signup.herokuapp.com).

Today’s photo is from a recent dive trip to Honduras. More on that in soon-to-come edition of [_Passages & Glass_](/newsletter) (my other, less software oriented, much less frequent newsletter) if you’re interested in hearing more.

---

## At last, the beginning (#beginning)

Last week, async-await stabilized in Rust. It was a long road to get there. It started early exploration into a variety of concurrency models, including Go-like green threads that were eventually deliberated out in preference of an ultra-minimal runtime. Then came a terrible misadventure in using macros to build async-await as user-space abstractions. The result was [pure agony](/fragments/rust-brick-walls). Then, perhaps the longest phase: today’s idea was proposed, implemented, and went through a lengthy vetting process before arriving at where it is today. It’s been almost two years since the [original RFC](https://github.com/rust-lang/rfcs/pull/2394).

These days, with so many of us building network-based programs like APIs and web sites, plausible concurrency is a hallmark of any modern programming language. With its asynchronous mechanics finally finished, in a very practical sense this is Rust's true beginning. Higher level constructs will come next, starting with a release of Tokio this month, and culminating with fast, efficiency, and refined web frameworks. Exciting times lie ahead.

## The real cost of dependencies (#dependencies)

Russ Cox writes [about software dependencies](https://queue.acm.org/detail.cfm?id=3344149), a topic that’s infamously relevant to Go, and even more so him personally as one of the drivers of the new-ish [Go Module](https://blog.golang.org/using-go-modules) system.

His premise is that although the steady advances in the quality and ease-of-use of package managers is a good thing, their use is now so streamlined that modern developers rarely even hesitate before adding new dependencies. This is exemplified well by Node’s `escape-string-regexp` package, an 8-line package in use by 1,000+ other Node packages and an untold numbers of apps.

These external packages are tremendously useful for building features quickly, but have less obvious long term ramifications in maintenance, bugs, and security. A recent example was the introduction of malicious code into Node's `event-stream` package to steal from the Bitcoin wallets of Copay users. An even greater sensation was the exfiltration of data of 146M Americans from Equifax, which turned out to be due to a vulnerability in Apache Struts.

Cox suggests vetting dependencies before use by examining their design, code quality, state of maintenance, transitive dependencies, and security history. Depending on the results, you may decide to pick up the dependency, engage it more carefully through an isolation mechanism like a sandbox, or avoiding it. He notes that Apache Struts disclosed major remote code execution vulnerabilities in 2016, 2017, and 2018 -- projects with such sordid histories might be best avoided altogether.

A point of interest is that Go is experimenting with including dependency version manifests in compiled binaries, which would allow tooling to scan them for liabilities that should prompt an upgrade. It's a great idea given that the opaqueness of an already-compiled binary is one of the few downsides of deployment via static binary compared to alternatives.

## Static trees (#static-trees)

An exciting development in Apple's toolchain was the introduction of [SwiftUI](https://developer.apple.com/xcode/swiftui/), their take on a React-like system for building declarative, modular UIs, but unlike React, allowing you to do so in a nice modern language like Swift instead of untyped JSX templates. By most reports using it is [rough going](https://inessential.com/2019/10/21/swiftui_is_still_the_future) right now, but it looks promising.

[Static types in SwiftUI](https://www.objc.io/blog/2019/11/05/static-types-in-swiftui/) describes how Swift infers complex types based on views built with the framework. So something like:

``` swift
let stack = VStack {
    Text("Hello")
    Rectangle()
        .fill(myState ? Color.red : Color.green)
        .frame(width: 100, height: 100)
}
```

Is interpreted by the compiler as the nested, parameterized type `VStack<TupleView<(Text, ModifiedContent<_ShapeView<Rectangle, Color>, _FrameLayout>)>>`.

React-like frameworks re-render view trees as state they’re encapsulating changes. In React or Elm, re-renders typically involve a diff step, where old and new trees are compared so that the framework can determine if new elements need to be added or old ones removed. Having a static type to describe each view will often allow SwiftUI to minimize that work or even skip it altogether — the type system tells it in advance that elements remain constant, and it’s just their internal state that needs updating.

A more complex view might include a conditional that makes that proposition more complex:

```
if myState {
	Text(“Hello”)
}
```

SwiftUI handles this case by replacing a required `Text` field with an optional `Text?` in the static type. Some tree diffing is now necessary, but only a minimal amount where conditional elements are present. Full `if ... else ...` constructs are also supported by typing (they become `_ConditionalContent<A, B>`) and in cases where conditional logic becomes too complex to encode, the type-erased `AnyView` acts as a fall back to represent anything.

---

Goodbye.
