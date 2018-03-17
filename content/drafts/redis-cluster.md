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

```
# Gets the full name of a user
EVAL "return redis.call('GET', KEYS[1]) .. ' ' .. redis.call('GET', KEYS[2])"
    2 "user123.first" "user123.last"
```

## `MOVED` redirection (#moved)

```
GET x
-MOVED 3999 127.0.0.1:6381
```

[spec]: https://redis.io/topics/cluster-spec
