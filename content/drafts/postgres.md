---
title: "Postgres: Useful and Non-obvious Features"
published_at: 2016-08-25T18:58:54Z
hook: A listing of all useful Postgres features, tricks, and advice that I've
  accumulated over the years and which aren't easy to discover without someone
  telling you about them first.
location: San Francisco
---

## Schemas

### BIGINT and BIGSERIAL (#bigint-and-bigserial)

When adding an integer column or one that should contain a sequence, it's
tempting to reach for the defaults of `INT` and `SERIAL`.

Both of these are 32-bit data types. There are a few important things to know
about them:

1. If you run out of runway on a 32-bit `SERIAL`, you'll need to resize it.
   This is not a cheap operation.
2. On a 64-bit machine (which are pretty much ubiquitous these days, a 32-bit
   type isn't actually going to save you any space [1].

Save yourself some pain and always use `BIGINT` and `BIGSERIAL` instead.

### NOT NULL (#not-null)

Apply `NOT NULL` liberally and _early_.

An app's data starts small and uniform, but as it grows, those previously
unthinkable edge cases start to happen more and more often until the data still
follows a general _shape_, but with many exceptions that were never supposed to
be there. This will manifest as a problem during migrations for example, where
a script encounters something unexpected and crashes, forcing you to go in an
examine it manually.

The beauty of a relational database is that it can apply constraints even when
an application talking to hit has a bug. Those constraints can form a final and
impenetrable layer of defense to ensure data that's universally consistent.

Constraints are good, and the more the better. While they might slow your
development down in the early stages of a project, they'll save you thousands
of hours of manual data correction later on. Always default columns that should
be required to `NOT NULL`.

### Text Fields (#text)

In Postgres, `varchar` [has absolutely no performance or storage advantage over
`text`][text]. Over a certain size of data, all values end up [in TOAST][toast]
(The Oversized-Attribute Storage Technique).

`varchar` can cost you though, in that it can cause significant churn if you
realize later that you need a different size and have to alter the table.

Always use `text`. If you need to constraint its check (and you _absolutely_
should), use a `CHECK` constraint (see below). A `CHECK` constraint can be
modified instantly at any time.

### JSONB (#json)

Occasionally you might want to have a grab bucket of arbitrary values in a
table and resort to one of the `json`, `jsonb`, or `hstore` types.

`json` is functionally equivalent to `hstore` except better because it supports
the same basic types as found in a normal JSON documents (`hstore` values can
only be strings).

In the same way, `jsonb` is functionally identical to `json` except better:
when operating on a `jsonb` column Postgres doesn't need to reparse its
contents; they're already available in a format that can be loaded directly
into memory. `jsonb` also supports indexing.

There's a simple heuristic: always pick `jsonb`.

### Partial Indexes

Postgres allows an index to be created with a `WHERE` clause. This can save a
lot of storage space in a large table by allowing you to only build indexes
along predicates that we care about.

But where partial indexes really shine is when they're combined with `UNIQUE`.
Imagine this situation:

``` sql
CREATE UNIQUE INDEX index_users_on_email WHERE deleted_at IS NULL;
```

This an incredibly easy way to ensure that no two users are created with the
same email. You can even check this at the application level as well to give
most users a nice error message, but rely on Postgres to guarantee that two
requests to create a user with the same email only a millisecond apart can't
both succeed.

### Timestamps (#timestamps)

Use `timestamptz` (otherwise known as `timestamp with time zone`) everywhere
and store everything as UTC.

This might seem obvious, but it's surprisingly easy to make mistakes when it
comes to times. I managed an old database that started storing times in Pacific
and it was such an incredible pain amount of effort to fix (there were hundreds
of time columns in total) that even years later it never happened.

### Check Constraints

SQL allows `CHECK` constraints to be defined on a table. These will verify that
an arbitrary boolean condition is true before allowing an insert or update.

For example, we might verify that a price is a positive value:

``` sql
CHECK (price > 0)
```

Or that a `text` field complies to a certain length (similar to what a
`VARCHAR` gets you except that it can be changed instantly):

``` sql
CHECK (char_length(name) <= 255)
```

Constraints are good. Even if you have equivalent application-level validation
logic, use `CHECK` liberally.

## psql (#psql)

While there are a few decent GUIs available that will talk to Postgres, I'd
suggest becoming intimately familiar with Postgres' built-in psql tool. It's an
incredibly efficient way to quickly get visibility into and manage a running
database.

### Editor

Use `$EDITOR` variable to look at queries:

```
\e
```

### Automatic Results Formatting (#x-auto)

```
\x auto
```

### Interactive Error Rollback (#on-error-rollback-interactive)

Imagine a situation where you're in a transaction, finish executing a
long-lived command, and go to perform a minor check, make a typo, and get one
of these:

```
ERROR: current transaction is aborted, commands ignored until end of transaction block
```

Sets a savepoint for every command, but only for interactive Psql sessions (so

```
\set ON_ERROR_ROLLBACK interactive
```

## Operations

### Raise and Drop Indexes Concurrently

The one caveat is that you can't run concurrent operations from inside a
transaction.

### Dangerous Operations

## Features

### Listen/Notify

## SQL

### Intervals

``` sql
SELECT now() - '1 month'::interval;
```

### With Clauses

Names subselects

PVH: You can actually share a SQL query with another human being [2].

``` sql
WITH ()
```

### Window Functions

``` sql
OVER PARTITION BY
```

[1] This is true in the general case, but if a 32-bit type is compounded with
another 32-bit type in an index, there is _some_ potential for space saving.

[2] See [Postgres: The Bits You Haven’t Found](https://vimeo.com/61044807)
(2013).

### pg_stat_activity

See running processes.

[text]: https://www.postgresql.org/docs/current/static/datatype-character.html
[toast]: https://www.postgresql.org/docs/current/static/storage-toast.html
