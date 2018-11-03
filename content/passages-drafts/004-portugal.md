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
through this and find plenty more closer to the end.

## PGConf EU (#pgconf)

## Bloat (#bloat)

TODO

## Zheap (#zheap)

One of the most exciting (and largest!) developments today
in Postgres is the development of Zheap, a new heap
implementation that promises to make dramatically reduce
the impact of bloat. Zheap follows in the footsteps of
Oracle by introducing the idea of a per-page _undo log_

TOOD

## Pluggable storage (#pluggable-storage)

Changing a storage engine in a database as widely deployed
as Postgres isn't like Indiana Jones swapping out the idol
with one hand for a replacement in the other. Even with a
lengthy testing phase, there's still a substantial risk of
regression *somewhere* that'd be impossible to ever fully
address. Zheap also changes the tradeoffs made by MVVC --
while workloads requiring short-lived transactions will get
faster, access to historical data by old transactions gets
more costly as the system travels back in time by applying
changes in the undo log. If the engine was changed out
wholesale, some people's installations would slow down.

To mitigate the risk involved in its introduction, a new
pluggable storage system will be introduced and Zheap will
come in as an alternate engine that's part of it. A new
layer of abstraction called an "table access manager" comes
in between the executor and heap for which both Zheap and
the traditional heap get their own implementations. The
underlying storage will be selectable at table granularity
using a new `WITH` syntax:

``` sql
CREATE TABLE account (...) WITH zheap;
```

The table access manager is the C equivalent of an
interface, a struct of function pointers:

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

Andres gave a great talk on all of this. See his [slides
here][pluggable].

TODO: Verify all of this.

## On the web: fast defaults and connection counts (#articles)

With the release of Postgres 11 right around the corner, I
wrote a piece on how the new version will be able to [add
columns with default values _quickly_](/postgres-default).
Previously an exclusive table lock needed to be held while
a value was written for every existing row which was enough
to tank a production system, and made adding any new `NOT
NULL` so much effort that most of us didn't bother.

Speaking of operational problems, I also put pen to paper
on [managing connections](/postgres-connections). The
relatively modest connection limits in Postgres (most cloud
providers limit them to 500 even on the largest instances)
makes running out of them one of the first major problems
with Postgres that users encounter. The article above
covers why connections are limited and how to manage them
efficiently with connection pools.

## To Porto and back again (#porto)

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

[fademuseum]: https://www.blueowl.us/blogs/fade-museum
[pluggable]: http://anarazel.de/talks/2018-10-25-pgconfeu-pluggable-storage/pluggable.pdf
[rabelos]: https://en.wikipedia.org/wiki/Rabelo_boat
[rawdenim]: https://www.heddels.com/2011/09/the-essential-raw-denim-breakdown-our-100th-article/
