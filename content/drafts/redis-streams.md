---
title: Redis Streams and the Unified Log
published_at: 2017-11-01T17:24:20Z
location: San Francisco
hook: TODO
---

## At least once design (#at-least-once)

## Redis streams (#redis-streams)

### Configuring Redis for durability (#redis-durability)

## Streaming Rocket Rides (#rocket-rides-streaming)

### The streamer (#streamer)

### A transactional consumer (#consumer)

### Non-transactional consumers (#non-transaction)

Consumers don't necessarily have to be transactional as
long as the work they do can be applied cleanly given
at-least-once semantics.

Notably our example here wouldn't work without the
transaction. With the transaction removed, if the consumer
successfully updated `total_count` but failed to set the
checkpoint, then it would double-count the distance of
those records when it retried consuming them.

But if a consumer using input records to execute operations
which are _idempotent_, then a wrapping transaction isn't
necessarily needed. An example of this is a consumer that's
reading a stream to add or remove information into a data
warehouse. As long as creation records are treated as
something like an upsert instead of `INSERT` and a deletion
is tolerant if the target doesn't exist, then all
operations can safely be considered to be idempotent.

### Simulating failure (#simulating-failure)

Operating at our small scale we're unlikely to see many
problems in the system, so the processes are written to
simulate some. 10% of the time, the streamer will
double-send every event in a batch. This models it failing
midway through sending a batch and having to retry the
operation.

Likewise, each consumer will crash 10% of the time after
handling a batch but before committing its transaction.

Despite these artificial problems, because the system's
designed to handle these edge cases, the results will
always be correct and consistent.

Run `forego start` and leave the fleet of processes running
for a while. Despite the occasional double sends and each
consumer failing randomly and independently, no matter how
long you wait, the consumers should always show the same
`total_distance` reading for any given consumed ID.

For example, here's `consumer0` and `consumer1` showing an
identical number for ride ID `521`:

```
consumer0.1 | Consumed record: {"id":521,"distance":539.836923415231} total_distance=257721.7m
consumer1.1 | Consumed record: {"id":521,"distance":539.836923415231} total_distance=257721.7m
```
