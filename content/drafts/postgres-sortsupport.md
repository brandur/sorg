---
title: "SortSupport: Sorting in Postgres Postgres at Speed"
published_at: 2019-01-24T19:19:18Z
location: San Francisco
hook: TODO
---

Most often, there's a trade off involved in optimizing
software. The cost of better performance is the time that
it takes to an optimization in place, and the additional
cost of maintenance for code that becomes more complex and
more difficult to understand.

Very often in the business world it's common to optimize
for product development velocity instead of runtime speed.
Time is generally spent building new things instead of
making existing things faster, and code is kept simple and
easy to understand so that adding new features or fixing
bugs stays easy.

That's not the case in all domains though. I've always
found reading game code interesting because it's one place
where it's common practice to optimize code even at the
cost of modularity and maintainability. One way to
accomplish that is to inline code in certain critical
sections even to the point of absurdity. CryEngine,
open-sourced a few years ago, has a few good examples of
this, with [tick functions like this one][cryengine] that
are 800+ lines long with 14 levels of indentation in
places.

Another class of software where it's common to think about
optimizations is databases, and Postgres in particular has
lots of interesting ones. Databases are an example of
software that's extremely leveraged -- if you can find a
way to make sorting rows or building indexes 10% faster, it
won't be an improvement that affects just a couple users,
it's one that'll energize millions of installations around
the world. Advantageous enough that the enhancement is very
often worth it, even at the cost of a challenging
implementation and additional code complexity.

An optimization in Postgres that's interested me for a
while is **SortSupport**, a technique for localizing the
information needed to compare data into places where it can
be accessed very quickly, thereby making sorting data much
faster. In some cases sorting gets as much as twice as fast
(or more), which speeds up common database operations like
using `ORDER BY`, `DISTINCT`, and building indexes. Let's
take a closer look at how it works.

## Sorting with abbreviated keys (#abbreviated-keys)

While sorting, Postgres builds a series of tiny structures
that represent the data set being sorted. These tuples have
space for a value the size of a native pointer (i.e. 64
bits on a 64-bit machine) which is enough to fit the
entirety of some common types like booleans or integers
(known as pass-by-value types), but not for others that are
larger than 64 bits or arbitrarily large. In their case,
Postgres will follow a references back to the heap when
comparing values (and they're therefore appropriately
called pass-by-reference types). Postgres is very fast, so
that's still a fast operation, but it's obviously slower
than comparing short values that are readily available in
memory.

TODO: Diagram of SortTuples lined up memory

SortSupport augments pass-by-reference types by bringing
some information about their heap value right into the sort
tuple to save trips to the heap. Because sort tuples
usually don't have the space to store the entirety of the
value, SortSupport generates a "digest" of the full value
called an **abbreviated key**, and stores it instead. The
contents of an abbreviated key vary by type, but they'll
aim to store as much sorting-relevant information as
possible while remaining faithful to the sorting rules of
type.

Abbreviated keys should never produce an incorrect
comparison, but it's okay if one can't be fully resolved by
what's in the abbreviated key. If two abbreviated keys look
equal, Postgres will fall back to comparing their full heap
values to make sure it gets the right result (usually
called an "authoritative comparison").

TODO: Diagram of abbreviated key point back to heap

Implementing an abbreviated key turns out to be quite
straightforward in many cases. UUIDs are a good example: at
128 bits they're always larger than the pointer size even
on a 64-bit machine, but we can get a very good sample of
their full value by just pulling in the first 64 bits (or
32 on a 32-bit machine). Especially for V4 UUIDs which are
entirely random, the first 64 bits will be enough to
definitively determine the order for all but unimaginably
large data sets. Indeed, [the patch that brought in
SortSupport for UUIDs][uuidpatch] made sorting them about
twice as fast!

String-like types (e.g. `text`, `varchar`) aren't too much
harder: just pack as many characters from the front of the
string in as possible (although made somewhat more
complicated by locales). My only ever patch to Postgres was
implementing SortSupport for the `macaddr` type, which was
quite easy because although it's a pass-by-reference type,
its values are only six bytes long [1]. On a 64-bit machine
we have room for all six bytes, and on 32-bits we sample
the MAC address' first four bytes.

## A glance at the implementation (#implementation)

I'm going to try to give you a basic idea of how
SortSupport is implemented by exposing a narrow slice of
source code. Sorting in Postgres is extremely complex and
involves thousands of lines of code, so fair warning that
I'm going to simplify some things and skip *a lot*, but we
can still upon a few interesting parts.

