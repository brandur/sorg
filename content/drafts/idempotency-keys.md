---
title: Idempotency & Idempotency Keys
published_at: 2017-01-12T03:11:50Z
hook: Detect problems and inconsistencies in production without affecting
  users.
location: San Francisco
---

As many in people in the technology industry can tell you,
networks are unstable features that are prone to a variety
of problems including outages, routing problems, and other
intermittent failures.

Consider a call between any two nodes. There are a variety
of interesting failures that can occur:

* The initial connection could fail as the client tries to
  connect to its remote.

* The call could fail midway through the remote's
  operation, leaving the work in limbo.

* The call could succeed, but the connection break before
  the remote can tell the client about it.

This often leaves the client in an uncertain situation. It
can sometimes intuit what happened and know whether it's
safe to retry the operation, but from its perspective some
cases are indistinguishable. If a connection breaks midway
through, it has no idea whether the work it was trying to
do completed successfully or not.

This problem is a classic staple of distributed systems
that everyone running one has to deal with. Keep in mind
too that even your integration with Stripe is a distributed
system!

## Idempotency (#idempotency)

The easiest way to solve the problem is to make remote
endpoints idempotent, meaning that if they can be called
any number of times with no harmful side effects. If a
client sees any kind of error, they can converge their
state by retrying the call until it goes through.

As an example, take the API for a basic DNS provider that
allows us to add subdomains via `PUT
/domains/:domain/subdomains/:subdomain/`:

    PUT /domains/example.com/subdomains/s3.example.com
    { "type": "CNAME", "value": "example.s3.amazonaws.com", "ttl": 3600 }

All the information needed to create a record is included
in the call, and it's perfectly safe for a client to invoke
it any number of times.

In a RESTful API, an idempotent call normally uses the
`PUT` verb to signify that we're _replacing_ the target
resource as opposed to simply modifying it.

## Idempotency Keys (#idempotency-keys)

But what if we have an API endpoint that we need to be sure
is invoked exactly once and no more? In this case pure
idempotency might not be suitable. An example of this type
of operation might be to charge a customer money;
accidentally calling such an endpoint twice would lead to
the customer double-charged, which is obviously very bad.

This is where **idempotency keys** come into play. The
basic idea is that when performing a request, a client
generates a special ID just for that operation and sends it
up to a remote along with the normal payload. The remote
receives the ID and correlates it with the state of the
request on its end. If the client notices a failure, it
retries the request with the same ID, and from there it's
up to the remote to figure out what to do with it.

If we consider our sample network failure cases from above:

* On retrying a connection failure, on the second request
  the remote will see the ID for the first time, and
  process it normally.

* On a failure midway through an operation, the remote
  picks up the work and carries it through. The exact
  behavior is heavily dependent on implementation, but if
  the previous operation was successfully rolled back by
  way of an ACID database, it'll be safe to retry it
  wholesale. Otherwise, state is recovered and the call is
  continued.

* On a response failure (i.e. the operation executed
  successfully, but the client couldn't get the result),
  the server simply replies with a cached result of the
  successful operation.

The Stripe API implements idempotency keys on all endpoints
by allowing clients to pass a unique value in with the
special `Idempotency-Key` header, which allows us to
guarantee the safety of distributed operations.

## Wrapping Up (#wrapping-up)

[stripe-keys]: https://stripe.com/docs/api?lang=curl#idempotent_requests
