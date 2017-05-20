---
title: A Comparison of Modern Cloud Databases
published_at: 2017-05-19T15:54:21Z
location: San Francisco
hook: A non-exhaustive primer of some modern cloud database
  solutions (as of 2017).
---

Here's some introductory text.

I came up with a list of key characteristics to list in my
comparison chart based on their importance in building and
operating software. This is fairly subjective, and comes
with a limitless number of caveats in what's been left
unsaid. I provide some additional context in the sections
below, but even there there's far more details and
subtleties than I could possibly include in this article.

## Comparison matrix (#matrix)

<figure>
  <table class="overflowing">
    <tr>
      <th>Database</th>
      <th>ACID</th>
      <th>HA</th>
      <th>Horizontally Scalable</th>
      <th>Automatic Data Distribution</th>
      <th>Low Latency</th>
      <th>Notes</th>
    </tr>
    <tr>
      <td><strong>Amazon Aurora</strong></td>
      <td>✔</td>
      <td>✔</td>
      <td>✔ Disk only</td>
      <td>✔</td>
      <td>✔</td>
      <td></td>
    </tr>
    <tr>
      <td><strong>CitusDB</strong></td>
      <td>✔</td>
      <td>✔</td>
      <td>✔</td>
      <td>✘</td>
      <td>✔</td>
      <td>ACI* is node local</td>
    </tr>
    <tr>
      <td><strong>CockroachDB</strong></td>
      <td>✔</td>
      <td>✔</td>
      <td>✔</td>
      <td>✔</td>
      <td>✘</td>
      <td>Open source</td>
    </tr>
    <tr>
      <td><strong>Google Spanner</strong></td>
      <td>✔</td>
      <td>✔</td>
      <td>✔</td>
      <td>✔</td>
      <td>✘</td>
      <td></td>
    </tr>
    <tr>
      <td><strong>Microsoft Cosmos</strong></td>
      <td>✘ </td>
      <td>✔</td>
      <td>✔</td>
      <td>✘</td>
      <td>✘</td>
      <td></td>
    </tr>
    <tr>
      <td><strong>MongoDB</strong></td>
      <td>✘</td>
      <td>✔</td>
      <td>✔</td>
      <td>✘</td>
      <td>✔</td>
      <td>Open source; see notes, but not recommended given
        modern alternatives</td>
    </tr>
    <tr>
      <td><strong>Postgres</strong></td>
      <td>✔</td>
      <td>✔</td>
      <td>✘</td>
      <td></td>
      <td>✔</td>
      <td>Open source; HA through Amazon RDS, Heroku
        Postgres, or Azure Database</td>
    </tr>
  </table>
  <figcaption>Feature matrix and notes of cloud
    databases.</figcaption>
</figure>

I compared on these characteristics:

* ***ACID***: Whether the database supports ACID
  guarantees across multiple operations. ACID is a
  [powerful tool for system correctness](/acid), and until
  recently has been a long sought but illusive chimera for
  distributed databases.

* ***HA***: Whether the database is highly available (HA).
  I've marked every one on the list as HA, but some are
  "more HA" than others with CockroachDB, Cosmos, and
  Spanner leading the way in this respect. The others rely
  on a single node failovers.

* ***Horizontal Scalable***: Whether the database can be
  scaled horizontally out to additional nodes. Everything
  on the list except Postgres is, but I've included the
  column to call out the fact that unlike the others,
  Aurora's scalability is disk only. That's not to say that
  it's unsuitable for use, but it has some caveats (see
  [Amazon Aurora](#aurora) below for details).

* ***Automatic Data Distribution:*** Distinguishes
  databases where data partitioning and balancing is
  handled manually by the user versus automatically by the
  database. As an example of a "manual" database, in
  CitusDB you explicitly tell the database that you want a
  table to be distributed and tell it what key should be
  used for partitioning (e.g. `user_id`). Both options are
  workable, but manual distribution has more operational
  overhead and without a lot of care, can lead to
  unbalanced sharding where larger nodes run
  disproportionately hot.

* ***Low latency***: The extra inter-node coordination
  overhead used by CockroachDB Cosmos, and Spanner to
  ensure consistency comes at the cost of being unsuitable
  where very low latency operations are needed (~1 ms). I
  cover this in a little more detail below in [Time-based
  consistency](#time-consistency).

## Additional considerations (#additional-considerations)

### CAP (#cap)

The CAP theorem dictates that given _consistency_, _100%
availability_, and _partition tolerance_, any given
database can satisfy a maximum of two of the three.

I purposely didn't include CAP in the table above. To
explain why, I'll quote [Eric Brewer (Google VP
Infrastructure) writing about Spanner][spanner-truetime]:

