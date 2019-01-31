---
title: "SortSupport: Fast Comparisons in Postgres Indexes"
published_at: 2019-01-24T19:19:18Z
location: San Francisco
hook: TODO
---

An interesting aspect of systems like Postgres is that
optimizations that might not make sense in other contexts
because of how time consuming they are to write or the
additional complexity stemming from their implementation
*do* make sense because of the incredible leverage
involved. You don't just speed up a couple of
installations -- you speed up _millions_ of them around the
world.

One optimization in this vein that I want to talk about
today is **SortSupport**. In essence SortSupport is a
technique for localizing the information needed to compare
data into places where it can be accessed very quickly,
thereby making sorting data much faster. This is useful in
places like using an `ORDER BY`, `DISTINCT`, and building
indexes.

## Sorting with abbreviated keys (#abbreviated-keys)

The basic idea is pretty simple. While sorting, Postgres
builds a series of tiny structures that represent the data
set being sorted. These tuples have space for a value the
size of a native pointer (i.e., 64 bits on a 64-bit
machine) which is enough to fit the entirety of some common
types like booleans or integers (known as pass-by-value
types). But for types that are larger than 64 bits or which
can be arbitrarily large, it's not enough. In their case,
Postgres will follow a references back to the heap when
comparing values (called pass-by-reference types). Postgres
is very fast, so that's still a fast operation, but it's
obviously slower than comparing short values that are
readily available in memory.

TODO: Diagram of SortTuples lined up memory

SortSupport allows pass-by-reference types to be augmented
by bringing some information about a value right into the
sort tuple to potentially save trips to the heap. Because
sort tuples usually don't have the space to store the
value's entirety, a "digest" version called an
**abbreviated key** is stored instead. What's stored in
abbreviated key varies by type, but the overarching goal is
to store as much sorting-relevant information as possible
while remaining faithful to the sorting rules of type.

Abbreviated keys should obviously never produce an
incorrect comparison, but it's okay if one can't be fully
resolved by what's in the abbreviated key. If two
abbreviated keys look equal, Postgres will fall back to the
type's authoritative comparison function to make sure it
has the right result.

TODO: Diagram of abbreviated key point back to heap

Implementing an abbreviated key turns out to be quite
straightforward in many cases. UUIDs are a good example
because at 128 bits they're always larger than the pointer
size even on a 64-bit machine, it's pretty obvious what to
do about that -- just pull in their first 64 bits of
information (or 32 bits on a 32-bit machine). Especially
for V4 UUIDs which are entirely random, the first 64 bits
will be enough to definitively determine the order for all
but unimaginably large data sets. Indeed the patch that
brought in SortSupport for UUIDs reduced typical sort time
by about 50% -- that's twice as fast! (TODO: verify this)

String-like types (e.g. `text`, `varchar`) aren't too much
harder: just pack as many characters from the front of the
string in as possible (although made somewhat more
complicated by locales). My only ever patch to Postgres was
implementing SortSupport for the `macaddr` type, which was
quite easy because although it's a pass-by-reference type,
its values are only six bytes long [1].

## A glance at the implementation (#implementation)

Let's take a quick look at how SortSupport is implemented.
For brevity I'm going to simplify and skip some things so
just remember that the code is always canonical! Go take a
look if you're interested.

An good type to start with is `Datum`, the pointer-sized
type (32 or 64 bits) used for sort comparisons. It stores
entire values for pass-by-value types, and abbreviated keys
for SortSupport. You can see it defined in
[`postgres.h`][datum]:

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

`SortTuple`s are the tiny structures that Postgres sorts in
memory. It holds a reference to the "true" tuple, a
`Datum`, and a flag to indicate whether or not the first
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

`tuple` above is often an `IndexTuple` (from [`itup.h`][indextuple]):

``` c
/*
 * Index tuple header structure
 */
typedef struct IndexTupleData
{
    ItemPointerData t_tid;        /* reference TID to heap tuple */

    /* ---------------
     * t_info is laid out in the following fashion:
     *
     * 15th (high) bit: has nulls
     * 14th bit: has var-width attributes
     * 13th bit: AM-defined meaning
     * 12-0 bit: size of tuple
     * ---------------
     */

    unsigned short t_info;        /* various info about tuple */

} IndexTupleData;                /* MORE DATA FOLLOWS AT END OF STRUCT */

typedef IndexTupleData *IndexTuple;
```

