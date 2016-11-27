---
hook: Notes on Rust after spending a few weeks with it.
location: San Francisco
published_at: 2016-11-08T04:56:03Z
title: Reflections on Rust
---

I spent the last few weeks learning Rust so that I could use it to build a
Redis module. The result was [redis-cell][redis-cell] and a lot of opinions on
the language, which I've journaled here given that I was still looking at it
with relatively fresh eyes.

It was an interesting project because Redis provides only C APIs to interact
with its module system with the expectation that modules will only be written
in C. I wrote a simple bridge to the API using Rust's FFI (foreign function
interface) module, and wrapped it with a higher level API that would was more
pleasant to use and memory safe. This sort of embedded work is Rust's sweet
spot, and although still relatively new, it's already better suited to it than
any other language that I know of.

I'm going to make a lot of Go comparisons throughout this document. This isn't
because I feel any particular indignation towards the language, but rather
because that besides Rust, it's the most recent language that I spent the time
to learn (and [detailed that experience as well](/go)).

## The Good (#good)

### Error Handling (#error-handling)

Rust has nailed error handling to a greater degree than anyone else. It's not
that its approach was innovative, but more that the right model was baked in
right from the beginning. It's safe, ergonomic, and used uniformly throughout
the language, standard library, and third party packages.

Go introduced the nice C-style convention of moving away from exceptions in
favor of returning errors:

``` go
res, err := getAPIResults()
if err != nil {
    return nil, err
}
```

This is a good idea, but ends up littering your code with an endless amount of
junk boilerplate. Besides making code hard to read and slow to write, it also
makes bugs more likely. For example, I've reversed comparisons by accident a
few times by including `==` instead of `!=`:

``` go
res, err := getAPIResults()
if err == nil {
    return nil, err
}
```

The result is buggy code, and the compiler can't help you find the problem.

Rust's error mechanic is conceptually identical to Go's, but dedicated
in-language constructs make it less prone to problems, easier to read, and very
fast to write. It's core is the `Result` type, which holds either a result or
an error in a type-safe way. Pattern matching makes it trivial to deal with:

``` rust
let res = match get_api_results() {
    Ok(x) => x,
    Err(e) => {
        return e;
    }
};
```

That's still almost as much code as Go, except it's far safer because the
compiler will tell you if you made a mistake.

But the real kicker is the addition of the `try!` macro (now available using
just the `?`) shorthand which boils the above to just:

``` rust
let res = try!(get_api_results());

// or the new and even more efficient syntax
let res = get_api_results()?;
```

Rust also makes other helpers like `unwrap()` available in places where error
handling is really not necessary because an error would be representative of a
profoundly serious application bug. The result is that the error system is
always helping you, and never getting in your way.

### Eliminating Null (#null)

Much like `Result` perfected error handling, the `Option` type allows Rust to
stripe the idea of null by allowing only a value or `None`. If a `None` would
only be indicative of a program bug, you can short circuit things with
`unwrap()`, but otherwise pattern matching makes it easy to handle, and the
compiler won't let you forget to handle both cases.

### Compiler Help (#compiler-help)

Help provided by the compiler in the form of instructional error messages has
gotten ridiculously good. This is hugely important because Rust is a complex
language and beginners need all the help they can get.

Often they'll tell you exactly what you need to do to fix your problem. Take
this one for example, where it gives me the exact line that I need to include
to get my project building (`use std::error::Error`):

```
$ cargo build
   Compiling redis-cell v0.1.0 (file:///Users/brandur/Documents/projects/redis-cell)
error: no method named `description` found for type
`error::CellError` in the current scope
  --> src/redis/mod.rs:65:69
   |
65 |         format!("Cell error: {}\0", e.description())
   |                                       ^^^^^^^^^^^
   |
   = help: items from traits can only be used if the trait
     is in scope; the following trait is implemented but not
     in scope, perhaps add a `use` for it:
   = help: candidate #1: `use std::error::Error`
```

### Pattern Matching (#pattern matching)

