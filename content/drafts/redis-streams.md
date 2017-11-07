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
everything is in place, probably a better one compared to
Redis due to its design around high availability and other
advanced features.

What's exciting about Redis streams is that they bring an
easy way of building a unified log in a small and/or
inexpensive app. Kafka is infamously difficult to configure
and get running, and is expensive to operate once you do.
Pricing for a small Kafka cluster on Heroku costs $100 a
month and climbs steeply from there. It's temping to think
you can do it more cheaply yourself, but after factoring in
server and personnel costs along with the time it takes to
build working expertise in the system, it'll cost more.

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
  # XADD mystream * data <JSON-encoded blob>
  RDB.xadd(STREAM_NAME, "*", "data", JSON.generate(data))
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

Recall that in our simple example, consumers add up the
`distance` of every ride created on the platform. They
track a Redis ID so that they know their location in the
Redis stream.

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

### Non-transactional consumers (#non-transaction)

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

## Other considerations (#considerations)

### Versus logical replication (#logical-replication)

### Are delivery guarantees absolute? (#absolute)

[1] The expectation currently is that streams will be
available in the Redis 4.0 series by the end of the year.

[mongodurability]: /fragments/mongo-durability
[persistence]: https://redis.io/topics/persistence
[thelog]: https://engineering.linkedin.com/distributed-systems/log-what-every-software-engineer-should-know-about-real-time-datas-unifying
[streams]: http://antirez.com/news/114
[unifiedrides]: https://github.com/brandur/rocket-rides-unified
