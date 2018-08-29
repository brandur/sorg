---
title: A Comparison of Advanced, Modern Cloud Databases
published_at: 2017-05-22T17:24:13Z
location: San Francisco
hook: A non-exhaustive primer of modern cloud database
  solutions like Aurora, Cosmos, and Spanner.
---

In the last few years we've seen the emergency of some
impressive cloud technology ranging from databases like
very reminiscent of current RDMSes except with better
scalability (Aurora), to those with novel new designs that
take advantage of custom hardware for the guarantees that
they need to scale (Spanner). While unbounded data growth
may still have a logistical problem for many organizations,
the tools that we have today to manage it have never been
better.

It can be a little hard to keep track of the new entrants
and track how exactly they differ from one another, so here
I've tried to summarize various offerings and how they
compare to one another.

It's impossible to rate any one as a clear winner because
like any consideration in technology, there are trade offs
for everything, and organizations will largely have to
select technology based on what will be valuable to them.
Many of the characteristics in my comparison matrix below
were selected based on their importance in building robust
software, but I admit there's some bias there. There's also
some bias in which technologies even show up. There are
dozens of options and I've excluded the vast majority; the
list is scoped down to the some of the most general
purpose, most practical, and most interesting.

I put some opinions on favorites in ["Closing
thoughts"](#closing-thoughts) below.

## Comparison matrix (#matrix)

<figure>
  <div class="table-container">
    <table class="overflowing">
      <tr>
        <th>Database</th>
        <th>Concurrent ACID</th>
        <th>HA</th>
        <th>Horizontally Scalable</th>
        <th>Automatic Scalability</th>
        <th>Low Latency</th>
        <th>Notes</th>
      </tr>
      <tr>
        <td><strong><a href="#aurora">Amazon Aurora</a></strong></td>
        <td>✓</td>
        <td>✓</td>
        <td>✓ Disk only</td>
        <td>✓ Single node; see notes</td>
        <td>✓</td>
        <td></td>
      </tr>
      <tr>
        <td><strong><a href="#citusdb">Citus</a></strong></td>
        <td>✓</td>
        <td>✓</td>
        <td>✓</td>
        <td></td>
        <td>✓</td>
        <td>Open source; ACI* is node local</td>
      </tr>
      <tr>
        <td><strong><a href="#cockroachdb">CockroachDB</a></strong></td>
        <td>✓</td>
        <td>✓</td>
        <td>✓</td>
        <td>✓</td>
        <td></td>
        <td>Open source</td>
      </tr>
      <tr>
        <td><strong><a href="#spanner">Google Spanner</a></strong></td>
        <td>✓</td>
        <td>✓</td>
        <td>✓</td>
        <td>✓</td>
        <td></td>
        <td></td>
      </tr>
      <tr>
        <td><strong><a href="#cosmos">Microsoft Cosmos</a></strong></td>
        <td></td>
        <td>✓</td>
        <td>✓</td>
        <td></td>
        <td></td>
        <td></td>
      </tr>
      <tr>
        <td><strong><a href="#mongodb">MongoDB</a></strong></td>
        <td></td>
        <td>✓</td>
        <td>✓</td>
        <td></td>
        <td>✓</td>
        <td>Open source; not recommended given modern
          alternatives (see notes)</td>
      </tr>
      <tr>
        <td><strong><a href="#postgres">Postgres</a></strong></td>
        <td>✓</td>
        <td>✓</td>
        <td></td>
        <td>N/A</td>
        <td>✓</td>
        <td>Open source; HA through Amazon RDS, Heroku
          Postgres, or Azure Database</td>
      </tr>
    </table>
  </div>
  <figcaption>Feature matrix and notes of cloud
    databases.</figcaption>
</figure>

Here's the meaning of each column:

* ***Concurrent ACID:*** Whether the database supports ACID
  (atomicity, consistency, isolation, and durability)
  guarantees across multiple operations. ACID is a
  [powerful tool for system correctness](/acid), and until
  recently has been a long sought but illusive chimera for
  distributed databases. I use the term "concurrent ACID"
  because technically Cosmos guarantees ACID, but only
  within the confines of a single operation.

* ***HA:*** Whether the database is highly available (HA).
  I've marked every one on the list as HA, but some are
  "more HA" than others with CockroachDB, Cosmos, and
  Spanner leading the way in this respect. The others tend
  to rely on a single node failovers.

* ***Horizontally Scalable:*** Whether the database can be
  scaled horizontally out to additional nodes. Everything
  on the list except Postgres is, but I've included the
  column to call out the fact that unlike the others,
  Aurora's scalability is disk only. That's not to say that
  it's unsuitable for use, but it has some caveats (see
  ["Amazon Aurora"](#aurora) below for details).

* ***Automatic Scalability:*** Distinguishes databases
  where data partitioning and balancing is handled manually
  by the user versus automatically by the database. As an
  example of a "manual" database, in Citus or MongoDB you
  explicitly tell the database that you want a table to be
  distributed and tell it what key should be used for
  sharding (e.g. `user_id`). For comparison, Spanner
  automatically figures out how to distribute any data
  stored to it to the nodes it has available, and
  rebalances as necessary. Both options are workable, but
  manual distribution has more operational overhead and
  without care, can lead to unbalanced sharding where
  larger nodes run disproportionately hot.

* ***Low latency:*** The extra inter-node coordination
  overhead used by CockroachDB Cosmos, and Spanner to
  ensure consistency comes at the cost of being unsuitable
  where very low latency operations are needed (~1 ms). I
  cover this in a little more detail below in ["Time-based
  consistency"](#time-consistency).

## Additional considerations (#additional-considerations)

### CAP (#cap)

The CAP theorem dictates that given _consistency_, _100%
availability_, and _partition tolerance_, any given
database can satisfy a maximum of two of the three.

To explain why I didn't include CAP in the table above,
I'll quote [Eric Brewer (Google VP Infrastructure) writing
about Spanner][spanner-truetime]:

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
_A_ (with some caveats [1]).

### Time-based consistency (#time-consistency)

Sophisticated distributed systems like Spanner and
CockroachDB tend to need a little more time to coordinate
and verify the accuracy of the results that will be
returned from any given node, and this makes them less
suitable for low latency operations.

Quizlet suggests that the minimum latency for a Spanner
operation [is ~5 ms][spanner-quizlet]. The [Spanner
paper][spanner-paper] describes the details of the
coordination for various operations in sections 4.1. and
4.2. CockroachDB states very explicitly in their FAQ that
[it's not as good of a choice where low latency reads and
writes are critical][cockroach-not-good-choice].

The design of Microsoft's Cosmos isn't as transparent, but
its documentation seems to suggest similar performance
characteristics with [the median time for reads and writes
at 5 ms][cosmos-99th].

## The contenders (#contenders)

### Amazon Aurora (#aurora)

[Aurora][aurora] is a managed relational database that has
an SQL interface that's compatible with MySQL and Postgres.
One of its biggest selling points is performance, and it
claims to provide 5x the throughput of MySQL and 2x of
Postgres running on the same hardware.

Aurora is quite distinctive from any other option on this
list because it's not horizontally scalable at the node
level, and its clusters more resemble those of a
traditional RDMS with a primary and read replicas. Instead,
Amazon has devised a storage-level scaling scheme that
allows its tables to grow to sizes significantly larger
than you'd see with a traditional RDMS; up to 64 TB per
table.

This storage-based scaling has the disadvantage that
compute and memory resources (for writes or consistent
reads) are limited to a single vertically scaled node
[2], but it also has significant advantages as well: data
is always colocated so query latency is very low. It also
means that you can't make a mistake choosing a partition
scheme and end up with a few hot shards that need to be
rebalanced (which is _very_ easy to do and _very_ hard to
fix). It may be a more appropriate choice than solutions
like CockroachDB or Spanner for users looking for extensive
scalability, but who don't need it to be infinite.

### Citus (#citusdb)

[Citus][citus] is a distributed database built on top of
Postgres that allows individual tables to be sharded and
distributed across any number of nodes. It provides clever
concepts like _reference tables_ to help ensure data
locality to improve query performance. ACID guarantees are
scoped to particular nodes, which is often adequate given
that partitioning is designed so that data is colocated.

Most notably, Citus is open source and runs using the
Postgres extension API. This reduces the risk of lock in,
which is a considerable downside of most of the other
options on this list. Compared to Aurora, it also means
that you're more likely to see features from new Postgres
releases make it into your database.

A downside compared to CockroachDB and Spanner is that it
data is sharded manually, which as noted above, can lead to
balancing problems. Another consideration is that it's
built by a fairly new company with a yet unproven business
model. Generally when selecting a database, it's good for
peace of mind to know that you're using something that's
almost certainly going to be around and well-maintained in
ten years time. You can be pretty confident of that when
the product is made by a behemoth like Amazon, Google, or
Microsoft, but less so for smaller companies.

### CockroachDB (#cockroachdb)

[CockroachDB][cockroach] is a product built out of
Cockroach Labs, a company founded by ex-Googlers who are
known to have been influential in building Google File
System and Google Reader. It's based on the design laid out
by the original Spanner paper, and like spanner, uses a
time-based mechanic to achieve consistency, but without the
benefit of Google's GPS and atomic clocks.

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
themselves][cockroach-not-good-choice]). Like Citus above,
the fact that it's built by a small company with an
unproven business model is a downside.

