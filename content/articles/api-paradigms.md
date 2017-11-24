---
title: Is GraphQL the Next Frontier for Web APIs?
published_at: 2017-03-29T21:00:36Z
location: San Francisco
hook: Musings on the next API technology, and whether REST-ish
  JSON over HTTP is just "good enough" to never be displaced
  in a significant way.
hn_link: https://news.ycombinator.com/item?id=14003134
---

For a long time the API community spent a lot of effort
evangelizing [hypermedia][hypermedia], which promised to
make web APIs more discoverable through clients that could
follow links like a human, and more future proof by
re-using existing hypertext and HTTP semantics like `<a>`
links and 301 redirects.

But hypermedia had a very hard time building any kind of
traction. At least some of the trouble was technical, but
its biggest problem wasn't that it wasn't useful, but that
it wasn't useful _enough_. It may have had some nominal
advantages over the REST-ish JSON APIs that are seen most
often today, but they weren't valuable enough to justify
the extra overhead.

After five years of strong evangelism at every API
conference in the world and very little actual adoption,
it's a pretty safe bet that hypermedia isn't going to be
the next big thing. But that leads us to the question of
what will be? Does the world even need a new API paradigm?
REST-ish JSON over HTTP has proven itself to be quite
practical and resilient to obsolescence; it might just be
"good enough" to reign supreme for a long time to come.

## DX is paramount (#dx)

As a service provider, it might be tempting to think that
your choice of backend technology is going to make a big
difference to your users, but isn't necessarily true. As
long as you meet a minimum bar of non-offensiveness, this
is almost never the case (SOAP and OAuth 1 being two
examples of technologies that don't).

If there are reasonable tools available to keep the
developer experience (DX) in integrating pretty good, users
tend to be flexible. Avoid anything that's too heavy,
exotic, or obnoxious, and you'll find that your users will
care about the quality of your documentation far more than
they do about the technology you use.

## REST is okay (#rest)

Roy Fielding's original ideas around REST are elegant and
now quite widespread, but it's worth considering that the
paradigm's actual advantages in APIs for developers are
fairly unremarkable. Practically speaking, REST's strongest
points are its widespread interoperability (every language
has an HTTP client) and its **conventions** [1]. URLs are
usually resources. Resources often have standard CRUD
operations that are mapped to HTTP verbs like `PATCH` and
`DELETE`. Status codes usually convey information. Commonly
needed API mechanics like authentication and encoding are
sometimes integrated into standard HTTP headers like
`Authorization` and `Content-Encoding`. This is all very
good; convention allows a developer to learn something once
and then re-use that information to figure out how other
things will probably work.

<figure>
  <div class="table-container">
    <table class="overflowing">
      <tr>
        <th>Action</th>
        <th>HTTP Verb</th>
        <th>URL</th>
      </tr>
      <tr>
        <td>Create</td>
        <td><code>POST</code></td>
        <td><code>/customers</code></td>
      </tr>
      <tr>
        <td>Replace</td>
        <td><code>PUT</code></td>
        <td><code>/customers/:id</code></td>
      </tr>
      <tr>
        <td>Update</td>
        <td><code>PATCH</code></td>
        <td><code>/customers/:id</code></td>
      </tr>
      <tr>
        <td>Delete</td>
        <td><code>DELETE</code></td>
        <td><code>/customers/:id</code></td>
      </tr>
    </table>
  </div>
  <figcaption>The conventions of REST. URLs are resources
    and CRUD maps to HTTP verbs.</figcaption>
</figure>

If convention in REST has one problem, it's that there
isn't enough of it. I use words like _usually_, _often_,
and _sometimes_ above because although these practices are
recommended by the spec, they may or may not be followed.
In real life, most APIs are REST-ish at best. At Stripe for
example, our resource updates should use `PATCH` instead of
`PUT`, but for historical reasons they don't, and it's
probably not worth changing at this point. Developers will
need to read the documentation anyway, and they'll find out
about our ubiquitous use of the `POST` verb when they do.

REST also has other problems. Resource payloads can be
quite large because they return everything instead of just
what you need, and in many cases they don't map well to the
kind of information that clients actually want, forcing
expensive `N + 1` query situations. This is especially bad
for clients on more limited networks like mobile, where
bandwidth and bad latency conspire to make REST an
expensive proposition.

## What's next (#whats-next)

While the world sticking to the status quo for a long time
to come is a strong possibility, the inefficiencies of REST
might mean that there's room for something to come next.
Let's explore a few possibilities.

### GraphQL (#graphql)

[GraphQL][graphql] is a fan favorite right now. Conceived
at Facebook, it's gotten some significant traction from at
least one "big API" adopting it in the form of GitHub. More
exciting though is the organic uptake, with many smaller
companies with a better opportunity to greenfield starting
with it instead of REST.

!fig src="/assets/api-paradigms/graphql.jpg" caption="GraphQL produces an API that can be queried in complex ways."

It has many advantages: built-in introspection so
developers can use tools to navigate through an API they're
about to use. Data can be organized in a free form way that
doesn't necessarily tie it to heavy resources. A
well-organized graph can allow a client request the exact
set of data they need to do their work, with little waste
in number of calls or payload size. Imagine having to send
only a single API request out to load a page instead of
dozens. It's great for service operators too, because its
[explicitness allows them to get a better understanding of
exactly what their users are trying to do](/api-upgrades).

But GraphQL's future is still uncertain. Most notably,
Facebook itself hasn't adopted it for their public API,
which brings into question their commitment to the idea.
Also, despite its strengths, it might find itself in a
similar place as hypermedia in that its edge just isn't
worthwhile enough to a large enough audience who are more
than happy to keep using REST-ish JSON.

Possibly most importantly, GraphQL might not provide the
superior developer experience that many of us are looking
for. At the end of the day it involves writing up a big
query blob with minimal typing and guarantees on the
client-side. It's structured, but individual operations
still need to be looked up in a query explorer or reference
documentation. Its closest analog is SQL; although this is
a technology that many of us use every day, maintaining SQL
query blobs are a painful enough experience that most of us
turn to ORMs to wrap them. The same could certainly be done
for GraphQL, but in that case we might have to ask
ourselves just how much it's really accomplished for us.

### RPC (#rpc)

A very strong argument could be made that if most APIs are
REST-**ish** instead of REST-**ful**, and assuming that
most of the conventions that we're actually using boil down
to making URLs consistent and basic CRUD, then just maybe
REST really isn't buying us all that much. It may be an
elegant idea, but as a developer my foremost value is ease
of integration; an API's ideological integrity is a distant
tertiary concern.

One possibility for the next big paradigm in APIs is just
to make them a set RPCs (remote procedure calls) and use
something like [GRPC][grpc] to generate libraries in a huge
variety of languages that users can pull down and use at
their leisure. Under GRPC, data moves around in the
efficient protocol buffers format over HTTP2, making
transport performance of a GRPC-based API good by default,
and with no effort required on the parts of either users or
operators.

A well-organized set of RPC methods could offer strong
enough conventions to be competitive with what REST gives
us (e.g. `create_charge()`, `update_customer(id:)`,
`delete_subscription(id:)`), and as a developer be just as
pleasant to use.

!fig src="/assets/api-paradigms/rpc-vs-rest.svg" caption="By dropping a resource-based world view, we can better map endpoints to the user actions."

Dropping REST's resource-based world view also has the
advantage of letting designers better map their endpoints
to the actions that their users are really trying to take.
For example, if almost everyone who creates a customer in
your API will want to immediately create a charge for that
customer, then those two operations could be rolled up into
one despite requiring two separate steps in REST (1.
`POST /customers`, 2. `POST /charges`).

But the real world-shifting power of RPC would only become
apparent if you could get _everyone_ to agree on it.
Imagine if AWS, GitHub, Stripe, etc. were all on GRPC, and
integrating with a new API was as simple as downloading a
new set of protobuf definitions and writing code
immediately because all your supporting infrastructure
(i.e. initialization, libraries, configuration, ...) was
already in place and ready to go.

### Bespoke clients (#bespoke-clients)

Going back to the idea of developer experience being of
utmost importance, it may be that the future of big APIs
are custom-built libraries in all the major languages that
their users care about. The maintenance overhead of this
route is obviously significantly worse for service
providers, but these libraries could be designed according
to the local conventions of each language, making them a
pleasure to integrate with.

!fig src="/assets/api-paradigms/bespoke.jpg" caption="Some tooling to help create a bespoke leather product."

Strong typing could be used to make sure that the compiler
catches as many bugs as possible without a round trip to
the API server even needed. For example, we could make it
impossible to make an API call unless an API key was
provided, or require that an email parameter be provided to
create a new user. Documentation could be provided within
the language's own ecosystem (Godoc for example), or even
inline while writing code if there's a good IDE.

In this world, the API's design (i.e. whether it's on REST,
GraphQL, etc.) would be totally opaque to end users and
purely up to the discretion of the library maintainers.
Maybe function invocations translate directly into REST
requests, but the library could also compile each one into
a specially crafted GraphQL query (or mutation) to maximize
its network efficiency.

