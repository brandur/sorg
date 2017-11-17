---
title: Using Postgres WAL to Eradicate Stale Reads
published_at: 2017-11-11T23:54:10Z
hook: TODO
---

Read replicas are a common pattern in databases to help
scale workload without having to resort to more complex
strategies like partitioning. Most relational databases
like Postgres, MySQL, and SQL Server are single master
systems, and so all writes have to go the primary. Read
operations however can plausibly be routed to the primary
_or_ any of its replicas [1].

This is useful because it allows an application to start
distributing a considerable amount of its load amongst all
available database nodes. It's especially useful when you
consider that while _most_ write operations have relatively
predictable performance because they're often insert,
updating, or deleting just single records, reads are often
much more elaborate and by extension, expensive for the
database to perform.

TODO: Diagram of writes to primary and reads on replicas.

Even as part of a normal application's workload (barring
analytical queries that can be even more complex), we might
join on two or three different tables in order to perform
an eager load, or even just have to read out a few dozen
rows as a user views a single page. With a suitably
configured application, all of this work can be offloaded
to replicas.

## Stale reads (#stale-reads)

While running reads on replicas is a high-impact and
low-effort win for scalability, it's not without its
challenges. The technique introduces the possibility of
***stale reads*** that occur when an application reads from
replica before it's had a chance to receive information
that's been committed to the primary. To a user this might
look like them updating some information, but then when
trying to view what they updated seeing stale data
representing pre-update state.

TODO: Diagram of stale read.

Stale reads are a race condition. Modern databases
operating over low latency connections are able to keep
replicas following their primary _very_ closely, and may
spend most of their time less than a second out of date,
meaning that even systems using read replicas without any
techniques for mitigating stale reads will probably produce
correct results most of the time.

As software engineers interested in building highly robust
systems, most of the time isn't good enough. We can do
better. Let's take a look at how we can ensure that stale
reads _never_ occur. We'll use Postgres's understanding of
its own state of replication and some in-application
intelligence around connection management to do it.

## The Postgres WAL (#postgres-wal)

In order to come up with a working strategy for avoiding
stale reads, we'll first need to understand a little bit
about how replication works in Postgres.

Postgres commits all changes to a ***WAL*** (write-ahead
log) for durability reasons. Every change is written out as
a new entry in the WAL and it acts the canonical reference
as to whether any change in the system occurred --
committed information is written to a data directory like
you might expect, but is only considered visible if the WAL
confirms that its associated transaction is committed (see
[How Postgres makes transactions
atomic](/postgres-atomicity) for more on this subject).

Changes are written to the WAL one entry at a time and each
one is assigned a ***LSN*** (log sequence number). Changes
are batched in 16 MB ***WAL segments***.

### The WAL's role in replication (#wal-replication)

A Postgres database can dump a representation of its
current state to a base backup which can be used to
initialize replica. From there, the replica can stay in
lockstep with its primary by consuming the changes that it
finds in its emitted WAL. A base backup comes with a
pointer to the current LSN so that when a replica starts to
consume the WAL, it knows where to start.

!fig src="/assets/postgres-reads/replicas-and-wal.svg" caption="A replica being initialized from base backup and consuming its primary's WAL."

