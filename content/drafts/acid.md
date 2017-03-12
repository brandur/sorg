---
title: Building Robustly With ACID, or Why to Avoid MongoDB
published_at: 2017-03-12T17:59:02Z
hook: TODO
location: San Francisco
---

In the last decade we've seen the emergence of a number of
new data stores that trade ACID away for other flashy
features like higher availability, streaming changesets, or
JavaScript APIs. Although these features might be desirable
in a few situations (HA in particular), in the vast
majority of use cases projects should prefer the use of a
database that offers ACID guarantees to help ensure the
scalability of their software.

ACID databases are by far the most important tool in
existence for ensuring data correctness in an online
system.

## Optimizing for the Wrong Thing (#optimizing)

An often cited features document data stores is that they
allow you to bootstrap quickly and get to a prototype
because they don't bog you down with schema design.

Keeping in mind that this claim isn't actually true -- a
developer reasonably competent with their RDMS of choice
and armed with an ORM and migration framework can easily
keep up with their document store-oriented counterpart, but
more importantly, this is optimizing for exactly the wrong
thing.

While building a prototype quickly might be important for
the first two weeks of a system's lifespan, the next ten
years will be about keeping it running correctly by
minimizing bugs and data consistency problems that will
lead directly to user pain and attrition. Valuing miniscule
short-term gains over long-term maintainability is an
incredibly pathological way of doing anything, let alone
building software.

But how does an RDMS help with maintainability? Well, it
turns out that ACID guarantees combined with strong
constraints are very valuable tools. Lets take a closer
look.

## Atomicity (#atomicity)

Atomicity is the "A" in ACID. It dictates that within a
given database transaction, the changes to be committed
will be all or nothing. If the transaction fails partway
through, the initial database state is left unchanged.

It's a favorite claim of products like MongoDB and
RethinkDB to say that transactions in their systems are
"atomic" -- as long as you only need atomicity inside a
single document update. This is marketing-speak for "the
system isn't atomic at all" -- if your data store doesn't
guarantee atomicity at the document level, you're in for
real trouble indeed.

Within the context of building web applications, atomicity
in the sense of ACID is incredibly valuable. Software is
buggy by nature and introducing problems that
unintentionally fail requests is inevitable. By wrapping
requests in transactions, we can ensure that even in the
these worst case scenarios, state is left undamaged, and
it's safe for other requests to proceed in the system.

It's never desirable to fail requests that we expected to
commit, but atomicity cancels the expensive fallout.

### In a World Without (#without-atomicity)

So what happens in a world without ACID guarantees where
any failed request leaves invalid state behind?

!fig src="/assets/mongodb/request-failure.svg" caption="Demonstration of how without an atomicity guarantee, a failed request results in an invalid state of data."

You might hope that you'd have specialized safeguards to
protect against failed state and even try to roll it back
where possible. While this probably does exist somewhere,
it's much more likely that nothing likely that operators on
the inside just optimistically hope that it won't happen.
I've seen places where the standard operating procedure
after a set of requests failed due to a bug is to go in and
manually correct thousands of requests worth of data.

## Consistency (#consistency)

## Isolation (#isolation)

!fig src="/assets/mongodb/pessimistic-locking.svg" caption="Demonstration of pessimistic locking showing 3 requests to the same resource. Each blocks the next in line."

## Constraints (#constraints)

Fast prototyping.

## Scaling (#scaling)
