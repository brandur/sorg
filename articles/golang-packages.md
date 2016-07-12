---
hook: Understanding the benefits of Golang's restrictive (but simple) import and package
  management system.
location: San Francisco
published_at: 2014-10-20T00:54:44Z
title: Package Management in Go
---

Go's strategy for package management is a little untraditional by the standards of what most language ecosystems are setting today. Nearly every other language that's under active development today has opted for an approach to dependency management that involves central repositories accompanied by a recipe that's checked in with projects with instructions on how to rebuild the dependency tree that they need to run: Rubygems in Ruby, NPM in Node, Maven for the JVM, Cargo for Rust, etc. Go's more exotic approach can be a little harder for new entrants to understand, especially if they're coming in from other languages.

Personally, some basic principles left me reeling. Here's an [excerpt from the Go FAQ](http://golang.org/doc/faq#get_version) suggesting that maintainers of public packages fork their projects if they're introducing a backwards incompatible change:

> Packages intended for public use should try to maintain backwards compatibility as they evolve. The Go 1 compatibility guidelines are a good reference here: don't remove exported names, encourage tagged composite literals, and so on. If different functionality is required, add a new name instead of changing an old one. If a complete break is required, create a new package with a new import path.

You have to fork project to change the API?! It's hard to believe how this could be considered correct in any world, and many Go articles are so apologetic that downsides of this approach are never addressed. For those who were as confused about this as I was, let's address a few of the special characteristics of Go's import and packaging system.

## The Workspace (#workspace)

At the heart of Go's conventions is [the workspace](https://golang.org/doc/code.html#Workspaces). A workspace is simply any directory that your `$GOPATH` is currently referencing, and which has a basic layout like the following:

```
bin/
pkg/
src/
```

Whatever project you're working on should be in _a_ workspace at a location like `$GOPATH/src/github.com/brandur/heroku-agent`. When you introduce a dependency in an `import` statement, you're actually telling the compiler to resolve it within your workspace, even if it looks like that dependency comes from a remote host:

```
import (
    "github.com/brandur/my-dependency"
)
```

`go get` can retrieve your project's dependencies from a variety of providers and store them to your workspace for compilation.

## DVCS & the Central Repository (#dcvs)

A common line in the community is that Go's package system is "built on top of distributed version control", which reads like it's adding some extra layer of robustness on top of a more traditional packaging system. Although nominally true, the far and away popular convention is to reference DVCS hosts like GitHub or Google Code, which are no more distributed or robust than RubyGems. The operation of `go get` isn't too radically different than a `bundle install --local`.

This doesn't eliminate any central point of failure, but it does have the notable advantage of making the community less dependent on a single central repository that's quite expensive to maintain. Central repositories like RubyGems and NPM owe their continued existence and development to largely charitable sponsorship. Although this has traditionally worked quite well, this may not last forever, especially if either community loses some of the impressive public support that they currently enjoy. Go's approach allows support for the DCVS providers du jour to be added or removed as necessary; at the end of the day, the only common functionality required of a provider or source control system is to be able to check out source code to a known path.

## No Relative Imports (#no-relative-imports)

One of the most astonishing aspects of Go's import for me was that any kind of [relative import is strongly discouraged](https://groups.google.com/forum/#!topic/golang-nuts/_usbgS9LeS8) (relative import is only allowed outside a workspace). This seems reasonable when referencing external dependencies, but is pretty inconvenient when building more complex Go projects that are divided into subpackages where convention is still to fully qualify everything:

``` go
import (
    "github.com/goraft/raft/protobuf"
)
```

This syntax leaves a lot of open questions: am I actually referencing the master branch on GitHub? Do I have to push changes in my subpackages up before I can use them in main? Well, yes to the former, but not the latter. Once again, this approach is made tenable by workspaces: `import` statements always reference code within your workspace, but can contain a provider so that they can be fetched by `go get`. The code within the current workspace is what gets used for compilation, even if it's deviated from what's in the origin's master branch.

Removing the concept of a relative important has advantages as well: understanding of the local file hierarchy is no longer required to build a package, making paths in Go packages easier to reason about. This guarantees nice convention between projects; at no point in time is it necessary to detangle a project's exotic approach to organization.

It's also astoundingly good for open source in that it's trivial to find, inspect, and manipulate your project's dependencies when it becomes necessary to do so. Every package's location is explicitly implied by its import path and easy to find.

## Vendoring (#vendoring)

So importing from master is great, but any kind of non-trivial program will eventually need a way to make deployments repeatable, which means locking down dependencies. Again, the [Go FAQ](http://golang.org/doc/faq#get_version) somewhat surprisingly recommends that dependencies be locked by vendoring them into a project:

> If you're using an externally supplied package and worry that it might change in unexpected ways, the simplest solution is to copy it to your local repository. (This is the approach Google takes internally.) Store the copy under a new import path that identifies it as a local copy. For example, you might copy "original.com/pkg" to "you.com/external/original.com/pkg". Keith Rarick's goven is one tool to help automate this process.

This technique raises questions around repository cleanliness in that a lot of extraneous source code gets ported around with the import pieces. Go diffs can be nightmarish to read.

However, it does have the advantage of making builds not dependent on the availability of external services. It also avoids any dependency hell type problems where two dependencies rely on different versions of a third.

## It's About Simplicity (#simplicity)

Like everything in Go, the import system is based on the same fundamental principle of simplicity that the [rest of the language encourages](http://bradgignac.com/2014/09/24/avoiding-complexity-with-go.html). Packages are resolved using he same popular version control systems that you use to store your source code. Packages are housed in the same location as your project (the `$GOPATH`). There is no versioning of any kind; the compiler ingests whatever code is on disk. Everything can be resolved and built by the Go compiler without any other special tooling.

The merits of Go's approach compared to other languages is certainly disputable, but the refreshing minimality of Go's system can't be easily dismissed. Having recently spent an hour loosening version constraints in upstream the dependencies of a large Ruby app, and another two trying to have Maven resolve a common HTTP library, working with a version control system that I can easily reason about is an attractive prospect indeed.
