---
title: Redis Streams and the Unified Log
published_at: 2017-11-08T15:15:31Z
location: San Francisco
hook: Building a log-based architecture that's fast,
  efficient, and resilient on the new stream data structure
  in Redis.
hn_link: https://news.ycombinator.com/item?id=15653544
---

Years ago an article came out of LinkedIn [about the
unified log][thelog], a useful architectural pattern for
services in a distributed system share state with one
another. In the log's design, services emit state changes
into an ordered data structure in which each new record
gets a unique ID. Unlike a queue, a log is durable across
any number of reads until it's explicitly truncated.

Consumers track changes in the wider system by reading the
log. Each one maintains the ID of the last record it
successfully consumed and aims to read every record at
least once -- nothing should ever be missed. When a
consumer is knocked offline, it restarts and looks up the
last ID that it saw, and continues reading the log from
there.

The log is ***unified*** because it acts as a master ledger
of state changes in a wider system. All components that are
making critical changes write to it, and all components
that need to track distributed state are subscribed.

LinkedIn's article is sober enough to point out that this
design is nothing new: we've been using logs in various
forms in computer science for decades. [Journaling file
systems][journalfs] use them to protect data against
corruption. Databases use them in places like the
[write-ahead log (WAL)][wal] in Postgres as they stream
changes to their read replicas.

!fig src="/assets/redis-streams/unified-log.svg" caption="The unified log: a producer emits to the stream and consumers read from it."

Even so, the unified log was a refreshingly novel idea when
the article was written, and still is. File systems and
databases use the structure because it's effective, and it
lends itself just as well to distributed architectures.
Kafka is more prevalent in 2017, but most of us are still
gluing components together with patches and duct tape.

Chatty services exchange high-frequency messages back and
forth in a way that's slow (they rely on synchrony),
inefficient (single messages are passed around), and
fragile (every individual message introduces some
possibility of failure). In contrast, the log is
asynchronous, its records are produced and consumed in
batches, and its design builds in resilience at every turn.

## Redis streams (#redis-streams)

This brings us to Redis. I was happy to hear recently that
the project will soon [1] be shipping with a new data
structure that's a perfect foundation for a unified log:
[streams][streams]. Unlike a Redis list, records in a
stream are assigned with addressable IDs and are indexed or
sliced with those IDs instead than a relative offset (i.e.
like `0` or `len() - 1`).This lends itself well to having
multiple consumers reading out of a single stream and
tracking their position within it by persisting the ID of
the last record they read.

The new `XADD` command appends to one:

```
> XADD rocket-rides-log * id 123 distance 456.7
1506871964177.0
```

A record with `id = 123` and `distance = 456.7` is appended
to the stream `rocket-rides-log`. Redis responds with a
unique ID for the record that's made up of a timestamp and
a sequence number (`.0`) to disambiguate entries created
within the same millisecond.

`XRANGE` is `XADD`'s counterpart. It reads a set of records
from a stream:

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
that we haven't consumed yet. Specifying `COUNT 2` lets us
bound the number of records read so that we can process the
stream in efficient batches.

### Versus Kafka (#kakfa)

Kafka is a popular system component that also makes a nice
alternative for a unified log implementation; and once
everything is in place, probably a better one compared to
Redis thanks to its sophisticated design around high
availability and other advanced features.

Redis streams aren't exciting for their innovativeness, but
rather than they bring building a unified log architecture
within reach of a small and/or inexpensive app. Kafka is
infamously difficult to configure and get running, and is
expensive to operate once you do. Pricing for a small Kafka
cluster on Heroku costs $100 a month and climbs steeply
from there. It's tempting to think you can do it more
cheaply yourself, but after factoring in server and
personnel costs along with the time it takes to build
working expertise in the system, it'll cost more.

Redis on the other hand is probably already in your stack.
Being the Swiss army knife of cloud persistence, it's
useful for a multitude of things including caching, rate
limiting, storing user sessions, etc. Even if you don't
already have it, you can compile it from source and get it
configured and running in all of about thirty seconds.
Dozens of cloud providers (including big ones like AWS)
offer a hosted version.

Once you're operating at serious scale, Kafka might be the
right fit. In the meantime, Redis streams make a great (and
economic) alternative.

### Configuring Redis for durability (#redis-durability)

One highly desirable property of a unified log is that it's
***durable***, meaning that even if its host crashes or
something terrible happens, it doesn't lose information
that producers think they had persisted.

By default Redis is not durable; a sane configuration
choice when it's been used for caching or rate limiting,
but not when it's being used for a log. To make Redis fully
durable, tell it to keep an append-only file (AOF) with
`appendonly` and instruct it to perform fsync on every
command written to the AOF with `appendfsync always` (more
details [in the Redis documentation on
persistence][persistence]):

