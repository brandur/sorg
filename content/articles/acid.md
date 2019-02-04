---
title: Building Robust Systems With ACID and Constraints
published_at: 2017-05-16T14:03:01Z
location: San Francisco
hook: On ensuring system integrity, operability, and
  correctness through a solid foundational database, and
  how ACID transactions and strong constraints work in your
  favor. Why to prefer Postgres over MongoDB.
tags: ["postgres"]
---

In 1983, Andreas Reuter and Theo Härder coined the acronym
ACID as shorthand for _atomicity_, _consistency_,
_isolation_, and _durability_. They were building on
earlier work by Jim Gray who'd proposed atomicity,
consistency, and durability, but had initially left out the
_I_. ACID is one of those inventions of the 80s that's not
only just still in use in the form of major database
systems like Postgres, Oracle, and MSSQL, but which has
never been displaced by a better idea.

In the last decade we've seen the emergence of a number of
new flavors of data store that come with untraditional
features like streaming changesets, JavaScript APIs, or
nestable JSON documents. Most of them assume that the need
for horizontal partitioning is a given in this day and age
and therefore ACID is put to the altar (this doesn't
necessarily have to be the case, see below). Every decision
comes with trade offs, but trading away these powerful
guarantees for the novelties du jour or an unexamined
assumption that horizontal scaling will very soon be a
critically required feature is as raw of a deal as you'll
ever see in the technical world.

But why the mismatch in values? I think it's because many
of us have taught ourselves programming on frameworks like
Rails, or been educated in environments where ACID
databases were a part of the package, and we've taken them
for granted. They've always been there, and we've never
necessarily considered closely exactly what they can do for
us and why they're important. In many cases this also leads
to their most powerful features being underutilized.

I want to convince you that ACID databases are one of the
most important tools in existence for ensuring
maintainability and data correctness in big production
systems. Lets start by digging into each of their namesake
guarantees.

## Atomicity (#atomicity)

The "A" in ACID. It dictates that within a given database
transaction, the changes to be committed will be all or
nothing. If the transaction fails partway through, the
initial database state is left unchanged.

Software is buggy by nature and introducing problems that
unintentionally fail some operations is inevitable. Any
sufficiently large program is eventually going to want to
have an operation that writes two or more objects
consecutively, and by wrapping that operation in a
transaction, we get to ensure that even in the these worst
case scenarios state is left undamaged. Every subsequent
operation will start with safe initial state.

It's never desirable to fail transactions that we hoped to
commit, but atomicity cancels the expensive fallout.

!fig src="/assets/acid/transactions-in-requests.svg" caption="Some requests. Each wraps its database operations using an atomic transaction so that they either all commit, or none of them do."

### The janitorial team (#janitorial-team)

Many products will claim "document-level atomicity" (e.g.
MongoDB, RethinkDB, CouchBase, ...) which means that
writing any one row is atomic, but nothing beyond that.
What happens in a world like this where any failed
operation leaves invalid state behind?

The default will be that a subsequent retry won't be able
to reconcile the broken state, and that the data will need
to be repaired before it's usable again.

Here's an example of a simple GitHub-like service. When a
user opens a pull request, we have a number of objects that
we have to save in succession before finishing the request:
a pull request modeling the created resource, a webhook to
fire off to any listeners on the repository, a reviewer
record mapping to whomever we've assigned review, and an
event to store in the audit log.

!fig src="/assets/acid/request-failure.svg" caption="Demonstration of how without an atomicity guarantee, a failed request results in an invalid state of data."

A request that fails after the first two saves fails to
create a valid set of objects, but with transactional
atomicity can't revert the changes it did make. The result?
An invalid pull request. A subsequent request that tries to
look it up might error as the code tries to load state that
was only partially created.

You might hope that projects in this position would have
automated protections in place to try and roll back bad
partial transactions. While this may exist somewhere, it's
much more likely that the overarching strategy is an
optimistic sense of hope that these kinds of problems won't
happen very often. Code paths begin to mutate to load data
defensively so that they handle an exponential number of
combinations of bad state that have accumulated in the data
store over time.

