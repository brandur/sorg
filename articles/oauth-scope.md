---
hook: Designing scope for Heroku OAuth, and a brief tour of other implementations
  on the web.
location: San Francisco
published_at: 2013-07-23T14:54:44Z
title: Scoping and OAuth 2
---

Yesterday marked the beta release of [OAuth for the Heroku Platform
API](https://blog.heroku.com/archives/2013/7/22/oauth-for-platform-api-in-public-beta),
which we hope will empower users to develop apps against the API by providing a
simple and powerful authentication framework that's consistent with other
providers across the web.

One interesting discussion that developed while we were building this out was
around OAuth scoping, the mechanism that allows OAuth clients to tell an
authorization server what permissions they'll need on resources they're
accessing. I thought this might be a good opportunity to talk a little about
OAuth scoping, what the spec has to say about it, how it's implemented elsewhere
on the web, and our own design considerations.

## The Spec (#spec)

[RFC 6749](http://tools.ietf.org/html/rfc6749#section-3.3) describes how scope
should be implemented according to the proposed OAuth 2 standard. I've tried to
summarize the main points presented in the document:

* Scope is specified on either the authorization or token endpoints using the
  parameter `scope`.
* Scope is expressed as a set of _case-sensitive_ and _space-delimited_ strings.
* The authorization server may override the scope request, in this case it must
  include `scope` in its response to inform a client of their actual scope.
* When a scope is not specified, the server may either fallback to a
  well-documented default, or fail the request.

The spec describes the format that a scope should have and how the server
should handle it, but is open-ended in respect to what strings in a scope
should actually look like. This decision allows providers to define their own
strings, and gives them enough flexibility to ensure that OAuth 2 scoping is a
good fit for accessing a wide variety of different resources.

## From Around the Web (#samples)

The open-ended spec has resulted in all kinds of creative implementations
across the web, with no two being exactly alike. I've compiled a few examples
to demonstrate the range of ideas out there.

### App.net (#app-net)

App.net allows developers to define a basic set of scopes in snake_case. This
pretty standard scoping implementation is simple and effective.

    basic stream update_profile

http://developers.app.net/docs/authentication/#scopes

### Facebook (#facebook)

Facebook deviates from spec a bit by suggesting that scope strings be
comma-delimited. The two other interesting characteristics of Facebook scopes
are that more specific strings are namespaced under their broader category
(e.g. `user_actions.video`), and that some strings are dynamic (e.g.
`APP_NAMESPACE` scopes to a particular app in `user_actions:APP_NAMESPACE`).
Facebook also offers a very extensive variety of available scopes so that apps
can be very precise about what powers they'll require.

    email,read_stream,user_actions.video,user_actions:APP_NAMESPACE

https://developers.facebook.com/docs/reference/login/#permissions

### GitHub (#github)

GitHub provides a concise set of scopes with some namespacing using the colon
character. For example, `user:email` is a subset of the permissions allowed by
`user`.

```
gist repo user user:email
```

Another interesting innovation here is that for any API requests, GitHub passes
back the response headers `X-OAuth-Scopes` and `X-Accepted-OAuth-Scopes` to
indicate to the user what scope strings their token has, and what strings this
endpoint will accept. This makes their APIs self-documenting in that it
provides users an easy alternative to looking up documentation when choosing
scope for their apps and tokens.

http://developer.github.com/v3/oauth/#scopes

### Google (#google)

Google mandates that scopes should start with the `openid` string, then include
either or both of `email` and `profile`. From there, scope is extended across
Google's flourishing ecosystem by defining other strings as extensible URIs.

    openid profile email https://www.googleapis.com/auth/drive.file

https://developers.google.com/accounts/docs/OAuth2Login

### Instagram (#instagram)

Another fairly simple implementation, with the notable use of plus signs rather
than spaces for delimitation.

    likes+comments

http://instagram.com/developer/authentication

### LinkedIn (#linkedin)

LinkedIn reserves the underscore to separate types of resources from the
read/write permissions to that type, with an `r` specifying read privileges and
`w` write.

    r_basicprofile r_emailaddress rw_groups w_messages

https://developer.linkedin.com/documents/authentication#granting

### Salesforce (#salesforce)

An uncommon trait here is that Salesforce requires a particular scope string
for the privilege of being granted a refresh token.

    api refresh_token web

http://help.salesforce.com/help/doc/en/remoteaccess_oauth_scopes.htm

### Shopify (#shopify)

Shopify also mixes read and write permissions into scope strings. Their system
is fairly intuitive in that `write_` also implies `read_` permission, so that
developers don't need to specify both.

    read_customers write_script_tags, write_shipping

http://docs.shopify.com/api/tutorials/oauth

### Windows Live ID (#windows-live)

Defines scope strings that are prefixed with `wl.` (for Windows Live);
presumably so that scopes are unique across Microsoft's entire product space.

    wl.basic wl.offline_access wl.contacts_photos

http://msdn.microsoft.com/en-us/library/live/hh243646.aspx

## Heroku (#heroku)

The end product for Heroku OAuth scope was shaped by a few major product and
engineering design goals:

* The set of scope strings should be minimal so that we still have the power to
  evolve scoping as we continue to build out our product. Even if we completely
  redesign our scope strings, all the old strings should be general enough to
  easily map to the new system. Adding things is easy, but deprecating is hard.
* Some app resources are sensitive enough that even if a scope grants almost
  universal permission to manage an app, they still need to be protected. A
  good example of this are an app's config vars, which contain secrets like
  database connection strings. The scoping system must take this into account.
* We should provide a very minimal scope that provides basic user information
  and nothing else. This is useful in systems that will use OAuth to identify a
  user and little else like the [Heroku Forums](https://discussion.heroku.com).

Taking these goals into account, along with the spec and the web's other
implementations, we came up with a starting point for our scope system which is
what was released yesterday:

* `identity`: Allows access to `GET /account` for basic user info, but nothing
  else.
* `read`: Read access to all a user's apps and their subresources, except for
  protected subresources like config vars and releases.
* `write`: Write access to apps and unprotected subresources. Superset of
  `read`.
* `read-protected`: Read including protected subresources. Superset of `read`.
* `write-protected`: Write including protected subresources. Superset of
  `read-protected` and `write`.
* `global`: Global access encompassing all other scope.

These strings map to more a much more granular set of permissions in the
backend, which will allow us to continue evolving the public interface as need
be.

Like a few other providers, we also elected for self-documenting API endpoints
that help developers along by specifying their accepted scope strings as
response headers (we tend to drop any `X-` prefixes as [they're effectively
deprecated](http://tools.ietf.org/html/draft-ietf-appsawg-xdash-03)):

```
Oauth-Scope: global
Oauth-Scope-Accepted: global read read-protected write write-protected
```

This isn't a finalized design, and are looking forward to collecting
requirements from the community and internal consumers and iterating on it, the
hope being that we can provider a system which offers a powerful amount of
flexibility and granularity, while staying simple to use and true to the
original spec.
