---
title: Stripe Webhook Signatures & DX
published_at: 2017-05-10T02:59:44Z
hook: TODO
---

I'm really excited to say that last week we released
support for signed webhooks. This has been a feature that
users have requested for a long time, and which we've
wanted to provide for approximately forever. We're
certainly not doing anything new or novel by providing
them, but signatures are an incremental improvement that
makes the product just a little bit better.

As a bit of background, webhooks are a very convenient
transport mechanism for a real time update, but also
provide fairly weak security by default -- if someone
discovers the URL where you receive your webhooks, it's
fairly trivial for them to exploit it by posting their own
fake events.

Previously, to check authenticity we'd recommended that
users make a request back to the Stripe API with any event
IDs they receive. This model is perfectly secure if
followed correctly, but there's no forcing function to
ensure that our users are actually making that second
request, and it also adds an undesirable second HTTP call.

Signed webhooks are better.

Hopefully this won't come across too self indulgent, but I
can promise that beyond a few code reviews, I had no hand
in the hard work that it took to get this across the finish
line.

## Batteries included (#batteries-included)

A key fundamental to any good developer experience is to
_make things easy_. We've tried to do that by providing
built-in signature verification functions for every
programming language that we support (currently at five out of
six, with Go support coming very soon).

Our libraries provide methods that will verify an incoming
webhook and construct a Stripe event out of the result if
the check was successful:

``` ruby
event = Stripe::Webhook.construct_event(
  payload, sig_header, endpoint_secret
)
```

The idea here is to make checking webhook authenticity as
easy as possible by needing only a single method to do so.
In fact, it's easier to check for a valid signature than it
is not to check, because the method also hydrates an
incoming webhook to a usable event model.

## Help with future proofing (#future-proofing)

You can talk about how something *might* change in the
future and that you should tolerate it, but unless there's
something more substantial to look at, they're empty words.

This is especially difficult with matters related to
cryptography (like a webhook signature). The only reason
that we'd ever have a reason that you'd ever change your
current scheme is if a weakness is detected in it, and that
could take years to happen, or maybe never. This is made
doubly difficult because if that's ever the case, you
really need people to upgrade expediently.

Stripe signature schemes have an associated version (the
current being `v1`) so that we're forwards compatible in
case we need to change something down the line. That in
itself isn't notable, but what's more interesting is that
in testmode we provide a fake `v0` scheme to test against.
This lets developers more easily build their integrations
to be tolerant of the existence of other schemes in keeping
with [the robustness principle][robustness] (_"be
conservative in what you do, be liberal in what you accept
from others"_).

Here's an example of the `v1` scheme and fake `v0` scheme
bundled up into a single header value:

```
Stripe-Signature: t=1492774577,
    v1=5257a869e7ecebeda32affa62cdca3fa51cad7e77a0e56ff536d0ce8e108d8bd,
    v0=6ffbb59b2300aae63f272406069a9788598b792a944a07aba816edb039989a39
```

## Hacker friendly (#hacker-friendly)

Our libraries will be a good fit for most users, but not
for every user; some people will want to implement
signature verification themselves because they don't want
to pull in a Stripe library as a dependency, or they're using
a language that we don't support.

We've tried to support their case by providing precise
instructions on [how to verify a webhook signature
manually][verify-manually] including parsing the signature
header, calculating an expected signature, and verifying
that the included timestamp is fresh. From a cryptographic
standpoing we're using HMAC SHA-256 which has the very nice
property of being knowingly strong, but also common enough
that implementations will be available across a broad set
of languages and frameworks.

[libraries]: https://stripe.com/docs/webhooks#verify-official-libraries
[robustness]: https://en.wikipedia.org/wiki/Robustness_principle
[verify-manually]: https://stripe.com/docs/webhooks#verify-manually
