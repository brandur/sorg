+++
published_at = 2018-10-22T03:04:37Z
title = "Postgres & Portugal"
+++

I started writing this edition of _Passages & Glass_ on the
way back to Lisbon from Porto, right after a failed attempt
to go up north to Braga. The trip had been going smoothly
so far, but a last-second strike for the next day from its
national rail Comboios de Portugal would have left me
stranded up there, and unable to get back to Portugal in
time for my flight home.

I'm _finishing_ writing this edition of _Passages & Glass_
three months later from San Francisco, and somehow that
quarter-year feels like just the blink of an eye. Time
movies quickly.

Portugal's a lovely country. Its architecture is
impressive, its streets lively, and its [nata][nata] (an
egg tart dusted with cinnamon), delicious. It's reminiscent
of its contemporaries in Europe, but with a lower price
tag. The weather in October was perfect -- warm, but not
too hot, and with the occasional morning rain for to
revitalize its streets.

It's been so long since I sent something to this list that
you might not even remember signing up for it. As usual, if
you don't want to it anymore, you can [unsubscribe in one
easy click][unsubscribe].

![Trams in Lisbon](/assets/images/passages/004-portugal/trams@2x.jpg)

## Lisbon (#lisbon)

Lisbon is a city of hills, and in that way it felt like
I hadn't left San Francisco. To help people navigate them,
over the years the city's built a number of elevators (the
most famous being the Elevador de Santa Justa, which has a
regular lineup around the block) and hill-climbing
funiculars, reminding me of Norway and Japan.

Lisbon's signature image is a yellow streetcar making its
way uphill along a narrow street. Over the years you've
probably seen it around in random wall hangings like in
this [Edelvik from IKEA][ikea], even if you didn't know at
the time which city it was. Those trams are vintage cars
which are still in regular operation, and once you're in
Lisbon there's no need to go look for them -- you see them
running around _everywhere_ in the inner city. Their
traditional line is the 28, going from Martim Moniz to
Campo Ourique.

![The Elevador de Santa Justa](/assets/images/passages/004-portugal/lisbon-elevador@2x.jpg)

Claire and Maciek had been spending part of a well-deserved
sabbatical living in Lisbon, and I was lucky enough to have
them show me around the city when I arrived. Amongst other
things, we stopped at open mic night at a bar where the
staff knew them by name, and a Japanese izakaya which was
amazingly authentic given that it was on the other side of
the world from its mother country.

![Lisbon at night](/assets/images/passages/004-portugal/lisbon-night@2x.jpg)

## Postgres (#postgres)

And now a not-so-brief interlude to talk about Postgres. If
you want to see more photos of Portugal, you can skip
through this and find more of that closer to the end.

I was in Lisbon to attend PGConf EU, a conference that
rotates through Europe, and which landed in in Portugal for
2018. I've been using and writing about Postgres for years
now, but this was my first time attending a Postgres event
bigger than a local meetup. Talks covered a variety of
topics, but the center of gravity was heap bloat and some
of the projects in flight to fight it.

### Pitfalls in production (#production)

Bloat is the Achilles heel of production Postgres,
especially where it's used for [OLTP][oltp] (many fast,
small transactions, as opposed to OLAP, which is
analytics). It's an inherent artifact of Postgres'
implementation of multi-version concurrency control (MVCC),
which isolates concurrent transactions from each other, and
guarantees the ability to roll them back when necessary.

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
the long-running query could conceivably access them, so
they can't be permanently removed. Having to visit these
rows over and over degrades the performance of current
transactions in the system, many of which need to be fast
to keep applications healthy. I have [my own story about
bloat][queues] from my time at Heroku involving a
database-based job queue.

### Zheap and undo logs (#zheap)

One of the most exciting (and as far as committed
resources, maybe one of the largest ever) developments
today in Postgres is the development of Zheap, a new heap
implementation that promises to dramatically reduce the
impact of bloat. Zheap introduces an _undo log_, an idea
inspired by the design of databases like MySQL and Oracle.

