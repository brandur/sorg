---
title: Living APIs, and the Case for GraphQL
location: San Francisco
published_at: 2018-06-08T19:26:48Z
hook: Why it makes sense to model APIs as graphs, and what
  GraphQL can do for us to help with discovery, batch
  operations, and gradual enhancement.
hn_link: https://news.ycombinator.com/item?id=17269028
---

It's hard to read exactly where GraphQL stands in the API
world right now. Available publicly since 2015, trends in
APIs aren't obviously moving in its favor, but not
obviously moving against it either. Interest from the
developer community has been steady throughout even if the
technology isn't spreading like wildfire.

Its biggest third party proponent is GitHub, who released
the fourth version of their API as GraphQL in 2016 with an
[engineering post][githubpost] speaking about it very
favorably. It also has a other vocal users in the form of
Shopify and Yelp, both of whom offer public GraphQL APIs.
But beyond those big three, other big providers are
somewhat harder to find. [This repository][graphqllist]
keeps a list of publicly available GraphQL APIs, and most
well-known API providers are notably absent, including
Facebook themselves [1].

Most publicly proffered APIs are still "REST-ish" -- with
resources and actions offered over HTTP -- including those
from almost every name you'd recognize in the space:
Amazon, Dropbox, Google, Microsoft, Stripe, and Twilio.
Momentum plays a huge part in that the pattern is
widespread and developers are used to it both on the parts
of integrators using APIs, and those who are building them.
Some arguments are still made that strict adherence to REST
and hypermedia will open a wide world of automatic
discoverability and adaptation, but lack of real world
precedent despite years of opportunity seems to be a strong
empirical suggestion that this vision is a
will-o'-the-wisp.

GraphQL's biggest problem may be that although it's better,
it's not "better enough". The bar set by REST is low, but
it's high enough to work, and is adequate for most
purposes.

I've been doing a lot of thinking about what a new
generation of web APIs would look like (or if there will be
one at all), and I for one, would like to see more GraphQL.
I'll try to articulate a few arguments for why it's a good
idea that go beyond the common surface-level selling
points.

## The surface (#surface)

I'll defer to the [official introduction][intro] as a good
resource to get familiar with GraphQL's basics, but it has
a few important core ideas that are worth touching upon.

With GraphQL, fields and relationships must be requested
**explicitly**. Here we ask for a user object including the
`currency`, `email`, and `subscriptions` fields:

``` js
getUser(id: "user_123") {
  currency,
  email,
  subscriptions
}
```

There's no wildcard operator like a `SELECT *` from SQL.
Compared to REST, this has an advantage of reducing payload
size (especially helpful for mobile), but more importantly,
it establishes an explicit contract between the client and
server which allow APIs to be evolved more gracefully.
We'll talk about this more below.

GraphQL is automatically **introspectable** online. By
using the special `__type` operator, any client can get a
detailed understanding of a type and all its fields and
documentation:

``` js
{
  __type(name: "User") {
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

Every common implementation supports introspection (it's
required in [the GraphQL spec][spec]) and tooling can be
built to rely on it being available. Unlike REST, there's
no need to retrofit an unstandardized description language
like OpenAPI (or its myriad of competitors). Even today,
these are usually not available, and often not completely
accurate because the description isn't tied directly to the
implementation.

Finally, GraphQL is **typed**. Types often come in the form
of complex objects (e.g., `User`) or JSON scalars (e.g.,
int, string), but the type system also supports more
sophisticated features like enumerations, interfaces, and
union types. Nullability is baked in, which happens to work
out incredibly well when building APIs in languages that
don't allow null (like Rust) because every field comes out
as non-nullable by default. This additional constraint
makes handling API responses more deterministic and less
prone to error.

!fig src="/assets/graphql/village.jpg" caption="The relationships between people in a town are a graph. This is a stretch (but I like this photo)."

## The graph (#graph)

As its name would suggest, GraphQL models objects as a
graph. Technically, the graph starts with a root node that
branches into query and mutation nodes, which then descend
into API-specific resources.

GraphQL takes existing API paradigms to a logical
conclusion. Almost every REST API that exists today is
already a graph, but one that's more difficult to traverse.
Resources reference other resources by IDs (or links in
APIs which most strongly adhere to the principles of REST),
and relations are fetched with new HTTP requests. Making
relationships explicit is conceptually sound, and lets
consumers get work done with fewer API calls.

Stripe's API has a concept called [object
expansion][expand] that lets a user tell the server that it
would like an ID (e.g., `cus_123`) expanded into its full
object representation by passing an `expand[]=...`
parameter in with the request. Expansions are chainable, so
I can ask for `charge.customer` on a dispute to reveal the
dispute's associated charge, and that charge's customer.
The feature's most common effect is saving API calls --
instead of having to request two objects separately, just
one request can be made for the first object with the
second embedded. Users make extensive use of this feature
-- we constrain expansions to three levels deep, but get
regular requests to allow up to four levels.

## Discovery and exploration (#discovery)

A core challenge of every API is making it approachable to
new users, and providing interactive way to explore them
and make ad-hoc requests is a great way to address that.
GraphQL provides an answer to this in the form of
[GraphiQL][graphiql], an in-browser tool that lets users
read documentation and build queries.

I'd highly recommend taking a look at Shopify's [public
installation][shopifygraphiql] and trying some for
yourself. Remember to use the "Docs" link in the upper
right to pop open and explore the documentation. You should
find yourself being able to build a query that delves 4+
relations deep without much trouble.

!fig src="/assets/graphql/graphiql.png" caption="Using GraphiQL to explore an API and graph."

A vanilla installation of GraphiQL is a more powerful
integration tool for users than what 99% of REST providers
have, and it's available automatically (modulo a little
configuration for authentication, CORS, etc.), and for
free.

It's also worth remembering that GraphiQL's features are
built right onto the standard GraphQL introspection
primitives -- it's just an HTML and JavaScript file that
can be hosted statically. For a big provider, building a
custom version of it that's tailored to the features and
layout of a specific API is well within reason.

## Batch operations (#batch)

Every sufficiently long-lived web API that responds to user
feedback will eventually evolve a batch API.

In REST APIs, that involves building a custom batch
specification because there's nothing even close to wide
standardization for such a thing. Users adapt to each
exotic implementation by reading a lot of documentation. In
GraphQL, batch queries are built right in. Here's a
document containing multiple operations on the same query
and which uses aliases (`userA`, `userB`) so that the
results are disambiguated in the response:

``` js
userA: getUser(id: "user_123") {
  email
}

userB: getUser(id: "user_456") {
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
explicitly and that there's no SQL-like glob operator
(`SELECT *`) to get everything. This might be GraphQL's
most interesting feature because it lends itself so well to
API versioning and enhancement.

In a REST API, an API provider must assume that for any
given API resource, _every_ field is in use by every user
because they have no insight at all into which ones they're
actually using. Removing any field must be considered a
breaking change and [an appropriate versioning
system][versioning] will need to be installed to manage
those changes.

In GraphQL, every contract is explicit and observable.
Providers can use something like a [canonical log
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
`deprecated` annotation. From there, providers may choose
to even further restrict their use by gating in users who
were already consuming them, and disallowing everyone else,
possibly with an automatic process to remove those gated
exceptions as users upgrade organically over time and move
away from those deprecated fields. After a long grace
period, their use can be analyzed, and product teams can
start an active outreach campaign for total retirement
before removing them entirely.

Similarly, new fields are introduced one at a time and
their adoption can be observed immediately. Like a living
thing, the API changes little by little. New features are
added and old mistakes are fixed. It trends towards
maturity incrementally in a distant perfect form.

!fig src="/assets/graphql/living.jpg" caption="In the ideal case, we produce APIs that grow and improve like living things. My hands were really cold when I shot this."

## Shared convention and leverage (#convention)

GraphQL introduces many powerful ideas, and because it was
written in response to extensive real-world experience, it
addresses API scaling problems that most would-be API
designers wouldn't think about until it was too late.

It comes with a [comprehensive spec][spec] to help avoid
ambiguities. The result is that most GraphQL APIs look very
similar and features are widespread throughout all common
implementations. I'd personally like to see its designers
take an even more opinionated stance on conventions like
naming, mutation granularity, and pagination, but even
without, it's still a far more sophisticated set of
constraints than what we have with REST. This forced
consistency leads to leverage in the form of tools like
GraphiQL (and many more to come) that can be shared amongst
any of its implementations.

REST's momentum may appear unstoppable, but underdesign and
loose conventions leave a lot to be desired. We'd be doing
ourselves a favor by keeping our gaze on the horizon.

[1] Common wisdom is that GraphQL at Facebook is
sequestered to internal APIs only, although the public API
they do offer is graph-like and could fairly be called
proto-GraphQL.

[expand]: https://stripe.com/docs/api/expanding_objects
[githubpost]: https://github.blog/2016-09-14-the-github-graphql-api/
[graphiql]: https://github.com/graphql/graphiql
[graphqllist]: https://github.com/APIs-guru/graphql-apis
[intro]: https://graphql.org/learn/
[shopifygraphiql]: https://help.shopify.com/en/api/custom-storefronts/storefront-api/graphql-explorer
[spec]: https://graphql.github.io/graphql-spec/
[versioning]: https://stripe.com/blog/api-versioning
