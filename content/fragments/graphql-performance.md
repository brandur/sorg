+++
hook = "Dispelling the common misconception that GraphQL APIs are inherently non-performant."
published_at = 2017-04-25T13:53:38Z
title = "GraphQL APIs and performance"
+++

I published an article on [API paradigms](/api-paradigms) a
few weeks ago that touched fairly heavily on GraphQL.
Reactions were generally very positive, but a misconception
came up more than a few times around how GraphQL isn't
appropriate for public APIs because it implies unknown and
unbounded considerations around backend performance.

It's true that performance can be a challenge, but it's
not true that GraphQL somehow forces you to expose routes
that are not performant. A GraphQL schema is designed the
same way as your traditional REST-ish API in that a service
operator only reveals routes where they can ensure good
quality of service for users, and good stability for their
own backend.

For a general feel of the API that GraphQL implementations
provide, here's the most basic possible example from the
graphql-js README:

``` js
var schema = new GraphQLSchema({
  query: new GraphQLObjectType({
    name: 'RootQueryType',
    fields: {
      hello: {
        type: GraphQLString,
        resolve() {
          return 'world';
        }
      }
    }
  })
});
```

Note the `resolve` function. The library doesn't figure out
how to map onto your data's internal structure, _you tell
it how to_. [Here's a more complex example that shows the
same thing][star-wars].

Of course, a schema allowing rich data access will be
beneficial for its users, but operators have the freedom to
increase that complexity at their own pace, ensuring that
revealed queries, subqueries, or fields are suitably
performant as they go. The API is powerful, and just like
in REST, a service's internal structure stays decoupled
from its public API.

[star-wars]: https://github.com/graphql/graphql-js/blob/c87fbd787c2b04c478f9535225d56cfea5a710cc/src/__tests__/starWarsSchema.js#L258-L293
