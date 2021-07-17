+++
hook = "Making `DELETE` endpoints in web APIs not just _technically_ idempotent, but friendly for clients that may need to call them multiple times in a retry loop."
published_at = 2021-07-17T19:35:36Z
title = "Retry-friendly idempotent delete in web APIs"
+++

I've been thinking about idempotency in APIs a lot this week as I was working to improve the idempotency story within one of our internal services so that we could retry requests to it and minimize data inconsistencies.

For our public API, I'll likely implement something resembling the [`Idempotency-Key` IETF draft](/idempotency-key-draft), at least for `POST` and `PATCH` verbs. We're not likely to have idempotency keys for all internal services though, so I've been leveraging alternate idempotency techniques instead. A good example of an alternative technique is to allow clients to inject an ID that they generated when creating a resource, which easily allows the receiver to determine whether they've already processed the request. This might not work so well for public APIs (it introduces the possibility of malicious IDs), but it works fine internally.

After tackling idempotency on our create endpoints, I moved on to `DELETE`. This is an interesting verb because `DELETE` is supposed to be idempotent according to [RFC7231](https://datatracker.ietf.org/doc/html/rfc7231) (along with `OPTIONS`, `HEAD`, `GET`, and `PUT`) and in many implementations _technically_ is, it's often somewhat unhelpful about what it signals back to the caller.

Here's a delete implementation from a Rails generated controller:

``` ruby
def destroy
  @article = Article.find(params[:id])
  @article.destroy

  redirect_to root_path
end
```

A first call to `destroy` looks up and destroys the record, then `302`s the user back to root. A second request will fail to find the object and `404` instead.

## Correct v. friendly (#correct-friendly)

Now technically, this _is_ idempotent because idempotency doesn't guarantee anything with respect to status codes returned -- it just concerns itself with side effects of each operation, and that second request had no additional side effects compared to the first. So from a strict correctness perspective, it's correct.

However, it's also not very helpful. If a caller lost the response of the first request and got a `404` back on the second, it's hard for them to tell whether one of their requests did something, or were invalid as that object was never present in the first place.

## Acts as paranoid (#acts-as-paranoid)

On this go around, I'm going to try and do something with our `DELETE` APIs. In many product implementations some kind of "acts as paranoid" facility is common, where you add a `deleted_at` timestamp to tables or keep a separate "tombstone" collection to soft delete objects instead of culling them permanently, and we've had that since well before I started.

Conveniently, soft deletion acts as an easy stepping stone do something a little smarter with our delete implementations:

* A "normal" first time deletion returns a `204 No Response`, as is our usual convention.
* If we find that the object is already deleted, but did exist, we return a `410 Gone` instead.
* An object that never existed sends back a `404 Not Found`.

As a client, you can now reliably use either a `204` or `410` to know that the request succeeded, and that you'll get back one of these responses even in the presence of multiple attempts and intermittent network errors. A `404` tells you something unexpected happened. This is also nicely symmetrical with the convention of returning either a `201` or `200` on an idempotent create endpoint depending on whether the object already existed or not.

This is also not to say that those old objects need to be saved forever so that a `410` is always returned. The value of something like this or `Idempotency-Key` is not to have a permanent archive of old requests, they're to be able to reliably make sense of recent ones, usually those that have occurred just in the last few seconds as you're engaged in a retry loop. It's fair game for that `410` to become a `404` after 24 hours [1] or so as old objects are pruned.

I'll be the first to acknowledge that this isn't a ground breaking new idea, and may not even be worth the effort in many cases -- for the most part, clients are mostly just concerned that objects are gone, and getting a `404` back is good enough -- but it is a small convenience on the road to a retry-friendly (and friendly in general) API.

[1] 24 hours being the lifetime of an `Idempotency-Key` at Stripe before it gets pruned, although I'm sure this was chosen somewhat arbitrarily.