```
appendonly yes
appendfsync always
```

There's an inherent tradeoff between durability and
performance (ever wonder [how MongoDB performed so well on
its early benchmarks?][mongodurability]). Redis doing the
extra work to keep an AOF and performing more fsyncs will
make commands slower (although still very fast). If you're
using it for multiple things, it might be useful to make a
distinction between places where ephemerality is okay and
where it isn't, and run two separate Redises with different
configuration.

## Unified Rocket Rides (#rocket-rides-unified)

We're going to be returning to the Rocket Rides example
that we talked about while implementing [idempotency
keys](/idempotency-keys). As a quick reminder, Rocket Rides
is a Lyft-like app that lets its users get rides with
pilots wearing jetpacks; a vast improvement in speed and
adrenaline flow over the every day banality of a car.

As new rides come in, the _Unified Rocket Rides_ API will
emit a new record to the stream that contains the ID of the
ride and the distance traveled. From there, a couple
different consumers will read the stream and keep a running
tally of the total distance traveled for every ride in the
system that's been taken.

!fig src="/assets/redis-streams/streaming-model.svg" caption="Clients sending data to the API which passes it onto the stream and is ingested by stream consumers."

Both producer and consumers will be using database
transactions to guarantee that all information is correct.
No matter what kind of failures occur in clients, API,
consumers, or elsewhere in the system, the totals being
tracked by consumers should always agree with each other
for any given Redis or ride ID.

A working version of all this code is available in the
[_Unified Rocket Rides_][unifiedrides] repository. It might
be easier to download that code and follow along that way:

``` sh
git clone https://github.com/brandur/rocket-rides-unified.git
```

### At-least once design (#at-least-once)

For systems powered by a unified log, resilience and
correctness are the name of the game. Consumers shouldn't
just get most messages that a producer sends, they should
get _every_ message. To that end programs are built to
guarantee ***at-least once*** delivery semantics. Messages
are usually sent once, but in cases where there's
uncertainty around whether the transmission occurred, a
message will be sent as many times as necessary to be sure.

At-least once delivery is opposed to ***best-effort
delivery*** where messages will be received once under
normal conditions, but may be dropped in degraded cases.
It's also opposed by ***exactly-once delivery***; a panacea
of distributed systems. Exactly-once delivery is a
difficult guarantee to make, and even if possible, would
add costly coordination overhead to transmission. In
practice, at-least once semantics are robust and easy to
work with as long as systems are built to consider them
from the beginning.

### The API (#api)

The _Unified Rocket Rides_ API receives requests over HTTP
for new rides from clients. When it does it (1) creates a
ride entry in the local database, and (2) emits a record
into the unified log to show that it did.

``` ruby
post "/rides" do
  params = validate_params(request)

  DB.transaction(isolation: :serializable) do
    ride = Ride.create(
      distance: params["distance"]
    )

    StagedLogRecord.insert(
      action: ACTION_CREATE,
      object: OBJECT_RIDE,
      data: Sequel.pg_jsonb({
        id:       ride.id,
        distance: ride.distance,
      })
    )

    [201, JSON.generate(wrap_ok(
      Messages.ok(distance: params["distance"].round(1))
    ))]
  end
end
```

Rather than emit directly to Redis, a "staged" record is
created in Postgres. This indirection is useful so that in
case the request's transaction rolls back due to a
serialization error or other problem, no invalid data (i.e.
data that was only relevant in a now-aborted transaction)
is left in the log. This principle is identical to that of
[transactionally-staged job drains](/job-drain), which do
the same thing for background work.

The staged records relation in Postgres look like:

``` sql
CREATE TABLE staged_log_records (
    id     BIGSERIAL PRIMARY KEY,
    action TEXT      NOT NULL,
    data   JSONB     NOT NULL,
    object TEXT      NOT NULL
);
```

### The streamer (#streamer)

The streamer moves staged records into Redis once they
become visible outside of the transaction that created
them. It runs as a separate process, and sends records in
batches for improved efficiency.

``` ruby
def run_once
  num_streamed = 0

  # Need at least repeatable read isolation level so that our DELETE after
  # enqueueing will see the same records as the original SELECT.
  DB.transaction(isolation_level: :repeatable_read) do
    records = StagedLogRecord.order(:id).limit(BATCH_SIZE)

    unless records.empty?
      RDB.multi do
        records.each do |record|
          stream(record.data)
          num_streamed += 1

          $stdout.puts "Enqueued record: #{record.action} #{record.object}"
        end
      end

      StagedLogRecord.where(Sequel.lit("id <= ?", records.last.id)).delete
    end
  end

  num_streamed
end

#
# private
#

# Number of records to try to stream on each batch.
BATCH_SIZE = 1000
private_constant :BATCH_SIZE

private def stream(data)
  # XADD mystream MAXLEN ~ 10000  * data <JSON-encoded blob>
  #
  # MAXLEN ~ 10000 caps the stream at roughly that number (the "~" trades
  # precision for speed) so that it doesn't grow in a purely unbounded way.
  RDB.xadd(STREAM_NAME,
    "MAXLEN", "~", STREAM_MAXLEN,
    "*", "data", JSON.generate(data))
end
```

