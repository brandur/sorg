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
_or_ any of its replicas.

This is useful because it allows an application to start
distributing a considerable amount of its load amongst all
available database nodes. It's especially useful when you
consider that while _most_ write operations have relatively
predictable performance because they're often insert,
updating, or deleting just single records, reads are often
much more elaborate and by extension, expensive. Even as
part of a normal application's workload (barring analytical
queries that can be even more complex), we might join on
two or three different tables in order to perform an eager
load, or even just have to read out a few dozen rows as
part of a normal page load.

## Stale reads (#stale-reads)

But reading from a replica isn't without its challenges --
the technique introduces the possibility of ***stale
reads*** that occur when an application reads from replica
before it's had a chance to receive information that's been
committed to the primary. To a user this might look like
them updating some information, but then when trying to
view what they updated seeing stale data representing
pre-update state.

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

First, let's look at the WAL.

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
