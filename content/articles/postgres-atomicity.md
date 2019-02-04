---
title: How Postgres Makes Transactions Atomic
published_at: 2017-08-16T14:20:53Z
location: San Francisco
hook: A dive into the mechanics that allow Postgres to
  provide strong atomic guarantees despite the chaotic
  entropy of production.
tags: ["postgres"]
hn_link: https://news.ycombinator.com/item?id=15027870
---

Atomicity (in the sense of "ACID") states that for a series
of operations performed against a database, either every
one of them commits together, or they're all rolled back;
no in between states are allowed. For code that needs to be
resilient to the messiness of the real world, it's a
godsend.

Instead of bugs that make it to production changing data
and then leaving it permanently corrupt, those changes are
reverted. The long tail of connections that are dropped
midway from intermittent problems and other unexpected
states while handling millions of requests might cause
inconvenience, but won't scramble your data.

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
request, it opens the file, reads some information, and
writes the changes back. Things are mostly working fine,
but then one day you decide to enhance your database with a
sophisticated new feature,
multi-client support!

Unfortunately, the new implementation is immediately
plagued by problems that seem to especially apparent when
two clients are trying to access data around the same time.
One opens the CSV file, reads, modifies, and writes some
data, but that change is immediately clobbered by another
client trying to do the same.

!fig src="/assets/postgres-atomicity/csv-database.svg" caption="Data loss from contention between two clients."

This is a problem of concurrent access and it's addressed
by introducing _concurrency control_. There are plenty of
naive solutions. We could ensure that any process takes out
an exclusive lock on a file before reading or writing it,
or we could push all operations through a single flow
control point so that they only run one at a time. Not only
are these workarounds slow, but they won't scale up to
allow us to make our database fully ACID-compliant. Modern
databases have a better way, MVCC (multi-version
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
captures the state of a database at that moment in time.
Every transaction in the database is applied in _serial_
order, with a global lock ensuring that only one is being
confirmed committed or aborted at a time. A snapshot is a
perfect representation of the database's state between two
transactions.

To avoid the neverending accumulation of rows that have
been deleted and hidden, databases will eventually remove
obsolete data by way of a _vacuum_ process (or in some
cases, opportunistic "microvacuums" that happen in band
with other queries), but they'll only do so for information
that's no longer needed by open snapshots.

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

`xmin` and `xmax` are internal concepts, but they can be
revealed as hidden columns on any Postgres table. Just
select them explicitly by name:

``` sql
# SELECT *, xmin, xmax FROM names;

 id |   name   | xmin  | xmax
----+----------+-------+-------
  1 | Hyperion | 27926 | 27928
  2 | Endymion | 27927 |     0
```

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
different purpose. This `xmin` is a lower boundary for data
visibility. Tuples created by a transaction with `xid <
xmin` are visible to the snapshot.

It also defines an `xmax`, which is set to the last
committed `xid` plus one. `xmax` tracks the upper bound of
visibility; transactions with `xid >= xmax` are invisible
to the snapshot.

Lastly, a snapshot defines `*xip`, an array of all of the
`xid`s of transactions that were in progress when the
snapshot was created. `*xip` is needed because even though
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
bookkeeping in place, but it will defer more expensive
operations as long as it can. For example, the new
transaction isn't assigned a `xid` until it starts
modifying data to reduce the expense of tracking it
elsewhere in the system.

The new transaction also won't immediately get a snapshot.
It will when it runs its first query, whereupon
`exec_simple_query` ([in `postgres.c`][execsimplequery])
will push one onto a stack. Even a simple `SELECT 1;` is
enough to trigger it:

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

Creating the new snapshot is where the machinery really
starts coming to life. Here's `GetSnapshotData` ([in
`procarray.c`][getsnapshotdata]):

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
updated if the `xid` is higher than what it already holds.
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

The first few IDs are reserved as special identifiers, so
we always skip those and start at `3`.

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

