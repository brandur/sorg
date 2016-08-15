---
title: Don't Use MongoDB
hook: Why you should almost certainly use an ACID-compliant data store, even at scale.
location: San Francisco
published_at: 2016-08-01T00:23:52Z
---

After its initial release in 2009, MongoDB enjoyed quite some time in the
spotlight, and could be credited with re-popularizing the idea of a
document-oriented database. The team focused specifically on claims of superior
performance compared to its relational cousins, and the scalability made
possible by its shard-based design. From there it fell on harder times as its
performance numbers were [debunked][broken-by-design], and more people started
to understand that scalability has inherent downsides (see the infamous
"MongoDB is web scale" dialog). The system is still widely available, but most
developers have a much more measured opinion of it compared to the peak of its
glory days.

I spent many years operating a large Postgres installation before moving over
to being a regular user of a large MongoDB cluster. Needless to say, it wasn't
an improvement. I now had access to out-of-the-box sharding, but had lost
access to hundreds of other features that helped ease development, operations,
and most importantly, ensuring correctness.

These days many developers understand that MongoDB is somewhat suspect, but
when prompted, they'll more often than not cite problems with its potential for
data loss. There was a time when data loss was a real concern, but by and large
that's no longer the case. I take the position that MongoDB is _never_ an
appropriate choice for a new project, but not for the usual reasons.

I don't write this to be mean-spirited, but rather to balance some of the hype
initiatives that are still ongoing to sell MongoDB to young projects (the
["MEAN stack"][mean] for example). Migrating between data stores is incredibly
costly, and I'll consider this writing a success if I can help even one nascent
project or company avoid starting out on MongoDB, only to realize the mistake
much later when it's almost impossible to fix.

## Lightning Rods (#lightning-rods)

Lets start with some issues that tend to draw a lot of fire. While everything
here is a valid critique, none of it is what makes the choice of MongoDB a
truly costly architectural mistake.

### Data Integrity (#data-integrity)

For a very long time, MongoDB considered a write to be complete and fully
persisted [as soon as it had been buffered in the outgoing socket buffer of the
client host][broken-by-design]. The behavior wasn't even a partial guarantee of
reliable integrity, and could easily result in data loss.

Although it was _years_ before the problem was ever addressed satisfactorily,
I'm willing to give it a pass so as not to detract from more important matters.
As of version 3, MongoDB clients now default their [`w` "write concern" to
1][write-concerns], meaning that writes are not considered persisted until
confirmed by a standalone server or replica set primary. This is a sane change
that's in-line with how other databases work, and it'll ensure that data will
largely be safe on modern versions of MongoDB.

### Dishonest Benchmarks (#dishonest-benchmarks)

One of the early hot features of MongoDB was its speed, and particularly how
well it performed compared to relational equivalents. As it turns out, these
incredible speed benchmarks had more to do with its quesionable approach to
data integrity (as discussed above) rather than any advancement made by the
10gen team. After version 3 was released with safer write defaults, [it quickly
became obvious that MongoDB had lost the performance edge that it had
originally touted][broken-by-design].

Once again, I'm giving MongoDB a pass here. The application of [Hanlon's
Razor][hanlons-razor] suggests that it's much more likely that the original
MongoDB developers didn't understand that the way they were confirming writes
was problematic. They ran some benchmarks, and believing the good numbers to be
the inherent result of the system's extraordinary engineering, flouted them for
the world to see. Later, they realized that guaranteeing data integrity was
something that people cared about and which they weren't providing, and over
time curbed their bold claims.

### Distributed Trouble (#trouble)

MongoDB has performed poorly in Jepsen tests (see [inaccessible
primary][jepsen1] and [stale reads][jepsen2]). While this is undoubtedly a
problem, it's not likely to be what's going to cause you grief on a day-to-day
basis.

## Problems (#problems)

Let's talk about why MongoDB is actually a poor choice for your new production
system. It's almost entirely to do with a set of basic guarantees that have a
memorable acronym coined back in the early 80s, and which you'd probably heard
before: **ACID** (atomicity, consistency, isolation, and durability).

MongoDB historially failed to comply to every letter in ACID, but as of version
3, now only misses three out of four. Here I'll explain why the one they
finally have (durability) is good progress, but nowhere near enough.

### No Atomicity (A) (#no-atomicity)

In an ACID-comliant store, the _atomicity_ (the "A" in ACID) property
guarantees all the work execute in a single transaction either succeeds
entirely or fails, leaving the database unchanged.

MongoDB guarantees atomicity for updates on a single document, but not beyond.
That's enough if a single document contains all the data that needs to be
changed, but any non-trivial application inevitably involves a web of relations
involving many interacting models. MongoDB can't guarantee that changes between
any two of these entities are safe.

Their documentation recommends that you solve this problem by [implementing
two-phase commits throughout your application][two-phase], an idea so misguided
that it's comically depraved. Putting your own two-phase commit into even one
place is time consuming and complex. A real-life product may have hundreds of
interacting domain objects; putting two-phase commit in every time you want
guaranteed consistency between two of them is a recipe for multiplying your
project's development time and error propensity by 100x for no good reason at
all.

