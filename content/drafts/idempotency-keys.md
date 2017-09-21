---
title: Implementing Stripe-like Idempotency Keys in Postgres
published_at: 2017-09-20T11:46:10Z
location: Calgary
hook: Hardening services by identifying foreign state
  mutations and grouping local changes into atomic phases.
---

We talked about idempotency

What is idempotency?

## Rocket Rides (#rocket-rides)

We have a great app

<p><img src="/assets/idempotency-keys/rocketrides-ios-ride.png" data-rjs="2"></p>

The [Rocket Rides repository][rocketrides] comes with a
simple server implementation, but software tends to grow
with time, so to be more representative of what a real
service would look like we're going to complicate things a
little bit by adding a few embellishments.

Most importantly, we're going to build our own idempotency
key system so that if an API call from the mobile app to
the backend fails, we can safely retry the operation. We'll
be charging a user's credit card, and we absolutely can't
take the risk of charging them twice.

More specifically, when a new rides comes in we'll perform
this set of operations:

1. Insert an idempotency key record.
2. Create a ride record to track the ride that's about to
   happen.
3. Create an audit record referencing the ride.
4. Make an API call to Stripe to charge the user for the
   ride.
5. Update the ride record with the created charge ID.
6. Send the user a receipt via email.
7. Update idempotency key with results.

!fig src="/assets/idempotency-keys/api-request.svg" caption="A typical API request to our embellished Rocket Rides backend."

### The entropy of production (#failure)

Most of the time we can expect every our Rocket Rides API
calls to go swimmingly, and every operation will succeed
without a problem. However, when we reach the scale of
thousands of API calls a day, we'll start to notice a few
problems appearing here and there; requests failing due to
poor cellular connectivity, API calls to Stripe failing
occasionally, or bad turbulence caused by rocketing through
the air at supersonic speeds knocking connections totally
offline. After we reach the scale of millions of API calls
a day, we'll be seeing these sorts of things happening all
the time.

Let's look at a few examples of things that can go wrong:

* Inserting the idempotency key or ride record could fail
  due to a constraint violation or a database connectivity
  problem.
* Our call to Stripe could timeout, leaving it unclear
  whether our charge when through.
* Contacting Mailgun to send the receipt could fail,
  leaving the user with a credit card charge but no formal
  notification of the transaction.
* The client could disconnect, cancelling the operation
  midway through.

## Foreign state mutations (#foreign-state)

To shore up our backend, it's key to identify where we're
making ***foreign state mutations***; that is, calling out
and manipulating data on another system. This might be
creating a charge on Stripe, adding a DNS record, or
sending an email.

Some foreign state mutations are idempotent by nature (e.g.
adding a DNS record), some are not idempotent but can be
made idempotent with the help of an idempotency key (e.g.
charge on Stripe, sending an email), and some operations
are not idempotent, most often because a foreign service
hasn't designed them that way and doesn't provide a
mechanism like an idempotency key.

The reason that the local vs. foreign distinction matters
is that unlike a local set of operations where can just
roll back a result that we didn't like, once we make our
first foreign state mutation, we're committed one way or
another [1]. We've pushed data into a system beyond our own
boundaries and we shouldn't lose track of it.

We're using an API call to Stripe as a common example, but
remember that even foreign calls within your own
infrastructure count! It's tempting to treat emitting a
record to Kafka as part of an atomic operation because it
has such a high success rate that it feels like it is. It's
absolutely not, and should be treated like any other
foreign state mutation.

## Atomic phases (#atomic-phases)

An ***atomic phase*** is a set of local state mutations
that occur in transactions _between_ foreign state
mutations. We say that they're atomic because we can use an
RDMS like Postgres to guarantee that either all of them
will occur, or none will.

We should endeavor to commit our atomic phases before
initiating a foreign state mutation so that if it fails,
our local state will still contain a record of it happening
that we can use to retry the operation.

### Savepoints (#savepoints)

At the last action of an atomic phase, the savepoint record
should be updated.

## Background jobs and job staging (#background-jobs)

Foreign state mutations make a request slower and more
difficult to reason about, so they should be avoided when
possible. In many cases it's possible to defer this type of
work to after the request is complete by sending it to a
background job queue.

To re-use the examples above, a charge to Stripe probably
_can't_ be deferred to the background because we want to
know whether it succeeded in-band (and deny the request if
not). Sending an email _can_ and should be sent to the
background.

By using a [_transactionally-staged job
drain_](/job-drain), we can hide jobs from workers until
we've confirmed that they're ready to be worked by
isolating them in a transaction. This also means that the
background work becomes part of an atomic phase and greatly
simplifies the operational. Work should always be offloaded
to background queues when possible.

## Building Rocket Rides for ultra-resilience (#rocket-rides-resilience)

### The idempotency key relation (#idempotency-key)

`locked`

### Other schema (#other-schema)

### Designing atomic phases (#rocket-rides-phases)

!fig src="/assets/idempotency-keys/atomic-phases.svg" caption="API request to Rocket Rides broken into foreign state mutations and atomic phases."

## Murphy in action (#murphys-law)

Now let's look at a perfectly degenerate case.

## Non-idempotent foreign state mutations (#non-idempotent)

If we know that a foreign state mutation is an idempotent
operation or it supports an idempotency key (like Stripe
does), we know that it's safe to retry any failures that we
see.

Unfortunately, not every service will make this guarantee.
If we try to make a non-idempotent foreign state mutation
and we see a failure, we may have to persist this operation
as permanently errored. In many cases we won't know whether
it's safe to retry or not, and we'll have to take the
conservative route and fail the operation.

The exception is if we got an error back from the
non-idempotent API, but one that tell us explicitly that
it's okay to retry. Indeterminate errors like a connection
reset or timeout will have to be marked as failed.

This is why you should implement idempotency and/or
idempotency keys on all your services!

[1] There is one caveat that it may be possible to
implement [two-phase commit][2pc] between a system and all
other systems where it performs foreign state mutations.
This would allow distributed rollbacks, but is complex and
time-consuming enough to implement that it's rarely seen
with any kind of ubiquity in real software environments.

[2pc]: https://en.wikipedia.org/wiki/Two-phase_commit_protocol
[rocketrides]: https://github.com/stripe/stripe-connect-rocketrides
