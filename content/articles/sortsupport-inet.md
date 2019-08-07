+++
hook = "Making the sorting speed of network types in Postgres twice as fast by designing SortSupport abbreviated keys compatible with their existing sort semantics."
location = "Vancouver"
published_at = 2019-08-07T16:50:44Z
tags = ["postgres"]
title = "Doubling the Sorting Speed of Postgres Network Types with Abbreviated Keys"
+++

A few months ago, I wrote about [how SortSupport works in
Postgres](/sortsupport) to vastly speed up sorting on
large data types [1] like `numeric` or `text`, and
`varchar`. It works by generating **abbreviated keys** for
values that are representative of them for purposes of
sorting, but which fit nicely into the pointer-sized value
(called a "**datum**") in memory that Postgres uses for
sorting. Most values can be sorted just based on their
abbreviated key, saving trips to the heap and increasing
sorting throughput. Faster sorting leads to speedup on
common operations like `DISTINCT`, `ORDER BY`, and `CREATE
INDEX`.

A [patch][patch] of mine was recently committed to add
SortSupport for the `inet` and `cidr` types, which by my
measurement, a little more than doubles sorting speed on
them. `inet` and `cidr` are the types used to store network
addresses or individual hosts and in either IPv4 or IPv6
(they generally look something like `1.2.3.0/24` or
`1.2.3.4`).

`inet` and `cidr` have some important subtleties in how
they're sorted which made designing an abbreviated key that
would be faithful to those subtleties but still efficient,
a non-trivial problem. Because their size is limited,
abbreviated keys are allowed to show equality even for
values that aren't equal (Postgres will fall back to
authoritative comparison to confirm equality or tiebreak),
but they should never falsely indicate inequality.

## Network type anatomy, and inet vs. cidr (#inet-cidr)

A property that's not necessarily obvious to anyone
unfamiliar with them is that network types (`inet` or
`cidr`) can either address a single host (what most people
are used to seeing) or an entire subnetwork of arbitrary
size. For example:

* `1.2.3.4/32` specifies a 32-bit netmask on an IPv4 value,
  which is 32 bits wide, which means that it defines
  exactly one address: `1.2.3.4`. `/128` would work
  similarly for IPv6.

* `1.2.3.0/24` specifies a 24-bit netmask. It identifies
  the network at `1.2.3.*`. The last byte may be anywhere
  in the range of 0 to 255.

* Similarly, `1.0.0.0/8` specifies an 8-bit netmask. It
  identifies the much larger possible network at `1.*`.

We'll establish the following common vocabulary for each
component of an address (and take for example the value
`1.2.3.4/24`):

1. A **network**, or bits in the netmask (`1.2.3.`).
2. A **netmask size** (`/24` which is 24 bits). Dictates
   the number of bits in the network.
3. A **subnet**, or bits outside of the netmask (`.4`).
   Only `inet` carries non-zero bits here, and combined
   with the network, they identify a single **host**
   (`1.2.3.4`).

The netmask size is a little more complex than commonly
understood because while it's most common to see byte-sized
blocks like `/8`, `/16`, `/24`, and `/32`, it's allowed to
be any number between 0 and 32. It's easy to mentally
extract a byte-sized network out of a value (like `1.2.3.`
out of `1.2.3.4/24`) because you can just stop at the
appropriate byte boundary, but when it's not a nice byte
multiple you have to think at the binary level. For
example, if I have the value `255.255.255.255/1`, the
network is just the leftmost bit. 255 in binary is `1111
1111`, so the network is the bit `1` and the subnet is 31
consecutive `1`s.

!fig src="/assets/images/sortsupport-inet/inet-cidr-anatomy.svg" caption="The anatomy of inet and cidr values."

The difference between `inet` is `cidr` is that `inet`
allows a values outside of the netmasked bits. The value
`1.2.3.4/24` is possible in `inet`, but illegal in `cidr`
because only zeroes may appear after the network like
`1.2.3.0/24`. They're nearly identical, with the latter
being more strict.

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

In Postgres, `inet`/`cidr` sort according to these rules:

