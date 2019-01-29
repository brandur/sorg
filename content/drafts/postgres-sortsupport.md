---
title: "SortSupport: Fast Comparisons in Postgres Indexes"
published_at: 2019-01-24T19:19:18Z
location: San Francisco
hook: TODO
---

Postgres sorts with structures called "sort tuples", which
are by design tiny so that many of them can be held in
memory at the same time.

`SortTuple` holds a value of type `Datum` which is the same
size as a pointer -- either 32 or 64 bits depending on the
system's architecture:

The Postgres sorting algorithms use datums when performing
a sort. Occasionally a datum can hold the entirety of a
value), when dealing with a TODO for example (called
"pass-by-value" types), but very often it can't because
values are too large to fit in 32 or 64 bits. In these
cases (called "pass-by-reference" types) the datum holds a
pointer to the full value in Postgres' physical storage
(known as _the heap_).

Postgres is happy to go to the heap to compare values, but
there's a cost associated with that -- it'd be much faster
to compare values directly in the index if possible. But as
mentioned previously, this is troublesome because many
types commonly stored in indexes are much too large to make
this practical.

SortSupport is a clever feature that for many cases manages
to achieve the best of both worlds by keeping sort tuples
small, but also allowing a great majority of comparisons to
be performed without going to the heap. It does so by
introducing a third type of value that can be stored in a
sort tuple's datum called an _abbreviated key_.

Abbreviated keys act as proxies for heap values which
Postgres can try to use on an initial sorting pass with the
understanding that because we can't fit the entire value on
a datum, it may have to fall back to the full value on the
heap if the abbreviated key isn't enough to determine
inequality. In short, they try fit as much of their value's
entropy into a 32 or 64 bit datum as possible while still
maintaining the invariant that sorting with them can never
produce an inaccuracy.

The implementation of abbreviated keys varies based on data
type. Some are relatively simple, like `text` or `uuid`,
while others like `numeric` need to use more exotic
encoding schemes (we'll take a quick look at those
implementations in the sections below).

## Abbreviated B-tree comparison (#comparison)

[`postgres.h`][datum]:

``` c
/*
 * A Datum contains either a value of a pass-by-value type or a pointer to a
 * value of a pass-by-reference type.  Therefore, we require:
 *
 * sizeof(Datum) == sizeof(void *) == 4 or 8
 *
 * The macros below and the analogous macros for other types should be used to
 * convert between a Datum and the appropriate C type.
 */

typedef uintptr_t Datum;

#define SIZEOF_DATUM SIZEOF_VOID_P
```

[`tuplesort.c`][sorttuple]

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
    void       *tuple;            /* the tuple itself */
    Datum        datum1;            /* value of first key column */
    bool        isnull1;        /* is first key column NULL? */
    int            tupindex;        /* see notes above */
} SortTuple;
```

[`itup.h`][indextuple]

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

[`itemptr.h`][itempointer]

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

[`tuplesort.c`][comparetup]

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

If the initial comparison showed equality, we need to keep
working. Next, if we were comparing abbreviated keys from
SortSupport, we go the heap and compare the full values of
the first key in the index:

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

If the comparison is still showing equal, we start to
compare other keys in the index. Recall that it's common to
create indexes on multiple keys like `CREATE INDEX
user_organization_id_email ON user (organization_id,
email)`. When running comparisons, Postgres won't look past
the first key as long as values there aren't equal, but
looks onto other keys when necessary:

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

## Implementation for UUID (#uuid)

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

## More exotic implementations (#exotic)

The implementations for `uuid` and `text` are pretty easy
to understand, but packing meaningful comparison data in 32
or 64 bits isn't always quite so straightforward.

## Summary (#summary)

My one and only patch to Postgres involved implementing
`SortSupport` for the `macaddr` data type.

[comparetup]: src/backend/utils/sort/tuplesort.c:3953
[datum]: src/include/postgres.h:357
[itempointer]: src/include/storage/itemptr.h:20
[indextuple]: src/include/access/itup.h:22
[sorttuple]: src/backend/utils/sort/tuplesort.c:138
[uuidconvert]: src/backend/utils/adt/uuid.c:367
[varstrconvert]: src/backend/utils/adt/varlena.c:2317
