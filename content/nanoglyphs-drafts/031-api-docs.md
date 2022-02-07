+++
image_alt = "Ocean Beach"
image_url = "/photographs/nanoglyphs/031-api-docs/ocean-beach-cropped@2x.jpg"
published_at = 2022-02-06T11:10:21Z
title = "Docs! Docs! Docs!"
+++

The past couple of weeks: _a lot_ of work building out and tightening up documentation, with special emphasis on API reference.

We've had an API reference live for a while, but before now it'd always been maintained manually. The loop was: hopefully a product person would remember that a new API feature was shipping, and would poke a dev rel person to update the docs. The dev rel person would write some documentation about a feature they understood by way of two hops worth of telephone, hit some endpoints with cURL, and try to capture what came back as best as they could.

You may be able to imagine that the results were a little rough. Maybe half our total API endpoints were documented. Amongst those that were, fields were often missing or extraneous as the responses had changed since the docs were originally written. Some of it was just flat out wrong.

But although imperfect, in a major sense it did what it was supposed to: serving as a stopgap long enough to give us a chance to build something better. That took getting a few foundational blocks laid -- a process that took about six months all in -- but we got there.

---

## Hyper-schema forever (#hyperschema)

This is the third major push to build API documentation that I've been involved in, each of which has produced a very different pipeline.

Heroku's API reference is found [on Devcenter](https://devcenter.heroku.com/articles/platform-api-reference). Here's how it gets built:

* The core API stack is written in Ruby. It's Ruby of reasonable calibre, but at no point was there a push to put in an API framework that might be able to describe itself. It was pretty vanilla -- a few dozen Sinatra modules composed together, and within each route requests and responses were generated in the normal imperative way. e.g.

    ``` ruby
    post '/apps' do
      req = parse_request
      app = create_app(req)
      respond generate_respnose(app)
    end
    ```