## Maximizing productivity through convention and tooling (#productivity)

In the end, we shouldn't forget what REST-ful APIs did for
us in terms of providing a set of conventions that helped
us be more productive because what we knew was transferable
as we looked at new APIs.

GraphQL, RPC, and bespoke clients are all pretty plausible
ways forward to a post-REST world, but whichever we choose,
we shouldn't forget the lessons that REST taught us in that
convention and widespread consistency are powerful things.
If we do adopt something new, we should aim to make it as
ubiquitous as possible so that we don't worsen developer
experience by fracturing technologies. Unfortunately
though, realistically that might mean just sticking to
REST.

Here's one final strong opinion: religious adherence to
REST is overrated and its perceived advantages have never
materialized as fully as its proponents hoped. Whatever
we choose next should aim to be flexible and efficient, and
GraphQL seems like a good candidate for that. We should
move to GraphQL as a backend and combine it with great
language-specific libraries that leverage good type systems
to catch integration mistakes before the first HTTP call
flies, and which allow developers to use their own tooling
to auto-complete (in the sense of VS IntelliSense or Vim's
YouCompleteMe) to success.

[1] I realize that REST is designed to provide much greater
    facilities in the form of discovery and content
    negotiation, but in practice these just don't see a lot
    of use, which is why I normally say that convention is
    REST's strongest attribute.

[grpc]: http://www.grpc.io/
[graphql]: http://graphql.org/
[hypermedia]: https://en.wikipedia.org/wiki/Hypermedia
