---
title: The Case for GraphQL
location: San Francisco
published_at: 2018-06-05T13:51:26Z
hook: TODO
---

It's hard to read exactly where GraphQL stands in the API
world right now. Available publicly since 2015, it
continues to generate a lot of interest in developer
communities, but API trends aren't measurably moving in its
favor.

By far its biggest third-party public proponent is GitHub,
who released the fourth version of their API as GraphQL in
2016 with an [engineering post][githubpost] speaking about
it very favorably (and condemning the hypermedia
implementation of the previous version). It also has a few
other vocal users in the form of Shopify and Yelp, both of
whom offer GraphQL APIs, but other big providers are harder
to find. [This repository][graphqllist] keeps a list of
publicly available GraphQL APIs, and almost every
well-known API provider is notably absent, including
Facebook themselves [1].

Most publicly proffered APIs are still "REST-ish" -- with
resources and actions offered over HTTP -- including those
from almost every name you'd recognize in the space:
Amazon, Dropbox, Google, Microsoft, Stripe, and Twilio.
Momentum plays a huge part in that the pattern is
widespread and developers are used to it both on the parts
of integrators using APIs, and those who are building them.
A few arguments are still posited that strict adherence to
REST and hypermedia will open a wide world of automatic
discoverability and adaptation, but total lack of real
world precedent seems to be a strong empirical suggestion
that these are will-o'-the-wisps.

GraphQL's biggest problem is quite possibly that although
it's better, it's not "better enough". The bar set by a
REST-ish API is low, but it's high enough to work, and is
adequate for most purposes.

I've been doing a lot of thinking about what the next
generation of web APIs will look like (or if there will be
one at all), and I for one, would like to see more GraphQL.
Here I'll try to articulate a few arguments for why it's a
good idea that go beyond the surface-level selling points
on the tin.

## The surface (#surface)

I'll defer to the [official introduction][intro] as a good
resource to get familiar with GraphQL's basic, but it has a
few important core ideas that are worth touching upon.

With GraphQL, fields and relationships must be requested
**explicitly**. Here we ask for a customer object including
the `currency`, `email`, and `subscriptions` fields:

``` js
getCustomer(id: "cus_123") {
  currency,
  email,
  subscriptions
}
```

There's no wildcard operator like a `SELECT *` in SQL.
Compared to REST, this has an advantage of reducing payload
size (especially helpful for mobile), but more importantly,
it establishes a very explicit contract between the client
and server which allows APIs to be evolved more gracefully.
We'll talk about this more below.

GraphQL is automatically **introspectable**. By using the
special `__type` operator, any client can get a detailed
understanding of a type and all its fields:

``` js
{
  __type(name: "Customer") {
    name
    fields {
      name
      type {
        name
      }
    }
  }
}
```

Every common implementation supports introspection and
tooling can be built to rely on it being available. Unlike
REST, there's no need to retrofit description languages
like OpenAPI which are sometimes available and often not
completely accurate.

Finally, GraphQL is also **typed**. Types often come in the
form of complex objects (e.g., `Customer`) or JSON scalars
(e.g., int, string), but the type system also supports more
sophisticated features like enumerations, interfaces, and
union types. Nullability is baked right in, which works out
incredibly well when building APIs in languages that don't
allow null (like Rust) because every field comes out as
non-nullable by default which makes handling API responses
much less prone to error.

!fig src="/assets/graphql/village.jpg" caption="This is a stretch: the relationships within in a town are a graph – of people."

## The graph (#graph)

As its name would suggest, GraphQL models objects as a
graph. Technically, the graph starts with a root node that
branches into query and mutation nodes, which then descend
into API-specific resources.

## Exploration (#exploration)

GraphiQL is powered by built-in GraphQL introspection
primitives that come right in the spec and are common to
every common implementations (it’s just a static HTML and
JS file). There’s no need for meta-tools like OpenAPI to
retrofit this on.

OpenAPI

!fig src="/assets/graphql/graphiql.png" caption="Using GraphiQL to explore an API and graph."

## Batch operations (#batch)

Every sufficiently long-lived web API that responds to user
feedback will eventually evolve batch operations.

``` js
customerA: customer(id: "cus_123") {
  email
}

customerB: customer(id: "cus_456") {
  email
}
```

The availability of this feature doesn't necessarily give
users the ability to run costly batch requests with wild
abandon. Remember that as an API provider, you can still
put restricts on this within reason. For example, by
allowing only five operations per request, or even just
one.

## Explicitness and graceful evolution (#explicitness)

Canonical lines.

### Living APIs (#living-apis)

## Summary (#summary)

Consistency.

--> Leverage

[1] Common wisdom is that GraphQL at Facebook is
sequestered to internal APIs only, although the public API
they do offer is graph-like and could fairly be called
proto-GraphQL.

[githubpost]: https://githubengineering.com/the-github-graphql-api/
[graphqllist]: https://github.com/APIs-guru/graphql-apis
[intro]: https://graphql.org/learn/
