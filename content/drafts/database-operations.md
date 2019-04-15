---
title: Database Operations
published_at: 2019-03-26T17:00:18Z
location: San Francisco
hook: TODO
tags: ["postgres"]
---

We build database bootstraps:

* Slow query log.
* Slow query killer.
* Index builder.
* Index verifier: Tries to make sure that indexes are built
  in production before the code that uses them is deployed.
* Migration runner: Handles running migrations in larger
  development environment where many developers may want to
  execute them at the same time.
* Migration verifier: Verify that DDL in a migration won't
  be dangerous if run in production. For example, before
  Postgres 11 [adding a new non-nullable column would take
  an exclusive table lock](/postgres-default).
* Data migration runner: Moves data around at a controlled
  pace so as not to introduce too much load on the database
  which might affect other operations. A common pattern for
  this is to move data in small batches and throttle those
  operations.
* Statistics views: Easy interfaces to view the size of
  various tables and their indices without having to run
  commands against them.
* Analytics without affected affecting database health:
  Dataclips, etc.

* HA failover.
