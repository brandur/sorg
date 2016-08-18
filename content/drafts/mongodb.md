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
data loss. I'd like to set the record straight: there was a time when data loss
was a real concern, but by and large that's no longer the case. I take the
position that MongoDB is _never_ an appropriate choice for a new project, but
not for the usual reasons.

I don't write this to be mean-spirited, but rather to balance some of the hype
initiatives that are still ongoing to sell MongoDB to young projects (the
["MEAN stack"][mean] for example). Migrating between data stores is incredibly
costly, and I'll consider this writing a success if I can help even one nascent
project or company avoid starting out on MongoDB, only to realize the mistake
much later when it's almost impossible to fix.

## Lightning Rods (#lightning-rods)

Lets start with some issues that tend to draw a lot of fire. They're all valid
critiques, but none of them is what makes MongoDB a truly costly mistake.

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
the inherent result of the system's extraordinary engineering, showed them to
the world. Later, they realized that guaranteeing data integrity was something
that most people cared about, and over time curbed their claims.

### Distributed Trouble (#trouble)

MongoDB has performed poorly in Jepsen tests (see [inaccessible
primary][jepsen1] and [stale reads][jepsen2]). While this is undoubtedly a
problem, it's not likely to be what's going to cause you grief on a day-to-day
basis.

## Problems (#problems)

Let's talk about why MongoDB is actually a poor choice for your new production
system. It's almost entirely to do with a set of basic guarantees that have a
memorable acronym coined back in the early 80s, and which you've probably heard
before: **ACID** (atomicity, consistency, isolation, and durability).

MongoDB historically failed to comply to every letter in ACID, but as of version
3, now only misses three out of four. Here I'll explain why the one they
finally have (durability) is good progress, but nowhere near enough.

### No Atomicity (A) (#no-atomicity)

In an ACID-compliant store, the _atomicity_ (the "A" in ACID) property
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

#### Atomicity Example: Request Failure (#request-failure)

So now that we've established that MongoDB transactions are not atomic outside
of a single document, let's consider what happens in the very common case
of an application that manipulates multiple objects during the course of a
single request.

What happens when a request to a big MongoDB-based production system fails
halfway due to an internal error or a client disconnect? Well, the answer is
exactly what you'd hope it's not: you're left with data in an inconsistent
state.

!fig src="/assets/mongodb/request-failure.svg" caption="Demonstration of how without an atomicity guarantee, a failed request results in an invalid state of data."

There are a few techniques that can be used to try and solve the problem.
Two-phase commits between every two objects is possible, but unlikely. Another
is to build an automated system that runs out of band and tries to repair
inconsistencies by looking for certain problematic patterns. But the practical
implications of that can be daunting: failures are possible between _any two
changes_ in the system, creating an innumerable set of possibilities for
corruption that need to be identified and addressed.

More likely than either of these, it'll be an operator's job to go through and
manually fix data after failures occur; not exactly robust system design.

### No Consistency (C) (#no-consistency)

In an ACID-compliant store, the _consistency_ (the "C" in ACID) property
guarantees that for any given transaction, the system will always transition
from one valid state to another. Mechanisms like constraints, cascades, and
triggers have all fired as expected before a new state is considered valid.

In practice, that means you can do a lot of useful things:

* By adding a uniqueness constraint, you can guarantee that a user's credit
  card is never double-charged for the same purchase even if they accidentally
  clicked the "Submit" button on the purchase form twice.
* Say that any single account owns many apps. By using a foreign key constraint
  with `ON DELETE CASCADE`, you can guarantee that no app will ever be orphaned
  if its parent account is deleted.
* Say that any single account belongs to a team. By using a foreign key
  constraint with `ON DELETE RESTRICT`, you can guarantee that a team can never
  be deleted as long as any accounts under it are still active.
* Say that you want to produce an auditing record every time an account is
  deleted. By using a database trigger, you can guarantee that an audit trail
  is produced when an account is removed.

With MongoDB, you won't get access to any of these powerful tools for ensuring
correctness.

