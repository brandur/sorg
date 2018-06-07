---
title: Living APIs, and the Case for GraphQL
location: San Francisco
published_at: 2018-06-05T13:51:26Z
hook: Why it makes sense to model APIs as graphs, and what
  GraphQL can do for us to help with discovery, batch
  operations, and gradual enhancmeent.
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

Finally, GraphQL is **typed**. Types often come in the form
of complex objects (e.g., `Customer`) or JSON scalars
(e.g., int, string), but the type system also supports more
sophisticated features like enumerations, interfaces, and
union types. Nullability is baked right in, which works out
incredibly well when building APIs in languages that don't
allow null (like Rust) because every field comes out as
non-nullable by default which makes handling API responses
much less prone to error.

!fig src="/assets/graphql/propeller.jpg" caption="This is a stretch. Let's call it something that looks like a graph, kind of."

## The graph (#graph)

As its name would suggest, GraphQL models objects as a
graph. Technically, the graph starts with a root node that
branches into query and mutation nodes, which then descend
into API-specific resources.

GraphQL just takes existing API paradigms to a logical
conclusion. Almost every REST-ish API that exists today is
already a graph, with resources referencing other resources
by IDs, or links for the APIs which most dogmatically
adhere to the principles of REST. Making these
relationships explicit just makes sense, but it also lets
consumer get their work done with fewer API calls.

Our API at Stripe has a concept called [object
expansion][expand] that lets a user tell the server that it
would like an ID (e.g., `cus_123`) expanded into its full
object representation by passing an `expand[]=...`
parameter in with the request. Expansions are chainable, so
I can use ask for `charge.customer` on a dispute to reveal
the associated charge, and that charge's customer. The
feature's most common effect is saving API calls -- instead
of having to request two objects separately, just one
request can be made for the first object with the second
embedded. Users love this feature -- we constrain
expansions to three levels deep, but get regular requests
to allow up to four levels.

## Discovery and exploration (#discovery)

A core challenge of every API is how to make it
approachable to new users, and providing interactive way to
explore them and make ad-hoc requests is a good way to
address this problem. GraphQL provides an answer to this
in the form of [GraphiQL][graphiql], an in-browser tool
that lets users read documentation and build queries.

I'd highly recommend taking a look at Shopify's [public
installation][shopifygraphiql] and trying some for
yourself. Remember to use the "Docs" link in the upper
right to pop open and explore the documentation.

!fig src="/assets/graphql/graphiql.png" caption="Using GraphiQL to explore an API and graph."

A vanilla installation of GraphiQL is a more powerful
integration tool for users than what 99% of REST-ish
providers have, and it's available automatically (modulo a
little configuration for authentication, CORS, etc.), and
for free.

It's also worth remembering that GraphiQL's features are
built right onto the standard GraphQL introspection
primitives -- it's just an HTML and JavaScript file that
can be hosted statically. For a big provider, building a
custom version of it that's tailored to the features and
layout of a specific API is well within reason.

## Batch operations (#batch)

Every sufficiently long-lived web API that responds to user
feedback will eventually evolve a batch API.

In REST-ish APIs, that involves building a custom batch
specification because there's nothing even close to wide
standardization for such a thing. Users adapt to each
exotic implementation by reading a lot of documentation. In
GraphQL, batch queries are built right in. Here's a
document containing multiple operations on the same query
and which uses aliases (`customerA`, `customerB`) so that
the results can be disambiguated in the response:

``` js
customerA: getCustomer(id: "cus_123") {
  email
}

customerB: getCustomer(id: "cus_456") {
  email
}
```

Batch mutations are also allowed.

The availability of this feature doesn't necessarily give
users free reign the ability to run costly batch requests.
Remember that as an API provider, you can still put
restrictions on this within reason. For example, by
allowing only five operations per request (if that's the
right fit for you), or even just one.

## Explicitness and graceful enhancement (#explicitness)

I mentioned above how fields in GraphQL must be requested
explicitly and that there's no SQL-like glob operator to
get everything. This might be GraphQL's most interesting
feature because it lends itself so well to API versioning
and enhancement.

In a REST API, an API provider must assume that for any
given API resource, _every_ field is in use by every user
because they have no insight at all into which ones
consumers are actually using. Removing any field must be
considered a breaking change and [an appropriate versioning
system][versioning] will need to be installed to manage
those changes.

In GraphQL, every contract is explicit and observable.
Provides can use something like a [canonical log
line](/canonical-log-lines) to get perfect insight into the
fields that are in use for every request, and use that
information to make decisions around product development,
API changes, and retirement. For example, when introducing
a new field, we can explicitly measure its use over time to
see how successful it is. Alternatively, if we notice that
a field is only in use by a tiny fraction of users and it
fits poorly into the API's design or is expensive to
maintain, it's a good candidate for deprecation and
eventual removal.

### Living APIs (#living-apis)

The REST model of little insight tends to produce APIs with
a strong tendency to ossify, with broad and abrupt changes
made intermittently with new versions. GraphQL produces an
environment that evolves much more gradually.

Fields that need to be phased out can be initially hidden
from documentation by marking them with GraphQL's built-in
`deprecated` annotation. From there, provides may choose to
even further restrict them by gating in users who were
already consuming them, and disallowing use for everyone
else, with an automatic process to remove gating as users
stop using them organically. After some deprecation period,
their use can be analyzed, and product teams can either
start an active outreach campaign for retirement, or remove
them entirely.

Similarly, new fields are introduced one at a time and
their adoption can be observed immediately. Like a living
thing, the API changes little by little. New features are
added and old mistakes are fixed. Its surface trends slowly
towards perfect maturity.

!fig src="/assets/graphql/living.jpg" caption="In the ideal case, we produce APIs that grow and improve like living things. My hands were really cold when I shot this."

## Leverage and convention (#leverage)

GraphQL introduces many powerful ideas, and because it was
written in response to extensive real-world experience, it
addresses scaling problems that most would-be API designers
wouldn't realize were problems until it was too late.

It comes with a [comprehensive spec][spec] to help avoid
ambiguities. The result is that most GraphQL APIs look very
similar and features are widespread throughout all common
implementations. I'd personally like to see its designers
take an even more opinionated stance on conventions like
naming and pagination, but even without, it's still a far
more sophisticated set of constraints than what we have
with REST. This forced consistency leads to leverage in the
form of tools like GraphiQL (and many more to come) that
can be shared amongst any of its implementations.

REST's momentum may appear to unstoppable, but its
underdesign and loose conventions leave a lot to be
desired. We'd do ourselves a favor to keep our gaze on
the horizon.

[1] Common wisdom is that GraphQL at Facebook is
sequestered to internal APIs only, although the public API
they do offer is graph-like and could fairly be called
proto-GraphQL.

[expand]: https://stripe.com/docs/api#expanding_objects
[githubpost]: https://githubengineering.com/the-github-graphql-api/
[graphiql]: https://github.com/graphql/graphiql
[graphqllist]: https://github.com/APIs-guru/graphql-apis
[intro]: https://graphql.org/learn/
[shopifygraphiql]: https://help.shopify.com/api/custom-storefronts/storefront-api/graphql-explorer/graphiql
[spec]: http://facebook.github.io/graphql/
[versioning]: https://stripe.com/blog/api-versioning