#### Example: Manual Request Clean-up

**TODO: This section is not complete.**

What happens in a big MongoDB-based production system when a request that
commits multiple documents fails halfway through? Well, you're left with
inconsistent data.

You can try to build an automated background system to compensate for this by
looking for patterns in your database that are likely to be problematic, but
the difficulty of this problem can be daunting. Failures are possible between
_any two changes_ in the system, which creates a factorial set of possibilities
for corruption that need to be found and addressed.

More likely, it'll be an operator's job to go through and manually fix problems
according to their human intuition.

### No Consistency (C) (#no-consistency)

In an ACID-compliant store, the _consistency_ (the "C" in ACID) property
guarantees that for any given transaction, the system will always transition
from one valid state to another. Mechanisms like constraints, cascades, and
triggers have all fired as expected before a new state is considered valid.

In practice, that means you can do a lot of useful things:

* By adding a uniqueness constraint, you can guarantee that two accounts cannot
  be created with the same email address, even if two requests try to do so
  simultaneously.
* Say that any single account owns many apps. By using a foreign key constraint
  with `ON DELETE CASCADE`, you can guarantee that no app will ever be orphaned
  if its parent account is deleted.
* Say that any single account belongs to a team. By using a foreign key
  constraint with `ON DELETE RESTRICT`, you can guarantee that a team can never
  be deleted as long as any accounts under it are still active.
* Say that you want to produce an auditing record every time an account is
  deleted. By using a database trigger, you can guarantee that an audit trail
  is produced when an account is removed.

With MongoDB, you won't get a single one of these. Ever.

If you want to check email uniqueness, you'll need to implement a locking
system for new addresses, or run a background processor that looks for and
alerts on duplicate records. To check data constraints you'll need locking
combined with application-level conditional statements sprinkled throughout
your codebase. To produce an audit trail, you'll need to implement your own
two-phase commit along with checks everywhere to make sure that nothing is
accessing uncommitted data (i.e. partially deleted account records where the
audit trail has not yet been confirmed).

By using MongoDB, you're throwing away an invaluable tool for guaranteeing that
no matter what happens in your database, data is _always_ valid. It's not
impossible to do this from application-level code, but trying to do so is
entering a world of needless complication, buggy implementations, and corner
cases.

#### Example: Orphaned Data

**TODO: This section is not complete.**

### No Isolation (I) (#no-isolation)

An ACID-compliant store can guarantee that operations spanning multiple records
are safe through _isolation_ (the "I" in ACID). Even if two transactions are
modifying the same set of records simultaneously, the database will ensure
their correctness by hiding their changes from one another. In the case where
those changes end up being incompatible, only one of those transactions is
allowed to commit.

MongoDB of course can only guarantee an atomic change at the level of a single
document. So how do you safely modify sets of related records despite MongoDB
without isolation? Well, you implement your own application-level locking
scheme of course.

Instead of having your mature data store take care of this tremendously
difficult and hugely error prone problem for you, you pull it into your own
complex, operationally heavy, and probably buggy code. And don't think for a
minute that you're going to build in the incredibly sophisticated optimistic
locking features you get with any modern RDMS; no, to simplify the complicated
problem and save time, you're going to build a pessimistic locking scheme. That
means that simultaneous accesses on the same sets of resources will block on
each other to modify data, and make your system irreparably slower.

Lack of isolation can lead to other types of even more subtle problems as well.
[Meteor wrote a good post about how MongoDB can fail to return results][meteor]
[1] that are in the process of being updated despite their data matching about
a query's search predicates before and after the update.

#### Example: Test Data Deletion

**TODO: This section is not complete.**

Note: Maybe use request locking as the example here and test data deletion for
consistency?

Data is instantaneously inconsistent as a deletion job is running through it.

### Analytics (#analytics)

By committing to MongoDB, with its sharded nature and inscrutable querying
syntax, you're also implicitly commiting to building out a secondary
warehousing system and ingestion pipeline so that it's possible to run
analytics and other types of reporting in one place with a well-known query
language like SQL. By sticking to an RDMS, you can get this almost for free by
simply keeping a non-production follower available for this use [2].

While building a data warehouse will almost certainly be eventually
appropriate, it can be a significant advantage especially for smaller companies
to avoid committing the engineering and maintenance effort necessary to
accomplish this for as long as possible so that those resources can be
allocated to more critical projects.

## Anti-features (#anti-features)

### Fast prototyping. (#fast-prototyping)

**TODO: This section is not complete.**

### The Oplog is sure cool. (#oplog)

MongoDB offers a feature called called the oplog that's used for the primary in
a replica set to stream change information which is then consumed by each
secondary to stay up-to-date. The oplog is exposed via a MongoDB API so that it
can also be read by your own services.

The oplog has traditionally been hailed as a feature that for which Postgres
has no equivalent because its physical WAL is unsuitable to be consumed by
anything but Postgres. While this may have been true before, Postgres now has
"logical" WAL options like [pglogical][pglogical] have been introduced that
will provide essentially the same functionality.

