---
title: "SortSupport: Sorting in Postgres at Speed"
published_at: 2019-02-04T16:56:52Z
location: San Francisco
hook: How Postgres makes sorting really fast by comparing
  small, memory-friendly abbreviated keys as proxies for
  arbitrarily large values on the heap.
tags: ["postgres"]
---

Most often, there's a trade off involved in optimizing
software. The cost of better performance is the opportunity
cost of the time that it took to write the optimization,
and the additional cost of maintenance for code that
becomes more complex and more difficult to understand.

Many projects prioritize product development over improving
runtime speed. Time is spent building new things instead of
making existing things faster. Code is kept simpler and
easier to understand so that adding new features and fixing
bugs stays easy, even as particular people rotate in and
out and institutional knowledge is lost.

But that's certainly not the case in all domains. Game code
is often an interesting read because it comes from an
industry where speed is a competitive advantage, and it's
common practice to optimize liberally even at some cost to
modularity and maintainability. One technique for that is
to inline code in critical sections even to the point of
absurdity. CryEngine, open-sourced a few years ago, has a
few examples of this, with ["tick" functions like this
one][cryengine] that are 800+ lines long with 14 levels of
indentation.

Another common place to find optimizations is in databases.
While games optimize because they have to, databases
optimize because they're an example of software that's
extremely leveraged -- if there's a way to make running
select queries or building indexes 10% faster, it's not an
improvement that affects just a couple users, it's one
that'll potentially invigorate millions of installations
around the world. That's enough of an advantage that the
enhancement is very often worth it, even if the price is a
challenging implementation or some additional code
complexity.

Postgres contains a wide breadth of optimizations, and
happily they've been written conscientiously so that the
source code stays readable. The one that we'll look at
today is **SortSupport**, a technique for localizing the
information needed to compare data into places where it can
be accessed very quickly, thereby making sorting data much
faster. Sorting for types that have had Sortsupport
implemented usually gets twice as fast or more, a speedup
that transfers directly into common database operations
like `ORDER BY`, `DISTINCT`, and `CREATE INDEX`.

## Sorting with abbreviated keys (#abbreviated-keys)

While sorting, Postgres builds a series of tiny structures
that represent the data set being sorted. These tuples have
space for a value the size of a native pointer (i.e. 64
bits on a 64-bit machine) which is enough to fit the
entirety of some common types like booleans or integers
(known as pass-by-value types), but not for others that are
larger than 64 bits or arbitrarily large. In their case,
Postgres will follow a references back to the heap when
comparing values (they're appropriately called
pass-by-reference types). Postgres is very fast, so that
still happens quickly, but it's slower than comparing
values readily available in memory.

!fig src="/assets/sortsupport/sort-tuples.svg" caption="An array of sort tuples."

SortSupport augments pass-by-reference types by bringing a
representative part of their value into the sort tuple to
save trips to the heap. Because sort tuples usually don't
have the space to store the entirety of the value,
SortSupport generates a digest of the full value called an
**abbreviated key**, and stores it instead. The contents of
an abbreviated key vary by type, but they'll aim to store
as much sorting-relevant information as possible while
remaining faithful to pre-existing sorting rules.

Abbreviated keys should never produce an incorrect
comparison, but it's okay if they can't fully resolve one.
If two abbreviated keys look equal, Postgres will fall back
to comparing their full heap values to make sure it gets
the right result (called an "authoritative comparison").

!fig src="/assets/sortsupport/abbreviated-keys.svg" caption="A sort tuple with an abbreviated key and pointer to the heap."

Implementing an abbreviated key is straightforward in many
cases. UUIDs are a good example of that: at 128 bits long
they're always larger than the pointer size even on a
64-bit machine, but we can get a very good proxy of their
full value just by sampling their first 64 bits (or 32 on a
32-bit machine). Especially for V4 UUIDs which are almost
entirely random [1], the first 64 bits will be enough to
definitively determine the order for all but unimaginably
large data sets. Indeed, [the patch that brought in
SortSupport for UUIDs][uuidpatch] made sorting them about
twice as fast!

