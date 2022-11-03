+++
hook = "Keeping high-volume and high-throughput data out of the main database for fewer operational headaches and faster recovery."
published_at = 2022-11-02T16:48:23Z
title = "Ephemeral DB, a sacrificial database line for high-throughput data"
+++

We tried a new operational technique recently that we're calling "ephemeral DB". It's pretty simple: break our service's data up into two separate databases, with a standard "main" database that stores most of everything, and is the default go-to for most new additions, but with a second database, the ephemeral DB, which stores data that's higher volume or which changes frequently, and in which in case of emergency, we can afford to lose.

Keeping high throughput data colocated with everything else has sizable downsides that can go unnoticed until something goes wrong:

* Big tables take more work to query, leading to more costly queries that run longer, and which will contend with other queries in the system.

* Longer running queries also mean more tuples and tuple versions retained, leading to bloat, and a [higher likelihood of problems elsewhere](/postgres-queues). This is a side-effect of Postgres' MVCC design -- records must be retained until the last transaction that could possibly see them has finished running.

* It may exhaust the system's available I/O. Again, contending with more important work.

* The database will take longer to restore because not only are base backups bigger, there's also more [WAL](https://www.postgresql.org/docs/current/wal-intro.html) to apply, stemming from the frequent inserts/updates.

* In case of an HA failover, a new HA standby will take longer to come online for the same reason, leading to higher risk of an outage in the unlucky case that another HA failover has to occur.

To date we've managed to keep the main database very small, with the whole thing occupying less than 150 MB. But the topic had come up of adding higher volume data (in this case, metrics), and the question emerged as to whether this should go in the existing database or a separate one.

We polled a number of our internal engineers with a lot of database experience, and the unanimity of the response was shocking: put it in new database. Rationale varied somewhat from person to person, but was all along the lines of the points listed above. And thus, the ephemeral DB was born.

Our application code opens two database pools when it comes online, and is written so that each service can start a transaction on either one:

``` go
type BaseService struct {
	// Begin begins a database transaction. We send this
	// in instead of a full connection or connection pool
	// to encourage use of transactions in service handlers.
	Begin func(context.Context) (*db.Transaction, error)

	// BeginEphemeral is similar to Begin, but begins a
	// transaction on the ephemeral database line.
	BeginEphemeral func(context.Context) (*db.Transaction, error)
}
```

There are of course some tradeoffs:

* Two databases are harder to look after than one.

* No foreign keys between databases. You can still have IDs in one that reference the other, but you're in Mongo world in the sense that the existence of the object they reference is no longer guaranteed by the data layer.

* Development complexity is increased somewhat in that we now have two separate migration lines, and engineers have to think about what they're doing when adding changes to one or the other. This doesn't come out to be too important though because it can be almost entirely solved by a little investment in tooling (we have a single `make db` command that'll get your local databases back to current state regardless of where they started).

The loose set of decision rules we came up with as to when to put a new table in the ephemeral DB instead of the main DB are (1) it should be high volume, and (2) it should be data that we can afford to lose (e.g. metrics). In case of a major production problem wherein recovery isn't going well, we'd be willing to jettison the ephemeral DB completely and spin up a brand new one to replace it. Currently, relations number about 50:1 in favor of being in the main database, so the ephemeral DB is only used sparingly.

We still have [a single-dependency stack](/fragments/single-dependency-stacks) in that Postgres is the only persistence layer we use (even if the addition of a second database complicates the equation). So in many ways the ephemeral DB will be fulfilling the role that a Redis might elsewhere [1].

I'm fairly confident in saying that having had something similar would have saved us from countless operational problems that we had with Heroku's API, the vast majority of which were of the shape of "analytical/billing query runs long churning through vast volumes of data, leading to more dead tuples and WAL, and thus bleeding into everything else and causing cascading failure". Having a second database would've meant that even if we'd had the same problems, those could've been isolated, and only a small part of the total API functionality would've been degraded instead of knocking the whole thing offline.

[1] With the important caveat that operationally, Postgres is much less default safe compared to Redis, so the ephemeral DB will be fulfilling some Redis functions, but _carefully_.