In accordance with at-least once design, the streamer only
removes staged records once their receipt has been
confirmed by Redis. If part of the workflow fails then the
process will run again and select the same batch of records
from `staged_log_records` a second time. They'll be
re-emitted into the stream even if it means that some
consumers will see them twice.

Records are sent to the stream with ascending ride `id`s.
It's possible for a record with a smaller `id` to be
present after one with a higher `id`, but _only_ in the
case of a double-send. With the exception of that one
caveat, consumers can always assume that they're receiving
`id`s in order.

#### Log truncation (#truncation)

Unlike a queue, consumers don't remove records from a log,
and without management it would be in danger of growing in
an unbounded way. In the example above, the streamer uses
the `MAXLEN` argument to `XADD` to tell Redis that the
stream should have a maximum length. The tilde (`~`)
operator is an optimization that indicates to Redis that
the stream should be truncated to _approximately_ the
specified length when it's possible to remove an entire
node. This is significantly faster than trying to prune it
to an exact number.

This rough truncation will work well in most of the time,
but lack of safety measures means that it's _possible_ that
records might be removed which have not yet been read by a
consumer that's fallen way behind. A more resilient system
should track the progress of each consumer and only
truncate records that are no longer needed by any of them.

### Consumers & checkpointing (#consumers)

Consumers read records out of the log in batches and
consume them one-by-one. When a batch has been successfully
processed, they set a ***checkpoint*** containing the ID of
the last record consumed. The next time a consumer restarts
(due to a crash or otherwise), it reads its last checkpoint
and starts reading the log from the ID that it contained.

Checkpoints are stored as a relation in Postgres:

``` sql
CREATE TABLE checkpoints (
    id            BIGSERIAL PRIMARY KEY,
    name          TEXT      NOT NULL UNIQUE,
    last_redis_id TEXT      NOT NULL,
    last_ride_id  BIGINT    NOT NULL
);
```

Recall that in our simple example, consumers add up the
`distance` of every ride created on the platform. We'll
keep this running tally in a `consumer_states` table which
has an entry for each consumer:

``` sql
CREATE TABLE consumer_states (
    id             BIGSERIAL        PRIMARY KEY,
    name           TEXT             NOT NULL UNIQUE,
    total_distance DOUBLE PRECISION NOT NULL
);
```

The code for a consumer to iterate the stream and update
its checkpoint and state will look a little like this:

``` ruby
def run_once
  num_consumed = 0

  DB.transaction do
    checkpoint = Checkpoint.first(name: name)

    # "-" is a special symbol in Redis streams that dictates that we should
    # start from the earliest record in the stream. If we don't already have
    # a checkpoint, we start with that.
    start_id = "-"
    start_id = self.class.increment(checkpoint.last_redis_id) unless checkpoint.nil?

    checkpoint = Checkpoint.new(name: name, last_ride_id: 0) if checkpoint.nil?

    records = RDB.xrange(STREAM_NAME, start_id, "+", "COUNT", BATCH_SIZE)
    unless records.empty?
      # get or create a new state for this consumer
      state = ConsumerState.first(name: name)
      state = ConsumerState.new(name: name, total_distance: 0.0) if state.nil?

      records.each do |record|
        redis_id, fields = record

        # ["data", "{\"id\":123}"] -> {"data"=>"{\"id\":123}"}
        fields = Hash[*fields]

        data = JSON.parse(fields["data"])

        # if the ride's ID is lower or equal to one that we know we consumed,
        # skip it; this is a double send
        if data["id"] <= checkpoint.last_ride_id
          $stdout.puts "Skipped record: #{fields["data"]} " \
            "(already consumed this ride ID)"
          next
        end

        state.total_distance += data["distance"]

        $stdout.puts "Consumed record: #{fields["data"]} " \
          "total_distance=#{state.total_distance.round(1)}m"
        num_consumed += 1

        checkpoint.last_redis_id = redis_id
        checkpoint.last_ride_id = data["id"]
      end

      # now that all records for this round are consumed, persist state
      state.save

      # and persist the changes to the checkpoint
      checkpoint.save
    end
  end

  num_consumed
end
```

Like the streamer, the consumer is designed with at-least
once semantics in mind. In the event of a crash, neither
the `total_distance` or the checkpoint is updated because a
raised exception aborts the transaction that wraps the
entire set of operations. When the consumer restarts, it
happily consumes the last batch again with no ill effects.

