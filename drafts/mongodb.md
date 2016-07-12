---
hook: Why MongoDB is never the right choice.
location: San Francisco
published_at: 2016-06-08T03:57:09Z
title: MongoDB
---

Operated a large Postgres installation, then moved to being a regular user of a
large MongoDB cluster.

They've heard the stories and imagine that using MongoDB is pretty bad. What
they don't expect is what I tell them next, "it's much worse than you think."

_M:I reference?_

## Non-issues

### Data Loss

I'll give them a pass; every early database has this problem.

### Bad Benchmarks

I'm willing to give them the benefit of the doubt here by applying Hanlon's
Razor: I think it's far more likely that the botched benchmarks were the result
of incompetence than malice.

### Failure to Comply to CAP

## Problems

### No Transactions

What happens in a big MongoDB-based production system when a request that
commits multiple documents fails halfway through? Well, it's exactly what you
would think given a few moments to think about it: Mongo only guarantees
consistency within updates of a single document, so if you fail between
documents, you're left with inconsistent data.

In the optimal system, you have an automated process that attempts to identify
this class of failure and clean them up by reverting data to a consistent
state. But here in the real world, with deadlines and scarce engineering time,
you'll almost certainly have a human operator that dives in and _manually_
repairs that bad data. Remember that the process could have been cut off
between _any_ two Mongo commits, so you could be left with an innumerable
number of edge cases that are difficult to compensate for with an automated
repair system.

Mongo recommends that you solve this problem by [implementing two-phase commits
in your application][two-phase]. This is certifiably _insane_. Putting your own
two-phase commit into even one place is time consuming and complex. A real-life
product may have hundreds of interacting domain objects; putting two-phase
commit in every time you want guaranteed consistency between two of them is a
recipe for multiplying your project's development time by 100x for no good
reason at all.

Serialization transactions are magic.

### No Atomicity

Mongo supports atomic operations at the document level. Despite what you might
read in their documentation, in a system anchored in the real world,
document-level atomic operations are about as useful as _no atomic operations
at all_. That's because any non-trivial computation is almost certainly going
to operate on multiple documents, and not having strong atomicity guarantees is
going to bring you into a world of contention, failure, and pain.

So how do you deal with this in a Mongo-based production system? _You implement
locking yourself_.

Yes, you read that right. Instead of having your mature data store take care of
this tremendously difficult problem for you, you pull it into your own
almost-certainly-buggy application-level code. And don't think for a minute
that you're going to build in the incredibly sophisticated optimistic locking
features you get with any modern RDMS; no, to simplify the complicated problem
and save time, you're going to build a pessimistic locking scheme. That means
that simultaneous accesses on the same resource will block on each other to
modify data, and make your system irreparably slower.

### No Constraints

### Analytics

## Non-solutions

### The Oplog is sure cool.

If you're tailing an oplog, you're communicating between components using
private implementation details. Your entire system becomes inherently fragile
because internal changes to how data is stored can take down everything else.
Don't do it. Use public APIs instead.

That's the easy and highly ideal answer. Perhaps worse yet, 

You should not be using the oplog except in very specialized storage-related
cases.

### But at least it's scalable right?

Junk data.

### Well, if nothing else, at least it's HA!

Everyone imagines that the major dangers to the availability of a data store
are disk failures and network partitions.

_Sometimes_ those things do cause problems. However, here's what you actually
need to worry about in real life:

* Operator error in which somebody accidentally mangles some piece of critical
  data and brings an online system to its knees.
* Overly expensive migrations locking schema/data or eating all available
  resources.
* Poorly vetted deploys in which new code expects a certain schema or data
  before it's actually updated, and which cause a failure once they go out.
* Long-lived transactions or other jobs that appropriate resources from other
  online operations.

In practice, an HA data store helps you a bit, but not as much as you'd think.
I've seen as much or more downtime on a large Mongo system as I have on a
Postgres system of similar scale.

## Summary

If you're already on MongoDB, it may be very difficult to migrate off of and
staying on it might be the right choice for your organization. I can relate.
But if you're not in that situation, and are considering using MongoDB in a new
project, please, please, _please_ reconsider. MongoDB for any new system is
_never_ the right choice.

**Do you need document-style storage (i.e. nested JSON structures)?** You
probably don't, but if you really _really_ do, you should use the `jsonb` type
in Postgres instead of Mongo. You'll get the flexibility that you're after [2],
but also an ACID-compliant system and the ability to introduce constraints [3].

**Do you need incredible scalability that Postgres can't possibly provide?**
Unless you're Google or Facebook, you probably don't, but if you really
_really_ do, you should store your core data (users, apps, payment methods,
servers, etc.) in Postgres, and move those data sets that need super
scalability out into separate scalable systems _as late as you possibly can_.
The chances are that you'll never even get to that point, and if you do, you
may still have to deal with some of the same problems that are listed here, but
at least you'll have a stable core.

## References

https://news.ycombinator.com/item?id=11857674

http://cryto.net/~joepie91/blog/2015/07/19/why-you-should-never-ever-ever-use-mongodb/

Analytics failure:
https://www.linkedin.com/pulse/mongodb-32-now-powered-postgresql-john-de-goes

[1] Okay, this is embellished for dramatic effect. Sometimes a resource failure
    will take a system down, but in my experience, these incidents are dwarfed
    by those incited by user error.

[2] Although, this sort of flexibility may not be as good of an idea as you
    might think.

[3] In Postgres, try creating a `UNIQUE` index on a predicate that uses a JSON
    selector to query into a JSON document stored in a field. It works, and is
    incredibly cool.

[two-phase]: https://docs.mongodb.com/manual/tutorial/perform-two-phase-commits/
