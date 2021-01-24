+++
hook = "A new Go package for the WaniKani API, transitioning to pkg.go.dev, and what makes APIs nice to use."
published_at = 2021-01-23T00:12:01Z
title = "A WaniKani API Go package"
+++

I wrote Go language bindings for the [WaniKani API](https://docs.api.wanikani.com/20170710/):

https://github.com/brandur/wanikaniapi

Their design is largely derived from my time with [stripe-go](https://github.com/stripe/stripe-go), but with the peripheral resource-specific packages dropped and the removal of anything resembling package-level configuration -- all API calls are made by first instantiating a `wanikaniapi.Client`.

It has most of the usual features you'd expect from a client library: automatic pagination, contexts, error handling, automatic retries, etc. The big one I wanted to get right was decent documentation, so the README contains snippets for most common things people will need to do.

## READMEs, back en vogue (#readmes)

A recent development in the Go ecosystem is that documentation is migrating from [godoc.org](https://godoc.org/) to [pkg.go.dev](https://pkg.go.dev/) early this year. The two systems are similar, but an important difference is that pkg.go.dev documentation now includes the contents of a README when present in the project.

{{Figure "pkg.go.dev screenshot showing included README." (ImgSrcAndAltAndClass "/assets/images/fragments/wani-kani-go-api-package/pkg-go-dev.png" "pkg.go.dev screenshot showing included README." "overflowing")}}

This is a great change. Previously package authors had the option of putting detailed usage information in (1) the package documentation, (2) the README, or (3) in both, and trying to juggle any updates between them. Many Go authors opted for (1), which was particularly frustrating as a user because you'd land on a GitHub project page with no information, and be expected to link through to Godoc. [READMEs are awesome](https://tom.preston-werner.com/2010/08/23/readme-driven-development.html), and packages authors can now flesh them out with useful information while having that reflected in package documentation as well.

Unfortunately snippets in Go documentation are not compiled ([unlike Rust](https://doc.rust-lang.org/rustdoc/documentation-tests.html)), so I wrote [testable examples](https://blog.golang.org/examples) for each one, and then copied those back into the docs. It'll still be possible for the documentation versions to fall out of date, but this maximizes the chances they're right to start, and I'll try to keep examples updated in lockstep with documentation.

## WaniKani (#wani-kani)

WaniKani proves to be a very good test ground for API-related projects. It's a pretty typical REST-ish API, with [great documentation](https://docs.api.wanikani.com/20170710/), good self-consistency, is minimal with about a dozen total resources, and is stable as the product is more or less "done".

It's also got some nice quality of life and efficiency niceties like support for `If-None-Match` and `If-Modified-Since`, large page sizes, and list endpoints that all take an `updated_after` parameter so that clients can request just what's changed since their last run.

While designing Heroku's V3 API, we gave ourselves an edict that the API should be powerful enough that Heroku's own products like the Dashboard could be implemented on it, with no help from private endpoints or other internal features. WaniKani's API apparently had a similar design goal, and it's allowed fully-featured 3rd party clients to be developed, like [Tsurukame on iOS](https://github.com/davidsansome/tsurukame).

All of this put together makes WaniKani one of the best examples of a good API that I know of, and I'll surely be citing it in future work.