> Despite being a global distributed system, Spanner claims
> to be consistent and highly available, which implies
> there are no partitions and thus many are skeptical. Does
> this mean that Spanner is a CA system as defined by CAP?
> The short answer is “no” technically, but “yes” in effect
> and its users can and do assume CA.
>
> The purist answer is “no” because partitions can happen
> and in fact have happened at Google, and during (some)
> partitions, Spanner chooses C and forfeits A. It is
> technically a CP system. We explore the impact of
> partitions below.
>
> Given that Spanner always provides consistency, the real
> question for a claim of CA is whether or not Spanner’s
> serious users assume its availability. If its actual
> availability is so high that users can ignore outages,
> then Spanner can justify an “effectively CA” claim. This
> does not imply 100% availability (and Spanner does not
> and will not provide it), but rather something like 5 or
> more “9s” (1 failure in 10^5 or less). In turn, the real
> litmus test is whether or not users (that want their own
> service to be highly available) write the code to handle
> outage exceptions: if they haven’t written that code,
> then they are assuming high availability. Based on a
> large number of internal users of Spanner, we know that
> they assume Spanner is highly available.

In other words, modern techniques can achieve CP while
still keeping availability that's incredibly good. Like
_five or more 9s of good_. This result is so optimal that
modern databases seem to be converging on it. Every
database on the list above is _CP_ with varying levels of
_A_ [1].

### Time-based consistency (#time-consistency)

## The contenders (#contenders)

### Amazon Aurora (#aurora)

Aurora is a managed relational database that has an SQL
interface that's compatible with MySQL and Postgres. One of
its biggest selling points is performance, and it claims to
provide 5x the throughput of MySQL and 2x of Postgres
running on the same hardware.

Aurora is quite distinctive from any other option on this
list because it's not horizontally scalable at the node
level, and its clusters more resemble those of a
traditional RDMS with a primary and read replicas. Instead,
Amazon has devised a storage-level scaling scheme that
allows its tables to grow to sizes larger than an RDMS (up
to 64 TB per table).

This storage-based scaling has the disadvantage that
compute and memory resources (for writes or consistent
reads) are limited to a single vertically scaled node
[2], but it also has significant advantages as well: data
is always colocated so query latency is very low. It also
means that you can't make a mistake choosing a partition
scheme and end up with a few hot shards that need to be
rebalanced (which is _very_ easy to do and _very_ hard to
fix). As such, it may be a more appropriate choice than
solutions like CockroachDB or Spanner for users looking for
extensive scalability, but who don't need it to be
infinite.

### CitusDB (#citusdb)

CitusDB is a distributed database build on top of Postgres
that allows individual tables to be sharded and distributed
across a number of different nodes. It also provides a
clever concept called _reference tables_ to help ensure
data locality to improve query performance.

Most notably, CitusDB is open source and runs using
Postgres' standard extension API. This reduces the risk of
lock in, which is a considerable downside of most of the
other options on this list. Compared to Aurora, it also
means that you're more likely to see new features from new
Postgres releases make it into your database.

Downsides compared to CockroachDB and Spanner are that it
relies on a central coordinator, which means that its HA is
more akin to that of Aurora or Postgres, and that data is
sharded manually, which as noted above, can lead to
balancing problems. Another consideration is that it's
built by an upstart company with a yet unproven business
model. Generally when selecting a database, it's nice to
have something that's going to be still around and
well-maintained in ten years time, and you can have pretty
good assurance of that with something made by Amazon,
Google, or Microsoft.