String-like types (e.g. `text`, `varchar`) aren't too much
harder: just pack as many characters from the front of the
string in as possible (although made somewhat more
complicated by locales). Adding SortSupport for them made
operations like `CREATE INDEX` [about three times
faster][textblog]. My only ever patch to Postgres was
implementing SortSupport for the `macaddr` type, which was
fairly easy because although it's pass-by-reference, its
values are only six bytes long [2]. On a 64-bit machine we
have room for all six bytes, and on 32-bit we sample the
MAC address' first four bytes.

Some abbreviated keys are more complex. The implementation
for the `numeric` type, which allows arbitrary scale and
precision, involves [excess-K coding][excessk] and breaking
available bits into multiple parts to store sort-relevant
fields.

## A glance at the implementation (#implementation)

Let's try to get a basic idea of how SortSupport is
implemented by examining a narrow slice of source code.
Sorting in Postgres is extremely complex and involves
thousands of lines of code, so fair warning that I'm going
to simplify some things and skip *a lot* of others.

A good place start is with `Datum`, the pointer-sized type
(32 or 64 bits, depending on the CPU) used for sort
comparisons. It stores entire values for pass-by-value
types, abbreviated keys for pass-by-reference types that
implement SortSupport, and a pointer for those that don't.
You can find it defined in [`postgres.h`][datum]:

``` c
/*
 * A Datum contains either a value of a pass-by-value type or a pointer
 * to a value of a pass-by-reference type.  Therefore, we require:
 *
 * sizeof(Datum) == sizeof(void *) == 4 or 8
 */

typedef uintptr_t Datum;

#define SIZEOF_DATUM SIZEOF_VOID_P
```

### Building abbreviated keys for UUID (#uuid)

The format of abbreviated keys for the `uuid` type is one
of the easiest to understand, so let's look at that. In
Postgres, the struct `pg_uuid_t` defines how UUIDs are
physically stored in the heap (from [`uuid.h`][uuid]):

``` c
/* uuid size in bytes */
#define UUID_LEN 16

typedef struct pg_uuid_t
{
    unsigned char data[UUID_LEN];
} pg_uuid_t;
```

You might be used to seeing UUIDs represented in string
format like `123e4567-e89b-12d3-a456-426655440000`, but
remember that this is Postgres which likes to be as
efficient as possible! A UUID contains 16 bytes worth of
information, so `pg_uuid_t` above defines an array of
exactly 16 bytes. No wastefulness to be found.

SortSupport implementations define a conversion routine
which takes the original value and produces a datum
containing an abbreviated key. Here's the one for UUIDs
(from [`uuid.c`][uuidconvert]):

``` c
static Datum
uuid_abbrev_convert(Datum original, SortSupport ssup)
{
    pg_uuid_t *authoritative = DatumGetUUIDP(original);
    Datum      res;

    memcpy(&res, authoritative->data, sizeof(Datum));

    ...

    /*
     * Byteswap on little-endian machines.
     *
     * This is needed so that uuid_cmp_abbrev() (an unsigned integer 3-way
     * comparator) works correctly on all platforms.  If we didn't do this,
     * the comparator would have to call memcmp() with a pair of pointers to
     * the first byte of each abbreviated key, which is slower.
     */
    res = DatumBigEndianToNative(res);

    return res;
}
```

`memcpy` ("memory copy") extracts a datum worth of bytes
from a `pg_uuid_t` and places it into `res`. We can't take
the whole UUID, but we'll be taking its 4 or 8 most
significant bytes, which will be enough information for
most comparisons.

!fig src="/assets/sortsupport/uuid.svg" caption="Abbreviated key formats for the `uuid` type."

The call `DatumBigEndianToNative` is there to help with an
optimization. When comparing our abbreviated keys, we could
do so with `memcmp` ("memory compare")  which would compare
each byte in the datum one at a time. That's perfectly
functional of course, but because our datums are the same
size as native integers, we can instead choose to take
advantage of the fact that CPUs are optimized to compare
integers really, really quickly, and arrange the datums in
memory as if they were integers. You can see this integer
comparison taking place in the UUID abbreviated key
comparison function:

``` c
static int
uuid_cmp_abbrev(Datum x, Datum y, SortSupport ssup)
{
    if (x > y)
        return 1;
    else if (x == y)
        return 0;
    else
        return -1;
}
```

