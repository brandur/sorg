---
hook: Why MongoDB is never the right choice.
location: San Francisco
published_at: 2016-07-26T17:48:35Z
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

By using sharding, MongoDB allows a large data set to be spread across many
different compute nodes. This by extension distributes the workloads on that
data across those nodes as well, resulting in reduced stress on any one
machine's resources.

> _I suppose it is tempting, if the only tool you have is a hammer, to treat
> everything as if it were a nail._
>
> &mdash; Abraham Maslow, 1966

While a great idea in theory, in my experience that when easily available, it's
vastly more likely for sharding to be abused than used appropriately.
Enthusiastic engineers will inevitably shard prematurely and unwisely, and the
darker sides of sharing quickly start to become apparent:

* Cross-record ACID guarantees are gone forever.
* Data becomes difficult to find and cross-reference because it exists across a
  number of different systems. This makes the job of every operator more
  difficult forevermore.
* Application code becomes bloated with sharding logic and slowly becomes
  hugely complex and bloated.
* It becomes apparent that a poor sharding strategy was used (early sharding
  decisions are especially prone to this), and certain nodes start to run
  disporportionately "hot". In many situations, this problem can be nearly
  impossible to fix, especially when a single node starts to push the limits of
  vertical scalability.

Of course I wouldn't go as far to say that sharding is _never_ appropriate, but
it can be avoided by most companies in most situations. Some good alternatives
are:

1. **Delete old information.** This is by far the best option if at all
   possible because it keeps systems lean and simple.
2. If data can't be deleted, **archive it to other scalable data stores**. If
   it almost never needs to be accessed, batches in S3 are perfect. If it only
   ever needs to be accessed internally, Redshift is great. If it needs to be
   accessed occasionally by the public, DynamoDB might be appropriate.

The underlying question for any data set should be, _what do we actually need
to keep?_ Spending some time in answering it will almost certainly result in a
lean core with fringe data moved to scalable stores, and investing in that
model will pay out in dividends in reduced resources and engineering burden
over time.

#### Example: Webhooks

A company I've worked for decided to implement WebHooks. Because sharding was
readily available, the engineers in charge decided that it wouldn't be bad idea
to just store every WebHook notification that ever went out, and all the
interactions that we'd had with remote servers trying to deliver it. Worse yet,
all of this was exposed through an API that some subset of customers started to
depend on.

We're now stuck in a complex situation where we have to manage an online data
set on the scale of 10s of TBs, most of which is _never_ accessed, but which
should ideally remain online to maximize backwards compatibility. This is
hugely expensive in both computing resources and engineer time.

What _should_ have happened from the beginning is that if it was very important
to have a WebHooks paper trail going back to the beginning of time, old events
should have been moved offline to an archive like S3 so that they could be
audited at some later time. More practically, it's probably not even worth
going that far, and events could conceivably just be purged completely after a
reasonable 30 or 90 day timeframe, leaving a data set small enough to run on a
single node forever for everyone except a Google-sized system.

### Well, if nothing else, at least it's HA!

It's true that MongoDB implements a form of easy high availability (HA) by
allowing automatic failover when the current primary becomes unavailable by
electing a secondary to primary in its place. It's worth nothing though that
this isn't too much different from the HA schemes implemented in Postgres by
services like [AWS][aws-ha] or [Heroku][heroku-ha], which essentially do the
exactly the same thing as Mongo by looking for an unhealthy primary and if
found, promoting a follower in its place.

More imporantly though, HA is often not as much about the technical
sophistication of a sytem as it is about the technical processes surrounding
it. Everyone imagines that the major dangers to the availability of a data
store are disk failures and network partitions, and one in a while those things
really do happen. However, in most day-to-day development, a database has many
more existential issues problems that are vastly more likely to occur:

* Operator error in which somebody accidentally mangles some piece of critical
  data and brings an online system to its knees.
* Overly expensive migrations locking schema/data or eating all available
  resources.
* Poorly vetted deploys in which new code expects a certain schema or data
  before it's actually updated, and which cause a failure once they go out.
* Long-lived transactions or other jobs that starve other online operations of
  resources.

In practice, an HA data store helps you, but not as much as you'd think. I've
seen as much or more downtime on a large Mongo system as I have on a Postgres
system of similar scale; none of it was due to problems at the network layer.

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

[aws-ha]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.MultiAZ.html
[heroku-ha]: https://devcenter.heroku.com/articles/heroku-postgres-ha
[two-phase]: https://docs.mongodb.com/manual/tutorial/perform-two-phase-commits/