All problems need to be solved at the application level: if you want to check
for doubled charges, you'll need to implement a secondary non-sharded
collection that references the main sharded one. To check data constraints
you'll need locking combined with application-level conditional statements
sprinkled throughout your codebase. To produce an audit trail, you'll need to
implement your own two-phase commit along with checks everywhere to make sure
that nothing is accessing uncommitted data (i.e. partially deleted account
records where the audit trail has not yet been confirmed).

By using MongoDB, you're throwing away an invaluable mechanic for guaranteeing
that no matter what happens in your database, data is _always_ valid. It's not
impossible to do this from application-level code, but trying to do so is
entering a world of needless complication, buggy implementations, and corner
cases.

#### Consistency Example: User Signup (#user-signup)

If we have a service that accepts user signups, a very common constraint to put
on it is that we don't want two accounts to share the same email address. While
MongoDB has a partial solution to this with a [unique index][unique-index],
that mechanism breaks down as soon as a collection is sharded because indexes
are local to each shard.

!fig src="/assets/mongodb/sharded-user-collection.svg" caption="A collection of split across many shards, each with its own local indexes."

Uniqueness can be enforced for shard keys, but email isn't a good one because
shard key values are immutable, and we'd like users to be able to change their
email. Using it as a shard key is a poor solution at any rate because there can
only be one shard key, but there may be other characteristics that should be
unique across an entire collection.

The solution to this problem in MongoDB is to maintain a smaller, secondary
collection that's not sharded and which contains email addresses constrained
with a unique index along with a pointer to the record in the main collection.
While this is functional, it's a significant about of overhead, and MongoDB's
lack of atomicity makes it a dangerous pattern; if a failure occurs and one
collection is updated but not the other, the system can be left with two user
records with the same email, or an email record without a corresponding user.

Admittedly this example is somewhat contrived because this problem isn't
specific to MongoDB: any database will have a difficult time guaranteeing
consistency after data is stored between multiple nodes. However, MongoDB at
best provides no advantage over better databases, and at worst encourages
sharding prematurely and without adequate education around the consistency
problems that will inevitably arise from that sort of decision.

### No Isolation (I) (#no-isolation)

An ACID-compliant store can guarantee that operations spanning multiple records
are safe through _isolation_ (the "I" in ACID). Even if two transactions are
modifying the same set of records simultaneously, the database will ensure
their correctness by hiding their changes from one another. In the case where
those changes end up being incompatible, only one of those transactions is
allowed to commit.

MongoDB of course can only guarantee an atomic change at the level of a single
document. This introduces a variety of problems, but the most noticeable is
that it it's not possible to safely operate on related sets of records between
multiple processes.

It can also lead to other types of even more subtle problems. [Meteor wrote a
good post about how MongoDB can fail to return results][meteor] [1] that are in
the process of being updated despite their data matching about a query's search
predicates before and after the update.

#### Isolation Example: Test Data Deletion (#test-data-deletion)

Imagine you have an invoicing system that'd designed to bill customers on a
regular schedule. It contains a **biller** process that looks through active
subscriptions and creates invoices for them. Each invoice consists of multiple
line items, and must have line items to be valid (a $0 invoice is meaningless).

To test your system, you have a CI-based consumer that goes through and creates
new test-mode invoices, pays them, and verifies the results. The fake invoices
created by CI don't go away on their own though, so you have another
**cleaner** process that periodically cycles through the system and removes
them.

The cleaner looks for valid cleaning targets, and removes them by deleting
their line items first, then deleting the invoice itself.

The problem occurs when the biller and cleaner happen to be running
simultaneously for the same resource. The cleaner may have already removed an
invoice's line items by the time the biller gets to it, thus leaving the biller
with an invalid invoice (and probably a crash).

!fig src="/assets/mongodb/state-conflict.svg" caption="Demonstration of how state conflicts between processes are possible without consistency."

