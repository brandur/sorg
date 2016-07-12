---
hook: A set of patterns to make APIs more accessible to developers; lowering the barrier
  of entry for new users, and easing the maintenance of consuming applications.
location: San Francisco
published_at: 2013-09-01T15:59:44Z
title: Developer Accessible APIs
---

Many providers get users started off on their API by pointing them to an
extensive set of documents, designed to help build a conceptual model for those
users to leverage later when they start to build something with it. Because
asking users to consume significant documentation upfront is a lot to ask these
days, a common pattern is to include a set of examples and tutorials to help
get users up and running as quickly possible. These are effective and
time-proven techniques, but there's still room for improvement.  Consider these
problems:

* API behavior must be documented exhaustively as developers have no easy way
  of checking particular cases for themselves. This documentation is expensive
  to create and maintain, and without a rigid process to keep it updated, will
  inevitably fall into disrepair.
* Documentation will never be able to cover every possible case. At the end of
  the day, the best way to be sure of the API's behavior in some corner case is
  to try the API itself.
* Tutorials and examples must often make assumptions based on what languages
  and tools users are expected to use. Users left without relevant documents will
  often have to project one of the available documents onto their own toolset,
  for example a developer coding in Clojure and following a Ruby tutorial.

It's possible to solve some of these problems by ensuring that APIs are
**developer accessible**, meaning that rather than only optimizing for the case
of those applications that will be consuming them over the long run, they also
cater to those developers that are learning the API in order to build new
products with it. Developers using the API this way will be coming in and
making manual one-off calls with their tools of choice, before transcribing
those calls to the more permanent medium of whatever applications they build.

This kind of accessibility isn't just good for jumpstarting new developers
either. As the existing API economy becomes ever more prominent, we have to
consider that over a long enough period of time, changes in either clients or
the APIs themselves will inexorably lead to breakages. In such situations, we
should aim to make it as easy as possible for a developer to jump in with their
toolkit and quickly figure out what's going on so that the problem can be
remedied.

Developer accessibility is more of an idea than any particular method, and as
such there's no definite way of implementing it, but a number of patterns that
we see in the wild can help to illustrate the concept.

## Patterns (#patterns)

### OAuth 2 (#oauth2)

Much of the added complexity around OAuth 1 stems from the extra layer of
security that was built on top of it. OAuth 2 simply relies on HTTPS to take
care of this, and the result is a much more usable protocol. While OAuth 1 APIs
often need to be called through specialized consumer libraries, calls to OAuth
2 APIs can generally be made with any generic client, including plain old Curl,
which significantly lowers the barrier of entry to an API.

Even Twitter, well known for its restrictive APIs [has an easy way of procuring
an OAuth 2 access token](https://gist.github.com/brandur/5845931).

### Bearer Token Authorization (#bearer-tokens)

A very simple pattern for use with OAuth 2 is allowing users to authorize with
a bearer token via the `Authorization` header. This ensures that any client
that can send an HTTP header has an easy way in without needing to do
base64-encoding.

    curl -H "Authorization: Bearer 01234567-89ab-cdef-0123-456789abcdef" ...

### Curlability (#curl)

A consistent theme across many patterns is simply that an API should be
accessible to any generic HTTP client, with Curl occupying the place of that
baseline tool in many of our kits due to its relative ubiquity. Good Curl
accessibility is useful for both new developers who can start experimenting
with an API immediately, and for the API owners themselves, who can take
advantage of it during the development of new API features as well.

A very simple pattern of immediately improving an API's Curlabiliity is to
prettify JSON output for Curl clients as [I've previously
described](https://mutelight.org/pretty-json).

### In-response Scope Hints (#scope-hints)

It can be quite helpful to return metadata about the current request and the
current endpoint for a developer to digest while they're testing calls against
an API.

For example, a fairly general problem with APIs being consumed by OAuth-enabled
apps is that apps will often request scopes with more liberal permissions than
the app actually needs, which isn't ideal from a security perspective. By
returning a header like `OAuth-Scope-Accepted` below, we give developers an
easy way to determine what permissisions are needed on the endpoints they're
accessing, allowing them to lock down their scope before releasing an app.

    Oauth-Scope: global
    Oauth-Scope-Accepted: global identity

### In-response Ordering Hints (#ordering-hints)

For our [V3 platform
API](https://devcenter.heroku.com/articles/platform-api-reference) at Heroku,
list ordering is accomplished by specifying that order through a `Range`
header, but ordering can only be carried out on particular fields. Those fields
can either be looked up in the reference documentation, or a developer can
easily check which ones are supported by inspecting the `Accept-Ranges` header
that comes back with list responses:

    Accept-Ranges: id, name
    Range: id ..

### Ship a Service Stub (#service-stubs)

I've previously talked about how Rack service stubs can be [used to improve the
development and testing experience](https://brandur.org/service-stubs) of apps
that are heaviliy dependent on external APIs. An API can also ship its own
service stub, which allows developers to try API calls that might otherwise
mutate data when done in a production environment. See the [Heroku API
stub](https://github.com/heroku/heroku-api-stub) for an example of this
technique.

### Programmatic Maps (#programmatic-maps)

An interesting Hypermedia-related technique that's gaining some traction is to
provide a set of links at an API's root that point to other available
endpoints. Coupled with strong RESTful conventions, this might allow a
developer to skip the reference documentation completely by learning the API by
navigating around it with Curl.

Try GitHub's root to see this in the real world:

    curl https://api.github.com
    
    {
      "current_user_url": "https://api.github.com/user",
      "authorizations_url": "https://api.github.com/authorizations",
      "emails_url": "https://api.github.com/user/emails",
      "emojis_url": "https://api.github.com/emojis",
      "events_url": "https://api.github.com/events",
      "feeds_url": "https://api.github.com/feeds",
      "following_url": "https://api.github.com/user/following{/target}",
      "gists_url": "https://api.github.com/gists{/gist_id}",
      "hub_url": "https://api.github.com/hub",
      "issue_search_url": "https://api.github.com/legacy/issues/search/{owner}/{repo}/{state}/{keyword}",
      "issues_url": "https://api.github.com/issues",
      "keys_url": "https://api.github.com/user/keys",
      "notifications_url": "https://api.github.com/notifications",
      "organization_repositories_url": "https://api.github.com/orgs/{org}/repos/{?type,page,per_page,sort}",
      "organization_url": "https://api.github.com/orgs/{org}",
      "public_gists_url": "https://api.github.com/gists/public",
      "rate_limit_url": "https://api.github.com/rate_limit",
      "repository_url": "https://api.github.com/repos/{owner}/{repo}",
      "repository_search_url": "https://api.github.com/legacy/repos/search/{keyword}{?language,start_page}",
      "current_user_repositories_url": "https://api.github.com/user/repos{?type,page,per_page,sort}",
      "starred_url": "https://api.github.com/user/starred{/owner}{/repo}",
      "starred_gists_url": "https://api.github.com/gists/starred",
      "team_url": "https://api.github.com/teams",
      "user_url": "https://api.github.com/users/{user}",
      "user_organizations_url": "https://api.github.com/user/orgs",
      "user_repositories_url": "https://api.github.com/users/{user}/repos{?type,page,per_page,sort}",
      "user_search_url": "https://api.github.com/legacy/user/search/{keyword}"
    }
