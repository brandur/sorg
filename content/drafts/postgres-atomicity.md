---
title: How Postgres Makes Transactions Atomic
published_at: 2017-08-06T17:10:18Z
location: San Francisco
hook: Most of us have heard that transactions in Postgres
  are atomic, but given the hazardous conditions of real
  life, how does does it provide such a strong guarantee?
---

Atomicity (in the sense of "ACID") states that for a series
of operations performed against a database, either every
one of them commits together, or they're all rolled back.
No in between states are allowed. For code that needs to be
resilient to the messiness of the real world, it's a
godsend.

Instead of bugs that make it to production changing data
and then leaving it permanently corrupt, those changes are
reverted. The long tail of connections that are dropped
midway from intermittent problems and other unexpected
states while handling millions of requests might cause
inconvenience, but won't scramble your data.

Having used MongoDB in production for a few years now and
seeing first-hand the operational catastrophe inherent to
this type of non-ACID data store ([more detail on those
problems](/acid)), I've taken an interest in the subject of
data correctness. My curiosity was piqued: software is
fallible and there are a lot of things in a computer that
can go wrong, so how are some databases able to offer such
strong data guarantees?

Postgres's implementation in particular is known to provide
powerful transaction semantics with little overhead. And
while I've used it for years, it's never been something
that I've understood. Postgres works reliably enough that
I've been able to treat it as a black box -- wonderfully
useful, but with inner workings that are a mystery.

This article looks into how Postgres keeps the books on its
transactions, how they're committed atomically, and some
concepts that are key to understanding it all [1].

## Managing concurrent access (#mvcc)

Say you build a simple database that reads and writes from
an on-disk CSV file. When a single client comes in with a
request, it opens the file and writes some information.
Things are mostly working fine, but then one day you decide
to enhance your database with a sophisticated new feature,
multi-client support!

Unfortunately, the new implementation is immediately
plagued by problems that are especially noticeable when two
clients are trying to access data around the same time. One
opens the CSV file, writes some data, and that change is
immediately clobbered by another client doing its own
write.

!fig src="/assets/postgres-atomicity/csv-database.svg" caption="Data loss from contention between two clients."

This is a problem of concurrent access and it's addressed
by introducing _concurrency control_. There are plenty of
naive solutions. We could ensure that any process takes out
an exclusive lock on a file before reading or writing it,
or push all operations through a single flow control point.
Not only are these workarounds slow, but they won't scale
up to allow us to make our database fully ACID-compliant.
Modern databases have a better way, MVCC (multi-version
concurrency control).

Under MVCC, statements execute inside of a
***transaction***, and instead of overwriting data
directly, they create new versions of it. The original data
is still available to other clients that might need it, and
any new data stays hidden until the transaction commits.
Clients are no longer in direct contention, and data stays
safely persisted because they're not overwriting each
other's changes.

When a transaction starts, it takes a ***snapshot*** that
captures the state of a database at that moment in time. To
avoid the neverending accumulation of rows that have been
deleted, databases will eventually remove obsolete data by
way of a background "vacuum" process, but they'll only do
so for information that's no longer needed by open
snapshots.

Postgres manages concurrent access with MVCC. Lets take a
look at how it works.

## Transactions, tuples, and snapshots (#snapshots-transactions)

Here's the data structure that Postgres uses to represent a
transaction (from [proc.c][pgxact]):

``` c
typedef struct PGXACT
{
    TransactionId xid;   /* id of top-level transaction currently being
                          * executed by this proc, if running and XID
                          * is assigned; else InvalidTransactionId */

    TransactionId xmin;  /* minimal running XID as it was when we were
                          * starting our xact, excluding LAZY VACUUM:
                          * vacuum must not remove tuples deleted by
                          * xid >= xmin ! */

    ...
} PGXACT;
```

Transactions are identified with a `xid` (transaction, or
"xact" ID). As an optimization, Postgres will only assign a
transaction a `xid` if it starts to modify data because
it's only at that point where other processes need to start
tracking its changes. Readonly transactions can execute
happily without ever needing a `xid`.

