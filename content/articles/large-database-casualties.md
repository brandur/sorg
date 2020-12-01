+++
hook = "Databases shed important RDMS features as they get big. Examining why this tends to be the case, and some ideas for preventing it."
location = "San Francisco"
published_at = 2020-12-01T20:06:51Z
title = "Casualties of Large Databases"
+++

Big data has an unfortunate tendency to get messy. A few years in, a growing database that use to be small, lean, and well-designed, has better odds than not of becoming something large, bloated, and with best practices tossed aside and now considered unsalvageable.

There's a few common reasons that this happens, some better than others:

* **Technological limitation:** The underlying tech doesn't support the scale. Say transactions or referential integrity across partitions in a sharded system.
* **Stability:** Certain operations come to be considered too risky. e.g. Batch update operations that have unpredictable performance properties.
* **Cost/effort:** Doing things the right way is too hard or too expensive. e.g. Back-migrating a large amount of existing data.
* **Convenience:** Similar to the previous point, poor data practice is simply by far the easiest thing to do, and gets your immediate project shipped more quickly, even if it makes future projects more difficult.

The loss of these features is unfortunate because they're the major reason we're using sophisticated databases in the first place. In the [most extreme cases](https://eng.uber.com/schemaless-part-one-mysql-datastore/), advanced databases end up as nothing more than glorified key/value stores, and the applications they power lose important foundations for reliability and correctness.

## The casualties of large applications/data (#casualties)

### Transactions (#transactions)

ACID transactions tend to be one of the first things to go, especially since the value they provide isn't immediately obvious in a new system that's not yet seeing a lot of traffic or trouble. Between that and the facts that they add some friction in writing code quickly, and can lead to locking problems in production mean that they're often put in the chopping block early, especially when less experienced engineers are involved.

Losing transactions is bad news for an applications future operability, but as this subject's already covered extensively elsewhere ([including by me](/acid)), I won't go into depth here.

### Referential integrity (#referential-integrity)

Referential integrity guarantees that if a key exists somewhere in a database, then the object its referencing does as well. Foreign keys allow developers to control deletions such that if an object is being removed, but is still referenced, than that deletion should be blocked (`ON DELETE RESTRICT`), or, that referencing objects should be removed with it (`ON DELETE CASCADE`).

It's a powerful tool for correctness -- having the database enforcing certain rules makes code easier to write. _Not_ having it tends to bleed out into code. Suddenly anytime a referenced object is loaded _anywhere_, the case that it came up without a result must be handled:

``` ruby
user = User.load(api_key.user_id)
if !user
  raise ObjectNotFound, "couldn't find user!"
end
```

Sacrificing referential integrity is rationalized away in a number of ways. Sometimes it's due to technological limitation, sometimes due to reliability concerns (a benign-looking delete triggering an unexpectedly large cascade), but more often it's for the simple-and-not-good reason that maintaining good hygiene around foreign key relations takes discipline and effort.

### Nullable, as far as the eye can see (#nullable)

Relations in large databases tend to have a disproportionate number of nullable fields. This is a problem because in application code it's more difficult to work with objects that have a poorly defined schema. Every nullable field needs to be examined independently, and a fallback designed for it it didn't have a value. This takes time and introduces new vectors for bugs.

There's a number of reasons that nullable-by-default is so common. The simplest is simply that nullable columns are literally the default in DDL -- you'll get one unless you're really thinking about what you're doing and explicitly use `NOT NULL`.

A more common reason is that non-nullable columns often require that existing data be migrated, which is difficult, time consuming, and maybe even operationally fraught on nodes which are running very hot and which a migration unexpectedly pushes over the edge.

Lastly, there are often technological limitations as well. In Postgres for example, even after running a migration, taking that last step of changing a nullable column to non-nullable (`SET NOT NULL`) isn't safe. Postgres needs to verify that there are no nulls in the table, requiring a full table scan that blocks other operations. On a small table that'll run in an instant. On a large one, it could be the downfall of production.

### Suboptimal indexing (#indexing)

Indexes are the easiest thing in the world to work with until they're not. In a large system, they might get complicated because:

* They need to be built on multiple clusters instead of just one.
* Building them on very hot nodes gets risky as the build interferes with production operations. Internal teams may need to build tools to throttle or pause builds.
* Data gets so large that building them takes a long time.
* Data gets so large that each index is a significant non-trivial cost to store.

Reduced performance is the most obvious outcome, but expensive index operations can have less obvious ones too. I worked on a project recently where product design was being driven by whether options would necessitate raising a new index on a particularly enormous collection which would take weeks and cost a large figure every year in storage costs alone.

### Restricted APIs (#restricted-apis)

SQL is the most expressive language ever for querying and manipulating data, and in the right hands, that power can make hard things easy.

However, the more complex the SQL statement, the more likely it is to impact production through problems like unpredictable performance or unanticipated locking. A common solution is for storage teams to simply ban non-trivial SQL wholesale, and constrain developers to a vastly simplified API -- e.g. single row select, multi row select with index hint, single row update, single row delete.

``` ruby
# a simplified storage API
def insert(data:); end
def delete_one(id:); end
def load_many(predicate:, index:, limit:); end
def load_one(id:); end
def update_one(id:, data:); end
```

At a previous job, our MySQL DBA banned any database update that affected more than one row, even where it would be vastly beneficial to performance, due to concerns around them delaying replication to secondaries. This might have helped production, but had the predictable effect of reduced productivity along with some truly heinous workarounds for things that should have been trivial, and which instead resulted in considerable tech debt.

Where I work now, even with the comparative unexpressiveness of Mongo compared to SQL, every select in the system must be named and statically defined along with an index it expects to use. This is so that we can verify at build time that the appropriate index is already available in production.

## Ideas for scalability (#scalability-ideas)

There's a divide between the engineers who run big production systems and the developers who work on open-source projects in the data space, with neither group having all that much visibility into the other. Engineers who run big databases tend to adopt a nilhist outlook that every large installation inevitably trends towards a key/value store -- at a certain point, the niceties available to smaller databases must get the axe. Open-source developers don't tend to value highly the features that would help big installations.

I don't think the nilhist viewpoint should be the inevitable outcome, and there's cause for optimism in the development of systems like Citus, Spanner, and CockroachDB, which enable previously difficult features like cross shard transactions. We need even more movement in that direction.

There's a variety of possible operations-friendly features that might be possible to counteract the entropic dumbing down of large databases. Some ideas:

* Make index builds pauseable so that they can be easily throttled in emergencies.
* Make it easy to make a nullable field non-nullable, *not* requiring a problematic and immediate full table scan.
* A "strict" SQL dialect that makes specifying fields as `NOT NULL` default, and specifying foreign keys required.
* A communication protocol that allows the query to signal out-of-band with a query's results that it didn't run particularly efficiently, say that it got results but wasn't able to make use of an index. This would allow a test suite to fail early by signaling the problem to a developer instead of finding out about it in production.
* A migrations framework built into the database itself that makes migrations easier and faster to write while also guaranteeing stability by allowing long-lived migration-related queries to be deprioritized and paused if necessary.

Ideally, we get to a place where large databases enjoy all the same benefits as smaller ones, and we all get to reap the benefits of software that gets more stable and more reliable as a result.
