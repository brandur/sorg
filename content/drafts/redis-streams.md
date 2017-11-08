---
title: Redis Streams and the Unified Log
published_at: 2017-11-01T17:24:20Z
location: San Francisco
hook: TODO
---

Years ago, LinkedIn [wrote an article about the unified
log][thelog], a useful architectural pattern for services
in a distributed system converge state with one another. In
the log's design, services emit state changes into an
ordered data structure where each new record gets a unique
ID. Unlike a queue, a log is durable across any number of
reads until it's explicitly truncated.

Consumers track changes in the wider system by consuming
the log. Each one maintains the ID of the last record it
successfully consumed and aims to consume every record at
least once -- no records should be missed. When a consumer
is knocked offline, it looks up the last ID that it
consumed, and continues reading the log from there.

The article is sober enough to point out that this design
is nothing new: we've been using logs in various forms in
computer science for decades. Journaling file systems use
the idea for data correctness. Databases use it with ideas
like the write-ahead log (WAL) in Postgres as they stream
changes to their read replicas.

!fig src="/assets/redis-streams/unified-log.svg" caption="The unified log: a producer emits to the stream and consumers read from it."

Even so, the unified log was a refreshingly novel idea when
the article was written, and still is. File systems and
databases use the structure because it's an effective
pattern, and it lends itself just as well to distributed
architectures. Kafka is more prevalent in 2017, but most of
us are still gluing components together with patches and
duct tape.

Chatty services exchange high-frequency messages back and
forth in a way that's slow (they rely on synchrony),
inefficient (single messages are passed back and forth),
and fragile (every message introduces some possibility of
failure). In contrast, the log is asynchronous, its records
are produced and consumed in batches, and its design builds
in resilience at every turn.

## Redis streams (#redis-streams)

This brings us to Redis. I was happy to hear recently that
the project will soon [1] be shipping with a new data
structure that's a perfect scaffold for a unified log:
[streams][streams]. Unlike a Redis list, records in a
stream are assigned with addressable IDs and are indexed or
sliced with those IDs instead than a relative offset (i.e.
like `0` or `len() - 1`).This lends itself well to having
multiple consumers reading out of a single stream and
tracking their position within it by persisting the ID of
the last record they read.

Records are added to a stream with the new `XADD` command:

```
> XADD rocket-rides-log * id 123 distance 456.7
1506871964177.0
```

A record with `id = 123` and `distance = 456.7` is appended
to the stream `rocket-rides-log`. Redis responds with a
unique ID for the record that's made up of a timestamp and
a sequence number (`.0`) to disambiguate within the
millisecond.

`XRANGE` is the counterpart of `XADD`. It reads a set of
records from a stream:

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

Kafka is a popular system component that also makes a great
backend for a unified log implementation, and once
everything is in place, probably a better one compared to
Redis thanks to its sophisticated design around high
availability and other advanced features.

The most exciting feature of Redis streams isn't their
novelty, but rather than they bring building a unified log
architecture within reach of a small and/or inexpensive
app. Kafka is infamously difficult to configure and get
running, and is expensive to operate once you do. Pricing
for a small Kafka cluster on Heroku costs $100 a month and
climbs steeply from there. It's temping to think you can do
it more cheaply yourself, but after factoring in server and
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

Once you're operating at serious scale, consider switching
to Kafka. In the meantime, Redis streams make a great (and
economic) alternative.

### Configuring Redis for durability (#redis-durability)

One highly desirable property of a unified log is that it's
***durable***, meaning that even if its host crashes or
something terrible happens, it doesn't lose any information
that producers think has been persisted.

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
where it isn't, and run two separate Redis's with different
configuration.

## Unified Rocket Rides (#rocket-rides-unified)

We're going to be returning to the Rocket Rides example
that we talked about while implementing [idempotency
keys](/idempotency-keys). As a quick reminder, Rocket Rides
is a Lyft-like app that lets its users get rides with
pilots wearing jetpacks; a vast improvement over the
every day banality of a car.

As new rides come in, the Unified Rocket Rides API will
emit a new record to the stream that contains the ID of the
ride and the distance traveled. From there, a couple
different consumers will read the stream and keep a running
tally of the total distance traveled for every ride in the
system that's ever been taken.

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

### The API (#api)

The API receives requests over HTTP for new rides from
clients. When it does it creates a ride entry in the local
database, and also emits a record into the unified log to
show that it did.

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
created and add to Postgres. This small indirection is
useful so that in case the request's transaction rolls back
due to a serialization error or other problem, no invalid
data is left in the log. This idea is further expanded on
in [transactionally-staged job drains](/job-drain) which
applies the same idea to background jobs.