`xmin` is always set immediately to the smallest `xid` of
any transactions that are still running when this one
starts. Vacuum processes calculate the minimum boundary of
data that they need to keep by taking the minimum of the
`xmin`s of all active transactions 

### Lifetime-aware tuples (#tuples)

Rows of data in Postgres are often referred to as
***tuples***. While Postgres uses common lookup structures
like B-trees to make retrievals fast, indexes don't store a
tuple's full set of data or any of its visibility
information. Instead, they store a `tid` (tuple ID) that
can be used to retrieve a row from physical storage,
otherwise known as "the heap". The `tid` gives Postgres a
starting point where it can start scanning the heap until
it finds a tuple that satisfies the current snapshot's
visibility.

Here's the Postgres implementation for a _heap tuple_ (as
opposed to an _index tuple_ which is the structure found in
an index), along with a few other structs that represent
its header information ([from `htup.h`][tuple] [and
`htup_details.h`][tupleheaders]):

``` c
typedef struct HeapTupleData
{
    uint32          t_len;         /* length of *t_data */
    ItemPointerData t_self;        /* SelfItemPointer */
    Oid             t_tableOid;    /* table the tuple came from */
    HeapTupleHeader t_data;        /* -> tuple header and data */
} HeapTupleData;

/* referenced by HeapTupleData */
struct HeapTupleHeaderData
{
    HeapTupleFields t_heap;

    ...
}

/* referenced by HeapTupleHeaderData */
typedef struct HeapTupleFields
{
    TransactionId t_xmin;        /* inserting xact ID */
    TransactionId t_xmax;        /* deleting or locking xact ID */

    ...
} HeapTupleFields;
```

Like a transaction, a tuple tracks its own `xmin`, except
in the tuple's case it's recorded to represent the first
transaction where the tuple becomes visible (i.e. the one
that created it). It also tracks `xmax` to be the _last_
transaction where the tuple is visible (i.e. the one that
deleted it) [2].

!fig src="/assets/postgres-atomicity/heap-tuple-visibility.svg" caption="A heap tuple's lifetime being tracked with xmin and xmax."

### Snapshots: xmin, xmax, and xip (#snapshots)

Here's the snapshot structure ([from snapshot.h][snapshot]):

``` c
typedef struct SnapshotData
{
    /*
     * The remaining fields are used only for MVCC snapshots, and are normally
     * just zeroes in special snapshots.  (But xmin and xmax are used
     * specially by HeapTupleSatisfiesDirty.)
     *
     * An MVCC snapshot can never see the effects of XIDs >= xmax. It can see
     * the effects of all older XIDs except those listed in the snapshot. xmin
     * is stored as an optimization to avoid needing to search the XID arrays
     * for most tuples.
     */
    TransactionId xmin;            /* all XID < xmin are visible to me */
    TransactionId xmax;            /* all XID >= xmax are invisible to me */

    /*
     * For normal MVCC snapshot this contains the all xact IDs that are in
     * progress, unless the snapshot was taken during recovery in which case
     * it's empty. For historic MVCC snapshots, the meaning is inverted, i.e.
     * it contains *committed* transactions between xmin and xmax.
     *
     * note: all ids in xip[] satisfy xmin <= xip[i] < xmax
     */
    TransactionId *xip;
    uint32        xcnt; /* # of xact ids in xip[] */

    ...
}
```

A snapshot's `xmin` is calculated the same way as a
transaction's (i.e. the lowest `xid` amongst running
transactions when the snapshot is created), but for a
different prupose. This `xmin` is a lower boundary for data
visibility. Tuples created by a transaction with `xid <
xmin` are visible to the snapshot.

It also defines an `xmax`, which is set to the last
commited `xid` plus one. `xmax` tracks the upper bound of
visibility; transactions with `xid >= xmax` are invisible
to the snapshot.

Lastly, a snapshot defines `*xip`, an array of all of the
`xid`s of transactions that were in progress when the
snapshot was created. It's created because even though
there's already a visibility boundary with `xmin`, there
may still be some transactions that are already committed
with `xid`s greater than `xmin`, but _also_ greater than a
`xid` of an in-progress transaction (so they couldn't be
included in `xmin`).