1. IPv4 always appears before IPv6.
2. The bits in the network are compared (`1.2.3.`).
3. Netmask size is compared (`/24`).
4. All bits are compared. Having made it here, we know that
   the network bits are equal, so we're in effect just
   comparing the subnet (`.4`).

These rules combined with the fact that we're working at
the bit level produce ordering that in cases may not be
intuitively. For example, `192.0.0.0/1` sorts *before*
`128.0.0.0/2` despite 192 being the larger number -- when
comparing them, we start by looking at the common bits
available in both networks, which comes out to just one bit
(`min(/1, /2)`). That bit is the same in the networks of
both values (remember, 192 = `1100 0000` and 128 = `1000
0000`), so we fall through to comparing netmask size. `/2`
is the larger of the two, so `128.0.0.0/2` is the larger
value.

## Designing an abbreviated key (#designing-keys)

Armed with the structure of `inet`/`cidr` and how their
sorting works, we can now design an abbreviated key for
them. Remember that abbreviated keys need to fit into the
pointer-sized Postgres datum -- either 32 or 64 bits
depending on target architecture. The goal is to pack in as
much sorting-relevant information as possible while staying
true to existing semantics.

We'll be breaking the available datum into multiple parts,
with information that we need for higher precedence sorting
rules occupying more significant bits so that it compares
first. This allows us to compare any two keys as integers
-- a very fast operation for CPUs (faster even than
comparing memory byte-by-byte), and also a common technique
in other abbreviated key implementations like the one for
[UUIDs][uuid].

### 1 bit for family (#family)

The first part is easy: all IPv4 values always appear
before all IPv6 values. Since there's only two IP families,
so we'll reserve the most significant bit of our key to
represent a value's family. 0 for IPv4 and 1 for IPv6.

!fig src="/assets/images/sortsupport-inet/ip-family.svg" caption="One bit reserved for IP family."

It might seem short-sighted that we're assuming that only
two IP families will ever exist, but luckily abbreviated
keys are not persisted to disk (only in the memory of a
running Postgres system) and their format is therefore
non-binding. If a new IP family were to ever appear, we
could allocate another bit to account for it.

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

!fig src="/assets/images/sortsupport-inet/network-bits.svg" caption="Number of bits available to store network per datum size and IP family."

But there is one case where we have some space left over:
IPv4 on a 64-bit machine. Even after storing all 32
possible bits of network, there's still 31 bits available.
Let's see what we can use them for.

### IPv4 on 64-bit: network size and a few subnet bits (#ipv4-64bit)

As datums are being compared for IPv4 on a 64-bit machine,
we can be sure that having looked at the 33 bits that
we've designed so far -- IP family (1 bit) and 
network (32 bits) -- are equal. That leaves us with 31
bits (64 - 33) left to work with, and lets us move onto
the next comparison rule -- netmask size. The largest
possible netmask size for an IPv4 address is 32, which
conveniently fits into only 6 bits (`32 = 10 0000`) [2].

After adding netmask size to the datum we're left with 25
bits (31 - 6), which we can use for the next sorting rule
-- subnet. Subnets can be as large as 32 bits for a `/0`
value, so we'll have to shift any that are too large to fit
down to the size available. That will only ever happen for
netmask sizes of `/6` or smaller -- for all commonly seen
netmask sizes like `/8`, `/16`, or `/24` we can fit the
entirety of the subnet into the datum.

With subnet covered, we've used up all the available key
bits, but also managed to cover every sorting rule -- with
most [3] real-world data, Postgres should be able to sort
almost entirely with abbreviated keys without falling back
to authoritative comparison. The final key design looks
like this:

!fig src="/assets/images/sortsupport-inet/key-design.svg" caption="The design of abbreviated keys for inet and cidr."

## Bit gymnastics in C (#gymnastics)

Now that we have an encoding scheme for each different
case, we can build an implementation that puts everything
into place. This involves the use of many bitwise
operations that are common in C, but which many of us who
program in high-level languages day-to-day aren't as used
to.

I'll go through this implementation step-by-step, but you
may prefer to refer to the completed version in the
[Postgres source][source], which we've made an effort to
comment comprehensively.

### Ingesting bytes as an integer (#integer)

Recall that an IP component is stored as a 16-byte
`unsigned char` array in the backing network type:

``` c
typedef struct
{
    ...
    unsigned char ipaddr[16];  /* up to 128 bits of address */
} inet_struct;
```

Our abbreviated keys will be compared as if they were
integers (one of the reasons that they're so fast), so the
first step is to extract a datum's worth of bytes from
`ipaddr` into an intermediate representation that'll be
used to more easily separate out the final components.
We'll use `memcpy` to copy it out byte-by-byte:

``` c
Datum ipaddr_datum;
memcpy(&ipaddr_datum, ip_addr(authoritative), sizeof(Datum));
```

`ipaddr` is laid out most significant byte first, which
will be fine when representing an integer on a big-endian
machine, but no good on one that's little-endian (like most
of our Intel processors), so do a byte-wise position swap
to re-form it (more detail on this talking about [`uuid`'s
abbreviated key implementation][uuid]:

``` c
/* Must byteswap on little-endian machines */
ipaddr_datum = DatumBigEndianToNative(ipaddr_datum);
```

And for IPv6, make sure to shift a 1 bit into the leftmost
position so that it sorts after all IPv4 values:

```
Datum res;
res = ((Datum) 1) << (SIZEOF_DATUM * BITS_PER_BYTE - 1);
```

### Extracting network via bitmask (#network-bitmask)

Next we'll extract the leading **network** component using a
technique called bitmasking. This common technique involves
using a bitwise-AND to extract a desired range of bits:

```
  1010 1010 1010 1010       (original value)
& 0000 1111 1111 0000       (bitmask)
  -------------------
  0000 1010 1010 0000       (final result)
```

We're going to create a bitmask for the **subnet** portion
of the value (reminder: that's the last part _after_ the
network), and it's size depends on how many subnet bits we
expect to see in `ipaddr_datum`. For example, if the
network component occupies bits equal or greater to the
datum's size, then the subnet bitmask will be zero.

The code's broken into three separate conditionals. This
first section handles the case of no bits in the network
components. The subnet bitmask should be all ones, which we
get by starting with 0, subtracting 1, and allowing the
value to roll over to its maximum value:

``` c
Datum subnet_bitmask,
      network;

subnet_size = ip_maxbits(authoritative) - ip_bits(authoritative);
Assert(subnet_size >= 0);

if (ip_bits(authoritative) == 0)
{
    /* Fit as many ipaddr bits as possible into subnet */
    subnet_bitmask = ((Datum) 0) - 1;
    network = 0;
}
```

The next section is the case where there are some bits for
both the network and subnet. We use a trick to get the
bitmask which involves shifting a 1 left out by the subnet
size, then subtracting one to get 1s in all positions that
were right of it:

```
  0000 0001 0000 0000       (1 << 8)
-                   1       (minus one)
  -------------------
  0000 0000 1111 1111       (8-bit mask)
```

Getting the network's value then involves ANDing the IP's
datum and the _negated_ form of the subnet bitmask
(`ipaddr_datum & ~subnet_bitmask`):

``` c
else if (ip_bits(authoritative) < SIZEOF_DATUM * BITS_PER_BYTE)
{
    /* Split ipaddr bits between network and subnet */
    subnet_bitmask = (((Datum) 1) << subnet_size) - 1;
    network = ipaddr_datum & ~subnet_bitmask;
}
```

The final case represents no bits in the subnet. Set
`network` to the full value of `ipaddr_datum`:

``` c
else
{
    /* Fit as many ipaddr bits as possible into network */
    subnet_bitmask = 0;        /* Unused, but be tidy */
    network = ipaddr_datum;
}
```

### Shifting things into place for IPv4 on 64-bit (#shifting)

Recall that IPv4 on a 64-bit architecture is by far the
most complex case because we have room to fit a lot more
information. This next section involves taking the network
and subnet bitmask that we resolved above and shifting it
all into place.

The order of operations is:

1. `network`: Shift the network left 31 bits to make room
   for netmask size and 25 bits worth of subnet.
2. `network_size`: Shift the network size left 25 bits to
   make room for the subnet.
3. `subnet`: Extract a subnet using the bitmask calculated
   above.
4. `subnet`: If the subnet is longer than 25 bits, shift it
   down to just occupy 25 bits.
5. `res`: Get a final result by ORing the values from (1),
   (2), and (4) above.

``` c
#if SIZEOF_DATUM == 8
    if (ip_family(authoritative) == PGSQL_AF_INET)
    {
        /*
         * IPv4 with 8 byte datums: keep all 32 netmasked bits, netmask size,
         * and most significant 25 subnet bits
         */
        Datum        netmask_size = (Datum) ip_bits(authoritative);
        Datum        subnet;

        /* Shift left 31 bits: 6 bits netmask size + 25 subnet bits */
        network <<= (ABBREV_BITS_INET4_NETMASK_SIZE +
                     ABBREV_BITS_INET4_SUBNET);

        /* Shift size to make room for subnet bits at the end */
        netmask_size <<= ABBREV_BITS_INET4_SUBNET;

        /* Extract subnet bits without shifting them */
        subnet = ipaddr_datum & subnet_bitmask;

        /*
         * If we have more than 25 subnet bits, we can't fit everything. Shift
         * subnet down to avoid clobbering bits that are only supposed to be
         * used for netmask_size.
         *
         * Discarding the least significant subnet bits like this is correct
         * because abbreviated comparisons that are resolved at the subnet
         * level must have had equal subnet sizes in order to get that far.
         */
        if (subnet_size > ABBREV_BITS_INET4_SUBNET)
            subnet >>= subnet_size - ABBREV_BITS_INET4_SUBNET;

        /*
         * Assemble the final abbreviated key without clobbering the ipfamily
         * bit that must remain a zero.
         */
        res |= network | netmask_size | subnet;
    }
    else
#endif
```

### Everything else (#everything-else)

The three other cases (refer to the figure above) are much
simpler because we only have room for network bits. Shift
them right by 1 bit to not clobber our previously set IP
family, then OR with `res` for the final result:

``` c
#endif
    {
        /*
         * 4 byte datums, or IPv6 with 8 byte datums: Use as many of the
         * netmasked bits as will fit in final abbreviated key. Avoid
         * clobbering the ipfamily bit that was set earlier.
         */
        res |= network >> 1;
    }
```

## Speed vs. sustainability (#speed-vs-sustainability)

The abbreviated key implementation here is complex enough
that in most contexts I'd probably consider it a poor trade
off -- added speed is nice to have, but there is a cost in
the ongoing maintenance burden of the new code and its
understandability by future contributors.

However, Postgres is a highly leveraged piece of software.
This patch makes sorting and creating indexes on network
types _~twice as fast_, and that improvement will trickle
down automatically to hundreds of thousands of Postgres
installations around the world as they're upgraded to the
next major version. If there's one place where trading some
more complexity for speed is worth it, it's cases like this
one where only very few have to understand the code, but
very many will reap its benefits. We've also made sure to
add extensive comments and test cases to keep future code
changes as easy as they can be.

Thanks to Peter Geoghegan for seeding the idea for this
patch, as well as for advice and very thorough
testing/review, and Edmund Horner for review.

[1] Technically, pass-by-reference types. Generally those
    that can't fit their entire value in a datum.

[2] I originally thought that by subtracting one from 32 I
    could fit netmask size into only 5 bits (31 = `1
    1111`), but that's not possible because 0-bit netmasks
    are allowed and we therefore need to be able to
    represent the entire range of 0 to 32. For example,
    `1.2.3.4/0` is a legal value in Postgres.

[3] Authoritative comparison will still be needed in the
    case of equal network values and values with short
    networks (`/6` or less) that share many leading bits.

[inet]: https://github.com/postgres/postgres/blob/12afc7145c03c212f26fea3a99e016da6a1c919c/src/include/utils/inet.h:23
[patch]: https://www.postgresql.org/message-id/CABR_9B-PQ8o2MZNJ88wo6r-NxW2EFG70M96Wmcgf99G6HUQ3sw%40mail.gmail.com
[source]: https://github.com/postgres/postgres/blob/12afc7145c03c212f26fea3a99e016da6a1c919c/src/backend/utils/adt/network.c#L561
[uuid]: /sortsupport#uuid