Pattern matching is just as cool as the Erlang and Haskell people have been
saying for years. It's not just that it's powerful and succinct, but also that
it's safer because the compiler will force you to handle every possible branch.

### Safety (#safety)

On that note, one of the best aspects of Rust is it's safely; most around
types, but also when it comes to memory, conditionals, and elsewhere. Every
trivial problem will be caught at compile time so that you spend your time
fixing logical bugs instead of typos, segmentation faults, panics, bad casts,
etc.

This is _hugely_ pleasant if you're coming from a Ruby, JavaScript, or Python
where it's necessary to spend hours writing tests for every line of code to
make sure that you didn't type `emiter` instead of `emitter`. But furthermore,
it's nice even coming from Go; no more `nil` or `interface{}`.

### Immutability (#immutability)

Declared variables in Rust are immutable by default:

``` rust
let foo = 7;
```

Trying to assign a new value to `foo` yields:

```
$ cargo test
   Compiling sandbox v0.1.0 (file:///Users/brandur/Documents/projects/rust_sandbox)
error[E0384]: re-assignment of immutable variable `foo`
 --> src/lib.rs:6:9
  |
5 |         let foo = 7;
  |             --- first assignment to `foo`
6 |         foo = 8;
  |         ^^^^^^^ re-assignment of immutable variable

error: aborting due to previous error
```

Luckily, this is easy to fix. The `mut` modifier makes a variable mutable:

``` rust
let mut foo = 7;
```

I figured going in that this would be a huge pain to work with, but it's
amazing how little this gets in your way. I toggle a few variables to be
mutable while mostly everything just stays as the immutable default, helping to
save me from entire classes of logical bugs that could otherwise occur.

### FFI (#ffi)

Rust's FFI (foreign function interface) lets you invoke C programs and have C
programs invoke your Rust program. This module, combined with Rust's very
minimal runtime and lack of a garbage collector, means that it's trivial to
embed Rust into another program's runtime.

As noted above, I built a Redis module in Rust that compiled down to a `.so`
library which Redis could load in and invoke. It's a positive delight to see
this working for the first time. It opens the door to doing systems-level work
in a language besides C/C++ that conveys huge additional safety while trading
off very little performance.

### Destructors (#destructors)

Rust provides destructors via the [`Drop`][drop] trait. For anyone that hasn't
worked with C/C++ recently, destructors are a special function on an object
that gets called when it goes out of scope:

``` rust
impl<T> Drop for Box<T> {
    fn drop(&mut self) {
        // clean-up resources
    }
}
```

This feature is especially useful for building safe wrappers around potentially
unsafe foreign constructs when using the FFI. For example, Redis' C API
requires that every string allocated through it be freed at some point.
Normally this would require an explicit invocation of "free string" before any
possible return in a function that allocates a string; an obviously error-prone
approach that will eventually lead to a memory leak. But by wrapping a string
pointer in a struct and giving it a `drop`, we can get a much stronger
guarantee of an error-free implementation, and with less visual clutter to
boot!

Many languages like Ruby and C# provide functions that get called when an
object is garbage collected ("finalizers") that are similar in spirit to a
destructor, but the destructor's deterministic nature unlocks some neat
approaches. For example, it can be used to implement the [lock section of a
mutex][mutex-guard].

Go's `defer` is an alternative whose implementation I mostly like, but is not
as good because it's so easy to forget to include or remove a defer statement
accidentally. The first Go program that I ever wrote leaked file descriptors
because I'd forgotten to call `Close` on an HTTP response's `Body` (normally
done with a `defer`).

### Macros (#macros)

Macros scare me because their potential for abuse is almost unlimited, but if
their use is reasonably bounded, they can be _very_ useful. I eliminated some
big sections of boilerplate (mostly around debug logging) by defining a couple
simple ones.

### Community (#community)

