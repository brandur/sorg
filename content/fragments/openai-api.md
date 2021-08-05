+++
hook = "How to closely follow Stripe's API and documentation conventions, but iterate on the bad stuff."
published_at = 2021-07-29T16:28:00Z
title = "API design showcase: OpenAI"
+++

In API design (as in many things) it's a good idea to keep a north star in mind -- a few companies that do it better than everybody else, and whose conventions can be used as a baseboard the next time you build something greenfield.

For years, the go-to example of good API design has been Stripe, which along with a powerful API, is also well-known for being a major innovator with its [API reference documentation](https://stripe.com/docs/api) and client libraries.

Stripe's still good, but many other parties have been catching up. These days, for my money, one of the best example of good API design in the wild is [OpenAI](https://beta.openai.com/docs/api-reference/answers/create).

You'll immediately notice similarities between OpenAI's documentation and Stripe's. OpenAI and Stripe have considerable interrelated history, and very likely that was a factor in OpenAI drawing influence from Stripe. This is a big part of what makes OpenAI's API design good -- it reuses many of Stripe's conventions, but refines the faults in Stripe's API that have gone unaddressed.

{{FigureSingleWithClass "OpenAI API documentation." "/photographs/fragments/openai-api/openai-api-docs.png" "overflowing"}}

## Common conventions (#common-conventions)

Notice these commonalities between the OpenAPI and Stripe APIs:

* All objects have an `id` field. IDs get an object prefix to help disambiguate them. e.g. `file-123` on OpenAI versus `file_123` on Stripe.
* All objects have a `type` field indicating their unique type name. e.g. `answer`, `classification`, `list`, etc.
* Most objects have a `created` timestamp. Not that unusual, but like Stripe this is a Unix epoch integer rather than RFC3339, the latter being the more common format choice these days.
* All list objects are top-level objects (rather than arrays), and nest a `data` array.
* A `/v1/` in the path, largely an ornamental embellishment which like Stripe's, isn't likely to be incremented.

They're subtle, but there are enough ways to design an API that these are clear signs of inspiration.

## Refinements (#refinements)

But OpenAI's carefully chosen places to diverge from Stripe's conventions to improve on them:

* Request payloads are encoded as JSON instead of `application/x-www-form-urlencoded`. I'll write more on this separately, but the latter, although human-convenient in some cases, was a major pain point within Stripe over the years.
* All parameters are clearly documented with a type that indicates whether they're supposed to be a `string`, `integer`, `boolean`, etc.
* Optional parameters that will get a default have that default clearly shown. This is obviously useful, and amazingly rare.
* A clear distinction is made between path parameters, query parameters, and parameters that belong in the request body.
* High-fidelity examples of what requests and responses are supposed to look like which are clearly human-curated. Compare this to Stripe where these exist, but which vary wildly in quality. (A known problem internally, but one which was difficult to fix for a variety of reasons.)

{{FigureSingleWithClass "An example API response from OpenAI." "/photographs/fragments/openai-api/example.png" "overflowing"}}

* Every request parameter gets lovingly crafted and detailed documentation. Even with Stripe, many parameter docs are largely unchanged from whatever was thrown together by the engineer who originally introduced the feature, with little follow up refinement. Other APIs are much worse.

Note that most of these improvements are related to API documentation rather than the API itself. This goes to show that as long an API has a decent baseline quality to it, it's the documentation that'll really make it shine.

## @todo (#todo)

OpenAI's API quality is great, but not perfect. Some noteworthy blemishes:

* Perhaps the most glaring (and oddest) omission is that response objects aren't documented.
* Even given a relative newness of the API, some inconsistencies are already sneaking in -- created timestamps are named `created` or `created_at` depending on the object.
* No pagination off the get go, something that might be painful to bring in later.
* Still no client library support beyond Python. This isn't so much an OpenAI thing as an everyone-but-Stripe thing -- there are still remarkably few good answers for generalized client library generation.

Still, nothing wildly terrible. Even with them OpenAI rests comfortably in the top few percent of well-designed web APIs.
