---
title: Scaling Postgres with Read Replicas & Using WAL to
  Counter Stale Reads
location: Osaka
published_at: 2017-11-17T22:02:56Z
hook: Scaling out operation with read replicas and avoiding
  the downside of stale reads by observing replication
  progress.
tags: ["postgres"]
hn_link: https://news.ycombinator.com/item?id=15726376
---

A common technique when running applications powered by
relational databases like Postgres, MySQL, and SQL Server
is offloading read operations to readonly replicas [1],
helping to distribute load between more nodes in the
system by re-routing queries that don't need to run on the
primary. These databases are traditionally single master,
so writes have to go to the primary that's leading the
cluster, but reads can go to any replica as long as it's
reasonably current.

Spreading load across more servers is good, and the pattern
shows itself to be even more useful when considering that
although write operations might be numerous, most of them
have predictable performance -- they're often inserting,
updating, or deleting just a single record. Reads on the
other hand are often more elaborate, and by extension, more
expensive.

!fig src="/assets/postgres-reads/replica-reads.svg" caption="Writes on the primary and reads on its replicas."

Even as part of a normal application's workload (barring
analytical queries that can be even more complex), we might
join on two or three different tables in order to perform
an eager load, or even just have to read out a few dozen
rows to accurately render a response. A mature application
might execute hundreds of queries to fulfill even a single
request, and farming these out to replicas would yield huge
benefits in reducing pressure on the primary.

## A complication: stale reads (#stale-reads)

Running reads on replicas is a pretty good high-impact and
low-effort win for scalability, but it's not without its
challenges. The technique introduces the possibility of
***stale reads*** that occur when an application reads from
replica before that replica has received relevant updates
that have been committed to the primary. A user might
update some key details, and then go to view their changes
and see stale data representing the pre-update state.

!fig src="/assets/postgres-reads/stale-read.svg" caption="A stale read that went to a replica that hadn't yet applied changes from the primary."

Stale reads are a race condition. Modern databases
operating over low latency connections can keep replicas
trailing their primary _very_ closely, and probably spend
most of their time less than a second out of date. Even
systems using read replicas without any techniques for
mitigating stale reads will produce correct results most of
the time.

But as software engineers interested in building
bulletproof systems, "most of the time" isn't good enough,
and we can do better. Let's take a look at a technique to
make sure that stale reads _never_ occur. We'll use
Postgres's own understanding of its replication state and
some in-application intelligence around connection
management to accomplish it.

## The Postgres WAL (#postgres-wal)

First, we're going to have to understand a little bit about
how replication works in Postgres.

Postgres commits all changes to a ***WAL*** (write-ahead
log) for durability reasons. Every change is written out as
a new entry in the WAL and it acts the canonical reference
as to whether any change in the system occurred --
committed information is written to a data directory like
you might expect, but is only considered visible to new
transactions if the WAL confirms that it's committed (see
[How Postgres makes transactions
atomic](/postgres-atomicity) for more on this subject).

Changes are written to the WAL one entry at a time and each
one is assigned a ***LSN*** (log sequence number). Changes
are batched in 16 MB ***WAL segments***.

### The WAL's role in replication (#wal-replication)

A Postgres database can dump a representation of its
current state to a *base backup* which can be used to
initialize replica. From there, the replica stays in
lockstep with its primary by consuming changes in its
emitted WAL. A base backup comes with a pointer to the
current LSN so that when a replica starts to consume the
WAL, it knows where to start.

!fig src="/assets/postgres-reads/replicas-and-wal.svg" caption="A replica being initialized from base backup and consuming its primary's WAL."

There are a few ways for a replica to consume WAL. The
first is "log shipping": completed WAL segments (16 MB
chunks of the WAL) are copied from primary to replicas and
consumed as a single batch. This has the major advantage of
efficiency (it's fast to copy files around, and has
negligible cost to the primary), but with a tradeoff of how
closely any secondary can be following its primary --
secondaries will be at least as behind as the current
segment that's still being written.

Another common configuration for consuming WAL is
"streaming", where WAL is emitted by the primary to
replicas over an open connection. This has the advantage of
secondaries being very current at the cost of some extra
resource consumption.

Based on their respective aptitude's for becoming primary
at a moment's notice, replicas consuming WAL with log
shipping are also known as "warm standbys" while those
using streaming are called "hot standbys". Hot standbys are
often seen in production setups because maintain state that
closely matches their primary and make great targets to
fail over to at a moment's notice. The technique we're
going to discuss works better with streaming, but should
yield at benefits with either method.

## Routing reads based on replica WAL position (#routing-reads)

By routing read operations only to replicas that are caught
up enough to run them accurately, we can eliminate stale
reads. This necessitates an easy way of measuring how far
behind a replica is, and the WAL's LSN is perfect for this
use.

When mutating a resource in the system we'll store the
last committed LSN for the entity making the request. Then,
when we subsequently want to fulfill a read operation for
that same entity, we'll check which replicas have consumed
to that point or beyond it, and randomly select one from
the pool. If no replicas are sufficiently advanced (i.e.
say a read operation is being run very closely after the
initial write), we'll fall back to the master. Stale reads
become impossible regardless of the state of any given
replica.

