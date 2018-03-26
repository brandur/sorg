---
title: Sharding Stripe's Rate Limiting Stack With Redis
  Cluster
location: San Francisco
published_at: 2018-03-17T16:52:38Z
hook: Flattening a single very hot vertical Redis node out
  out horizontally with Redis Cluster.
---

Until recently, Stripe's rate limiting stack ran on a
single very hot instance of Redis. It had a followers in
place that were ready to be promoted in case the primary
failed, but at any given time, it was one node that handled
every operation.

Not having a strategy in place for scaling horizontally
wasn't a good thing, but it was impressive that Redis could
even handle on the order of thousands of operations per
second (and up [1]).

It's made somewhat more impressive when you know that Redis
is for all practical purposes a single-threaded system.
This is a design that's baked deeply into its architecture
in that handling one operation at time is also how it
guarantees that each is performed atomically and free of
conflict from other operations. If you log into a Redis
node operating near capacity, you'll see one CPU totally
saturated, and the rest essentially totally idle.

TODO: Diagram on Redis blocking

## Intersecting failures (#intersecting-failures)

This model worked well for a long time in large part
because Redis happens to degrade quite gracefully -- as a
box starts to brush up against its maximum capacity, the
vast majority of operations still complete successfully. A
few will fail or block longer than expected, but we'd
configured clients with low connect and read timeouts (~0.1
seconds), so most of the time even in the event of a
problem, they'd fail open and continue serving the request
with no ill effect.

Once in a while, Redis would find itself in a particularly
bad spot, and produce a spike of failed operations. Even
when this wasn't the case, we could still see the ambient
error level climbing slowly as months passed and traffic
increased.

Even during an error spike, we were still fine most of the
time -- the stack would fail open and users wouldn't
notice. The real problem occurred only during the
intersection of two unusual events where Redis was shedding
operations _and_ we had a user throwing traffic our way
far above normal levels.

In the beginning the latter was relatively rare -- you'd
see the occasional script run awry and we'd shed most of
the traffic by responding with `429 Too many requests`, but
as your userbase grows, you get to a point where you see
this kind of thing more often -- eventually _somebody_ is
doing it almost all the time.

So the likelihood of that intersection of failures
eventually became reasonably high -- if Redis got itself
into trouble, we had to rely on luck alone that we didn't
have someone hitting us with too much traffic. That was
obviously untenable, so we started a project to scale the
system our horizontally. Redis Cluster was a natural choice
-- the entire system was already heavily Redis-based (the
data structures built into Redis lend themselves
particularly well to rate limiting operations), and AWS had
built-in support for it through ElastiCache.

## Redis Cluster's sharding model (#sharding)

A core design value of Redis is speed, and Redis Cluster is
structured so as not to compromise that. Unlike many other
distributed models, nodes in a Redis Cluster aren't
generally talking to each other to build consensus on the
result of an operation, and instead look a lot more like a
set of independent Redis' sharing a workload by divvying up
the total space of possible work.

In Redis Cluster, the total set of possible keys are
divided into 16,384 _slots_, where a key's slot is
calculated with a simple hashing function that all clients
know ahead of time:

```
HASH_SLOT = CRC16(key) mod 16384
```

Each node in the cluster will handle some fraction of those
total 16,384 slots, with the exact number depending on the
number of nodes in the cluster. Nodes communicate with each
other to coordinate partitioning and slot rebalancing (if
need be).

TODO: Diagram of sharding model.

A set of `CLUSTER` commands are available so that clients
can query the cluster's state. For example, a common
operation is to issue `CLUSTER LIST` (TODO: this might not
be right) to get a list of nodes in the cluster and see see
which slots are being handled by which nodes.

TODO: Mappings example (get from Splunk)

### `MOVED` redirection (#moved)

If a node receives a command for a key in a slot that it
doesn't handle, it makes no attempt to forward that command
to get a success, and instead tells the client to take care
of it. It sends back a `MOVED` response with the address of
the node that can handle the operation:

```
GET foo
-MOVED 3999 127.0.0.1:6381
```

During a cluster rebalancing, slots can migrate from one
node to another, and `MOVED` is an important signal that
servers use to tell a client its local mappings of slots to
nodes are stale.

TODO: Diagram of `MOVED`.

