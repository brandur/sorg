---
title: A Dive Into How Postgres Makes Transactions Atomic
published_at: 2017-08-06T17:10:18Z
location: San Francisco
hook: TODO
---

## MVCC (#mvcc)

## Snapshots and Transactions (#snapshots-transactions)

## Beginning a transaction (#begin)

TODO: Where the hell does a transaction start?! And how
does that call into GetSnapshotData?!

The meat of beginning a transaction is creating its
snapshot, which is performed by [`GetSnapshotData` in
`procarray.c`][getsnapshotdata]:

``` c
Snapshot
GetSnapshotData(Snapshot snapshot)
{
    ...
```

This function does a lot of initialization work, but like
we talked about, some of its most important work is set to
the snapshot's `xmin`, `xmax`, and `*xip`. The easiest of
these is `xmax`, which is retrieved from shared memory
managed by the postmaster, which is tracking the `xid`s of
any transactions that complete (more on this later):

``` c
/* xmax is always latestCompletedXid + 1 */
xmax = ShmemVariableCache->latestCompletedXid;
Assert(TransactionIdIsNormal(xmax));
TransactionIdAdvance(xmax);
```

Notice that it's the function's responsibility to add one
to the last `xid`. This isn't quite as trivial as
incrementing it because transaction IDs in Postgres are
allowed to wrap. A transaction ID is defined as a simple
unsigned 32-bit integer (from [c.h][xid]):

``` c
typedef uint32 TransactionId;
```

Even though `xid`s are assigned only opportunistically (as
mentioned above, reads don't need one), a system doing a
lot of transaction throughput can easily hit the bounds of
32 bits, so the system needs to be able to wrap to "reset"
the `xid` sequence as necessary. This is handled by some
preprocessor magic (in [transam.h][xidadvance]):

``` c
#define InvalidTransactionId		((TransactionId) 0)
#define BootstrapTransactionId		((TransactionId) 1)
#define FrozenTransactionId			((TransactionId) 2)
#define FirstNormalTransactionId	((TransactionId) 3)

...

/* advance a transaction ID variable, handling wraparound correctly */
#define TransactionIdAdvance(dest)	\
	do { \
		(dest)++; \
		if ((dest) < FirstNormalTransactionId) \
			(dest) = FirstNormalTransactionId; \
	} while(0)
```

Note that the first few IDs are reserved as special
identifiers, so we always skip those and start at `3`.

Back in `GetSnapshotData`, we get `xmin` and `xip` by
iterating over all running transactions:

``` c
/*
 * Spin over procArray checking xid, xmin, and subxids.  The goal is
 * to gather all active xids, find the lowest xmin, and try to record
 * subxids.
 */
for (index = 0; index < numProcs; index++)
{
    volatile PGXACT *pgxact = &allPgXact[pgprocno];
    TransactionId xid;
    xid = pgxact->xmin; /* fetch just once */

    /*
     * If the transaction has no XID assigned, we can skip it; it
     * won't have sub-XIDs either.  If the XID is >= xmax, we can also
     * skip it; such transactions will be treated as running anyway
     * (and any sub-XIDs will also be >= xmax).
     */
    if (!TransactionIdIsNormal(xid)
        || !NormalTransactionIdPrecedes(xid, xmax))
        continue;

    if (NormalTransactionIdPrecedes(xid, xmin))
        xmin = xid;

    /* Add XID to snapshot. */
    snapshot->xip[count++] = xid;

    ...
}

...

snapshot->xmin = xmin;
```

`xmin`'s purpose isn't to tell us everything we need to
know about what's visible to a snapshot, but it acts as a
useful horizon beyond which we know nothing is visible.
Therefore it's calculated as the minimum `xid` of all
running transactions when a snapshot is created.

## Committing a transaction (#commit)

## Checking visibility and hint bits (#visibility)

[getsnapshotdata]: src/backend/storage/ipc/procarray.c:1508
[xid]: src/include/c.h:397
[xidadvance]: src/include/access/transam.h:31
