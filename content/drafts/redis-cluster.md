---
title: Sharding Stripe's Rate Limiting Stack With Redis
  Cluster
location: San Francisco
published_at: 2018-03-17T16:52:38Z
hook: Flattening a single very hot vertical Redis node out
  out horizontally with Redis Cluster.
---

Redis is the often unspoken workhorse of production. It's
not often used as a primary data store, but it has a sweet
spot in storing and accessing ephemeral data whose loss can
be tolerated -- metrics, session state, caching -- and it
does so _fast_, providing not only optimal performance, but
efficient algorithms on a useful set of built-in data
structures. It's a common staple in the modern technology
stack.

Stripe's rate limiters are built on top of Redis, and until
recently, they ran on a single _very hot_ instance of
Redis. The server had followers in place for failover, but
at any given time, one node was handling every operation.

In retrospect, it was impressive that this was even
possible. Every rate limiting operation requires multiple
Redis commands to run, and every API request passes
through a series of rate limiters. One node was handling on
the scale of tens of thousands of operations per second and
up [1].

The node's saturation meant that there was an ambient level
of failures occurring. Most of the time this was okay
because clients were specifically written to tolerate Redis
unavailability, but the severity could escalate under
certain conditions. We eventually migrated to a 10-node
[Redis Cluster][rediscluster] with good success. Impact on
performance was negligible, and we now have an easy knob to
turn for horizontal scalability.

The before and after error cliff [2]:

!fig src="/assets/redis-cluster/errors.png" caption="Errors subsiding after a transition to Redis Cluster."

## The limits of operation (#limits)

Before replacing a system, it's worth understanding the
cause and effect that led the original to fail.

A property of Redis that's worth understanding is that it's
a single threaded program. This isn't _strictly_ true
anymore because background threads handle some operations
like object deletion, but it's spiritually true in that all
executing operations block on access to a single flow
control point. It's relatively easy to understand how this
came about too -- Redis' guarantee around the atomicity of
any given operation (be it a single command, `MULTI`, or
`EXEC`) stems from the fact that it's only executing one of
the them at a time. Even so, there are some obvious
parallelism opportunities, and [notes in the FAQ][threaded]
suggest that the intention is to start investigating a more
threaded design beyond 4.0.

Redis' single-threaded model was indeed our bottleneck. You
could log onto the original node and see a single core
pegged at 100% usage.

### Intersecting failures (#intersecting-failures)

Even operating right at maximum capacity, we found Redis to
degrade quite gracefully. The main manifestation was an
increased rate of baseline connectivity errors as observed
from the nodes talking to Redis -- in order to be tolerant
of a malfunctioning Redis they were constrained with
aggressive connect and read timeouts (~0.1 seconds), and
couldn't establish a connection of execute an operation
within that time when dealing with a overstrained target
host.

Although not optimal, this situation was mostly okay. The
real problem came in when we were targeted with a major
spike of illegitimate traffic (i.e., an order of magnitude
or more over allowed limits) from a _legitimate_ user who
could authenticate successfully and run an expensive
database operation. That's _expensive_ in the relative
sense -- even returning a set of objects from a list
endpoint is far more expensive than denying the request
with a `401` because its authentication wasn't valid, or a
`429` because it's over limit.

These traffics spikes would lead to a proportional increase
in error rate, and much of that traffic would be allowed
through because the rate limiters defaulted to allowing the
request given an error condition. That would put increased
pressure on our backend database, and when it's overloaded
it doesn't fail as gracefully as Redis, and is prone to
partitions becoming inoperable and timing out a sizable
number of the requests made to them.

## Redis Cluster's sharding model (#sharding)

A core design value of Redis is speed, and Redis Cluster is
structured so as not to compromise that. Unlike many other
distributed models, operations in Redis Cluster aren't
confirming on multiple nodes before reporting a success,
and instead look a lot more like a set of independent
Redis' sharing a workload by divvying up the total space of
possible work. This sacrifices availability in favor of
keeping operations fast -- indeed, the additional overhead
of running an operation against a Redis Cluster is
negligible compared to a standard Redis.

The total set of possible keys are divided into 16,384
_slots_. A key's slot is calculated with a stable hashing
function that all clients know how to do:

```
HASH_SLOT = CRC16(key) mod 16384
```

Each node in a cluster will handle some fraction of those
total 16,384 slots, with the exact number depending on the
number of nodes. Nodes within the cluster communicate with
each other to coordinate slot distribution and rebalancing.

!fig src="/assets/redis-cluster/hash-slots.svg" caption="The set of hash slots spread across nodes in a cluster."

A set of `CLUSTER` commands are available so that clients
can query the cluster's state. For example, a common
operation is to issue `CLUSTER NODES`
to get a list of nodes in the cluster and see see
which slots are being handled by which nodes.

Clients use the `CLUSTER` family of commands to query a
cluster's state. A common operation is `CLUSTER NODES` to
get a mapping of slots to nodes, the result of which is
generally cached locally as long as it stays fresh.

