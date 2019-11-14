+++
published_at = 2019-10-03T15:05:23Z
title = "Diving Roatan, 11s, 12s"
+++

![On the way to West Bay](/assets/images/passages/004-roatan/beach@2x.jpg)

Hi everyone --

Long time readers may recall that in [the last edition of Passsages](/passages/003-koya) I promised to send these bursts a little more often after a longtime hiatus. Well, obviously that didn’t work out. You can consider me sufficiently chagrined through self-flagellation, or not — feel free to reply and give me a hard time about it.

If you don’t know why you’re getting this email, it’s probably because you signed up for my newsletter on my website [brandur.org](https://brandur.org/newsletter) at some point, but they’re sent so infrequently that it may well have been in the distant past. If you never want to see another one, all it takes to permanently exorcise your inbox is a single click to [this unsubscribe link][unsubscribe].

Otherwise, here we go. In this edition I’ll say a few words about a recent trip to Honduras, the iPhone 11, and Postgres 12.

## Roatan (#roatan)

I had a bit of a _Grand Budapest Hotel_ experience [1], which up until then, was something that I didn’t think happened in real life. It was low season thanks to an upcoming rainy season, and aside from me, the hotel and grounds were almost empty.

I’d show up in the morning for dive expeditions, and be the sole passenger on the sole boat that went out for the day (of a fleet of ~nine). It was a great recipe for seeing marine life (divemasters have little else to do except keep an eye out), and a little weird — with no one else to pay attention to, our boat's deck hand would assemble every piece of my equipment up to and including defogging my goggles, proceed to mount that equipment _onto me_, and lead me over the side of the boat like a show pony. Highly convenient, but in the way that being a doted on five year old is highly convenient.

![The Seagrape's dock](/assets/images/passages/004-roatan/dock@2x.jpg)

The area's fauna is impressive. Eagle rays were the signature macro life, one we saw spanning about seven feet wide, and with its long tail trailing far behind it. We saw sea turtles nearly every dive, the biggest a full meter in length. I saw my first [mantis shrimp][oatmeal], and was duly amazed by the loud clacks of its snapping claws -- fully audible underwater despite the animal being a few inches long (the shrimp strikes so quickly that it produces vapor-filled bubbles between its appendages and target called "cavitation bubbles", the collapse of which causes a small shock wave). Aside from those: seahorses, [nudibranchs][nudibranch], neck crab, morays.

I got into a comfortable routine: wake up early. Eat. Dive. Boat to shore. Wait 20 minutes. Boat to sea. Dive again. Rinse equipment. Finish with the ocean by 11 o'clock. Read in hammock or go look around the island.

### Rustic living (#rustic-living)

With Yelp not-a-thing and few other people around to telegraph good restaurants through business, I found places to eat the _really_ old-fashioned way — word of mouth, and walking around town peering into patios and looking at menus. Sometimes you got a good one; often you didn’t; but it was mechanically satisfying in the same way as making pour-over coffee or reading a paperback. Occasionally I'd find a really treasure, like the _Cafe de Palo_ below, and the serendipity would produce a little turbocharge of unbridled excitement.

![Cafe de Palo](/assets/images/passages/004-roatan/cafe-de-palo@2x.jpg)

Although walking infrastructure was typically Central American (bad), it was a pleasant surprise to find that it was possible to walk the length of the beach from the West End (where I was staying — a slightly grungy, more down-to-earth part of the island, think San Francisco's Mission district) to the West Bay (a series of more traditional high end coastal resorts) ending upon reaching the cliffs overlooking the Grand Roatan at the end of the beach. Just follow the coast, duck under a volleyball net, and when you hit a series of rocks that look semi-impassable, climb over them anyway. Ten kilometers round trip. Highly recommended.

![West Bay](/assets/images/passages/004-roatan/west-bay@2x.jpg)

### Hunting invaders (#lionfish)

In a way, the lionfish is a striking beauty, but, when it comes to their presence in most of the world, they’re one of the most formidable and harmful invasive species ever known. Native to the Indo-Pacific, in recent decades they’ve been making steady progress taking over the west Atlantic, Caribbean, and Mediterranean. Venomous spines make them a near perfect predator — so much so that a _single_ lionfish on a reef has been observed to [reduce juvenile fish populations by 79%](https://today.oregonstate.edu/archives/2010/apr/lionfish-invasion-continuing-expand).

![A lionfish (from the Creative Commons)](/assets/images/passages/004-roatan/lionfish@2x.jpg)

To help control their numbers, many governments have blessed divers with special hunting rights. Given that for most of my dives my divemasters’ only charge was me, they’d often take advantage of them.

The state of the art in lionfish hunting technology is a short three-pronged spear attached to an elastic band, as shockingly primitive as it is effective. The hunter pulls the spear back along the length of the elastic, aims, and ***POW*** -- invader skewered. I never saw a miss. My divemasters’ carried a rusty pair of pliers in their BCDs and got started on trimming a catch’s spines immediately. We'd then find something interesting to feed it to, like a moray eel, unless it was a large catch -- those were specially earmarked for feeding to humans.

Even in their home part of the world, lionfish have astonishingly few predators. Some larger animals (e.g. morays, groupers, sharks) are known to prey on them, but not consistently. It’s speculated that animals that feed on juvenile lionfish and larvae may be one of the key factors in controlling home populations, but who or what does that job is still poorly understood. That fact, combined with the problem that diver control efforts are limited to shallow, human-frequented depths, make the prospect of meaningful population control difficult.

![Seagrape door](/assets/images/passages/004-roatan/padi-door@2x.jpg)

## The triclops (#triclops)

I picked up an iPhone 11 on its first day of available about 14 hours before my flight, so for a time I can probably claim to have had the only one in Honduras. It’s largely the same as the iPhone X launched two years ago, with the notable exception of the new camera system.

Apple has introduced a new 0.5x wide angle lens to make the phone's visual array a triplet. At first glance I assumed it was a gimmick, but that doesn't seem to be the case. After deploying the phone's camera for a few days, I realized that 0.5x was more often than not the field of view that I wanted. Like any wide angle lens it distorts things, but that distortion is useful: it lets you capture a field of view that gives the viewer a much better feeling of what it's like to actually be there.

![Sunset over the Seagrape](/assets/images/passages/004-roatan/sunset@2x.jpg)

Night mode works well too. The Halide people wrote an interesting [piece on the iPhone 11 camera][halide] that chalks the improvements up to the factors you'd expect: faster lenses, improved ISO range (33% improved on the "main" lens), and smarter software that composes better final results.

I like the camera so much that when I was carrying just my phone, I didn't get my normal FOMO of not having a real camera system with me. People say that with every new iPhone release, but it really is more true with every one. We’ve long since passed the point where most people’s phone is their camera, but hobbyist photographers who tend to invest in more traditional camera equipment are going to be attriting to just smartphones too. That will likely eventually include yours truly -- notably, every photo in this edition was taken on an iPhone [2].

## Postgres 12 (#postgres-12)

A few weeks back spelled the official release of Postgres 12. See the [full release notes][postgres12] for all the details, but here's a couple big ones. [CTEs][cte] no longer act as an optimization barrier for queries, which in some situations can result in huge performance gains. Take the following:

``` sql
WITH x AS (
    SELECT * FROM t
)
SELECT * FROM x
    WHERE id = 1000;
```

It's relatively obvious to a human that this simple query should be quite easy for a database to execute. Although we've selected all the results of `t` into the `x` CTE, the final clause only cares about a single row, so we should be able to reach right into `t` to get it. However, in previous versions of Postgres, `x` was treated as a black box, and its results would have to be materialized in a temporary buffer somewhere before being queried (at great expense). Postgres 12 allows the query to reach right into `t` to avoid those temporary results and take advantage of its indexes. Not every CTE will benefit to this extent, but it'll be huge for some.

### More `CONCURRENTLY` (#reindex-concurrently)

The new `REINDEX CONCURRENTLY` command has been introduced to supplement `CREATE INDEX CONCURRENTLY` and `DROP INDEX CONCURRENTLY`. As with the latter two, this has the effect of making things fundamentally better and fundamentally safer. Postgres 12 also introduces some space utilization and read/write performance improvements to B-tree indexes, so rebuilding some of your indexes is an easy way to make your database a little smaller, and a little faster.

### Even better partitioning (#partitioning)

Lastly I'll call out that some incremental performance improvements to partitioned tables came in including better query performance on tables with thousands of partitions, faster `INSERT`s, and non-blocking `ATTACH PARTITION`. Continued work on making partitioning better is some of the most important progress that the Postgres project can make because degenerate performance on large tables is an operational problem that every user with a non-trivial use case is likely to run into, so it's great to see this progress happening.

---

With every new release I wonder when Postgres will be “done”. Surely at some point all the most interesting features for a database will be under the bridge and the project will move onto much more incremental progress. And yet, every time there’s something interesting to talk about. Here’s to hoping that never changes.

Well, that’s it for now. I’ve got some ideas for another couule of these queued up, so I’ll give you the usual refrain about how it shouldn’t be so long before the next one, but I’ll be the first to admit that my track record doesn’t inspire confidence.

Until next time.

[1] “What few guests we were had quickly come to recognize each other by sight as the only living souls residing in the vast establishment. We were a very reserved group it seemed, and without exception, solitary.” -- _The Grand Budapest Hotel_

[2] Except that of the lionfish, although let's hope that iPhone waterproofing continues its admirable trend towards allowing that one day.

[cte]: https://www.postgresql.org/docs/current/queries-with.html
[halide]: https://blog.halide.cam/inside-the-iphone-11-camera-part-1-a-completely-new-camera-28ea5d091071
[nudibranch]: https://en.wikipedia.org/wiki/Nudibranch
[oatmeal]: https://theoatmeal.com/comics/mantis_shrimp
[postgres12]: https://www.postgresql.org/docs/release/12.0/
[unsubscribe]: %unsubscribe_url%
