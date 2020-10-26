+++
image_alt = "UC Berkeley, the birthplace of Postgres"
image_url = "/assets/images/nanoglyphs/016-postgres-13/uc-berkeley-clock-tower@2x.jpg"
published_at = 2020-10-26T15:19:48Z
title = "Postgres 13, Shrunken Indexes, Planet Earth"
+++

Readers --

I hope all is well. For those who don't recognize it, the photo you're looking at is the clock tower at UC Berkeley, birthplace of Postgres in the mid 80s. I took it a few months ago while walking down to the university from its surrounding hillsides, where you can find breathtaking views of the campus and the town, and which are surprisingly steep for a little urban hike.

This is the third issue of _Nanoglyph_ sent this month. Nothing to brag about, but it's my best ever run by a good margin. I've had a slow dawning of realization that a serious defect in my publication schedule is lack of anything even _fuzzily_ resembling routine. I _intend_ to write something down every day, but in practice, get up, mull around aimlessly for an hour with calisthenics, breakfast, and dishes, end up reading miscellaneous internet esoterica, and by the time I'm ready to do anything, it's time to go to work. Evenings are terribly unproductive. The unenviable aggregate body of literature I produce on an average day amounts to about 5,000 words worth of Slack messages and JIRA comments.

---

With an eye towards fixing it, I started looking around for some inspiration, and found this nice [digest of daily routines](https://jamesclear.com/daily-routines-writers) from well-known authors. From Haruki "Man of Steel" Murakami:

> When I’m in writing mode for a novel, I get up at four a.m. and work for five to six hours. In the afternoon, I run for ten kilometers or swim for fifteen hundred meters (or do both), then I read a bit and listen to some music. I go to bed at nine p.m.
>
> I keep to this routine every day without variation. The repetition itself becomes the important thing; it’s a form of mesmerism. I mesmerize myself to reach a deeper state of mind.

Hemingway, the most efficient wordsmith ever to grace the Earth, gives us his typical eloquence (and I agree with him about mornings):

> When I am working on a book or a story I write every morning as soon after first light as possible. There is no one to disturb you and it is cool or cold and you come to your work and warm as you write. You read what you have written and, as you always stop when you know what is going to happen next, you go on from there.

E.B. White (_Charlotte's Web_) reminds us that it's not always clean or easy, and you just have to do it anyway:

> In consequence, the members of my household never pay the slightest attention to my being a writing man — they make all the noise and fuss they want to. If I get sick of it, I have places I can go. A writer who waits for ideal conditions under which to work will die without putting a word on paper.

---

## Lucky number 13 (#postgres-13)

Postgres continues its annual cadence of major version pushes by giving us Postgres 13. The database is already so mature that every year I think “surely this must be the last set of major features” -- I mean, there _must_ a point where it’s so refined that all the big wins are behind it, and we’re left with just small stuff. But, like clockwork, every year I’m wrong, with a major innovation coming out of left field and reminding us why Postgres is the established leader in the database world. Like with Chrome and Firefox [1], AMD and Intel, or Python and Ruby, there was a time when Postgres could've been said to be in lockstep for features and mindshare with similar products like MySQL and Oracle, but those days are over, Postgres having pulled ahead into a comfortable lead.

I’m covering some favorite highlights, but see the [full release notes](https://www.postgresql.org/docs/release/13.0/) because I’m not even hitting all the big ones, some of which require a little too much backstory for this newsletter. As with every Postgres release there are dozens of improvements of all sizes from a 100+ contributors. Postgres demonstrates by example how an open-source product can be far better than a commercial one, as almost none of these major wins would have floated to the top of a private company’s stack ranking, let alone the smaller ones.

### Index deduplication (#index-deduplication)

Postgres 13's headliner is [deduplication in B-tree indexes](https://git.postgresql.org/gitweb/?p=postgresql.git;a=commit;h=0d861bbb702f8aa05c2a4e3f1650e7e8df8c8c27), which allows duplicate values in an index to share space and thereby vastly reduce index sizes where repeat values exist.

Claiming specific numbers on space savings is fraught because it depends so much on the characteristics of the fields being indexed and on workload, but the patch suggests a sample index where each unique value is duplicated 10-15 times might be reduced in size by 2.5 to 4x (meaning an index occupying 100 GB would become 25 GB). This is a very big deal because in the real world Postgres tables tend to have a lot of indexes, many of which are heavy on duplicates.

The implementation of Postgres’ MVCC makes the change even more important than it seems on the surface. Rows being deleted or changed will have duplicated index values pointing to different versions of the same logical record, and even when those duplicates are no longer needed, the system would generally have to wait for a vacuum to reclaim them, before which there was ample opportunity for page splits and bloat. Databases where updates lead to many simultaneously living versions of the same row will see major benefits from this change.

#### Why it’s important (#important)

Postgres users with large installations will already know this, but one should underestimate the importance of index size. They start small, but can get so big that they cause their own existential crises.

For the last six months I’ve been dealing with a problem at work involving data size and retention. In our one particular problematic Mongo collection, we have _just indexes_ that are many 10s of TBs in size. These large indexes are a somewhat inherent byproduct of the underlying indexed data being large, but its not _as much_ larger as you'd think -- the data itself is only about an order of magnitude larger than the indexes.

Its indexes are now so enormous that they now drive product decisions -- we’ll introduce hacks of considerable proportion or skip proper fixes to avoid having to raise a new index, which takes on the order of weeks to create, and is a large figure with a lot of zeros on the end every month in infrastructure costs to maintain.

We'd sure love index deduplication, but will probably never get it. Mongo delenda est [2].

![A grove at UC Berkeley](/assets/images/nanoglyphs/016-postgres-13/uc-berkeley-grove@2x.jpg)

### Parallel vacuum (#parallel-vacuum)

Postgres provides its transactional guarantees by carefully controlling data visibility. A deleted row isn't immediately deleted, but rather flagged as invisible for any transaction beginning after the one which deleted it. Similarly, updated rows as inserted as new versions alongside their originals, so new transactions see new data, but older ones still see the data they started with.

Vacuum is the janitor that does the sweeping up. Without it running, deleted and outdated data would accumulate until they eventually bloated the table beyond usability. It's important that it runs quickly and often.

Version 13 introduces [parallel vacuum](https://www.postgresql.org/message-id/E1itMsg-0005Kj-7h%40gemulon.postgresql.org) which does exactly what you think it does. Previously, vacuum was a sequential operation (quite a fast/efficient one). Now, its work can be broken up across multiple workers to be run in parallel, with the predictable result of an overall faster runtime. Parallel vacuum plugs into Postgres' home-grown parallel worker framework which already existed to power parallel queries (aggregates, joins, sorts, and scans) and the parallel index builds introduced in 11.

There is overhead to spinning up workers, so like many things in an SQL database, Postgres uses a variety of heuristics to be smart about it. Indexes which are too small (default: < 512 kB) don't participate in parallel vacuum, with the idea being that startup overhead would be more costly than vacuuming sequentially. Otherwise, a number of workers equal to the number of indexes on the table are spun up, topping out at a preset maximum (default: 8).

Performance gain will depend a lot on specific circumstances. Small tables will see none. But in a larger table with multiple indexes, even the addition of one extra parallel worker will cut vacuum time almost by half because the work parallelizes so well. There's obvious diminishing returns, but a realistic table that's large, with many indexes, and a lot of row churn could see its vacuum times slashed by up to three quarters.

### UUIDs included (#uuids)

This one is tiny, but still a great improvement. [`gen_random_uuid`](https://www.postgresql.org/docs/current/functions-uuid.html) generates a V4 UUID (the kind that everyone uses which are random), and is built-in.

``` sql
# SELECT gen_random_uuid();
           gen_random_uuid
--------------------------------------
 f426d205-3b81-4d7c-9bac-aff89897ebdc
```

Previously, generating a UUID involved installing the [`uuid-ossp`](https://www.postgresql.org/docs/current/uuid-ossp.html) extension, which was annoying, confusing to new users, and in plain honesty -- just pretty dumb. 99% of its users were trying to do exactly one thing. Having a simple method available to generate UUIDs out the box is a great improvement.

### Yes Postgres, I wanted to do what I said I wanted to do (#force-dropdb)

Another small one: the drop database command now accepts a force flag.

A very common development scenario was one where you wanted to drop a database (most often to immediately re-create it), but were stopped in your tracks when Postgres wouldn’t let that happen even if one other client was connected. This would lead you down the rabbit hole of figuring out which terminal tab or running program was still holding an open connection so that you could run it down. Now imagine that happening twenty times a week. It's theoretically a useful safety feature, but in practice mostly just ate time.

The CLI version of `dropdb` now takes `--force`:

``` sh
$ dropdb --force
```

And the SQL has a `FORCE` option:

``` sql
# DROP DATABASE my_db WITH ( FORCE );
```

Chalk up another win for ergonomics.

### Fast sorting of network types (#fast-network-sorting)

This is very self-serving, but faster sorting of network types is my own small contribution to Postgres 13. By packing bits into pointer-sized datums with very specific rules, sorting `inet`/`cidr` values (which store IPs like `1.2.3.4` or IP ranges like `1.2.3.0/24`) becomes roughly twice as fast because the system no longer has to go the heap for comparisons.

The nuance around how IP addresses and ranges sort relative to each other, and supporting all the rules around IPv4 as well as IPv6 means that this change is a little more difficult than it might appear. It sure stretched my remedial skills in C and bit shifting right up to their breaking point. I wrote a [detailed article](/sortsupport-inet) on how it all works.

---

## A life on our planet (#life-on-our-planet)

_Planet Earth_ and its follow up _Planet Earth II_ are some of the best video content ever produced, maybe _the_ best. Narrated by David Attenborough's soothing English tones, they capture scenes of the natural kingdom so extraordinary that they don't seem possible.

![A Life On This Planet -- Fields](/assets/images/nanoglyphs/016-postgres-13/life-planet-fields@2x.jpg)

![A Life On This Planet -- Prairies](/assets/images/nanoglyphs/016-postgres-13/life-planet-prairies@2x.jpg)

![A Life On This Planet -- Ice](/assets/images/nanoglyphs/016-postgres-13/life-planet-ice@2x.jpg)

A scene from the _Islands_ episode of _II_ where a hatching group of marine iguanas on Fernandina Island have to make their way passed a cliffside full of [Galapagos racer snakes](https://theconversation.com/in-defence-of-racer-snakes-the-demons-of-planet-earth-ii-theyre-only-after-a-meal-68514) to their new life on the rocky coastlines has stuck with me to this day. You see the entirety of the journey of specific iguanas -- the unlucky of which are ravaged in scenes of unfiltered natural violence -- and others as they sneak by (the snakes have poor eyesight) or blow by (if the snakes have noticed them) their would-be predators to escape to the safety of the sea. Like, how is it even possible to get those shots? Let alone in crystal clear high definition and cinemagraphic perfection down to the tiniest detail. Mind-blowing.

But if there's one criticism to be levied against _Planet Earth_, it's that its producers have gone to incredible lengths to gloss over the human impact on these ecosystems. Humanity is rarely even mentioned, let alone shown on-screen. This is done as a courtesy to the viewer, who has enough to worry about in their life, and doesn't want to spend their evenings watching more doom and gloom, but seeing blue whales, snow leopards, and river dolphins in apparently pristine natural environments has the effect of misleading us into thinking that there's plenty of natural world left, and that hey, maybe we're not so bad after all.

![A Life On This Planet -- Jungle](/assets/images/nanoglyphs/016-postgres-13/life-planet-jungle@2x.jpg)

![A Life On This Planet -- Polar](/assets/images/nanoglyphs/016-postgres-13/life-planet-polar@2x.jpg)

![A Life On This Planet -- Temples](/assets/images/nanoglyphs/016-postgres-13/life-planet-temples@2x.jpg)

Last week I watched Attenborough's latest _A Life On Our Planet_ (a single film rather than a series) and would recommend it. It's the same beautiful footage that we've come to expect from _Planet Earth_ and nominally framed as a biography of Attenborough himself, but he's careful to remind us that humanity's impact has been significant, even within just the bounds of his own lifetime, having vastly increased our population, decreased the wild habitat left on Earth, and produced astounding levels of pollution. He describes how noticeably more difficult it is to find wildlife for his current productions compared to those earlier in his career -- Earth's biodiversity has been shrinking, and still is.

He goes on to point out that although it's a difficult problem, there are quite a few well-understood mitigations known to be effective. Richer and better-educated societies have fewer children -- reducing global poverty and educating young girls would stabilize human population. Protecting even a third of coastal areas from fishing would produce a disproportionately large opportunity for marine populations to stabilize. Diets with a heavier emphasis on plants over meats are vastly more eco-efficient. Countries like Costa Rica act as real world role models in how it's possible to reverse deforestation.

![A Life On This Planet -- Chernobyl](/assets/images/nanoglyphs/016-postgres-13/life-planet-chernobyl@2x.jpg)

![A Life On This Planet -- City](/assets/images/nanoglyphs/016-postgres-13/life-planet-city@2x.jpg)

Very few people explicitly think of themselves as nihilists, but practically all of us are -- we all know more or less about these macro-scale problems, and know that what we're doing is unsustainable, but have unconsciously given up on getting traction on any of the major changes needed to solve them. Well, everyone except Bill Gates maybe.

I'm reminded of Alfonso Cuarón's excellent _Children of Men_ where Clive Owen's character asks his cousin, "In a hundred years, there won't be one sad fuck to look at any of this. What keeps you going?"

He replies, "You know what it is Theo? I just don't think about it."

Until next week.

![Children of Men](/assets/images/nanoglyphs/016-postgres-13/children-of-men@2x.jpg)

[1] Not that I'm happy about this one. We should probably all be getting back on Firefox before Google has de facto control over web technology.

[2] See Cato’s [Carthago delenda est](https://en.wikipedia.org/wiki/Carthago_delenda_est).
