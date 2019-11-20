+++
published_at = 2019-11-19T19:53:29Z
title = "Async, Awaited"
+++

![Seagrape](/assets/images/nanoglyphs/004-async-awaited/seagrape@2x.jpg)

The folding keyboard [from last week’s](/nanoglyphs/003-12-factors) been scrapped, and I’m now using a similar mobile setup involving Apple’s Magic Keyboard. Not designed at all to be carried around, but surprisingly mobile. So far it's been a rare holdout in a company with the single overarching value of thinner products at any cost, even if that means eviscerating usability. Its battery life is good. No butterfly keys. The only downside is that I have to keep it in a Ziplock bag when it goes into my backpack. Otherwise, its sheen of the purest white becomes a muddied black faster than you say "reality distortion".

Welcome to Nanoglyph, a software weekly, issue no. four. One point of interest is that I migrated the system that handles this newsletter’s testing and deployment from Travis to GitHub Actions, which was a fun mini-project. See a few of my notes on that at the end of this newsletter.

Today’s photo is from a recent dive trip to Honduras. More on that in soon-to-come edition of [_Passages & Glass_](/newsletter) (my other, less software oriented, much less frequent newsletter) if you’re interested in hearing more.

---

## The beginning at last (#beginning)

Last week, async-await stabilized in Rust. It was a long road to get there. There were some early forays into alternative concurrency models like green threads that were eventually deliberated out. Then came a terrible misadventure in building async-await as user-space abstractions on top of macros. The result was [pure agony](/fragments/rust-brick-walls). Then, perhaps the longest phase: today’s idea was pitched, implemented, and travelled through a lengthy vetting process before getting to where it is today. It’s been almost two years since the [original RFC](https://github.com/rust-lang/rfcs/pull/2394).

Plausible concurrency is such an important aspect for a modern programming language that in a big way, this is really Rust’s beginning. From here, the higher level constructs can come in on top of the async-await primitives, starting with a release of Tokio this month, and culminating with refined (and very fast and very low resource) web frameworks. There are some exciting times for this language ahead. 

## The real cost of dependencies (#dependencies)


Russ Cox writes [about software dependencies](https://queue.acm.org/detail.cfm?id=3344149), a topic that’s infamously relevant to the Go community at large, and even more so him personally as one of the main drivers of the new-ish [Go Module](https://blog.golang.org/using-go-modules) system.

The article’s premise is that although the steady advances in the quality and ease-of-use of package managers is a good thing, their use is now so streamlined that modern developers rarely even hesitate before adding a new one, maybe best exemplified by Node’s `escape-string-regexp` package, an 8-line package depended on by 1,000+ other packages and untold numbers of apps. 

Dependencies are useful for quick features, but can have less obvious long term ramifications in maintenance burden, bugs, and security. A direct example was the injection of malicious code to access the Bitcoin wallets of Copay users into Node’s `event-stream` package. An even larger sensation was the exfiltration of data of 146M Americans from Equifax, which turned out to be due to a vulnerability in Apache Struts.

Cox suggests vetting potential dependencies before using them by closely examining their design, code quality, state of maintenance, transitive dependencies, and security history. Depending on the results, you may want to engage the dependency, engage it more carefully through isolation like a sandbox, or avoiding it. He notes that Apache Struts disclosed major remote code execution vulnerabilities in 2016, 2017, and 2018 — it might be best to avoid the use of projects with such patchy histories altogether.

One point of interest is that he mentioned that Go is experimenting with including dependency version manifests into compiled binaries, the purpose of which would be to allow deployment and security tools to scan them for potential liabilities that should spur an upgrade. This strikes me as a very good idea given that the opaqueness of an already-compiled binary is one of their few downsides compared to alternative approaches.

## Static trees (#static-trees)

One of the most exciting software developments in Apple’s world recently was [SwiftUI](https://developer.apple.com/xcode/swiftui/), their take on a React-like system for building modular UIs, but allowing you to try write them in a nice modern language like Swift instead of untyped JSX templates. By most reports it's [pretty rough right now](https://inessential.com/2019/10/21/swiftui_is_still_the_future), but it looks promising.

[Static types in SwiftUI](https://www.objc.io/blog/2019/11/05/static-types-in-swiftui/) describes how Swift infers complex types based on views built in the framework. So something like this:

``` swift
let stack = VStack {
    Text("Hello")
    Rectangle()
        .fill(myState ? Color.red : Color.green)
        .frame(width: 100, height: 100)
}
```

Translates to the nested, parameterized type `VStack<TupleView<(Text, ModifiedContent<_ShapeView<Rectangle, Color>, _FrameLayout>)>>`.

These React-like frameworks re-render view trees as the state they’re encapsulating changes. In React or Elm, re-renders typically involve a diff step, where the old and new view trees are compared so that the framework can determine if new elements need to be added or old ones removed. Having a static type to describe each view will often allow SwiftUI to skip or minimize the work done during this step — it knows in advance that the elements can remain constant, and it’s just their internal state that needs to be updated.

A more complex view include a conditional that could make the typing system more complex:

```
if myState {
	Text(“Hello”)
}
```

SwiftUI handles this case by replacing a previously required `Text` field with an optional `Text?` in the static type, so it can still benefit from the typing by only diffing in places like this where it’s possible for elements to change. It also has a scheme for handling conditional branches like with `if ... else ...` (they become `_ConditionalContent`) and in cases where the logic becomes too complex to encode in types, can fall back to a type-erased `AnyView` to represent anything.

## Migrating to Actions (#github-actions)

Ever since Travis untimely [acquisition by a holding company](https://news.ycombinator.com/item?id=18978251) and the departure of a large number of its engineering staff, a number of us have been keeping an eye out for what to use next. For the time being Travis is still running, but the clock’s likely inching closer to midnight.

[GitHub Actions](https://github.com/features/actions) were an extremely timely addition to the company’s product surface. Although described in most places in such grandiose vernacular so as to obfuscate what it actually does, to put it simply, it allows you describe jobs that will run on certain repository events like a push, opened pull request, or cron — perfect for CI. It’s major differentiating feature is that the steps of a job can be defined as similar shell commands, they can also defined a Docker container to run. This makes the whole system modular — units of work of arbitrary size can live in their own project, providing good encapsulation and easy updates.

