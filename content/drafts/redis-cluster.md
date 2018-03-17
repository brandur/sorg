---
title: Scaling Stripe's Rate Limiting Stack With Redis
  Cluster
location: San Francisco
published_at: 2018-03-17T16:52:38Z
hook: TODO
---

## Distribution model (#distribution)

16384 slots.

```
HASH_SLOT = CRC16(key) mod 16384
```

This is what I really like about Redis: I've spent almost
no time at all studying it source code, but most of the
mechanics it employs (even for complex features like
distributed sharding) are simple enough for even a
layperson like myself to understand and reason about.

## Localizing multi-key operations with hash tags (#hash-tags)

It's quite common in Redis to run operations that operate
on multiple keys through the use of the `EVAL` command and
a custom Lua script. This is an especially important
feature for implementing rate limiting, because all the
work dispatched via a single `EVAL` is guaranteed to be
atomic -- this allows us to correctly calculate remaining
quotas even where there's other concurrent operations in
flight that would otherwise conflict.

A distributed model would make this type of multi-key
operation difficult. Because the slot of each key is
calculated via hash, there'd be no guarantee that related
keys would map to the same slot. My keys
`user123.first_name` and `user123.last_name`, obviously
meant to belong to the same relation, could end up on two
completely different nodes in the cluster. An `EVAL` that
read from both of them wouldn't be able to run on a single
node without an expensive remote fetch from another.

To concrete this with a simple example, let's say we have
an `EVAL` operation that concatenates a first and last name
to produce a full name:

```
# Gets the full name of a user
EVAL "return redis.call('GET', KEYS[1]) .. ' ' .. redis.call('GET', KEYS[2])"
    2 "user123.first_name" "user123.last_name"
```

Here's a sample invocation:

```
> SET "user123.first_name" William
> SET "user123.last_name" Adama

> EVAL "..." 2 "user123.first_name" "user123.last_name"
"William Adama"
```

This script would have trouble running on Redis Cluster if
it didn't provide a mechanic to allow it. Luckily it does
through the use of "hash tags".

The Redis Cluster answer to `EVAL`s that would require
cross-node operations is to disallow them (a choice that
optimizes for speed). Instead, it's the user's jobs to
ensure that the keys that are part of any particular `EVAL`
map to the same slot by hinting how a key's hash should be
calculated with a hash tag. Hash tags look like curly
braces in a key's name, and they dictate that only the
surrounded part of the key is used for hash calculation.

We'd fix our script above by redefining our keys to only
use their shared `user123` content for hashing:

```
> EVAL "..." 2 "{user123}.first_name" "{user123}.last_name"
```

`{user123}.first_name` and `{user123}.last_name` are now
guaranteed to map to the same slot, and `EVAL` operations
that contain both of them will be trouble-free. And
although this is only a basic example, the exact same
concept maps all the way up to a complex rate limiter
implementation.

## `MOVED` redirection (#moved)

```
GET x
-MOVED 3999 127.0.0.1:6381
```

[spec]: https://redis.io/topics/cluster-spec
