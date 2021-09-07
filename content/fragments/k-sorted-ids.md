+++
hook = "A list of roughly sorted ("K-sorted") ID generation algorithms including ULID and UUID V6."
published_at = 2021-09-07T13:19:21Z
title = "K-sorted ID algorithms"
+++

In the last two editions of _Nanoglyph_ ([026](/nanoglyphs/026-ids) and [027](/nanoglyphs/027-15-minutes)), I wrote about a class of identifiers that you might called "K-sorted ID algorithms". A [K-sorted sequence](https://en.wikipedia.org/wiki/K-sorted_sequence) is one that is "roughly" ordered. Elements may not be exactly where they should be, but no element is very far off from where it more precisely should be. The easiest-to-grok example is to look at the IDs of tweets -- as new ones go out, they're not ordered perfectly down to the exact moment they were created, but tweets created around the same time have IDs that are close together, and the overall sequence is ascending.

In general, K-sorted IDs are two components squashed together:

    <bits of timestamp><random bits>
		
The timestamp bits keep things roughly ordered and the random bits disambiguate within the same timestamp.

Not only are ordered IDs a nicety for users, but they also tend to [perform better](https://www.2ndquadrant.com/en/blog/sequential-uuid-generators/). Newly generated IDs are inserted close to each other, which means they touch fewer pages of cache and produce less WAL.

For your convenience, here's a list of the various major K-sorted IDs, ordered approximately according to relevancy today (using my best judgement):

* [**ULID**](https://github.com/ulid/spec): 128-bit UUID compatible, with 48 bits timestamp, 80 bits random.

* [**Snowflake ID**](https://en.wikipedia.org/wiki/Snowflake_ID): 64-bit, with 41 bits timestamp, 10 bits machine ID, 12 bits per-machine sequence. This may be the oldest, but one of the most relevant as it was invented by Twitter, and both Discord and Instagram are using variations of it.

* [**UUID V6**](http://gh.peabody.io/uuidv6/): 128-bit UUID, with 64 bits timestamp (minus 4 bits UUID version embedded in it), 2 bits UUID variant, 14 bits "clock sequence", and 48 bits random. Has a [non-expired IETF draft](https://datatracker.ietf.org/doc/html/draft-peabody-dispatch-new-uuid-format).

* [**KSUID**](https://segment.com/blog/a-brief-history-of-the-uuid/) (K-Sortable Unique IDentifier): 160-bit, with 32 bits timestamp and 128 bits random. Created by Segment and probably mostly in use internally to them as well.

* [**Flake**](https://github.com/boundary/flake): 128-bit, with 64 bits timestamp, 48 bits machine ID, and 16 bits sequence. Created by Boundary (which is under new management) in Erlang, and best considered deprecated.

We'd already long since been using UUIDs for IDs and needed something UUID-compatible. I hope that something like UUID V6 eventually becomes a standard, but went with ULIDs because it seems to be the most widely used of the bunch, with some 50 implementations across any programming language you could wish for.
