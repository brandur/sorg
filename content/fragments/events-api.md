+++
hook = "Event APIs are a good idea (and probably a _better_ idea), but do come with a few downsides."
published_at = 2021-07-21T19:05:42Z
title = "Considerations for event APIs (versus webhooks)"
+++

An article hit last week that [argued in favor of events API over webhooks](https://blog.syncinc.so/events-not-webhooks), specifically talking about consuming Stripe's `/v1/events`, which similarly to webhooks, captures all changes that occur in an account. Whenever an API resource is added, updated, or deleted, a new `/v1/events` record is made available to show the change.

I'm not a big fan of webhooks either. They're easy to use, but have some major usability downsides for consumers:

* They require exposing a public endpoint, which is inconvenient in development and in more complex deployment configurations involving VPCs, ingress rules, reverse proxies, etc.
* Configuring them usually involves logging into a web dashboard and manually entering URLs.
* They don't provide any kind of historical archive, and a missed webhook might be hard to identify.
* They're not ordered.

And from the standpoint of the webhook provider:

* Their security is questionable. You can include a signature, but you can't check that a consumer is validating it.
* They're inefficient. It's conventional to issue thousands of noisy requests which could easily have been batched together for better economy.
* Disappearing endpoints and slow endpoints inevitably cause trouble. It's possible to build a robust webhooks implementation, but it's much harder than most people think.

With all that said, it's worth nothing that an events endpoint isn't a free lunch. Let's look at a few pitfalls.

## Horizon trimming (#horizon-trimming)

In Stripe's implementation, events have IDs that are roughly ordered like `evt_123`, `evt_124`, `evt_125`, etc., and these IDs are used for pagination by asking for the next page like `?starting_after=evt_123`. But although they're roughly ordered, they are not _exactly_ ordered. IDs are generated as `evt_<time component><random component>` so that they're mostly ordered by time, but include a random component to avoid collisions within the same epoch. Moreover, IDs aren't necessarily generated at the time the record is inserted -- it's possible that one is generated early on in a long API call, and inserted at the end of it, resulting in many seconds of delay.

This means that events may be inserted out of order. When a consumer consumes up to a given ID like `evt_125`, it can't assume that when checking again that this is the new ceiling -- `evt_124` may later be inserted retroactively. A robust consumer should page ~60 seconds back in time on each fetch, even if it means covering already consumed events.

The problem is particularly insidious because especially when initially implementing a consumer and there's relatively few records, it won't even be happening  -- events are spaced far enough from each other in time that IDs will be ordered. This will likely lead most developers to assume that IDs are always ordered, when they're not. I'd venture a guess that most implementations will be wrong by default, and only corrected later after the problem is noticed (likely after a prolonged debugging phase).

It's possible for a provider to avoid the problem by building an events API based on a central system that can guarantee that IDs are assigned monotonically, but it needs to be considered carefully. A Postgres sequence could do the job, but might become a scaling bottleneck. Backing it with Kafka would also work, but then you're running Kafka.

## Overload (#overload)

One of the virtues of webhooks is that they parallelized quite well. If a user gets very big and needs to receive a lot of webhooks, all they need to do is scale up the endpoint that receives them, and because it's a web API, they've probably already got an easy way to do that.

With events, not so much. Stripe's maximum page size is 100, and with cursors, pagination is generally not parallelizable. Each list request is not particularly fast (Stripe's probably take a few hundred milliseconds or so per request), so if you're generating on the order of 100+ events per second, you might find that it's not possible to use `/v1/events` to keep up [1].

Again, this is solvable, probably with something like `?partition=` parameter and multiple cooperating consumers, but like the horizon trimming problem, it requires significantly more upfront design, a more elaborate API, and a more complex implementation. On the provider's side, many large and frequent fetches are likely to put more pressure on the database, whereas that would be minimal with webhooks (you can often just fire payloads into a Kafka topic and scale out consumers as necessary).

## What to do today (#today)

Now, despite these considerations, the downsides of webhooks are so consideration (in my opinion), that building from the ground up today I'd still be most tempted to come up with solutions and try to make an events API work.

[1] I'll caveat that there are ways to [parallelize cursor-based pagination](/fragments/offset-pagination), but it gets messy.
