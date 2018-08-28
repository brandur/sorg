---
title: "Postgres 11: Fast Column Creation with Default
  Values"
published_at: 2018-08-27T22:16:00Z
hook: The biggest little feature of Postgres 11.
location: San Francisco
---

Looking over the release notes for the [upcoming Postgres
11][notes], I was pleasantly surprised to see a feature
that I've been hoping would appear roughly forever:

> Many other useful performance improvements, including
> making ALTER TABLE .. ADD COLUMN with a non-null column
> default faster

Despite being buried at the bottom of a list, it's one of
the most important operational improvements that Postgres
has made in years, although it might not be that obvious
why. Let's start with a little background.

## Table alterations and exclusive locks (#alterations)

Consider a simple statement to add a new column to a table:

``` sql
ALTER TABLE users ADD COLUMN credit_balance bigint;
```

Although we're altering the table's schema, any modern
database is sophisticated enough to make this operation
practically instantaneous. Instead of rewriting the
existing representation of the table, information on the
new column is added to the system catalog, which is a cheap
operation. New rows may now be written with values for the
new column, but the system is smart enough to return a
`NULL` for existing rows where no value is available.

But things change when we execute the same statement with
an added `DEFAULT` clause:

``` sql
ALTER TABLE users ADD COLUMN credit_balance bigint NOT NULL DEFAULT 0;
```

It looks similar, but where the previous operation was
trivial, this one requires a full rewrite of the table and
all its indexes so that the new default value can be added
for every existing row.

That's a very expensive operation, but that doesn't mean
it's not efficient -- Postgres will do the work about as
quickly as it's possible to do. On smaller databases, the
operation will still appear to execute instantly.

It's larger installations where it's a problem. Rewriting a
table with lots of existing data will take as long as you'd
expect. In the meantime, that rewrite will have taken an
[`ACCESS EXCLUSIVE` lock][locking] on the table which will
block every other operation until it's released -- even
`SELECT` statements have to wait. In any system with lots
of ongoing access to the table it's not just a problem,
it's totally unacceptable.

TODO: Diagram of conflict

Accidentally locking access to a table when adding a column
was a common pitfall for new Postgres operators because
there's nothing in the SQL to tip them off to the
additional expense of `DEFAULT`. It takes a close reading
of [the manual][altertable] to find out.

## Constraints relaxed by necessity (#constraints)

Because it wasn't possible to cheaply add a `DEFAULT`
column, it also wasn't possible to add a column set to `NOT
NULL`. By definition non nullable columns need to have
values for every row, and it's not possible to add one to a
non-empty table without specifying what values the existing
data should have with a `DEFAULT`.

You could still get a non nullable column by first adding
it as nullable, running a migration to add values to every
existing row, then alter the table with `SET NOT NULL`.
That's still an expensive operation that requires a full
table scan (also `ACCESS EXCLUSIVE`), but it's much faster
than a rewrite.

The amount of effort involved in getting a new non nullable
column in any large relation meant that in practice you
usually wouldn't bother. It was either too time consuming
or too dangerous.

## Why bother with non nullable columns anyway? (#why-bother)

One of the biggest reasons to use relational databases over
document stores and other less sophisticated storage
technology is data integrity. Columns are strongly typed
with the likes of `INT`, `DECIMAL`, or `TIMESTAMPTZ`.
Values are constrained with `NOT NULL`, `VARCHAR` (length),
or [`CHECK` constraints][check]. Foreign key constraints
guarantee [referential integrity][referential].

You can rest assured that your data is in a high quality
state -- the very database is assuring it. This makes it
easier to work with, and prevents bugs caused by data
existing in an unexpected state.

Fast `DEFAULT` values on `ADD COLUMN` fill a gaping hole in
Postgres' operational story. Enthusiasts like me have
always argued in favor of strong data constraints, but knew
also that new non nullable fields weren't usually possible
in any systems running at scale. Postgres 11 will take a
big step forward in addressing that.

## Appendix: How it works (#how-it-works)

The [patch][commit] adds two new fields to
[`pg_attribute`][pgattribute], a system table that tracks
information on every column in the database. One is
`atthasmissing` which is set to `true` when there's a
missing `DEFAULT` value and one is `attmissingval` which
contains its value. As scans are returning rows, they check
these new fields and return missing values where
appropriate. New rows inserted into the table pick up the
default values, and there's no need to check
`atthasmissing` when returning their contents.

TODO: Diagram of table and pg_attribute

The new fields are only used as long as they have to be. If
at any point the table is rewritten, Postgres takes the
opportunity to insert the default value for every row and
unset `atthasmissing` and `attmissingval`.

Note that due to the relative simplicity of
`attmissingval`, this optimization only works for default
values that are _non-volatile_ (i.e., single, constant
values). A default value that calls `NOW()` for example,
can't take advantage of the optimization.

[altertable]: https://www.postgresql.org/docs/10/static/sql-altertable.html
[check]: https://www.postgresql.org/docs/current/static/ddl-constraints.html#DDL-CONSTRAINTS-CHECK-CONSTRAINTS
[commit]: https://github.com/postgres/postgres/commit/16828d5c0273b4fe5f10f42588005f16b415b2d8
[locking]: https://www.postgresql.org/docs/current/static/explicit-locking.html
[notes]: https://www.postgresql.org/docs/11/static/release-11.html
[pgattribute]: https://www.postgresql.org/docs/current/static/catalog-pg-attribute.html
[referential]: https://en.wikipedia.org/wiki/Referential_integrity
