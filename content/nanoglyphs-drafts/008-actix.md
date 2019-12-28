+++
published_at = 2019-12-27T06:29:34Z
title = "Actix Web: Optimizing Amongst Optimizations"
+++

![Rust](/assets/images/nanoglyphs/008-actix/rust@2x.jpg)

---

Many in the web development community will already be familiar with the TechEmpower [web framework benchmarks](https://www.techempower.com/benchmarks/) which pit various web frameworks in various languages against each other.

Benchmarks like this tend to draw fire because although results are presented definitely, they can be misleading. Two languages/frameworks may have similar performance properties, but one of the two has sunk a lot more time into optimizing their benchmark implementation, allowing it to pull ahead of its mate. This tends to be less of an issue over time as benchmarks mature and every implementation gets more optimized, but they should be taken with some grain of salt.

That said, even if benchmark games don't tell us everything, they do tell us _something_. For example, no matter how heavily one of the Ruby implementations is optimized, it'll never beat PHP, let alone a fast language like C++, Go, or Java -- the inherent performance disparity is too great. Results might not be perfect, but they do give us a rough idea of relative performance.

## Round 18 (#round-18)

An upset during the latest round ([no. 18](https://www.techempower.com/benchmarks/#section=data-r18&hw=ph&test=fortune), run last July) was [Actix web](https://github.com/actix/actix-web) pulling ahead of the rest of the pack by a respectable margin:

![Fortunes round 18 results](/assets/images/nanoglyphs/008-actix/fortunes@2x.jpg)

TechEmpower runs a few different benchmarks, and this is specifically_Fortunes_ where implementations are tasked with a simple but realistic workload that exercises a number of facets like database communication, HTML rendering, and unicode handling. Here's the official description:

> In this test, the framework's ORM is used to fetch all rows from a database table containing an unknown number of Unix fortune cookie messages (the table has 12 rows, but the code cannot have foreknowledge of the table's size). An additional fortune cookie message is inserted into the list at runtime and then the list is sorted by the message text. Finally, the list is delivered to the client using a server-side HTML template. The message text must be considered untrusted and properly escaped and the UTF-8 fortune messages must be rendered properly.

Fortunes is the most interesting of TechEmpower's series of benchmarks because it does more. Those that do something more simplistic like send a canned JSON or plaintext response have more than a dozen frameworks that perform almost identically to each other because they've all done a good job in ensuring that one piece of the pipeline is well optimized.

## Actix web (#actix)

Actix web is a light framework written in Rust. I wrote about using it for a project [a year and a half ago](/rust-web), and I'm very happy to say that unlike some of Rust's other web frameworks, Actix web has been consistently well-maintained during that entire time and stayed up-to-date with new language features and conventions. Notably, it's 2.0 release which integrates Rust's newly stable standard library futures [shipped this week](https://github.com/actix/actix-web/releases/tag/web-v2.0.0).

When asked about Actix's new lead, Nikolay ([@fafhrd91](https://github.com/fafhrd91)), Actix's creator and stalwart stewards, described the improvements that lead to it in his usual laconic style:

![Actix results explanation](/assets/images/nanoglyphs/008-actix/actix-explanation@2x.jpg)

Personally, I found this information fascinating, and the rest of this edition is dedicated to digging into each of these points in a little more depth. There are a variety of competing Actix implementations (`actix-diesel`, `actix-pg`, etc.) that vary in their components -- I'll be looking specifically at `actix-core`, the current top performer.

---

### Native Postgres drivers (#tokio-postgres)

> new async rust postgres driver. it is rust implementation of psql protocol

Rust has had a [natively-implemented Postgres driver](https://github.com/sfackler/rust-postgres) for quite some time now, but more recently it also has an _asynchronous_ natively-implemented Postgres driver (`tokio-postgres`, a new crate but part of the same project). It plays nicely with Rust's [newly stable](https://blog.rust-lang.org/2019/11/07/Async-await-stable.html) async-await syntax, standard library `Future` trait, and `tokio` 0.2, a new version of `tokio` providing a revamped runtime.

Its asynchronous nature allows pauses on I/O to be optimized away by having other tasks in the runtime be worked during the wait. It also supports [pipelining](https://docs.rs/tokio-postgres/0.5.1/tokio_postgres/#pipelining) which allows the client to be waiting on the responses of many concurrently-issued queries, and which activates automatically whenever multiple query futures are polled concurrently.

### Bottleneck-free concurrency and parallelism (#parallel)

> actix is single threaded, it runs in multiple threads but each thread is independent, so no synchronization is needed

To maximize performance, Actix leverages a combination of threading for true parallelism _and_ an asynchronous runtime for in-thread concurrency. Starting an Actix web server spins up a number of worker threads (by default [equal to the number of logical CPUs](https://actix.rs/docs/server/#multi-threading) on the system) and creates a tokio runtime for async/await handling inside each one.

No state is shared between threads by default, although users can use various Rust primitives to synchronize access to some shared state if they'd like to. In the case of `actix-core`, each thread spawned get its own local Postgres connection, so no thread waits on any other as it's serving requests.

``` rust
struct App {
    db: PgConnection,
    ...
}

fn new_service(&self, _: ()) -> Self::Future {
    const DB_URL: &str = "postgres://...";

    Box::pin(async move {
        let db = PgConnection::connect(DB_URL).await;
        Ok(App {
            db,
            ...
        })
    })
}
```

### High performance escaping (#simd-escaping)

> fortunes template uses simd instructions for html escaping

Some low-level tasks in software are performed so often that they're worth optimizing aggressively, even if the resulting implementation is fiendishly complex compared to the original.

Recall from the description of _Fortunes_ above that fortune messages must be treated as untrusted, and therefore they should be properly HTML escaped before being HTML rendered. Actix escapes using the [`v_htmlescape`](https://github.com/botika/v_escape) crate, which optimizes HTML escaping by leveraging [<acronym title="Streaming SIMD Extensions 4">SSE4</acronym>](https://en.wikipedia.org/wiki/SSE4#SSE4.2), a set of Intel <acronym title="Single instruction, multiple data">SIMD</acronym> instructions that parallelize operations on data at the CPU level. SSE 4.2 in particular adds <acronym title="String and Text New Instructions">STTNI</acronym>, a set of instructions that perform character searches and comparison on two inputs of 16 bytes at a time. Although designed to help speed up the parsing of XML documents, it turns out they can be leveraged for optimizing web-related features as well.

See [`sse.rs`](https://github.com/botika/v_escape/blob/56549a196fbff38a4d3fb7354b8ada586fe074eb/v_escape/src/sse.rs) in the `v_escape` project to get a feel for how this works. Notably, although the code is `unsafe`, the entire implementation is still written in Rust -- no need to drop down to assembly or C as would be common practice in other languages.

### Static dispatch (#static-dispatch)

> actix uses generics extensively. compiler is able to use static dispatch for a lot of function calls

Actix uses generics essentially wherever it's possible to use them, which means that the functions that need to be called as part of an application's lifecycle are determined precisely during compile time. This is a form of [static dispatch](https://en.wikipedia.org/wiki/Static_dispatch), and means that the program has less work to do during runtime. The alternative is dynamic dispatch, where a program uses runtime information to determine where it should route a function call.

For example, an Actix server is built by specifying a `ServiceFactory` whose job it is to instantiate a `Service`, each of which is a trait implemented by a user-defined type (code from `actix-core`):

``` rust
struct App {
    db: PgConnection,
    ...
}

impl actix_service::Service for App {
    ...
}

#[derive(Clone)]
struct AppFactory;

impl actix_service::ServiceFactory for AppFactory {
    type Service = App;
}

fn main() -> std::io::Result<()> {
    Server::build()
        .backlog(1024)
        .bind("techempower", "0.0.0.0:8080", || {
            HttpService::build()
                .h1(AppFactory)
                .tcp()
        })?
        .start();

    ...
}
```

### Request and response pools (#pools)

> actix uses object pools for requests and responses

### Swisstable (#swisstable)

> also it uses high performance hash map, based on google's swisstable

---
