---
title: Postgres & Portugal
published_at: 2018-10-22T03:04:37Z
---

I started writing this latest edition of _Passages & Glass_
taking on a from Porto back to Lisbon to catch a flight out
of Portugal. I was supposed to be on my way to Braga at the
time, but a last-second rail strike was about to cancel all
service for the next day, so I needed to get back if I
wanted to get on my plane.

Portugal's a lovely country. Its architecture is
splendorous, its streets lively, and its [nata][nata] (an
egg tart dusted with cinnamon), delicious. It's reminscent
of its contemporaries in Europe, but with a lower price
tag. The weather in October was just about perfect -- warm,
but not too hot, and with the occasional morning rain for
to refresh its cities.

As usual, if you don't want to get this newsletter anymore,
you can [unsubscribe in one easy click][unsubscribe] to
never get it again.

![Trams in Lisbon](/assets/passages/004-portugal/trams@2x.jpg)

## Lisbon (#lisbon)

Lisbon is a city of hills, and in that respect it felt like
I hadn't left San Francisco. To help people navigate them,
over the years the city's built a number of elevators (the
most famous being the Elevador de Santa Justa, which has a
regular lineup around the block) and funiculars, which
reminded me of Norway and Japan.

Lisbon's signature image is that of a yellow streetcar
making its way uphill along a narrow street. Over the years
you've probably seen it more than a few times in random
photos on the wall like in this [Edelvik from IKEA][ikea],
even if you didn't know at the time where it was taken. The
streetcars are vintage cars are still in regular operation
and there's no need to go look for them -- you see them
running around _everywhere_ in the inner city. Their
traditional line is the 28, going from Martim Moniz to
Campo Ourique.

![The Elevador de Santa Justa](/assets/passages/004-portugal/lisbon-elevador@2x.jpg)

Claire and Maciek had been spending part of a well-deserved
sabbatical living in Lisbon, and I was lucky enough to have
them show me around the city when I arrived. Amongst other
things, we stopped at open mic night at a bar where the
staff knew them by name, and the most authentic Japan-style
izakaya that I'd been to outside of Japan.

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
implementation of multi-version concurrency control (MVCC),
which isolates the results of concurrent transactions from
each other, and guarantees the ability to roll them back
when necessary.

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
around. They're invisible to practically everything, but
the long-running query might still need them, so they can't
be permanently removed. Having to visit these rows over and
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
introduces an _undo log_, an idea inspired by the design of
databases like MySQL and Oracle.

Instead of old row versions staying in-line in the heap,
they're replaced with the new version and moved to the undo
log, which exists in a specially-reserved section of each
heap page. When a transaction needs an old version of the
row, it follows history back through the undo log until it
finds it. Likewise, when a transaction rolls back, the
now-invalidated version in the heap is replaced by one from
the undo log.

The practical effect is that young transactions get to deal
with mostly fresh data. Even if a heap contains a lot of
history, it's kept out-of-band and most transactions won't
have to look through it, saving themselves work and keeping
things fast. There's also the possibility that it could
eliminate (or at least vastly reduce the necessity of)
`VACUUM`, which would go a long way towards making the
database's performance more deterministic.

By aiming to explain Zheap in three paragraphs I've glossed
over about 1,000 subtleties of its implementation. Amit
Kapila's slides (the project's development lead) on the
subject offer [a wealth of more detailed
information][zheap].

## Pluggable storage (#pluggable-storage)

Changing a storage engine in a database as widely deployed
as Postgres isn't like Indiana Jones swapping out the idol
with one hand for a replacement in the other. Even with a
complete implementation and lengthy testing phase, there'd
still be a substantial risk of regression *somewhere*
that'd be difficult to ever fully address. Zheap also
changes the tradeoffs made by MVVC -- while workloads
requiring short-lived transactions will get faster,
rollbacks and access to historical data gets more costly as
the new system has to travel back in time by applying
changes from undo. If the engine was changed out wholesale,
some applications would slow down.

To mitigate the risk involved in its introduction, Postgres
will be getting a new pluggable storage system, and Zheap
will be its first alternative engine. A new layer of
abstraction called an "table access manager" comes in
between the executor and heap for which both Zheap and the
traditional heap get separate implementations. When
creating a table, the underlying engine will be selectable
at table granularity using a new `WITH` syntax:

``` sql
CREATE TABLE account (...) WITH zheap;
```

Way before programming languages ever got interfaces AKA
traits AKA protocols, you could emulate one with a struct
of function pointers. C never got any of the former, so
that's a struct of function pointers it is:

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

Table access managers like Zheap instantiate the struct and
assign its pointers to their own implementations.

Andres' slides on the subject [are here][pluggable].

## On the web: fast defaults and connection counts (#articles)

