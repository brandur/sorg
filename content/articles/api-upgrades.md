---
title: Why Doesn't Stripe Automatically Upgrade API Versions?
location: San Francisco
published_at: 2017-03-17T15:30:50Z
hook: Given a wide API surface area are small changes, most
  API upgrades are safe for most users. Could we upgrade
  their fixed API version automatically?
---

I got an interesting question while talking to a friend
last night about API versioning, "Why doesn't Stripe
automatically do API upgrades for its users?" The idea is
interesting enough that it's worth digging into a little
more.

Some readers may be aware already that [at Stripe we take a
somewhat unconventional][upgrades] "safety first" approach
to API versioning. A new version is introduced every time a
backwards incompatible change is made, and there's enough
of them that they're named after dates like `2017-02-14`.
There tends to be on the order of 10s of new versions in a
year.

The first time a user account makes a request to the API,
their account is automatically locked to the current
version of the API. They can override it by passing a
`Stripe-Version` header, but any requests without will
implicitly be assigned the version fixed to their account.

A **backwards incompatible** change is one that could
potentially break an existing integration. For example,
removing a field from a response, or changing its JSON
type. Most day-to-day changes like adding a new API
endpoint or a new field to an existing response are
considered forwards compatible, and a version isn't cut for
them.

## Most upgrades are safe (#upgrade-safety)

You can see by [perusing the API changelog][changelog] that
most changes are fairly minor. Upgrades can be painful and
time consuming for users, so we try our hardest to get the
design of the API right the first time. In cases that we
don't, the changes that are made are relatively minor. For
example:

* The response on `/v1/accounts` no longer returns the
  `currencies_supported` field.
* Disputes on charge resources used to be expanded by
  default, but are now collapsed without an explicit
  request for expansion.
* Requests with insufficient permissions now return a 403
  status code instead of 401.
* The `name` field under bank account responses was renamed
  to `account_holder_name`.

The Stripe API has a lot of surface area with dozens of
resources and ~130 endpoints. Most people are using only a
small subset of that, and any given change isn't likely to
affect them. Even if they are using an affected endpoint,
it's possible, and even likely, that they're not using any
of the fields that changed.

## Automatic upgrades (#automatic-upgrades)

If most upgrades are safe for most users, then it stands to
reason that we could potentially upgrade people
automatically so that they wouldn't have to do it
themselves. It would also make retiring older API versions
possible, which is desirable, but not currently done [1].

In at least some cases automatic upgrades would be
possible. We have good information around which users call
into which endpoints, and if we noticed that the endpoints
changed by newly introduced upgrade and the endpoints
called by a user were perfectly exclusive, we could roll
them forward onto a new version. From an example above, if
`/v1/accounts` is changing, but a user only creates charges
and customers, they could be upgraded.

But in other cases it's difficult: if disputes under charge
resources start getting collapsed by default and we know
that a user makes calls on charge endpoints, we can't
measure the safety of the upgrade. It's possible that it's
still perfectly safe because although they get charge
responses they never look at their disputes, but it's also
possible that they do, and we have no way of knowing.

Automatic upgrades are a great idea and they'd be a nice
feature, but too many changes fall into this ambiguous
area, so we don't.

## Other schemes (#other-schemes)

Regardless of design, all RESTful APIs will be more or less
stuck in the same place because it's so standard to respond
with the entire serialized form of API resources on any
request. Even [hypermedia][hypermedia], which theoretically
allows for greater flexibility through the use of HTTP
content negotiation and smarter clients, doesn't have an
answer for the problem.

That's not to say though that there aren't other ways.
[GraphQL][graphql] is a popular API paradigm that's seeing
some pretty good uptake. One of the things I really
appreciate about it is that it requires all fields to be
requested explicitly:

``` json
{
  human(id: "1000") {
    name
    height
  }
}
```

There is no equivalent to `SELECT * FROM ...`.

Even accounting for some error where users request fields
that they don't really need, this still gives you an
ability to profile incoming requests that's leaps and
bounds better than REST, and better flexibility as a
result.

[changelog]: https://stripe.com/docs/upgrades#api-changelog
[graphql]: http://graphql.org/learn/queries/
[hypermedia]: https://en.wikipedia.org/wiki/HATEOAS
[upgrades]: https://stripe.com/docs/upgrades

[1] Given the choice, most people won't be very proactive
    about upgrading (and really, why should they be?), so
    in practice old versions tend to have at least some
    usage forever.