There are a few methods available to a replica to consume
WAL from its primary. One is "log shipping" where completed
WAL segments are copied from a primary server to
secondaries and consumed as a single batch. This has the
major advantage of efficiency, but at the cost of how
closely any secondary can be following its primary
(secondaries will be at least as behind as the current
segment that's still being written). Another is streaming,
where WAL is emitted to secondaries over an open
connection. This has the advantage of secondaries being
very up to date at the cost of some extra resource usage.
It also conveys some other advantages like having
secondaries ready to fail over at a moment's notice, and
allowing secondaries to keep their primary appraised of
their progress (hot standby feedback).

Due to their respective capabilities in being ready to
become a primary, replicas consuming WAL with log shipping
are also known as "warm standbys" while those using
streaming are called "hot standbys". The latter is often
seen in production setups due to its very nice property of
being ready to take the reins at a moment's notice.  The
technique we're going to discuss will be able to make
replica reads more often when WAL is being streamed, but
will work with either method.

## Routing reads based on replica WAL position (#routing-reads)

We can avoid stale reads by making sure to route read
queries only to replicas that are caught up enough to
accurately fulfill them. To do this, we'll need a way of
measuring how far behind a replica is, and the WAL's LSN
makes for a very convenient metric by which to measure
this.

When mutating a resource in the system we'll store the
latest LSN for the entity making the request. When
fulfilling a read operation for that same entity, we'll
check which replicas have consumed to that point or beyond
it, and randomly select one from the pool. If no replicas
are sufficiently up to date (i.e. say a read operation is
being run very closely after the initial write), we'll fall
back to the master. Using this technique, stale reads
become impossible regardless of the state of any given
replica.

!fig src="/assets/postgres-reads/routing.svg" caption="Routing read operations based on replica progress in the WAL."

### Scalable Rocket Rides (#rocket-rides)

To build a working demonstrating of this concept we'll be
returning to the same toy application that we used to show
off an implementation for [idempotency
keys](/idempotency-keys) and [the unified
log](/redis-streams) -- _Rocket Rides_. As a quick
reminder, Rocket Rides is a Lyft-like app that lets its
users get rides with pilots wearing jetpacks; a vast
improvement over the everyday banality of a car.

_Scalable Rocket Rides_ has an `api` process that writes to
a Postgres database. It's configured with a number of read
replicas that receive changes with the WAL. When performing
a read, it tries to route it to one of a random replica
that's sufficiently caught up to fulfill the operation for
a particular user.

We'll be using the Sequel gem, which can be configured with
a primary and any number of different replicas which are
assigned names (e.g. `replica0`) and address with the
`server(...)` method:

``` ruby
DB = Sequel.connect("postgres://localhost:5433/rocket-rides-scalale",
  servers: {
    replica0: { port: 5434 },
    replica1: { port: 5435 },
    replica2: { port: 5436 },
    ...
  }

# routes to primary
DB[:users].update(...)

# routes to replica0
DB[:users].server(:replica0).select(...)
```

A working version of all this code is available in the
[_Scalable Rocket Rides_][scalablerides] repository. We'll
walk through the project with a number of extracted
snippets, but it might easier to download the code and
follow along that way:

``` sh
git clone https://github.com/brandur/rocket-rides-scalable.git
```

### Bootstrapping a cluster (#cluster)

For demo purposes it's useful to create a small cluster
with a primary and a number of read replicas, and the
project [includes a small script to help do
so][createcluster]. It initializes and starts a primary,
and for a number of times equal to the `NUM_REPLICAS`
environment variable performs a base backup and starts a
replica with it

Processes are started as children of the script with Ruby's
`Process.spawn`, and all Postgres daemons will shut down
when it's stopped. The setup's designed to be ephemeral and
any data added to the primary is removed when the cluster
bootstraps itself again on the script's next run.

### The Observer: tracking replication status (#observer)

To save every `api` process from having to reach out and
check on the status of every replica for itself, we'll have
a single process called an `observer` that periodically
refreshes the state of every replica and stores it to a
Postgres table.

The table contains a common `name` for each replica (e.g.
`replica0`) and a `last_lsn` field that stores a sequence
number as Postgres's native `pg_lsn` data type:

``` sql
CREATE TABLE replica_statuses (
    id       BIGSERIAL    PRIMARY KEY,
    last_lsn PG_LSN       NOT NULL,
    name     VARCHAR(100) NOT NULL UNIQUE
);
```

Keep in mind that this status information could really go
anywhere. If we have Redis available, we could put it in
there for fast access, or have every `api` worker cache it
in-process periodically for even faster access. Postgres is
convenient, and as we'll see momentarily makes lookups
quite elegant, but it's not necessary.

The `observer` runs in a loop, and on every iteration
executes this:

``` ruby
# exclude :default at the zero index
replica_names = DB.servers[1..-1]

last_lsns = replica_names.map do |name|
  DB.with_server(name) do
    DB[Sequel.lit(<<~eos)].first[:lsn]
      SELECT pg_last_wal_replay_lsn() AS lsn;
    eos
  end
end

insert_tuples = []
replica_names.each_with_index do |name, i|
  insert_tuples << { name: name.to_s, last_lsn: last_lsns[i] }
end

# update all replica statuses at once with upsert
DB[:replica_statuses].
  insert_conflict(target: :name, update: { last_lsn: Sequel[:excluded][:last_lsn] }).
  multi_insert(insert_tuples)

$stdout.puts "Updated replica LSNs: results=#{insert_tuples}"
```

A connection is made to every replica in sequence and
`pg_last_wal_replay_lsn()` is used to see its current
location in the WAL. When all statuses are available, we
use Postgres upsert (`INSERT INTO ... ON CONFLICT ...`) to
store the entire set to `replica_statuses`.

### Saving minimum LSN (#min-lsn)

Knowing the status of our replicas is half the equation,
with the other half being knowing the minimum required
replication progress for any given user so that they never
see a stale read. This is determined by saving the
primary's current LSN whenever the user performs an action.

We'll model this as a `min_lsn` field on our `users`
relation:

``` sql
CREATE TABLE users (
    id      BIGSERIAL    PRIMARY KEY,
    email   VARCHAR(255) NOT NULL UNIQUE,
    min_lsn PG_LSN
);
```

For any action that will later affect reads, we touch the
user's `min_lsn` by setting it to the value of the
primary's `pg_current_wal_lsn()`. This is performed in
`update_user_min_lsn` in this simple implementation:

``` ruby
post "/rides" do
  user = authenticate_user(request)
  params = validate_params(request)

  DB.transaction(isolation: :serializable) do
    ride = Ride.create(
      distance: params["distance"],
      user_id: user.id,
    )
    update_user_min_lsn(user)

    [201, JSON.generate(serialize_ride(ride))]
  end
end

def update_user_min_lsn(user)
  User.
    where(id: user.id).
    update(Sequel.lit("min_lsn = pg_current_wal_lsn()"))
end
```

### Selecting a replica (#select-replica)

Now that replication status and minimum WAL progress for
every user is being tracked, `api` processes need a way to
select an eligible replica candidate for read operations.
Here's an implementation of a `select_replica` method that
does just that:

``` ruby
def select_replica(user)
  # If the user's `min_lsn` is `NULL` then they haven't performed an operation
  # yet, and we don't yet know if we can use a replica yet. Default to the
  # primary.
  return :default if user.min_lsn.nil?

  # exclude :default at the zero index
  replica_names = DB.servers[1..-1].map { |name| name.to_s }

  res = DB[Sequel.lit(<<~eos), replica_names, user.min_lsn]
    SELECT name
    FROM replica_statuses
    WHERE name IN ?
      AND pg_wal_lsn_diff(last_lsn, ?) >= 0;
  eos

  # If no candidates are caught up enough, then go to the primary.
  return :default if res.nil? || res.empty?

  # Return a random replica name from amongst the candidates.
  candidate_names = res.map { |res| res[:name].to_sym }
  candidate_names.sample
end
```

`pg_wal_lsn_diff()` returns the difference between two
`pg_lsn` values, and we use it to compare the stored status
of each replica in `replica_statuses` to the `min_lsn`
value of the current user (`>= 0` means that the replica is
ahead of the user's minimum). We take the name of a random
replica from the returned set. If the set was empty, then
no replica is advanced enough for our purposes, so we fall
back to the primary.

Here's `select_replica` in action on an API endpoint:

``` ruby
get "/rides/:id" do |id|
  user = authenticate_user(request)

  name = select_replica(user)
  $stdout.puts "Reading ride #{id} from server '#{name}'"

  ride = Ride.server(name).first(id: id)
  if ride.nil?
    halt 404, JSON.generate(wrap_error(
      Messages.error_not_found(object: "ride", id: id)
    ))
  end

  [200, JSON.generate(serialize_ride(ride))]
end
```

And that's it! The repository also comes with a simulator
that creates a new ride and then immediately tries to read
it. Running the whole constellation of programs will show
that most of the time these reads will be served from a
replica, but occasionally from the primary (`default` in
Sequel) as replication falls behind or the `observer`
hasn't performed its work loop in some time:

```
$ forego start | grep 'Reading ride'
api.1       | Reading ride 96 from server 'replica0'
api.1       | Reading ride 97 from server 'replica0'
api.1       | Reading ride 98 from server 'replica0'
api.1       | Reading ride 99 from server 'replica1'
api.1       | Reading ride 100 from server 'replica4'
api.1       | Reading ride 101 from server 'replica2'
api.1       | Reading ride 102 from server 'replica0'
api.1       | Reading ride 103 from server 'default'
api.1       | Reading ride 104 from server 'default'
api.1       | Reading ride 105 from server 'replica2'
```

## Should I do this? (#should-i)

[1] A note on terminology: I use the word "replica" to
refer to a server that's tracking changes on a primary.
Common synonyms for the word include "standby", "slave",
and "secondary", but I'll stick to "replica" for
consistency.

[createcluster]: https://github.com/brandur/rocket-rides-scalable/tree/master/scripts/create_cluster
[scalablerides]: https://github.com/brandur/rocket-rides-scalable
