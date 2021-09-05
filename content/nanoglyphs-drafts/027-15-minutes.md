+++
image_alt = "Looking down the coastline towards Pacifica"
# image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/027-15-minutes/red-rock@2x.jpg"
published_at = 2021-09-04T18:50:46Z
title = "Red Rocks + 15 Minutes + K-sorted IDs"
+++

I spent the last couple of weeks in Denver. On my last day there I walked through various Lakewood parks, down [Dinosaur Ridge](https://en.wikipedia.org/wiki/Dinosaur_Ridge), and up to the [Red Rock Amphitheatre](https://en.wikipedia.org/wiki/Red_Rocks_Amphitheatre). Maybe the most unique concert venue on Earth, its first rock-and-roll show is considered to be from The Beatles, on tour in '64, which also notably was the only show in the United States that wasn't sold out.

In '71, a five-year ban was enacted on rock shows there after a thousand rabid fans showed up to a sold-out show for Jethro Tull ... without tickets. After being denied entry, they charged police and started lobbing rocks at them. Police responded by discharging tear gas at the gate crashers, which accidentally had the misfortune of being lifted by the wind, carrying over the hills, and right into the amphitheatre and stage. Since then, Red Rocks has been played by thousands of acts including The Grateful Dead, U2, and even The Blues Brothers, which were apparently a real band before being a Hollywood blockbuster franchise.

I never got to see the inside of place. After walking 20 km through Colorado hills without doing enough advanced research, I arrived sweaty, dusty, and dehydrated, only to be told that they'd just shut down to prepare for the night's show. I was informed by the lady watching the gate that I could still buy a ticket, but that the event was for an unenviable genre called "electronic dance music" (this explained why a parking lot in the middle of nowhere was full of burner lookalikes with hula hoops). With another 15 km of return still ahead of me and fast-dwindling water reserves, I had to decline, but promised to be back.

But even with that disappointment, the area's rock formations (whose red-ness isn't exaggerated) are still a magnificent sight from the outside, and hiking up to them procures some of the best views. What a beautiful place.

<img src="/photographs/nanoglyphs/027-15-minutes/red-rock-landscape@2x.jpg" alt="The Red Rocks Amphitheatre and surrounding landscape" class="wide" loading="lazy">

_(This shot from the other side gives you a much better idea of the scale of these monsters. The one on the right is "Ship Rock" which makes up one side of the amphitheatre. The park was previously named "Garden of the Titans", which is very apt.)_

<img src="/photographs/nanoglyphs/027-15-minutes/rezz-rocks@2x.jpg" alt="Poster for Rezz Rock III" class="wide" loading="lazy">

