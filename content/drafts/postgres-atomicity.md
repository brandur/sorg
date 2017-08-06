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

Transactions committed through [`CommitTransaction` (in
`xact.c`)][commit]. This function is monstrously complex,
but again, I'm going to simplify it and call out a couple
of the most important parts:

``` c
static void
CommitTransaction(void)
{
    ...

    /*
     * We need to mark our XIDs as committed in pg_xact.  This is where we
     * durably commit.
     */
    latestXid = RecordTransactionCommit();

    /*
     * Let others know about no transaction in progress by me. Note that this
     * must be done _before_ releasing locks we hold and _after_
     * RecordTransactionCommit.
     */
    ProcArrayEndTransaction(MyProc, latestXid);

    ...
}
```

Postgres is entirely designed around the idea of
durability, which dictates that even in extreme events like
a crash or power loss, committed transactions should stay
committed. Like many good systems, it uses a write-ahead
log (WAL, or often called a "clog" or "xlog" in Postgres
lingo) to achieve this durability. Every committed change
is written and flushed to disk, and even in the event of
sudden termination, Postgres can replay what it finds in
the WAL to recover any changes didn't make it into its data
files.

`RecordTransactionCommit` from the snippet above handles
getting a change to the WAL:

``` c
static TransactionId
RecordTransactionCommit(void)
{
    bool markXidCommitted = TransactionIdIsValid(xid);

    /*
     * If we haven't been assigned an XID yet, we neither can, nor do we want
     * to write a COMMIT record.
     */
    if (!markXidCommitted)
    {
        ...
    } else {
        XactLogCommitRecord(xactStopTimestamp,
                            nchildren, children, nrels, rels,
                            nmsgs, invalMessages,
                            RelcacheInitFileInval, forceSyncCommit,
                            MyXactFlags,
                            InvalidTransactionId /* plain commit */ );

        ....
    }

    if ((wrote_xlog && markXidCommitted &&
         synchronous_commit > SYNCHRONOUS_COMMIT_OFF) ||
        forceSyncCommit || nrels > 0)
    {
        XLogFlush(XactLastRecEnd);

        /*
         * Now we may update the CLOG, if we wrote a COMMIT record above
         */
        if (markXidCommitted)
            TransactionIdCommitTree(xid, nchildren, children);
    }

    ...
}
```

Another core Postgres philosophy is performance. If a
transaction was never assigned a `xid` because it didn't
affect the state of the database, Postgres skips writing it
to the wAL. If a transaction was aborted, we write it to
the WAL, but don't bother to send a flush because even
though it's completed, it doesn't affect data so it's not
catastrophic if we lose it.

Note also that the WAL is written in two parts. We write
the bulk of the information out in `XactLogCommitRecord`,
and if the transaction committed, we go through in a second
pass with `TransactionIdCommitTree` and set the status of
each record to "committed". It's only after this operation
completes that we can formally say that the transaction was
durably committed.

`TransactionIdCommitTree` commits a "tree" because a commit
may have subcommits. I won't get into subcommits too much,
but it's worth nothing that because
`TransactionIdCommitTree` cannot be guaranteed to be
atomic, each subcommit is recorded as committed and then
the parent is recorded as a final step. When Postgres is
reading the WAL on recovery, subcommits aren't considered
to be committed even if they're marked as such until the
parent record is read and confirmed committed. This is
because the system could have successfully recorded every
subcommit, but then crashed before it could write the
parent.

## Checking visibility and hint bits (#visibility)

[commit]: src/backend/access/transam/xact.c:1939
[getsnapshotdata]: src/backend/storage/ipc/procarray.c:1508
[xid]: src/include/c.h:397
[xidadvance]: src/include/access/transam.h:31
