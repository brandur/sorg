---
title: MongoDB's Discovered "D"
published_at: 2017-05-04T23:44:24Z
hook: A data store's long path to durability.
---

Anyone running in database circles will probably have at
some point heard someone joke about how MongoDB loses data.
Unsurprisingly, data persistence is something that database
people tend to take seriously; so seriously in fact, that
it's even got a name: ***durability***, most frequently
observed in the wild as the "D" at the end of "ACID". The
property states that after a transaction has committed, it
will remain committed even in the event of power loss,
crashes, or other errors.

## Not too concerned about writes (#write-concern)

MongoDB clients have a setting called `WriteConcern` which
dictates how they handle persistence. For the first four
years of the data store's life its default setting was `0`
which means that the clients didn't wait to acknowledge
writes. Confirming that the requests had made it to the
outgoing socket buffer of the client host was considered
"good enough".

It doesn't take much to see the problem here. Any number of
degenerate cases could cause that data to be lost: the
client machine crashing, the failure of the connected
mongod instance, or an interruption in the network
connection that leads to a communication error.

## Writes to disk are not webscale (#journaling)

Possibly even more egregious was that until version 1.8
(released March 2011), MongoDB didn't have journaling.
Changes were committed in memory and for performace reasons
flushed to disk about once a minute. Again, the problems
here are obvious in that a crash would lose you a minute's
worth of data that you'd thought was committed.

In 1.8 journaling was added, but still came with a caveat:
it was only synced to disk on a regular "commit interval"
of ~100 ms. Once again, a crash would cause the loss of any
data that wasn't written during that interval. A second
`WriteConcern` option called `j` (for "journaling"; the
first is named `w`) was added that lets clients specify
that they want to wait for the journal to sync to disk
before returning.

## A more durable future (#future)

The DBAs who joked about MongoDB losing data were right.
For a long time it had multiple options at its disposal to
throw your data out the window. On the plus side, they were
great help for the benchmarks that help fuel its initial
hype; persistence operations are very fast when you don't
wait to see whether they worked.

To their credit, the company did eventually close most of
these holes. MongoDB 1.8 brought journaling, and as of
November 2012 its client libraries set `WriteConcern` to
`1` by default, which tells them to only acknowledge the
write after it's been confirmed to have propagated to a
replica set's primary. Full durability is possible by
setting the `j` option, although the data store's design
around sync intervals continues to make its use not
performant.

MongoDB may still be missing three letters of "ACID", but
these days it's got one on the board.
