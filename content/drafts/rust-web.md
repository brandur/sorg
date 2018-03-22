---
title: A Tour of a Fast and Complete Web Service in Rust
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

I built my service on [`actix-web`][actixweb], a web
framework built on [`actix`][actix], an actor library for
Rust similar to what you might see in a language like
Erlang, except which makes heavy use of Rust's
sophisticated type and concurrency systems.

You may recognize the name because `actix-web` has made its
way to the top of the [TechEmpower benchmarks][techempower]
for web frameworks. Apps built for these sorts of
benchmarks usually turn out to be a little contrived, but
its now contrived Rust code that's sitting right up at the
list with contrived C++ code, and _above_ contrived Java
code. Regardless of how you feel about that, the takeaway
is that `actix-web` is _fast_.

!fig src="/assets/rust-web/techempower.png" caption="Rust is consistently ranking alongside C++ and Java on TechEmpower."

It appears to be one of these projects written by a prodigy
-- it's only about six months old, and not only is already
more feature-complete and with better APIs than web
frameworks seen in other open source languages, but more so
than the frameworks bankrolled by large organizations.
Niceties like HTTP/2, WebSockets, steaming responses,
graceful shutdown, HTTPS, cookie support, static file
serving, and good testing infrastructure are all available
out of the box. And although the documentation is still a
bit rough, I've yet to run into a single bug.

### Diesel and compile-time query checking (#diesel)

I've been using [`diesel`][diesel] as an ORM to talk to
Postgres. Using `diesel` has its ups and downs, but most
importantly it's an ORM written by someone with a lot of
past experience with building ORMs (in this case
`ActiveModel`), and it's quite feature complete. A lot of
great design decisions were baked right into it from the
start, like foregoing a specialized DSL for migrations in
favor of raw SQL, or supporting powerful Postgres-specific
features like upsert or `jsonb` in the core library.

Most of my database queries are written using `diesel`'s
type-safe DSL. If I misreference a field, try to insert a
tuple into the wrong table, or even produce an impossible
join, the compiler tells me about it. A typical operation
looks a little like this:

``` rust
time_helpers::log_timed(&log.new(o!("step" => "upsert_episodes")), |_log| {
    Ok(diesel::insert_into(schema::episode::table)
        .values(ins_episodes)
        .on_conflict((schema::episode::podcast_id, schema::episode::guid))
        .do_update()
        .set((
            schema::episode::description.eq(excluded(schema::episode::description)),
            schema::episode::explicit.eq(excluded(schema::episode::explicit)),
            schema::episode::link_url.eq(excluded(schema::episode::link_url)),
            schema::episode::media_type.eq(excluded(schema::episode::media_type)),
            schema::episode::media_url.eq(excluded(schema::episode::media_url)),
            schema::episode::podcast_id.eq(excluded(schema::episode::podcast_id)),
            schema::episode::published_at.eq(excluded(schema::episode::published_at)),
            schema::episode::title.eq(excluded(schema::episode::title)),
        ))
        .get_results(self.conn)
        .chain_err(|| "Error upserting podcast episodes")?)
})
```

Some more complex SQL is difficult to represent using the
DSL, but luckily we have a great alternative. Rust's built
in `include_str!` macro allows us to ingest a file's
contents during compilation, and from there we can use
`diesel` for parameter binding and execution:

``` rust
diesel::sql_query(include_str!("../sql/cleaner_directory_search.sql"))
    .bind::<Text, _>(DIRECTORY_SEARCH_DELETE_HORIZON)
    .bind::<BigInt, _>(DELETE_LIMIT)
    .get_result::<DeleteResults>(conn)
    .chain_err(|| "Error deleting directory search content batch")
```

Meanwhile, the more elaborate query lives in a separate
`.sql` file:

``` sql
WITH expired AS (
    SELECT id
    FROM directory_search
    WHERE retrieved_at < NOW() - $1::interval
    LIMIT $2
),
deleted_batch AS (
    DELETE FROM directory_search
    WHERE id IN (
        SELECT id
        FROM expired
    )
    RETURNING id
)
SELECT COUNT(*)
FROM deleted_batch;
```

We lose compile-time SQL checking with this approach, but
we gain direct access to the raw power of SQL's semantics,
and great syntax highlighting in your favorite editor (as
it detects the `.sql` extension and highlights
appropriately).

## An adequate model for concurrency (#concurreny-model)

`actix-web` is powered by [`tokio`][tokio], Rust's fast
event loop library [1]. When starting an HTTP server, it
spawns a number of workers (by default the number is equal
to the number of logical cores on the server), each in
their own thread, and each with their own `tokio` reactor.

HTTP handlers can be written in a variety of ways. For
example, we might write one that returns some content
synchronously:

``` rust
fn index(req: HttpRequest) -> Bytes {
    ...
}
```

This will block the underlying `tokio` core until it's
finished, but that's appropriate in situations where no
other blocking calls need to be made; for example,
rendering a static view, or returning a health check
response.

We can also write an HTTP handler that returns a boxed
future for operations that may need to make another
asynchronous call:

``` rust
fn index(req: HttpRequest) -> Box<Future<Item=HttpResponse, Error=Error>> {
    ...
}
```

Examples of this might be responding with a file that we're
reading from disk (which is blocking, albeit minimally), or
waiting on a response from a database operation. While
waiting on a future's result, the underlying `tokio`
reactor will happily perform other work.

!fig src="/assets/rust-web/concurrency-model.svg" caption="An example of the concurrency model with actix-web."

### Synchronous actors (#sync-actors)

Support for futures in Rust is widespread, but not
universal. Notably, `diesel` doesn't support asynchronous
operations, so any operations you use it to make will block
as they wait for a database operations. Running one from
within a synchronous HTTP handler would lock up that
thread's `tokio` reactor.

Luckily, `actix` has a great solution for this problem in
the form of _synchronous actors_. These are actors that
expect to run their workloads synchronously, and so each is
started in its own thread. `actix` provides the
`SyncArbiter` abstraction to easily start a number of them.
The cluster shares a single message queue (referenced by
`addr` below) so that it's trivial to dispatch work into
it:

``` rust
// Start 3 `DbExecutor` actors, each with its own database
// connection, and each in its own thread
let addr = SyncArbiter::start(3, || {
    DbExecutor(SqliteConnection::establish("test.db").unwrap())
});
```

Although the work within a synchronous actor is blocking,
other actors in the system don't need to wait for it --
they get a future back that represents the message result
so that they can keep doing other work.

In my web service, fast workloads like parsing parameters
and rendering views is performed right inside a handler,
and synchronous actors are never invoked if they don't need
to be. When a response requires a database operation, a
message is dispatched to a synchronous actor, and the HTTP
worker's underlying `tokio` reactor serves other traffic
while waiting for the future to complete.

### Connection management (#connection-management)

At first glance, introducing synchronous actors might seem
like purely a disadvantage because there's a limit to the
number of parallel operations that they can handle at any
given time.

But this limit can also be seen as a major advantage. While
scaling up a database like Postgres one of the first limits
that you're likely to run into is the one around the
maximum number of simultaneous connections. Even the
biggest instances on Heroku or GCP max out at 500
connections, and the smaller instances have limits that are
_much_ lower. Specifying the number of synchronous actors
in my service also by extension specifies the maximum
number of connections that the service will use, which
gives me perfect control over its connection usage.

!fig src="/assets/rust-web/connection-management.svg" caption="Connections are held only when a synchronous actor needs one."

Synchronous actors also check out individual connections
from a connection pool ([`r2d2`][r2d2]) when starting work
and check them back in after they're done, so when the
service is idle, starting up, or shutting down, it uses
zero connections. This is in sharp contrast to many web
frameworks where the convention is to open a database
connection as soon as a worker starts up, and to keep it
open as long as the worker is alive. That approach has a 2x
connection requirement for graceful restarts because all
workers being phased in immediately establish a connection,
even while all workers being phased out are still holding
onto one.