That said, you almost certainly shouldn't be consuming either the oplog or a
Postgres logical stream except under very special circumstances. Tracking
record-level changes means that you're inherently tying yourself into a
service's internal implementation, and any changes to the way it handles data
will either break integrations or require very careful and time-consuming
coordination. Don't do it.

Instead, expose public representations of data through an API. If you need a
stream, send that _public_ representation through a system like Kafka or
Kinesis.

### But at least it's scalable right? (#scalability)

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
darker sides of sharding become apparent immediately:

* Cross-record ACID guarantees and constraints are gone forever.
* Data becomes difficult to find and cross-reference because it exists across a
  number of different systems. This makes the job of every operator more
  difficult.
* Application code becomes riddled with sharding logic and trends towards
  becoming bloated and hugely complex.
* It becomes apparent that a poor sharding strategy was used (early sharding
  decisions are especially prone to this), and certain nodes start to run
  disporportionately "hot". In many situations, this problem can be nearly
  impossible to fix, especially when a single node starts to push the limits of
  vertical scalability.

Of course I wouldn't go as far to say that sharding is _never_ appropriate, but
it can be avoided by most companies in most situations. Some good alternatives:

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

#### Example: Webhooks (#webhooks)

A company I've worked for decided to implement Webhooks. Because sharding was
readily available, the engineers in charge decided that it wouldn't be bad idea
to just store every Webhook notification that ever went out, and all the
interactions that we'd had with remote servers trying to deliver it. Worse yet,
all of this was exposed through an API that some subset of customers started to
depend on.

We're now stuck in a complex situation where we have to manage an online data
set on the scale of 10s of TBs, most of which is _never_ accessed, but which
should ideally remain online to maximize backwards compatibility. This is
hugely expensive in both computing resources and engineer time.

What _should_ have happened from the beginning is that if it was very important
to have a Webhooks paper trail going back to the beginning of time, old events
should have been moved offline to an archive like S3 so that they could be
audited at some later time. More practically, it's probably not even worth
going that far, and events could conceivably just be purged completely after a
reasonable 30 or 90 day timeframe, leaving a data set small enough to run on a
single node forever for everyone except a Google-sized system.

### Well, if nothing else, at least it's HA! (#high-availability)

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
store are disk failures and network partitions, and once in a while those
things really do happen. However, in most day-to-day development, a database
has many more existential issues problems that are vastly more likely to occur:

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
system of similar scale; none of it had anything to do with problems at the
network layer.

## Summary (#summary)

If you're already on MongoDB, it may be very difficult to migrate off of and
staying on it might be the right choice for your organization. I can relate.
But if you're not in that situation, and are evaluating the use of MongoDB in a
new project, you should reconsider. MongoDB for any new system is _never_ the
right choice.

**Do you need document-style storage (i.e. nested JSON structures)?** You
probably don't, but if you really _really_ do, you should use the `jsonb` type
in Postgres instead of Mongo. You'll get the flexibility that you're after [3],
but also an ACID-compliant system and the ability to introduce constraints [4].

**Do you need incredible scalability that Postgres can't possibly provide?**
Unless you're Google, you probably don't, but if you really _really_ do, you
should store your core data (users, apps, payment methods, servers, etc.) in
Postgres, and move those data sets that need super scalability out into
separate scalable systems (or shard just those resources) _as late as you
possibly can_. The chances are that you'll never even get to that point, and if
you do, you may still have to deal with some of the same problems that are
listed here, but at least you'll have a stable core. If you actually are
Google, then sure, just use Bigtable or whatever.

[1] [Hacker News commentary][meteor-hn] for Meteor's article.

[2] It's worth noting that when using a Postgres follower for analytics, it's a
    good idea to keep systems in place to look for long-running transactions to
    avoid putting backpressure on production databases. See my article on
    [Postgres Job Queues](/postgres-queues) for more information.

[3] Although, this sort of flexibility may not be as good of an idea as you
    might think.

[4] In Postgres, try creating a `UNIQUE` index on a predicate that uses a JSON
    selector to query into a JSON document stored in a field. It works, and is
    incredibly cool.

[aws-ha]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.MultiAZ.html
[broken-by-design]: http://hackingdistributed.com/2013/01/29/mongo-ft/
[hanlons-razor]: https://en.wikipedia.org/wiki/Hanlon's_razor
[heroku-ha]: https://devcenter.heroku.com/articles/heroku-postgres-ha
[jepsen1]: https://aphyr.com/posts/284-call-me-maybe-mongodb
[jepsen2]: https://aphyr.com/posts/322-call-me-maybe-mongodb-stale-reads
[mean]: https://en.wikipedia.org/wiki/MEAN_(software_bundle)
[meteor]: https://engineering.meteor.com/mongodb-queries-dont-always-return-all-matching-documents-654b6594a827
[meteor-hn]: https://news.ycombinator.com/item?id=11857674
[pglogical]: https://2ndquadrant.com/en/resources/pglogical/
[two-phase]: https://docs.mongodb.com/manual/tutorial/perform-two-phase-commits/
[write-concerns]: https://docs.mongodb.com/manual/reference/write-concern/
