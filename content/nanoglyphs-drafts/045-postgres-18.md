+++
image_alt = "Pike Place"
# image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/045-postgres-18/pike-place-loback@2x.jpg"
published_at = 2025-09-27T11:31:51-07:00
title = "Postgres 18"
hook = "A few features from Postgres 18 and a note on upcoming travel to Indonesia."
+++

Readers --

It's Christmas week for databases, with [Postgres 18 released last Thursday](https://www.postgresql.org/about/news/postgresql-18-released-3142/). Postgres only ships new features once a year, so the annual major version cut is a time to stop, take note, and experiment a little. We usually wait for one minor release before upgrading production so that'll be another month or two, but it's something to start looking into.

There's the usual long laundry list of new features and enhancements. I find a little humor in that the one which will touch the most peoples' lives is UUIDv7, which sums up tidily to only a [few hundred lines of code](https://git.postgresql.org/gitweb/?p=postgresql.git;a=commitdiff;h=78c5e141e), half of which are comments. It could've shipped a year ago but for the Postgres' team's insistence on waiting for perfect geologic stability in the RFC akin to Yucca Mountain before letting it out the door. But better late than never.

Let's take a high level look at some headline features, and with a short note about Indonesia at the bottom.

## 1 --- Async I/O (#async-io)

Postgres 18 brings in a new async I/O subsystem based on [io_uring](https://man7.org/linux/man-pages/man7/io_uring.7.html). It's still off by default, but can easily be configured on Linux (and Linux only):

``` txt
io_method = io_uring
```

Async I/O doesn't automatically take effect for any arbitrary I/O operation. So far Postgres can only take advantage of async I/O for sequential scans, bitmap heap scans, and `VACUUM`, but with foundations now in place, the hope is its use can be expanded in future releases.

I've played with async I/O minimally, and am not quite sure what to realistically expect out of it yet. Synthetic benchmarks show 2-3x speedups in some circumstances, but the real world improvements are probably more modest. As one test, I tried [River's benchmark suite](https://riverqueue.com/docs/benchmarks) against Postgres 18 with `io_method = io_uring` on, and it made no observable difference.

Lukas [wrote a more in-depth article](https://pganalyze.com/blog/postgres-18-async-io) containing an example case where async I/O would be more help.

I like the idea that even as our chips and disks get faster, enhancements are simultaneously being pushed through on the software level through the likes of io_uring. So even while every newly available hardware cycle in web or mobile apps is immediately burned away through increasingly deep stacks of JavaScript nonsense, at least our databases and server software should be getting strictly faster.

## 2 --- UUIDv7 (#uuidv7)

This one's been written about once a week for a year now, so I'll stop myself from going into too much detail, but in essence: UUIDv7 is what most people have wanted out of UUIDs since the early 2000s. They're generated in an ascending sequence which is good UX from a user's perspective, but also good for B-tree insert and WAL performance. Downsides are supremely marginal. Use them, especially for new projects. Competing alternatives like ULID can be safely retired.

Generate them with the new `uuidv7` function, and store them into the preexisting `uuid` data type:

``` sql
# SELECT uuidv7();
                uuidv7
--------------------------------------
 019986a7-8745-7b6e-bf94-63b91465cb1a
```

I did a short writeup last year on [how the Postgres implementation guarantees monotonicity](/fragments/uuid-v7-monotonicity), which is clever, and surprisingly simple. A little of the relevant C code:

``` c
/*
 * Generate UUID version 7 per RFC 9562, with the given timestamp.
 *
 * UUID version 7 consists of a Unix timestamp in milliseconds (48
 * bits) and 74 random bits, excluding the required version and
 * variant bits. To ensure monotonicity in scenarios of high-
 * frequency UUID generation, we employ the method "Replace
 * LeftmostRandom Bits with Increased Clock Precision (Method 3)",
 * described in the RFC. This method utilizes 12 bits from the
 * "rand_a" bits to store a 1/4096 (or 2^12) fraction of sub-
 * millisecond precision.
 *
 * ns is a number of nanoseconds since start of the UNIX epoch.
 * This value is used for time-dependent bits of UUID.
 */
static pg_uuid_t* generate_uuidv7(int64 ns) {

...

/*
 * sub-millisecond timestamp fraction (SUBMS_BITS bits, not
 * SUBMS_MINIMAL_STEP_BITS)
 */
increased_clock_precision = ((ns % NS_PER_MS) * (1 << SUBMS_BITS)) / NS_PER_MS;

/* Fill the increased clock precision to "rand_a" bits */
uuid->data[6] = (unsigned char) (increased_clock_precision >> 8);
uuid->data[7] = (unsigned char) (increased_clock_precision);

/* fill everything after the increased clock precision with random bytes */
if (!pg_strong_random(&uuid->data[8], UUID_LEN - 8))
    ereport(ERROR,
            (errcode(ERRCODE_INTERNAL_ERROR),
            errmsg("could not generate random values")));
```

UUIDs are 128 bits. UUIDv7 dictates an initial 48 bits that encodes a timestamp down to millisecond precision. A millisecond is a short amount of time for a human, but quite long for a computer, and many UUIDs could easily be generated within the span of one ms.

``` txt
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                      48 bits unix_ts_ms                       |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|   48 bits unix_ts_ms (cont)   |  ver  |    12 bits rand_a     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|var|                    62 bits rand_b                         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                     62 bits rand_b (cont)                     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

Postgres's UUIDv7 implementation solves the problem by repurposing 12 bits of the UUID’s random component to increase the precision of the timestamp down to nanosecond granularity (filling `rand_a` above), which in practice is too precise to contain two UUIDv7s generated in the same process. It means that a repeated UUID between processes is technically possible, but there’s still 62 bits of randomness left to make use of, so collisions are vastly unlikely.

## 3 --- OLD and NEW in RETURNING (#old-new-returning)

I published a ~5 paragraph post for Crunchy's blog on [OLD and NEW rows becoming available in RETURNING statements](https://www.crunchydata.com/blog/postgres-18-old-and-new-in-the-returning-clause).

I usually don't read the release notes in advance, so this one was a nice surprise. It's a little niche, but an example of one place it's really useful is determining whether a row being returned from upsert is new or not. Previously, the only way to determine this was a sketchy check on `xmax = 0` that relied on an [implementation detail in Postgres' locking mechanism](https://stackoverflow.com/a/39204667) to work:

``` sql
INSERT INTO webhook (
    id,
    data
) VALUES (
    @id,
    @data
)
ON CONFLICT (id)
    DO UPDATE SET id = webhook.id -- force upsert to return a row
RETURNING webhook.*,
    (xmax = 0) AS is_new;
```

In Postgres 18, we can just check on whether `OLD IS NULL`. More legible, and a part of the stable public API that won't suddenly change out from under us in a future version:

``` sql
INSERT INTO webhook (
    id,
    data
) VALUES (
    @id,
    @data
)
ON CONFLICT (id)
    DO UPDATE SET id = webhook.id -- force upsert to return a row
RETURNING webhook.*,
    (OLD IS NULL)::boolean AS is_new;
```

## Dispatches from the Pacific (#pacific)

Next week I'm heading out on a month-long trip in Indonesia where I'll be stopping at three separate dive sites and with a stop at Komodo.

My tech giant's insistence on the use of "secure endpoints" [1] will work in my favor because while normally I'd think nothing of bringing a laptop along to check in at work occasionally, it'd take significant persuasive skill to convince me to bring a *second* laptop.

The last time I visited Indonesia in 2019, the islands where the dive resorts were had point-to-point radio internet links to the larger islands with about 300 baud of bandwidth. If you were the very first person to wake up (amongst divers, this is competitive), you might be able to get on there and check your email for a few minutes. But as soon as the *second* person at the resort woke up, that puts enough contention on the uplink that both of you will just be staring at loading screens for the rest of the day.

That's a long way of saying: I'm going to try and write a few dispatches of this newsletter while I'm out there, technology permitting, and in that case expect it to be transformed on a temporary basis into an Indonesia blog. It might not work, so in the equally likely case that technology is _not_ permitting, I'll send photos when I get back.

Until next week.

<img src="/photographs/nanoglyphs/045-postgres-18/coral-eye@2x.webp" alt="Coral Eye coastline" class="wide" loading="lazy">

[1] Secure endpoint = laptop purposely pre-provisioned with a rootkit and more spyware than you can shake a stick at. A common feature at large companies.