Along with a Redis stream ID, a checkpoint also tracks the
last consumed ride ID. This is so consumers can handle
records that were written to the stream more than once. IDs
can always be assumed to be ordered and if a consumer sees
an ID smaller or equal to one that it knows that it
consumes, it safely skips to the next record.

### Simulating failure (#simulating-failure)

I've claimed this system is fault-tolerant, but it's more
believable if I can demonstrate it. Operating at our small
scale we're unlikely to see many problems, so processes are
written to simulate some. 10% of the time, the streamer
will double-send every event in a batch. This models it
failing midway through sending a batch and having to retry
the entire operation.

Likewise, each consumer will crash 10% of the time after
handling a batch but before committing the transaction that
would set its state and checkpoint.

The system's been designed to handle these edge cases and
despite the artificial problems, it will manage itself
gracefully. Run `forego start` (after following the
appropriate setup in `README.md`) and leave the fleet of
processes running. Despite the double sends and each
consumer failing randomly and independently, no matter how
long you wait, the consumers should always stay roughly
caught up to each other and show the same `total_distance`
reading for any given ID.

Here's `consumer0` and `consumer1` showing an identical
total for ride `521`:

```
consumer0.1 | Consumed record: {"id":521,"distance":539.836923415231}
              total_distance=257721.7m
consumer1.1 | Consumed record: {"id":521,"distance":539.836923415231}
              total_distance=257721.7m
```

## Other considerations (#considerations)

### Non-transactional consumers & idempotency (#non-transaction)

Consumers don't necessarily have to be transactional as
long as the work they do while consuming records can be
re-applied cleanly given at-least once semantics. Another
way of saying this is that a consumer doesn't need a
transaction as long as every operation it applies is
***idempotent***.

Notably our example here wouldn't yield correct results
without being nested in a transaction: if it successfully
updated `total_count` but failed to set the checkpoint,
then it would double-count the distance of those records
the next time it tried to consume them.

But if all operations are idempotent, we could remove the
transaction. An example of this is a consumer that's
reading a stream to add, update, or remove information in a
data warehouse. As long as creation records are treated as
something like an upsert instead of `INSERT` and a deletion
is tolerant if the target doesn't exist, then all
operations can safely be considered to be idempotent.

### Versus Postgres logical replication (#logical-replication)

Postgres aficionados might notice that what we've built
looks pretty similar to [logical replication][logicalrepl]
in Postgres 10, which can similarly guarantee that all
emitted data makes it from producer to consumer.

There are a few advantages to using a stream over logical
replication:

* It's possible to have multiple producers move information
  to a single stream without sharing a database.
* Producers can stream a public representation of data
  instead of one tied to their internal schema. This allows
  producers to change their internal schema without
  breaking consumers.
* You're less likely to leave yourself tied into fairly
  esoteric internal features of Postgres. Understanding how
  to best configure and operate subscriptions and
  replication slots won't be trivial.

### Are delivery guarantees absolute? (#absolute)

Nothing in software is absolute. We've built architecture
based on powerful primitives in system design like ACID
transactions and at-least once delivery semantics, and in
practice, it's likely to be quite robust. But not even a
transaction can protect us from every bug, and eventually
something's going to go wrong enough that these safety
features won't be enough -- for example, the code to stage
stream records in the API might be accidentally removed
through human error. Even if it's noticed and fixed
quickly, some inconsistency will have been introduced into
the system.

Consumers that require absolute precision will need a
secondary mechanism that they can run occasionally to
reconcile their state against canonical sources. In
_Unified Rocket Rides_, we might run a nightly job that
reduces distances across every known ride and emits a tuple
of `(total_distance, last_ride_id)` that consumers can use
to reset their state before continuing to consume the
stream.

## Log-based architecture (#log-architecture)

Log-based architecture provides an effective backbone for
distributed systems by being fast, efficient, and
resilient. Redis streams will provide (when available at
roughly the end of the year) a user-friendly and ubiquitous
log implementation with which to it. Even while Kafka will
continue to be beneficial to the largest of web platforms,
a stack built on Redis and Postgres will serve quite well
right up until that point.

[1] The expectation currently is that streams will be
available in the Redis 4.0 series by the end of the year.

[journalfs]: https://en.wikipedia.org/wiki/Journaling_file_system
[logicalrepl]: https://www.postgresql.org/docs/10/static/logical-replication.html
[mongodurability]: /fragments/mongo-durability
[persistence]: https://redis.io/topics/persistence
[thelog]: https://engineering.linkedin.com/distributed-systems/log-what-every-software-engineer-should-know-about-real-time-datas-unifying
[streams]: http://antirez.com/news/114
[unifiedrides]: https://github.com/brandur/rocket-rides-unified
[wal]: https://www.postgresql.org/docs/current/static/wal.html
