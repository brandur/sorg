+++
hook = "Defaulting all linters to on, and some callouts for some very useful lints which are not enabled by default."
published_at = 2021-09-26T22:20:08Z
title = "Tightening the bolts with golangci-lint"
+++

I've been using [golangci-lint](https://golangci-lint.run/) on work projects for a few months now. I managed to go an improbable amount of time not knowing about this project, but took to it immediately. Having been using Golint previously, which enforces _some_ code standards, but many of are questionable value (e.g. every exported type/function should have a docblock), and which catches very little in the way of bugs, Golangci-lint by comparison is hugely configurable with a plethora of useful linters, and catches many bugs.

Before now, I was using only its default set of linters, which should be good for most projects, and relatively easy to get turned out. Last weekend, I cranked things up by enabling every possible linter, and then went through each set of failures and banned only the ones that were unworkable. It took a few hours, but I got there.

Here's a couple non-default linters that are particularly good (see the [whole list here](https://golangci-lint.run/usage/linters/)):

* `exhaustive`: For enum-like constants used in a `switch` statement, checks to make sure that there are cases to handle all values exhaustively. Enums are a weak point in Go. This makes things a bit better.

* `forbidigo`: Great for making sure that you're only using one package for error wrapping everywhere ([we use `xerrors`](/fragments/go-xerror)).

* `gci`: Allows you to group import statements into stdlib versus internal versus 3rd party using the `local-prefixes` option. Very small improvement, but nice.

* `wrapcheck`: Makes sure that all errors incoming from 3rd party code are error-wrapped. This is important to always get a stacktrace when using something like [`xerrors`](/fragments/go-xerror). _Very_ useful, although hard to get turned on for existing code.

## Obnoxious (#obnoxious)

Here's my final configuration:

``` yaml
linters:
  disable:
    # obnoxious
    - cyclop
    - dupl
    - exhaustivestruct
    - forcetypeassert
    - funlen
    - gochecknoinits
    - gochecknoglobals
    - gocognit
    - gocyclo
    - godox
    - gomnd
    - nlreturn
    - paralleltest
    - testpackage
    - wsl

    # deprecated
    - golint
    - interfacer
    - maligned
    - scopelint
  enable-all: true

linters-settings:
  forbidigo:
    forbid:
      - '^errors\.Wrap$'
      - '^errors\.Wrapf$'
      - '^fmt\.Errorf$'
  gci:
    local-prefixes: github.com/brandur

  gocritic:
    disabled-checks:
      - commentFormatting

  wrapcheck:
    ignorePackageGlobs:
      - github.com/brandur/*
```

I bucketed disabled linters into either "deprecated" (`enable-all` activates even deprecated linters), or "obnoxious", which are linters that are hard to get activated and which run checks that are not obviously useful.

## Borderline (#borderline)

There are a couple I'm not sure about and left turned off, although maybe there's a case for them:

* `paralleltest`: Requires that your tests specify `t.Parallel()`. As a reminder, `go test` will run tests for different packages in parallel, but tests within a package sequentially. That is, unless you mark tests explicitly with `t.Parallel()`, which allow them to run in parallel with other test cases marked with `t.Parallel()`.

    Although parallelism is good in general, I couldn't find any well-argued cases online that all test cases should get a `t.Parallel()`, with the best articulation I saw being that it should generally just be reserved for specific cases that are known to be slow so that they can be made to run alongside other slow tests. The default inter-package `go test` concurrency should be enough to keep processors busy in most projects. The [project itself](https://github.com/kunwardeep/paralleltest) makes no attempt to argue its own case.
	
* `testpackage`: Requires that package tests be in a separate package like `my_package_test` compared to `my_package`. This is another one that seems conceptually sound -- it forces tests to only use exported APIs. That said, the number of times I've found it useful to package tests up with code so I can more easily exercise an internal function that shouldn't be exported is _innumerable_. And although it's not generally  considered good practice, it hasn't caused many problems for me over the years.
