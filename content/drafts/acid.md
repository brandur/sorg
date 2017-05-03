---
title: Building Robustly With ACID, or Why to Avoid MongoDB
published_at: 2017-03-12T17:59:02Z
location: San Francisco
hook: TODO
hook_image: true
---

In the last decade we've seen the emergence of a number of
new data stores that trade ACID away for other flashy
features like higher availability, streaming changesets,
JavaScript APIs, or nestable JSON documents. Although these
features might be desirable in a few situations (HA in
particular), in the vast majority of use cases projects
should prefer the use of a database that offers ACID
guarantees to help ensure the scalability of their
software.

ACID databases are by far the most important tool in
existence for ensuring data correctness in an online
system.

## Optimizing for the wrong thing (#optimizing)

An often cited features document data stores is that they
allow you to bootstrap quickly and get to a prototype
because they don't bog you down with schema design.

Keeping in mind that this claim isn't actually true -- a
developer reasonably competent with their RDMS of choice
and armed with an ORM and migration framework can easily
keep up with their document store-oriented counterpart (and
probably outpace them), but more importantly, it's
optimizing for exactly the wrong thing.

While building a prototype quickly might be important for
the first two weeks of a system's lifespan, the next ten
years will be about keeping it running correctly by
minimizing bugs and data consistency problems that will
lead to user and operator pain and attrition. Valuing
miniscule short-term gains over long-term maintainability
is a pathological way of doing anything; it's a sin when
building critical production software.

But how does an RDMS help with maintainability? Well, it
turns out that ACID guarantees combined with strong
constraints are valuable tools. Lets take a closer look.

## Atomicity (#atomicity)

Atomicity is the "A" in ACID. It dictates that within a
given database transaction, the changes to be committed
will be all or nothing. If the transaction fails partway
through, the initial database state is left unchanged.

It's a favorite claim of products like MongoDB and
RethinkDB to say that transactions in their systems are
"atomic" -- as long as you only need atomicity inside a
single document update. This is marketing-speak for "the
system isn't atomic at all". If your data store doesn't
guarantee atomicity at the document level, you're in for
a real ride.

Within the context of building web applications, atomicity
is incredibly valuable. Software is buggy by nature and
introducing problems that unintentionally fail requests is
inevitable. By wrapping requests in transactions, we get to
ensure that even in the these worst case scenarios, state
is left undamaged, and it's safe for other requests to
proceed in the system.

It's never desirable to fail requests that we expected to
commit, but atomicity cancels the expensive fallout.

### The janitorial team (#without-atomicity)

So what happens in a world without ACID guarantees where
any failed request leaves invalid state behind?

The default will be that a subsequent retry won't be able
to reconcile the broken state, and that the data will need
to be repaired before it's usable again.

!fig src="/assets/mongodb/request-failure.svg" caption="Demonstration of how without an atomicity guarantee, a failed request results in an invalid state of data."

You might hope that companies in this position would have
automated protections in place to try and roll back bad
state where possible. While this may exist somewhere, it's
much more likely that the overarching strategy is an
optimistic sense of hope that these kinds of problems won't
happen very often. This is often combined with a
laissez-faire philosophy that all systems have some bad
data in them, and there's no point in agonizing too much
over a few bad tuples.

Particularly bad incidents will necessitate manual operator
intervention, or even a specially crafted "fixer script" to
clean up state and get everything back to normal. After a
certain size, this sort of thing will be happening
frequently, and your engineers will start to spend more and
more of their time acting as janitors.

## Consistency (#consistency)

Consistency is the "C" in ACID. It dictates that every
transaction will bring a database from one valid state to
another valid state; there's no potential for anything in
between.

It might be a little hard to imagine what this can do for a
real world app in practice, but consider one the very
common case where a user signs up for a service with their
email address `foo@example.com`. We don't want to two
accounts with the same email in the system, so when
creating the account we'd use an initial check:

1. Look for any existing `foo@example.com` users in the
   system. If there is one, reject the request.

2. Create a new record for `foo@example.com`.

Regardless of data store, this will generally work just
fine until you have a system with enough traffic to start
revealing edge cases. If we have two requests trying to
register `foo@example.com` that are running almost
concurrently, then the above check can fail us because both
could have validated step one successfully before moving on
to create a duplicated record.

!fig src="/assets/acid/consistency.jpg" caption="Consistency."

You can solve this problem on an ACID database in multiple
ways:

1. You could use a strong isolation level like
   `SERIALIZABLE` on your transactions, which would
   guarantee that only one `foo@example.com` creation would
   be allowed to commit.

2. You could put a uniqueness check on the table itself (or
   on an index) which would prevent a duplicate record from
   being inserted.

### Fix it later. Maybe. (#without-consistency)

Without ACID, its up to your application code to solve the
problem. You could implement some a locking system of sorts
to guarantee that only one registration for any given email
address can be in flight at once. Realistically, many
providers on non-ACID databases will probably elect to just
not solve the problem. Maybe later, _after_ it causes
painful fall out in production.

