---
title: Middleware Boundary Bleeding and Enforced Modularity in Rust
location: San Francisco
published_at: 2018-03-01T18:26:25Z
hook: TODO
---

Lately I've been playing with the idea with building a web
application in Rust. I don't know whether it's a good idea
-- the language is still new, still pretty impractical due
to a steep learning curve and complex type system, and
major components like a concurrency story are still under
active development and seeing rapid change.

I'm doing it because I have a question above software that
I want answered: can we build more reliable systems with
programming languages that provide better checks and
stronger constraints? I've been working on large projects
in interpreted languages for years and the neverending game
of whack-a-mole is tiring. You fix one problem only to have
a new one pop up somewhere else. Any new development comes,
no matter how careful, produces a stream of bugs that will
be shored up over the course of years.

Rust is interesting because its sophisticated type system
unlocks the potential for very strong compile-time checks
-- for example to check the correctness of an SQL query
with [Diesel][diesel] or an HTML view with
[Horrorshow][horrorshow]. A pedantic compiler isn't a new
idea, but I'm also attracted to Rust because of its modern
toolchains, strong conventions, pleasant syntax and
explicitness, and the attention to detail put into its
upfront design (compared to say a Haskell, which in this
author's opinion, has lapsed on most of these fronts).

One are that I've looked into lately is web middleware.
Middleware is a great idea -- perfectly modular components
that can be written once and shared between projects, but
like many good ideas, in practice they still often violate
encapsulation and lead to bugs. I found some of the ideas
for middleware in Rust's web frameworks to be pretty good
at counteracting this effect. Let's take a closer look.

## Recap: the middleware interface (#recap)

A middleware is a module that gets a chance to perform some
operation before an HTTP request and after it. Here's the
basic interface:

``` ruby
class MyMiddleware
  def initialize(app)
    @app = app
  end

  def call(env)
    #
    # request pre-processing here
    #

    status, headers, data = @app.call(env)

    #
    # request post-processing here
    #

    [status, headers, data]
  end
end
```

`app` is the application on which the middleware is
installed, but it may also be a reference to an application
that's already wrapped in another middleware. Middleware is
nested as a stack so that any number of them can be used,
each calling into the next component.

Rack receives a request and invokes `call` on the top
middleware of an app with the request's environment (which
contains method, path, and headers among other things).
Each middleware gets a chance to perform some arbitrary
operation before it invokes `@app.call(env)`, which
transfers control to the next middleware in the stack, and
eventually the app itself. After the app responds, the
middleware stack unwinds and each one gets a chance to
modify the response before passing control back to its
parent.

A great example of a perfectly modular piece of middleware
is [`Rack::Deflater`][deflater] which, among other things,
will Gzip a response when a client has signaled that they
support compression by sending an `Accept-Encoding` header.

Note that we've looked at a _Ruby_ middleware interface,
but the idea has spread to many languages and frameworks,
and is implemented similarly elsewhere.

## The promiscuity of unchecked keys (#unchecked-keys)

A common pattern in middleware (and possibly undesirable,
but still widespread), is to share information by injecting
new data into a request `env`, usually under an
application-specific key with a prefix like `app.*`. This
is useful because it allows downstream components to use
facilities provided by their parent middleware upstream.

For a simple example, consider these two middleware:

* `LogInitializerMiddleware` that initializes a common
  `Logger` to be used for the lifetime of a request.
* `RequestIDMiddleware` that generates a [request
  ID](/request-ids) to identify the request.

Both store a key into `env`, and the downstream
`RequestIDMiddleware` emits a debug log message with the
token it's generated.

``` ruby
class LogInitializerMiddleware
  def initialize(app)
    @app = app
  end

  def call(env)
    logger = Logger.new(STDOUT)
    env['app.logger'] = logger
    @app.call(env)
  end
end

class RequestIDMiddleware
  def initialize(app)
    @app = app
  end

  def call(env)
    request_id = SecureRandom.uuid
    env['app.request_id'] = request_id
    env['app.logger'].debug "Generated request ID: #{request_id}"
    @app.call(env)
  end
end
```