With the release of Postgres 11, I wrote a few words on how
the new version will be able to [add columns with default
values _instantly_](/postgres-default). Previously an
exclusive table lock needed to be held while a value was
written for every existing row -- enough to sink a
production system and cause a sizable outage. There were
workarounds for adding new `NOT NULL` columns, but they
were so complicated and exhausting that most of us didn't
bother.

Speaking of operational problems in Postgres, I also wrote
about [managing connections](/postgres-connections). The
database's relatively modest connection limits (most cloud
providers limit them to 500 even on the largest instances)
makes running out of them another often seen pitfall. This
writeup talks about why connections are limited and how to
manage them efficiently with connection pools and minimum
viable checkouts.

## To Porto and back again (#porto)

After the conference I jumped on a train north. My first
stop was Coimbra, a picturesque city built onto and
surrounding a single large hill next to the Mondago River.
It's the same site as the ancient Roman city of Aeminium,
and the city still has a few artifacts from that era
including a beautifully-preserved aqueduct.

![Coimbra's aqueduct](/assets/passages/004-portugal/coimbra-aqueduct@2x.jpg)

The city is well-known for its university, which is the
oldest in the Portuguese-speaking world, founded circa
1290. Built at the city's highest point, it feels like a
plateau on top of the world, with some of the best views in
town.

![Coimbra's university](/assets/passages/004-portugal/coimbra-university@2x.jpg)

![Looking out over Coimbra](/assets/passages/004-portugal/coimbra-view@2x.jpg)

The university's crown jewel is the [Biblioteca
Joanina][joanina], which towers high above the city. It's
famed for being a classic example of unique Baroque
architecture ... but also for its unusual methods of pest
control. The library is home to a small colony of
pipistrelle bats who emerge at night to feed on the insects
that would otherwise be harmful to preserving the
institution's ancient books. In turn, human workers cover
furniture with large sheets of leather before going home to
preserve the library itself from corrosive bat guano.

The resident bats fascinate people -- maybe a little too
much. Visitors to the library tend to become so fixated on
this one point of interest that after it's mentioned, it's
hard for them to think about anything else, and the locals
are tired of it. From a WSJ article [1] on the subject:

> Catarina Freire, a Joanina guide for 16 years, said she
> doesn’t mention bats until the end of her tours, to avoid
> spending the whole time answering questions about them.
>
> “Sometimes I think to myself: Enough of the bats!” she
> said. “They should be a detail in the midst of this
> splendorous temple of knowledge.”

![The library of Coimbra](/assets/passages/004-portugal/coimbra-library@2x.jpg)

### Porto & port (#porto)

I continued north from Coimbra to Porto. Portugal is
cholk-full of beautiful cities, but Porto takes the award
for being the most so. Built around the Douro River's
estuary (where it dumps into the Atlantic Ocean), Porto
features Portugal's usual hilly landscapes and charming
architecture. It also has a wide river bordered by high
cliffs that provide incredible views, impressive bridges to
span it, and a tasting room for every brand of port you've
ever heard of.

And speaking of port, I went in confused about the
technicalities of where it comes from and what exports are
allowed to have that name. Here are a few facts on port in
a form that's as easy to digest (*-tif?*) as you're likely
to find:

* Porto, or _Oporto_, derives its name from _o_ ("the") +
  _porto_ ("port, harbour"). Literally because it's a big
  port city.

* Port, the fortified wine, is traditionally produced in
  Portugal's Douro Valley, about 100 kilometers east of
  Porto. It gets its name from Porto because that's where
  it would brought downriver for aging and export.

* Like cognac or champagne, port falls under the EU's
  protected origin guidelines, but they allow anything out
  of Portugal to be designated port, not just what comes
  out of Porto or the Douro Valley.

* The US doesn't have a system to recognize geographic
  indications, and no one has a registered certification
  mark for port, so around here you may find your ports
  originating from anywhere.

![The Dom Luís in Porto](/assets/passages/004-portugal/porto-bridge@2x.jpg)

The area's port estates are all clustered closely together,
and technically in the municipality of Vila Nova de Gaia,
but given that all it takes is walking across the famous
Dom Luis bridge to get there from Porto, it all feels like
the same city.

While touring a cellar, I was most impressed by the huge
barrels used to reduce surface area during the aging
process. It's hard to tell from the picture, but this one
was half again as tall as I was.  The number in the upper
right is the number of liters in the batch. Thirty-two
_thousand_.

![Port barrel](/assets/passages/004-portugal/port-barrel@2x.jpg)

Parked along Vila Nova de Gaia's shores are [barcos
rabelos][rabelos], flat-bottomed cargo ships that were in
popular use to transport port from the Douro Valley down
along the Douro River. Hydroelectric dams and locks built
in the 50s and 60s put an end to their use, but they're
still a pretty sight from another age.

