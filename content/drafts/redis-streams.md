---
title: Redis Streams and the Unified Log
published_at: 2017-11-01T17:24:20Z
location: San Francisco
hook: TODO
---

Years ago, LinkedIn [wrote an article about the unified
log][thelog], a useful architectural pattern for services
in a distributed system converge state with one another. It
was a refreshingly novel idea at the time, and still is:
Kafka may be more prevalent in 2017, but most of us are
still gluing components together with little more than
patches and duct tape.

In the log's design, services emit state changes into an
ordered data structure where each new record gets a unique
ID. Unlike a queue, a log is durable across any number of
reads until the log is explicitly truncated.

TODO: Diagram

Consumers track changes in the wider system by consuming
the log. Each one maintains the ID of the last record it
successfully consumed and aims to consume every record at
least once -- no records should be missed. When a consumer
is knocked offline, it looks up the last ID that it
consumed, and continues reading the log from there.

The article above points out that this design is nothing
new -- we've been using logs in various forms in computer
science for decades. A common example is how a Postgres
installation streams changes to its replicas over the WAL
(write-ahead log). Changes in the database's physical
structure are written as records to the WAL, and each
replica reads records and applies them to their own state.
Tuple `123` was added, tuple `124` was updated, tuple `125`
was deleted. The WAL is saved in segments and in a
production environment often uploaded to a service like S3
for durable access and high availability.

## At-least once design (#at-least-once)

On systems powered by a unified log, resilience and
correctness are the name of the game. Consumers should get
every message that a producer sends, and to that end
processes are built to guarantee **at-least once** delivery
semantics. Messages are usually sent once, but in cases
where there's uncertainty around whether the transmission
occurred, a message will be send as many times as necessary
to be sure.

Exactly-once delivery is a panacea of distributed systems,
but even if possible, it would be a costly guarantee to
make in the additional overhead that would be needed in
consumers and producers. In practice, at-least once
semantics are fine to handle as long as consumers are built
to handle it from the beginning.

## Redis streams (#redis-streams)

Kafka is a popular system component that's a great log
implementation. Unfortunately, Kafka is heavy software
that's difficult to get configured and costly to run.
Pricing on Heroku costs $100 a month, and once you factor
in server and personnel costs, it's probably going to cost
you more to do it yourself. There are many alternatives,
but most of them are either also non-trivial to setup and
maintain, obscure, or design problems that make their use
awkward (e.g. [poor fanout in Kinesis][fivereads]).

### Configuring Redis for durability (#redis-durability)

## Streaming Rocket Rides (#rocket-rides-streaming)

### The streamer (#streamer)

### Consumers & checkpointing (#consumers)

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
consumer0.1 | Consumed record: {"id":521,"distance":539.836923415231}
              total_distance=257721.7m
consumer1.1 | Consumed record: {"id":521,"distance":539.836923415231}
              total_distance=257721.7m
```

[fivereads]: /kinesis-in-production#five-reads
[thelog]: https://engineering.linkedin.com/distributed-systems/log-what-every-software-engineer-should-know-about-real-time-datas-unifying
