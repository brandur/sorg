---
title: Webhooks, Operability, and the Future of Streaming APIs
published_at: 2017-06-06T03:00:11Z
location: San Francisco
hook: TODO
---

The term "webhook" was coined back in 2007 by Jeff Lindsay
as a "hook" (or callback) for the web; meant to be a
general purpose system to allow Internet systems to be
composed in the same spirit as the Unix pipe. By speaking
HTTP and being symmetrical to common HTTP APIs, they were
an elegant answer to a problem without many options at the
time -- WebSockets wouldn't be standardized until 2011 and
would only see practical use much later, and more
contemporary streaming options were still a only speck on
the horizon.

At Stripe, webhooks may be one of the best known features
of our API. They can used to configure multiple receivers
that receive customized sets of events, and they also work
for connected accounts, allowing platforms build on Stripe
to tie into the activity of their users. They're useful and
are a feature that's not going anywhere, but are also far
from perfect, especially from an operational perspective.
In the spirit of avoiding [accidental
evangelism](/accidental-evangelist), lets briefly discuss
whether they're a good pattern for new API providers to
emulate.

## User ergonomics (#user)

While think the stronger case against webhooks can be made
from the perspective of an operator, it's worth noting
first of all that the user and developer experience around
webhooks isn't without its own problems. None of them are
insurmountable, but alternatives might get more things
right more easily.

### Endpoint management (#endpoints)

Getting an HTTP endpoint provisioned to receive a webhook
isn't technically difficult, but it can be occasionally
frustrating.

The classic example is the large enterprise where getting a
new endpoint exposed to the outside world might be a
considerable project involving negiotiations with
infrastructure and security teams. In the worst cases,
webhooks might be wholly incompatible with an
organization's security model where user data is
uncompromisingly kept within a secured perimiter at all
times.

**TODO:** Cartoon of perimiter security.

Development is another difficult case. There's no perfectly
fluid way of getting an endpoint from a locally running
environment exposed for a webhook provider to access.
Workarounds like Ngrok are the best option, but still add a
step and complication that wouldn't be necessary otherwise.

### Uncertain security (#security)

Because webhook endpoints are publicly accessible HTTP
APIs, it's up to providers to build in a security scheme
to ensure that an attacker can't issue malicious requests
containing forged payloads. There's a variety of common
techniques:

1. ***Webhook signing:*** Sign webhook payloads and send the
   signature via HTTP header so that users can verify it.
2. ***HTTP authentication:*** Force users to provide HTTP
   basic auth credentials when they configure endpoints to
   receive webhooks.
3. ***API retrieval:*** Provide only an event identifier in
   webhook payload and force recipients to make a
   synchronous API request to get the full message.

!fig src="/assets/webhooks/signing-secrets.png" caption="Endpoint signing secrets in Stripe's dashboard."

Although security is possible, the fundamental problem with
webhooks is that it's difficult as a provide to _ensure_
that your users are following best practices (not the case
for synchronous APIs where you can mandate certain keys and
practices). Of the three options above, only the third
guarantees strong security; even if you provide signatures
you can't know for sure that your users are verifying them,
and if forced to provide HTTP basic auth credentials, many
users will opt for weak ones.

### Development and testing (#development)

It's relatively easy to provide a stub or live testmode for
a synchronous API, but a little more difficult for
webhooks because the user needs some mechanic to request
that a test webhook be sent.

At Stripe, we provide a "Send test webhook" function from
the dashboard. This provides a reasonable developer
experience in that at least testing an endpoint is
possible, but it's quite manual and would be difficult to
integrate into a CI suite (for example).

!fig src="/assets/webhooks/send-test-webhook.png" caption="Sending a test webhook in Stripe's dashboard."

### No ordering guarantees (#order)

Transmission failures, variations in latency, and quirks in
the provider's implementation mean that even though
webhooks are sent to an endpoint roughly ordered, there are
no guarantees that they'll be received that way. A lot of
the time this isn't a big problem, but it does mean that a
"delete" event for a resource could be received before its
"create" event, and consumers must be able to tolerate this
sort of inconsistency.

Ideally speaking, a real time stream would be reliable
enough that a consumer could use it as an [ordered
append-only log][log] which could be used to manage state
in a database. Webhooks are not this system.

### Version upgrades (#versioning)

For providers that version their API like we do at Stripe,
version upgrades can be a problem. Normally we allow users
to explicitly request a new version with an API call so
that they can verify that their integrations work before
upgrading their account, but with webhooks the provider has
to decide on the version. Often this leads to users trying
to write code that's compatible across multiple versions,
and then flipping the upgrade switch and praying that it
works (often it doesn't and the upgrade has to be rolled
back).

Once again, this can be fixed with great tooling, but
that's more infrastructure that a provider needs to
implement for a good webhook experience. We recently added
a feature that lets users configured the API version that
gets sent to each of their webhook endpoints, but for a
long time upgrades were much more awkward.

!fig src="/assets/webhooks/upgrade-version.png" caption="Upgrading the API version sent to a webhook endpoint in Stripe's dashboad."

## Operator downsides (#operator)

### Retries (#retries)

### Misbehavior is on the provider (#misbehavior)

### Chattiness and communication efficiency (#chattiness)

## A strength: load balancing (#balancing)

## What's next (#whats-next)

### GraphQL subscriptions (#graphql)

### GRPC streaming RPC (#grpc)

## The future (#future)

[log]: https://engineering.linkedin.com/distributed-systems/log-what-every-software-engineer-should-know-about-real-time-datas-unifying