However, pretending that some consecutive bytes in memory
are integers introduces some complication. Integers might
be stored like `data` in `pg_uuid_t` with the most
significant byte first, but that depends on the
architecture of the CPU. We call architectures that store
numerical values this way **big-endian**. Big-endian
machines exist, but the chances are that the CPU you're
using to read this article stores bytes in the reverse
order of their significance, with the most significant at
the highest address. This layout is called
**little-endian**, and is in use by Intel's X86, as well as
being the default mode for ARM chips like the ones in
Android and iOS devices.

If we left the big-endian result of the `memcpy` unchanged
on little-endian systems, the resulting integer would be
wrong. The answer is to byteswap, which reverses the order
of the bytes, and corrects the integer.

!fig src="/assets/sortsupport/endianness.svg" caption="Example placement of integer bytes on little and big endian architectures."

You can see in [`pg_bswap.h`][pgbswap] that
`DatumBigEndianToNative` is defined as a no-op on a
big-endian machine, and is otherwise connected to a
byteswap ("bswap") routine of the appropriate size:

``` c
#ifdef WORDS_BIGENDIAN

        #define        DatumBigEndianToNative(x)    (x)

#else

    #if SIZEOF_DATUM == 8
        #define        DatumBigEndianToNative(x)    pg_bswap64(x)
    #else
        #define        DatumBigEndianToNative(x)    pg_bswap32(x)
    #endif

#endif
```

#### Conversion abort & HyperLogLog (#abort)

Let's touch upon one more feature of `uuid_abbrev_convert`.
In data sets with very low cardinality (i.e, many
duplicated items) SortSupport introduces some danger of
worsening performance. With so many duplicates, the
contents of abbreviated keys would often show equality, in
which cases Postgres would often have to fall back to the
authoritative comparator. In effect, by adding SortSupport
we would have added a useless additional comparison that
wasn't there before.

To protect against performance regression, SortSupport has
a mechanism for aborting abbreviated key conversion. If the
data set is found to be below a certain cardinality
threshold, Postgres stops abbreviating, reverts any keys
that were already abbreviated, and disables further
abbreviation for the sort.

Cardinality is estimated with the help of
[HyperLogLog][hyperloglog], an algorithm that estimates the
distinct count of a data set in a very memory-efficient
way. Here you can see the conversion routine adding new
values to the HyperLogLog if an abort is still possible:

``` c
uss->input_count += 1;

if (uss->estimating)
{
    uint32        tmp;

#if SIZEOF_DATUM == 8
    tmp = (uint32) res ^ (uint32) ((uint64) res >> 32);
#else
    tmp = (uint32) res;
#endif

    addHyperLogLog(&uss->abbr_card, DatumGetUInt32(hash_uint32(tmp)));
}
```

And where it makes an abort decision (from [`uuid.c`][uuidabort]):

``` c
static bool
uuid_abbrev_abort(int memtupcount, SortSupport ssup)
{
    ...

    abbr_card = estimateHyperLogLog(&uss->abbr_card);

    /*
     * If we have >100k distinct values, then even if we were
     * sorting many billion rows we'd likely still break even,
     * and the penalty of undoing that many rows of abbrevs would
     * probably not be worth it. Stop even counting at that point.
     */
    if (abbr_card > 100000.0)
    {
        uss->estimating = false;
        return false;
    }

    /*
     * Target minimum cardinality is 1 per ~2k of non-null inputs.
     * 0.5 row fudge factor allows us to abort earlier on genuinely
     * pathological data where we've had exactly one abbreviated
     * value in the first 2k (non-null) rows.
     */
    if (abbr_card < uss->input_count / 2000.0 + 0.5)
    {
        return true;
    }

    ...
}
```

It also covers aborting the case where we have a data set
that's poorly suited to the abbreviated key format. For
example, imagine a million UUIDs that all shared a common
prefix in their first eight bytes, but were distinct in
their last eight [3]. Realistically this will be extremely
unusual, so abbreviated key conversion will rarely abort.

### Tuples and data types (#tuples)

**Sort tuples** are the tiny structures that Postgres sorts
in memory. They hold a reference to the "true" tuple, a
datum, and a flag to indicate whether or not the first
value is `NULL` (which has its own special sorting
semantics). The latter two are named with a `1` suffix as
`datum1` and `isnull1` because they represent only one
field worth of information. Postgres will need to fall back
to different values in the event of equality in a
multi-column comparison. From [`tuplesort.c`][sorttuple]:

