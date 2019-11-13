+++
published_at = 2019-11-13T02:21:52Z
title = "Async, Awaited"
+++

![Seagrape](/assets/images/nanoglyphs/004-async-awaited/seagrape@2x.jpg)

The folding keyboard from last week’s been scrapped, and I’m now using a similar mobile setup involving Apple’s Magic Keyboard. Quite mobile, and yet curiously it’s managed to avoid having its key mechanism downgraded to butterfly by an overzealous product manager inside Apple. The only downside is that I have to keep it inside a Ziplock bag because otherwise its sheen of the purest white turns black so quickly.

Welcome to Nanoglyph, a software weekly, issue number four.

Today’s photo is from a recent dive trip to Honduras. More on that in soon-to-come edition of [_Passages & Glass_](https://brandur.org/newsletter) (my other, less software oriented newsletter) if you’re interested in hearing more.

---

## The beginning, finally (#async-await)

Last week, async-await stabilized in Rust. It was a long road to get there. There were some early forays into alternative concurrency techniques like green threads that were eventually deliberated out. Then came a misadventure in building async-await as user-space abstractions with macros. The result was [pure pain]. Then perhaps the longest phase: today’s idea was pitched, developed, and travelled through a lengthy vetting process before getting to where it is today. It’s been X years since the original blog post. 

Plausible concurrency is such an important aspect for a modern programming language that in a big way, this is really Rust’s beginning. From here, the higher level constructs can come in on top of the async-await primitives, starting with a release of Tokio this month, and culminating with refined (and very fast and very low resource) web frameworks. There are some exciting times for this language ahead. 

## The real cost of dependencies (#dependencies)

## Static types in SwiftUI (#swiftui-types)
