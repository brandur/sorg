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
struct AccessManager {
    void CreateTable(...);
}
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

![Port barrel](/assets/passages/004-portugal/port-barrel@2x.jpg)

![Third wave coffee in Lisbon](/assets/passages/004-portugal/third-wave@2x.jpg)

Train strike.

## Port (#port)

A lot of classics.

The number in the upper right is liters. Reduce surface
area.

Rule of thirds?

## Raw denim (#raw-denim)

Fade museum.

[pluggable]: http://anarazel.de/talks/2018-10-25-pgconfeu-pluggable-storage/pluggable.pdf