### Microsoft Cosmos (#cosmos)

[Cosmos][cosmos] is Microsoft's brand-new cloud database.
Its sales pitch tends to come on a little strong. For
example, here's an excerpt where they sell schemaless
design, which put most generously is a well known trade
off, and less so, an anti-feature:

> Both relational and NoSQL databases force you to deal
> with schema & index management, versioning and migration
> [...] But don’t worry -- Cosmos DB makes this problem go
> away!

That said, it's got a pretty good set of features:

* Fast and easy geographical distribution.
* A configurable consistency model that allows anything
  from strong serializability all the way down to eventual
  consistency which trades the possibility of out-of-order
  reads for speed.
* SLAs on operation timing that guarantees reads under 10
  ms and indexed writes under 15 ms at the 99th percentile.

Like with CockroachDB and Spanner, the distribution of
Cosmos makes it less suitable for work requiring very low
latency operations. Their documentation suggests a median
read and write latency of ~5 ms.

Cosmos is able to provide ACID through [the use of stored
procedures in JavaScript][cosmos-acid]. This seems to be by
virtue of having only one JavaScript runtime running on the
primary so that only one script is being handled at a time,
but it's also doing some bookkeeping the ensure that any
writes can be reverted, thereby ensuring true atomicity
(unlike say `EVAL` in Redis). Still, this approach is not
as sophisticated as the MVCC engines used by other
databases on this list because it can't provide concurrent
use.

