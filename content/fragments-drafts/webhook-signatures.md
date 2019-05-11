+++
hook = "TODO"
published_at = 2017-05-10T02:59:44Z
title = "Stripe webhook signatures & DX"
+++

I'm really excited to say that last week we released
support for signed webhooks. This has been a feature that
users have requested for a long time, and which we've
wanted to provide forever. Signatures aren't a novel
feature given that the idea is widespread amongst other
APIs that provide webhooks, but they're an incremental
improvement that we think makes the product a little better
and a little easier to use.

## Webhook verification models (#verification-models)

Webhooks are a convenient transport mechanic for a real
time update, but their story around security is weak by
default. If someone discovers the URL where you receive
your webhooks, it's easy for them to exploit it by posting
their own maliciously constructed events.

Previously, to check authenticity we'd recommended that
users make a request back to the Stripe API with any event
IDs they receive. It's a model that's secure if followed,
but there's no forcing function to ensure that users run
that check. It also adds another round trip over the
network, which is undesirable if it can be avoided.

With signed webhooks users start with a secret that's share
with Stripe. It's used along with a cryptographic function
to generate a signature for a received payload which can
then be compared against Stripe's own signature that we've
calculated and injected through a header as we post the
webhook. The webhooks is verified conclusively without an
extra round trip.

The system includes a few minor niceties around developer
experience (DX) that I want to highlight.

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
begin
  event = Stripe::Webhook.construct_event(
    payload, signature_header, endpoint_secret
  )
rescue Stripe::SignatureVerificationError
  $stderr.puts "Invalid webhook!"
end
```

The idea here is to make checking webhook authenticity as
easy as possible by needing only a single method to do so.
In fact, it's easier to check for a valid signature than it
is not to check, because the method also hydrates an
incoming webhook to a usable event model -- a function that
you'd otherwise have to write yourself.

## Help with future proofing (#future-proofing)

You can talk about how something *might* change in the
future and that you should tolerate it, but unless there's
something more substantial to look at, they're empty words.
If the test of that forwards compatibility happens well
beyond when the code was originally written, there's a good
chance that it's not going to work.

This is especially difficult with matters related to
cryptography. The main reason to change an existing scheme
is if a weakness is detected in it, and that could take
years. This is made doubly difficult because if that's ever
the case, you really want people to be able to upgrade
expediently.

Stripe signature schemes have an associated version (the
current being `v1`) to give us room to add a new
implementation in case we have to later. That in itself
isn't notable, but what's a little more interesting is that
in testmode we send a fake `v0` scheme to test against.
This lets developers more easily build their integrations
to be tolerant of the existence of other schemes, even if
no other schemes exist yet. This way programs can be built
in keeping with [the robustness principle][robustness]
(_"be conservative in what you do, be liberal in what you
accept from others"_).

Here's a `v1` signature along with a fake `v0` signature
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
that the included timestamp is fresh. We're using HMAC
SHA-256 which has the very nice property of being
cryptographically secure, but also common enough that
implementations will be available across a broad set of
languages and frameworks.

## Verify today! (#verify-today)

We hope that you'll give our new signature system a try.
Head over to [our docs][signatures] for full instructions
on how to use them.

[libraries]: https://stripe.com/docs/webhooks#verify-official-libraries
[robustness]: https://en.wikipedia.org/wiki/Robustness_principle
[signatures]: https://stripe.com/docs/webhooks#signatures
[verify-manually]: https://stripe.com/docs/webhooks#verify-manually
