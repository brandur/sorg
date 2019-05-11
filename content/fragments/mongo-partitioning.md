+++
hook = "A day in trench life."
published_at = 2017-04-28T22:30:13Z
title = "Partitioning in MongoDB (or lack thereof)"
+++

A day doesn't pass of using MongoDB that doesn't make you
reminisce of time spent on a real database.

I needed a way of partitioning a data set (specifically,
our idempotency keys) so that _N_ workers could
independently work their own segment. A common/easy way to
do this is to simply modulo by the number of partitions and
have each worker switch off the result:

``` ruby
# number of partitions
N = 10

# worker 0
work.select { |x| x.id % N == 0 }

# worker 1
work.select { |x| x.id % N == 1 }

...
```

Here I didn't have an integer key that I could use, but I
did have a suitable string key in the form of each object's
ID which looks like `idr_12345`. That's okay though,
because the ID is unique and makes a fine modulo candidate;
all we have to do is hash its value and convert some of the
resulting bytes to an integer.

Unfortunately, you can't cast in a MongoDB query. The best
you can do is fall back to JavaScript:

``` js
// This doesn't actually work -- you'd need to find some
// hashing scheme that's available from JavaScript which
// is an adventure of its own.
db.idempotent_keys.
    find("parseInt(md5(this.id).substring(0,8), 8) % 10 == 0")
```

But now that you're working in JavaScript, MongoDB can't
use any of its indexes.

With a sane database, I'd cast right in the query:

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
expressions][indexed-expressions], a feature that I took
for granted for a long time, but whose utility comes into
sharp relief after you lose it.

One quick command and the problem is definitively solved:

``` sql
CREATE INDEX index_idempotent_keys_on_id_partitioned ON idempotent_keys
    ((('x'||substr(md5(id),1,8))::bit(32)::int % 10));
```

That's a little ugly. No problem though; we can easily use
an `IMMUTABLE` function to nicen things up, and call it
from right within `CREATE INDEX`.

``` sql
CREATE FUNCTION text_to_integer_hash(str text) RETURNS integer AS $$
    BEGIN
        RETURN ('x'||substr(md5(str),1,8))::bit(32)::int;
    END;
$$
LANGUAGE plpgsql
IMMUTABLE;

CREATE INDEX index_idempotent_keys_on_id_partitioned ON idempotent_keys
    ((text_to_integer_hash(id) % 10));
```

Back in MongoDB world, I had to add a column specifically
for partitioning, modify my code to write to it, backfill
old values, and even after all that I *still won't have a
usable index!*

Choose your technology carefully folks. You can spend your
time either making mediocre databases operable, or solving
actual problems.

[indexed-expressions]: https://www.postgresql.org/docs/current/static/indexes-expressional.html
