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

## Comparison chart (#chart)

<figure>
  <table class="overflowing">
    <tr>
      <th>Database</th>
      <th>ACID</th>
      <th>HA</th>
      <th>Scalability</th>
      <th>Partitioning Scheme [1]</th>
      <th>Notes</th>
    </tr>
    <tr>
      <td><strong>Amazon Aurora</strong></td>
      <td>✔</td>
      <td>✔</td>
      <td>✔ Disk only</td>
      <td>✔ Automatic</td>
      <td></td>
    </tr>
    <tr>
      <td><strong>CitusDB</strong></td>
      <td>✔</td>
      <td>✔</td>
      <td>✔ CPU/memory/disk</td>
      <td>✘ Manual</td>
      <td>ACI* is node local</td>
    </tr>
    <tr>
      <td><strong>CockroachDB</strong></td>
      <td>✔</td>
      <td>✔</td>
      <td>✔ CPU/memory/disk</td>
      <td>✔ Automatic</td>
      <td>5 ms floor</td>
    </tr>
    <tr>
      <td><strong>Google Spanner</strong></td>
      <td>✔</td>
      <td>✔</td>
      <td>✔ CPU/memory/disk</td>
      <td>✔ Automatic</td>
      <td>5 ms floor</td>
    </tr>
    <tr>
      <td><strong>MongoDB</strong></td>
      <td>✘ </td>
      <td>✔</td>
      <td>✔ CPU/memory/disk</td>
      <td>✘ Manual</td>
      <td>Open source; not recommended (see notes)</td>
    </tr>
    <tr>
      <td><strong>Postgres</strong></td>
      <td>✔</td>
      <td>✔</td>
      <td>✘  Single node</td>
      <td></td>
      <td>Open source; HA through Amazon RDS, Heroku
      Postgres, or Azure Database</td>
    </tr>
  </table>
  <figcaption>Comparison of cloud databases.</figcaption>
</figure>

## The contenders (#contenders)

### Amazon Aurora (#aurora)

### CitusDB (#citusdb)

### CockroachDB (#cockroachdb)

Ex-Google people.

Written in Go.

> Postgres interface and ORM support

> Interface with a robust SQL API that supports distributed ACID transactions, foreign keys, secondary indexes, JOINs, aggregations, and zero-downtime schema changes.

Known implementations (nice!):

https://www.cockroachlabs.com/docs/known-limitations.html

Clocks:

> However, CockroachDB does require moderate levels of clock synchronization for correctness. If clocks drift past a maximum threshold, nodes will be taken offline. It’s therefore highly recommended to run NTP or other clock synchronization software on each node.

Scaling:

> CockroachDB scales horizontally with minimal operator overhead. You can run it on your local computer, a single server, a corporate development cluster, or a private or public cloud. Adding capacity is as easy as pointing a new node at the running cluster.

> At the key-value level, CockroachDB starts off with a single, empty range. As you put data in, this single range eventually reaches a threshold size (64MB by default). When that happens, the data splits into two ranges, each covering a contiguous segment of the entire key-value space. This process continues indefinitely; as new data flows in, existing ranges continue to split into new ranges, aiming to keep a relatively small and consistent range size.

> When your cluster spans multiple nodes (physical machines, virtual machines, or containers), newly split ranges are automatically rebalanced to nodes with more capacity. CockroachDB communicates opportunities for rebalancing using a peer-to-peer gossip protocol by which nodes exchange network addresses, store capacity, and other information.

Serializable:

> CockroachDB guarantees the SQL isolation level “serializable”, the highest defined by the SQL standard. It does so by combining the Raft consensus algorithm for writes and a custom time-based synchronization algorithms for reads. See our description of strong consistency for more details.

CAP:

> CockroachDB is a CP (consistent and partition tolerant) system. This means that, in the presence of partitions, the system will become unavailable rather than do anything which might cause inconsistent results. For example, writes require acknowledgements from a majority of replicas, and reads require a lease, which can only be transferred to a different node when writes are possible.

### Google Spanner (#spanner)

CAP:

>  Second, the actual theorem is about 100% availability, while the interesting discussion here is about the tradeoffs involved for realistic high availability.

> The purist answer is “no” because partitions can happen and in fact have happened at Google, and during (some) partitions, Spanner chooses C and forfeits A. It is technically a CP system.

> Given that Spanner always provides consistency, the real question for a claim of CA is whether or not Spanner’s serious users assume its availability. If its actual availability is so high that users can ignore outages, then Spanner can justify an “effectively CA” claim. This does not imply 100% availability (and Spanner does not and will not provide it), but rather something like 5 or more “9s” (1 failure in 105 or less). In turn, the real litmus test is whether or not users (that want their own service to be highly available) write the code to handle outage exceptions: if they haven’t written that code, then they are assuming high availability. Based on a large number of internal users of Spanner, we know that they assume Spanner is highly available.

Transactions:

> A transaction in Cloud Spanner is a set of reads and writes that execute atomically at a single logical point in time across columns, rows, and tables in a database.

> Cloud Spanner supports two transaction modes:

> Locking read-write. This type of transaction is the only transaction type that supports writing data into Cloud Spanner. These transactions rely on pessimistic locking and, if necessary, two-phase commit. Locking read-write transactions may abort, requiring the application to retry.
> Read-only. This transaction type provides guaranteed consistency across several reads, but does not allow writes. Read-only transactions can be configured to read at timestamps in the past. Read-only transactions do not need to be committed and do not take locks.

> Because read-only transactions are much faster than locking read-write transactions, we strongly recommend that you do all of your transaction reads in read-only transactions if possible, and only use locking read-write under the scenarios described in the next section.

Serializability:

> Cloud Spanner provides 'serializability', which means that all transactions appear as if they executed in a serial order, even if some of the reads, writes, and other operations of distinct transactions actually occurred in parallel. Cloud Spanner assigns commit timestamps that reflect the order of committed transactions to implement this property. In fact, Cloud Spanner offers stronger guarantees than serializability - transactions commit in an order that is reflected in their commit timestamps, and these commit timestamps are "real time" so you can compare them to your watch. Reads in a transaction see everything that has been committed before the transaction commits, and writes are seen by everything that starts after the transaction is committed.

### Microsoft Cosmos (#cosmos)

### MongoDB (#mongodb)

I've included MongoDB on the list for purposes of
comparison, but even taking into account that it's HA and
horizontally scalable, any of CitusDB, CockroachDB, or
Spanner have a strict superset of its features. They also
support ACID transactions where it doesn't. At this point I
hope that it's a relatively uncontentious thing to say that
technology's moved on; new projects should not start with
MongoDB, and old projects should think about migrating off
it.

### Postgres (#postgres)

HA isn't built in, but is available through Amazon's RDS,
Heroku's Postgres offering, or Azure Database. Google Cloud
SQL also has Postgres support; it doesn't offer HA, but
there are rumors that it's coming soon.

## Final thoughts (#final-thoughts)

[1] I say "automatic scaling" to refer to a data
partitioning scheme that's handled by the database
automatically rather than by the user manually. For
example, when adding data to Cockroach or Spanner, it's
distributed by the system automatically where as with
CitusDB or Mongo, it's up to you to define and maintain a
partitioning scheme.

The latter can be problematic because
it's easy to pick a partitioning scheme that results in
unbalanced partitions; consider if you partition by user,
but some users are far larger than others.