### MongoDB (#mongodb)

MongoDB is a NoSQL data store that stores data as
schemaless JSON documents. It doesn't support ACID
transactions, and if that wasn't enough, since its release
in 2009 has had a number of valid criticisms around core
database competencies like
[durability](/fragments/mongo-durability),
[security][mongo-security], and
[correctness][mongo-correctness].

I've included it for purposes of comparison and because it
still seems to have a lot of mindshare, but it's not at the
same level of sophistication as other systems on this list.
Most others have a strict superset of its functionality
(albeit with trade offs in a few cases), but also support
other critically important features like ACID guarantees.
New projects shouldn't start on MongoDB, and old projects
should be thinking about migrating off of it.

### Postgres (#postgres)

[Postgres][postgres] is the trusty workhorse of traditional
RDMSes. HA isn't built in, but is available through
offerings from Amazon RDS, Heroku Postgres, or Azure
Database (and hopefully Google Cloud SQL soon).

Even though it's not a perfect fit for the rest of this
list, I've included it anyway because it's often still the
best option for many use cases. Most organizations don't
have data that's as big as they think it is, and by
consciously restricting bloat, they can get away with a
vertically scaled Postgres node. This will lead to a more
operable stack, and more options in case it's ever
necessary to migrate between clouds and providers. You can
also easily run Postgres locally or in testing, which is
very important for friction-free productivity.

## Closing thoughts (#closing-thoughts)

Opinion time: the best choice for most people will be to
start with Postgres. It's a battle-tested database with a
spectacular number of features and few limitations. It's
open source and widely available so it can easily be run in
development, CI, or migrated across every major cloud
provider. Vertical scaling will go a long way for
organizations [who curate their data and offload lower
fidelity information to more scalable
stores](/acid#scaling).

After you're at the scale of AirBnB or Uber, something like
Aurora should look interesting. It seems to have many of
the advantages of Postgres, and yet still manages to
maintain data locality and scalable storage (at the costs
of loss of dev/production parity and vendor lock in).
Organizations at this tier who run hot and need compute and
memory resources that are scalable beyond a single node
might benefit from something like Citus instead.

After you're at the scale of Google, something closer to
Spanner is probably the right answer. Although less
suitable for low latency operations, its scalability
appears to be practically limitless.

The only databases on the list that I've seen running in
production are MongoDB and Postgres, so take these
recommendations with a grain of salt. There's almost
certainly hidden caveats to any of them that will only be
uncovered with a lot of hands on experience.

[1] The _CAP_ properties of Cosmos and MongoDB are
configurable as they can both be made to be eventually
consistent.

[2] Aurora nodes are currently scalable to 32 vCPUs and 244
GB of memory. Although that is "only" one node, it's
nothing to scoff at and should provide enough runway for
the vast majority of use cases.

[aurora]: https://aws.amazon.com/rds/aurora/
[citus]: https://www.citusdata.com/
[cockroach]: https://www.cockroachlabs.com/
[cockroach-limitations]: https://www.cockroachlabs.com/docs/known-limitations.html
[cockroach-not-good-choice]: https://www.cockroachlabs.com/docs/frequently-asked-questions.html#when-is-cockroachdb-not-a-good-choice
[cosmos]: https://docs.microsoft.com/en-us/azure/cosmos-db/introduction
[cosmos-acid]: https://docs.microsoft.com/en-us/azure/documentdb/documentdb-programming#database-program-transactions
[cosmos-99th]: https://docs.microsoft.com/en-us/azure/cosmos-db/introduction#low-latency-guarantees-at-the-99th-percentile
[mongo-correctness]: https://blog.meteor.com/mongodb-queries-dont-always-return-all-matching-documents-654b6594a827
[mongo-security]: https://krebsonsecurity.com/2017/01/extortionists-wipe-thousands-of-databases-victims-who-pay-up-get-stiffed/
[postgres]: https://www.postgresql.org/
[spanner-paper]: https://research.google.com/archive/spanner.html
[spanner-quizlet]: https://quizlet.com/blog/quizlet-cloud-spanner
[spanner-truetime]: https://research.google.com/pubs/pub45855.html