### The ergonomic advantage of synchronous code (#ergonomics)

There's also something to be said for synchronous
operations and their benefits for development velocity.
While the performance aspects of futures are nice, getting
them properly composed together is very time consuming, and
the compiler messages they generate when you make a mistake
is truly the stuff of nightmares. Writing synchronous code
is faster and easier, and I'm personally fine with a
slight performance hit if it means I can implement my core
domain logic more quickly.

It's also worth noting that even if the synchronous actor
model performs poorly compared to a purely-asynchronous
stack (i.e., futures everywhere), compared to almost any
other framework and programming language, it's really,
_really_ fast. At the end of the day your database is
always going to be a bottleneck for parallelism, and this
model supports about as much parallelism as we can expect
to get from it, while also supporting high concurrency for
any actions that don't need database access.

## Error handling (#error-handling)

Like any good Rust program, APIs almost everywhere
throughout return the `Result` type. Futures also plumb
through their own version of `Result` containing either a
successful result or an error.

I'm using [error-chain][errorchain] to define my errors.
Most are internal, but I've defined a certain group with
the explicit purpose of being user facing:

``` rust
error_chain!{
    errors {
        //
        // User errors
        //

        BadRequest(message: String) {
            description("Bad request"),
            display("Bad request: {}", message),
        }
    }
}
```

When a failure should be surfaced to a user, I make sure to
map it to one of my user error types:

``` rust
Params::build_from_post(log, bytes.as_ref()).map_err(|e|
    ErrorKind::BadRequest(e.to_string()).into()
)
```

After waiting on a synchronous actor and after attempting
to construct a successful HTTP response, I handle a
potential user error and render that as the HTTP response.
The implementation turns out to be quite elegant (note that
in future composition, `then` differs from `and_then` in
that it handles a success _or_ a failure by receiving a
`Result`, as opposed to `and_then` which only receives a
success):

``` rust
let message = server::Message::new(&log, params);

// Send message to synchronous actor
sync_addr
    .send(message)
    .and_then(move |actor_response| {
        // Transform actor response to HTTP response
    }
    .then(|res: Result<HttpResponse>|
        server::transform_user_error(res, render_user_error)
    )
    .responder()
```

Errors not intended to be seen by the user get logged and
`actix-web` surfaces them as a `500 Internal server error`.

`transform_user_error` looks something like this (a
`render` function is abstracted so that we can reuse this
generically between an API that renders JSON responses and
a web server that renders HTML):

``` rust
pub fn transform_user_error<F>(res: Result<HttpResponse>, render: F) -> Result<HttpResponse>
where
    F: FnOnce(StatusCode, String) -> Result<HttpResponse>,
{
    match res {
        Err(e @ Error(ErrorKind::BadRequest(_), _)) => {
            // `format!` activates the `Display` traits and shows our error's `display`
            // definition
            render(StatusCode::BAD_REQUEST, format!("{}", e))
        }
        r => r,
    }
}
```

## Middleware (#middleware)

`actix-web` supports middleware as well. Here's one that
initializes a per-request logger and installs it into the
request's `extensions` (a collection of request state that
will live for as long as the request does):

``` rust
pub mod log_initializer {
    pub struct Middleware;

    pub struct Extension(pub Logger);

    impl<S: server::State> actix_web::middleware::Middleware<S> for Middleware {
        fn start(&self, req: &mut HttpRequest<S>) -> actix_web::Result<Started> {
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

    /// Shorthand for getting a usable `Logger` out of a request. It's also
    /// possible to access the request's extensions directly.
    pub fn log<S: server::State>(req: &mut HttpRequest<S>) -> Logger {
        req.extensions().get::<Extension>().unwrap().0.clone()
    }
}
```