A good type to start is `Datum`, the pointer-sized type (32
or 64 bits, depending on the CPU's architecture) used for
sort comparisons. It stores entire values for pass-by-value
types, abbreviated keys for pass-by-reference types that
implement SortSupport, and a pointer for those that don't.
You can see it defined in [`postgres.h`][datum]:

``` c
/*
 * A Datum contains either a value of a pass-by-value type or a pointer
 * to a value of a pass-by-reference type.  Therefore, we require:
 *
 * sizeof(Datum) == sizeof(void *) == 4 or 8
 *
 * The macros below and the analogous macros for other types should be
 * used to convert between a Datum and the appropriate C type.
 */

typedef uintptr_t Datum;

#define SIZEOF_DATUM SIZEOF_VOID_P
```

### Building abbreviated keys for UUID (#uuid)

As noted above, the format of abbreviated keys for the
`uuid` type is probably the easiest to understand, so let's
take a look at that. In Postgres, the struct `pg_uuid_t`
defines how UUIDs are physically stored in the heap (from
[`uuid.h`][uuid]):

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
remember that this is Postgres, which likes to be as
efficient as possible! A UUID contains exactly 16 bytes
worth of information, so `pg_uuid_t` above defines an array
of 16 bytes (for those unfamiliar with C, a `char` is one
byte).

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

`memcpy` (read "memory copy") extracts a datum worth of
bytes from a `pg_uuid_t` and places it into `res`. We can't
take the whole UUID, but we'll be taking its 4 or 8 most
significant bytes, which will be enough information for
most comparisons.

The call `DatumBigEndianToNative` helps with an
optimization and a little more difficult to understand.
When comparing our abbreviated keys, we could do so with
`memcmp` (read "memory compare")  which would compare each
byte in the datum one at a time. That works of course, but
because our datums are the same size as native integers, we
can take advantage of the fact that CPUs can compare
integers really, really quickly (faster even than `memcmp`)
by arranging them in memory *like* integers. You can see
this integer comparison taking place in the UUID
abbreviated key comparison function:

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

But pretending that some consecutive bytes in memory are
integers introduces some complication. Integers might be
stored like `data` in `pg_uuid_t` with the most significant
byte first, but only on systems which are big-endian.
Little-endian machines store an integer's bytes in reverse
order, with the most significant at the highest address. If
we just left the result of `memcpy`, integer comparisons on
little-endian systems would come out wrong. The answer is
to byteswap, which reverses the order of the bytes, and
corrects the integer.

TODO: Diagram of endian

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

#### Abbreviated key conversion abort & HyperLogLog

Let's touch upon one more feature of `uuid_abbrev_convert`.
In data sets with very low cardinality (i.e, many
duplicated items) SortSupport introduces some danger of
worsening performance. With so many duplicates, the
contents of abbreviated keys would often show equality, in
which cases Postgres would often have to fall back to the
authoritative comparator. In effect, by adding SortSupport
we would have just added an additional comparison that
wasn't there before.

To protect against the possibility of that performance
regression, SortSupport has a mechanism for aborting
abbreviated key conversion. If the data set is found to be
below a certain cardinality threshold, Postgres stops
abbreviating, reverts any keys that it had already
abbreviated, and disables further abbreviation for the
sort.

Cardinality is estimated with the help of
[HyperLogLog][hyperloglog], an algorithm that estimates the
distinct count of a data set in a very memory-efficient
way. Here you can see the conversion routine adding new
values to the HyperLogLog if it's still considering
aborting:

``` c
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

It also covers aborting the case where we have some
degenerate data set that's poorly suited to the abbreviated
key format. For example, imagine a million UUIDs that all
shared a common prefix in their first eight bytes, but were
distinct in their last eight. Realistically this should be
extremely unusual, so abbreviated key conversion will
rarely abort.

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
go into, but you can see that it contains an
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
`strcmp`: when comparing `(a, b)`, `-1` indicates `a < b`,
0 indicates equality, and `1` indicates `a > b`.

The algorithm returns immediately if inequality (`!= 0`)
was detected. Otherwise, it checks to see if abbreviated
keys were used, and if so applies the authoritative
comparison (comparing full values from the heap) if they
were. Because space in abbreviated keys is limited, two
being equal doesn't necessarily indicate that the values
that they represent are.

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
of an index:

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
last step is to `return 0`, indicating in libc-style that
the two tuples are really, fully equal.

## Summary (#summary)

My one and only patch to Postgres involved implementing
`SortSupport` for the `macaddr` data type.

[1] The new type `macaddr8` was later introduced to handle
    EUI-64 MAC addresses, which are 64 bits long.

[comparetup]: src/backend/utils/sort/tuplesort.c:3909
[cryengine]: https://github.com/CRYTEK/CRYENGINE/blob/release/Code/CryEngine/CryPhysics/livingentity.cpp#L1275
[datum]: src/include/postgres.h:357
[heaptuple]: src/include/access/htup_details.h:152
[hyperloglog]: https://en.wikipedia.org/wiki/HyperLogLog
[itempointer]: src/include/storage/itemptr.h:20
[pgbswap]: src/include/port/pg_bswap.h:143
[sorttuple]: src/backend/utils/sort/tuplesort.c:138
[uuid]: src/include/utils/uuid.h:17
[uuidconvert]: src/backend/utils/adt/uuid.c:367
[uuidpatch]: https://www.postgresql.org/message-id/CAM3SWZR4avsTwwNVUzRNbHk8v36W-QBqpoKg%3DOGkWWy0dKtWBA%40mail.gmail.com
[varstrconvert]: src/backend/utils/adt/varlena.c:2317
