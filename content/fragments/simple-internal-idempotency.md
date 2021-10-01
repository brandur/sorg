+++
hook = "Avoiding resource leakage between internal systems by using IDs for idempotency."
published_at = 2021-10-01T20:18:35Z
title = "Simple internal idempotency by ID"
+++

Our cloud layer consists of two "fat" services -- a "Platform" layer that internalizes most of our domain logic including cluster metadata, accounts, subscriptions, etc. and an internal service called Owl that specializes solely in the management of Postgres clusters.

Most actions in the Bridge product involve the two services talking to each other, and like with all network communication, there's occasionally a hiccup while doing so.

We wanted to put in some idempotent guarantees so that it'd be safe to have Platform retry failed operations to Owl. The normal way to accomplish this is with [idempotency keys](/fragments/idempotency-key-draft), but the team has a lot going on, and wasn't keen on bringing in that amount of new infrastructure to address a tail risk. Unlike an outfit like Stripe, we're not handling millions of API calls, so while we'd observe occasional resource leaks, they weren't common.

But ... we're proponents in keeping our systems as hands-free as possible, so we found another way that was easier to pull off. We have the benefit of Owl never having to serve any external requests, so we can design a system that's not perfectly robust against abuse.

All our object keys [are UUIDs](/fragments/k-sorted-ids), which have the benefit of global uniqueness. So here's what we did:

* The top-level service (Platform) becomes solely responsible for assigning all new IDs for public objects.

* When creating a new object in owl, Platform generates an ID and pushes it down.

	* If the object doesn't already exist, Owl creates the new object and sends it back with a 201 [1].

	* If the object did exist, Owl sends it back with a 200.

``` ruby
DB.transaction do
  created = T.let(false, T::Boolean)
  rule = T.let(nil, T.nilable(FirewallRule))

  if body.id
    rule = FirewallRule[Id.d(T.must(body.id))]
  end

  if !rule
    created = true
    network = T.must(cluster.network)
    rule = FirewallRule.new network: network, cidr: body.rule
    rule.id = Id.d(T.must(body.id)) if body.id
    rule.save
  end

  respond(body: GetFirewallRuleResponse.from_model(T.must(rule)),
    status: created ? 201 : 200)
end
```

So the generated ID from Platform functionally acts like an idempotency key. This wouldn't be such a good idea for a public-facing API for a couple reasons:

* You're not protecting against user error. With well-implemented idempotency keys, a key and incoming parameters are checked against the parameters that were sent last time the key was used, protecting the user from accidentally reusing the same key across requests that were supposed to be distinct.

* Malicious users would aim for key collisions.

But given Owl's internal nature and robust protection against ID reuse in Platform, this tradeoff works for us -- not quite as much protection, but worth it in return for a much simplified implementation.

The ID injection is used only on `POST`/create endpoints. Owl's `GET`s and `PUT`s are implemented to be idempotent as their HTTP semantics suggest. `DELETE`s are idempotent as well, and we [return `410`s as a usability nicety](/fragments/idempotent-delete) so that retries don't manifest as `404`.

[1] The 201 versus 200 status codes are an aesthetic embellishment and not really important to how the system works.