``` c
/*
 * The objects we actually sort are SortTuple structs.  These contain
 * a pointer to the tuple proper (might be a MinimalTuple or IndexTuple),
 * which is a separate palloc chunk --- we assume it is just one chunk and
 * can be freed by a simple pfree() (except during merge, when we use a
 * simple slab allocator).  SortTuples also contain the tuple's first key
 * column in Datum/nullflag format, and an index integer.
 */
typedef struct
{
    void       *tuple;          /* the tuple itself */
    Datum       datum1;         /* value of first key column */
    bool        isnull1;        /* is first key column NULL? */
    int         tupindex;       /* see notes above */
} SortTuple;
```

In the code we'll look at below, `SortTuple` may reference
a **heap tuple**, which has a variety of different struct
representations. One used by the sort algorithm is
`HeapTupleHeaderData` (from [`htup_details.h`][heaptuple]):

``` c
struct HeapTupleHeaderData
{
    union
    {
        HeapTupleFields t_heap;
        DatumTupleFields t_datum;
    }            t_choice;

    ItemPointerData t_ctid; /* current TID of this or newer tuple (or a
                             * speculative insertion token) */

    ...
}
```

Heap tuples have a pretty complex structure which we won't
cover, but you can see that it contains an
`ItemPointerData` value. This struct is what gives Postgres
the precise information it needs to find data in the heap
(from [`itemptr.h`][itempointer]):

``` c
/*
 * ItemPointer:
 *
 * This is a pointer to an item within a disk page of a known file
 * (for example, a cross-link from an index to its parent table).
 * blkid tells us which block, posid tells us which entry in the linp
 * (ItemIdData) array we want.
 */
typedef struct ItemPointerData
{
    BlockIdData ip_blkid;
    OffsetNumber ip_posid;
}
```

### Tuple comparison (#comparison)

The algorithm to compare abbreviated keys is duplicated in
the Postgres source in a number of places depending on the
sort operation being carried out. We'll take a look at
`comparetup_heap` (from [`tuplesort.c`][comparetup]) which
is used when sorting based on the heap. This would be
invoked for example if you ran an `ORDER BY` on a field
that doesn't have an index on it.

``` c
static int
comparetup_heap(const SortTuple *a, const SortTuple *b, Tuplesortstate *state)
{
    SortSupport sortKey = state->sortKeys;
    HeapTupleData ltup;
    HeapTupleData rtup;
    TupleDesc     tupDesc;
    int           nkey;
    int32         compare;
    AttrNumber    attno;
    Datum         datum1,
                  datum2;
    bool          isnull1,
                  isnull2;


    /* Compare the leading sort key */
    compare = ApplySortComparator(a->datum1, a->isnull1,
                                  b->datum1, b->isnull1,
                                  sortKey);
    if (compare != 0)
        return compare;
```

`ApplySortComparator` gets a comparison result between two
datum values. It'll compare two abbreviated keys where
appropriate and handles `NULL` sorting semantics. The
return value of a comparison follows the spirit of C's
`strcmp`: when comparing `(a, b)`, -1 indicates `a < b`,
0 indicates equality, and 1 indicates `a > b`.

The algorithm returns immediately if inequality (`!= 0`)
was detected. Otherwise, it checks to see if abbreviated
keys were used, and if so applies the authoritative
comparison if they were. Because space in abbreviated keys
is limited, two being equal doesn't necessarily indicate
that the values that they represent are.

``` c
if (sortKey->abbrev_converter)
{
    attno = sortKey->ssup_attno;

    datum1 = heap_getattr(&ltup, attno, tupDesc, &isnull1);
    datum2 = heap_getattr(&rtup, attno, tupDesc, &isnull2);

    compare = ApplySortAbbrevFullComparator(datum1, isnull1,
                                            datum2, isnull2,
                                            sortKey);
    if (compare != 0)
        return compare;
}
```

Once again, the algorithm returns if inequality was
detected. If not, it starts to look beyond the first field
(in the case of a multi-column sort):

``` c
    ...

    sortKey++;
    for (nkey = 1; nkey < state->nKeys; nkey++, sortKey++)
    {
        attno = sortKey->ssup_attno;

        datum1 = heap_getattr(&ltup, attno, tupDesc, &isnull1);
        datum2 = heap_getattr(&rtup, attno, tupDesc, &isnull2);

        compare = ApplySortComparator(datum1, isnull1,
                                      datum2, isnull2,
                                      sortKey);
        if (compare != 0)
            return compare;
    }

    return 0;
}
```

