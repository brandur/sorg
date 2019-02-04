---
title: How Postgres Makes Transactions Serializable
published_at: 2018-09-07T15:36:33Z
location: San Francisco
hook: TODO
tags: ["postgres"]
---

In my decades working with computers, `SERIALIZABLE`
transactions are one of the few things that have truly
astounded me. A lot of technology is impressive, but it's
usually possible to get a rough idea for how it works when
you think about it a bit. Serializable transactions were
different for me in that I had _no idea_ how what they were
doing was even possible.

The easiest way to demonstrate this is through example.
Take the common task of creating a new user account. It's
possible that the account has already been created, so we
want to check that the same user doesn't already exist
before creating a new one. The SQL generated from an
application doing this might look a little like this:

``` sql
BEGIN;

SELECT * FROM users
    WHERE email = 'jane@example.com';

INSERT INTO users (email)
    VALUES ('jane@example.com');

COMMIT;
```

Let's pretend for a moment that there is no `UNIQUE`
constraint on `email`. It's immediately obvious that
there's a potential race condition here: two transactions
could be executing nearly simultaneously, and both be
running their `SELECT` before either has called called
`INSERT` and committed. The result is that we have two user
entries for the same person instead of one, and indeed if
you try, you'll be able to produce this anomaly.

But the results change if we set the transaction's
isolation level to `SERIALIZABLE` [1]:

``` sql
BEGIN;
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
```

The multiple user anomaly describe above has become
impossible. When multiple of these transactions are running
simultaneously all but one of them will be aborted when it
tries to commit with an error like this:

```
ERROR:  could not serialize access due to read/write dependencies among transactions
DETAIL:  Reason code: Canceled on identification as a pivot, during commit attempt.
HINT:  The transaction might succeed if retried.
```

This is just a very simple example (in practice, you'd just
add a `UNIQUE` constraint on `email` because that's cheaper
to enforce), but `SERIALIZABLE` will additionally protect
against every arbitrary serialization anomaly that can
occur, including those which are much more subtle and
complex. This makes it one of the most powerful
abstractions for building reliable software in existence,
and it can be a huge boon for reducing bugs and improving
data consistency in well-designed systems.

But `SERIALIZABLE` is also a complex idea, and until
relatively recently systems that wanted to provide used a
heavy strict two-phase locking (S2PL, more on this below)
to implement it which offered poor performance in
transaction throughput. In 2006, the paper [_Serializable
Isolation for Snapshot Databases_][paper] by Cahill et al.
described a new approach to the problem called SSI
(serializable snapshot isolation) which made the same
serializable guarantees, but far more efficiently than
S2PL. The Postgres community was quick to adopt it, and an
implementation of SSI was committed in 2011 and made in
into the [9.1 release][postgres91].

## Serializability (#serializability)

Let's briefly define ***serializability***, term that's
used broadly in the science of database transaction
processing, and far beyond just Postgres.

I find that a really helpful way to think about it is to
imagine a very naive database implementation that can only
process one transaction at time. If a transaction is being
processed already when another come in, the new transaction
has to wait in a queue until the first one finishes.

TODO: Diagram of serially running transactions.

Although too limited for real-world use, this naive
database has at least one property that's extremely
desirable: there are never any surprises. Transactions
might normally have to worry about writes being made by
concurrently running transactions, if `tx2` wrote a new
user record that `tx1` had already written for example, but
not so here. All transactions see a perfectly consistent
state elsewhere in the database throughout their entire
lifespan. `tx2` runs strictly after `tx1` and sees that the
user record has already been written.

Of course such a database is hopelessly untenable for most
use cases. Transaction throughput is an important feature,
and we can't have transactions queueing up one at a time to
run because the wait time would be incredible in any high
volume system. This is why any modern database system
handles transactions in parallel. This opens up the
possibility of side effects caused by the concurrency, so
these databases build in various mechanics to compensate
for them.

Serializability is the strongest transaction concurrency
guarantee that a database can offer. **We say that a number
of concurrently running transactions are serializable if
their outcome is the same as if the transactions had all
executed serially** (i.e., not concurrently) like in our
naive implementation above.