### CockroachDB (#cockroachdb)

CockroachDB is a product built out of Cockroach Labs, a
company founded by ex-Googlers who are known to have been
influencial in building Google File System and Google
Reader. It's based on the same ideas as Spanner, and like
spanner, uses a time-based mechanic to achieve consistency,
but without the benefit of Google's GPS and atomic clocks.

It provides serializable distributed transactions, foreign
keys, and secondary indexes. It's open source and written
in Go which gives it the nice property of being easily
installable and runnable in a development environment.
Their documentation is refreshingly well-written, easily
readable, and honest. Take for example their [list of known
limitations][cockroach-limitations].

Like Spanner, the additional overhead of guaranteeing
distributed consistency means that it's a poor choice where
low latency operations are needed ([they admit as much
themselves][cockroach-not-good-choice]). Like CitusDB
above, the fact that it's built by a small company with an
unproven business model is a con worth considering.

### Microsoft Cosmos (#cosmos)

[Cosmos][cosmos] is Microsoft's brand-new cloud database.
Its got a few major selling points:

* Fast and easy geographical distribution.
* A configurable consistency model that allows anything
  from strong serializability all the way down to eventual
  consistency which trades the possibility of out-of-order
  reads for speed.
* SLAs on operation timing that guarantees reads under 10
  ms and indexed writes under 15 ms at the 99th percentile.

Like with CockroachDB and Spanner, the distribution of
Cosmos makes it less suitable for work requiring very low
latency operations. Their documentation suggests a media
read and write latency of ~5 ms.

The sales pitch for Cosmos comes on pretty strong. Here's
an excerpt where they sell schemaless design, which put
most generously is a well known trade off, and less so, an
anti-feature:

> Both relational and NoSQL databases force you to deal
> with schema & index management, versioning and migration
> [...] But don’t worry -- Cosmos DB makes this problem go
> away!

Cosmos is also unable to provide ACID transactions, which
puts it at a distinct disadvantage compared to some of the
other options on this list.

### MongoDB (#mongodb)

I've included MongoDB on the list for purposes of
comparison, but it's not at the same level of
sophistication as other systems on this list. Most others
have a strict superset of its functionality (albeit with
trade offs in a few cases), but also support other
critically important features like ACID guarantees. New
projects shouldn't start on MongoDB, and old projects
should be thinking about migrating off of it.

### Postgres (#postgres)

Postgres is the trusty workhorse of traditional RDMSes. HA
isn't built in, but is available through offerings from
Amazon RDS, Heroku Postgres, or Azure Database (and
hopefully Google Cloud SQL soon).

Even though it's not a perfect fit for the rest of this
list, I've included it anyway because it's often still the
best option for most use cases. [Most organizations don't
have data that's as big as they think it
is](/acid#scaling), and by keeping it small through
curation, they can get away with vertically scaled
Postgres. This will lead to a more operable stack, and more
options in case it's ever necessary to migrate between
clouds and providers. You can also easily run Postgres
locally or in testing which is _hugely important_ for
friction-free productivity.

## Final thoughts (#final-thoughts)

[1] The _CAP_ properties of Cosmos and MongoDB are
configurable as they can both be made to be eventually
consistent.

[2] Aurora nodes are currently scalable to 32 vCPUs and 244
GB of memory. Although that is "only" one node, it's
nothing to scoff at and should provide enough runway for
the vast majority of use cases.

[cockroach-limitations]: https://www.cockroachlabs.com/docs/known-limitations.html
[cockroach-not-good-choice]: https://www.cockroachlabs.com/docs/frequently-asked-questions.html#when-is-cockroachdb-not-a-good-choice
[cosmos]: https://docs.microsoft.com/en-us/azure/cosmos-db/introduction
[spanner-truetime]: https://research.google.com/pubs/pub45855.html