After finding abbreviated keys to be equal, full values to
be equal, and all additional sort fields to be equal, the
last step is to `return 0`, indicating in classic libc
style that the two tuples are really, fully equal.

## Fast code and leveraged software (#leverage)

SortSupport is a good example of the type of low-level
optimization that most of us probably wouldn't bother with
in our projects, but which makes sense in an extremely
leveraged system like a database. As implementations are
added for it and Postgres' tens of thousands of users like
myself upgrade, common operations like `DISTINCT`, `ORDER
BY`, and `CREATE INDEX` get twice as fast, for free.

Credit to Peter Geoghegan for some of the original
exploration of this idea and implementations for UUID and a
generalized system for SortSupport on variable-length
string types, Robert Haas and Tom Lane for adding the
[necessary infrastructure][infrastructure], and Andrew
Gierth for a [difficult implementation][numeric] for
`numeric`. (I hope I got all that right.)

[1] A note for the pedantic that V4 UUIDs usually have only
    122 bits of randomness as four bits are used for the
    version and two for the variant.

[2] The new type `macaddr8` was later introduced to handle
    EUI-64 MAC addresses, which are 64 bits long.

[3] A data set of UUIDs with common datum-sized prefixes is
    a pretty unlikely scenario, but it's a little more
    realistic for variable-length string types, where users
    are storing much more free-form data.

[comparetup]: https://github.com/postgres/postgres/blob/08ecdfe7e5e0a31efbe1d58fefbe085b53bc79ca/src/backend/utils/sort/tuplesort.c#L3508
[cryengine]: https://github.com/CRYTEK/CRYENGINE/blob/release/Code/CryEngine/CryPhysics/livingentity.cpp#L1275
[datum]: https://github.com/postgres/postgres/blob/08ecdfe7e5e0a31efbe1d58fefbe085b53bc79ca/src/include/postgres.h#L367
[excessk]: https://en.wikipedia.org/wiki/Offset_binary
[heaptuple]: https://github.com/postgres/postgres/blob/08ecdfe7e5e0a31efbe1d58fefbe085b53bc79ca/src/include/access/htup_details.h#L152
[hyperloglog]: https://en.wikipedia.org/wiki/HyperLogLog
[infrastructure]: https://git.postgresql.org/gitweb/?p=postgresql.git;a=commit;h=c6e3ac11b60ac4a8942ab964252d51c1c0bd8845
[itempointer]: https://github.com/postgres/postgres/blob/08ecdfe7e5e0a31efbe1d58fefbe085b53bc79ca/src/include/storage/itemptr.h#L36
[numeric]: https://git.postgresql.org/gitweb/?p=postgresql.git;a=commit;h=abd94bcac4582903765be7be959d1dbc121df0d0
[pgbswap]: https://github.com/postgres/postgres/blob/08ecdfe7e5e0a31efbe1d58fefbe085b53bc79ca/src/include/port/pg_bswap.h#L143
[sorttuple]: https://github.com/postgres/postgres/blob/08ecdfe7e5e0a31efbe1d58fefbe085b53bc79ca/src/backend/utils/sort/tuplesort.c#L169
[textblog]: http://pgeoghegan.blogspot.com/2015/01/abbreviated-keys-exploiting-locality-to.html
[uuid]: https://github.com/postgres/postgres/blob/08ecdfe7e5e0a31efbe1d58fefbe085b53bc79ca/src/include/utils/uuid.h#L20
[uuidabort]: https://github.com/postgres/postgres/blob/08ecdfe7e5e0a31efbe1d58fefbe085b53bc79ca/src/backend/utils/adt/uuid.c#L301
[uuidconvert]: https://github.com/postgres/postgres/blob/08ecdfe7e5e0a31efbe1d58fefbe085b53bc79ca/src/backend/utils/adt/uuid.c#L367
[uuidpatch]: https://www.postgresql.org/message-id/CAM3SWZR4avsTwwNVUzRNbHk8v36W-QBqpoKg%3DOGkWWy0dKtWBA%40mail.gmail.com
[varstrconvert]: https://github.com/postgres/postgres/blob/08ecdfe7e5e0a31efbe1d58fefbe085b53bc79ca/src/backend/utils/adt/varlena.c#L2373