And you can see the `ItemPointerData` it contains give
Postgres the precise information it needs to find data in
the heap (from [`itemptr.h`][itempointer]):

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

### Comparison (#comparison)

A good place to look at how comparisons take place while
sorting is the comparison function used for a B-tree index
(from [`tuplesort.c`][comparetup]):

``` c
static int
comparetup_index_btree(const SortTuple *a, const SortTuple *b,
                       Tuplesortstate *state)
{
    ...

    /* Compare the leading sort key */
    compare = ApplySortComparator(a->datum1, a->isnull1,
                                  b->datum1, b->isnull1,
                                  sortKey);
    if (compare != 0)
        return compare;
    ...

```

`ApplySortComparator` gets a comparison result between two
values. It'll compare two abbreviated keys where
appropriate (it may full back to authoritative comparison
in cases where key abbreviation has been aborted) and
handles `NULL` sorting semantics. Comparisons occur in the
spirit of C's `strcmp`: when comparing `(a, b)`, `-1`
indicates `a < b`, 0 indicates equality, and `1` indicates
`a > b`.

The algorithm returns immediately if inequality was
detected. Otherwise, it checks to see if abbreviated keys
were used, and if so applies the authoritative if they
were. Because information in abbreviated keys is limited,
two being equal doesn't necessarily indicate that the
values that they represent are.

``` c
/* Compare additional sort keys */
tuple1 = (IndexTuple) a->tuple;
tuple2 = (IndexTuple) b->tuple;
keysz = state->nKeys;
tupDes = RelationGetDescr(state->indexRel);

if (sortKey->abbrev_converter)
{
    datum1 = index_getattr(tuple1, 1, tupDes, &isnull1);
    datum2 = index_getattr(tuple2, 1, tupDes, &isnull2);

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
SortSupport sortKey = state->sortKeys;

for (nkey = 2; nkey <= keysz; nkey++, sortKey++)
{
    datum1 = index_getattr(tuple1, nkey, tupDes, &isnull1);
    datum2 = index_getattr(tuple2, nkey, tupDes, &isnull2);

    compare = ApplySortComparator(datum1, isnull1,
                                  datum2, isnull2,
                                  sortKey);
    if (compare != 0)
        return compare;        /* done when we find unequal attributes */
}
```

If two index tuples are *still* equal after that, it falls
back to using the block and offset from `ItemPointer` which
will always produce a non-equal comparison:

``` c
/*
 * If key values are equal, we sort on ItemPointer.  This does not affect
 * validity of the finished index, but it may be useful to have index
 * scans in physical order.
 */
{
    BlockNumber blk1 = ItemPointerGetBlockNumber(&tuple1->t_tid);
    BlockNumber blk2 = ItemPointerGetBlockNumber(&tuple2->t_tid);

    if (blk1 != blk2)
        return (blk1 < blk2) ? -1 : 1;
}
{
    OffsetNumber pos1 = ItemPointerGetOffsetNumber(&tuple1->t_tid);
    OffsetNumber pos2 = ItemPointerGetOffsetNumber(&tuple2->t_tid);

    if (pos1 != pos2)
        return (pos1 < pos2) ? -1 : 1;
}
```

### Building abbreviated keys for UUID (#uuid)

[`uuid.c`][uuidconvert]

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

### More exotic implementations (#exotic)

The implementation for `uuid` is pretty easy to understand,
but packing meaningful comparison data in 32 or 64 bits
isn't always quite so straightforward.

## Summary (#summary)

My one and only patch to Postgres involved implementing
`SortSupport` for the `macaddr` data type.

[1] The new type `macaddr8` was later introduced to handle
    EUI-64 MAC addresses, which are 64 bits long.

[comparetup]: src/backend/utils/sort/tuplesort.c:3953
[datum]: src/include/postgres.h:357
[itempointer]: src/include/storage/itemptr.h:20
[indextuple]: src/include/access/itup.h:22
[sorttuple]: src/backend/utils/sort/tuplesort.c:138
[uuidconvert]: src/backend/utils/adt/uuid.c:367
[varstrconvert]: src/backend/utils/adt/varlena.c:2317
