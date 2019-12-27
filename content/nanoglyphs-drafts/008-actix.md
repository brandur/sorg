+++
published_at = 2019-12-27T06:29:34Z
title = "Optimizing Amongst Optimizers"
+++

![Rust](/assets/images/nanoglyphs/008-actix/rust@2x.jpg)

---

Benchmarks of this nature tend to draw a lot of fire because although the results seem to be presented definitely, they can be misleading. Two languages/frameworks may have similar performance properties, but one of the two has sunk a lot more time into optimizing their benchmark implementation, allowing it to pull ahead of its mate. The results tend to get more accurate over time as the benchmark grows in popularity and receives more contributions, but they should be considered critically.

But although these benchmark games don't tell us everything, they do tell us _something_. For example, no matter how heavily one of the Ruby implementations is optimized, it'll never beat PHP, let alone a fast language like C++, Go, or Java -- the inherent performance disparity is too great.

[round 18 of the Fortunes benchmark](https://www.techempower.com/benchmarks/#section=data-r18&hw=ph&test=fortune).

> In this test, the framework's ORM is used to fetch all rows from a database table containing an unknown number of Unix fortune cookie messages (the table has 12 rows, but the code cannot have foreknowledge of the table's size). An additional fortune cookie message is inserted into the list at runtime and then the list is sorted by the message text. Finally, the list is delivered to the client using a server-side HTML template. The message text must be considered untrusted and properly escaped and the UTF-8 fortune messages must be rendered properly.

Fortunes might be the most interesting of the TechEmpower benchmarks because it does a fair bit of work. There are other benchmarks for sending with a simple JSON or plaintext response, but the trouble with them is that the workload is so simple that more than a dozen frameworks perform almost identically to each other.

![Fortunes results](/assets/images/nanoglyphs/008-actix/fortunes@2x.jpg)

