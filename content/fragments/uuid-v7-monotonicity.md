+++
hook = "How Postgres' v7 UUIDs are made monotonic, and why that's a great feature."
# image = ""
published_at = 2024-12-31T15:32:43-07:00
title = "Postgres UUIDv7 + per-backend monotonicity"
+++

An implementation for [UUIDv7 was committed to Postgres](https://git.postgresql.org/gitweb/?p=postgresql.git;a=commitdiff;h=78c5e141e9c139fc2ff36a220334e4aa25e1b0eb) earlier this month. These have all the benefits of a v4 (random) UUID, but are generated with a more deterministic order using the current time, and perform considerably better on inserts using ordered structures like B-trees.

A nice surprise is that the random portion of the UUIDs will be monotonic within each Postgres backend:

> In our implementation, the 12-bit sub-millisecond timestamp fraction
> is stored immediately after the timestamp, in the space referred to as
> "rand_a" in the RFC. This ensures additional monotonicity within a
> millisecond. The rand_a bits also function as a counter. We select a
> sub-millisecond timestamp so that it monotonically increases for
> generated UUIDs within the same backend, even when the system clock
> goes backward or when generating UUIDs at very high
> frequency. Therefore, the monotonicity of generated UUIDs is ensured
> within the same backend.

This is a hugely valuable feature in practice, especially in testing. Say you want to generate five objects for testing an API list endpoint. It's possible they're generated in-order by virtue of being across different milliseconds or by getting lucky, but probability is against you, and the likelihood is that some will be out of order. A test case has to generate the five objects, then do an initial sort before making use of them. That's not the end of the world, but it's more test code and adds noise.

``` ruby
test_accounts = 5.times { TestFactory.account }

# maybe IDs were in order, but maybe not, so do an initial sort
test_accounts.sort_by! { |a| a.id }

# API endpoint will return accounts ordered by ID
resp = make_api_request :get, "/accounts"
expect(resp.map { _1["id"] }).to eq(test_accounts.map(&:id))
```

With Postgres ensuring monotonicity for UUIDv7s, the five generated objects get five in-order IDs, making the test safer [1] and faster to write. Montonicity isn't guaranteed across backends, but that's okay in well written test suites. Patterns like [test transactions](/fragments/go-test-tx-using-t-cleanup) will guarantee that each test case speaks to exactly one backend.

## 12 bits more clock (#12-bits-more-clock)

My grasp on monotonicity has always been tenuous at best, so I was curious how it was implemented here. I looked at the patch, and its approach was more obvious than I expected:

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

UUIDv7 dictates an initial 48 bits that encodes a timestamp down to millisecond precision. A millisecond is a short amount of time for a human, but quite long for a computer, and many UUIDs could easily be generated with the space of a single ms.

```
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

The Postgres patch solves the problem by repurposing 12 bits of the UUID's random component to increase the precision of the timestamp down to nanosecond granularity (filling `rand_a` above), which in practice is too precise to contain two UUIDv7s generated in the same process. It makes a repeated UUID between processes more likely, but there's still 62 bits of randomness left to make use of, so collisions remain vastly unlikely.

## The wait is on (#wait)

UUIDv7s are going to make a great core addition to Postgres, and I can't wait to start using them. Quite unfortunately, their commit was delayed past the freeze for Postgres 17, so they won't make it into an official version until Postgres 18 is cut in late 2025. So now, we wait.

[1] A common scenario is to get lucky when writing a test initially, but then having to investigate breakages later in CI as more runs reveal intermittency.
