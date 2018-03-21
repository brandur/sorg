---
title: A Tour of a Fast and Complete Web Backend in Rust
published_at: 2018-03-21T16:30:04Z
hook: TODO
---

I'm having a crisis of faith in interpreted languages right
now. They're fun and fast to work in at small scale, but
when you have a project that gets big, they lose their
veneer pretty quickly. Working with a big project in Ruby
or JavaScript feels like a never ending game of
whack-a-mock -- you fix one problem only to have your
refactor cause a new one to appear somewhere else. No
matter how many tests you write or how well-disciplined
your team is, any new development is sure to introduce a
stream of bugs that will only be shored up over the course
of months or years.

I'm running a personal experiment right now that stems from
the question: can we build more reliable systems with
programming languages that provide better checks and
stronger constraints?

To that end I've skewed all the way to polar opposite end
of the spectrum and have been building a web service in
Rust, a language infamous for its uncompromising compiler.
I don't know whether it's a good idea -- the language is
still new and still deeply impractical, but it's been an
interesting learning experience.

It's been a slog learning how to work with the compiler's
strict rules around types, ownership, and lifetimes, but I
now have a web service that works, and I can say already
that its proved my thesis at least partially correct and
that runtime errors are way down. I want to show off some
of the more novel features of the language, core libraries,
and various frameworks that I used to build it.

## The foundation (#foundation)

Actix-web.

It's really fast

More feature-complete than frameworks that come out of large organizations.

e.g. Cookies/sessions, HTTP/2, WebSockets, streaming
responses, static file serving, good testing
infrastructure.

Diesel.

Query safety, etc.

## The concurrency model (#concurreny-model)

Tokio. One core per machine (?).

### Sync executors (#sync-executors)

Diesel isn't async.

Too much concurrency is not always a good thing --
especially where databases are concerned. This allows the
size of the connection pool to be easily managed.

## Error handling (#error-handling)

Error chain. Define user errors for public consumption.

Implementation with futures is quite elegant:

``` rust
let message = server::Message::new(&log, params);

// Send message to sync executor
sync_addr
    .send(message)
    .and_then(move |executor_response| {
        // Transform executor response to HTTP response
    }
    .then(|res|
        server::transform_user_error(res, render_user_error)
    )
    .responder()
```

## Middleware (#middleware)

Types can keep middleware state constrained to module (just
don't export it).

Middleware can be async too!

So for example, you might have a rate limiting middleware
that makes async calls to Redis, or a user authentication
that uses its own Sync Executor to talk to the database.

## Testing (#testing)

There's a few testing strategies, I settled on unit testing through the HTTP stack.

`json!` macro.

[actixweb]: https://github.com/actix/actix-web
[techempower]: https://www.techempower.com/benchmarks/#section=data-r15&hw=ph&test=plaintext
