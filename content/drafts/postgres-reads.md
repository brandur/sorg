---
title: Using Postgres WAL to Eradicate Stale Reads
published_at: 2017-11-11T23:54:10Z
hook: TODO
---

Read replicas are a common pattern in databases to help
scale workload without having to resort to more complex
strategies like partitioning. Most relational databases
like Postgres, MySQL, and SQL Server are single master
systems, and so all writes have to go the primary. Read
operations however can plausibly be routed to the primary
_or_ any of its replicas [1].

This is useful because it allows an application to start
distributing a considerable amount of its load amongst all
available database nodes. It's especially useful when you
consider that while _most_ write operations have relatively
predictable performance because they're often insert,
updating, or deleting just single records, reads are often
much more elaborate and by extension, expensive for the
database to perform.

TODO: Diagram of writes to primary and reads on replicas.

Even as part of a normal application's workload (barring
analytical queries that can be even more complex), we might
join on two or three different tables in order to perform
an eager load, or even just have to read out a few dozen
rows as a user views a single page. With a suitably
configured application, all of this work can be offloaded
to replicas.

## Stale reads (#stale-reads)

While running reads on replicas is a high-impact and
low-effort win for scalability, it's not without its
challenges. The technique introduces the possibility of
***stale reads*** that occur when an application reads from
replica before it's had a chance to receive information
that's been committed to the primary. To a user this might
look like them updating some information, but then when
trying to view what they updated seeing stale data
representing pre-update state.

TODO: Diagram of stale read.

Stale reads are a race condition. Modern databases
operating over low latency connections are able to keep
replicas following their primary _very_ closely, and may
spend most of their time less than a second out of date,
meaning that even systems using read replicas without any
techniques for mitigating stale reads will probably produce
correct results most of the time.

As software engineers interested in building highly robust
systems, most of the time isn't good enough. We can do
better. Let's take a look at how we can ensure that stale
reads _never_ occur. We'll use Postgres's understanding of
its own state of replication and some in-application
intelligence around connection management to do it.

## The Postgres WAL (#postgres-wal)

In order to come up with a working strategy for avoiding
stale reads, we'll first need to understand a little bit
about how replication works in Postgres.

Postgres commits all changes to a ***WAL*** (write-ahead
log) for durability reasons. Every change is written out as
a new entry in the WAL and it acts the canonical reference
as to whether any change in the system occurred --
committed information is written to a data directory like
you might expect, but is only considered visible if the WAL
confirms that its associated transaction is committed (see
[How Postgres makes transactions
atomic](/postgres-atomicity) for more on this subject).

Changes are written to the WAL one entry at a time and each
one is assigned a ***LSN*** (log sequence number). Changes
are batched in 16 MB ***WAL segments***.

### The WAL's role in replication (#wal-replication)

A Postgres database can dump a representation of its
current state to a base backup which can be used to
initialize replica. From there, the replica can stay in
lockstep with its primary by consuming the changes that it
finds in its emitted WAL. A base backup comes with a
pointer to the current LSN so that when a replica starts to
consume the WAL, it knows where to start.

!fig src="/assets/postgres-reads/replicas-and-wal.svg" caption="A replica being initialized from base backup and consuming its primary's WAL."

There are a few methods available to a replica to consume
WAL from its primary. One is "log shipping" where completed
WAL segments are copied from a primary to secondaries and
consumed as a single batch. This has the major advantage of
efficiency, but at the cost of how closely any secondary
can be following its primary (secondaries will be at least
as behind as the current segment that's still being
written). Another is streaming, where WAL is emitted to
secondaries over an open connection. This has the advantage
of secondaries being very up to date at the cost of some
extra resource usage. It also conveys some other advantages
like having secondaries ready to fail over at a moment's
notice, and allowing secondaries to keep their primary
appraised of their progress (hot standby feedback).

Streaming replication is probably far and away the most
common method in online Postgres installations due to its
advantage of keeping replicas ready to take over from their
primary. The technique we're about to discuss will be able
to make replica reads more often when WAL is being
streamed, but will work with either method.

## Monitoring log position (#log-position)

### Tracking replica locations (#replica-locations)

LSN = Log sequence number

Earlier location when querying on replica:

```
mydb=# select pg_last_wal_replay_lsn();
 pg_last_wal_replay_lsn
------------------------
 0/15E88D0
(1 row)
```

Later location:

```
mydb=# select pg_last_wal_replay_lsn();
 pg_last_wal_replay_lsn
------------------------
 0/160A580
(1 row)
```

On primary:

```
mydb=# select pg_last_wal_replay_lsn();
 pg_last_wal_replay_lsn
------------------------

(1 row)
```

Get current location on primary:

```
mydb=# select pg_current_wal_lsn();
 pg_current_wal_lsn
--------------------
 0/160A580
(1 row)
```

Get how far ahead or behind one location is compared to another:

```
mydb=# select pg_wal_lsn_diff('0/160A580', '0/160A580');
 pg_wal_lsn_diff
-----------------
               0
(1 row)

mydb=# select pg_wal_lsn_diff('0/160A580', '0/15E88D0');
 pg_wal_lsn_diff
-----------------
          138416
(1 row)

mydb=# select pg_wal_lsn_diff('0/15E88D0', '0/160A580');
 pg_wal_lsn_diff
-----------------
         -138416
(1 row)

```

https://www.postgresql.org/docs/current/static/wal-intro.html
https://www.postgresql.org/docs/current/static/functions-admin.html
http://sequel.jeremyevans.net/rdoc/files/doc/sharding_rdoc.html

[1] A note on terminology: I use the word "replica" to
refer to a server that's tracking changes on a primary.
Common synonyms for the word include "standby", "slave",
and "secondary", but I'll stick to "replica" for
consistency.