The staged records relation in Postgres look like this:

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
them. It runs as a separate process and for better
efficiency, sends records in large batches.

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
removes staged records once they've been confirmed in
Redis. If part of the workflow fails then the process will
run again and reselect the same batch of records a second
time. Those records will be re-emitted into the stream even
if it means that some consumers will see them twice.

Records are emitted with ascending `id`s. It's possible for
a record with a smaller `id` to be emitted after one with a
higher `id`, but _only_ in the case of a double-send. With
the exception of that one manageable caveat, consumers can
always assume that they're receiving `id`s in order.

### Consumers & checkpointing (#consumers)

Consumers pull records out of the log in batches and
consumes them one-by-one. When a batch has been
successfully processed, they set a ***checkpoint*** that
contains the IDs of the last record they consumed to. The
next time a consumer restarts (due to a crash or
otherwise), it reads its last checkpoint and starts
consuming the log from the `id` that it contains.

Checkpoints are modeled as a relation in Postgres:

``` sql
CREATE TABLE checkpoints (
    id            BIGSERIAL PRIMARY KEY,
    name          TEXT      NOT NULL UNIQUE,
    last_redis_id TEXT      NOT NULL,
    last_ride_id  BIGINT    NOT NULL
);
```

Recall that in our simple example, consumers add up the
`distance` of every ride created on the platform. They
track a Redis ID so that they know their location in the
Redis stream.

Total distance is stored to a Postgres relation along with
the name of the consumer:

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
consumes, it skips to the next record.

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
designed to handle these edge cases and will tolerate them
gracefully. By running the simulation for a while, we can
improve our confidence that results will always be correct
and consistent.

Run `forego start` (after following the appropriate setup
in `README.md`) and leave the fleet of processes running.
Despite the occasional double sends and each consumer
failing randomly and independently, no matter how long you
wait, the consumers should always show the same
`total_distance` reading for any given consumed ID.

For example, here's `consumer0` and `consumer1` showing an
identical total for ride ID `521`:

```
consumer0.1 | Consumed record: {"id":521,"distance":539.836923415231}
              total_distance=257721.7m
consumer1.1 | Consumed record: {"id":521,"distance":539.836923415231}
              total_distance=257721.7m
```

## Other considerations (#considerations)

### Non-transactional consumers & idempotency (#non-transaction)

Consumers don't necessarily have to be transactional as
long as the work they do can be applied cleanly given
at-least once semantics.

Notably our example here wouldn't yield correct results
without being nested in a transaction: if it successfully
updated `total_count` but failed to set the checkpoint,
then it would double-count the distance of those records
the next time it tried to consume them.

But if a consumer using input records to execute operations
which are _idempotent_, a wrapping transaction isn't
necessarily needed. An example of this is a consumer that's
reading a stream to add or remove information into a data
warehouse. As long as creation records are treated as
something like an upsert instead of `INSERT` and a deletion
is tolerant if the target doesn't exist, then all
operations can safely be considered to be idempotent.

### Versus Postgres logical replication (#logical-replication)

Postgres aficionados might notice that what we've built
looks pretty similar to [logical replication][logicalrepl]
in Postgres 10, which can similarly guarantee that all
emitted data makes it from producer to consumer.

There are a few advantages to this approach over logical
replication:

* It's possible to have multiple producers move information
  to a single stream without sharing a database.
* Producers can stream a public representation of data
  instead of one tied to their internal schema. This allows
  producers to change their internal schema without
  breaking consumers.
* You're less likely to leave yourself tied into fairly
  esoteric internal features of Postgres.

### Are delivery guarantees absolute? (#absolute)

Nothing in software is absolute. We've built a system based
on powerful primitives like ACID transactions and in
practice, it's likely to be quite robust. But not even a
transaction can protect us against every bug, and
eventually something's going to go wrong enough that these
safety features won't be enough -- the code to stage stream
records in the API might be accidentally removed for
example. Even if it's noticed and fixed quickly, some
inconsistency will have been introduced into the system.

Consumers that require absolute precision will need a
secondary mechanism that they can run occasionally to
reconcile their state against canonical sources. In our
Rocket Rides Unified example, we might run a nightly job
that reduces distances across every known ride and emits a
tuple of `(total_distance, last_ride_id)` that consumers
can use to reset their state before continuing to consume
the stream.

[1] The expectation currently is that streams will be
available in the Redis 4.0 series by the end of the year.

[logicalrepl]: https://www.postgresql.org/docs/10/static/logical-replication.html
[mongodurability]: /fragments/mongo-durability
[persistence]: https://redis.io/topics/persistence
[thelog]: https://engineering.linkedin.com/distributed-systems/log-what-every-software-engineer-should-know-about-real-time-datas-unifying
[streams]: http://antirez.com/news/114
[unifiedrides]: https://github.com/brandur/rocket-rides-unified