```
127.0.0.1:30002 master - 0 1426238316232 2 connected 5461-10922
127.0.0.1:30003 master - 0 1426238318243 3 connected 10923-16383
127.0.0.1:30001 myself,master - 0 0 1 connected 0-5460
```

I've simplified the output above, but the important parts
are the host addresses in the first column and the numbers
in the last. `5461-10922` means that this node handles the
range of slots starting at `5461` and ending at `10922`.

### `MOVED` redirection (#moved)

If a node in a Redis Cluster receives a command for a key
in a slot that it doesn't handle, it makes no attempt to
forward that command to get a success. Instead, the client
is told to try again somewhere else. This comes in the form
of a `MOVED` response with the address of the new target:

```
GET foo
-MOVED 3999 127.0.0.1:6381
```

During a cluster rebalancing, slots migrate from one
node to another, and `MOVED` is an important signal that
servers use to tell a client its local mappings of slots to
nodes are stale.

!fig src="/assets/redis-cluster/moved-redirection.svg" caption="A slot migrating from one node to another."

Sending `MOVED` to a client instead of having the server
try to transparently retrieve the result from a sibling is
also an important design choice for keeping performance
deterministic -- operations are always executed against the
server which the client is immediately talking to, and
because on the whole slots will rarely be moving around,
the extra coordination overhead is generally negligible.

### How clients execute requests (#client)

Redis clients need a few extra features to support Redis
Cluster, with the most important ones being support for the
key hashing algorithm, and a scheme to maintain slot to
node mappings so that they know where to dispatch commands.

Generally, a client will operate like this:

1. On startup, connect to a node and get a mapping table
   with `CLUSTER NODES`.
2. Execute commands normally, targeting servers according
   to key slot and slot mapping.
3. If `MOVED` is received, return to 1.

A multi-threaded client can be optimized by having it
merely mark the mappings table dirty when receiving
`MOVED`, and have threads executing commands follow `MOVED`
responses with new targets while a background thread
refreshes the mappings asynchronously. In practice, even
while rebalancing the likelihood is that most slots won't
be moving, so this model allows _most_ commands to continue
executing with no overhead.

### Localizing multi-key operations with hash tags (#hash-tags)

It's quite common in Redis to run operations that operate
on multiple keys through the use of the `EVAL` command and
a custom Lua script. This is an especially important
feature for implementing rate limiting, because all the
work dispatched via a single `EVAL` is guaranteed to be
atomic -- this allows us to correctly calculate remaining
quotas even where there's other concurrent operations in
flight that might conflict.

A distributed model would make this type of multi-key
operation difficult. Because the slot of each key is
calculated via hash, there'd be no guarantee that related
keys would map to the same slot. My keys
`user123.first_name` and `user123.last_name`, obviously
meant to belong to the same relation, could end up on two
completely different nodes in the cluster. An `EVAL` that
read from both of them wouldn't be able to run on a single
node without an expensive remote fetch from another.

Say for example we have an `EVAL` operation that
concatenates a first and last name to produce a person's
full name:

```
# Gets the full name of a user
EVAL "return redis.call('GET', KEYS[1]) .. ' ' .. redis.call('GET', KEYS[2])"
    2 "user123.first_name" "user123.last_name"
```

A sample invocation:

```
> SET "user123.first_name" William
> SET "user123.last_name" Adama

> EVAL "..." 2 "user123.first_name" "user123.last_name"
"William Adama"
```

This script wouldn't run correctly if Redis Cluster didn't
provide a way for it to do so. Luckily it does through the
use of ***hash tags***.

The Redis Cluster answer to `EVAL`s that would require
cross-node operations is to disallow them (a choice that
once again optimizes for speed). Instead, it's the user's
jobs to ensure that the keys that are part of any
particular `EVAL` map to the same slot by hinting how a
key's hash should be calculated with a hash tag. Hash tags
look like curly braces in a key's name, and they dictate
that only the surrounded part of the key is used for
hashing.

We'd fix our script above by redefining our keys to only
hash their shared `user123`:

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
using it can get away with a simple standalone instance.
Error rates took a nosedive, and we're confident that we've
got a lot of runway for continued growth.

More philosophically, I'm a fan of Redis Cluster's simple
and yet powerful design. Especially when it comes to
distributed systems, many implementations are exceedingly
complicated, and failure to understand the specifics of
their operation can be catastrophic when encountering a
tricky edge in production. Redis Cluster is scalable, and
yet with few enough moving parts that even a layperson like
myself can mostly wrap their head around it. Its design
docs [are comprehensive and yet approachable][spec].

We need more software like this.

!fig src="/assets/redis-cluster/sharding.jpg" caption="Your daily dose of tangentially related photography: Stone at the top of Massive Mountain in Alberta sharding into thin flakes."

[1] The number of operations per second is left
    intentionally vague.

[2] Notably, we're not error-free. There are enough
    operations in flight that some level of intermittent
    failure is unavoidable.

[client]: TODO
[rediscluster]: https://redis.io/topics/cluster-tutorial
[spec]: https://redis.io/topics/cluster-spec
[threaded]: https://redis.io/topics/faq#redis-is-single-threaded-how-can-i-exploit-multiple-cpu--cores