Instead of old row versions kept in-line with live ones in
the heap, they're replaced with the new version and moved
to the undo log, which exists in a reserved segment of each
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

### Pluggable storage (#pluggable-storage)

Changing a storage engine in a database as widely deployed
as Postgres isn't as easy Indiana swapping out an idol with
one hand for a replacement in the other. Even with a
complete implementation and lengthy testing phase, there'd
still be a substantial risk of regression *somewhere*
that'd be difficult to ever fully address. Zheap also
changes the tradeoffs made by MVVC -- while workloads
requiring short-lived transactions will get faster,
rollbacks and access to historical data gets more costly as
the new system has to travel back in time by applying
changes from undo. If the engine was changed out wholesale,
some uses of it would slow down.

To mitigate the risk involved in its introduction, Postgres
will bring it in with a new pluggable storage system. A new
layer of abstraction called an "table access manager" gets
added between the executor and heap, and both Zheap and the
traditional heap get separate implementations. Existing
tables stay on the traditional heap. The engine can be
selected for new tables using a new `WITH` syntax:

``` sql
CREATE TABLE account (...) WITH zheap;
```

Way before programming languages ever got interfaces (AKA
traits AKA protocols), you could emulate one with a struct
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
were so exhausting that most of us didn't bother.

And speaking of operational problems in Postgres, I also
wrote about [managing connections](/postgres-connections).
The database's relatively modest connection limits (most
cloud providers limit them to 500 even on the largest
instances) makes running out of them another often seen
pitfall. This writeup talks about why connections are
limited and how to manage them efficiently with connection
pools and minimum viable checkouts.

## To Porto and back again (#porto)

After the conference I traveled north by train. My first
stop was Coimbra, a picturesque city built onto and
surrounding a single large hill next to the Mondago River.
It's the same site as the ancient Roman city of Aeminium,
and the city still has a few artifacts from that era
including a beautifully-preserved aqueduct.

![Coimbra's aqueduct](/assets/images/passages/004-portugal/coimbra-aqueduct@2x.jpg)

The city is well-known for its university, which is the
oldest in the Portuguese-speaking world, founded circa
1290. Built at the city's highest point, when you visit it
you feel like you're standing on a plateau on top of the
world, giving you a vantage over the wider area. Not too
many schools have the best views in town.

![Coimbra's university](/assets/images/passages/004-portugal/coimbra-university@2x.jpg)

![Looking out over Coimbra](/assets/images/passages/004-portugal/coimbra-view@2x.jpg)

The university's crown jewel is the [Biblioteca
Joanina][joanina], which towers high above the city. It's
famed for being a classic example of unique Baroque
architecture -- but also for its unusual methods of pest
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

![The library of Coimbra](/assets/images/passages/004-portugal/coimbra-library@2x.jpg)

### Porto & port (#porto)

I continued north from Coimbra to Porto. Portugal is
cholk-full of beautiful cities, but Porto is the most so.
Built around the Douro River's estuary (where it dumps into
the Atlantic Ocean), Porto features Portugal's usual hilly
landscapes and charming architecture. It also has a wide
river bordered by high cliffs that provide incredible
views, impressive bridges to span it, and a tasting room
for every brand of port you've ever heard of.

And speaking of port, I went in confused about the
technicalities of where it comes from and what exports are
allowed to be called "port". Here are a few facts on port
in a form that's as easy to digest (*-tif?*) as I could
make them:

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

![The Dom Luís in Porto](/assets/images/passages/004-portugal/porto-bridge@2x.jpg)

The area's port estates are all clustered closely together,
and technically in the municipality of Vila Nova de Gaia,
but given that all it takes is a walk across a bridge (the
famous Dom Luis) to get there from Porto, it feels like the
same city.

While touring a cellar, I was most impressed by the huge
barrels used to reduce surface area during the aging
process. It's hard to tell from the picture, but this one
was half again as tall as I was.  The number in the upper
right is the number of liters in the batch. Over thirty-two
_thousand_.

