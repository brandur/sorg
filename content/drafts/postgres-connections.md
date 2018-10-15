---
title: How to Manage Connections Efficiently in Postgres, or Any Database
published_at: 2018-09-25T16:12:19Z
location: San Francisco
hook: TODO
---

You start building your new project. You've heard good
things about Postgres, so you choose it as your database.
As advertised, it proves to be a very satisfying tool and
your progress is good. You put your first version of the
project into production and just like you'd hoped, things
go smoothly as it also turns out to be well-suited for
production use.

The first few months go great and your traffic continues to
ramp up when suddenly a deluge of failures appears. You dig
into the cause and see that your application is failing to
open connections to its database. You check your logs and
find this littered throughout:

```
FATAL: remaining connection slots are reserved for
non-replication superuser connections
```

This is one of the first major operational problems that
users are likely to encounter with Postgres, and one that
might continue to be frustratingly persistent. The database
is indicating that its total number of connection slots are
limited, and that the limit has been by connections that
are already open. The ceiling is controlled by the
`max_connections` key in configuration, which defaults to
100.

Almost every cloud Postgres provider like Google Cloud
Platform or Heroku limit the number pretty carefully, with
the largest databases topping out at 500 connections and
the smaller ones at 20 or 25.

At first sight this might seem a little counterintuitive.
If connection limits are a known problem, why not just
configure the biggest number possible? It turns out that
there are a number of factors that will limit the maximum
number of connections that are practically possible; some
obvious, and some more subtle. Let's take a closer look.

## The practical limits of database concurrency (#concurrency-limits)

The most direct constraint, but also probably the least
important, is memory. Postgres is designed around a process
model where a central Postmaster accepts incoming
connections and forks child processes to handle them. Each
of these "backend" processes starts out at around 5 MB in
size, but will grow to be much larger depending on the data
they're accessing.

TODO: Postmaster diagram.

But these days it's pretty easy to procure a system where
memory is abundant, so usually memory isn't a main limiting
factor. A more subtle one is that the Postmaster and its
backend processes use shared memory for communication, and
parts of that shared space are global bottlenecks. For
example, here's the structure that tracks every ongoing
process and transaction:

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

Operations that might happen in any backend requires
walking the entire list of processes or transactions.
Adding a new process to the proc array necessitates taking
an exclusive lock:

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

There are a few such bottlenecks throughout Postgres, and
this is of course in addition to the contention you'd
expect to find around normal system resources like I/O or
CPU.

The cumulative effect is that within any given backend,
performance is proportional to the number of all active
backends in the wider system. I wrote a [benchmark] to
demonstrate this effect: it spins up a cluster of parallel
workers that each use their own connection to perform a
transaction that inserts ten times, selects ten times, and
deletes ten times before committing [1]. Parallelism starts
at 1, ramps up to 1000, and timing is measured for every
transaction. You can see from the results that performance
degrades slowly but surely as more active clients are
introduced into the system:

!fig src="/assets/postgres-connections/contention.png" caption="Performance of a simple task degrading as the number of active connections in the database increases."

So while it might be a little irking that platforms like
Google Cloud and Heroku limit the total connections even on
very big servers, they're actually trying to help you.
Performance in Postgres isn't reliable when it's scaled up
to huge numbers of connections. Once you start brushing up
against a big connection limit like 500, the right answer
probably isn't to increase it -- it's to re-evaluate how
they're being used to and try to manage them more
efficiently.

## Techniques for efficient use of connections (#techniques)

### Connection pools (#connection-pool)

A connection pool is a cache of database connections
usually local to a specific process. It's main advantage is
improved performance -- there's a certain amount of
overhead implicit to opening a new database connection both
on the client and the server. By checking a connection back
into a connection pool instead of discarding it after
finishing with it, that connection can be reused the next
time one is needed. Connection pooling is built into many
database adapters including Go's
[`database/sql`][databasesql], Java's [JDBC][jdbc], or
Active Record in Ruby.

TODO: Diagram of per-node connection pools pointing back to

Another reason to use a connection pool is to help manage
connections more efficiently. They allow a maximum number
of connections to be configured for any particular node
which makes the total number of connections that you can
expect your deployment to use deterministic. Those
connections local to each node can also be shared by a much
larger number of total service processes. Applications
should be written so that they only acquire a connection
when they're serving a request, and therefore idle
processes don't need to hold a connection at all.

This is usually a bigger problem for applications that rely
on forking process for concurrency. In Rails for example,
Active Record bakes in the idea of a connection pool, but
it's of course only effective within the same process so
only a threaded server can really take advantage of it. To
work around Ruby's properties when it comes to parallelism,
Ruby workers are often deployed across processes using
something like Unicorn or Puma. Each process get its own
connection pool, and sharing becomes much less efficient.

### Minimum viable checkouts (#mvc)

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

### PgBouncer & inter-node pooling (#pgbouncer)

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
[databasesql]: https://godoc.org/database/sql
[jdbc]: https://en.wikipedia.org/wiki/Java_Database_Connectivity