Bad incidents will necessitate heavy manual intervention by
operators, or even a specially crafted "fixer script" to
clean up state and get everything back to normal. After a
certain size, this sort of thing will be happening
frequently, and your engineers will start to spend less
time as engineers, and more time as data janitors.

!fig src="/assets/acid/pillars.jpg" caption="A grid of pillars at the Jewish Museum in Berlin. Real world consistency at its best."

## Consistency (#consistency)

The "C" in ACID. It dictates that every transaction will
bring a database from one valid state to another valid
state; there's no potential for anything in between.

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

!fig src="/assets/acid/consistency.svg" caption="Without guaranteed consistency, there's nothing to stop the database from transitioning to an invalid state."

You can solve this problem on an ACID database in multiple
ways:

1. You could use a strong isolation level like
   `SERIALIZABLE` on your transactions, which would
   guarantee that only one `foo@example.com` creation would
   be allowed to commit.

2. You could put a uniqueness check on the table itself (or
   on an index) which would prevent a duplicate record from
   being inserted.

### Fix it later. Maybe. (#later-maybe)

Without ACID, its up to your application code to solve the
problem. You could implement some a locking system of sorts
to guarantee that only one registration for any given email
address can be in flight at once. Realistically, many
providers on non-ACID databases will probably elect to just
not solve the problem. Maybe later, _after_ it causes
painful fall out in production.

## Isolation (#isolation)

The "I" in ACID. It ensures that two simultaneously
executing transactions that are operating on the same
information don't conflict with each other. Each one has
access to a pristine view of the data (depending on
isolation level) even if the other has modified it, and
results are reconciled when the transactions are ready to
commit. Modern RDMSes have sophisticated multiversion
concurrency control systems that make this possible in ways
that are correct and efficient.

<figure>
  <div class="table-container">
    <table class="overflowing">
      <tr>
        <th>Isolation Level</th>
        <th>Dirty Read</th>
        <th>Nonrepeatable Read</th>
        <th>Phantom Read</th>
        <th>Serialization Anomaly</th>
      </tr>
      <tr>
        <td><strong>Read uncommitted</strong></td>
        <td>Allowed</td>
        <td>Possible</td>
        <td>Possible</td>
        <td>Possible</td>
      </tr>
      <tr>
        <td><strong>Read committed</strong></td>
        <td>Not possible</td>
        <td>Possible</td>
        <td>Possible</td>
        <td>Possible</td>
      </tr>
      <tr>
        <td><strong>Repeatable read</strong></td>
        <td>Not possible</td>
        <td>Not possible</td>
        <td>Allowed</td>
        <td>Possible</td>
      </tr>
      <tr>
        <td><strong>Serializable</strong></td>
        <td>Not possible</td>
        <td>Not possible</td>
        <td>Not possible</td>
        <td>Not possible</td>
      </tr>
    </table>
  </div>
  <figcaption>Transaction isolation levels and the
    contention phenomena that they allow. See the <a
    href="https://www.postgresql.org/docs/current/static/transaction-iso.html">Postgres
    docs</a> if you want to learn more.</figcaption>
</figure>

Concurrent resource access is a problem that every real
world web application is going to have to deal with. So
without isolation, how do you deal with the problem?

### Slow, bad, and buggy custom locking schemes (#bad-locking)

The most common technique is to implement your own
pessimistic locking system that constrains access to some
set of resources to a single operation, and forces others to
block until it's finished. So for example, if our core
model is a set of user accounts that own other resources,
we'd lock the whole account when a modification request
comes in, and only unlock it again after we've finished our
work.

!fig src="/assets/acid/pessimistic-locking.svg" caption="Demonstration of pessimistic locking showing 3 requests to the same resource. Each blocks the next in line."

This approach is _all_ downsides:

1. ***It's slow.*** Operations waiting on a lock may have
   to wait for very extended periods for resources to
   become available. The more concurrent access, the worse
   it is (which probably means that your large users will
   suffer the most).

2. ***It's inefficient.*** Not every blocking operation
   actually needs to wait on every other operation. Because
   the models you lock on tend to be broad to reduce the
   system's complexity, many operations will block when
   they didn't necessarily have to.

3. ***It's a lot of work.*** A basic locking system isn't
   too hard to implement, but if you want to improve its
   speed or efficiency then you quickly need to move to
   something more elaborate which gets complicated fast.
   With an ACID database, you'll get a fast, efficient, and
   correct locking system built-in for free.