!fig src="/assets/postgres-reads/routing.svg" caption="Routing read operations based on replica progress in the WAL."

The technique is inspired by [GitLab's article on scaling
their database][gitlab], where they refer to it as "sticky
connections". Their large Postgres installation is still
unpartitioned, and using replicas for extra read capacity
is key in managing its considerable load.

### Scalable Rocket Rides (#rocket-rides)

To build a working demo we'll be returning to the same toy
application that we used to show off an implementation for
[idempotency keys](/idempotency-keys) and [the unified
log](/redis-streams) -- _Rocket Rides_. As a quick
reminder, _Rocket Rides_ is a Lyft-like app that lets its
users get rides with pilots wearing jetpacks; a vast
improvement over the everyday banality of a car.

Our new _Scalable Rocket Rides_ demo has an `api` process
that writes to a Postgres database. It's configured with a
number of read replicas that are configured with Postgres
replication to receive changes from the primary. When
performing a read, the `api` tries to route it to one of a
random replica that's sufficiently caught up to fulfill the
operation for a particular user.

We'll be using the Sequel gem, which can be configured with
a primary and any number of read replicas. Replicas are
assigned names like `replica0`, and operations are sent to
them with the `server(...)` helper:

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
snippets, but if you prefer, you can download the code and
follow along:

``` sh
git clone https://github.com/brandur/rocket-rides-scalable.git
```

### Bootstrapping a cluster (#cluster)

For demo purposes it's useful to create a small
locally-running cluster with a primary and some replicas.
The project [includes a small script to help with
that][createcluster]. It initializes and starts a primary,
and for a number of times equal to the `NUM_REPLICAS`
environment variable performs a base backup and boots a
replica with it

Postgres daemons are started as children of the script with
Ruby's `Process.spawn` and will all die when it's stopped.
The setup's designed to be ephemeral and any data added to
the primary is removed when the cluster bootstraps itself
again on the script's next run.

### The Observer: tracking replication status (#observer)

To save every `api` process from having to reach out and
check on the replication status of every replica for
itself, we'll have a process called an `observer` that
periodically refreshes the state of every replica and
stores it to a Postgres table.

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
convenient, and as we'll see momentarily, makes lookups
quite elegant, but it's not necessary.

The `observer` runs in a loop, and executes something like
this on every iteration:

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
  insert_conflict(target: :name,
    update: { last_lsn: Sequel[:excluded][:last_lsn] }).
  multi_insert(insert_tuples)

$stdout.puts "Updated replica LSNs: results=#{insert_tuples}"
```

A connection is made to every replica and
`pg_last_wal_replay_lsn()` is used to see its current
location in the WAL. When all statuses have been collected,
Postgres upsert (`INSERT INTO ... ON CONFLICT ...`) is used
to store the entire set to `replica_statuses`.

### Saving minimum LSN (#min-lsn)

Knowing the status of our replicas is half of the
implementation. The other half is knowing the minimum
replication progress for every user that will give us the
horizon beyond which stale reads are impossible. This is
determined by saving the primary's current LSN whenever the
user makes a change in the system.

We'll model this as a `min_lsn` field on our `users`
relation (and again use the built-in `pg_lsn` data type):

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

### Selecting an eligible replica (#select-replica)

Now that replication status and minimum WAL progress for
every user is being tracked, `api` processes need a way to
select an eligible replica candidate for read operations.
Here's an implementation that does just that:

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
it. Running the constellation of programs will show that
most of the time these reads will be served from a replica,
but occasionally from the primary (`default` in Sequel) as
replication falls behind or the `observer` hasn't performed
its work loop in a while:

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

Maybe. The implementation's major downside is that each
user's `min_lsn` needs to be updated every time an action
that affects read results is performed. If you squint just
a little bit, you'll notice that this looks a lot like
cache invalidation -- a technique infamous for working well
until it doesn't. In a more complex codebase save hooks and
update triggers can be useful in helping to ensure
correctness, but given enough lines of code and enough
people working on it, _perfect_ correctness can be
frustratingly elusive.

Projects that produce only moderate database load (the
majority of all projects) shouldn't bother, and keep their
implementations simple by running everything against the
primary. Projects that need infinitely scalable storage
(i.e. disk usage is expected to grow well beyond what a
single node can handle) should probably look into a more
elaborate partitioning scheme ([like Citus][citus]).

There is a sweet spot of projects that can keep their
storage within a single node, but still want to scale out
on computation. For this sort of use moving reads to
replicas can be quite beneficial because it greatly expands
the runway for scalability while also avoiding the
considerable overhead and operational complexity of
partitioning.

[1] A note on terminology: I use the word "replica" to
refer to a server that's tracking changes on a "primary"
(A.K.A. "leader", "master"). Common synonyms for a replica
include "standby", "slave", and "secondary", but I'll stick
to "replica" for consistency.

[citus]: https://www.citusdata.com/
[createcluster]: https://github.com/brandur/rocket-rides-scalable/tree/master/scripts/create_cluster
[gitlab]: https://about.gitlab.com/2017/10/02/scaling-the-gitlab-database/#sticky-connections
[scalablerides]: https://github.com/brandur/rocket-rides-scalable
