+++
hook = "`Idempotency-Key` is a convention long in use by Stripe to provide idempotency on non-idempotent HTTP verbs, and it now has its own IETF standards draft."
published_at = 2021-07-05T01:49:41Z
title = "`Idempotency-Key` IETF Standards Draft"
+++

From HN today: an IETF draft for a [standardized `Idempotency-Key` header](https://datatracker.ietf.org/doc/html/draft-ietf-httpapi-idempotency-key-header-00), more or less identical to the convention used by Stripe.

This is a subject that's near and dear and which I've [written about](https://stripe.com/blog/idempotency) [extensively before](/idempotency-keys), and it's nice to see it getting some non-Stripe attention. For me, the draft's most interesting aspect is that it compiles a list of heavy hitters who are already using an `Idempotency-Key` almost exactly as proposed:

* Stripe
* Adyen
* Dwolla
* Interledger
* WorldPay
* Yandex

Along with a few smaller ones. One quirk is that the draft seems to be authored by a PayPal employee, and although PayPal has an idempotency concept, it notably does _not_ appear to use `Idempotency-Key` (preferring `PayPal-Request-Id` instead).

The [HN discussion](https://news.ycombinator.com/item?id=27729610) goes a little off the rails -- mostly going to show that idempotency is a nuanced enough subject that even hacker types don't understand it very well ("just use PUT!", "well _I_ for one don't think a standard adds value ...", "it should be called `Request-Id`", etc.). A Google employee argues that that an idempotency key more appropriately belongs in a request's payload -- almost certainly a post-hoc rationalization of a design decision made because Google uses protobuf/GRPC everywhere, and those don't gel well with HTTP headers.

Any API where a leaked resource could make a difference (that is to say,  most of them) should consider implementing `Idempotency-Key`. I can't speak for most of the companies on the list above, but Stripe is living proof that an extremely simple implementation goes a long way. Here's how it works:

1. In a middleware, insert a "partial" idempotency key record to indicate that a request with this key is in progress, keyed to `account_id` + `idempotency_key`. Store when the request started and its parameters.
2. Process the request.
3. As the middleware unwinds, update the record to include the response that's being sent back.
    * You generally store responses for errored requests as well, but want to make sure to only do so for non-intermittent ones. A `422 Unprocessable entity` is definitive because it's determined based on the request's parameters, but a `429 Too many requests` isn't -- those should remove the idempotency key's partial record to give the client a chance to try again.

Now, when another request comes in with the same idempotency key, do one of the following:

* If a previous request with the same key has already been completed, load its response and return it.
* If a previous request with the key is still in a partial state (meaning it's still processing), return `409 Conflict`, which indicates to a client to retry later for a more definite answer.

Stripe used a unique index for the job, detecting duplicates atomically by handling a duplicate key error. Unique indexes in most databases would work well, with the only major consideration being that idempotency keys should generally expire, and not all implementations easily allow this at large volumes (running big `DELETE` jobs can be a problem). Mongo [TTL indexes](/fragments/ttl-indexes) do the trick, as would Redis/Redis Cluster where key expiry is a core feature.

Personally, I think that a more sophisticated idempotency key approach than Stripe's is warranted, but it's a case of not letting perfect be the enemy of good, and many API providers should consider including at least a basic version of the idea.
