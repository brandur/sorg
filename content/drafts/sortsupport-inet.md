---
title: "Designing SortSupport Abbreviated Keys for
  inet/cidr"
published_at: 2019-02-03T20:27:44Z
location: San Francisco
hook: TODO
tags: ["postgres"]
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

`inet` and `cidr` have some important subtleties in how
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
number between 0 and 32. It's easy to mentally a byte-sized
network out of a value (like `1.2.3.` out of `1.2.3.4/24`)
because you can just stop at the appropriate byte boundary,
but when it's not a nice byte multiple you have to think at
the binary level. For example, if I have the value
`255.255.255.255/1`, the network is just the leftmost bit.
255 in binary is `1111 1111`, so the network is the bit `1`
and the subnet is 31 consecutive `1`s.

!fig src="/assets/sortsupport-inet/inet-cidr-anatomy.svg" caption="The anatomy of inet and cidr values."

An address whose entire value is in the network (`/32` for
IPv4 or `/128` for IPv6) specifies just a single host, and
for display purposes the netmask size is usually omitted.
We'd show `1.2.3.4` instead of `1.2.3.4/32`.

The difference between `inet` is `cidr` is that `inet`
allows a values outside of the netmasked bits. The value
`1.2.3.4/24` is possible in `inet`, but illegal in `cidr`
because only zeroes may appear after the network like
`1.2.3.0/24`. They're nearly identical, with the latter
being strictly more constraining (and when working with
data, that's a good thing). Otherwise put, `cidr` values
never have a non-zero subnet.

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
   the network bits are equal, so we're in effect just
   comparing the subnet `.4`.

These rules combined with the fact that we're working at
the bit level produce ordering that in cases may not be
intuitively all that obvious to a human reader. For
example, `192.0.0.0/1` sorts *before* `128.0.0.0/2` despite
192 being the larger number. The reason is that when
comparing them, we start by looking at the common bits
available in both networks, which comes out to just one bit
(`min(/1, /2)`). That bit is the same in the networks
of both values (remember, 192 = `1100 0000` and 128 = `1000
0000`), so we fall through to comparing netmask size. `/2`
is the larger of the two, so `128.0.0.0/2` is deemed to be
the larger value.

## Designing an abbreviated key (#designing-keys)

Now that we understand the basics of `inet`/`cidr` and how
their sorting rules work, it's time to design an
abbreviated key for them. Remember that abbreviated keys
need to fit into the pointer-sized Postgres datum -- either
32 or 64 bits depending on target architecture. The goal is
to pack in as much sorting-relevant information as possible
while staying true to the existing sorting semantics.

Because of the subtle sorting semantics for these types,
we'll be breaking the available datum into multiple parts,
with information that we need for higher precedence sorting
rules occupying more significant bits so that it compares
first.

### 1 bit for family (#family)

The first part is easy: all IPv4 values always appear
before all IPv6 values. Since there's only two IP families,
so we'll reserve the most significant bit of our key to
represent a value's family. 0 for IPv4 and 1 for IPv6.

!fig src="/assets/sortsupport-inet/ip-family.svg" caption="One bit reserved for IP family."

At first glance it might seem short-sighted that we're
assuming that only two IP families will ever exist. Luckily
though, abbreviated keys are not persisted to disk (they're
only generated in the memory of a running Postgres system)
and their format is therefore non-binding. If a new IP
family were to ever appear, we could allocate another bit
to account for it.

### As many network bits as we can pack in (#network)

The next comparison that needs to be done is against a
value's network bits, so we should include those in the
datum.

The less obvious insight is that we can *only* include
network bits in this part. Think back to our example of
`192.0.0.0/1` and `128.0.0.0/2`: if we included 192's full
bits of `1100 0000`, then when comparing it to 128's `1000
0000`, it would sort higher when it needs to come out
lower. In order to guarantee our keys will comply with the
rules, we have to truncate values to just what appears in
the network.

Both `192.0.0.0/1` and `128.0.0.0/2` would appear as `1000
0000` (two of 128's bits were extracted, but it has a 0 in
the second position) and would appear equal when
considering this part of the abbreviated key. In cases
where that's all the space in the key we have to work with,
Postgres will have to fall back to authoritative comparison
(which would be able to move on and compare netmask size)
to break the tie.

The network bits are where we need to stop for most of our
use cases because that's all the space in the datum there
is. An IPv6 value is 128 bits -- after reserving 1 bit in
the datum for family, we have 31 bits left on a 32-bit
machine and 63 bits on a 64-bit machine, which will be
filled entirely with network. An IPv4 value is only 32
bits, but that's still more space than we have left on a
32-bit machine, so again, we'll pack in 31 of them.

The only case with space leftover is IPv4 on a 64-bit
machine. Even after storing all 32 possible bits of
network, there's still 31 bits available. Let's see what we
can use them for.

### IPv4 on 64-bit: network size and a few subnet bits (#ipv4-64bit)

As datums are being compared for IPv4 on a 64-bit machine,
we can be sure that that having looked at the 33 bits that
we've designed so far that IP family and available network
bits are equal. That lets us think about the next
comparison rule -- netmask size, which will fit nicely into
our datum. The largest possible netmask size for an IPv4
address is 32, which conveniently into only 6 bits [1] (`10
0000`).

After adding netmask size to the datum we're left with 25
bits, which we can use for subnet. Subnets can be as large
as 32 bits for a `/0` value, so we'll have to shift any
that are too large to fit down to the size available. That
will only ever happen for netmask sizes of `/6` or smaller
-- for all commonly seen netmask sizes like `/8`, `/16`,
or `/24` we can fit the entirety of the subnet into the
datum.

The final abbreviated key design looks like this:

!fig src="/assets/sortsupport-inet/key-design.svg" caption="The design of abbreviated keys for inet and cidr."

## Bit gymnastics in C (#gymnastics)

IP family:

``` c
static Datum
network_abbrev_convert(Datum original, SortSupport ssup)
{
    ...

    res = (Datum) 0;
    if (ip_family(authoritative) == PGSQL_AF_INET6)
    {
        /* Shift a 1 over to the datum's most significant bit. */
        res = ((Datum) 1) << (SIZEOF_DATUM * BITS_PER_BYTE - 1);
    }
```

### Ingesting bytes like an integer (#integer)

Ingesting 4 or 8 bytes of a value:

``` c
/*
 * Create an integer representation of the IP address by taking its first
 * 4 or 8 bytes. We take 8 bytes of an IPv6 address on a 64-bit machine
 * and 4 bytes on a 32-bit. Always take all 4 bytes of an IPv4 address.
 *
 * We're consuming an array of char, so make sure to byteswap on little
 * endian systems (an inet's IP array emulates big endian in that the
 * first byte is always the most significant).
 */
if (ip_family(authoritative) == PGSQL_AF_INET6)
{
    ipaddr_datum = *((Datum *) ip_addr(authoritative));
    ipaddr_datum = DatumBigEndianToNative(ipaddr_datum);
}
else
{
    uint32		ipaddr_datum32 = *((uint32 *) ip_addr(authoritative));
#ifndef WORDS_BIGENDIAN
    ipaddr_datum = pg_bswap32(ipaddr_datum32);
#endif
}
```

### Bit masking (#masking)

Separating network and subnet bits:

``` c
/*
 * Number of bits in subnet. e.g. An IPv4 that's /24 is 32 - 24 = 8.
 *
 * However, only some of the bits may have made it into the fixed sized
 * datum, so take the smallest number between bits in the subnet and bits
 * in the datum which are not part of the network.
 */
datum_subnet_size = Min(ip_maxbits(authoritative) - ip_bits(authoritative),
                        SIZEOF_DATUM * BITS_PER_BYTE - ip_bits(authoritative));

/* we may have ended up with < 0 for a large netmask size */
if (datum_subnet_size <= 0)
{
    /* the network occupies the entirety `ipaddr_datum` */
    network = ipaddr_datum;
    subnet = (Datum) 0;
}

...
```

The else case is more interesting:

``` c
...

else
{
    /*
     * This shift creates a power of two like `0010 0000`, and subtracts
     * one to create a bitmask for an IP's subnet bits like `0001 1111`.
     *
     * Note that `datum_subnet_mask` may be == 0, in which case we'll
     * generate a 0 bitmask and `subnet` will also come out as 0.
     */
    subnet_bitmask = (((Datum) 1) << datum_subnet_size) - 1;

    /* and likewise, use the mask's complement to get the netmask bits */
    network = ipaddr_datum & ~subnet_bitmask;

    /* bitwise AND the IP and bitmask to extract just the subnet bits */
    subnet = ipaddr_datum & subnet_bitmask;
}
```

Why are we suddenly not concerned with endianness now?

### Shifting things into place (#shifting)

IPv6 on 64-bit:

``` c
#if SIZEOF_DATUM == 8

if (ip_family(authoritative) == PGSQL_AF_INET6)
{
    /*
     * IPv6 on a 64-bit machine: keep the most significant 63 netmasked
     * bits.
     */
    res |= network >> 1;
}

...
```

This looks pretty similar to IPv4 or IPv6 on 32-bit:

``` c
...

#else /* SIZEOF_DATUM != 8 */

/*
 * 32-bit machine: keep the most significant 31 netmasked bits in both
 * IPv4 and IPv6.
 */
res |= network >> 1;

#endif
```

Refer to source for IPv4 on 64-bit.

## Summary (#summary)

[1] I originally thought that by subtracting one from 32 I
    could fit netmask size into only 5 bits (31 = `1
    1111`), but that's not possible because 0-bit netmasks
    are allowed and we therefore need to be able to
    represent the entire range of 0 to 32. For example,
    `1.2.3.4/0` is a legal value in Postgres.

[inet]: src/include/utils/inet.h:23
[patch]: TODO
