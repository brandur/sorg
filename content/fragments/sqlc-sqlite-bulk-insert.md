+++
hook = "Using `json_each` and `json_extract` to get bulk inserts in SQLite with sqlc."
# image = ""
published_at = 2026-06-06T13:15:27+02:00
title = "SQLite bulk insert with sqlc"
+++

One of [sqlc](/sqlc)'s long-standing quirks is that it doesn't support `INSERT INTO table ( ... ) VALUES ( ... )` with multiple tuples. This forces you to find a workaround where you can inject multiple values as a single value, like using an array in Postgres and `unnest`ing it in SQL.

Back when I first added SQLite support to River, I had a *really* hard time finding a way to do this. Not so much because SQLite didn't support it (though it doesn't have arrays), but rather due to a number of longstanding bugs in sqlc that prevented the use of [table-valued functions](https://sqlite.org/vtab.html#tabfunc2) like `json_each`. I ended up giving up on bulk insert in SQLite and looping over single insert operations instead.

But that was over a year ago. Happily, I took a look at this again today and found the sqlc bugs fixed.

To insert a series of single values, inject a JSON array of scalars as a binary blob, then iterate over it with `json_each` and select `value`:

``` sql
-- name: RiverMigrationInsertMany :many
INSERT INTO /* TEMPLATE: schema */river_migration (
    line,
    version
)
SELECT
    @line,
    value
FROM json_each(cast(@versions AS blob))
RETURNING *;
```

To insert a series of multiple values, inject a JSON array of objects as a binary blob, then iterate over it with `json_each` and extract each value with `json_extract`:

``` sql
-- name: NotificationInsertMany :exec
INSERT INTO /* TEMPLATE: schema */river_notification (
    payload,
    topic
)
SELECT
    json_extract(value, '$.payload'),
    json_extract(value, '$.topic')
FROM json_each(cast(@notifications AS blob));
```

If using `jsonb` instead of `json`, you'll use `jsonb_each` and `jsonb_extract` instead.

I'm not sure about the use of sqlc anymore these days -- it seems to have largely fallen out of maintenance, and isn't quite as necessary in the LLM age. In River, we haven't had a strong reason to migrate off it yet, so for now we're holding the line.
