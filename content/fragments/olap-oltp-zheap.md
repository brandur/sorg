---
title: OLAP, OLTP, Zheap, and tradeoffs in Postgres MVCC
published_at: 2018-10-26T18:51:39Z
location: Coimbra
hook: TODO
---

A few talks at a recent Postgres conference that I attended
gave me some ideas around the developing [Zheap storage
engine][zheap] in Postgres [1], how to think about it with
respect to the tradeoffs it makes, and how those fit into
broader database design.

You may have heard database enthusiasts talk about OLAP and
OLTP. They're wordy terms for the uninitiated, but like
many overwrought acronyms, actually represent relatively
simple ideas, and are useful as unambiguous names for two
important concepts:

* **OLAP** is "online analytical processing" and refers to
  a workload of more complex and long-lived analytical
  queries design to glean insight from a data set, like
  you'd see in a normal data warehouse.

* **OLTP** is "online transaction processing" and refers to
  databases tasked with handling large volumes of
  short-lived transactions incoming from users in real
  time. Think of your typical production app.

## Workload specialization (#specialization)

Many databases are suitable for OLAP or OLTP, but not
_both_. For example:

* FoundationDB [limits transactions to 5 seconds][fivesec]
  and supports only key/value storage. That effectively
  makes OLAP impossible, and specializes it for OLTP
  specifically.

* Redshift is capable of crunching huge volumes of data
  into useful answers, but queries to it are infamously
  slow (on the order of seconds or much longer), making it
  good for OLAP but not OLTP which requires much faster
  responses.

* MongoDB encourages document-oriented storage (meaning
  data is not normalized) and uses a homegrown query
  language that's a fraction as expressive as SQL. This
  makes it unsuitable for OLAP (and only precariously
  suitable for OLTP).

Some traditional RDMSes like Postgres are not as
specialized, and do a good job of both OLAP and OLTP. Users
can write extremely complex OLAP queries in SQL (involving
joins, aggregations, [CTEs][cte], etc.) and rely on the
underlying engine to find efficient ways to execute them.
Inserts, updates, deletes, and simple (and well-indexed)
selects are consistently fast, usually finishing in
milliseconds, making them great for OLTP as well.

## Bloat and its effects on OLTP (#bloat)

One of Postgres' most severe and long-standing operational
weaknesses is bloat. In its [MVCC][mvcc] (multiversion
concurrency control) model, both deleted rows and old
versions of updated rows are kept right alongside current
rows in the heap, which is the physical storage where the
contents of tables are stored. Dead rows are eventually
reaped by [vacuum], but only after they're no longer needed
for any running transaction (regardless of how old) and
vacuum gets a chance to run.

It's actually quite an elegant implementation. It's easy to
reason about, and all row versions are readily accessible
by the transactions that need them (I've previously gone
into the details of [its inner workings
here](/postgres-atomicity)). Unfortunately though, it leads
to degenerate performance when relations become "bloated"
because an old transaction is forcing many old versions to
be held which greatly increases their size. Those rows are
still visible to the old transaction and can't be cleaned
out.

We could say that Postgres has optimized for OLAP at the
expense of OLTP. Long-lived transactions have easy access
to contemporaneous row versions, but that same feature
degrades the performance of current, short-lived
transactions.

## Moving the needle: Zheap (#zheap)

Zheap is one of the largest in-flight projects in Postgres
right now. It's a new storage engine that rethinks how MVCC
works. With Zheap, rows are often be updated in place, with
old versions replaced by new ones. Instead of sticking
around in the heap, old versions are moved to an "undo log"
in the current page which acts a historical record. The
idea is inspired by other databases that already use a
similar technique, like Oracle.

If a transaction aborts, old versions from the undo log are
applied until the right version is reached. Similarly, old
transactions that need an old version follow the undo log
back until they find the right one.

In essence, Zheap shifts the balance of the tradeoff made
in Postgres' MVCC. The "old" heap put all transactions on
equal footing by making access to both old and new versions
of rows about the same amount of work. Zheap moves old
versions out of band making current versions very easy to
access by fresh transactions, but old versions more costly
to access by stale ones. It optimizes a little more for
OLTP at the expense of OLAP, and that will be good news for
production apps that run Postgres.

## Striking the right balance (#balance)

The current plan for Zheap's implementation includes
introducing "pluggable storage" that allows a storage
engine to be selected at table-level granularity (`CREATE
TABLE my_table (...) WITH zheap`) [2]. A big reason for
this is because Zheap is too complex of a change to be
introduced wholesale and will need to be eased into the
system, but it will also give users some control over the
tradeoff they want to make.

Finally, I should also note that while Zheap will be hugely
useful for addressing bloat, it will also help with other
Postgres operational problems. The "write amplication" in
indexes (popularized by Uber) becomes far less severe
because tuples are updated in place. Indexes will only need
to be updated if there was a change in a column that they
cover (previously all indexes needed to be updated for
every change). See [the slides][zheapslides] from a recent
talk on the subject at PGConf EU for complete details.

[1] Zheap is still under very active development and is not
yet slated for any upcoming Postgres release.

[2] Details on pluggable storage in Andres' [slides
here][pluggable].

[cte]: https://www.postgresql.org/docs/current/static/queries-with.html
[fivesec]: https://apple.github.io/foundationdb/known-limitations.html
[mvcc]: https://en.wikipedia.org/wiki/Multiversion_concurrency_control
[pluggable]: http://anarazel.de/talks/2018-10-25-pgconfeu-pluggable-storage/pluggable.pdf
[vacuum]: https://www.postgresql.org/docs/current/static/sql-vacuum.html
[zheap]: https://github.com/EnterpriseDB/zheap
[zheapslides]: https://www.postgresql.eu/events/pgconfeu2018/sessions/session/2104/slides/93/zheap-a-new-storage-format-postgresql.pdf
