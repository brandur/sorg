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

Due to the way data is aligned on a 64-bit machine, 32-bit types like `INT` and
`SERIAL` won't actually save you any space. They can be actively destructive
though in that if you run out of runway on these types, you'll have to an
expensive table alteration later.

_Always_ use `BIGINT` and `BIGSERIAL`.

### NOT NULL (#not-null)

Constraints are good. Always use `NOT NULL`.

### TEXT (#text)

Always use `TEXT`.

Use a CHECK constraint for length if so desired. This can be cheaply modified
at a later time.

See Toast.

### JSONB (#json)

Always use `JSONB` over `JSON` or `HSTORE`.

### Partial Indexes

``` sql
CREATE INDEX index_users_on_email WHERE deleted_at IS NULL;
```

## psql (#psql)

### Automatic Results Formatting (#x-auto)

```
\x auto
```

### Interactive Error Rollback (#on-error-rollback-interactive)

Imagine a situation where you're in a transaction, finish executing a
long-lived command, and go to perform a minor check, make a typo, and get one
of these:

```
ERROR:  current transaction is aborted, commands ignored until end of transaction block
```

Sets a savepoint for every command, but only for interactive Psql sessions (so

```
\set ON_ERROR_ROLLBACK interactive
```

## Operations

### Raise and Drop Indexes Concurrently

The one caveat is that you can't run concurrent operations from inside a transaction.

### Dangerous Operations
