---
title: "A Missing Link in Postgres 11: Fast Column Creation with Defaults"
published_at: 2018-08-27T22:16:00Z
hook: How a seemingly minor enhancement in Postgres 11
  fills one of the system's biggest operational holes.
location: San Francisco
---

If you read through the release notes for [upcoming
Postgres 11][notes], you might find a somewhat
inconspicuous feature tucked away at the bottom of the
enhancements list:

> Many other useful performance improvements, including
> making `ALTER TABLE .. ADD COLUMN` with a non-null column
> default faster

It's not a flagship feature of the new release, but even
so, it's one of the most important operational improvements
that Postgres has made in years, even though it might not
be immediately obvious why. The short version is that it's
eliminated a limitation that used to make correctness in
schema design difficult, but let's take a look at the
details.

## Alterations and exclusive locks (#alterations)

Consider for a moment one of the simplest database
statements possible, one that adds a new column to a table:

``` sql
ALTER TABLE users
    ADD COLUMN credits bigint;
```

Although it's altering the table's schema, any modern
database is sophisticated enough to make this operation
practically instantaneous. Instead of rewriting the
existing representation of the table (thereby forcing all
existing data to be copied over at great expense),
information on the new column is added to the system
catalog, which is cheap. That allows new rows to be written
with values for the new column, and the system is smart
enough to return `NULL` for current rows where no value
previously existed.

But things get complicated when we add a `DEFAULT` clause
to the same statement:

``` sql
ALTER TABLE users
    ADD COLUMN credits bigint NOT NULL DEFAULT 0;
```

The SQL looks so similar as to be almost identical, but
where the previous operation was trivial, this one requires
a full rewrite of the table and all its indexes. Because
there's now a non-null value involved, the database ensures
data integrity by going back and injecting it for every
existing row, which is very expensive work.

Despite that expense, Postgres is still capable of doing
the rewrite efficiently, and on smaller databases it'll
appear to happen instantly.

It's bigger installations where it becomes a problem.
Rewriting a table with a large body of existing data will
take about as long as you'd expect, and in the meantime,
the rewrite will take an [`ACCESS EXCLUSIVE` lock][locking]
on the table. `ACCESS EXCLUSIVE` is the coarsest
granularity of table lock possible, and it'll block _every_
other operation until it's released; even simple `SELECT`
statements have to wait. In any system with a lot of
ongoing access to the table, that's a huge problem.

!fig src="/assets/postgres-default/blocking.svg" caption="Data loss from contention between two clients."

Historically, accidentally locking access to a table when
adding a column has been a common pitfall for new Postgres
operators because there's nothing in the SQL to tip them
off to the additional expense of adding that `DEFAULT`
clause. It takes a close reading of [the
manual][altertable] to find out, or the pyrrhic wisdom
acquired by causing a minor operational incident.

## Constraints, relaxed by necessity (#constraints)

Because it wasn't possible to cheaply add a `DEFAULT`
column, it also wasn't possible to add a column set to `NOT
NULL`. By definition non-null columns need to have values
for every row, and you can't add one to a non-empty table
without specifying what values the existing data should
have, and that takes `DEFAULT`.

You could still get a non-nullcolumn by first adding it as
nullable, running a migration to add values to every
existing row, then altering the table with `SET NOT NULL`.
Even that isn't perfectly safe because `SET NOT NULL`
requires a full stable scan (as it verifies the new
constraint across all existing data) which is also `ACCESS
EXCLUSIVE`, but it's faster than a rewrite.

The amount of effort involved in getting a new non-null
column in any large relation meant that in practice you
wouldn't bother. It was either too time consuming or too
dangerous.

## Why bother with non-null anyway? (#why-bother)

One of the biggest reasons to prefer relational databases over
document stores and other less sophisticated storage
technology is data integrity. Columns are strongly typed
with the likes of `INT`, `DECIMAL`, or `TIMESTAMPTZ`.
Values are constrained with `NOT NULL`, `VARCHAR` (length),
or [`CHECK` constraints][check]. Foreign key constraints
guarantee [referential integrity][referential].

With a good schema design you can rest assured that your
data is in a high quality state because the very database
is ensuring it. This makes querying or changing it easier,
and prevents an entire class of application-level bugs
caused by data existing in an unexpected state. Enthusiasts
like me have always argued in favor of strong data
constraints, but knew also that new non nullable fields
often weren't possible in Postgres when it was running at
scale.

Postgres 11 brings in a change that makes `ADD COLUMN` with
`DEFAULT` values fast by marshaling them for existing rows
only as necessary. The expensive table rewrite and long
hold on `ACCESS EXCLUSIVE` are eliminated, and a gaping
hole in Postgres' operational story is filled. It will now
be possible to have both strong data integrity and strong
operational guarantees.

## Appendix: How it works (#how-it-works)

The change adds two new fields to
[`pg_attribute`][pgattribute], a system table that tracks
information on every column in the database:

* `atthasmissing`: Set to `true` when there are missing
  default values.
* `attmissingval`: Contains the missing value.

As scans are returning rows, they check these new fields
and return missing values where appropriate. New rows
inserted into the table pick up the default values as
they're created so that there's no need to check
`atthasmissing` when returning their contents.

!fig src="/assets/postgres-default/implementation.svg" caption="Data loss from contention between two clients."

The new fields are only used as long as they have to be. If
at any point the table is rewritten, Postgres takes the
opportunity to insert the default value for every row and
unset `atthasmissing` and `attmissingval`.

Due to the relative simplicity of `attmissingval`, this
optimization only works for default values that are
_non-volatile_ (i.e., single, constant values). A default
value that calls `NOW()` can't take advantage of the
optimization.

There's nothing all that difficult conceptually about this
change, but its implementation wasn't easy because the
system is complex enough that there's a lot of places where
the new missing values have to be considered. See [the
patch][commit] that brought it in for full details.

[altertable]: https://www.postgresql.org/docs/10/static/sql-altertable.html
[check]: https://www.postgresql.org/docs/current/static/ddl-constraints.html#DDL-CONSTRAINTS-CHECK-CONSTRAINTS
[commit]: https://github.com/postgres/postgres/commit/16828d5c0273b4fe5f10f42588005f16b415b2d8
[locking]: https://www.postgresql.org/docs/current/static/explicit-locking.html
[notes]: https://www.postgresql.org/docs/11/static/release-11.html
[pgattribute]: https://www.postgresql.org/docs/current/static/catalog-pg-attribute.html
[referential]: https://en.wikipedia.org/wiki/Referential_integrity
