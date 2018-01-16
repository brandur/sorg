---
title: "Lambda: Why Go Is the Perfect Fit For Serverless"
location: San Francisco
published_at: 2018-01-16T16:57:55Z
hook: TODO
---

Yesterday, in a move foreshadowed by a hint at Amazons'
re:Invent conference late last year, [AWS released support
for Go on its Lambda platform][announce].

This means that Go users can now compile programs against
the [`aws-lambda-go` SDK][sdk], which contains typed
structs representing Lambda event sources and common
responses. This can then be compiled, bundled up into a
["Lambda deployment package"][package] (as simple as a ZIP
file with a binary in it), and added to a new function
(where "Go 1.x" is now selectable as a runtime):

!fig src="/assets/go-lambda/create-function.png" caption="Prompt for creating a new function on Lambda."

Go fans around the world are undoubtedly celebrating the
move, but this is a big step forward for everyone, Go
programmer or not. Although not a perfect language by any
means, Go has a few properties that make it absolutely
ideal for use in a serverless environment like Lambda.

## Lambda runtimes as special snowflakes (#runtimes)

Lambda's exact implementation details have always been a
little mysterious, but we know a few things about it. We
know that user processes are started in sandboxed
containers. Containers that have finished their execution
may be kept around to service a future invocation of the
same function (but might not be), and that between
invocations containers are frozen that no user code is
allowed to execute.

Containers also flavored as one of the [preconfigured
runtimes][runtimes] allowed by Amazon. Here's the current
list (as of January 16th):

* **Node.js** – v4.3.2 and 6.10.3
* **Java** – Java 8
* **Python** – Python 3.6 and 2.7
* **.NET Core** – .NET Core 1.0.1 and .NET Core 2.0
* **Go** – Go 1.x

That's a pretty good variety of languages, but maybe more
interesting is what's *missing* from the list. While .NET
Core and Python are relatively up-to-date, Java 9 is
missing from the list, along with any recent major version
of Node (7.x, 8.x, or 9.x). Notably, major features like
`async/await` (which arrived Node ~7.6) are still not
available on Lambda even a year after release.

This tells us something else about Lambda: that new
runtimes are non-trivial to create, run, and maintain,
which means that versions often lag far behind their
initial availability. Given that Lambda [will be four years
old this year][history], it seems unlikely that Amazon will
be able to address this deficiency anytime soon.

!fig src="/assets/go-lambda/mountain.jpg" caption="Go 1.x's longevity is so impressive that it feels like a part of the landscape."

## The remarkable longevity of 1.x (#longevity)

That brings us back to Go. Note above that the Go runtime
specifies version "1.x". At first glance that might not
look all that different from other languages on the list,
but consider that Go 1 was released almost _six years ago_
in [March 2012][releases]!

Since then, Go has followed up with nine more releases on
the 1.x line (and with a tenth expected soon), each of
which carried significant improvements and features. And
while it's rare to ever have a _perfect_ release that
doesn't break anyone, Go's done as good of a job as is
practically possible, and generally new releases are as
pain and worry-free as bumping a number in your
`.travis.yml`.

This level and length of API stability for a programming
language is all but unheard of, and it's made yet more
impressive given that Go is far from stagnant; it's one of
the most actively developed projects out there. The only
way it's been made possible is that having experienced the
amount of pain involved in most language upgrades, Go's
team has made stability a core philosophical value. Here's
[an entire article][go1] dedicated to the policies around
stability for the 1.x line. An excerpt:

> It is intended that programs written to the Go 1
> specification will continue to compile and run correctly,
> unchanged, over the lifetime of that specification. At
> some indefinite point, a Go 2 specification may arise,
> but until that time, Go programs that work today should
> continue to work even as future "point" releases of Go 1
> arise (Go 1.1, Go 1.2, etc.).
>
> The APIs may grow, acquiring new packages and features,
> but not in a way that breaks existing Go 1 code.

This might sound like normal semver, but semver only
dictates what to do in the event of a breaking change, not
about committing to not making them. Go's track record in
this area puts it provably ahead of just about any other
project.

That brings us back to Lambda. If we look back at our list
of runtimes, the supported versions across languages might
not look all that different, but it's a reasonably safe bet
that the "Go 1.x" in that list is going to outlive every
other one of those options, probably by a wide margin.

## Static binaries, dependencies, and forward compatibility (#static)

[The Lambda guide for Go][guide] suggests creating a
function by building a statically-linked binary (the
standard for Go), zipping it up, and uploading the whole
package to AWS:

``` sh
$ GOOS=linux go build -o main
$ zip deployment.zip main
$ aws lambda create-function ...
```

This is in sharp contrast to other support languages where
you send either source-level code (Node, Python), or
compiled bytecode (.NET Core, Java). It also has some major
advantages.

Statically linking makes dependency deployment absolutely
trivial. Anything that's needed by a final program is
linked in at compile time, and once a program needs to
execute, it doesn't need to think about project layout,
include paths, project entry points, or anything else. At
the source level dependency management has been a long
criticized blindspot of Go, but with the addition of the
`vendor/` directory in Go 1.6 and the uptake on the new
[`dep`][dep] dependency management tool, the future is
looking brighter than ever.

Static binaries also carry the promise of forward
compatibility. Unlike even a bytecode interpreter, when a
new version of Go is released, the Lambda runtime may not
necessarily need to be updated given that existing
containers will almost certainly be able to run the new
binary. (This comes with the caveat that `net/rpc` package
that [`aws-lambda-go` uses for the entrypoint][entrypoint]
remains stable across versions. This is reasonably likely
though given that the package has been [frozen][frozen] for
more than a year.)

## Stability as a feature (#stability)

It's rare to write software and not have it come back to
haunt you in a few year's time as it needs to be fixed and
upgraded. Go is one of the only places in the software
world that you'll find anything akin to stability.

[announce]: https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/
[dep]: https://github.com/golang/dep
[entrypoint]: https://github.com/aws/aws-lambda-go/blob/master/lambda/entry.go
[frozen]: https://go-review.googlesource.com/c/go/+/32112
[go1]: https://golang.org/doc/go1compat
[guide]: https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/
[history]: https://docs.aws.amazon.com/lambda/latest/dg/history.html
[package]: https://docs.aws.amazon.com/lambda/latest/dg/lambda-go-how-to-create-deployment-package.html
[releases]: https://golang.org/doc/devel/release.html#go1
[runtimes]: https://docs.aws.amazon.com/lambda/latest/dg/current-supported-versions.html
[sdk]: https://github.com/aws/aws-lambda-go
