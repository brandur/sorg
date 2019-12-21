+++
published_at = 2019-12-20T03:19:40Z
title = "Preventing the Collapse of Civilization"
+++

![Palace of the Fine Arts in 2015](/assets/images/nanoglyphs/007-civilization/palace@2x.jpg)

We held our annual internal conference last week, with our out of town employees all converging on San Francisco for the largest simultaneous gathering all year. Like this year, the first of these I went to years ago was held at the Palace of the Fine Arts, an event space in the Marina. But while in that first year we occupied only a tiny corner of the building, this year the massive event space felt positively crowded as the company’s 10x’ed in size compared to then — quite a turnaround.

Most of the event was superlative-Bingo troop rallying of the kind you might expect from a large company, but I was pleasantly surprised by the external talks that’d been lined up, including one from Jonathan Blow,  best known as the creator of _Braid_ and _The Witness_. Beyond games, Jonathan’s well known to have strong opinions on software development, and has gone so far as to build his own programming language — the yet unreleased JAI.

No slides were posted for this talk, but much of the same philosophy is conveyed in the excellent [_Preventing the Collapse of Civilization_](https://www.youtube.com/watch?v=pW-SOdj4Kkk). To sum it up in as few words as possible (and do it some injustice by doing so): programmers are taught early to build layers of abstraction, with the premise being that each new layer brings more safety and more productivity. Jonathan’s contention is that the model’s core assumption — that abstraction is free — is wrong. Layers have weight, and moving up an ever deepening stack compromises robustness and productivity. We've internalized this truth by not expecting software to work properly -- no one is surprised to encounter minor bugs in almost every program or web app we use, and we all do so on a daily basis.

I’ve previously called _Preventing the Collapse of Civilization_ the most important talk for the future of the field of software engineering, period. If we're going to make major advances 

Admittedly, I’m a fan, and proponents of elaborate constructs of dependencies, deep frameworks, and heavy build/transpilation pipelines aren’t going to enjoy the implications as much, but it’s worth an hour of anyone’s time.

---

## A scalability epic (#odyssey)

Yandex released [version 1.0 of Odyssey](https://github.com/yandex/odyssey/releases/tag/1.0), a fast connection pooler for Postgres. Unlike PgBouncer which is single-threaded, it’s built with a multi-threaded architectural model, and the various improvements that come in with 1.0 put its feature set on par with any comparable projects. The release notes indicate that Yandex is using Odyssey to run more than 1,000,000 requests per second across many hundreds of hosts.

Yandex has previously talked about migration their mail service [from Oracle to Postgres](https://www.pgcon.org/2016/schedule/attachments/426_2016.05.19%20Yandex.Mail%20success%20story.pdf) (PDF warning), and it suggests that they were managing 300 TB of core data in 2016, which makes it fair to call them one of the largest Postgres users in the wild. They also provide a managed Postgres service, so if any user is likely to have developed deep operational dexterity with Postgres, it’s them.

## Adaptive runtimes (#adaptive-runtimes)

The async-std project in Rust writes about their new [Go-inspired runtime](https://async.rs/blog/stop-worrying-about-blocking-the-new-async-std-runtime/). The current state of asynchronous Rust is pretty convoluted, so a quick refresher for anyone not living on the razor’s edge:

* The `async`/`await` language keywords have landed in stable Rust.
* Although these basic primitives are available in core, Rust notably does not include an asynchronous runtime. Users bring their own, the [Tokio runtime](https://docs.rs/tokio/0.2.6/tokio/runtime/index.html) being the most common.
* The async-std project aims to provide async equivalents for everything found in Rust’s standard library. For example, reading a file is a blocking operation, but async-std provides an async alternative. The project also includes a runtime that can be used as an alternative to Tokio.

A traditional problem with the async model (compared to threads or green threads) is _blocking_. Runtimes typically rely on functions finishing quickly (or pausing on other asynchronous work) so that the runtime can move on to something else. Node’s runtime for example is quite infamously single-threaded, so it can only be doing one thing at a time. A function that blocks through either a compute-heavy workload or by waiting on I/O in a way that’s not compatible with its reactor gums up the whole system.

Th async-std executor is interesting because it’s designed to detect when a single task has been blocking for too long and responds by spinning up a new executor thread to replace the old one. It starts single-threaded, but adaptively scales itself up to a multi-threaded model as its workload demands.

## Best cloud per buck (#cloud-report)

Cockroach Labs published their [2020 Cloud Report](https://www.cockroachlabs.com/blog/2020-cloud-report/), which compares performance across major cloud providers. They’ve been publishing the report for a few years now, and the most shocking result is how much ground Google’s covered over the last year. AWS outperformed GCP by 40% in the last report, which they attributed to Amazon’s [Nitro System](https://aws.amazon.com/ec2/nitro/) present in newer instance types. This year, that edge is gone.

Cockroach benchmarks using the realistic [TPC-C](http://www.tpc.org/tpcc/) which measures database OLTP performance by modeling warehouses executing transactions. In terms of throughput, AWS’ `c5d.4xlarge` came out on top, but only by a narrow margin relative to comparable products from Azure and GCP. More interestingly, the report doesn’t stop there — it goes on to also measure the performance of each cloud _relative to cost_, and along that dimension GCP’s `n2-highcpu-16` headlines the list.

Again, the differences are minor enough to the point where they’re not all that important, but the big takeaway here should be that for a realistic workload, there isn’t that much difference between the big three anymore.

---

LOLWUT