### Isolation levels and serializable (#isolation-levels)

The SQL standard defines four levels of transaction
isolation, shown in the table below. Rather than being
defined by the guarantees they make, isolation levels are
defined by the various phenomena caused by conflicts with
concurrent transactions that are _not allowed_.

<figure>
  <div class="table-container">
    <table class="overflowing">
      <tr>
        <th>Isolation level</th>
        <th>Dirty read</th>
        <th>Nonrepeatable read</th>
        <th>Phantom read</th>
        <th>Serialization anomaly</th>
      </tr>
      <tr>
        <td><strong>Read uncommitted</strong></td>
        <td>Allowed (but not in Postgres)</td>
        <td>Possible</td>
        <td>Possible</td>
        <td>Possible</td>
      </tr>
      <tr>
        <td><strong>Read committed</strong></td>
        <td>Not possible</td>
        <td>Possible</td>
        <td>Possible</td>
        <td>Possible</td>
      </tr>
      <tr>
        <td><strong>Repeatable read</strong></td>
        <td>Not possible</td>
        <td>Not possible</td>
        <td>Allowed (but not in Postgres)</td>
        <td>Possible</td>
      </tr>
      <tr>
        <td><strong>Serializable</strong></td>
        <td>Not possible</td>
        <td>Not possible</td>
        <td>Not possible</td>
        <td>Not possible</td>
      </tr>
    </table>
  </div>
  <figcaption>Transaction isolation levels and the
    contention phenomena that they allow. See the <a
    href="https://www.postgresql.org/docs/current/static/transaction-iso.html">Postgres
    docs</a> if you want to learn more.</figcaption>
</figure>

The meanings of the concurrent phenomena:

* **Dirty read:** A transaction reads data written by a
  concurrent uncommitted transaction.
* **Non-repeatable read:** When a transaction re-reads data
  it had read previously it finds that data has been
  modified by another transaction.
* **Phantom read:** When a transaction re-executes a query
  returning a set of rows for a search condition and finds
  that the set has changed due to another transaction.
* **Serialization anomaly:** The outcome of of a set of
  concurrently running transactions is inconsistent with
  all the possible orderings of running those transactions
  one at a time (the same as discussed above).

The SQL standard only defines the maximum phenomena allowed
for each isolation level and individual implementations are
allowed to provide stronger guarantees for each level if
they want to. Postgres does, and as a result there is
effectively no such thing as `READ UNCOMMITTED` in
Postgres. You can set it, but the result will be identical
to if `READ COMMITTED` had been used.

## Strict two-phase locking (S2PL) (#strict-two-phase-locking)

Before SSI, a common method to implement `SERIALIZABLE` was
strict two-phase locking (S2PL). Under this system,
transactions acquire a shared lock data before reading it,
and an exclusive lock before writing it. Acquiring locks
and writing data is the "first phase" suggested by the
name. After all the necessary locks have been acquired and
data written, they're all released atomically and the
transaction commits. That's phase two.

This is a costly implementation because transactions have
to block on each other waiting for locks to be released. It
also introduces the possibility of deadlocks as two
transactions each acquire their own lock and then try to
wait on the lock acquired by the other. This inefficiency
was an incentive that would lead to the design of SSI.

## Serializable snapshot isolation (SSI) (#ssi)

Cahill's paper builds on previous work in the field and
defines the basic building block of an interdependency
between transactions as an **rw-dependency**. An
rw-dependecy occurs between `tx1` and `tx2` if `tx2` writes
data after `tx1` has read it.

### Tracking dangerous structures (#dangerous-structures)

## Implementation in Postgres (#postgres)

[1] Postgres provides stronger guarantees on its isolation
    levels than necessary, so in practice both the
    `SERIALIZABLE` _and_ `REPEATABLE READ` levels will
    prevent this anomaly.

[paper]: https://dl.acm.org/citation.cfm?id=1620587
[postgres91]: https://www.postgresql.org/docs/10/static/release-9-1.html#id-1.11.6.131.3
