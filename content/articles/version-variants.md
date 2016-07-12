---
hook: A simple mechanism for managing changes to a web API and to help cheapen the
  disposal of prototypes.
location: San Francisco
published_at: 2015-02-17T22:13:48Z
title: Version Variants
---

An old problem with APIs of all kinds is that they're difficult to change. Although an API can be expanded without too much trouble, it becomes costly to remove any fields that may have been present previously in case some consumer has come to be dependent on them. As companies like Twitter found out, breaking API consumers on a consistent basis is the fastest way to foster a disparaging development community and a long-lasting infamy as an unreliable provider.

Modern web pundits might tell you to solve this problem with Hypermedia, but although this type of RESTful technique might provide some protection against the relocation of resources, it can do little to protect against the removal of fields on a resource or the removal of entire resource types.

Many providers try to give themselves a bit of freedom in this area by introducing versioning into their APIs so that in the case of required API changes the version can be bumped and old consumers can keep working. For example, we require calls to our modern API version to specify a version on every HTTP request via the `Accept` header:

    Accept: application/vnd.heroku+json; version=3

But when it comes to web APIs, versioning is problematic in its own way. Anytime that a version is incremented, an orphan is left behind. That orphan still has many consumers left on it, and will require considerable product and technical work in the form of a sophisticated deprecation schedule and strategy. To ensure that consumers aren't constantly chasing a moving target, that schedule may have to stay very conservative by allowing a grace period of months, or even years. For example, [our compatibility policy](https://devcenter.heroku.com/articles/api-compatibility-policy#production) states that production resources will remain available for twelve months after deprecation. Especially when it comes to building prototypes and experiments, this kind of expense associated with any kind of obsolescence is a hard pill to swallow.

## Variants (#variants)

To help improve on this situation, we introduced a concept that we've been using for a few months now called _version variants_. Variants are a simple way of hiding new API features behind a flag so that they stay out of the main API version. They have names that mirror their associated version that look like `version=3.new-feature` and are requested in a similar fashion:

    Accept: application/vnd.heroku+json; version=3.new-feature

Variants have a few important characteristics:

* **Additive:** Variants are purely additive. They can add new resources to the API or shadow existing resources by adding new fields to them, but they cannot remove fields on an existing resource. This is designed to mirror the guarantees that are required of any changes to the main API in that any kind of removal is considered a breaking change.
* **Explicit:** As seen above, requesting a version variant is a very explicit process in that all requests must specifically include the variant with every call. This helps signal to consumer that what they're requesting is a probably an experimental feature and as such, does not provide the same stability guarantees as they might expect from the mainline API.
* **Orthogonal:** Version variants are orthogonal to each other in that although they will include all features of the main API, they cannot be combined with other variants. This is designed to act as a forcing function to encourage variants to be pulled back into the mainline so that they can get access to new features. More importantly, it discourages developers from building spiderwebs of interconnected experimental features that depend on other experimental features to operate.

## Lifecycle (#lifecycle)

At their core, variants are a tool to ease the prototyping process by making the process of deprecating a prototype cheaper. Their common lifecycle looks a little like the following:

1. Fork a variant from mainline.
2. Continue to develop the prototype; introduce users and run experiments.
3. Finish the project by either:
    1. Declaring it widespread beta or GA: pull the variant into the mainline API.
    2. Declare the prototype obsolete: remove the variant and all associated implementation code.

We'll normally have an API engineer run a full review on the new APIs at the time of pulling a variant mainline, with only minimal manual guidance provided up to that point (we do of course encourage everyone to read our [general HTTP API design guidelines](https://github.com/interagent/http-api-design) before starting anything at all, and to come to us with any proposed designs that don't fit our existing patterns well). This helps to cheapen the cost of the prototype in that a team building a new feature doesn't have to swallow the process of an API audit with every change that they make to it.

Our API responds with mainline even for API variants that it doesn't know about (i.e. `3.*`). This makes the process of pulling variants to mainline safe in that even consumers that are still requesting the old variant have their requests filled appropriately until they can be updated.

In the latter case of a prototype's complete removal, some consumers may be broken just like if a major feature was removed from the mainline API, but hopefully the number of broken consumers will be fewer and the lowered stability expectations of those users will help them cope with the change. In any case, we'd still recommend announcing the deprecation at least a few weeks in advance to provide consumers with some grace time to help them react appropriately.

## Internal Expectations (#internal-expectations)

One anti-pattern that might manifest without careful consideration are prototypes in variants that are not made either generally available or deprecated appropriately, a common case for any project which is started but then loses steam and isn't finished. To help mitigate this, we're experimenting with requiring all variants to be assigned an expiry date, after which a variant may be removed liberally if the team that created it is no longer taking appropriate action to continue moving its lifecycle forward. This is modeled in part on the IETF's [guidelines for Internet drafts](http://www.ietf.org/ietf-ftp/1id-guidelines.txt) which require that an expiration date of 185 days from the date of submission is added to the first and last pages of any draft document.
