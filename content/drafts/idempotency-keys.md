---
title: Implementing a Stripe-like Idempotency Key System On Postgres
published_at: 2017-09-20T11:46:10Z
hook: Hardening services by identifying foreign state
  mutations and grouping other changes into atomic phases.
---

Rocket rides!

1. Insert an idempotency key record.
2. Create a ride record to track the ride that's about to
   happen.
3. Create an audit record referencing the created ride.
4. Make an API call to Stripe to charge the ride.
5. Update the ride record with the created charge ID.
6. Create an audit record referencing the successful
   charge.
7. Send an email to the user containing the charge details.

Let's look at some of the things that can go wrong:

## Foreign state mutations (#foreign-state)

## Atomic phases (#atomic-phases)

## Murphy's law (#murphys-law)

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