Especially after saying that the Go community [is one of the worst parts about
working with the language](/go#bad), I'd be remiss to point out that so far
I've found the Rust community to be very pleasant to work with. They're [very
transparent][rust-roadmap-2017] and focus on ease of adoption by new developers
as a core value. The [Rust RFC system][rust-rfcs] is an absolute gold standard,
with new features being discussed out in the open instead of dictated by a
ruling despot.

## Neutral (#neutral)

### Rustfmt (#rustfmt)

Much like Go's gofmt, Rust has an easily accessible rustfmt, which is very
useful for guaranteeing consistent coding conventions across a project.

One really neat thing that rustfmt does is rewrite code so that it falls within
an expected line length (to keep lines under 79 characters wide for example).
This has always been by far the hardest convention to enforce in a project, and
it's great that rustfmt tackled it. Changing options in a `rustfmt.toml` and
watching as `cargo fmt` rewrites vast swatches of code to be compliant is a
borderline magical:

``` toml
# cargo.toml
ideal_width = 79
max_width = 90
write_mode = "Overwrite"
```

One thing that rustfmt notably does not do compared to gofmt is alignment. For
example, like you'd see for fields on a struct definition or values when
initializing a map:

``` go
type Book struct {
	Author     string
	ISBN       string
	NumPages   int
	OccurredAt *time.Time
	Rating     int
	Title      string
}
```

I'm a little on the fence for this behavior because although it makes code look
really nice, it also results in huge whitespace diffs when a new field gets
added to a struct (which I assume is why it wasn't included in rustfmt). But
working in Rust, I found myself missing it. Rustfmt generally makes code look a
lot better than a C++ project, but doesn't leave it beautiful by any means.

### Pragmas (#pragmas)

You'll find any non-trivial Rust program littered with pragma-style directives
(i.e. lines prefixed with a `#`) throughout:

``` rust
#[derive(Debug, PartialEq)]
pub enum KeyMode {
    ...
}
```

They're generally used for two things:

1. Giving types basic implementations for certain common traits. `Debug` or
   `PartialEq` above for example.
2. Suppressing certain compiler warnings which the compiler might otherwise
   throw very liberally. I often find myself needing an `allow(dead_code)`
   annotation as I'm prototyping a function that's not yet called elsewhere.

The reasons that these exist are good, but it leaves code noisier than might be
otherwise preferable.

### Module System (#modules)

Rust's [crate and module system][module-system] is quite a novel
implementation compared to other languages in a few ways:

* Cargo's "crates" (known as packages in many languages) are a first-class
  language construct in that a dependency on one needs to be spelled out
  explicitly in code.
* Module implementations are based heavily on convention. A crate's root source
  file should be found at `src/lib.rs`, and an implementation for `my_module`
  should be at `my_module.rs` or `my_module/mod.rs`.

My read on it is that it works well enough after you understand it, but until
then, it's very confusing.

## The Bad (#bad)

### Batteries Not Included (#batteries-not-included)

Too much has been moved out of the standard library. A few examples of
utilities that you can easily find in other languages, but not in Rust:

* Getopt
* Creating temporary files and directories
* Working with deflate/zip/bzip2 compressed archives

These are all easily available as third party crates, but I think this is an
anti-pattern that has the very negative long-term effect of encouraging a
fractured community even around very common functions. New languages should be
absolutely trying to avoid situations like Ruby's 5+ "pretty standard" HTTP
libraries or Go's twenty different package managers.

### Concurrency (#concurrency)

A few years ago Rust made the decision to remove its [M:N threading
model][m-n-threading] (otherwise known as green threads) in favor of a lighter
runtime. The future of concurrency in the language appears to be "zero-cost
futures" built around the [futures crate][futures-rs] and eventually
[Tokio][tokio] for asynchronous I/O.

While futures are undoubtedly fast, they're the wrong concurrency model for
humans. Here's an example of what application code under the proposed
implementation will look like:

``` rust
id_rpc(&my_server).and_then(|id| {
    get_row(id)
}).map(|row| {
    json::encode(row)
}).and_then(|encoded| {
    write_string(my_socket, encoded)
})
```

It's reasonably readable, but is problematic because the entire codebase gets
littered with boilerplate to support the concurrency model. Even if it's clear
as to what's going on, there's still measurable cognitive overhead in
interpreting and modifying it.

Concurrency is hard to reason about and languages should be providing powerful
foundations to do whatever they can to make it easier. Green threading models
like Go's do exactly that; developers write code normally without thinking
about whether any line is going to have to block on asynchronous I/O or not.
Time is spent solving problems instead of rummaging around in the weeds.

I'd love to see Rust standardize a green threading model. It wouldn't have to
be in core (and therefore not pollute the default Slim runtime), but would be
readily available in cases where the additional overhead is worth it, like a
web service of any kind.

### HTTP (#http)

Rust's HTTP libraries are still not ready. [Rust-http][rust-http] originally
served as a useful stand-in for something more standardized. Later,
[teepee][teepee] appeared. Now it appears that [hyper][hyper] is the future and
has deprecated the first two libraries, but it's still not finished.

Working with HTTP in Rust is a moving target. I recognize that getting to a
feature-rich and stable API is a huge amount of work, but a somber fact of the
real world is that a huge number of developers are doing HTTP work these days,
and a language will be of limited utility until APIs for such are available.

Even once available, the product is likely to sit permanently on the sidelines
outside of the standard library, which I think is problematic for the reasons
listed in [Batteries Not Included](#batteries-not-included).

### Ownership and Cognitive Complexity (#ownership)

Rust uses a very clever ["ownership"][ownership] system to achieve its memory
safety without having to resort to reference counting or a garbage collector.
The ingenuity of this implementation can't be lauded enough. It works, does
exactly what it promises to do, and even manages to produce good compiler
errors when misused.

However, there's no question in my mind that ownership and its associated
complexities are going to curb Rust's adoption. I've spoken to a few friends
who are very senior developers that tried to learn the language, but failed
after the ownership model produced an amount of backpressure that they couldn't
overcome in a reasonable amount of time. I personally feel like I know it well
enough to be practically useful, but wouldn't for a second claim to be able to
explain its more subtle intricacies, even after weeks of use.

## Summary (#summary)

I love Rust. Its huge collection of useful features combined with an eminently
supportive and productive community is enough to make it the language that I'm
most excited about today. I get a Haskell-like feeling with it that the
compiler is trying to help me write bug-free code, but without having to deal
with the artificial obscurity and accumulated language cruft of the former.

While the complexity of its ownership model is likely to present a continued
barrier to adoption, good progress is being made towards improving the
situation with more helpful messages from the compiler (the difference between
trying to learn it now versus a year ago is tangible). I still harbor some
doubts that it's ever going to be practical in a corporate environment where
there's a huge diversity in developer skill levels, but I'm hopeful.

Unrefined HTTP APIs and lack of a modern concurrency model are the last two
missing puzzle pieces preventing me from moving to it from Go and Ruby for the
bulk of my projects. And even without these in places, it's still suitable for
a lot of things I do.

[drop]: https://doc.rust-lang.org/nomicon/destructors.html
[futures-rs]: https://github.com/alexcrichton/futures-rs
[hyper]: https://github.com/hyperium/hyper
[module-system]: https://doc.rust-lang.org/book/crates-and-modules.html
[m-n-threading]: https://mail.mozilla.org/pipermail/rust-dev/2013-November/006550.html
[mutex-guard]: https://doc.rust-lang.org/std/sync/struct.MutexGuard.html
[ownership]: https://doc.rust-lang.org/book/ownership.html
[redis-cell]: https://github.com/brandur/redis-cell
[rust-http]: https://github.com/chris-morgan/rust-http
[rust-rfcs]: https://github.com/rust-lang/rfcs
[rust-roadmap-2017]: https://github.com/aturon/rfcs/blob/roadmap-2017/text/0000-roadmap-2017.md
[servo-easy]: https://github.com/servo/servo/issues?q=is:open+is:issue+label:E-easy
[servo-less-easy]: https://github.com/servo/servo/issues?utf8=✓&q=is:open%20is:issue%20label:"E-less%20easy"%20
[teepee]: https://github.com/teepee/teepee
[tokio]: https://github.com/tokio-rs/tokio
