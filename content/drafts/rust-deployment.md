---
title: Productionizing a Rust Web Stack and Deploying it to
  Kubernetes
published_at: 2018-03-25T17:10:24Z
hook: TODO
location: San Francisco
---

## Structuring a program (#structure)

All one program:

```
$ podcore api
$ podcore crawl
$ podcore web
```

### Binary-embedded migrations with Diesel (#migrations)

`diesel` ships with a CLI separate from the main crate
which provides schema commands like migrations. This is a
somewhat nice idea because it minimizes the size of the
core crate and projects using it don't have to include that
scaffolding. However, it does make production deployments
inconvenient, especially on systems like Kubernetes where
you'd need a separate image just to include Diesel's
tooling.

``` rust
#[macro_use]
extern crate diesel_migrations;

// Migrations get pulled into the final binary. This makes it quite a bit
// easier to run them on remote clusters without trouble.
embed_migrations!("./migrations");

...

embedded_migrations::run_with_output(&*conn, &mut std::io::stdout())
```

`$ podcore migrate`

## Observability with structured logging (#observability)

[slog][slog]

```
Mar 25 10:27:05.938 DEBG Generated request ID, request_id: 50cc9a741e43447d82b8ecf40ad949de
request_id: 50cc9a741e43447d82b8ecf40ad949de
 step: build_params
  Mar 25 10:27:05.938 INFO Start
  Mar 25 10:27:05.938 INFO Finish, elapsed: 9.808µs
 step: handle_message
  Mar 25 10:27:05.938 INFO Start
  Mar 25 10:27:05.938 INFO Looking up podcast, id: 1
  Mar 25 10:27:05.941 INFO Finish, elapsed: 2.597ms
 step: render_view_model
  Mar 25 10:27:05.941 INFO Start
  Mar 25 10:27:05.941 INFO Finish, elapsed: 352.213µs
 Mar 25 10:27:05.941 INFO Request finished, status: 200, path: /podcasts/1, method: GET, elapsed: 3.795ms
```

In production, the system detects that `stdout` isn't a
TTY, and it switches to a form of asynchronous logging
instead. This format is searchable and is indifferent to
lines from different requests being interleaved:

```
Mar 25 10:29:14.662 DEBG Generated request ID, request_id: 59a12820537c4dcab7dbc9fd47e98cb5
Mar 25 10:29:14.662 INFO Start, step: build_params, request_id: 59a12820537c4dcab7dbc9fd47e98cb5
Mar 25 10:29:14.662 INFO Finish, elapsed: 50.645µs, step: build_params, request_id: 59a12820537c4dcab7dbc9fd47e98cb5
Mar 25 10:29:14.663 INFO Start, step: handle_message, request_id: 59a12820537c4dcab7dbc9fd47e98cb5
Mar 25 10:29:14.663 INFO Looking up podcast, id: 1, step: handle_message, request_id: 59a12820537c4dcab7dbc9fd47e98cb5
Mar 25 10:29:14.666 INFO Finish, elapsed: 2.972ms, step: handle_message, request_id: 59a12820537c4dcab7dbc9fd47e98cb5
Mar 25 10:29:14.666 INFO Start, step: render_view_model, request_id: 59a12820537c4dcab7dbc9fd47e98cb5
Mar 25 10:29:14.666 INFO Finish, elapsed: 450.816µs, step: render_view_model, request_id: 59a12820537c4dcab7dbc9fd47e98cb5
Mar 25 10:29:14.666 INFO Request finished, status: 200, path: /podcasts/1, method: GET, elapsed: 4.633ms, request_id: 59a12820537c4dcab7dbc9fd47e98cb5
```

## The Docker build (#docker-build)

[musl][musl] 

> Today’s Rust depends on libc, and on most Linuxes that
> means glibc. It’s technically challenging to fully
> statically link glibc, which presents difficulties when
> using it to produce a truly standalone binary.
> Fortunately, an alternative exists: musl, a small, modern
> implementation of libc that can be easily statically
> linked. Rust has been compatible with musl since version
> 1.1, but until recently developers have needed to build
> their own compiler to benefit from it.

Unfortunately, building a `musl` target in Rust isn't as
easy as specifying a cross-compile target and letting
`rustc` take care of it. There are still a few common
dependencies that will cause serious problems like OpenSSL,
needed by a number of different crates, and `libpq`, needed
by `diesel` for use with Postgres. Luckily, the amazing
[rust-musl-builder] project exists which bundles in static
versions of OpenSSL and `libpq` and provides a base image
which is incredibly easy to build a static musl Rust
program against.