_(The show I missed. Another Canadian -- [Rezz](https://en.wikipedia.org/wiki/Rezz). Included here because this poster is great.)_

---

## Shallow stacks (#shallow-stacks)

Last week, I trolled Rubyists and ex-colleagues with this tweet:

<a href="https://twitter.com/brandur/statuses/1431297226677133316"><img src="/photographs/nanoglyphs/027-15-minutes/15-minute-tweet@2x.png" alt="Tweet of 15 minutes to upgrade" class="img_constrained" loading="lazy"></a>

Twitter is the land of over-simplification and hyperbole, and I'm certainly guilty of that too. However, what I said here is accurate, and it's worth going into a little more detail.

First on the claim of 15 minutes in Go, which is less interesting but still important. 15 minutes wasn't exaggerating, and that includes writing code/config, getting it reviewed, upgrading staging, testing, upgrading production, and testing there too. It actually took longer than some of my historical Go upgrades because 1.16 contained a very rare (and very minor) backward incompatibility -- it got a little more strict about not allowing an HTTP response that's been declared a 204 (no content) to be written to. We had one test case where we were improperly doing this, and the 15 minutes also includes me debugging and fixing that.

One of Go's best features is that compatibility is taken _very_ seriously -- in many years of using it, I've never once encountered a non-trivial problem upgrading it. The common case is that you bump a number, and your app becomes incrementally faster, for free.

### Stripe-flavored Ruby (#stripe-ruby)

The 15 weeks for upgrading Ruby at Stripe is harder to measure. This is partly exaggeration and partly understatement. Upgrades took wildly differing amounts of time depending on what year it was, how motivated the upgrade was, and the people involved in making it. Upgrades faster than 15 weeks definitely happened, but especially by the later years, 15 weeks from inception to finish isn't too far off. And mind you, those are 15 weeks from when the project was undertaken -- we weren't tailing the latest stable Ruby or even close to it. Upgrades often wouldn't happen for months or even years following a new release.

There's a myriad of reasons that upgrades took so long. The biggest was that the entire stack was so deep and so heavily custom. Everything from the sheer amount of Ruby code written, to an incredibly complex production set up, to an immensely intricate Jenkins environment for CI worked as considerable upgrade back pressure. Bump that version of Ruby and things would break, _a lot_ of things. It was someone's job to go through them one by one and do the necessary work to get the project over the line.

Another big one is that very careful attention had to be paid to possible performance and memory regressions. The Stripe API runs as a single monolithic Ruby process, and a vast amount of Ruby code needs to be loaded to start it up. This is made much worse by Ruby's non-support for parallelism [1], making a forking multi-process in the style of Unicorn very common. For the longest time the API ran on a heavily customized [Thin web server](https://github.com/macournoyer/thin) combined with NIH [Einhorn](https://github.com/stripe-archive/einhorn) technology for Unicorn-like features.

For those not well-versed with Ruby deployments, the way this works is that a Ruby app is loaded into a single process which then forks itself many times to produce child processes that will handle requests. In theory, the child processes can share memory with their parent, but because Ruby's GC tends to taint memory pages quickly, in practice memory isn't shared for long, and all children balloon up to the same size as their parent [2]. In addition to parallelism, this set up also allows for graceful restarts -- the parent process will reload itself upon receiving a signal, and coordinate rotating out its children for new processes running updated code, while also giving each one time to finish what it's working on.

Process-based parallelism is fine, except that it's not a good fit combined with Ruby's memory profile. Because of the sheer quantities of Ruby involved, each API worker needed somewhere around a gigabyte of memory, and could handle exactly one request at a time. Thin was eventually retired in favor of Puma, but even then Puma's multi-threaded features were never used because hundreds of thousands of lines of code had never been vetted for thread safety. And even when a Ruby code base is thread safe, there's still an argument to be made that it shouldn't use threads because Ruby's GIL dictates that only one of them can be running Ruby code at a time. This is less big of a problem than it sounds because web applications tend to be IO-bound, but a processed-based Ruby application will always perform better than a thread-based one, so where performance is a critical consideration and cost is less of one, you might want to just throw money at the problem.

Back to the original point I was trying to make: Stripe servers ran big and hot, so even minor changes or regressions in Ruby's runtime or GC could have major effects in production. The only safe way to put a new version into play was to first deploy it to a Canary and have someone keep an eye on charts around latency, CPU usage, GC pressure, and memory for a few days. If anything came up, we'd have to dig into the problem and find a work around.

<img src="/photographs/nanoglyphs/027-15-minutes/denver-river@2x.jpg" alt="The South Platte River running through central Denver" class="wide" loading="lazy">

## K-sorted IDs ad nauseam (#k-sorted)

A few weeks ago I wrote about [primary IDs](/nanoglyphs/026-ids) in applications, UUIDs versus sequences, and more novel techniques like ULIDs and Stripe's generated IDs, both of which aim to introduce a time component so that they're roughly in ascending order.

I got one of my best newsletter responses ever (people care about their IDs!), so I'm following up here with a few of them. It turns out that generating random-ish IDs in roughly ascending order is far from a unique idea, with many examples of prior art besides the ones that I'd mentioned.

On a meta note, I've been wanting to include reader feedback/opinions since starting this newsletter, so keep it coming.

### UUID V6 (#uuid-v6)

Bo notes that UUID has a [V6 specification](http://gh.peabody.io/uuidv6/) that adds a leading time component (recall that UUID V4 is totally random).

Recall that a UUID is 128 bits long. A V6 UUID looks like:

* 64 bits of timestamp + 4 bits of version. UUIDs embed a version in a specific bit location, so the timestamp is massaged in around the version.
* 2 bits for UUID variant. This is a concept common to [all UUIDs](https://en.wikipedia.org/wiki/Universally_unique_identifier#Variants).
* 14 bits for "clock sequence". If a single machine could have generated more than one ID within the same hundred-nanosecond interval, this would field is incremented so that each one is unique.
* 48 bits for "node", or random bits. (The specification document also allows for these to be generated based on the local machine's MAC address, but you probably shouldn't do that.)

I have a hard time tracking what's going on with IETF drafts, but V6 UUIDs [have one](https://datatracker.ietf.org/doc/html/draft-peabody-dispatch-new-uuid-format) that's not expired. Although we have a lot of alternatives at this point, it'd certainly be nice if something like this would become a standard and lead to broad implementation in standard libraries.

## Snowflakes (#snowflakes)

Ben writes in about [Snowflake IDs](https://en.wikipedia.org/wiki/Snowflake_ID) (for the literalists out there, note that this is tongue and cheek):

> My one gripe with sequential UUID / ULID is that they are unnecessarily long. 32 characters! That's like 40% of my terminal. And it seems like that much entropy is overkill for ~anything.
>
> How do you feel about Snowflake IDs? 50% shorter, and unlike UUIDs, people won't look at you funny for base64'ing them, which takes them down to a mere 11 characters. I guess it's somewhat more cumbersome than ULIDs to generate snowflake IDs from an arbitrary service because of the shard ID, but if you just have a db it's fine.

A snowflake is half the size of a UUID at 64 bits, and made up of:

* 41 bits timestamp, getting to millisecond precision.
* 10 bits machine ID (preventing clashes between machines).
* 12 bits per-machine sequence for snowflakes generated in the same millisecond.

Although by no means a standard, they were coined by Twitter, and later Discord and Instagram picked them up as well, albeit with slight variations. If Twitter, Discord, and Instagram can get away with "only" 64-bit IDs, you probably can too.

## KSUID (#ksuid)

Michael writes in about [Segment's KSUIDs](https://segment.com/blog/a-brief-history-of-the-uuid/). From that article:

> Thus KSUID was born. KSUID is an abbreviation for **K**-**S**ortable **U**nique **ID**entifier. It combines the simplicity and security of UUID Version 4 with the lexicographic k-ordering properties of Flake [3].

A [K-sorted sequence](https://en.wikipedia.org/wiki/K-sorted_sequence) is one that is "roughly" ordered. Elements may not be exactly where they should be, but no element is very far off from its precise location. All the formats we've talked about so far -- ULIDs, Stripe IDs, UUID V6, Snowflakes, and KSUIDs -- are all K-sorted.

KSUIDs move the needle up to 160 bits with:

* 32 bits of timestamp from a custom epoch at 1 second resolution.
* 128 bits of randomness.

Once again, very similar to the formats we've covered so far. It's worth nothing that 160 bits might be overkill for most purposes -- a good chunk of Segment's rationalization for the new format is concerns around UUID V4 collisions due to implementation bugs. This is certainly possible, but in my experience, doesn't turn out to be a problem in practice.

### Pure randomness as a feature (#randomness)

Justin writes in with a very thorough article from [Cockroach Labs on choosing index keys](https://www.cockroachlabs.com/blog/how-to-choose-db-index-keys/). Cockroach maps its normal `SERIAL` type to a function called [`unique_rowid()`](https://www.cockroachlabs.com/docs/stable/functions-and-operators#id-generation-functions), which generates a 64-bit ID combining some timestamp and some randomness that should seem pretty familiar by now.

However, because CockroachDB involves having many cooperating nodes where writes can happen, a K-sorted ID won't make good utilization of available nodes in an insert-heavy system, and would perform much worse compared to a V4 UUID. Cockroach provides sharded keys to work around this problem and get the best of both worlds:

> Even though timestamps avoid the worst bottlenecks of sequential IDs, they still tend to create a bottleneck because all insertions are happening at around the current time, so only a small number of nodes are able to participate in handling these writes. If you need more write throughput than timestamp IDs offer but more clustering than random UUIDs, you can use sharded keys to spread the load out across the cluster and reduce hotspots.

Here's a simple example of Cockroach DDL where the K-ordered primary ID is hashed so that insert get random, uniform distribution:

```
CREATE TABLE posts (
    shard STRING AS (substr(sha256(id::string), 64)) STORED,
    id SERIAL,
    author_id INT8,
    ts TIMESTAMP,
    content TEXT,
    PRIMARY KEY (shard, id),
    INDEX (author_id, ts));
```

### ULIDs --> prod (#ulids-prod)

For my own purposes, I ended up putting [ULIDs](https://github.com/ulid/spec) [4] into production. I probably would have used UUID V6 if it was more standard and more broadly available, but for my money, ULID seems to be the K-sorted ID format with the most uptake and most language-specific implementations.

We were already using UUIDs so the format we chose needed to be UUID compatible. Even if didn't, being able to reuse the built-in Postgres `uuid` data type is very convenient -- drivers all support it out of the box, and there's very little friction in getting everything working. We're using [pgx](https://github.com/jackc/pgx) so our IDs are not only stored efficiently in Postgres as 16-byte arrays, but transferred as byte arrays using Postgres' binary protocol, treated as `[16]byte` in our Go code, and only rendered as strings at the last possible moment when data needs to be sent back to a user. (As opposed to in most languages/frameworks where UUIDs become a string before they leave the database.)

I wrote a simple UUID-compatible SQL generation function:

``` sql
CREATE OR REPLACE FUNCTION gen_ulid()
RETURNS uuid
AS $$
DECLARE
  timestamp  BYTEA = E'\\000\\000\\000\\000\\000\\000';
  unix_time  BIGINT;
  ulid       BYTEA;
BEGIN
  -- 6 timestamp bytes
  unix_time = (EXTRACT(EPOCH FROM NOW()) * 1000)::BIGINT;
  timestamp = SET_BYTE(timestamp, 0, (unix_time >> 40)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 1, (unix_time >> 32)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 2, (unix_time >> 24)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 3, (unix_time >> 16)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 4, (unix_time >> 8)::BIT(8)::INTEGER);
  timestamp = SET_BYTE(timestamp, 5, unix_time::BIT(8)::INTEGER);

  -- 10 entropy bytes
  ulid = timestamp || gen_random_bytes(10);

  -- Postgres makes converting bytea to uuid and vice versa surprisingly
  -- difficult. This hack relies on the fact that a bytea printed as a
  --  string is actually a valid UUID as long as you strip the `\x`
  -- off the beginning.
  RETURN CAST(substring(CAST (ulid AS text) from 3) AS uuid);
END
$$
LANGUAGE plpgsql
VOLATILE;
```

Table DDL then gets this `DEFAULT` annotation:

``` sql
ALTER TABLE access_token
    ALTER COLUMN id SET DEFAULT gen_ulid();
```

From some places in Go code we use the [Go ULID package](https://github.com/oklog/ulid). This has the ever-so-slight advantage of using a monotonic entropy pool for the random component that brings the chance of collision down from basically-zero to zero-zero. For our purposes it's definitely overkill, but also, why not.

<img src="/photographs/nanoglyphs/027-15-minutes/millenium-bridge@2x.jpg" alt="The Millenium Bridge in LoDo Denver" class="wide_portrait" loading="lazy">

## On Denver (#denver)

I'm in Denver. I think it surprises most people when my answer to "what are you doing there?" is "nothing, really". I've never been before, and I came mostly to look around the city, get a feel for what it's like, and maybe most importantly, to not be in California.

[1] Ruby does finally have real parallelism, but it's still preliminary, and existing apps don't get to take advantage of it for free. More on Ractors on [issue 018](/nanoglyphs/018-ractors).

[2] For more details on Ruby memory management, see ["The Limits of Copy-on-write"](/ruby-memory).

[3] [Flake](https://github.com/boundary/flake) is yet another K-sorted sequence format from Boundary that was inspired by Twitter's Snowflake. It jumps back to 128 bits.

[4] More on ULIDs in [issue 026](/nanoglyphs/026-ids).
