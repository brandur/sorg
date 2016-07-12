---
hook: ''
published_at: 2013-09-01T15:59:44Z
title: API Versioning
---

## Web API Versioning Schemes

### Broad Versioning

Either prefixed or through something like an `Accept` header.

It seems like a tempting idea, but in practice forcing users across a major
upgrade is so painful that it's avoided at all costs. GitHub is still V3.
Twitter is still on V1. Twilio is still on "2010-04-01". Heroku is still on V3
[1]. Stripe is still on V1 (although it its case the existence of prefix
versioning is largely vestigial).

### Implicit Versioning

### No Versioning

### Hypermedia

Despite years of aggressive and sustained evangelism, it's failed to gain much
of a foothold in any major APIs. It also has some major practical problems,
such as the increased expense due to its chatty nature.

## Stripe's Versioning Examined

Conveys major benefits, but has major downsides.

The traditional advantage conveyed to Stripe-style versioning is that it you
trade stability for your users for some additional pain on the part of the
operators (who need to maintain a large number of different API versions). In
retrospect, this is partly true, but the piece that's usually not mentioned is
that eventually user's still need to upgrade. And because your versioning
system has allowed them to stave this off for so long, the upgrade is
_especially_ painful.

[1] One could even argue that Heroku is still on V2 (which is still heavily
    used by the CLI and is the default version) because a major version upgrade
    turned out to be so difficult.
