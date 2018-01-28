---
title: "Living Interfaces: Why User Pinning Is Good for
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

A while ago I wrote a program that would run on a home
server and periodically update a `CNAME` record managed by
CloudFlare with the current value of my dynamic IP address.

Almost two years ago, they sent me an email saying that
they were [deprecating their V1 API in favor of the new V4
API][sunset] on November 9th, 2016. My program is small,
but not particularly important, and the prospect of
relearning an entire API to update it was a poor value
proposition. Even though I knew there was a high likelihood
that it would stop working, I did nothing.

Today in 2018, my program still works. It seems that at the
end of their deprecation period CloudFlare still had enough
users on V1 that they decided to leave it on.

Getting users to go through a major API upgrade is _hard_.
Even if you're deprecation is more successful than
CloudFlare's, designing and carrying out a sunsetting
strategy and running associated user support is going to be
a time-consuming affair.

When API designers are thinking about changes to their API,
this is always going to be at the back of their minds. Is
the change important enough to cut a new API version? The
usual answer is "no", so the current API is left with a
subpar design, albeit one that's backward compatible.

Most of those changes will be batched up and released all
at once as a new version eventually, but size of that
accumulated mass will paradoxically make any user upgrades
even harder because moving up to a API version is all or
nothing.

### Short hops with implicit (#minor-upgrades)

If something's broken, just fix it! You'll need to maintain
old users, but all new users have a more cohesive product.

## Graceful enhancement through living APIs (#living-apis)

Users are likely to be only using a subset of any
non-trivial API.

By understanding their request profiles, it's possible to
determine whether a change will affect them (paths only).

In many cases, you should be able to upgrade users through
most changes automatically. No effort on their part.

### Living APIs with GraphQL (#graphql)

Even better than REST implicit versioning because fields
are requested explicitly.

Light SDKs.

## Maintain only one version (#one-version)

Major upgrades are painful to users, which is a good enough
to avoid them already, but they're also painful to service
operators -- without major deprecation projects, people
don't leave old versions.

[sunset]: https://blog.cloudflare.com/sunsetting-api-v1-in-favor-of-cloudflares-current-client-api-api-v4/
