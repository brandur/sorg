+++
hook = "Operations in online databases are always a tend to be a little tricky to carry out. Renaming a table safely in Postgres isn't too hard, but there is some technique to it."
published_at = 2022-10-27T16:30:29Z
title = "Postgres: Safely renaming a table with no downtime using updatable views"
+++

At first glance, renaming entities in a database seems like it should be easy. The SQL is a dead simple one-liner of `ALTER ... RENAME TO ...`, so what could go wrong?

Well, it turns out a lot actually. Anyone who's run a production database before will recognize that outside of an academic context, it's actually kind of hard. The problem isn't in the database itself, but in database _clients_. Anything that was still running against the old name when a rename takes place will immediately break, causing downtime and major user impact.

An alternative would be to disable all clients temporarily and then do the rename, and indeed a "we're down for maintenance" screen was a pretty common sight in the 2000s, but serious services in the 2020s aim to never have downtime at all. It's annoying for users, and painful if a service is doing some business critical.

Practically speaking, the easiest way to administer a production database is to never rename anything, and live with the fact that some names are less-than-optical. I'd hazard to guess that this is how most shops run -- people would generally prefer to rename as appropriate, but in practice it's more time, risk, and effort than it's worth.

## Copy this code (#code)

But to schema hygiene fanatics out there (like myself), that's not a satisfactory answer. We'd like to rename things, but to also do so with zero downtime and zero user impact. Luckily, Postgres makes this possible relatively easily.

Here's some copy/pastable migration code for renaming a table (in this example, `chainwheel` -> `sprocket`):

``` sql
BEGIN;

ALTER TABLE chainwheel 
    RENAME TO sprocket;

CREATE VIEW chainwheel AS
    SELECT *
    FROM sprocket;
    
COMMIT;
```

And with a post-deploy (after all clients are rotated) follow up of:

``` sql
DROP VIEW chainwheel;
```

## The details (#details)

Why this works:

* `RENAME TO` immediately makes the table available under its new name of `sprocket`.

* To support programs still running against the old name, we create a view with the table's old name of `chainwheel` as a stand-in for the table.

* Postgres supports [updatable views](https://www.postgresql.org/docs/current/sql-createview.html#SQL-CREATEVIEW-UPDATABLE-VIEWS), meaning that with some caveats, views can support `INSERT`, `UPDATE`, and `DELETE` operations that will target their underlying table.

* Postgres supports **transactional DDL** so that the `RENAME TO` and `CREATE VIEW` happen _atomically_. There's never a moment for other database consumers where the table has been renamed but the new view isn't yet available. This is the linchpin feature that makes the change possible with no user impact.

## Proof-of-concept demo (#demo)

I made a [little demo project](https://github.com/brandur/postgres-table-rename-test) that demonstrates the safely of a rename even while a program is running.

We start a tiny app to read and write out of the table with its existing name:

``` sh
$ TABLE_NAME=chainwheel bundle exec ruby app.rb
Inserted and read 100 records of 'chainwheel'
Inserted and read 100 records of 'chainwheel'
Inserted and read 100 records of 'chainwheel'
Inserted and read 100 records of 'chainwheel'
Inserted and read 100 records of 'chainwheel'
Inserted and read 100 records of 'chainwheel'
Inserted and read 100 records of 'chainwheel'
...
```

While it's running, migrate to the new name:

``` sh
bundle exec sequel -m migrations/ -M 2 postgres://localhost:5432/postgres-table-rename-test
```

And notice how the app continues to run happily:

``` sh
...
Inserted and read 100 records of 'chainwheel'
Inserted and read 100 records of 'chainwheel'
Inserted and read 100 records of 'chainwheel'
...
```

We can then start a process that uses the new name:

``` sh
$ TABLE_NAME=sprocket bundle exec ruby app.rb
Inserted and read 100 records of 'sprocket'
Inserted and read 100 records of 'sprocket'
Inserted and read 100 records of 'sprocket'
...
```

In real life, that'd be a new deployment of the original app that's had its table references updated from `chainwheel` to `sprocket`. After it's online and the old processes've been cleared out, we'd run a follow up migration of `DROP VIEW chainwheel`.

## A clean workspace (#clean-workspace)

The demo goes on to show how an updatable view is useful for a rename, but isn't guaranteed to _stay_ updatable. If the schema of the original table is changed to require columns not in the view, like say a new `NOT NULL` column is added, insertions on the view will start to fail.

To avoid production problems make sure to _follow through_ by always taking the rename operation to completion. Do the rename and create the view, deploy and restart code running against the database, and drop the view. Don't leave it hanging around to become an eventual footgun for someone else in the future.