* In a time where web APIs were still hot and there were many competing API specifications (remember RAML? Blueprint?), we chose to write [JSON hyper-schema](https://json-schema.org/draft/2019-09/json-schema-hypermedia.html) to describe each endpoint, and over a period of many months, finished a description for the entire API. Hyper-schema didn't turn out to be a big winner in the spec wars, but at the end of the day, it's not _that_ different from OpenAPI (the latter makes heavy use of JSON schema, just not JSON _hyper_-schema), and it did the job. Hyper-schemas were written and maintained by hand.

* Because the schemas lived out-of-band of the API implementation, we needed a way to show that they were right as the implementation potentially diverged. I wrote [Committee](https://github.com/interagent/committee), which exposes test helpers that check test responses against the expected schema and flag inconsistencies. Committee might be my most successful open-source project ever in the sense that although I am not a good maintainer, its maintenance was picked up by a Japanese team who _are_ good maintainers, and who've been keeping a close eye on it ever since.

* A tool called [Prmd](https://github.com/interagent/prmd) transformed JSON hyper-schema into Markdown/HTML. The rendered Markdown was transferred into Devcenter.

<img src="/photographs/nanoglyphs/031-api-docs/heroku-api-reference@2x.png" alt="Screenshot of Heroku's API reference" class="wide" loading="lazy">

All in all, it wasn't a bad process, and especially given the state of the art at the time (this was ~2013), it wasn't half bad -- I would've pitted what we had against any other Silicon Valley companies with public APIs at the time.

But there were obvious downsides. A few that come to mind:

* Manually writing hyper-schema was time consuming and awful. It (along with other API specifications including OpenAPI) is much more suitably read _and written_ by a machine.

	  We'd also decided to be purists about the whole thing, and the JSON hyper-schema was _really written in JSON_, as opposed to a more human-friendly format like YAML or TOML. Put a brace in the wrong place? Forget a comma? Extra comma? Gods have mercy on your soul, friend.

* Furthermore, the whole process was just pretty manual overall. Even once the schema was written, it had to be maintained in step with API changes, with changes manually pushed to Prmd and then published through Devcenter.

---

{{NanoglyphSignup .InEmail}}

## AbstractAPIMethod (#abstract-api-method)

That brings us to Stripe, and the well known [stripe.com/docs/api](https://stripe.com/docs/api), a URL that I'll forever remember better than my own name.

Here's the rough loop there:

* By the time I got to the company, they'd already written a Ruby-based DSL that allowed API endpoints to describe themselves. It wasn't the best DSL ever created, and not particularly novel by today's standards, but it was _way_ ahead of its time. The DSL was introspected dynamically and combined with a healthy dose of highly custom Ruby code and ERB templates to render `/docs/api`.

* The DSL turned out to be a godsend because it allowed us to repurpose it and write our first API spec generators. The API was already pretty big, and having to retrofit a self-describing format would've been practically impossible, even back in 2015.

    Someone had already written a rough generator for JSON hyper-schema (possibly inspired by our evangelization out of Heroku) -- it loaded the Ruby codebase, iterated over every endpoint, and descended into each to pull request/response structure, types, documentation, and a host of other nick-nacks. I took it over the finish line and used the same framework to write an OpenAPI 2 generator (it was more apparent by then that JSON hyper-schema wasn't going to be a big long term winner). Later, when OpenAPI 3 was finalized, we wrote a third generator for that -- OpenAPI 2 was limited in quite a number of ways that prevented it from being able to fully express an API -- especially one with polymorphic responses like ours. Stripe's OpenAPI specs are still [publicly published to GitHub](https://github.com/stripe/openapi) to this day.

* At first, the OpenAPI spec was only used to write better test suites for our API libraries, which previously had major quality problems. But over the next few years, it came to be used almost everywhere:

    * [Flow](https://flow.org/) JavaScript bindings were generated from it for use in the Stripe Dashboard (frontend UI).
		
    * Later, it was used to build a private GraphQL endpoint which integrated well with various React state management frameworks the frontend team used for the next version of the Dashboard platform.
		
    * After a major push (and one that was years in coming), it was used to generate endpoint-specific code for Stripe's seven languages worth of API libraries. Previously these had all been manually maintained as API changes went live -- a job so colossal that it was nothing short of a miracle that it ever worked.
		
    * And of course, it eventually came to be used for `/docs/api` as well after it was rewritten from the ground up in 2019.
		
OpenAPI was a bit of an unauthorized skunkworks project when we started it, but it turned out to be one of the most high-leverage technical investments the company ever made. (Although I'll note that it's not anything particular to OpenAPI that makes this true -- the important part is to just have some kind of intermediary format that lives between the backend implementation and the generators that build derivatives.)

<img src="/photographs/nanoglyphs/031-api-docs/stripe-api-reference@2x.png" alt="Screenshot of Stripe's API reference" class="wide" loading="lazy">

It wasn't bad, but the process was far from perfect:

* The spec itself became a huge multi-megabyte file. It became difficult for a human to find anything in it, and it eventually became desirable to generate a JSON version in addition to the normal YAML due to considerations around how long it took to parse it in tests (YAML is very slow). We also had it checked into our Git repo, and it was a regular source of painful merge conflicts.

* CI would check to make sure you'd regenerated OpenAPI for API changes you'd made. CI was brutally slow (see [the Path of Madness](/nanoglyphs/029-path-of-madness)), and it was roughly the worst feeling ever to get build results back ten minutes later, only to find you'd forgotten to run one command.

* It was a huge API using slow technology stacks. Generating OpenAPI and derivatives took somewhere in the neighborhood of ~30 seconds to run -- brutally slow for a development loop.

* We never quite got producing good API resource examples right. There was a generator that could produce a sample object for a given API resource, but it was a shambling monstrosity of spaghetti code and produced very low quality results. We wanted to write something better, but it was a colossal job by the time we thought about it as there were 100s of complex resources to handle.

---

## Automating end-to-end (#end-to-end)

So that brings us to my latest stack at Crunchy, where having learnt from old mistakes, I was determined to build something not only accurate, but also fast, fluid, and as automatic as it could be.

Here's how it works:

* Go's HTTP primitives are very low level, so we built a light framework on top of them that's capable of defining endpoints, and also knows what requests, responses, and status codes they're expected to return.

* Along with a reflect package, Go's standard library ships with [one that can read docstring comments](https://pkg.go.dev/go/doc) on structs and fields, used to generate Godoc. We take advantage of it to have docstrings on structs for endpoints, requests, and responses act as canonical documentation for OpenAPI and doc.

    ``` go
    // Networks are a multi-provider abstraction of what would otherwise be called a
    // virtual private cloud (VPC) in most cloud providers. In a nutshell, they're
    // an encapsulated network where a cluster may be located which has a strong
    // boundary and which can specify its own rules around egress and ingress.
    type NetworkEndpointGroup struct {
        *apiendpoint.APIEndpointGroup
    }
    ```

    Docs could have been alternatively stored as Go strings, but using docstrings is especially useful for struct fields, where there'd be no clean way to attach a string to each one otherwise.

    ``` go
    // Network contains information on a virtual network.
    type Network struct {
        // Unique ID of the network.
        ID eid.EID `json:"id" validate:"required"`

        // A subnet block specifying the network's location and possible addresses
        // in CIDR4 (IPv4) notation.
        //
        // This property is only available after a network is fully created and will
        // be `null` early in a new network's lifecycle.
        CIDR4 *string `json:"cidr4" validate:"-"`

        // A subnet block specifying the network's location and possible addresses
        // in CIDR6 (IPv6) notation.
        //
        // This property is currently Owl-only and not yet released.
        CIDR6 *string `json:"-" validate:"-" openapi:"hide"`

        // Human-readable name of the network.
        Name string `json:"name" validate:"required"`

        // ID of the provider on which the network is located.
        ProviderID string `json:"provider_id" validate:"required"`

        // ID of the region where the network is located.
        //
        // Networks on GCP aren't affixed to a specific region and return the
        // special value of `global` in this field.
        RegionID string `json:"region_id" validate:"required"`

        // ID of the team which the network belongs to.
        TeamID eid.EID `json:"team_id" validate:"required"`
    }
    ```

* A program separate from the main API server initializes the full set of API endpoints, and iterates over them to produce OpenAPI. We have one spec for public endpoints, and a second including internal endpoints as well.

* The OpenAPI generator runs in a GitHub Action that's part of CI, and on a successful `master` build, stores the result to a publicly-accessible location.

* Another Go program in our docs repo ingests the OpenAPI spec and uses it to produce the API reference. It runs on a GitHub Action cron which will [autocommit any changes to the repo](/fragments/self-updating-github-readme).

* Our docs are generated by [Hugo](https://gohugo.io) (a popular static site generator), and a Heroku deploy pipeline automatically pushes new changes live every time a CI build succeeds.

* Back in API code, structs for API requests and responses implement a `SchemaExampler` interface with which they can generate a high quality sample for themselves. This is stored to OpenAPI's `example` field for their schemas, and pushed all the way through to docs. Importantly, these examples are stable, meaning that they're not changing around every time the spec is generated. So our GitHub Action job only commits new changes when there are actually API changes, not every time we build.

    ``` go
    var exampleNetwork = &Network{
        ID:         SampleNetworkID,
        CIDR4:      ptrutil.String("1.2.3.4/24"),
        Name:       "crunchy-production-network",
        ProviderID: "aws",
        RegionID:   "us-west-2",
        TeamID:     SampleTeamID,
    }

    func (*Network) SchemaExample() interface{} {
        return apiexample.ValidateAndMarshalToMap(exampleNetwork)
    }
    ```

* A small-but-worthwhile nuance is that the doc generator generates Hugo-friendly Markdown rather than HTML. That might not sound like much of a difference, but through a combination of CSS and use of [Hugo short codes](https://gohugo.io/content-management/shortcodes/), it's possible for non-engineers to iterate on the API ref's design without having to run any Go themselves. Instead, they just make changes to CSS or a short code's definition.

* In addition to doc generation, the OpenAPI spec is also used to generate TypeScript bindings for use in the Bridge Dashboard (our frontend UI), which is entirely API-driven. Generating language bindings is a pretty complicated beast, so we use a publicly available generator so as not to have write/maintain our own.

So to recap: on a successful merge to `master`, CI pushes a new OpenAPI to the web. A separate GitHub Action wakes up, runs the doc generator and commits any changes. That commit triggers a Heroku deployment and pushes the changes live. Aside from the initial merge on GitHub, no human intervention is required at any point. The result looks [like this](https://docs.crunchybridge.com/api/network).

<img src="/photographs/nanoglyphs/031-api-docs/bridge-api-reference@2x.png" alt="Screenshot of the Crunchy Bridge API reference" class="wide" loading="lazy">

We're also taking speed seriously. The OpenAPI generator runs in less than a second and that's without any optimization effort on my part. The docs generator is even faster. Go compiles quickly, so even when I'm iterating on either program, it's always fast. No thirty-second development loops in sight, and gods help me there never will be.

But while it's the best API reference pipeline I've been involved with yet, nothing is perfect. A few things that come to mind:

* We use Go docstrings for documentation. Go doc convention dictates that the name of the struct/field start a docstring like `// SchemaExample produces a schema example`. That sucks, so we don't do it, which means that our docstrings that go to API don't follow convention. Not a huge problem, but worth pointing out.

* When reading Go code, it's not necessarily obvious which doc strings will be emitted publicly and which ones won't. Once you understand the process it's pretty apparent (any docs on an endpoint, request, or response are going to OpenAPI), but it's hard for a new contributor to know this happening.

* Go's lack of support for some higher-level language features make some things difficult. For example, Go has no concept of an enum, but we would like enums in our OpenAPI/docs. We're able to accomplish it through a complicated introspection process that involves looking for "enum-like" types, and that works pretty well, but writing it wasn't exactly easy.

## A dash of humanity (#humanity)

So while generated docs are great from an effort standpoint, a hill I'm willing to die on is that even when generated, all docs should include a healthy dose of humanity. The computer handles iterating over endpoint/struct/field ad nauseum, but a human should augment what the machine would do to add as much background and context as possible.

Here's an example of the worst kind of documentation, unfortunately all to common all over the computing world:

``` go
// GenerateHTTPResponse generates an HTTP response.
func GenerateHTTPResponse([]byte, error) {
		...
}
```

Oh so _that's_ what `GenerateHTTPResponse` does. Hallelujah -- I was lost, but now I'm found. That documentation isn't just of no value, it's actually of negative value because someone might see there's documentation on a function and go there to read it, only to realize they've completely wasted their time.

So where possible, I encourage my peers to write docstrings that aren't just useful to us, but have enough context that they'd be useful to external users as well.

Check out the [Keycloak REST API](https://www.keycloak.org/docs-api/15.0/rest-api/index.html) for an example of what inhuman API documentation looks like -- exhaustive, but with zero context on what anything is or what it does. Frustratingly context-free in every possible way. I'm aiming very explicitly for our docs _not_ to look like that.

## Necromancers, by Scott Ridley (#necromancers)

Like sci-fi, want a TV recommendation, and still trust me after my horrible over-optimistic _Wheel of Time_ review from a few months back? [_Raised by Wolves_](https://en.wikipedia.org/wiki/Raised_by_Wolves_(American_TV_series\)), which just started airing season two. Directed by Ridley Scott, this is the purest science fiction to make its way to a big budget production in years, and one of the precious few original ideas to be found in modern culture. It's seriously crazy -- flying snakes, acid oceans, androids performing ad-hoc facial reconstructive surgery, Travis Fimmel reprising his role as Ragnor Lothbrok -- nothing makes sense, yet there's just enough there to make me believe there's a method to the madness. I never have any idea what's going to happen next.

Until next week.
