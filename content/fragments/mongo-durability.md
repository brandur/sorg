---
title: The Long Road to Mongo's Durability
published_at: 2017-05-07T14:53:09Z
hook: A data store's journey from data loss to durable
  storage.
---

Anyone running in database circles will probably have at
some point heard someone joke about how MongoDB loses data.
Unsurprisingly, data persistence is something that database
people tend to take seriously; so seriously that it's got a
name: ***durability***, perhaps best known as the "D" at
the end of "ACID". The property states that after a
transaction has committed, it will remain committed even in
the event of power loss, crashes, or other errors.

For a long time MongoDB wasn't durable and could easily
lose data, but has since mostly redeemed itself. Here's the
story of how MongoDB won its "D".

## Not too concerned about writes (#write-concern)

MongoDB clients have a setting called `WriteConcern` that
dictates the level of certainty that they should have
before considering data they add or change to be persisted.
For the first four years of the data store's life its
default setting was `0`, which meant that the clients
didn't even wait for server acknowledgement to consider a
write successful. Confirming that the requests had made it
to the outgoing socket buffer of the local host was "good
enough".

It doesn't take much to see the problem here. Any number of
common real life occurrences could cause that data to be
lost: the client machine crashing, the failure of the
connected mongod instance, or an interruption in the
network connection that leads to a communication error.

## Disks are not webscale (#journaling)

Possibly more egregious was that until version 1.8
(released March 2011), MongoDB didn't have
[journaling][journaling]. Changes were committed in memory
and for performance reasons only flushed to disk about once
a minute. Again, the problems here are obvious in that a
crash would lose you a minute's worth of data that you'd
thought was committed.

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
burn the data it managed. On the plus side, they were all
great help for the benchmarks that help fuel its initial
craze; persistence operations are very fast when you don't
wait to see whether they worked.

To their credit, the company did eventually close most of
these holes. MongoDB 1.8 brought journaling, and as of
November 2012 its client libraries set `WriteConcern` to
`1` by default; meaning that a write is acknowledged only
after it's been confirmed to have propagated to a replica
set's primary. Real durability is possible by setting the
`j` option, although the data store's design around sync
intervals continues to make its use not performant.

MongoDB may still be missing three letters of "ACID", but
these days it's got one on the board [1].

[1] MongoDB is durable assuming that the right
    configuration tweaks are in place (`WriteConcern` `w =
    1` and `j = true`).

[journaling]: https://en.wikipedia.org/wiki/Journaling_file_system
