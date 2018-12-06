---
title: Postgres & Portugal
published_at: 2018-10-22T03:04:37Z
---

![Trams in Lisbon](/assets/passages/004-portugal/trams@2x.jpg)

Writing from a train back down the coast to Lisbon.

![Lisbon at night](/assets/passages/004-portugal/lisbon-night@2x.jpg)

## Postgres (#postgres)

Let's take a not-so-brief interlude to talk about Postgres.
If you want to see more photos of Portugal, you can skip
through this and find more of that closer to the end.

I was lucky enough this year to attend PGConf EU, a
conference that rotates through Europe, and which was held
in Portugal for 2018. I've been using and writing about
Postgres for years now, but this was my first time
attending a Postgres-related event bigger than a local
meetup. Talks covered a variety of topics, but the center
of gravity was bloat and a few of the major projects in
flight to fight it.

## That old foe: bloat (#bloat)

Bloat is the Achilles heel of Postgres in production,
especially where it's used for [OLTP][oltp] (many fast,
small transactions, as opposed to OLAP, which is
analytics). Bloat is an inherent artifact of Postgres'
MVCC (multi-version concurrency control) implementation,
which isolates the results of concurrent transactions from
each other, and guarantees the ability to roll back a
transaction when necessary.

Every row that's visible to _any_ transaction that's still
running in the system has to be retained in the heap (the
name for a table's physical storage), even if it's been
subsequently deleted. When a transaction is searching the
heap for results, it checks every row's visibility to make
sure that it's still relevant before including it. Rows
that are deleted but retained because they're still visible
_somewhere_ are "bloat".

The not-uncommon degenerate scenario is for a very old
query (often an analytical one running for hours or even
days) to force a huge number of deleted rows to be kept
around. They're invisible to almost everything, but the
long-running query might still need them, so they can't be
permanently removed. Having to visit these rows over and
over degrades the performance of current transactions in
the system, many of which need to be fast to keep
applications healthy. I have [my own story about
bloat][queues] from my time at Heroku involving a
database-based job queue.

## A new contender: Zheap (#zheap)

One of the most exciting (and resource-wise, maybe one of
the largest ever) developments today in Postgres is the
development of Zheap, a new heap implementation that
promises to dramatically reduce the impact of bloat. Zheap
introduces an _undo log_, an idea inspired by the MVCC
implementation of databases like MySQL and Oracle.

Instead of old row versions staying in-line in the heap,
they're replaced with the new version and moved to the undo
log, which exists in a specially-reserved section of each
page. When an old transaction needs an old version of the
row, it follows history back through the undo log until it
finds it. Likewise, when a transaction rolls back, the
now-invalidated version in the heap is replaced by one from
the undo log.

The practical effect is that young transactions get to deal
with mostly fresh data. Even if a heap contains a large
quantity of history, it's kept out-of-band and most
transactions won't have to look at it, saving themselves
work and staying performant. There's also the possibility
that it could eliminate (or at least vastly reduce the
necessity of) `VACUUM`.

By aiming to explain Zheap in three paragraphs I've glossed
over about 1,000 subtleties of its implementation. Amit
Kapila's slides (the project's development lead) on the
subject offer [a wealth of more detailed
information][zheap].

## Pluggable storage (#pluggable-storage)

Changing a storage engine in a database as widely deployed
as Postgres isn't like Indiana Jones swapping out the idol
with one hand for a replacement in the other. Even with a
lengthy testing phase, there's still a substantial risk of
regression *somewhere* that'd be impossible to ever fully
address. Zheap also changes the tradeoffs made by MVVC --
while workloads requiring short-lived transactions will get
faster, rollbacks and access to historical data by old
transactions gets more costly as the new system has to
travel back in time by applying changes from undo. If the
engine was changed out wholesale, some applications would
slow down.

To mitigate the risk involved in its introduction, Postgres
will be getting a new pluggable storage system, and Zheap
will be its first alternative engine. A new layer of
abstraction called an "table access manager" comes in
between the executor and heap for which both Zheap and the
traditional heap get their own implementations. The
underlying storage will be selectable at table granularity
using a new `WITH` syntax:

``` sql
CREATE TABLE account (...) WITH zheap;
```

The table access manager is the C equivalent of an
interface, a struct of function pointers that are invoked
for heap-related functions:

``` c
typedef struct TableAmRoutine
{
    ...

    TupleInsert_function            tuple_insert;
    TupleInsertSpeculative_function tuple_insert_speculative;
    TupleUpdate_function            tuple_update;
    TupleDelete_function            tuple_delete;
    MultiInsert_function            multi_insert;
    TupleLock_function              tuple_lock;

    ...
} TableAmRoutine;
```

