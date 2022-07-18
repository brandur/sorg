+++
hook = "Tweaking error handling in our idempotency layer with a new transience flag."
published_at = 2022-07-18T17:59:22Z
title = "Idempotency: The `is_transient` property"
+++

I've written previously about [implementing idempotency keys](/fragments/idempotency-keys-crunchy), an idea stemming from APIs like Stripe's, but which has gotten wider adoption more recently including [it's own IETF draft](https://datatracker.ietf.org/doc/html/draft-ietf-httpapi-idempotency-key-header-00).

One of the quirks of idempotency keys is that while the responses of successful requests are always captured, errors are more nuanced. They fall into two categories:

* Errors that result from properties intrinsic to the request itself. For example, if a caller requests a resource be created with an invalid name, then barring a supremely rare change in naming policy, any number of retries will always result in the same failed response.

* Errors that result from some transitory aspect of the system's state at that time. For example, if a request conflicted with another in-flight request that was modifying the same data.

An idempotency layer can safely cache errors of the former type, but not the latter, which generally includes a set of status codes that are known to be transitory like 409 (conflict), 429 (too many requests -- usually rate limiting), and 500 (internal server error). This is how Stripe's implementation works for example.

Recently, a question came up around idempotency should work for errors that are "semi" transitory. For example, if a user tries to create a resource but they're overquota, we'd send back a 400 with an error message indicating so. And while we wouldn't expect their quota use to decrease immediately, we could expect it to change over the medium-term as they delete other resources.

I'm still not sure whether I agree that callers should be leaning on the idempotency layer to know not to cache for something like a filled quota, but the discussion got me thinking that it might still be useful to have the capability of treating errors that are traditionally transient as not, or those that are traditionally non-transient as transient, and it might also bring about a simplification opportunity.

## Signaling via `is_transient` (#is-transient)

Our API errors have a standard JSON shape including `message` containing a human-readable string and `status` for an embedded status code.

We've now added a new boolean property for `is_transient` that's always sent back with the standard shape as well. The initializers for our traditionally transient errors like 409, 429, and 500 all set the property by default, and others do not. But because it's a per-object setting, we still have the flexibility to set it on or off for any specific error being returned.

``` json
{
    "is_transient":true,
    "message":"Request rate limited. Please wait a short time before trying another.",
    "status":429
}
```

The idempotency layer's implementation changed so that it now only looks at this property to make an error caching decision. If `is_transient` is `false` then the error response is cached, and if `true`, it's not. The error's status code is no longer considered, effectively delegating the caching decision purely upstream.

``` go
// Consider certain types of errors to be transient (e.g. 429, 500,
// 503). If we encountered one of these, delete the idempotency key
// record and allow the user to retry.
if err == nil && apiErr.IsTransient {
    _, err := queries.IdempotencyKeyDeleteByID(ctx, key.ID)
    if err != nil {
        return xerrors.Errorf("error deleting idempotency key: %w", err)
    }
    return nil
}
```

## Benefits: ergonomics and compatibility (#benefits)

The property also serves as a useful signal to API callers. Previously, one had to read the documentation to know which types of errors could be expected to be cached and which were not. Now, a caller need only examine `is_transient` to know immediately whether the response was cached and whether it'd be useful to retry.

Along with simplifying API integrations, it's also better for compatibility -- authority on which errors are transient and which aren't is now fully delegated back to the API server itself instead of specific error types being baked into the integration. If the transience of an error changes, clients do the right thing automatically instead of needing an update to reclass it.

Lastly, our system is implemented such that a single API facade provides an idempotency layer for every other component. Because our API error shape is shared between all components, backend components that know very little about the idempotency layer can still control its behavior by flagging `is_transient` on or off.