And our composed app:

``` ruby
app = Rack::Builder.new do
  use LogInitializerMiddleware
  use RequestIDMiddleware
  run HelloWorldApp
end

Rack::Server.start app: app
```

Running the program and making a request produces exactly
the result that we'd expect:

```
D, [2018-03-02T07:50:27.213372 #89779] DEBUG -- : Generated request ID: 0884859f-9fc2-4838-909b-6efd27e539c5
```

### Dependencies and cross dependencies (#cross-dependencies)

Simple enough right? Our middlewares have dependencies, but
the dependency graph is still directed and acyclic so that
it's easy to reason about. Unfortunately for us though,
dependencies have a nefarious tendency to become cross
dependencies when there are no checks in place to ensure
that they shouldn't.

Say that we decided that we wanted to ensure that each
request's ID gets prefixed to every log line that it
produces, a perfectly rational thing to do. One way we
might do this is to pass a formatter to our `Logger`:

``` ruby
class LogInitializerMiddleware
  def initialize(app)
    @app = app
  end

  def call(env)
    logger = Logger.new(STDOUT)

    # Make sure that request ID gets prefixed to every log line!
    original_formatter = Logger::Formatter.new
    logger.formatter = ->(severity, datetime, progname, msg) {
      msg = "Request #{env['app.request_id']}: #{msg}"
      original_formatter.call(severity, datetime, progname, msg)
    }

    env['app.logger'] = logger
    @app.call(env)
  end
end

class RequestIDMiddleware
  def initialize(app)
    @app = app
  end

  def call(env)
    request_id = SecureRandom.uuid
    env['app.request_id'] = request_id
    env['app.logger'].debug "Generated request ID: #{request_id}"
    @app.call(env)
  end
end
```

It might be a bit of a surprise that despite each
middleware accessing keys that were set by the other, this
code still works:

```
D, [2018-03-01T12:32:28.228791 #1747] DEBUG -- : Request 7bdd0e36-b667-490b-b107-b425f788ad15: Generated request ID: 7bdd0e36-b667-490b-b107-b425f788ad15
```

This is because our closure is capturing the entire `env`
hash, so `app.request_id` is checked when a new log line is
emitted instead of when the formatter was originally
initialized.

Even though it works, the cross-dependency still makes this
code fragile because it's not obvious, not explicit, and
the interpreter won't do anything to help reveal problems
with it. Say another developer comes through and decides to
optimize the original code a bit:

``` ruby
prefix = "Request #{env['app.request_id']}: "
logger.formatter = ->(severity, datetime, progname, msg) {
  msg = prefix + message
  original_formatter.call(severity, datetime, progname, msg)
}
```

It's now broken. `app.request_id` isn't yet available when
the formatter is defined.

If this seems overly simplistic, it's meant to be. Try to
mentally scale up these two small and basic middleware into
a stack of 50 big and complicated ones, each with its own
implicit dependencies, and you'll have a better idea of
what a real production stack looks like.

This is roughly the number of middleware we're running, and
simply reordering them has caused production incidents. The
only way to see dependencies is to grep for specific
strings being set in `env`.

## Middleware modules and types in Rust (#rust)

Dependencies are explicit based on the type system, and
cross dependencies are impossible. Each middleware lives in
a separate module, and the compiler won't allow modules to
depend on each other.

TODO

