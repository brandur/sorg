---
title: A Tour of a Fast and Complete Web Stack in Rust
published_at: 2018-03-21T16:30:04Z
hook: TODO
---

## The foundation (#foundation)

Actix-web.

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