A really nice feature is that request state is keyed to a
_type_ instead of a string that goes into a shared hash
(e.g., Rack in Ruby). This has the benefit of type checking
at compile-time so you can't mistype a key, but it also
allows middleware to control their visibility. If we wanted
to strongly encapsulate the middleware above we could
remove the `pub` from `Extension`, thereby making it
private. Any other modules that tried to access its logger
would be prevented from doing so by the visibility rules
built into the compiler.

### Asynchrony all the way down (#asynchrony)

Like handlers, `actix-web` middleware can be asynchronous
by returning a future instead of a `Result`. This would
allow us to implement something like a rate limiting
middleware that made a call out to Redis in a way that
doesn't block anything else. Did I mention that `actix-web`
is pretty fast?

## HTTP testing (#testing)

`actix-web` documents a few recommendations for [HTTP
testing strategies][actixtesting]. I settled on a series of
unit tests that use `TestServerBuilder` to compose a
minimal app containing a single target handler, and then
execute a request against it. This is a nice compromise
because despite tests being minimal, they nonetheless
exercise an end-to-end slice of the HTTP stack, and are
therefore complete, but also fast:

``` rust
#[test]
fn test_handler_graphql_get() {
    let bootstrap = TestBootstrap::new();
    let mut server = bootstrap.server_builder.start(|app| {
        app.middleware(middleware::log_initializer::Middleware)
            .handler(handler_graphql_get)
    });

    let req = server
        .client(
            Method::GET,
            format!("/?query={}", test_helpers::url_encode(b"{podcast{id}}")).as_str(),
        )
        .finish()
        .unwrap();

    let resp = server.execute(req.send()).unwrap();

    assert_eq!(StatusCode::OK, resp.status());
    let value = test_helpers::read_body_json(resp);

    // The `json!` macro is really cool:
    assert_eq!(json!({"data": {"podcast": []}}), value);
}
```

I also want to call out `serde_json`'s (the standard Rust
JSON encoding and decoding library) `json!` macro, used on
the last line in the code above. If you look closely,
you'll notice that the in-line JSON is not a string --
`json!` lets me write actual JSON notation right into my
code that gets checked and converted to a valid Rust
structure by the compiler. This is _by far_ the most
elegant approach to testing HTTP JSON responses that I've
seen across any programming language, bar none.

## Summary (#summary)

It'd be fair to say that I could've written an equivalent
service in Ruby in a tenth of the time it took me to write
this one in Rust. Some of that is Rust's learning curve,
but a lot of it isn't -- the language is fairly succinct to
write, but appeasing the compiler often takes hours.

That said, over and over I've experienced passing that
final hurdle, running my program, and having it work
_exactly_ as I'd intended it to. Contrast that to an
interpreted language that only works on your 15th try, and
even then you have no idea whether you got the edge
conditions right. Changing things is also well-protected --
it's not unusual for me to refactor a thousand lines at a
time, and once again, have the program run perfectly
afterwards. Anyone who's seen an interpreted language at
production-scale knows that you never deploy a sizable
refactor except in miniscule chunks -- anything beyond that
is too risky.

Should you write your next web service in Rust? I'm not
sure yet, but the language is sure shaping up nicely.

[1] You can think of `tokio` a little like the event loop
    core to runtimes like Node.JS, except one which isn't
    limited to a single core, and with much lower overhead.

[actix]: https://github.com/actix/actix
[actixtesting]: https://actix.github.io/actix-web/guide/qs_8.html
[actixweb]: https://github.com/actix/actix-web
[diesel]: http://diesel.rs/
[errorchain]: https://github.com/rust-lang-nursery/error-chain
[r2d2]: https://github.com/sfackler/r2d2
[techempower]: https://www.techempower.com/benchmarks/#section=data-r15&hw=ph&test=plaintext
[tokio]: https://github.com/tokio-rs/tokio
