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
Tuple `123` was added; tuple `124` was updated; tuple `125`
was deleted. The WAL is saved in segments and in a
production environment often uploaded to a service like S3
for durable access and high availability.

## At-least once design (#at-least-once)

On systems powered by a unified log, resilience and
correctness are the name of the game. Consumers should get
every message that a producer sends, and to that end
processes are built to guarantee ***at-least once***
delivery semantics. Messages are usually sent once, but in
cases where there's uncertainty around whether the
transmission occurred, a message will be send as many times
as necessary to be sure.

At-least once delivery is opposed to ***best-effort
delivery*** where messages will be received once under
normal conditions, but may be dropped in extraordinary
cases. It's also opposed by ***exactly-once delivery***; a
classic panacea of distributed systems. Exactly-once
delivery is a difficult guarantee to make, and even if
possible, would add costly overhead to transmission. In
practice, at-least once semantics are fine to handle as
long as consumers are built with consideration for it from
the beginning.

## Redis streams (#redis-streams)

I was happy to hear recently that Redis will soon [1] be
shipping with a new data primitive that will make an
admirable backend for unified logs: [streams][streams].
They're distinct from Redis lists in that records in them
are assigned with addressable IDs, and they lend themselves
well to be being read more than once. They're a
near-perfect analog for what we've been talking about so
far.

about so far. Records are added to one with the new `XADD`
command:

```
> XADD rocket-rides-log * id 123 distance 456.7
1506871964177.0
```

This adds a record to the `rocket-rides-log` stream with
the fields `id = 123` and `distance = 456.7`. Redis
responds with a unique ID for the record within the stream
that's made up of a timestamp and a sequence number (`.0`)
to disambiguate within the millisecond.

The other important command is `XRANGE`, which allows us to
read records from the stream:

```
> XRANGE rocket-rides-log - + COUNT 2
1) 1) 1506871964177.0
   2) 1) "id"
      2) "123"
      3) "distance"
      4) "456.7"
2) 1) 1506872463535.0
   2) 1) "id"
      2) "124"
      3) "distance"
      4) "89.0"
```

The tokens `-` and `+` are special in that they tell Redis
to read from the first available record in the stream and
up to the last available record in the stream respectively.
Either one can be replaced with an ID like
`1506871964177.0` to read from or up to a specific record.
Using this capability allows us to slice out just records
that we haven't consumed yet.

Streams have a variety of other useful features that I
won't cover here, but [the original blog post covers
everything pretty well][streams].

### Versus Kafka (#kakfa)

Kafka is a popular system component that also makes a great
backend for a unified log implementation, and once
everything is in place, probably a better one than Redis
because of its design around high availability and advanced
features.

What's exciting about Redis streams is that they bring an
easy way of building a unified log in a small and/or
inexpensive app. Kafka is infamously difficult to configure
and get running, and is expensive to operate once you do.
Pricing for a small Kafka cluster on Heroku costs $100 a
month and climbs steeply from there. It's temping to think
you can do it more cheaply yourself, but after factoring in
server and personnel costs along with the time it takes to
build a working expertise in the system, it'll probably
cost more.

Redis on the other hand is probably already in your stack.
Being the Swiss army knife of cloud persistence, it's
useful for a multitude of things including caching, rate
limiting, or storing user sessions. Even if you don't have
it, you can compile it from source and get it configured in
running in about thirty seconds, and there are dozens of
cloud providers that will offer you a hosted version.

Once you're operating at serious scale, switch to Kafka. In
the meantime, Redis streams make a great and cheap
alternative.

### Configuring Redis for durability (#redis-durability)

One highly desirable property of a unified log is that it's
***durable***, meaning that even if its host crashes or
something else terrible happens, it doesn't lose any
information that producers that had been persisted.

By default Redis is not durable; a sane configuration
choice when it's been used for caching or rate limiting,
but not when it's being used for a log. To make Redis fully
durable, tell it to keep an append-only file (AOF) with
`appendonly` and instruct it to perform fsync on every
command that's written to the AOF with `appendfsync
always` (more details [in the Redis documentation on
persistence][persistence]):

```
appendonly yes
appendfsync always
```

Note that there's an inherent tradeoff between durability
and performance (ever wonder [how MongoDB performed so well
on its early benchmarks?][mongodurability]). Redis doing
the extra work to keep an AOL and performing more fsyncs
will make commands slower. If you're using it for multiple
things, it might be useful to make a distinction between
places where ephemerality is okay and where it isn't, and
run two separate Redis's with different configuration.

## Unified Rocket Rides (#rocket-rides-unified)

We're going to be returning to the Rocket Rides example
that we talked about while implementing [idempotency
keys](/idempotency-keys). As a quick reminder, Rocket
Rides is a small Lyft-like app that lets its users get
rides with pilots wearing jetpacks; a vast improvement over
the banality of a car.

As new rides come in, the Rocket Rides API will emit a new
record to the stream that contains the ID of the ride and
the distance traveled. From there, a couple different
consumers will read the stream and keep a running tally of
the total distance traveled for every ride in the system
that's ever been taken.

TODO: Diagram of client -> API -> stream -> consumers

Both producer and consumers will be using database and
transactions at-least once semantics to guarantee that all
information is correct. No matter what kind of failures
occur in clients, API, consumers, or elsewhere in the
system, the totals being tracked by consumers should always
agree with each other for any given Redis or ride ID.

A working version of all this code is available in the
[_Unified Rocket Rides_][unifiedrides] repository. It might
be easier to download that code and follow along that way:

``` sh
git clone https://github.com/brandur/rocket-rides-unified.git
```

### The API (#api)

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

It's well and good for me to claim that this system is
fault-tolerant, but it's a more believable claim when we
prove it. Operating at our small scale we're unlikely to
see many problems in this system, so I've written the
processes so that they simulate some. 10% of the time, the
streamer will double-send every event in a batch. This
models it failing midway through sending a batch and having
to retry the operation.

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
identical total for ride ID `521`:

```
consumer0.1 | Consumed record: {"id":521,"distance":539.836923415231}
              total_distance=257721.7m
consumer1.1 | Consumed record: {"id":521,"distance":539.836923415231}
              total_distance=257721.7m
```

[1] The expectation currently is that streams will be
available in the Redis 4.0 series by the end of the year.

[mongodurability]: /fragments/mongo-durability
[persistence]: https://redis.io/topics/persistence
[thelog]: https://engineering.linkedin.com/distributed-systems/log-what-every-software-engineer-should-know-about-real-time-datas-unifying
[streams]: http://antirez.com/news/114
[unifiedrides]: https://github.com/brandur/rocket-rides-unified
