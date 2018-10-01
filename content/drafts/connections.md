---
title: How to Manage Connections in Postgres (or Any Database)
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

The most direct limitation is memory. Postgres is designed
around a process model where a central Postmaster accepts
incoming connections and forks child processes to handle
them. Each of these "backend" processes starts out at
around 5 MB in size, but will grow to be much larger
depending on the data they're accessing.

TODO: Postmaster diagram.

But even in systems where memory is abundant, there are
more subtle limitations that make bounding the maximum
number of processes a good idea. The Postmaster and its
backend processes use shared memory for communication and
certain parts of that shared memory are global bottlenecks.
For example, here's a structure that tracks every ongoing
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

Some bookkeeping operations in any backend require walking
the entire list of processes or transactions meaning that
within any given backend, performance is proportional to
the total number of active backends.

(For example.)

The approach of the different Postgres cloud provides when
it comes to maximum connections varies. Google's GCP and
Heroku tend to tightly constrain that ceiling, with even
the biggest available instances offering only 500
connections. Amazon RDS will allow you to have many more,
but with the serious caveat that instead of reaching a
connection limit, users are likely to start hitting
performance bottlenecks that will force them to throttle
back on their connection count anyway.

## The connection pool (#connection-pool)

## Minimum viable checkouts (#mvc)

## In application (#applications)

Use a transaction pool. Check connections out for the
minimum amount of time needed, then check them back in.

TODO: Diagram of connection checkout lifetime over request.

Try to do as much before and after the checkout as you can.
Before: decoding and processing request payload, logging,
rate limiting, ... After: metrics emission, serialize
response, send response back to user (it's silly to keep a
connection checked out while waiting on this long-lived I/O
to complete)

TODO: Diagram of per-node connection pools pointing back to
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