This is one place where Go, having made the decision to
drop problematic dependencies like OpenSSL and `libpq` in
favor of pure Go versions, is dominating Rust. Building a
musl target from anywhere, including Mac OS, is literally a
one-liner. Rolling your own crypto seems like an eminently
bad idea at first glance, but so far Go's track record has
been good, and Go's purely native approach to dependencies
is very, very convenient. There's some hope for Rust here
in the form of pure-Rust implementations like
[Rustls][rustls], but it's probably some ways out.

```
#
# STAGE 1
#
# Uses rust-musl-builder to build a release version of the binary targeted for
# MUSL.
#

FROM ekidd/rust-musl-builder AS builder

# Add source code. Note that `.dockerignore` will take care of excluding the
# vast majority of files in the current directory and just bring in the couple
# core files that we need.
ADD ./ ./

# Fix permissions on source code.
RUN sudo chown -R rust:rust /home/rust

# Build the project.
RUN cargo build --release

#
# STAGE 2
#
# Use a tiny base image (alpine) and copy in the release target. This produces
# a very small output image for deployment.
#

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder \
    /home/rust/src/target/x86_64-unknown-linux-musl/release/podcore \
    /
COPY --from=builder \
    /home/rust/src/assets/ \
    /assets/

ENV PORT 8080
ENTRYPOINT ["/podcore"]
```

My image is coming out 24.5 MB these days. 

### Dockerignore (#dockerignore)

It's well the few minutes of time it'll take you to
optimize your `.dockerignore` file so that you don't have
to send as many files to your local daemon to start a
build. Mine is setup like a whitelist: `*` is used to
exclude all files and then particular files are reincluded
by prefixing them with `!`. Rust makes this particular easy
because by convention, all a crate's source is stored in
`src/`:

```
# exclude everything
*

# whitelist certain files
!/Cargo.*
!/assets/
!/migrations/
!/src/
```

## The Kubernetes configuration (#kubernetes)

### Deployment (#deployment)

> A pod (as in a pod of whales or pea pod) is a group of
> one or more containers (such as Docker containers), with
> shared storage/network, and a specification for how to
> run the containers.

> A Deployment controller provides declarative updates for
> Pods and ReplicaSets.

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: podcore-api
  labels:
    app: podcore-api
spec:
  replicas: 2
  template:
    metadata:
      labels:
        name: podcore-api
    spec:
      containers:
      - name: podcore
        image: gcr.io/podcore-194423/podcore:1.30
        command: ["/podcore", "api"]
        ports:
        - containerPort: 8082
```

### Service (#service)

> A Kubernetes Service is an abstraction which defines a
> logical set of Pods and a policy by which to access them
> - sometimes called a micro-service. The set of Pods
> targeted by a Service is (usually) determined by a Label
> Selector (see below for why you might want a Service
> without a selector).

```
apiVersion: v1
kind: Service
metadata:
  name: podcore-api
  labels:
    name: podcore-api
spec:
  ports:
  - port: 8082
    protocol: TCP
  selector:
    name: podcore-api
  type: NodePort
```

### Ingress (#ingress)

> Ingress is a Kubernetes resource that encapsulates a
> collection of rules and configuration for routing
> external HTTP(S) traffic to internal services.
>
> On Kubernetes Engine, Ingress is implemented using Cloud
> Load Balancing. When you create an Ingress in your
> cluster, Kubernetes Engine creates an HTTP(S) load
> balancer and configures it to route traffic to your
> application.

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: unified-ingress
  annotations:
    kubernetes.io/ingress.global-static-ip-name: podcore-static-ip
spec:
  rules:
  - host: api.podcore.mutelight.org
    http:
      paths:
      - backend:
          serviceName: podcore-api
          servicePort: 8082
  - host: podcore.mutelight.org
    http:
      paths:
      - backend:
          serviceName: podcore-web
          servicePort: 8083
```

### Graceful restarts (#graceful-restarts)

[musl]: https://en.wikipedia.org/wiki/Musl
[rust-musl-builder]: https://github.com/emk/rust-musl-builder
[rustls]: https://github.com/ctz/rustls
[slog]: 
