+++
published_at = 2019-12-27T06:29:34Z
title = "Actix Web: Optimizing Amongst Optimizations"
+++

![Rust](/assets/images/nanoglyphs/008-actix/rust@2x.jpg)

---

Many in the web development community will already be familiar with the TechEmpower [web framework benchmarks](https://www.techempower.com/benchmarks/) which pit various web frameworks in various languages against each other.

Benchmarks like this tend to draw a lot of fire because although results are presented definitely, they can be misleading. Two languages/frameworks may have similar performance properties, but one of the two has sunk a lot more time into optimizing their benchmark implementation, allowing it to pull ahead of its mate. This tends to be less of an issue over time as benchmarks mature and every implementation gets more optimized, but they should be taken with some grain of salt.

That said, although these benchmark games don't tell us everything, they do tell us _something_. For example, no matter how heavily one of the Ruby implementations is optimized, it'll never beat PHP, let alone a fast language like C++, Go, or Java -- the inherent performance disparity is too great.

## Round 18 (#round-18)

An upset during the latest round ([18](https://www.techempower.com/benchmarks/#section=data-r18&hw=ph&test=fortune)) was [Actix web](https://github.com/actix/actix-web) pulling ahead of the rest of the pack by a respectable margin:

![Fortunes round 18 results](/assets/images/nanoglyphs/008-actix/fortunes@2x.jpg)

TechEmpower runs a few different benchmarks, and this is specifically the _Fortunes_, where implementations are tasked with a simple but realistic workload that exercises a number of facets like database communication, HTML rendering, and unicode handling:

> In this test, the framework's ORM is used to fetch all rows from a database table containing an unknown number of Unix fortune cookie messages (the table has 12 rows, but the code cannot have foreknowledge of the table's size). An additional fortune cookie message is inserted into the list at runtime and then the list is sorted by the message text. Finally, the list is delivered to the client using a server-side HTML template. The message text must be considered untrusted and properly escaped and the UTF-8 fortune messages must be rendered properly.

Fortunes is the most interesting of TechEmpower's series of benchmarks because it does more. Those that do something more simplistic like send a canned JSON or plaintext response have more than a dozen frameworks that perform almost identically to each other because they've all done a good job in ensuring that one piece of the pipeline is well optimized.

## Actix web (#actix)

![Actix results explanation](/assets/images/nanoglyphs/008-actix/actix-explanation@2x.jpg)