Transactions are committed through [`CommitTransaction` (in
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

### Durability and the WAL (#durability)

Postgres is entirely designed around the idea of
durability, which dictates that even in extreme events like
a crash or power loss, committed transactions should stay
committed. Like many good systems, it uses a _write-ahead
log_ (_WAL_, or "xlog") to achieve this durability. All
changes are written and flushed to disk, and even in the
event of a sudden termination, Postgres can replay what it
finds in the WAL to recover any changes that didn't make it
into its data files.

`RecordTransactionCommit` from the snippet above handles
getting a change in transaction state to the WAL:

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

### The commit log (#commit-log)

Along with the WAL, Postgres also has a _commit log_ (or
"clog" or "pg_xact") which summarizes every transaction and
whether it committed or aborted. This is what
`TransactionIdCommitTree` is doing above -- the bulk of the
information is written out to WAL first, then
`TransactionIdCommitTree` goes through and sets the
transaction's status in the commit log to "committed".

Although the commit log is called a "log", it's really more
of a bitmap of commit statuses split across a number of
pages in shared memory and on disk. In an example of the
kind of frugality rarely seen in modern programming, the
status of a transaction can be recorded in only two bits,
so we can store four transactions per byte, or 32,768 in a
standard 8k page.

From [`clog.h`][clogstatuses] and [`clog.c`][clogbits]:

``` c
#define TRANSACTION_STATUS_IN_PROGRESS      0x00
#define TRANSACTION_STATUS_COMMITTED        0x01
#define TRANSACTION_STATUS_ABORTED          0x02
#define TRANSACTION_STATUS_SUB_COMMITTED    0x03

#define CLOG_BITS_PER_XACT  2
#define CLOG_XACTS_PER_BYTE 4
#define CLOG_XACTS_PER_PAGE (BLCKSZ * CLOG_XACTS_PER_BYTE)
```

### All sizes of optimization (#optimization)

While durability is important, performance is also a value
that's core to the Postgres philosophy. If a transaction
was never assigned a `xid`, Postgres skips writing it to
the WAL and commit log. If a transaction was aborted, we
still write its aborted status to the WAL and commit log,
but don't bother to immediately flush (fsync) because even
in the event of a crash, we wouldn't lose any information.
During crash recovery, Postgres would notice the unflagged
transactions, and assume that they were aborted.

### Defensive programming (#defensive-programming)

`TransactionIdCommitTree` (in [transam.c][committree], and
its implementation `TransactionIdSetTreeStatus` in
[clog.c][settreestatus]) commits a "tree" because a commit
may have subcommits. I won't go into subcommits in any
detail, but it's worth nothing that because
`TransactionIdCommitTree` cannot be guaranteed to be
atomic, each subcommit is recorded as committed separately,
and the parent is recorded as a final step. When Postgres
is recovering after a crash, subcommit records aren't
considered to be committed (even if they're marked as such)
until the parent record is read and confirmed committed.

Once again this is in the name of atomicity; the system
could have successfully recorded every subcommit, but then
crashed before it could write the parent.

Here's what that looks like [in `clog.c`][setpagestatus]:

``` c
/*
 * Record the final state of transaction entries in the commit log for
 * all entries on a single page.  Atomic only on this page.
 *
 * Otherwise API is same as TransactionIdSetTreeStatus()
 */
static void
TransactionIdSetPageStatus(TransactionId xid, int nsubxids,
                           TransactionId *subxids, XidStatus status,
                           XLogRecPtr lsn, int pageno)
{
    ...

    LWLockAcquire(CLogControlLock, LW_EXCLUSIVE);

    /*
     * Set the main transaction id, if any.
     *
     * If we update more than one xid on this page while it is being written
     * out, we might find that some of the bits go to disk and others don't.
     * If we are updating commits on the page with the top-level xid that
     * could break atomicity, so we subcommit the subxids first before we mark
     * the top-level commit.
     */
    if (TransactionIdIsValid(xid))
    {
        /* Subtransactions first, if needed ... */
        if (status == TRANSACTION_STATUS_COMMITTED)
        {
            for (i = 0; i < nsubxids; i++)
            {
                Assert(ClogCtl->shared->page_number[slotno] == TransactionIdToPage(subxids[i]));
                TransactionIdSetStatusBit(subxids[i],
                                          TRANSACTION_STATUS_SUB_COMMITTED,
                                          lsn, slotno);
            }
        }

        /* ... then the main transaction */
        TransactionIdSetStatusBit(xid, status, lsn, slotno);
    }

    ...

    LWLockRelease(CLogControlLock);
}
```

### Signaling completion through shared memory (#shared-memory)

With the transaction recorded to commit log, it's safe to
signal its completion to the rest of the system. This
happens in the second call in `CommitTransaction` above
([into procarray.c][endtransaction]):

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

You may be wondering what a "proc array" is. Unlike many
other daemon-like services, Postgres uses a process
forking model to handle concurrency instead of threading.
When it accepts a new connection, the Postmaster forks a
new backend ([in `postmaster.c`][backendstartup]). Backends
are represented by the `PGPROC` structure ([in
`proc.h`][pgproc]), and the entire set of active processes
is tracked in shared memory, thus "proc array".

Now remember how when we created a snapshot we set its
`xmax` to `latestCompletedXid + 1`? By setting
`latestCompletedXid` in global shared memory to the `xid`
of the transaction that just committed, we've just made its
results visible to every new snapshot that starts from this
point forward across any backend.

Take a look at the lock acquisition and release calls on
the lines with `LWLockConditionalAcquire` and
`LWLockRelease`. Most of the time, Postgres is perfectly
happy to let processes do work in parallel, but there are a
few places where locks need to be acquired to avoid
contention, and this is one of them. Near the beginning of
this article we touched on how transactions in Postgres
commit or abort in serial order, one at a time.
`ProcArrayEndTransaction` acquires an exclusive lock so
that it can update `latestCompletedXid` without having its
work negated by another process.

### Responding to the client (#client)

Throughout this entire process, a client has been waiting
synchronously for its transaction to be confirmed. Part of
the atomicity guarantee is that false positives where the
databases signals a transaction as committed when it hasn't
been aren't possible. Failures can happen in many places,
but if there is one, the client finds out about it and has
a chance to retry or otherwise address the problem.

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

    else if (XidInMVCCSnapshot(HeapTupleHeaderGetRawXmin(tuple), snapshot))
        return false;
    else if (TransactionIdDidCommit(HeapTupleHeaderGetRawXmin(tuple)))
        SetHintBits(tuple, buffer, HEAP_XMIN_COMMITTED,
                    HeapTupleHeaderGetRawXmin(tuple));

    ...

    /* xmax transaction committed */

    return false;
}
```

`XidInMVCCSnapshot` does an initial check to see whether
the tuple's `xid` is visible according to the snapshot's
`xmin`, `xmax`, and `xip`. Here's a simplified
implementation that shows the checks on each ([from
`tqual.c`][insnapshot]):

``` c
static bool
XidInMVCCSnapshot(TransactionId xid, Snapshot snapshot)
{
    /* Any xid < xmin is not in-progress */
    if (TransactionIdPrecedes(xid, snapshot->xmin))
        return false;
    /* Any xid >= xmax is in-progress */
    if (TransactionIdFollowsOrEquals(xid, snapshot->xmax))
        return true;

    ...

    for (i = 0; i < snapshot->xcnt; i++)
    {
        if (TransactionIdEquals(xid, snapshot->xip[i]))
            return true;
    }

    ...
}
```

Note the function's return value is inverted compared to
how you'd think about it intuitively -- a `false` means
that the `xid` _is_ visible to the snapshot. Although
confusing, you can follow what it's doing by comparing the
return values to where it's invoked.

After confirming that the `xid` is visible, Postgres checks
its commit status with `TransactionIdDidCommit` ([from
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
advertised. It calculates a location in the commit log from
the given transaction ID and reaches into it to get that
transaction's commit status. Whether or not the transaction
committed is used to help determine the tuple's visibility.

The key here is that for purposes of consistency, the
commit log is considered the canonical source for commit
status (and by extension, visibility) [3]. The same
information will be returned regardless of whether Postgres
successfully committed a transaction hours ago, or seconds
before a crash that the server is just now recovering from.

### Hint bits (#hint-bits)

`HeapTupleSatisfiesMVCC` from above does one more thing
before returning from a visibility check:

``` c
SetHintBits(tuple, buffer, HEAP_XMIN_COMMITTED,
            HeapTupleHeaderGetRawXmin(tuple));
```

Checking the commit log to see whether a tuple's `xmin` or
`xmax` transactions are committed is an expensive
operation. To avoid having to go to it every time, Postgres
will set special commit status flags called "hint bits" for
a heap tuple that is scanned. Subsequent operations can
check the tuple's hint bits and are saved a trip to the
commit log themselves.

## The box's dark walls (#black-box)

When I run a transaction against a database:

``` sql
BEGIN;

SELECT * FROM users WHERE email = 'brandur@example.com';

INSERT INTO users (email) VALUES ('brandur@example.com')
    RETURNING *;

COMMIT;
```

I don't stop to think about what's going on. I'm given a
powerful high level abstraction (in the form of SQL) which
I know will work reliably, and as we've seen, Postgres does
all the heavy lifting under the covers. Good software is a
black box, and Postgres is an especially dark one (although
with pleasantly accessible internals).

Thank you to [Peter Geoghegan][peter] for patiently
answering all my amateur questions about Postgres
transactions and snapshots, and giving me some pointers for
finding relevant code.

[1] A few words of warning: the Postgres source code is
pretty overwhelming, so I've glossed over a few details to
make this reading more digestible. It's also under active
development, so the passage of time will likely render some
of these code samples quite obsolete.

[2] Readers may notice that while `xmin` and `xmax` are
fine for tracking a tuple's creation and deletion, they
aren't to enough to handle updates. For brevity's sake, I'm
glossing over how updates work for now.

[3] Note that the commit log will eventually be truncated,
but only beyond the a snapshot's `xmin` horizon, and
therefore for the visibility check short circuits before
having to make a check in WAL.

[backendstartup]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/postmaster/postmaster.c#L4014
[clogbits]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/access/transam/clog.c#L57
[clogstatuses]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/access/clog.h#L26
[commit]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/access/transam/xact.c#L1939
[committree]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/access/transam/transam.c#L259
[didcommit]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/access/transam/transam.c#L124
[endtransaction]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/storage/ipc/procarray.c#L394
[execsimplequery]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/tcop/postgres.c#L1010
[gettup]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/access/heap/heapam.c#L478
[getsnapshotdata]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/storage/ipc/procarray.c#L1507
[insnapshot]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/utils/time/tqual.c#L1463
[peter]: https://twitter.com/petervgeoghegan
[pgproc]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/storage/proc.h#L94
[pgxact]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/storage/proc.h#L207
[satisfies]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/utils/time/tqual.c#L962
[setpagestatus]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/access/transam/clog.c#L254
[settreestatus]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/backend/access/transam/clog.c#L148
[snapshot]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/utils/snapshot.h#L52
[tuple]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/access/htup_details.h#L116
[tupleheaders]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/access/htup.h#L62
[xid]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/c.h#L397
[xidadvance]: https://github.com/postgres/postgres/blob/b35006ecccf505d05fd77ce0c820943996ad7ee9/src/include/access/transam.h#L31