This can be solved in MongoDB in a number of ways: user-space locks, search
predicates in process code that guarantees that the biller and cleaner never
look at the same resource, or a multi-phase deletion process (the cleaner flags
resources as deleted to hide them, and then garbage collects them at its own
leisure), but each of these fragile, and both time-consuming and repetitive to
implement.

In an ACID-compliant system, the isolation property would have ensured that the
biller and cleaner never interacted because each would operate within its own
snapshot. Conflict detection in [isolation levels][transaction-isolation] like
`SERIALIZABLE` are so incredibly sophisticated that their guarantees feel like
magic.

#### Isolation Example: HTTP Requests (#http-requests)

The example above demonstrates how without isolation, access between processes
to shared resources can be problematic, but it illustrates a slightly unusual
case. There's a much more common example that most of us are very familiar with
which can be used to show the true extent of the problem: HTTP requests.

A call to a web service on `PATCH /articles/:id` might update an article itself
along with child objects like comments, tags, or advertisements. These
resources are all a common collection that updates on the same article would
contend naturally for. So what do you do if two simultaneous requests to `PATCH
/articles/mongodb` try to update one article at the same time?

!fig src="/assets/mongodb/pessimistic-locking.svg" caption="Demonstration of pessimistic locking showing 3 requests to the same resource. Each blocks the next in line."

Those concurrent requests could lead to unacceptable state conflicts, so access
to those shared resources must be controlled. This is often accomplished in
MongoDB by implementing an application-level locking scheme.

Instead of having a mature RDMS take care of this tremendously difficult and
hugely error prone problem, it gets pulled in complex, operationally
burdensome, and probably buggy application code. This will of course be a
pessimistic locking scheme as well, so each request will block the next in line
until it's finished with its update, making the system irreparably slower.

### Analytics (#analytics)

By committing to MongoDB, with its sharded nature and obscure querying syntax,
you're also implicitly commiting to building out a secondary warehousing system
and ingestion pipeline so that it's possible to run analytics and other types
of reporting in one place with a well-known query language like SQL. By
sticking to an RDMS, you can get this almost for free by simply keeping a
non-production follower available for this use [2].

While building a data warehouse will almost certainly be eventually
appropriate, it can be a significant advantage especially for smaller companies
to avoid committing the engineering and maintenance effort necessary to
accomplish this for as long as possible so that those resources can be
allocated to more critical projects.

## Anti-features (#anti-features)

### Fast Prototyping (#fast-prototyping)

A perceived advantage of MongoDB is that its lack of structure and constraints
makes it advantageous for getting applications up and running quickly; no pesky
non-nullable columns, data types, `ALTER TABLE` statements, or foreign key
constraints to deal with here!

Every hour saved by not having to worry about data structure at the dawn of a
new project will cost 1000 hours throughout the rest of its lifetime as those
small inconsistencies start to add up and eventually build to the point where
the known state of any single object is not reliable. Constraints are good, and
I'll leave it at that.

### The Oplog (#oplog)

MongoDB offers a feature called called an oplog that's used for the primary in
a replica set to stream change information to its secondaries. The oplog is
exposed via a MongoDB API as JSON so that it can also be easily consumed by
your own services.

The oplog has traditionally been hailed as a feature that for which Postgres
has no equivalent because its physical WAL is unsuitable to be ingested by
anything but Postgres. While this may have been true before, Postgres now has
"logical" WAL options like [pglogical][pglogical] have been introduced that
will provide essentially the same functionality (tentatively slated for core
inclusion by Postgres 10, which will follow version 9.6).

That said, you almost certainly shouldn't be consuming either the oplog or a
Postgres logical stream except under very special circumstances. Tracking
record-level changes means that you're inherently tying yourself into a
service's internal implementation, and any changes to the way it handles data
will either break integrations or require very careful and time-consuming
coordination. Don't do it.

Instead, expose public representations of data through an API. If you need a
stream, send that _public_ representation through a system like Kafka or
Kinesis.

### Scalability (#scalability)

Okay, this sounds dubious as an anti-feature, but hear me out.