![Port barrel](/assets/images/passages/004-portugal/port-barrel@2x.jpg)

Parked along Vila Nova de Gaia's shores are [barcos
rabelos][rabelos], flat-bottomed cargo ships that were in
popular use to transport port from the Douro Valley down
along the Douro River. Hydroelectric dams and locks built
in the 50s and 60s put an end to their use, but they're
still a pretty sight from another age.

![Porto](/assets/images/passages/004-portugal/porto@2x.jpg)

## A wave of third wave (#third-wave)

You know how you can walk into any Costco and kind of know
where you're going? When I visited Europe for the first
time a decade ago, when I got homesick I'd stop by a local
Starbucks. Whether you're in Japan or the Czech Republic,
the menu and interior decor hailed back to any shop you'd
find in the Americas (or elsewhere in the world) with a
comforting familiarity.

I'm a little amused that these days you don't even need
Starbucks anymore -- you can do the same thing with third
wave coffee shops which have made their way out into the
wider world. Here's Fábrica in Lisbon. Polished concrete
floors, furniture built exclusively out of reclaimed wood,
Chemex/V60/Hario accessories decorating the walls. It's
like I never left SOMA.

![Third wave coffee in Lisbon](/assets/images/passages/004-portugal/third-wave@2x.jpg)

## Raw denim and romance in products (#raw-denim)

I've recently picked up an unhealthy (for my wallet)
fixation on the world of raw denim.

What is that you might ask? Well, almost every pair of
jeans you're likely to buy today will have been through a
[sanforization] process and pre-washed to shrink them and
produce a distressed effect. It's also been increasingly
more common for the very fabric of denim to be engineered
away -- originally pure cotton, your jeans are more likely
to contain plastics like spandex with every passing year.

Raw denim foregoes all these modern techniques. It's
unwashed, untreated, and largely untouched from the time it
rolls off the (usually selvedge) loom to when it's sold.
Compared to a normal pair of jeans it feels crispy and
stiff to the touch. The fabric will eventually soften to
feel more like the jeans we're used to, but only after
months of wear, and it's up to the new owner to break it
in.

Next, you might rightly ask, why would you ever want this?
Well there's a good argument that you wouldn't, but raw
denim's most interesting property is how it ages. Over time
the indigo dye fades, and that fading is most pronounced in
places like where the fabric creases at the knees, or
around regular items in pockets. Every pattern is unique to
the jean and wearer, and often produces some beautiful
effects. Just take a look at some of the models from [Blue
Owl's fade museum][fademuseum].

![Japan Blue fade](/assets/images/passages/004-portugal/jeans1@2x.jpg)

![Momotaro fade](/assets/images/passages/004-portugal/jeans2@2x.jpg)

There's also a case to be made for environmentalism -- the
industrial processes used in most jeans use water heavily,
and the plastics they contain often degrade to microscopic
particulates and end up in places like our oceans. I won't
pretend those are the main reasons enthusiasts are into raw
denim though -- it's a little like vinyl records or cooking
on cast iron -- there's no defensible objective advantage,
but there's some inherent je-ne-sais-quoi appeal to the art
and romance of a physical craft.

Heddel's has [a good writeup on the subject][rawdenim]. If
podcasts are your thing, the _99% Invisible_-adjacent show
_Articles of Interest_ did a [great episode on
it][articlesofinterest] as well.

If you're in San Francisco, you can visit the world's most
pre-eminent shop for raw denim in the form of [Self
Edge][selfedge]. This place really is the most boutique of
all boutique shops. Their prices will make your eyes water,
but it's a neat place to visit.

Over the Black Friday weekend I bought a pair of Pure Blue
Japan jeans. Buying jeans that were unsanforized, unshrunk,
and with no stretch, I sized up three sizes to a jean that
was described as "slim fit", but whose generous dimensions
would have been right at home on a skateboarder in the 90s.
I got home and soaked them almost immediately. Three days
later when I put them on, they'd shrunk to a near-perfect
fit.

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
