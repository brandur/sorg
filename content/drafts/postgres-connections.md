---
title: How to Manage Connections Efficiently in Postgres, or Any Database
published_at: 2018-09-25T16:12:19Z
location: San Francisco
hook: TODO
---

One of the first major operational problems a Postgres user
is likely to hit is this one:

```
FATAL: remaining connection slots are reserved for non-replication superuser connections
```

Postgres is indicating that its total number of connection
slots are limited, and that the limit has been reached by
connections that are already open. Increasing that limit is
easy, all it takes is changing the `max_connections` key in
configuration, but there are some serious caveats in doing
so, and not all of them are immediately obvious.

The approach of the different Postgres cloud provides when
it comes to maximum connections varies. Google's GCP and
Heroku tend to tightly constrain that ceiling, with even
the biggest available instances offering only 500
connections. Amazon RDS will allow you to have many more,
but with the serious caveat that instead of reaching a
connection limit, users are likely to start hitting
performance bottlenecks that will force them to throttle
back on their connection count anyway.

## The practical limits of database concurrency (#concurrency-limits)

There are a number of factors that limit the number of
active connections in Postgres. The most direct constraint,
but also the least important, is memory. Postgres is
designed around a process model where a central Postmaster
accepts incoming connections and forks child processes to
handle them. Each of these "backend" processes starts out
at around 5 MB in size, but will grow to be much larger
depending on the data they're accessing.

TODO: Postmaster diagram.

It's pretty easy to procure a system where memory is
abundant though, so most often memory isn't the main
limiting factor. A more subtle one is that the Postmaster
and its backend processes use shared memory for
communication, and parts of that shared space are global
bottlenecks. For example, here's the structure that tracks
every ongoing process and transaction:

``` c
typedef struct PROC_HDR
{
    /* Array of PGPROC structures (not including dummies for prepared txns) */
    PGPROC	   *allProcs;
    /* Array of PGXACT structures (not including dummies for prepared txns) */
    PGXACT	   *allPgXact;

    ...
}
```

Some bookkeeping operations that might happen in any
backend requires walking the entire list of processes or
transactions. Adding a new process to the proc array
necessitates taking an exclusive lock:

``` c
void
ProcArrayAdd(PGPROC *proc)
{
    ProcArrayStruct *arrayP = procArray;
    int            index;

    LWLockAcquire(ProcArrayLock, LW_EXCLUSIVE);

    ...
}
```

There are at least a few such bottlenecks throughout
Postgres, and this is of course in addition to the normal
contention you'd expect to find around I/O, CPU, and other
system resources.

The cumulative effective is that within any given backend,
performance is proportional to the number of all active
backends in the wider system. I wrote a [simple
benchmark][benchmark] to demonstrate this effect: it spins
up a number of parallel tasks that each use their own
connection to perform a simple transaction that inserts ten
times, selects ten times, and deletes ten times before
committing [1]. Parallelism starts at 1, ramps up to 1000,
and timing is measured for every task. You can see from the
results that performance slowly but surely degrades as more
active clients are introduced into the system:

!fig src="/assets/postgres-connections/contention.png" caption="Performance of a simple task degrading as the number of active connections in the database increases."

So while it might be a little irking that platforms like
Google Cloud and Heroku limit the total connections even on
very big machines, they're really trying to help you.
Performance isn't reliable once a Postgres system is scaled
up to huge numbers of connections.

One you start brushing up against a big connection limit
like 500, the right answer probably isn't to increase it --
it's to re-evaluate how you're using connections and try to
manage them more efficiently. Let's take a look at a few
techniques for doing so.

## Connection pools (#connection-pool)

TODO: Diagram of per-node connection pools pointing back to

This is usually a bigger problem for applications that rely
on forking process for concurrency. In Rails for example,
Active Record bakes in the idea of a connection pool, but
it's of course only effective within the same process so
only a threaded server can really take advantage of it. To
work around Ruby's properties when it comes to parallelism,
Ruby workers are often deployed across processes using
something like Unicorn or Puma. Each process get its own
connection pool, and sharing becomes much less efficient.

## Minimum viable checkouts (#mvc)

Check connections out for the minimum amount of time
needed, then check them back in.

TODO: Diagram of connection checkout lifetime over request.

Try to do as much before and after the checkout as you can.
Before: decoding and processing request payload, logging,
rate limiting, ... After: metrics emission, serialize
response, send response back to user (it's silly to keep a
connection checked out while waiting on this long-lived I/O
to complete)
the database.

## PgBouncer & inter-node pooling (#pgbouncer)

Argument that it should not be in core: because your
application should already be doing this.

However, can act as a global connection pool between nodes.
This is more important in memory-inefficient (or extremely
large) applications that need to be scaled horizontally to
a massive extent (e.g. Rails).

TODO: Diagram of per-node connection pools pointing back to
pgbounder which points back to the database.

[1] My simple benchmark is far from rigorous. While it
measures degradation, it makes no attempt to identify what
the core cause of that degradation is, whether it be locks
in Postgres or just I/O. It's mostly designed to prove that
the degradation exists.

[benchmark]: https://github.com/brandur/connections-test