We want the results any committed transactions with `xid >
xmin` to be visible, but the results of any that were in
flight hidden. `*xip` stores the list of transactions that
were active when the snapshot was created so that we can
tell which is which.

!fig src="/assets/postgres-atomicity/snapshot-creation.svg" caption="Transactions executing against a database and a snapshot capturing a moment in time."

## Beginning a transaction (#begin)

When you execute a `BEGIN`, Postgres puts some basic
bookeeping in place, but it will defer more expensive
operations as long as it can. For example, the new
transaction isn't assigned a `xid` until it starts
modifying data to reduce the expense of tracking it
elsewhere in the system.

The new transaction also won't immediately get a snapshot.
One won't be assigned until the transaction's first query,
whereupon `exec_simple_query` ([in
`postgres.c`][execsimplequery]) will push a snapshot onto a
stack. Even a simple `SELECT 1;` is enough to trigger it:

``` c
static void
exec_simple_query(const char *query_string)
{
    ...

    /*
     * Set up a snapshot if parse analysis/planning will need one.
     */
    if (analyze_requires_snapshot(parsetree))
    {
        PushActiveSnapshot(GetTransactionSnapshot());
        snapshot_set = true;
    }

    ...
}
```

Creating the new snapshot is where the transaction
machinery starts to come into effect. Here's
`GetSnapshotData` ([in `procarray.c`][getsnapshotdata]):

``` c
Snapshot
GetSnapshotData(Snapshot snapshot)
{
    /* xmax is always latestCompletedXid + 1 */
    xmax = ShmemVariableCache->latestCompletedXid;
    Assert(TransactionIdIsNormal(xmax));
    TransactionIdAdvance(xmax);

    ...

    snapshot->xmax = xmax;
}
```

This function does a lot of initialization, but like we
talked about, some of its most important work is set to the
snapshot's `xmin`, `xmax`, and `*xip`. The easiest of these
is `xmax`, which is retrieved from shared memory managed by
the postmaster. Every transaction that commits notifies the
postmaster that it did, and `latestCompletedXid` will be
updated if the `xid` is higher than what's already in there
(more on this later).

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
lot of throughput can easily hit the bounds of 32 bits, so
the system needs to be able to wrap to "reset" the `xid`
sequence as necessary. This is handled by some preprocessor
magic (in [transam.h][xidadvance]):

``` c
#define InvalidTransactionId        ((TransactionId) 0)
#define BootstrapTransactionId      ((TransactionId) 1)
#define FrozenTransactionId         ((TransactionId) 2)
#define FirstNormalTransactionId    ((TransactionId) 3)

...

/* advance a transaction ID variable, handling wraparound correctly */
#define TransactionIdAdvance(dest)    \
    do { \
        (dest)++; \
        if ((dest) < FirstNormalTransactionId) \
            (dest) = FirstNormalTransactionId; \
    } while(0)
```

Note that the first few IDs are reserved as special
identifiers, so we always skip those and start at `3`.

