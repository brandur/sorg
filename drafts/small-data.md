---
hook: Because everyone thinks that they're Google.
location: San Francisco
published_at: 2016-04-22T21:41:42Z
title: In Defense of Small Data
---

About a year ago at Heroku, we saw that our main database was approaching the 1
TB horizon. Postgres doesn't have any hard limits around how much data it can
handle, but tables beyond a certain size do start to become slow and unwieldy
[1].

We _could have_ taken the opportunity to initiate a migration project that
would take us over to Mongo or Cassandra so that we could scale out forever.
What we did instead was audit the information we stored in the database and
challenge our assumptions as to what we needed and what we didn't. Over the
next couple weeks we dropped out ~1 TB monster down to about 200 GB.

That was enough to get us well within our comfort levels, but the bulk of the
information that remained also belonged to a single field on a single table
that had a high-entropy data set that didn't lend itself well to compression
(in this case large, encrypted data blobs). If we'd gone through a little more
effort, we could have optimized it for storage at the application level and
knocked off another 150 GB or so, bringing a database that stores every user,
app, and release that's ever occurred on the Heroku platform down to a neat 50
GB.

This isn't to say that _everyone_ can compact their data so easily, but I'd
venture to guess that more people can than they think.

## Scability: A Blessing and a Curse (#scalability)

Being able to scale to infinity is a powerful tool, but one which is ripe for
abuse. It's easy to lead yourself to believe that it'll be useful to have hot
access to some particularly volumous data set. When you realize later that
maybe that's not the case, it can often be less effort to just add more nodes
to continue scaling the storage systems rather than just fix the problem.

In the example above, out of roughly 700 GB that were pruned from the database,
about 615 GB of it was a vast collection of tuples that we called "events",
they represented a change that had occurred in some kind of domain object and
included a full representation of the object at the time of the change. The
system had been intended for big things, but the vision hadn't been fully
realized and it had mostly become a convenient mechanism to power a few
non-critical systems.

Not only should we have not had 600+ GB of this data in our primary database,
we should have had exactly 0 bytes worth of it.

The vast majority of Mongo nodes ended up holding volumous, but low-quality data.

## What You'll Miss (#miss)

### You'll Miss ACID (#acid)

ACID is magic.

See transaction isolation levels.

Mongo's brand new WiredTiger database engine now supports document-level
concurrency control (what might be called row-level concurrency in an RDMS).
Compare this to Postgres, where you get similar guarantees of integrity not
just across throughout entire tables, but also across _the entire database_.

### Locking, or Lack of It (#locking)

### Constraints (#constraints)

Your database guarantees consistency so that you don't have to.

Lock down every goddamn field.

## Strategies (#strategies)

### Look For Junk (#junk)

### Consider Archiving (#archive)

[1] The Heroku Postgres systems would have to start growing our EBS volumes,
    which was also somewhat undesirable.
