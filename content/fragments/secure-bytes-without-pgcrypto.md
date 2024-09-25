+++
hook = "Avoiding the `pgcrypto` extension and its OpenSSL dependency by generating cryptographically secure randomness through `gen_random_uuid()`."
published_at = 2024-09-24T11:38:37-07:00
title = "A few secure, random bytes without `pgcrypto`"
+++

In Postgres it's common to see the SQL `random()` function used to generate a random number, but it's a pseudo-random number generator, and not suitable for cases where real randomness is required critical. Postgres also provides a way of getting secure random numbers as well, but only through the use of the `pgcrypto` extension, which makes `gen_random_bytes` available.

Pulling `pgcrypto` into your database is probably fine---at least it's a core extension that's distributed with Postgres itself---but while testing the RC version of [Postgres 17](https://www.crunchydata.com/blog/real-world-performance-gains-with-postgres-17-btree-bulk-scans) last week, I found that it was surprisingly difficult to build Postgres against OpenSSL, which unsurprisingly is required to build `pgcrypto`, thereby making `pgcrypto` itself hard to build.

I'm broadly against the use of Postgres extensions because they make upgrades harder and projects less portable [1], so we have a minimal posture when it comes to them, depending only on `btree_gist` and `pgcrypto`. Like `pgcrypto`, `btree_gist` is also distributed with Postgres, but unlike `pgcrypto`, doesn't have an OpenSSL dependency, making it trivial to build.

Rather than wasting more time trying to get OpenSSL configured, I did a quick code audit to find out where we were using `pgcrypto`, and found that we were using it in exactly one place to generate random bytes for use in [a ULID](/nanoglyphs/026-ids):

``` sql
-- 10 entropy bytes
ulid = timestamp || gen_random_bytes(10);
```

Needing a whole extension for generating a few random bytes seems like a waste, but unfortunately Postgres doesn't offer a built-in way to get cryptographically secure random bytes in any other way ... or does it?

## Secure bytes, just not for you (#secure-bytes)

Internally, Postgres has a module called `pg_strong_random.c` that exports a `pg_strong_random()` function that will use OpenSSL if available, but can fall back to `/dev/urandom` in case it's not, which is perfectly fine for our purposes:

``` c
/*
 * pg_strong_random & pg_strong_random_init
 *
 * Generate requested number of random bytes. The returned bytes are
 * cryptographically secure, suitable for use e.g. in authentication.
 *
 * Before pg_strong_random is called in any process, the generator must first
 * be initialized by calling pg_strong_random_init().
 *
 * We rely on system facilities for actually generating the numbers.
 * We support a number of sources:
 *
 * 1. OpenSSL's RAND_bytes()
 * 2. Windows' CryptGenRandom() function
 * 3. /dev/urandom
 *
 * Returns true on success, and false if none of the sources
 * were available. NB: It is important to check the return value!
 * Proceeding with key generation when no random data was available
 * would lead to predictable keys and security issues.
 */
 ```
 
So secure randomness is available without needing to dip into OpenSSL or `pgcrypto`. Postgres just doesn't make it available to you.

## Roundabout randomness (#roundabout-randomness)
 
Luckily, there's a workaround. `pg_strong_random()` is called through another function that's exported to userspace, Postgres 13's `gen_random_uuid()` which generates a V4 UUID that's secure, random data with the exception of six variant/version bits in the middle:
 
 ``` c
Datum
gen_random_uuid(PG_FUNCTION_ARGS)
{
    pg_uuid_t  *uuid = palloc(UUID_LEN);

    if (!pg_strong_random(uuid, UUID_LEN))
        ereport(ERROR,
                (errcode(ERRCODE_INTERNAL_ERROR),
                 errmsg("could not generate random values")));

    /*
     * Set magic numbers for a "version 4" (pseudorandom) UUID, see
     * http://tools.ietf.org/html/rfc4122#section-4.4
     */
    uuid->data[6] = (uuid->data[6] & 0x0f) | 0x40;    /* time_hi_and_version */
    uuid->data[8] = (uuid->data[8] & 0x3f) | 0x80;    /* clock_seq_hi_and_reserved */

    PG_RETURN_UUID_P(uuid);
}
```

Given our use of `pgcrypto` is so limited, and we only need ten random bytes at a time for a ULID, I changed our `gen_ulid()` implementation to find ten bytes of randomness by pulling five bytes off the front and back of a V6 UUID:

``` sql
-- 10 entropy bytes
--
-- We extract these by generating a random UUID and extracting
-- the first five bytes and last bytes out of it (thus avoiding
-- versioning bits in the middle). This is a roundabout way of
-- doing this, but is done to avoid a dependency on the pgcrypto
-- extension just to get `gen_random_bytes()`.
--
-- `uuid_send()` changes `uuid` to `bytea`.
random_uuid = uuid_send(gen_random_uuid());
ulid = timestamp ||
    substring(random_uuid FROM 1 FOR 5) ||
    substring(random_uuid FROM 12 FOR 5);
```

Which then lets us rid ourselves of `pgcrypto`, along with OpenSSL:

``` sql
DROP EXTENSION pgcrypto;
```

Making tests against a locally built version of Postgres considerably easier.

I'm hoping we can ditch this hack as soon as V7 UUIDs land in core (they didn't make Postgres 17, which is very sad), but in the mean time, this trick might be useful to someone else.
