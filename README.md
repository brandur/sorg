# sorg [![Build Status](https://travis-ci.org/brandur/sorg.svg?branch=master)](https://travis-ci.org/brandur/sorg)

A Go-based build script that compiles my [personal website][brandur]. This is
the site's second incarnation, with the original being [a Ruby/Sintra
stack][org] (sorg = "static org").

The site deploys automatically from its CI build in Travis as changes are
committed to the master branch.

## Clone

``` sh
git clone https://github.com/brandur/sorg.git
```

## Build

Install Go 1.9+, set up and run [blackswan][blackswan], then:

``` sh
go get -u github.com/ddollar/forego

cp .env.sample .env

# Used to run the test suite.
createdb sorg-test

# Compile Go executables.
make install

# Run an initial build of the site, look for build output in public/.
forego run make build

# Watch for changes in Go files and/or content and recompile and rebuild when
# one occurs.
forego start
```

The project can be deployed to s3 using:

``` sh
pip install awscli

export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export S3_BUCKET=...
make deploy
```

Cached photos can be fetched using:

``` sh
make photographs-download
```

## Development

Run the entire lifecycle like in CI:

``` sh
make
```

Run the test suite:

``` sh
make test
```

Run a single package's test suite or single test:

``` sh
go test ./markdown
go test ./markdown -run TestCollapseHTML
```

Get more verbose output while running tests:

```
go test -v ./markdown
```

## Vendoring dependencies

Dependencies are managed with dep. New ones can be vendored
using these commands:

    dep ensure -add github.com/foo/bar

[blackswan]: https://github.com/brandur/blackswan
[brandur]: https://brandur.org
[org]: https://github.com/brandur/org

<!--
# vim: set tw=79:
-->
