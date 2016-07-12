# sorg

[![Build Status](https://travis-ci.org/brandur/sorg.svg?branch=master)](https://travis-ci.org/brandur/sorg)

A Go-based build script that compiles my [personal website][brandur].

The site deploys automatically from its CI build in Travis as changes are
committed to the master branch.

## Build

Install Go 1.6, set up and run [black-swan][black-swan], then:

``` sh
go get -u github.com/ddollar/forego

cp .env.sample

# Compile Go executables.
make install

# Run an initial build of the site, look for build output in public/.
forego run make build

# Watch for changes in Go files and/or content and recompile and rebuild when
# one occurs.
forego start
```

Or an easy all-in-one:

``` sh
make install && forego run make build && forego start
```

The project can be deployed to s3 using:

``` sh
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
S3_BUCKET=...
make deploy
```

[black-swan]: https://github.com/brandur/black-swan
[brandur]: https://brandur.org
[org]: https://github.com/brandur/org