> _I suppose it is tempting, if the only tool you have is a hammer, to treat
> everything as if it were a nail._
>
> &mdash; Abraham Maslow, 1966

By using sharding, MongoDB allows a large data set to be spread across many
different compute nodes. This by extension distributes the workloads on that
data across those nodes as well, resulting in reduced stress on any one
machine's resources.

While a great idea in theory, I've already shown above how sharding can be
disadvantageous when it comes to ensuring consistency. Also, it's vastly more
likely for sharding to be abused than used appropriately. Enthusiastic
engineers will inevitably shard prematurely and unwisely, and the darker sides
of sharding become apparent immediately:

* Cross-record ACID guarantees and constraints are gone forever.
* Data becomes difficult to find and cross-reference because it exists across a
  number of different systems. This makes the jobs of operators more difficult.
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

A company I've worked implemented Webhooks as a streaming API to notify
customers of changes occurring in their account. Okay, so far so good.

Because sharding was readily available, the engineers in charge decided that it
wouldn't be bad idea to just store every Webhook notification that ever went
out for any change that had ever occurred anywhere in the system. Worse yet,
all of this was exposed through an API that some subset of customers started to
depend on.

!fig src="/assets/mongodb/webhooks.svg" caption="Visualization of Webhook storage."

Because API access was unrestricted, some consumers became dependent on having
unfettered access to all of those historic Webhooks, and disallowing it at this
point would break practical backwards compatibility, even if no such
compatibility was promised.

Sharding allowed the system to scale in the beginning with relative ease, but
consequently left a product catastrophe. Months worth of engineering time will
be eaten up figuring out how to make 10s of TBs of nearly worthless data
accessible without affecting production systems.

What _should_ have happened from the beginning is that even if it was very
important to have a Webhooks paper trail going back to the beginning of time,
old events should have been moved offline to an archive like S3 so that they
could be audited at some later time. More practically, it's probably not even
worth going that far, and events could conceivably just be purged completely
after a reasonable 30 or 90 day timeframe, leaving a data set small enough to
run on a single node forever for everyone except a Google-sized system.

### High Availability (#high-availability)

MongoDB implements a form of easy high availability (HA) by allowing automatic
failover when the current primary becomes unavailable by electing a secondary
to primary in its place.

But while other MongoDB has been catching up to other databases with regards to
durability, other databases have been catching up to MongoDB in HA. The
secondary election schema described isn't substantially different than the HA
schemes implemented in Postgres by services like [AWS][aws-ha] or
[Heroku][heroku-ha], which essentially do the exactly the same thing as Mongo
by looking for an unhealthy primary and if found, promoting a follower in its
place.

More importantly though, HA is often not as much about the technical
sophistication of a sytem as it is about the technical processes surrounding
it. Everyone imagines that the major dangers to the availability of a data
store are disk failures and network partitions, and once in a while those
things really do happen. However, in most day-to-day development, a database
has other existential problems that are vastly more likely to occur:

* Operator error in which somebody accidentally mangles some piece of critical
  data and brings an online system to its knees.
* Overly expensive migrations locking schema/data or eating all available
  resources.
* Long-lived transactions or other jobs that starve other online operations of
  resources.
* Bungled attempts at upgrading a database.
* Poorly vetted deploys in which new code expects a certain schema or data
  before it's actually updated, and which cause a failure once they go out.

In practice, an HA data store helps you, but not as much as you'd think. I've
seen as much or more downtime on a large MongoDB system as I have on a Postgres
system of similar scale; none of it had anything to do with problems at the
network layer.

## Summary (#tldr)

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
Google, then sure, just use Bigtable or Cassandra. You can afford it.

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
[transaction-isolation]: https://www.postgresql.org/docs/current/static/transaction-iso.html
[two-phase]: https://docs.mongodb.com/manual/tutorial/perform-two-phase-commits/
[unique-index]: https://docs.mongodb.com/manual/core/index-unique/
[write-concerns]: https://docs.mongodb.com/manual/reference/write-concern/
