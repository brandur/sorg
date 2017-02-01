---
title: Idempotency & Idempotency Keys
published_at: 2017-01-12T03:11:50Z
hook: Detect problems and inconsistencies in production without affecting
  users.
location: San Francisco
---

Whether they know it or not, almost anyone you'll meet can
atest to the fact that networks are unreliable; every one
of us has had trouble connecting to Wi-Fi, had a call drop
from our cellular carrier, or lost our ISP for a few hours
as they had trouble on their end or were performing some
scheduled maintenance.

Backend engineers who've dealt with a lot of networks can
bear a special level of witness to this. The networks
connecting our servers together are more reliable on
average compared to our consumer-level last miles like
cellular or home Internet packages, but given enough
information moving across the wire, they're still going to
fail in all kinds of exotic ways, even if those failures
are statistically unusual on the whole. Outages, routing
problems, and other intermittent failures are all bound to
be happenening all the time at some ambient background
rate.

Consider a call between any two nodes. There are a variety
of failure modes that can occur:

* The initial connection could fail as the client tries to
  connect to its remote.

* The call could fail midway through the remote's
  operation, leaving the work in limbo.

* The call could succeed, but the connection break before
  the remote can tell the client about it.

Any one of these leaves the client that made the request in
an uncertain situation. In some cases, the type of failure
is definitive enough that the client can know with good
certainty that it's safe to retry it wholesale. A total
failure to event establish a connection to the remote for
example. In many others though, the operation's success
from the perspective of the client is ambiguous and it
doesn't know whether retrying the operation is safe. A
connection terminating midway through message exchange is
an exampmle of this case.

This problem is a classic staple of distributed systems.
And keep in mind that when I'm talking about a "distributed
system" in this sense, the bar is low: as few as two
computers connection via a network that are passing each
other messages. Technically, even your servers calling out
to the Stripe API over the Internet is a distributed
system!

## Idempotency (#idempotency)

The easiest way to address inconsistencies in distributed
state caused by failures is to implement remote endpoints
so that they're _idempotent_, which means that they can be
called any number of times with no harmful side effects.

When a client sees any kind of error, they can converge
their state by retrying the call, and can keep trying until
it goes through successfully. This resolves the problem of
an ambiguous failure completely because the client knows
that it can safely handle any failure using one simple
technique.

As an example, take the API for a basic DNS provider that
allows us to add subdomains via `PUT
/domains/:domain/subdomains/:subdomain/`:

    PUT /domains/example.com/subdomains/s3.example.com
    { "type": "CNAME", "value": "example.s3.amazonaws.com", "ttl": 3600 }

All the information needed to create a record is included
in the call, and it's perfectly safe for a client to invoke
it any number of times. If the server receives a call that
it realizes is a duplicate because the domain already
exists, it simply ignores the request and responds with a
successful status code.

In a RESTful API, an idempotent call normally uses the
`PUT` verb to signify that we're _replacing_ the target
resource as opposed to simply modifying it (in modern
RESTful parlance, a modification would be represented by a
`PATCH`).

## Idempotency Keys (#idempotency-keys)

But what if we have an API endpoint that we need to be sure
is invoked exactly once and no more? In this case pure
idempotency might not be suitable. An example of this type
of operation might be to charge a customer money;
accidentally calling such an endpoint twice would lead to
the customer double-charged, which is obviously bad.

This is where **idempotency keys** come into play. When
performing a request, a client generates a unique ID to
identify just that operation and sends it up to a remote
along with the normal payload. The remote receives the ID
and correlates it with the state of the request on its end.
If the client notices a failure, it retries the request
with the same ID, and from there it's up to the remote to
figure out what to do with it.

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

The Stripe API implements idempotency keys on mutating
endpoints (i.e. anything under `POST` in our case) by
allowing clients to pass a unique value in with the special
`Idempotency-Key` header, which allows a client to
guarantee the safety of distributed operations.

## Being a Good Distributed Citizen (#citizen)

Safely handling failure is hugely important, but beyond
that it's also recommended that it be handled in a
considerate way. When a client sees that a network
operation has failed, there's a good chance that it's due
to an intermittent failure that'll be gone by the next
retry. However, there's also a chance that it's a more
serious problem that's going to be more tenacious; if the
remote service is in the middle of an incident that's
causing hard downtime for example. Not only will retries of
the operation not go through, but they may contribute to
further degradation of the already troubled remote.

It's usually recommended that clients follow something akin
to an [exponential backoff][exponential-backoff] algorithm
as they see errors. The client blocks for a brief initial
wait time on the first failure, but as the operation
continues to fail, it waits proportionally to 2^N, where
_N_ is the number of failures that have occurred. By
backing off exponentially, we can ensure that clients
aren't hammering on a downed remote and contributing to the
problem.

Furthermore, it's also a good idea to add an element of
randomness to the wait times. If a problem with a remote
causes a large number of clients to fail at close to the
same time, then even if they're backing off, the schedule
on which they do could still be close enough to each other
that every subsequent retry will hammer the downed service.
This is known as [the thundering herd
problem][thundering-herd]. However, if we add some amount
of random "jitter" to the wait time, then there's enough
schedule variance between clients that they're less likely
to be a problem.

The Stripe Ruby bindings retry on failure automatically
with an idempotency key using increasing backoff times and
jitter. The implementation for that is pretty simple, and
[you can refer to it yourself][stripe-ruby] to see exactly
how it works.

## Wrapping Up (#wrapping-up)

If you're going to take anything away from this article,
consider the following points:

1. Make sure that failures are handled. Not doing so could
   leave data managed by a remote service in an
   inconsistent state that will lead to later problems.

2. Make sure that failures are handled _safely_ using
   idempotency and idempotency keys.

3. Make sure that failures are handled _responsibly_ by
   using techniques like exponential backoff and random
   jitter. Be considerate of remote services that may be
   stuck in a degraded state.

[exponential-backoff]: https://en.wikipedia.org/wiki/Exponential_backoff
[stripe-keys]: https://stripe.com/docs/api?lang=curl#idempotent_requests
[stripe-ruby]: https://github.com/stripe/stripe-ruby/blob/98c42e0b69d2c9e3be64d62f13e8d7b6d44ee3e5/lib/stripe.rb#L441-L455
[thundering-herd]: https://en.wikipedia.org/wiki/Thundering_herd_problem
