---
title: Partitioning In MongoDB
published_at: 2017-04-25T15:55:53Z
hook: A day in trench life.
---

Use MongoDB for a day, and it's a rarity not to come across
some sort of deficiency that would have been trivially
tractable on a different system.

Yesterday, I needed a way of partitioning a data set
(specifically, our idempotency keys) so that each segment
could be worked independently by N workers. A common/easy
way to do this is to simply modulo by the number of
partitions and have each worker switch off the result:

``` ruby
# number of partitions
n = 10

# worker 0
work.select { |x| x.id % n == 0 }

# worker 1
work.select { |x| x.id % n == 1 }

...
```

Here I didn't have an integer key that I could use, but I
did have a suitable string key in the form of each object's
primary key which looks like `idr_12345`. That's fine
though, because we can still get an adequate modulo
candidate by hashing the string value and converting some
of the bytes to an integer.

But wait! You can't cast in a MongoDB query. The best you
can do is fall back to JavaScript:

``` js
// This doesn't actually work -- you'd need to find some
// hashing scheme that's available from JavaScript.
db.idempotent_keys.
    find("parseInt(md5(this.id).substring(0,8), 8) % 10 == 0")
```

But now that you're working in JavaScript, MongoDB can't
use any of its indexes.

With a sane database (i.e. Postgres), I'd cast right in the
query:

``` sql
--
-- the WHERE expression takes 8 bytes from the MD5 hash and
-- converts them to an 32-bit int
--
SELECT *
FROM idempotent_keys
WHERE (('x'||substr(md5(id),1,8))::bit(32)::int % 10) = 0;
```

But hold on, that also can't use the index on `id`.
Luckily, [Postgres supports building indexes on arbitrary
expression][indexed-expressions]. One quick command and the
problem is definitively solved:

``` sql
CREATE INDEX index_idempotent_keys_on_id_partitioned ON idempotent_keys
    ((('x'||substr(md5(id),1,8))::bit(32)::int % 10));
```

Back in MongoDB world, I'm going to have to add a new
column specifically for partitioning, modify my code to
write to it, backfill old values, and even after all that I
still won't have a usable index!

Choose your technology carefully folks. You can spend your
time either making mediocre databases operable, or solving
actual problems.

[indexed-expressions]: https://www.postgresql.org/docs/current/static/indexes-expressional.html
