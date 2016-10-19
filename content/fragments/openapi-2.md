---
title: OpenAPI 2.0
published_at: 2016-10-16T20:52:26Z
hook: Some thoughts on OpenAPI 2.0 after putting together a definition for
  Stripe's API.
---

Over at Heroku I spent some time and effort [spreading the good word on JSON
Hyper-schema](/elegant-apis) to programmatically describe APIs. I'm still a
strong believer in having a declarative schema that strongly defines an API's
contract, but since then I've come to believe that Hyper-schema has a few
problems including apparent stagnation, general lack of adoption, and poor
support for some commonly found API patterns like URL parameters.

[OpenAPI 2.0][openapi] (maybe still better known as Swagger 2.0) seemed to be a
good replacement candidate because while retaining many of the nice properties
of Hyper-schema, including JSON schema-based request and response definitions,
it lacked all of its most existential problems.

I added [OpenAPI 2.0 support to Committee][committee] and spent some time
putting together a reasonably complete representation of Stripe's API. Overall
the spec is quite good and easy to implement, especially if you're already
familiar with JSON schema. I especially appreciated the relatively terse syntax
and consideration for all parts of an HTTP request including error returns,
headers, and query and URL parameters.

The spec isn't not completely problem-free though, and along the way I ran into
a few major difficulties:

* It doesn't allow endpoints to respond with more than one type of resource.
  [Reference][responses].
* It doesn't allow complex types in endpoint request parameters.
  [Reference][parameters].

Neither problem exists in Hyper-schema which manages to avoid them while
simultaneously being quite a bit more more simple than the rather elaborate
OpenAPI specification. In the end I worked around them with some fairly
undesirable hacks, but came through with a finished product of reasonable
quality.

For anyone thinking about OpenAPI adoption, my recommendation is to wait a bit
longer. A new version of OpenAPI (probably 3.0) is currently under active
development and is slated to contain fixes for all my grievances with the
current version, and will also come with quite a few other nice features.

[committee]: https://github.com/interagent/committee/pull/101
[openapi]: https://github.com/OAI/OpenAPI-Specification
[parameters]: https://github.com/OAI/OpenAPI-Specification/issues/717
[responses]: https://github.com/OAI/OpenAPI-Specification/issues/270