## Isolation (#isolation)

Isolation is the "I" in ACID. It ensures that two
simultaneously executing transactions that are operating on
the same information don't conflict with each other. Each
one has access to a pristine view of the data (depending on
isolation level) even if the other has modified it, and
results are reconciled when the transactions are ready to
commit. Modern RDMSes have sophisticated multiversion
concurrency control systems that make this possible in ways
that are correct and efficient.

Concurrent resource access is a problem that every real
world web application is going to have to deal with. So
without isolation, how do you deal with the problem?

### Just lock the shit out of everything (#without-isolation)

The most common technique is to implement your own
pessimistic locking system that constrains access to some
set of resources to a single operation, and forces others to
block until it's finished. So for example, if our core
model is a set of user accounts that own other resources,
we'd lock the whole account when a modification request
comes in, and only unlock it again after we've committed
our work.

!fig src="/assets/mongodb/pessimistic-locking.svg" caption="Demonstration of pessimistic locking showing 3 requests to the same resource. Each blocks the next in line."

This approach is all downsides:

1. ***It's slow.*** Operations waiting on a lock may have
   to wait for very extended periods for resources to
   become available. The more concurrent access, the worse
   it is (which probably means that your large users will
   suffer the most).

2. ***It's inefficient.*** Not every blocking operation
   actually needs to wait on every other operation. Because
   the models you lock on tend to broad to reduce the
   system's complexity, many operations will block when
   they didn't necessarily have to.

3. ***It's a lot of work.*** A basic locking system isn't
   too hard to implement, but if you want to improve its
   speed or efficiency then you quickly need to move to
   something more elaborate which gets complicated fast.
   With an ACID database, you'll get a very fast, very
   efficient, and very correct locking system built-in for
   free.

3. ***It's probably not right.*** Locks and software are
   hard. Implementing your own system _is_ going to yield
   problems; it's just a question of what magnitude.

## Constraints are good (#constraints)

I talked before about how schemaless databases are often
misinterpreted as a feature because they enable fast
prototyping. Rich Hickey has a great talk where he makes [a
distinction between "simple" and "easy"][simple-made-easy],
with ***simplicity*** being an elegant process of boxing
powerful concepts into useful layers of abstraction,
whereas ***ease*** is nearly the opposite, where short term
gratification is favored to the detriment of long term
prosperity. Schemaless databases are not simple; they're
easy.

Data management in a service built on schemaless data store
will eventually become so painful that even its most
steadfast proponents will acquiesce to allow some form of
constraints. Life is artificially difficult when your
`User` records aren't even guaranteed to come with an `id`
or `email` field.

By the time an organization hits hundreds of models and
thousands of fields, they'll certainly be using some kind
of object modeling framework in a desperate attempt to get
a few assurances around data shape into place. By that
point though, things are probably already inconsistent
enough that it'll make migrations difficult in perpetuity,
and application code twisted and complicated as its built
to gracefully handle dozens of edge cases.

Throw away prototypes are the _only_ place that schemaless
data stores should be put to use (and again, even there I'd
question whether it's actually faster or has any measurable
merit). For services that you want to run in production,
the better defined your schema and the more self-consistent
your data, the easier your life is going to be.

## On scaling (#scaling)

A common criticism of ACID databases is that they don't
scale, and by extension horizontally scalable (and usually
non-ACID) data stores are the only valid choice.

First of all, despite unbounded optimism for growth, the
vast majority will be well-served by a single vertically
scalable node; probably forever. By offloading infrequently
needed "junk" data to scalable alternate data stores, it's
fairly reasonable to expect to vertically scale a service
for a very long time, even if it has somewhere on the order
of millions of users. Show me any databases that's on the
scale of TBs or larger, and I'll show you the 100s of GBs
that are in there when they don't need to be.

There are a few use cases that legitimately need
scalability, and for those you should choose a database
that gives you as many of these guarantees as possible,
even if it's on a per-partition scale.

## Build on solid ground (#solid-substrates)

There's a common theme to everything listed above:

* You can get away ***without atomicity***, but you end up
  hacking around it with cleanup scripts and lots of
  expensive engineer-hours.

* You can get away ***without consistency***, but only
  through the use of elaborate application-level schemes.

* You can get away ***without isolation***, but only by
  building your own probably slow, probably inefficient,
  and probably buggy locking scheme.

* You can get away ***without constraints*** and schemas,
  but only by internalizing a nihilistic understanding that
  your production data isn't consistent.

By choosing a non-ACID data store, you end up
reimplementing everything that it does for you in the user
space of your application, except far worse.

Your database can and should act as a foundational
substrate that offers your application profound leverage
for fast and correct online operation. Not only does it
provide these excellent features, but it provides them in a
way that's been battle-tested and empirically vetted by
millions of hours of running some of the heaviest
applications in the world.

[simple-made-easy]: https://www.infoq.com/presentations/Simple-Made-Easy
