+++
hook = "As of Go 1.21, Go fetches toolchains automatically, and it's easy to not be running the version that you thought you were running."
published_at = 2024-08-11T11:52:37-07:00
title = "Your Go version CI matrix might be wrong"
+++

We had an unpleasant surprise this week in [River's](https://github.com/riverqueue/river) CI suite. Since the project's inception we _thought_ we were supporting the latest two versions of Go (1.21 and 1.22), but it turns out that we never were.

As per common convention, we had a GitHub Actions CI matrix testing against both versions:

``` yaml
strategy:
  matrix:
    go-version:
      - "1.21"
      - "1.22"
```

That looks kosher, right? Wrong!

Builds were happily passing this whole time, but upon closer inspection of the install step, we see this:

``` txt
Run actions/setup-go@v5
Setup go version spec 1.21
Found in cache @ /opt/hostedtoolcache/go/1.21.12/x64
Added go to the path
Successfully set up Go version 1.21
go: downloading go1.22.5 (linux/amd64)
```

GitHub Actions had been downloading Go 1.21, then immediately upgrading itself to Go 1.22.

Since Go 1.21, Go has had a built in concepts called [toolchains](https://go.dev/doc/toolchain). An installed version of Go contains its own toolchain, but has the capacity to fetch and install other toolchains as well. Usually this is convenient feature because it means you can drop into any Go project and immediately get it running with a single command with no package or version managers in sight, but it has unexpected side effects.

## `go.mod` version and toolchain (#go-mod)

Along with toolchains, Go 1.21 also changed its treatment of `go` directives in `go.mod` so that instead of an advisory requirement, they're now a mandatory one. Any Go project needs to have its own `go` directive set to something at least as high as any modules it requires. So if a dependency requires Go 1.22.5, the project itself must be set to at least Go 1.22.5. Most of the time you won't even notice this because getting a new module with `go get` will handle updating a project's `go` directive automatically.

Given River is always a dependency, we want to provide as much leeway as possible on the minimum version bound, even while we'll always be using more modern versions of Go. `go.mod` files support a `go` directive along with a `toolchain` to specify a minimum bound along with a preferred toolchain:

``` txt
go 1.21

toolchain go1.22.5
```

Once again though, the presence of `toolchain` will cause CI jobs to upgrade themselves to 1.22 instead of running on the version of Go they're supposed to be targeting. We need one more magic env var to prevent this:

``` yaml
env:
  # The special value "local" tells Go to use the bundled Go
  # version rather than trying to fetch one according to a
  # `toolchain` value in `go.mod`. This ensures that we're
  # really running the Go version in the CI matrix rather than
  # one that the Go command has upgraded to automatically.
  GOTOOLCHAIN: local
```

## Take care with `go` directives (#go-directives)

A learning from this debacle is that Go modules that expect to be dependencies need to be very careful with the `go` directive in `go.mod` because it could have considerable downstream impact.

We're setting `go 1.21` which is the same as `go 1.21.0`, so any project that requires River will be able to use any patch version of Go 1.21 or 1.22.

Go's incredibly trigger happy when it comes to changing a `go.mod'`s `go` version, which it will happily and silently do at any opportunity. I'm legitimately amazed that we haven't seen more problems where dependencies accidentally upgrade to a new version of Go and break any downstream projects where that new version isn't yet available. This could even happen where a patch version changes as a brand new Go release comes out, but isn't yet available in everyone's build systems.

River's a multi-module project, and we hadn't even intentionally updated to Go 1.22.5, which spurred the bug report that led to discovery of the issue. I think what happened is that as we added new modules with `go mod init`, those would get assigned the latest patch release of Go, and then as we we required those from other modules, the new versions would proliferate. We'd see the change in diffs being reviewed, but didn't think much of it.

Along with patching all our directives to `go 1.21` we'll also be adding a CI check that verifies they all match up across modules to avoid any accidental version bumps in the future.