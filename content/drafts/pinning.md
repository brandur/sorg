---
title: "Living Interfaces: Why to Use User Pinning for
  Versioning Web APIs"
location: San Francisco
published_at: 2018-01-24T19:05:30Z
hook: TODO
---

## Explicit versioning (#explicit)

## Implicit versioning through user pinning (#implicit)

Seems less elegant

## The choice matters more for public APIs (#public-apis)

You can often roll forward on private APIs, but it's hard
on public.

## Major upgrades and API ossification (#ossification)

Changes need to be big enough to justify a new version
(there's a constant contention between whether a change is
worthwhile and the effort around maintaining a new version
line).

### Short hops with implicit (#minor-upgrades)

If something's broken, just fix it! You'll need to maintain
old users, but all new users have a more cohesive product.

## Graceful enhancement through living APIs (#living-apis)

Users are likely to be only using a subset of any
non-trivial API.

By understanding their request profiles, it's possible to
determine whether a chance will affect them (paths only).

In many cases, you should be able to upgrade users through
most changes automatically. No effort on their part.

## Living APIs with GraphQL (#graphql)

Even better than REST implicit versioning because fields
are requested explicitly.

Light SDKs.

## Maintain only one version (#one-version)

Major upgrades are painful to users, which is a good enough
to avoid them already, but they're also painful to service
operators -- without major deprecation projects, people
don't leave old versions.