3. ***It's probably not right.*** Locks and software are
   hard. Implementing your own system _is_ going to yield
   problems; it's just a question of what magnitude.

## Durability (#durability)

The "D" in ACID. It dictates that committed transactions
_stay_ committed, even in the event of a crash or power
loss. It's so important that even data stores that don't
support the rest of ACI* tend to get it right. I wrote a
separate article about [MongoDB's lengthy road to achieving
durability][mongo-durability] for example.

## Optimizing for saved seconds on a decade scale (#optimizing)

An often cited features document data stores is that they
allow you to bootstrap quickly and get to a prototype
because they don't bog you down with schema design. Rich
Hickey has a great talk where he makes [a distinction
between "simple" and "easy"][simple-made-easy], with
***simplicity*** meaning the opposite of complex, and
***ease*** meaning "to be at hand" or "to be approachable"
in that it may provide short term gratification, even if
it's to the detriment of long term maintainability.
Schemaless databases are not simple; they're easy.

First of all, the claim around faster prototyping isn't
actually true -- an experienced developer reasonably
competent with their RDMS of choice and armed with an ORM
and migration framework can keep up with their document
store-oriented counterpart (and almost certainly outpace
them), but even if it were true, it's optimizing for
exactly the wrong thing.

While building a prototype quickly might be important for
the first two weeks of a system's lifespan, the next ten
years will be about keeping it running correctly by
minimizing bugs and data consistency problems that will
lead to user and operator pain and attrition. Life is
artificially difficult when your `User` records aren't even
guaranteed to come with an `id` or `email` field, and even
steadfast schemaless enthusiasts will acquiesce to allow
some form of constraints.

By the time an organization hits hundreds of models and
thousands of fields, they'll certainly be using some kind
of object modeling framework in a desperate attempt to get
a few assurances around data shape into place.
Unfortunately by that point, things are probably already
inconsistent enough that it'll make migrations difficult in
perpetuity, and application code twisted and complicated as
it's built to safely handle hundreds of accumulated edge
cases.

For services that run in production, the better defined the
schema and the more self-consistent the data, the easier
life is going to be. Valuing miniscule short-term gains
over long-term sustainability is a pathological way of
doing anything; when building production-grade software,
it's a sin.

## Not "webscale" (#scaling)

A common criticism of ACID databases is that they don't
scale, and by extension horizontally scalable (and usually
non-ACID) data stores are the only valid choice.

First of all, despite unbounded optimism for growth, the
vast majority will be well-served by a single vertically
scalable node; probably forever. By offloading infrequently
needed "junk" to scalable alternate stores and archiving
old data, it's reasonable to expect to vertically scale a
service for a very long time, even if it has somewhere on
the order of millions of users. Show me any databases
that's on the scale of TBs or larger, and I'll show you the
100s of GBs that are in there when they don't need to be.

After reaching scale on the order of Google's, there's an
argument to be made for giving up aspects of ACID in return
for certain kinds of partitioning tolerance and guaranteed
availability, but advances in newer database technologies
that support some ACID along with scaling mean that you
don't have to go straight to building on top of a glorified
key/value store anymore. For example, Citus gives you
per-shard ACID guarantees. Google Spanner provides
distributed locking read-write transactions for when you
need them.

!fig src="/assets/acid/foundation.jpg" caption="For best results, build your app on solid foundations."

## Check your foundation (#foundation)

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
  your production data isn't cohesive.

By choosing a non-ACID data store, you end up
reimplementing everything that it does for you in the user
space of your application, except worse.

Your database can and should act as a foundational
substrate that offers your application profound leverage
for fast and correct operation. Not only does it provide
these excellent features, but it provides them in a way
that's been battle-tested and empirically vetted by
millions of hours of running some of the heaviest
workloads in the world.

My usual advice along these lines is that there's no reason
not to start your projects with an RDMS providing ACID and
good features around constraints. In almost every case the
right answer is probably to just use Postgres.

[mongo-durability]: /fragments/mongo-durability
[simple-made-easy]: https://www.infoq.com/presentations/Simple-Made-Easy
