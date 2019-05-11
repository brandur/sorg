+++
hook = "How Zheap shifts Postgres towards optimization for OLTP by making access to new row versions cheaper and old row versions more expensive."
location = "Coimbra"
published_at = 2018-10-28T08:33:07Z
title = "OLAP, OLTP, Zheap, and tradeoffs in Postgres MVCC"
+++

One of the most exciting developments in the world of
Postgres right now is work on the new [Zheap storage
engine][zheap] [1]. Recently, some sessions at a Postgres
conference helped crystallize it and the optimizations it
aims to make. Zheap promises to be a significant
advancement in the operational aspects of Postgres, even if
there's likely to be a tradeoff that will impact some uses
of the database.

You'll have to forgive the roundabout approach to talking
about Zheap here, but I thought it'd be interesting to look
at it in the broader of context of databases in general,
and what we use them for.

## Analytics versus production (#olap-oltp)

You may have heard database administrators or enthusiasts
talk about OLAP and OLTP. They're wordy terms for the
uninitiated, but like many overwrought acronyms, represent
relatively simple ideas, and are useful as unambiguous
names for two important concepts:

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
Inserts, updates, deletes, and simple (well-indexed)
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
vacuum gets a chance to run (I've previously gone into the
details of [its inner workings here](/postgres-atomicity)).

It's an elegant implementation -- easy to reason about, and
making all row versions readily accessible to the
transactions that need them, but it has its downsides.
Relations may become "bloated" when an old transaction is
forcing many old versions to be kept around, leading to
greatly increased table size. Bloated tables are slower to
use because transactions need to iterate through dead rows
to check whether they're visible or not. Those rows can't
be cleaned out until all transactions that could
potentially see them finish.

We could say that Postgres has optimized for OLAP at the
expense of OLTP (even if it was never a conscious
decision). Long-lived transactions have easy access to
contemporaneous row versions, but that same feature
degrades the performance of current, short-lived
transactions.

## Moving the needle: Zheap (#zheap)

Zheap is a new storage engine that rethinks how MVCC in
Postgres works. With Zheap, rows are often be updated in
place, with old versions replaced by new ones. Instead of
sticking around in the heap, old versions are moved to an
"undo log" in the current page which acts a historical
record. The idea is inspired by other databases that
already use a similar technique, like Oracle.

If a transaction aborts, old versions from the undo log are
applied until the right version is reached. Similarly, old
transactions that need an old version follow the undo log
back until they find the right one. Like the current heap,
old versions need only be kept as long as they're needed,
and their space is reclaimed as old transactions come to
an end.

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
every change). There's even talk about Zheap potentially
eliminating the need for vacuum because of the way it could
lazily reclaim slots in the undo log. See [the
slides][zheapslides] from a recent talk on the new engine
for more complete details.

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