Sending `MOVED` to a client instead of having the server
try to transparently retrieve the result from a sibling is
also an important design choice for keeping performance
deterministic -- operations are always executed against the
server which the client is immediately talking to, and
because on the whole slots will rarely be moving around,
on average using Redis Cluster will have a negligible
performance disadvantage compared to a single raw Redis
node.

### Client behavior (#client-behavior)

The major additions required to a Redis client to build in
Redis Cluster support are the key hashing algorithm and a
scheme to maintain a mapping of slots to nodes so that it
knows where to dispatch commands.

Generally, a client will operate like this:

1. On startup, connect to a node and get a mapping table
   with `CLUSTER LIST`.
2. Execute commands normally, targeting servers according
   to key slot and slot mapping.
3. If `MOVED` is received, return to 1.

An optimization for a multi-threaded client is for it to
merely mark the mappings table dirty when receiving
`MOVED`, and to have any given thread redirect a command to
the server address in the `MOVED` response. In practice,
even while rebalancing, _most_ slots won't be moving, so
this allows _most_ commands to continue executing normally
while a background thread issues `CLUSTER LIST` (TODO),
waits on the response, and refreshes the client's mappings
asynchronously.

### Localizing multi-key operations with hash tags (#hash-tags)

It's quite common in Redis to run operations that operate
on multiple keys through the use of the `EVAL` command and
a custom Lua script. This is an especially important
feature for implementing rate limiting, because all the
work dispatched via a single `EVAL` is guaranteed to be
atomic -- this allows us to correctly calculate remaining
quotas even where there's other concurrent operations in
flight that would otherwise conflict.

A distributed model would make this type of multi-key
operation difficult. Because the slot of each key is
calculated via hash, there'd be no guarantee that related
keys would map to the same slot. My keys
`user123.first_name` and `user123.last_name`, obviously
meant to belong to the same relation, could end up on two
completely different nodes in the cluster. An `EVAL` that
read from both of them wouldn't be able to run on a single
node without an expensive remote fetch from another.

To concrete this with a simple example, let's say we have
an `EVAL` operation that concatenates a first and last name
to produce a full name:

```
# Gets the full name of a user
EVAL "return redis.call('GET', KEYS[1]) .. ' ' .. redis.call('GET', KEYS[2])"
    2 "user123.first_name" "user123.last_name"
```

(The `2` in the line above is just the client telling the
server how many key arguments it's going to send.)

Here's a sample invocation:

```
> SET "user123.first_name" William
> SET "user123.last_name" Adama

> EVAL "..." 2 "user123.first_name" "user123.last_name"
"William Adama"
```

This script would have trouble running on Redis Cluster if
it didn't provide a mechanic to allow it. Luckily it does
through the use of "hash tags".

The Redis Cluster answer to `EVAL`s that would require
cross-node operations is to disallow them (a choice that
optimizes for speed). Instead, it's the user's jobs to
ensure that the keys that are part of any particular `EVAL`
map to the same slot by hinting how a key's hash should be
calculated with a hash tag. Hash tags look like curly
braces in a key's name, and they dictate that only the
surrounded part of the key is used for hash calculation.

We'd fix our script above by redefining our keys to only
use their shared `user123` content for hashing:

```
> EVAL "..." 2 "{user123}.first_name" "{user123}.last_name"
```

`{user123}.first_name` and `{user123}.last_name` are now
guaranteed to map to the same slot, and `EVAL` operations
that contain both of them will be trouble-free. And
although this is only a basic example, the exact same
concept maps all the way up to a complex rate limiter
implementation.

## Reliable and simple (#simple)

Transitioning over to Redis Cluster went remarkably
smoothly, with the most difficult part being shoring up one
of the Redis Cluster clients for production use. Even to
this day, good client support is somewhat spotty, which may
be an indication that Redis is fast enough that most people
using it can get away with a simple standalone instance. We
saw our error rates nosedive, and are reasonably confident
that we've got a lot of runway for continued growth.

This is what I really like about Redis: I've spent almost
no time at all studying it source code, but most of the
mechanics it employs (even for complex features like
distributed sharding) are simple enough for even a
layperson like myself to understand and reason about.

!fig src="/assets/redis-cluster/sharding.jpg" caption="Your daily dose of tangentially related photography: Stone at the top of Massive Mountain in Alberta sharding into thin flakes."

[1] I'll stay intentionally vague on the number, the we
serve many requests and each request goes through multiple
layers of rate limiters.

[client]: TODO
[spec]: https://redis.io/topics/cluster-spec
