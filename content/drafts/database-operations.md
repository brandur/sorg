+++
hook = "TODO"
location = "San Francisco"
published_at = 2019-03-26T17:00:18Z
tags = ["postgres"]
title = "Database Scaffolds: Ideas for more operable databases"
+++

There are two phases in starting to use a new database. The first phase is learning it, writing code that will use it, and getting that initial deployment out into production. It lasts somewhere on the order of weeks to months.

The next phase is much more difficult. It involves learning to operate the database as use and load continues to ramp up. Mistakes are made early on, and engineers learn from them. Principles of operation that aren’t obvious from reading documentation start to crystallize. Hard operational maneuvers that aren’t necessary in development are undertaken, like upgrading across major versions or recovering from major failures. Undocumented quirks are discovered. This phase lasts the rest of the organization’s lifetime.

As they scale, organizations end up building many tools to help with the above. For example, managing incremental changes to a database’s schema is such a common problem that practically every ORM in existence has a migration framework to help with it. But the relative ubiquity of high quality migrations are an unusual success case — because these facilities generally aren’t available within databases themselves and high quality open source solutions are often not available, they much more usually end up as rough, internal projects that get the job done, but not particularly well.

I’m going to refer to the collection of this sort of tooling as a ***database bolts*** — features that are not built into databases, but could be. Personally, I’d love to see more of them migrate down into databases so that all of us can benefit from shared tools of higher quality instead of duplicating work on low quality ones everywhere. This article is a list of some that come to mind.

## Queries (#queries)

### Slow query killer (#slow-query-killer)

A common operational problem is for long-running queries to cause trouble of various sorts in production ([for example](/Postgres-queues)) as the extra accounting the system has to do to support them affects current work. Databases like Postgres offer mechanisms that try to remediate this like `statement_timeout`, but they’re rarely sufficient — it’s still too easy for rogue users or services to omit, adjust, or disable their own timeouts.

The common remediation is to build a supervisor that monitors query runtime and kills any that have been going too long. It’s not that hard of a thing to do, but it’s so pervasive of a pattern that it’d be nice if databases provided it out of the box.

### Non-destructive analytics (#analytics)

Common OLTP operations are usually very fast, but analytics tend to be slow. Still, they’re so important that people want to do them anyway.

Because long-running queries can be so harmful to operational health, organizations often resort to exporting data from their main database into a separate data warehouse, which is expensive and inefficient (for a smaller organization at least).

A useful feature that a database could provide is a way to run queries that would be guaranteed not to affect production operations either because the same level of bookkeeping isn’t performed or them, they’re deniced to the point that they’re paused given too much load, or through some other facility — like a follower node that won’t affect the operation of its primary (as is common through the use of `hot_standby_feedback` in Postgres for example), even if it would mean falling behind in replication and having to self-sacrifice by detaching from the log.

### Query statistics (#query-statistics)

A long-standing dream of mine would be for queries to return statistics like their overall efficiency, whether they were able to take advantage of an index, and how much in-memory work they had too do in-band to the calling application alongside their results.

Today, applications generally have to resort to an out-of-band slow query log to find queries that perform poorly, but imagine if they were instead able to get immediate feedback on that and fix problems sooner. An application running in the test suite could detect an inefficient query,. and immediately fail without the underperforming query ever reaching production and putting its service at risk.

#### Locking (#locking)

In-band query statistics could be expanded further to include information on what kind of locks a query had to acquire during the course of an operation. It’s often not obvious by reading SQL when a query needs a full table lock versus when it doesn’t (information that’s normally acquired either through reading documentation, or burning your hand on the stove), and the database flagging problematic locks to applications talking to them would be a huge advancement.

### Syntax analyzer and banhammer (#syntax-analyzer)

At some point applications probably want to put bans on certain types of syntax. For example in Postgres, it rarely makes sense to run `CREATE INDEX` without the `CONCURRENTLY` modifier. `SELECT` clauses that don’t have an equality predicate on a unique field (like `WHERE id = 123`) should probably always have a `LIMIT` clause in an OLTP application.

## Indexes (#indexes)

### Index builder (#index-builder)

You can get away with building indexes with the standard migration system for quite some time, but this gets difficult when data gets so big that index builds can take hours, days, or weeks.

It’s common to eventually build a system that monitors any indexes that need to be built or dropped, and makes them easier to work with. For example, making sure builds are staggered so there aren’t too many running at once, or notifying their owners when a build is complete.

#### Slow index builder (#slow-index-builder)

Even better is an index builder that guarantees that builds don’t put too heavy of a load on the database so that they never affect other production operations. Timely new indexes can sometimes be very important, but they should almost always take a backseat to the system’s health.

