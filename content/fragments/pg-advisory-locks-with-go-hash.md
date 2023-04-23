+++
hook = "Using advisory locks in Go, taking advantage of the full 64-bit key space using Go's built in hashes."
published_at = 2023-04-22T21:19:22-07:00
title = "PG advisory locks in Go with built-in hashes"
+++

Postgres has a feature called [advisory locks](https://www.postgresql.org/docs/current/explicit-locking.html) that allow a client to take out specific locks whose meanings are defined by an application using Postgres rather than within Postgres itself. They're useful for app coordination, like ensuring that only one instance of a certain program starts.

Because Postgres tracks lock IDs as integers internally, the advisory lock functions take a 64-bit integer as a key:

```
pg_try_advisory_lock(int8)
```

But on the application side, it's common to want to use a string as a lock key rather than an arbitrarily defined integer, and that's not supported directly. Postgres provides some functions that can be used as quick workarounds:

```
SELECT pg_advisory_lock(hashtext('my_app'));
```

But that's not great either becauase `hashtext` produces a 32-bit output, making hash collisions a non-zero possibility.

Go provides a built-in [`hash` package](https://pkg.go.dev/hash) that lets us get an advisory lock quite elegantly. We invoke lock acquisition as an [sqlc](/sqlc) operation:

``` sql
-- name: PGTryAdvisoryLock :one
SELECT pg_try_advisory_lock(@key);
```

Then write a simple helper that produces 64-bit output using an [FNV](https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function), a non-crypotographic hash that Go provides convenient built-ins for.

``` go
// `pg_try_advisory_lock` takes a bigint rather than any kind of human-readable
// name. Just so we don't have to choose a random integer, hash a provided name
// to a bigint-compatible 64-bit uint64 and use that.
func keyNameAsHash64(keyName string) uint64 {
	hash := fnv.New64()
	_, err := hash.Write([]byte(keyName))
	if err != nil {
		panic(err)
	}
	return hash.Sum64()
}
```

Now we get a lock, taking advantage of the full 64-bit key space:

``` go
locked, err := dbsqlc.New(lockAndListenConn).PGTryAdvisoryLock(ctx, int64(keyNameAsHash64("worker")))
if err != nil {
    return xerrors.Errorf("error trying to acquire lock: %w", err)
}
```