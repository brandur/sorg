+++
hook = "How to optimize field order in a row to avoid padding (for when tables are expected to be really big) and a useful query to help."
published_at = 2023-10-12T11:23:22-07:00
title = "Optimizing row storage in 8-byte chunks"
+++

When Postgres stores physical rows, it aligns fields in 8-byte chunks [1]. For 8-byte data types like a UUID or `bigserial` you don't have to think about field ordering because it doesn't matter, but when a type's size isn't a multiple of eight, space is lost to padding unless fields are tetris'ed to maximize storage efficiency.

Say we have `CREATE TABLE foo (a int4, b int8, c int4)`. `int4`s are 4 bytes, so they're padded to get to 8-byte alignment:

![Row with space lost to padding](/assets/images/fragments/postgres-row-optimization/with-gaps.svg)

This can be avoided by reordering the `int4s` so they're together (`CREATE TABLE foo (a int4, c int4, b int8`):

![Row with fields tetris'ed for maximum storage efficiency](/assets/images/fragments/postgres-row-optimization/without-gaps.svg)

Most of the time it's not worth thinking about. Storage may not be maximally efficient, but it's efficient enough, and the largest values in a row like `text` or `jsonb` are stored out-of-band anyway. However, occasionally you have a situation where you want to store zillions of rows, and doing a little work to optimize tuple size yields returns (I came across a case like this recently which is why I'm writing about it). Eight wasted bytes is nothing, but what about eight times a million?

2nd Quadrant's already [written in good detail on column ordering](https://www.2ndquadrant.com/en/blog/on-rocks-and-sand/) ("On Rocks and Sand" -- I'm giving them the lifetime award for most dramatic Postgres title) so I'll avoid saying too much more on the subject, but I wanted to call out one of their queries that's a handy way to reveal the type size of each column in a table:

``` sql
SELECT a.attname, t.typname, t.typalign, t.typlen
  FROM pg_class c
  JOIN pg_attribute a ON (a.attrelid = c.oid)
  JOIN pg_type t ON (t.oid = a.atttypid)
 WHERE c.relname = 'user_order'
   AND a.attnum >= 0
 ORDER BY t.typlen DESC;

   attname   |   typname   | typalign | typlen 
-------------+-------------+----------+--------
 id          | int8        | d        |      8
 user_id     | int8        | d        |      8
 order_dt    | timestamptz | d        |      8
 ship_dt     | timestamptz | d        |      8
 receive_dt  | timestamptz | d        |      8
 item_ct     | int4        | i        |      4
 order_type  | int2        | s        |      2
 is_shipped  | bool        | c        |      1
 tracking_cd | text        | i        |     -1
 ship_cost   | numeric     | i        |     -1
 order_total | numeric     | i        |     -1
```

(`typalign` value meanings are `c` = char alignment, `s` = short alignment, `i` = int alignment (4 bytes on most machines), `d` = double alignment (8 bytes).)

You'll generally start by putting the larger 8-byte stuff at the beginning, then work in descending order of type size, and leaving the variable-length values at the end, so this query gives you an easy starting point.

## `NULL` (#null)

Another factor to consider is nullability. `NULL` values in Postgres aren't stored as a value, but rather as a flag in a ["null bitmap" located right after a heap tuple's header](https://www.postgresql.org/docs/current/storage-page-layout.html#STORAGE-TUPLE-LAYOUT). So if you're really trying to optimize for size, it'd make sense to order nullable fields whose values are often omitted after non-null fields.

[1] Technically chunks that are multiples of `MAXALIGN` which is pointer-sized, but I'm simplifying a little since the year is 2023.