``` rust
mod middleware {
    use actix_web;
    use actix_web::{HttpRequest, HttpResponse};
    use actix_web::middleware::Middleware as ActixMiddleware;
    use actix_web::middleware::{Response, Started};
    use common;
    use slog::Logger;

    pub mod log_initializer {
        use middleware::*;

        pub struct Middleware;

        /// The extension registered by this middleware to the request to make
        /// a `Logger `accessible.
        pub struct Extension(pub Logger);

        impl<S: common::State> ActixMiddleware<S> for Middleware {
            fn start(
                &self,
                req: &mut HttpRequest<S>,
            ) -> actix_web::Result<Started> {
                let log = req.state().log().clone();
                req.extensions().insert(Extension(log));
                Ok(Started::Done)
            }

            fn response(
                &self,
                _req: &mut HttpRequest<S>,
                resp: HttpResponse,
            ) -> actix_web::Result<Response> {
                Ok(Response::Done(resp))
            }
        }
    }

    pub mod request_id {
        use middleware::*;
        use uuid::Uuid;

        pub struct Middleware;

        /// The extension registered by this middleware to the request to make
        /// a request ID accessible.
        pub struct Extension(pub String);

        impl<S: common::State> ActixMiddleware<S> for Middleware {
            fn start(
                &self,
                req: &mut HttpRequest<S>,
            ) -> actix_web::Result<Started> {
                let request_id = Uuid::new_v4().simple().to_string();
                req.extensions().insert(Extension(request_id.clone()));

                let log = &req.extensions()
                    .get::<log_initializer::Extension>()
                    .unwrap()
                    .0;

                debug!(&log, "Generated request ID";
                    "request_id" => request_id.as_str());

                Ok(Started::Done)
            }

            fn response(
                &self,
                _req: &mut HttpRequest<S>,
                resp: HttpResponse,
            ) -> actix_web::Result<Response> {
                Ok(Response::Done(resp))
            }
        }
    }
}
```

We also define a `State` trait and `StateImpl`
implementation that will contain the program's root logger
along with `main` which composes the application stack:

``` rust
mod common {
    use slog::Logger;

    pub trait State {
        fn log(&self) -> &Logger;
    }

    pub struct StateImpl {
        pub log: Logger,
    }

    impl State for StateImpl {
        fn log(&self) -> &Logger {
            &self.log
        }
    }
}

fn main() {
    let sys = actix::System::new("middleware-rust");

    let _addr = actix_web::HttpServer::new(|| {
        actix_web::Application::with_state(common::StateImpl { log: log() })
            .middleware(middleware::log_initializer::Middleware)
            .middleware(middleware::request_id::Middleware)
            .resource("/", |r| {
                r.method(actix_web::Method::GET)
                    .f(|_req| actix_web::httpcodes::HTTPOk)
            })
    }).bind("127.0.0.1:8080")
        .expect("Can not bind to 127.0.0.1:8080")
        .start();

    println!("Starting http server: 127.0.0.1:8080");
    let _ = sys.run();
}
```

But what if we want to add the functionality that we did
for Ruby above where the request ID gets included with
every logged line of the request? To make that work we'll
have to shift responsibilities: instead of making it the
log middleware's job to find a request ID, we'll make it
the request ID middleware's job to inject one. Here's a
modified implementation of the request ID middleware:

``` rust
impl<S: common::State> ActixMiddleware<S> for Middleware {
    fn start(
        &self,
        req: &mut HttpRequest<S>,
    ) -> actix_web::Result<Started> {
        let request_id = Uuid::new_v4().simple().to_string();
        req.extensions().insert(Extension(request_id.clone()));

        // Remove the request's original `Logger`
        let log = req.extensions()
            .remove::<log_initializer::Extension>()
            .unwrap()
            .0;

        debug!(&log, "Generated request ID";
            "request_id" => request_id.as_str());

        // Insert a new `Logger` that includes the generated request ID
        req.extensions().insert(log_initializer::Extension(log.new(
            o!("request_id" => request_id),
        )));

        Ok(Started::Done)
    }

    fn response(
        &self,
        _req: &mut HttpRequest<S>,
        resp: HttpResponse,
    ) -> actix_web::Result<Response> {
        Ok(Response::Done(resp))
    }
}
```

## Safety-by-convention is not enough (#safety)

[deflater]: https://github.com/rack/rack/blob/master/lib/rack/deflater.rb
[diesel]: https://github.com/diesel-rs/diesel
[horrorshow]: https://github.com/Stebalien/horrorshow-rs