Andres' slides on the subject [are here][pluggable].

## On the web: fast defaults and connection counts (#articles)

With the release of Postgres 11, I wrote a few words on how
the new version will be able to [add columns with default
values _quickly_](/postgres-default). Previously an
exclusive table lock needed to be held while a value was
written for every existing row which was enough to sink a
production system, and made adding any new `NOT NULL` so
much effort that most of us didn't bother.

Speaking of operational problems in Postgres, I also wrote
about [managing connections](/postgres-connections). The
database's relatively modest connection limits (most cloud
providers limit them to 500 even on the largest instances)
makes running out of them another frequent pitfall. This
article talks about why connections are limited and how to
manage them efficiently with connection pools and minimum
viable checkouts.

## To Porto and back again (#porto)

[Biblioteca Joanina][joanina]

![The library of Coimbra](/assets/passages/004-portugal/coimbra-library@2x.jpg)

![Porto](/assets/passages/004-portugal/porto@2x.jpg)

![The Dom Luís in Porto](/assets/passages/004-portugal/porto-bridge@2x.jpg)

![Port barrel](/assets/passages/004-portugal/port-barrel@2x.jpg)

![Third wave coffee in Lisbon](/assets/passages/004-portugal/third-wave@2x.jpg)

Train strike.

## Port (#port)

A lot of classics.

Transported downriver in [barcos rabelos][rabelos], but the
construction of hydroelectric damns in the 50s and 60s
ended the practice, but some of the old boats are still on
display in central Porto.

> It is flat-bottomed, with a shallow draught, which was necessary to navigate the often shallow fast-flowing waters of the upper Douro prior to the construction of dams and locks from 1968 onwards.

The number in the upper right is liters. Reduce surface
area.

Rule of thirds?

## Raw denim (#raw-denim)

> Raw denim (aka dry denim) is simply denim fabric that remains unwashed, untreated, and virtually untouched from when it rolls off the loom to when it is sold to you. It’s denim in its purest form.

> Raw denim usually has a crispy and stiff feel and easily leaves traces of its indigo dye behind when it rubs against another surface–even your hands (this phenomenon is called crocking). Be careful what you rub up against when wearing a new pair of raw denim jeans, you might leave a bit of blue behind.

> So why go through all this hassle just for a new pair of jeans? One of the biggest benefits of raw denim, and the indigo loss, is that they develop and age based on what you do in them and to them. Every mile you walk, every scrape on the concrete, every item you keep regularly in your pocket leaves its mark. The dark indigo dye slowly begins to chip away revealing the light electric blue and eventually the white cotton core of the denim yarns the more you wear them. What’s left is a wholly unique garment that was formed and faded to you and you alone.

> It takes an awful lot of water to grow enough cotton for a pair of jeans, but washing and distressing them takes even more, an average of 42 liters per jean. By buying raw, none of that water needs to go to waste. It also doesn’t expose workers to the harmful chemicals often used to distress and wash denim.

![Japan Blue fade](/assets/passages/004-portugal/jeans1@2x.jpg)

![Momotaro fade](/assets/passages/004-portugal/jeans2@2x.jpg)

[Blue Owl's fade museum][fademuseum].

[Heddels article on raw denim][rawdenim].

If podcasts are your thing, the _99% Invisible_-adjacent
show _Articles of Interest_ did a [great episode on the
subject][articlesofinterest] recently.

If you happen to be in San Francisco, we're fortunate to
have what's probably the world's most pre-eminent shop for
raw denim in the [Self Edge][selfedge]. Their prices will
make your eyes water, but it's a neat place to visit.

[articlesofinterest]: https://99percentinvisible.org/episode/blue-jeans-articles-of-interest-5/transcript/
[fademuseum]: https://www.blueowl.us/blogs/fade-museum
[joanina]: https://en.wikipedia.org/wiki/Biblioteca_Joanina
[oltp]: https://brandur.org/fragments/olap-oltp-zheap
[pluggable]: http://anarazel.de/talks/2018-10-25-pgconfeu-pluggable-storage/pluggable.pdf
[queues]: https://brandur.org/postgres-queues
[rabelos]: https://en.wikipedia.org/wiki/Rabelo_boat
[rawdenim]: https://www.heddels.com/2011/09/the-essential-raw-denim-breakdown-our-100th-article/
[selfedge]: https://www.selfedge.com/
[zheap]: https://www.postgresql.eu/events/pgconfeu2018/sessions/session/2104/slides/93/zheap-a-new-storage-format-postgresql.pdf
