+++
published_at = 2019-12-26T23:19:29Z
title = "Preventing the Collapse of Civilization"
+++

![Palace of the Fine Arts in 2015](/assets/images/nanoglyphs/007-civilization/palace@2x.jpg)

Happy holidays! From my vantage point, they get a little closer together every year, but that existential malaise is mitigated by how much I look forward to them. It's already been an amazing few days of rich food and decadent desserts, and I hope everyone's break is going well.

You're reading 2019's final edition of _Nanoglyph_, an experimental newsletter on software. If you're viewing it on the web, as usual you can subscribe [here](/newsletter) to get it in the new decade. Personally, I can't wait to talk about how I lived through the 20s.

---

We recently held an annual internal conference, with out of town employees converging on San Francisco for the company's largest gathering all year. As it was this year, the first of these I went to years ago was held at the Palace of the Fine Arts, an event space in the Marina. The venue's interior is a one giant room of contiguous space. That first year we occupied only a tiny corner of it, with the rest consisting of nothing but echoing emptiness. This year, not only had the whole building been repurposed for our occupancy, but the event space felt positively crowded -- a bewildering turnaround to say the least.

Much of the event was the troop rallying you'd expect of a corporate conference, but the organizers secured some very interesting external talks including one from Jonathan Blow, best known as the creator of _Braid_ and _The Witness_. Beyond games, Jonathan’s well known to have strong opinions on software development, and has gone so far as to work his own programming language oriented towards performance and loading data (an important consideration for games) -- the yet unreleased JAI.

No slides were posted for this talk, but much of the same philosophy is conveyed in the excellent [_Preventing the Collapse of Civilization_](https://www.youtube.com/watch?v=pW-SOdj4Kkk). To sum it up in as few words as possible (and do it some injustice by doing so): programmers are taught early to build layers of abstraction, with the premise being that each new layer brings more safety and more productivity.

Jonathan’s contention is that the model’s core assumption -- that abstraction is free -- is wrong. Layers have weight, and moving up an ever deepening stack compromises robustness and productivity as complicated increases. We've all internalized this truth without knowing it -- no one really expects software to work properly, and we're never surprised by the minor bugs we encounter in almost every program of web app we use on a daily basis.

I’ve previously called _Preventing the Collapse of Civilization_ the most important talk for the future of the field of software engineering, period. We can do better than the crystalline constructs of dependencies, unwieldy frameworks, and sluggish build/transpilation pipelines that many of us are used to using today. The emergence of newer generations of tooling like Go and React and steps in the right direction, but even there they probably don't sufficiently compact the existing stack, and we'll need to do better. Admittedly, I'm predisposed to liking Jonathan's ideas, but the talk is worth an hour of anyone's time.

---

## A scalability epic (#odyssey)

Yandex released [version 1.0 of Odyssey](https://github.com/yandex/odyssey/releases/tag/1.0), a fast connection pooler for Postgres. Unlike PgBouncer which is single-threaded, it’s built with a multi-threaded architecture, and the various improvements that come in with 1.0 put its feature set on par with any comparable projects. The release notes indicate that Yandex is using Odyssey to run more than 1,000,000 requests per second across many hundreds of hosts.

Yandex has previously talked about migration their mail service [from Oracle to Postgres](https://www.pgcon.org/2016/schedule/attachments/426_2016.05.19%20Yandex.Mail%20success%20story.pdf) (PDF warning), and it suggests that they were managing 300 TB of core data in 2016, which makes it fair to call them one of the largest Postgres users in the wild. They also provide a managed Postgres service, so if any user is likely to have developed deep operational dexterity with Postgres, it’s them.

## Adaptive runtimes (#adaptive-runtimes)

The async-std project in Rust writes about their new [Go-inspired runtime](https://async.rs/blog/stop-worrying-about-blocking-the-new-async-std-runtime/). The current state of asynchronous Rust is a little convoluted, so a quick refresher for anyone not living on the razor’s edge:

* The `async`/`await` language keywords have landed in stable Rust.
* Although these basic primitives are available in core, Rust notably does not include an asynchronous runtime. Users bring their own, the [Tokio runtime](https://docs.rs/tokio/0.2.6/tokio/runtime/index.html) being the most common.
* The async-std project aims to provide async equivalents for everything found in Rust’s standard library. For example, reading a file is a blocking operation, but async-std provides an async alternative. The project also includes a runtime that can be used as an alternative to Tokio.

A traditional problem with the async model (compared to threads or green threads) is _blocking_. Runtimes typically rely on functions finishing quickly (or pausing on other asynchronous work) so that the runtime can move on to something else. Node’s runtime for example is infamously single-threaded, so it can only be doing one thing at a time. A function that blocks through either a compute-heavy workload, or by waiting on I/O in a way that’s not compatible with its reactor, gums up the whole system.

Th async-std executor is interesting because it’s designed to detect when a single task has been blocking for too long and respond by spinning up a new executor thread to replace the old one. It starts single-threaded, but can adaptively scale itself to a multi-threaded model as its workload demands. That sort of implicit behavior is likely to lead to resource usage problems for some production stacks, but has the distinct advantage of being quite practical.

## Most cloud per buck (#cloud-report)

Cockroach Labs published their [2020 Cloud Report](https://www.cockroachlabs.com/blog/2020-cloud-report/), which compares performance across major cloud providers. They’ve published the report for a few years now, and the most shocking result is how effectively Google’s closed to gap between itself and Amazon over the last year. AWS outperformed GCP by 40% in the last report, which they attributed to Amazon’s [Nitro System](https://aws.amazon.com/ec2/nitro/) present in newer instance types. This year, that edge has vanished completely.

The benchmarks using the realistic [TPC-C](http://www.tpc.org/tpcc/) suite, which measures database OLTP performance by modeling warehouses executing transactions. In terms of throughput, AWS’ `c5d.4xlarge` came out on top, but only by a narrow margin relative to comparable products from Azure and GCP. More interestingly, the report doesn’t stop there -- it goes on to also measure the performance of each cloud _relative to cost_, and along that dimension GCP’s `n2-highcpu-16` tops the list.

Again, the differences are minor enough to the point where they’re not all that important, but the big takeaway here should be that for a realistic workload, there's not much difference between the big three anymore.

---

I'm back in Calgary for the break, and nature's delivered us a picturesque winter -- lots of snow on the ground, but temperatures only a few degrees below zero. (A welcome surprise because December on the prairie can be cruel. I’m still reeling from a visit a few years ago where every day of my stay hovered at or below -30C.)

![Fish Creek in Calgary](/assets/images/nanoglyphs/007-civilization/fish-creek@2x.jpg)

Hopefully everyone's got a few relaxation goals (wishes?) for their time off. Personally, I'm hoping to eat and drink plenty, and with any luck recoup a little creative energy. Happy holidays, and see you in 2020.
