---
title: How to Manage Connections in Postgres (or Any Database)
published_at: 2018-09-25T16:12:19Z
location: San Francisco
hook: TODO
---

Connection limits are the first operational problem many
users hit.

Connection limits are low. Show table comparing different
providers.

This is because each connection is a forked process that
uses a non-trivial amount of memory. Also because each new
in flight transaction adds more potential bookkeeping
overhead to every other transaction.

> To get a bit more technical, the size of various data structures in postgres, such as the lock table and the procarray, are proportional to the max number of connections. These structures must be scanned by Postgres frequently.

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

## Example: Rails (#rails)

This article is written to be as programming language
agnostic as possible, but let's look at an example.

## pgbouncer (#pgbouncer)

Argument that it should not be in core: because your
application should already be doing this.

However, can act as a global connection pool between nodes.
This is more important in memory-inefficient (or extremely
large) applications that need to be scaled horizontally to
a massive extent (e.g. Rails).

TODO: Diagram of per-node connection pools pointing back to
pgbounder which points back to the database.