### Index verifier (#index-verifier)

An easy mistake to make is to send a migration that raises an index out with the same code that relies on that index being present to be performant, and that can lead to a production catastrophe as the index takes time to build and those queries slow to a crawl.

A countermeasure is a check that analyzes query cost based on the current production schema and tries to ensure that it can be performant, and which runs in CI so that bad code is kept out of `master`.

## Migrations (#migrations)

### Migration verifier (#migration-verifier)

Deploying code that relies on new schema is generally a two-phase process: (1) deploying and running schema alterations, and (2) deploying the code. Like the index verifier above, a migration verifier makes sure that code isn’t merged to `master` which relies on schema added by a migration that hasn’t yet run.

The migration verifier also checks the opposite conditions: fields should never be dropped while production code is still using them. This can get especially tricky if multiple services are sharing a database (you shouldn’t do this, but it’s not uncommon) because the state of all of their code will need to be checked.

### Migration cost analyzer and banhammer (#migration-cost-analyzer)

A big problem with SQL/DDL is that it’s not always obvious by looking at syntax how expensive a query will be to run. For example, before Postgres 11, [adding a new non-nullable column would take an exclusive table lock](/postgres-default), so running something like an `ALTER TABLE foo ADD COLUMN bar text NOT NULL` would block all other operations until it finished, which for a large table could be a long time. Accidentally trying that on a hot table would spell near-certain disaster.

A useful service for a database scaffold to provide is one which looks for DDL that contains knowingly dangerous operations, runs in CI, and bans any that it finds from coming into `master`.

### Migration runner (#migration-runner)

At smaller scales, everyone can run their own migrations to alter database schema and it’s not a big deal, but eventually systems get big enough that schema alterations need to be going out on a constant basis.

It’s convenient to have these schema alterations happen automatically. A service looks for migrations in `master` that still need to be run, and runs each one. Failed migrations are removed from the queue and handed back to their authors for mitigation.

### Data migration runner (#data-migration-runner)

**Data migrations** are slightly distinct in that they run some operation on data rather than just alter schema. Because they operate on existing data which may be quite large, they can take a long time and put a lot of load on the database as they’re running.

A common pattern to make sure they don’t overload the database is to move data manipulation into smaller batches or even single-row updates and controlling the speed at which those are executed (so think `UPDATE foo SET … WHERE id IN (?)` instead of `UPDATE foo SET …` without a predicate). At Stripe for example, we use a [token bucket](/rate-limiting) implementation to control data migration speed precisely.

## Command and control (#command-and-control)

### Statistics views (#statistics-views)

[Postgres recommends](https://wiki.postgresql.org/wiki/Disk_Usage) the following query to look at the size of all tables in a human-readable way:

``` sql
SELECT *, pg_size_pretty(total_bytes) AS total
    , pg_size_pretty(index_bytes) AS INDEX
    , pg_size_pretty(toast_bytes) AS toast
    , pg_size_pretty(table_bytes) AS TABLE
  FROM (
  SELECT *, total_bytes-index_bytes-COALESCE(toast_bytes,0) AS table_bytes FROM (
      SELECT c.oid,nspname AS table_schema, relname AS TABLE_NAME
              , c.reltuples AS row_estimate
              , pg_total_relation_size(c.oid) AS total_bytes
              , pg_indexes_size(c.oid) AS index_bytes
              , pg_total_relation_size(reltoastrelid) AS toast_bytes
          FROM pg_class c
          LEFT JOIN pg_namespace n ON n.oid = c.relnamespace
          WHERE relkind = 'r'
  ) a
) a;
```
That’s not fast or ergonomic. Because size information is important, it’s common to build an interface that allows it to be referenced quickly, and which provides a further detailed breakdown of each table, like having the size of each index inline.

### HA failover (#ha-failover)

This one’s obvious. It gets baked into every hosted solution (AWS RDS, Digital Ocean Managed Databases, Heroku Postgres, GCP Databases), but it’d be nice if the wheel didn’t have to be reinvented for every new service that appears. Postgres has had some progress on this in the form of [`pg_auto_failover`](https://github.com/citusdata/pg_auto_failover).

## Complete scaffolding (#complete-scaffolding)

Not everything in this list belongs in the core of a database, but the length of it should help to demonstrate that a database itself is often only part of the answer — it’s not in itself a safe abstraction and usually requires large augmentations to make it fully operable.

A lot of work is duplicated because many organizations have either built, or are in the process of building, the scaffolding above. Many of them will end up with something that works well enough, but bespoke internal code tends to be under-designed, not as solidly built, and less robust compared to general-purpose and well-maintained public solutions. So again, while not all this scaffolding belongs in the core, there’s probably room for more of it.
