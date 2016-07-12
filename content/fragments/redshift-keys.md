---
title: Showing Redshift Distkey & Sortkey
published_at: 2015-10-26T18:25:13Z
---

One slightly unfortunate aspect of how Postgres interacts with Redshift is that
standard tooling like `\d+` can't be used to inspect a table's distkey or
sortkey. As such, the recommended way of showing these is by querying the
[`pg_table_def`][pg-table-def] table.

```
SELECT "column", type, distkey, sortkey
FROM pg_table_def
WHERE schemaname = 'logs' AND
    tablename = 'bapi' 
AND distkey = true
    OR sortkey <> 0;

 column  |            type             | distkey | sortkey
---------+-----------------------------+---------+---------
 created | timestamp without time zone | f       |       1
```

**But wait!** There's one more gotcha here to think about. As described in the
documentation, only schemas that are contained in the user's
[search_path][search-path] will be shown in `pg_table_def`. If you try to query
for a table outside of a `search_path` schema, it may mysteriously come up
empty. Make sure to set `search_path` like so:

```
set search_path to '$user', infra, logs, public;
```

(And recall that this can go into your `~/.psqlrc`.)

[pg-table-def]: http://docs.aws.amazon.com/redshift/latest/dg/r_PG_TABLE_DEF.html
[search-path]: http://docs.aws.amazon.com/redshift/latest/dg/r_search_path.html
