+++
hook = "TODO"
location = "San Francisco"
published_at = 2019-08-27T16:25:46Z
title = "Building Failure-resistant Applications with Go Contexts"
+++

Your database is severely degraded. It's not hard down, but anything sent to
it is taking far longer to run than usual. While you might hope that only
_many_ of the requests to your service are failing, but the practical result
is that almost all of them are. Expensive operations are timing out midway as
might be expected, but even cheaper operations that might have a chance of
success are also timing out due to heavy resource contention with other
requests in the system.

The wishful answer to the scenario above is to not have a degraded database
in the first place, but these kinds of undesirable major problems happen to
everyone once in a while. A more pragmatic question that we should ask
ourselves is: given the presence of a major incident that's failing a lot of
production's traffic, how can we keep as much of it succeeding as possible?

## Culling the living dead (#living-dead)

There are techniques that might be applied to the problem, but the one we'll
focus on is the idea of **conserving resources by prematurely cancelling
requests that aren't likely to succeed**.

A very common practice in services is to have a **timeout** on requests so
that extremely long-lived requests don't stick around forever tarring up the
system. Normally a timeout kicking in is an exceptional case, but during an
incident (again, say the database is degraded), everything is taking longer
and requests that should have succeeded normally are hitting timeouts.

TODO: Diagram showing request taking longer than timeout.

**Premature cancellation** involves identifying requests
that are likely to time out and cancel them early so that we waste as few
resources as possible on the failure. The cancellation helps to minimize
database load as a request's latter operations don't occur, and frees up
worker capacity as workers with cancelled requests are now free to service
another.

Premature cancellation can happen anywhere, with the trick being to work it
into code so that it happens quite generally without developers working on
individual features having to think about it all the time. To that end, a
couple of good places to put checks for it are:

1. Before the main body of a service request starts. 2. Before performing a
   database operation. 3. Before performing an external service request.

## The Context struct (#context)

```go
ctx, cancel = context.WithCancel(context.Background())
```

## Sample application & building blocks (#sample-app)

## Powerful built-ins (#built-ins)
