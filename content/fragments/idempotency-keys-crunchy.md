+++
hook = "Probably overkill for a relatively low traffic API, but I couldn't help myself."
published_at = 2021-11-25T23:47:55Z
title = "Idempotency keys @ Crunchy"
+++

A few months ago I wrote about [the IETF draft for the `Idempotency-Key` header](/fragments/idempotency-key-draft), a way for an API to provide idempotent operation on non-idempotent verbs like `POST` and `PATCH`.

I'd been intending to an `Idempotency-Key` implementation in at work, and finally [got around to it](https://docs.crunchybridge.com/api/idempotency/).
`
The implementation largely follows what's it in the IETF draft, which itself is largely based off Stripe's original conventions. Similarly to Stripe:

* The `Idempotency-Key` is ignored for idempotent verbs like `GET`, `PUT`, and `DELETE`, but it's explicitly not harmful to send it with any verb so as to simplify client implementation. A client can naively always generate and send a key without worrying about the verb being requested.

* `Idempotency-Replay: true` is set for an operation that had already been completed and is now being replayed. I deviated on naming slightly from Stripe's `Idempotent-Replay` to match the prefix of `Idempotency-Key` (i.e. `-tent` versus `-ency`) because I have OCD for this kind of thing.

* Error semantics are similar -- a `409 Conflict` in case an original request with the same idempotency key is still being processed, or if a new request is made with a request payload which doesn't match.

* Some errors are considered transitory and their results aren't stored, such as a `429 Too many requests` or `503 Service unavailable`.

And some changes this time around:

* We require that keys are UUIDs, whereas Stripe accepts any kind of arbitrary input. This means that storage is more efficient (a tight 16 bytes in the DB) and index lookup is marginally faster thanks to abbreviated keys [1].

    There are some cases where a client might want to send a custom key based on some other non-UUID value they have on their end, which is a little less convenient with UUIDs, but still possible with one extra step -- hash the value and truncate it down to 16 bytes. e.g. SHA3-224 which produces 28 bytes, and keeping the first 16.

* Keys are valid for only one hour instead of 24. This means less to keep in the DB and also makes it clear that keys are only to be used for shorter-term retries. At Stripe we'd occasionally have users do something weird with idempotency keys like treating them as a database lookup so they could run overnight reconciliation jobs and the like.

* API actions are wrapped entirely in transactions meaning that the chances of a request failing on an internal error and leaving the idempotency key in an indeterminate state is much lower.

    Notably at Stripe, because we didn't know whether a failed operation could be safely retried, a 500'ed request would be saved to an idempotency key's result and could not be reprocessed. In our transaction-based implementation, we assume that a 500 can be retried (even if it might fail again), and treat one as a transitory error instead.

* Our deletion mechanism is much more efficient -- deleting in batches via CTE, and with nothing but a count coming back over the wire (see SQL below). At Stripe, we had to delete keys one-by-one for years thanks to limitations in Mongo and our homegrown ORM, with the predictable result that the key cleaner would constantly fall behind. We eventually moved to [TTL indexes](/fragments/ttl-indexes), which fixed that problem but created others [2].

``` sql
-- name: IdempotencyKeyDeleteBeyondHorizon :one
WITH deleted AS (
    DELETE FROM idempotency_key
    WHERE id IN (
        SELECT id
        FROM idempotency_key
        WHERE idempotency_key.created_at < @created_at_horizon
        LIMIT @max
    ) RETURNING *
)
SELECT count(*)
FROM deleted;
```

[1] Okay, lookup is probably similar as long as you made sure to use the "C" locale in Postgres, but very few people would know to do that.

[2] Namely, that there's no way to throttle a TTL index, so the Storage teams didn't have a knob to turn down when a cluster was in major trouble. A story for another day.