![Porto](/assets/passages/004-portugal/porto@2x.jpg)

## Coffee houses (#coffee-houses)

You know how you can walk into any Costco and kind of know
where you're going? When I visited Europe for the first
time a decade ago, if I homesick I'd stop by a local
Starbucks. Whether you're in Japan or the Czech Republic,
the menu and interior decor hailed back to any shop you'd
find in the Americas (or elsewhere in the world) with a
comforting familiarity.

I'm a little amused that these days you don't even need
Starbucks anymore -- you can do the same thing with third
wave coffee shops. Here's Fábrica in Lisbon. It's like I
never left SOMA.

![Third wave coffee in Lisbon](/assets/passages/004-portugal/third-wave@2x.jpg)

## Raw denim (#raw-denim)

Now for a final note that doesn't involve anything related
to travel or technology -- sort of the opposite in fact.
I've recently had an unhealthy (for my wallet at least)
fixation on the world of raw denim.

What is that you might ask? Well, almost every pair of
jeans you're likely to buy today will have been through
some kind of [sanforization] process and pre-washed to
shrink it or produce a distressed effect. It's also been
increasingly more common for the very fabric of denim to be
engineered away -- originally pure cotton, jeans are more
likely to contain plastics like spandex with every passing
year.

Raw denim foregoes all of this. It's unwashed, untreated,
and largely untouched from the time it rolls off the
(usually selvedge) loom to when it's sold. Compared to a
normal pair of jeans it feels crispy and stiff, and it's up
to the new owner to break it in.

Next, you might rightly ask, why would you ever want this?
Well there's a reasonable argument that you wouldn't, but
raw denim's most interesting property is how it ages. Over
time its indigo dye fades, most pronounced in places like
where the fabric creases at the knees, or around regular
items in pockets. Every pattern is unique to the jean and
wearer, and often produces quite an appealing. Just take a
look at some of the models from [Blue Owl's fade
museum][fademuseum].

![Japan Blue fade](/assets/passages/004-portugal/jeans1@2x.jpg)

![Momotaro fade](/assets/passages/004-portugal/jeans2@2x.jpg)

There's also an argument to be made for environmentalism --
the industrial processes used in most jeans use water
heavily, and the plastics they contain often degrade
microscopically and end up in our oceans. I won't pretend
those are the main reasons enthusiasts are into raw denim
though -- it's a little like vinyl records or cooking on
cast iron -- there's no defensible objective advantage, but
there's some inherent je-ne-sais-quoi appeal to the art and
romance of a physical craft.

Heddel's has [a good writeup on the subject][rawdenim]. If
podcasts are your thing, the _99% Invisible_-adjacent show
_Articles of Interest_ did a [great episode on
it][articlesofinterest] as well.

If you're in San Francisco, we're fortunate to have the
world's most pre-eminent shop for raw denim in the form of
[Self Edge][selfedge]. This place really is the most
boutique of all boutique shops. Their prices will make your
eyes water, but it's a neat place to visit.

Over the Black Friday weekend I personally bought a pair of
Pure Blue Japan jeans. Buying jeans that were unsanforized,
unshrunk, and with no stretch, I sized up three sizes to an
ostensibly slim fit jean whose fit at the time could only
be described as baggy. I got home and soaked them almost
immediately, and putting them on three days found that
they'd shrunk down to a near-perfect fit.

[1] Wall Street Journal. The bat article is behind a
paywall, but you can get through it by clicking through to
it from Google. Search for: _"The Bats Help Preserve Old
Books But They Drive Librarians, Well, Batty."_

[articlesofinterest]: https://99percentinvisible.org/episode/blue-jeans-articles-of-interest-5/transcript/
[fademuseum]: https://www.blueowl.us/blogs/fade-museum
[ikea]: https://www.ikea.com/pt/en/catalog/products/70420984/
[joanina]: https://en.wikipedia.org/wiki/Biblioteca_Joanina
[nata]: https://en.wikipedia.org/wiki/Pastel_de_nata
[oltp]: https://brandur.org/fragments/olap-oltp-zheap
[pluggable]: http://anarazel.de/talks/2018-10-25-pgconfeu-pluggable-storage/pluggable.pdf
[queues]: https://brandur.org/postgres-queues
[rabelos]: https://en.wikipedia.org/wiki/Rabelo_boat
[rawdenim]: https://www.heddels.com/2011/09/the-essential-raw-denim-breakdown-our-100th-article/
[sanforization]: https://en.wikipedia.org/wiki/Sanforization
[selfedge]: https://www.selfedge.com/
[unsubscribe]: %unsubscribe_url%
[zheap]: https://www.postgresql.eu/events/pgconfeu2018/sessions/session/2104/slides/93/zheap-a-new-storage-format-postgresql.pdf