Back in `GetSnapshotData`, we get `xmin` and `xip` by
iterating over all running transactions (again, see
[Snapshots](#snapshots) above for an explanation of what
these do):

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

## Committing a transaction (#commit)

Transactions committed through [`CommitTransaction` (in
`xact.c`)][commit]. This function is monstrously complex,
but here are a few of its important parts:

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
log (WAL, or often called a "clog", "xlog", or "transaction
log" in Postgres lingo) to achieve this durability. Every
committed change is written and flushed to disk, and even
in the event of sudden termination, Postgres can replay
what it finds in the WAL to recover any changes that didn't
make it into its data files.

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

Another core Postgres value is performance. If a
transaction was never assigned a `xid` because it didn't
affect the state of the database, Postgres skips writing it
to the WAL. If a transaction was aborted, we write it to
the WAL, but don't bother to flush its commit status
because even though it's completed, it doesn't affect data
so it's not catastrophic if we lose it (any transactions
found in the WAL without a commit status are assumed
aborted).

It's also worth pointing out that the WAL is written in two
parts. We write the bulk of the information out in
`XactLogCommitRecord`, and if the transaction committed, we
go through in a second pass with `TransactionIdCommitTree`
and set the status of each record to "committed". It's only
after this operation completes that we can formally say
that the transaction was durably committed.

### Defensive programming (#defensive-programming)

`TransactionIdCommitTree` (in [transam.c][committree], and
its implementation `TransactionIdSetTreeStatus` in
[clog.c][settreestatus]) commits a "tree" because a commit
may have subcommits. I won't go into subcommits in any
detail, but it's worth nothing that because
`TransactionIdCommitTree` cannot be guaranteed to be
atomic, each subcommit is recorded as committed separately,
and the parent is recorded as a final step. When Postgres
is reading the WAL on recovery, subcommit records aren't
considered to be committed (even if they're marked as such)
until the parent record is read and confirmed committed.

Once again this is in the name of atomicity; the system
could have successfully recorded every subcommit, but then
crashed before it could write the parent.

### Signaling completion through shared memory (#shared-memory)

With the transaction recorded to WAL, it's safe to signal
its completion to the rest of the system. This happens in
the second call in `CommitTransaction` above ([into
procarray.c][endtransaction]):

``` c
void
ProcArrayEndTransaction(PGPROC *proc, TransactionId latestXid)
{
    /*
     * We must lock ProcArrayLock while clearing our advertised XID, so
     * that we do not exit the set of "running" transactions while someone
     * else is taking a snapshot.  See discussion in
     * src/backend/access/transam/README.
     */
    if (LWLockConditionalAcquire(ProcArrayLock, LW_EXCLUSIVE))
    {
        ProcArrayEndTransactionInternal(proc, pgxact, latestXid);
        LWLockRelease(ProcArrayLock);
    }

    ...
}

static inline void
ProcArrayEndTransactionInternal(PGPROC *proc, PGXACT *pgxact,
                                TransactionId latestXid)
{
    ...

    /* Also advance global latestCompletedXid while holding the lock */
    if (TransactionIdPrecedes(ShmemVariableCache->latestCompletedXid,
                              latestXid))
        ShmemVariableCache->latestCompletedXid = latestXid;
}
```

Remember how when we created a snapshot we set its `xmax`
to `latestCompletedXid + 1`? By setting
`latestCompletedXid` to the `xid` of the transaction that
just committed, we've just made its results visible to
every new snapshot that starts from this point forward.

### Responding to the client (#client)

Throughout this entire process, a client has been waiting
synchronously for their transaction to be confirmed. Part
of the atomicity guarantee is that false positives where
the databases signals a transaction as committed when it
hasn't been aren't possible. Failures can happen in many
places, but if there is one, the client finds out about it
and has a chance to retry or otherwise address the problem.

## Checking visibility (#visibility)

We covered earlier how visibility information is stored on
heap tuples. `heapgettup` (in [heapam.c][gettup]) is the
method responsible for scanning the heap for tuples that
meet a snapshot's visibility criteria:

``` c
static void
heapgettup(HeapScanDesc scan,
           ScanDirection dir,
           int nkeys,
           ScanKey key)
{
    ...

    /*
     * advance the scan until we find a qualifying tuple or run out of stuff
     * to scan
     */
    lpp = PageGetItemId(dp, lineoff);
    for (;;)
    {
        /*
         * if current tuple qualifies, return it.
         */
        valid = HeapTupleSatisfiesVisibility(tuple,
                                             snapshot,
                                             scan->rs_cbuf);

        if (valid)
        {
            return;
        }

        ++lpp;            /* move forward in this page's ItemId array */
        ++lineoff;
    }

    ...
}
```

`HeapTupleSatisfiesVisibility` is a preprocessor macro that
will call into "satisfies" function like
`HeapTupleSatisfiesMVCC` ([in `tqual.c`][satisfies]):

``` c
bool
HeapTupleSatisfiesMVCC(HeapTuple htup, Snapshot snapshot,
                       Buffer buffer)
{
    ...

    else if (TransactionIdDidCommit(HeapTupleHeaderGetRawXmin(tuple)))
        SetHintBits(tuple, buffer, HEAP_XMIN_COMMITTED,
                    HeapTupleHeaderGetRawXmin(tuple));

    ...

    /* xmax transaction committed */

    return false;
}
```

And `TransactionIdDidCommit` ([from
`transam.c`][didcommit]):

``` c
bool /* true if given transaction committed */
TransactionIdDidCommit(TransactionId transactionId)
{
    XidStatus xidstatus;

    xidstatus = TransactionLogFetch(transactionId);

    /*
     * If it's marked committed, it's committed.
     */
    if (xidstatus == TRANSACTION_STATUS_COMMITTED)
        return true;

    ...
}
```

Further exploring the implementation of
`TransactionLogFetch` will reveal that it works as
advertised. It calculates a location in the WAL from the
given transaction ID and reaches into the WAL to get that
transaction's commit status. Whether or not the transaction
committed is used to help determine the tuple's visibility.

The key here is that for purposes of consistency, the WAL
is considered the canonical source for commit status (and
by extension, visibility) [3]. The same information will be
returned regardless of whether Postgres successfully
committed a transaction hours ago, or seconds before a
crash that the server is just now recovering from.

### Hint bits (#hint-bits)

`HeapTupleSatisfiesMVCC` from above does one more thing
before returning from a visibility check:

``` c
SetHintBits(tuple, buffer, HEAP_XMIN_COMMITTED,
            HeapTupleHeaderGetRawXmin(tuple));
```

Checking the WAL to see whether a tuple's `xmin` or `xmax`
transactions are committed is an expensive operation. To
avoid having to go to the WAL every time, Postgres will set
special commit status flags called "hint bits" for a heap
tuple that its scanned. Subsequent operations can check the
tuple's hint bits and are saved a trip to the WAL
themselves.

## The box's opaque walls (#black-box)

When I run a transaction against a database:

``` sql
BEGIN;

SELECT * FROM users 

INSERT INTO users (email) VALUES ('brandur@example.com')
RETURNING *;

COMMIT;
```

I don't really need to think about what's going on. I'm
given a powerful high level abstraction (in the form of
SQL) which I know will work reliably, and as we've seen,
Postgres does all the heavy lifting under the covers. Good
software is a black box, and Postgres is an especially dark
one (although with pleasantly accessible internals).

Thank you to [Peter Geoghegan][peter] for patiently
answering all my amateur questions about Postgres
transactions and snapshots, and giving me some pointers for
finding relevant code.

[commit]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/access/transam/xact.c#L1939
[committree]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/access/transam/transam.c#L259
[didcommit]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/access/transam/transam.c#L124
[endtransaction]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/storage/ipc/procarray.c#L394
[execsimplequery]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/tcop/postgres.c#L1010
[gettup]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/access/heap/heapam.c#L478
[getsnapshotdata]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/storage/ipc/procarray.c#L1507
[peter]: https://twitter.com/petervgeoghegan
[pgxact]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/storage/proc.h#L207
[satisfies]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/utils/time/tqual.c#L962
[settreestatus]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/access/transam/clog.c#L148
[snapshot]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/utils/snapshot.h#L52
[tuple]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/access/htup_details.h#L116
[tupleheaders]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/access/htup.h#L62
[xid]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/c.h#L397
[xidadvance]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/access/transam.h#L31

[1]  A few words of warning: the Postgres source code is
pretty overwhelming, so I've glossed over a few details to
make this reading more digestible. It's also under active
development, so the passage of time will likely render some
of these code samples quite obsolete.

[2] Readers may notice that while `xmin` and `xmax` are
fine for tracking a tuple's creation and deletion, they
aren't to enough to handle updates. For brevity's sake, I'm
glossing over how updates work for now.

[3] Note that changes will eventually be no longer
available in the WAL, but those will always be beyond a
snapshot's `xmin` horizon, and therefore the visibility
check short circuits before having to make a check in WAL.
