+++
image_alt = "Portland"
image_url = "/photographs/nanoglyphs/034-cloud-sqlite/portland@2x.jpg"
published_at = 2022-05-29T20:09:45Z
title = "Cloud SQLite"
+++

Seemingly by coincidence, two articles came out nearly consecutively which could turn out to be major landmarks for use of databases in the cloud:

* [**I'm all-in on Server-side SQLite**](https://fly.io/blog/all-in-on-sqlite-litestream/): Litestream is a companion project to SQLite which intercepts its normal WAL process to continuously archive segments to a blob store like S3. In this post, its maintainer talks about joining Fly.io and why he's optimistic about the future of the project.

* [**Announcing D1: Our first SQL database**](https://blog.cloudflare.com/introducing-d1/): Also based on SQLite, D1 is a database from CloudFlare aimed at providing a persistence layer for distributed applications. This is a big deal for them because previously their suggested storage mechanism was [a key/value store](https://developers.cloudflare.com/workers/runtime-apis/kv/).

Both Litestream and D1 appear to be a similar concept -- have an SQLite database archive continuously to one of the major cloud providers' blob stores, which have practically infinite scalability. Processes in other regions stay on top of new segments to make read replicas available that can be queried with far lower latency than the primary. Because SQLite databases are just a file, monolithic daemons in each region aren't necessary, with each edge worker effectively doing its own database legwork through its SQLite driver.

Imagine for a moment an idealized version of this. Fly.io operates in [21 regions around the world](https://fly.io/docs/reference/regions/) from Spain to Singapore. An app chooses one region as its primary, but practically effortlessly, gets automatic read replicas in the other twenty, requiring only lightweight archive tailers to rebuild local copies in each. With that in place, globally distributed workers have millisecond-level read latency while servicing users local to their edge, keeping everything snappy regardless of user locale. And because no large database processes are required, it wouldn't just be the world's largest apps that could be made global -- even small hobby apps could take advantage of it with minimal costs.

## The future is global (#future-is-global)

This is one of those moments where we might legitimately be looking at a turning point for web technology -- like when the combination of MySQL and PHP made it possible for anyone to build a dynamic, stateful application, or when Amazon changed the world with S3 and EC2. We could look back on this moment in ten years and think of monolithic programs like Postgres akin to how we think about Oracle today -- hopelessly outdated in our modern world of fully distributed, streaming databases.

SQLite has a lot going for it. It's the world's mostly widely deployed database by virtue of its widespread non-server use on integrated devices and elsewhere. It takes [backwards compatibility very seriously](https://www.sqlite.org/onefile.html), with modern versions interoperable all the way back to files created in 2004:

> All releases of SQLite version 3 can read and write database files created by the very first SQLite 3 release (version 3.0.0) going back to 2004-06-18. This is "backwards compatibility". The developers promise to maintain backwards compatibility of the database file format for all future releases of SQLite 3.

Especially compared to waning projects like MySQL, SQLite's done a good job of identifying and implementing newer SQL features that are likely to be most useful to its users:

* Adopted [upsert from Postgres](https://www.sqlite.org/lang_UPSERT.html) back in 2018.
* [CTEs](https://www.sqlite.org/lang_with.html).
* [JSON](https://www.sqlite.org/json1.html).
* [Full-text search](https://www.sqlite.org/fts5.html).
* Window functions.

It has the most comprehensive and well-regarded [testing methodology](https://www.sqlite.org/testing.html) of any software project in existence. Among other things:

* Multiple test harnesses: 50k TCL script parameterized test cases run under different permutations, C-level tests for branch coverage, SQL logic tests, and a fuzz tester.
* 100% branch coverage.
* Millions of test cases.
* Edge testing including out-of-memory tests, IO error tests, boundary, and crash tests.
* Static analysis via Valgrind.

In short, SQLite is one of the world's best run software projects, and it's not going anywhere.

## Is it hype? (#hype)

The purpose of this section isn't to cast aspersion on an interesting technology, but as a reminder to keep things in perspective because as an industry, we have a hype problem. This can be considered a good thing because it keeps us optimistic and moving forward, but over the years plenty of new technologies have been hailed as panacea, only to be relegated to the dustbin of history only a few years later. Having been consistently guilty of this myself (see [JSON Hyper-schema](/elegant-apis), [GraphQL](/graphql), or [cloud databases](/cloud-databases)), I try to be careful these days.

A few thoughts:

* The elephant in the room is write contention. SQLite has a [well-designed concurrency model](https://www.sqlite.org/lockingv3.html), but at the end of the day, when a large number of workers are trying to cooperate, it's at a _massive_ disadvantage compared to an in-memory system like Postgres that can coordinate concurrent operations far more easily.

    This is made somewhat worse by just how important the database is to most applications. I've yet to see a large-scale operation where the database isn't just the bottleneck, it's the most critical bottleneck _by far_, like orders of magnitude. Application processes tend to be trivially parallelizable compared to figuring out what to do with a hot DB.

* Backwards compatibility has tradeoffs. To maximize compatibility, SQLite only stores five types of values: `NULL`, integers, floats, text, and raw blobs. So for example,  although it has good support for JSON, that JSON is always stored as text. Compare this Postgres, which stores JSON as [preparsed `jsonb`](https://www.postgresql.org/docs/current/datatype-json.html). The storage limitations also mean you're giving up native support for a lot of other data types: datetimes, UUIDs, arrays, ranges, etc.

* The replication story for one of these SQLite-based systems compared to Postgres isn't _that_ different. In a cloud Postgres set up, something like [WAL-G](https://github.com/wal-g/wal-g) is used to send WAL archives to a durable blob store. Local processes in other regions ingest WAL and make copies available there with minimal latency. This only starts to break down when there's massive WAL scale to send over large geographic distances, which SQLite isn't going to have any easier of a time with.

A hallmark of branches of technology with incredibly bold claims which ended up far middling than their purveyors intended (Hypermedia APIs, serverless functions-as-a-service like AWS Lambda, dare I say Web3) is that ambitious goals were proclaimed, but never demonstrated much potential in the real world. A technology is never _disproven_ -- instead, it starts to lose its time in the limelight, and ever so slowly, fizzles out.

Personally what I'd like to see are some large-scale uses of server-side SQLite that demonstrate its viability as a modern backend compared to more traditional in-memory RDMSes. With any luck, we'll see a few over the coming years.

## Bridges and brews (#bridges-and-brews)

G'damn Portland is cool.

Earlier in the week I said to a colleague: "Do you ever get the feeling that Portlanders are, like, cosplaying Portlanders?" Walk around town, especially the southeast, every person is sporting loose vintage jeans, toque, Chrome bag, and piercings; half with full sleeve tattoos to boot. It's especially weird coming from San Francisco where not a single resident of the city has worn anything but athleisure going on three years now.

I walked up and down Belmont and Hawthorne last night, and with the number of breweries, weird bars, consignment shops, record stores (and even a ping pong bar) the place has, the only other city I could compare the area to are the best parts of Berlin. The energy on a Saturday night is amazing -- I stopped by a food cart "park" for some pad thai and chicken satay, and even at 10 PM, it was the busiest version of its kind that I've visited.

Quite by accident, I stayed at a hotel with Icelandic roots called Kex (Icelandic for "biscuit") which has a sister hostel in Reykjavík. Small rooms, but in true Nordic style, a sauna in the basement, and between luxurious space and fast wi-fi, the best main floor workspace I've ever seen. Like, if you found a neighborhood cafe like this, you'd go there every day.

<img src="/photographs/nanoglyphs/034-cloud-sqlite/kex-2@2x.jpg" alt="The Kex cafe (no. 2)" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/034-cloud-sqlite/kex@2x.jpg" alt="The Kex cafe" class="wide" loading="lazy">

## High-effort comedy (#high-effort-comedy)

Speaking of Iceland, my mind keeps wandering back back to one of the best marketing campaigns in years: Inspired by Iceland's [Outhorse your email](https://www.visiticeland.com/outhorse-your-email/).

You have to imagine that this is one of those projects that started out as a gag that some out-of-the-box marketers people came up with at the bar, but which was pushed to ever-more-extraordinary lengths.

<img src="/photographs/nanoglyphs/034-cloud-sqlite/outhorse-your-email@2x.jpg" alt="Ourhorse your email" class="wide" loading="lazy">

Layer one of effort would be to make a website, write some clever copy, and post some stock photography of Icelandic horses. Maybe a little Photoshop.

Check.

Layer two would be to construct a prop keyboard and film some horses walking over it in various exotic locations including beyond-majestic Icelandic coastlines and waterfalls.

_**Check.**_

Layer three would be to wire the prop with sensors and Raspberry Pis under each key, give it a battery power supply, and have it really transmit back to a PC, making it the world's largest wireless keyboard, for real, and making those horses typing content, for real.

[CHECK](https://vimeo.com/710288765/5e14861065).

How can you not admire that sort of commitment to a bit?

Until next week.
