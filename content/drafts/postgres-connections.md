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

!fig src="/assets/postgres-connections/process-model.svg" caption="A simplified view of Postgres' forking process model."

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

## Techniques for efficient connection use (#techniques)

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

!fig src="/assets/postgres-connections/connection-pooling.svg" caption="A deployment with a number of nodes, each of which maintains a local pool of connections for their workers to use."

Another reason to use a connection pool is to help manage
connections more efficiently. They allow a maximum number
of connections to be configured for any particular node
which makes the total number of connections that you can
expect your deployment to use deterministic. Those
connections local to each node can also be shared by a much
larger number of total service processes. Applications
should be written so that they only acquire a connection
when they're serving a request, so idle processes don't
need to hold a connection at all.

A limitation of connection pools is that they're usually
only effective within a single process. Rails implements a
connection pool in Active Record, but because Ruby isn't
capable true parallelism it's common to use forking servers
like Unicorn or Puma which makes a connection pool much
less effective because each process will need to maintain
its own [2]. Python's situation is similar.

### Minimum viable checkouts (#mvc)

It's possible to have a number of database connections
shared between a much larger pool of workers by making sure
to carefully manage the amount of time workers have a
connection checked out.

For any given span of work, very often it's possible to
identify a critical span in the middle where domain logic
is being run, and where a database connection needs to be
held. To take an HTTP request for example, there's usually
a phase at the beginning where a worker is reading a
request's body, decoding and validating its payload, and
performing other peripheral operations like rate limiting
before moving on to the application's core logic. After
that logic is executed there's a similar phase at the end
where it's serializing and sending the response, emitting
metrics, performing logging, etc.

!fig src="/assets/postgres-connections/minimum-viable-checkout.svg" caption="Workers should only hold connections as long as they're needed. There's work before and after core application logic where no connection is needed."

Workers should only have a connection checked out of the
pool while that core logic is executing. This **minimum
viable checkout** technique maximizes the efficient use of
connections by minimizing the amount of time any given
worker holds one.

### PgBouncer & inter-node pooling (#pgbouncer)

Connection pools and minimum viable checkouts will get you
a long way, but you still may reach a point where a hammer
is needed. When an application is scaled out to many
different nodes, connection pools can be used to maximize
the efficiency of connections local to any given node, but
can't do so between nodes. In most systems work should be
distributed between nodes roughly equally, but because it's
normal to use randomness to do that (through something like
HAProxy or other load balancer), an equal distribution is
by no means guaranteed.

If we have _N_ nodes and _M_ maximum connections per node,
we may have a configuration where _N_ × _M_ is greater than
the database's `max_connections` to protect against the
case where a single node is handling an outsized amount of
work. Because nodes aren't coordinating, if the whole
cluster is running close to capacity, it's possible for a
node trying to get a new connection to go over-limit and
get an error back from Postgres.

In this case it's possible to install
[PgBouncer][pgbouncer] which acts as a global connection
pool by proxying all connections to Postgres. It can be
configured with different modes of operation:

* **Session pooling:** A connection is assigned when a
  client opens a connection and unassigned when the client
  closes it.

* **Transaction pooling:** Connections are assigned only
  for the duration of a transaction, and may be shared
  around them. This comes with a limitation that
  applications cannot use features that change the "global"
  state of a connection like `SET`, `LISTEN`/`NOTIFY`, or
  prepared statements.

* **Statement pooling:** Connections are assigned only
  around individual statements. But this only works if an
  application gives up transactions, at which point it's
  losing a big advantage of using Postgres.

!fig src="/assets/postgres-connections/pgbouncer.svg" caption="Using PgBouncer to maintain a global connection pool to optimize connection use across all nodes."

Transaction pooling is the best strategy for applications
that are already making effective use of a node-local
connection pool, and will allow such an application that's
configured with an _N_ × _M_ greater than `max_connections`
to closely approach the maximum possible theoretical
utilization of available connections, and also to avoid
connection errors caused by going over-limit (although
delaying requests while waiting for a connection to become
available from PgBouncer is still possible).

Unfortunately, probably the more common use of PgBouncer is
to act as a node-local connection pool for applications
that can't do a good job of implementing their own, like a
Rails app deployed with Unicorn. Heroku provides a
buildpack that deploys a per-dyno PgBouncer for example.
It's better to use a more sophisticated technique if
possible.

## Connections as a resource (#resource)

There was a trend in frameworks for some time to try and
simplify software development for their users by
abstracting away the details of connection management. This
might work for a time, but in the long run anyone
deploying a large application on Postgres will have to
understand what's happening or they're likely to run into
trouble. It'll usually pay to understand them earlier so
that applications can be smartly architected to maximize
the efficient use of this scarce resource.



[1] My simple benchmark is far from rigorous. While it
measures degradation, it makes no attempt to identify what
the core cause of that degradation is, whether it be locks
in Postgres or just I/O. It's mostly designed to prove that
the degradation exists.

[2] Threaded deployments in Ruby are possible, but because
of Ruby's GIL (global interpreter lock), they'll be
fundamentally slower than using a process model.

[benchmark]: https://github.com/brandur/connections-test
[databasesql]: https://godoc.org/database/sql
[jdbc]: https://en.wikipedia.org/wiki/Java_Database_Connectivity
[pgbouncer]: https://pgbouncer.github.io/
