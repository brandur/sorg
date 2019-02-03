---
title: "Designing SortSupport Abbreviated Keys for
  inet/cidr"
published_at: 2019-02-03T20:27:44Z
location: San Francisco
hook: TODO
---

A few weeks ago, I wrote about [how SortSupport works in
Postgres](/sortsupport) to vastly speed up sorting on
pass-by-reference types like `numeric`, `text`, `uuid`, and
`varchar`. It works by generating abbreviated keys for
larger heap values that are representative of them for
purposes of sorting, but which fit nicely into the
pointer-sized value (called a "datum") in memory that
Postgres uses for sorting. Many values can be sorted just
based on their abbreviated key, saving trips to the heap
and increasing sorting throughput. Faster sorting leads to
faster operations like `DISTINCT`, `ORDER BY`, and `CREATE
INDEX`.

I recently [posted a patch][patch] to add SortSupport for
the `inet` and `cidr` types, which by my measurement, a
little more than doubles sorting speed on them. `inet` and
`cidr` are the types used to store network addresses or
individual hosts and in either IPv4 or IPv6 (they generally
look something like `1.2.3.0/24` or `1.2.3.4`).

`inet` and `cdir` have some important subtleties in how
they're sorted which made designing an abbreviated key that
would be faithful to those subtleties but still efficient,
somewhat challenging. Because their size is limited,
abbreviated keys are allowed to show equality even for
values that aren't equal (Postgres will fall back to
authoritative comparison to confirm equality or tiebreak),
but they should never falsely indicate *inequality*.

## The anatomy and differences of inet and cidr (#inet-cidr)

`inet` and `cidr` values have three components (and take
for example the value `1.2.3.4/24`):

1. A network, or bits in the netmask (`1.2.3.`).
2. A netmask size (`/24` which is 24 bits).
3. A subnet, or bits outside of the netmask (`.4`).

The netmask size dictates how many bits of the value belong
to the network. This can get a little confusing because
although it's most common to see byte-sized blocks like
`/8`, `/16`, `/24`, and `/32`, it's allowed to be any
number between 0 and 32. It's easy a byte-sized network out
of a value (like `1.2.3.`) because you just stop at a byte
boundary, but when it's not a round number you have to
think at the binary level. For example, if I have the value
`255.255.255.255/1`, the network is just the leading bit.
255 in binary is `1111 1111`, so the network is the bit
`1` and the subnet is 31 consecutive `1`s.

!fig src="/assets/sortsupport-inet/inet-cidr-anatomy.svg" caption="The anatomy of inet and cidr values."

An address whose entire value is in the network (`/32` for
IPv4 or `/128` for IPv6) specifies just a single host, and
for display purposes the netmask size is usually omitted.
We'd show `1.2.3.4` instead of `1.2.3.4/32`.

The difference between `inet` is `cdir` is that `inet`
allows a values outside of the netmasked bits. The value
`1.2.3.4/24` is possible in `inet`, but in `cidr` only
zeroes may appear after the network like `1.2.3.0/24`.
They're nearly identical, with the latter being strictly
more constraining (and when working with data, that's a
good thing).

In the Postgres source code, `inet` and `cidr` are
represented by the same C struct. Here it is in
[`inet.h`][inet]:

``` c
/*
 * This is the internal storage format for IP addresses
 * (both INET and CIDR datatypes):
 */
typedef struct
{
    unsigned char family;      /* PGSQL_AF_INET or PGSQL_AF_INET6 */
    unsigned char bits;        /* number of bits in netmask */
    unsigned char ipaddr[16];  /* up to 128 bits of address */
} inet_struct;
```

## Sorting rules (#sorting-rules)

In Postgres, `inet`/`cidr` sort with a specific set of
rules:

1. IPv4 always appears before IPv6.
2. The bits in the network are compared (`1.2.3.`).
3. Netmask size is compared (`/24`).
4. All bits are compared. Having made it here, we know that
   the network bits and network sizes are equal, so we're
   in effect just comparing the subnet `.4`.

These rules combined with the fact that we're working at
the bit level produce ordering that in cases may not be
intuitively all that obvious to a human reader. For
example, `192.0.0.0/1` sorts *before* `128.0.0.0/2` despite
192 being the larger number. The reason is that when
comparing them, we start by looking at the common bits
available in both networks which comes out to 1 (`min('/1',
'/2')`). That bit is the same for both values (remember,
192 = `1100 0000` and 128 = `1000 0000`), so we fall
through to comparing network size. `/2` is the larger of
the two, so `128.0.0.0/2` is deemed to be the larger
address.

## Designing an abbreviated key (#designing-keys)

!fig src="/assets/sortsupport-inet/key-design.svg" caption="The design of abbreviated keys for inet and cidr."

## Bit gymnastics in C (#gymnastics)

## Summary (#summary)

[inet]: src/include/utils/inet.h:23
[patch]: TODO
