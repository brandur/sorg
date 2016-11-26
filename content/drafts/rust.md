---
hook: Notes on Rust after spending a few weeks with it.
location: San Francisco
published_at: 2016-11-08T04:56:03Z
title: Reflections on Rust
---

## Good (#good)

### Error Handling (#error-handling)

I believe completely that Rust has nailed error handling to a greater degree
than any other language.

Go introduced the nice C-style convention of moving away from exceptions in
favor of returning errors:

``` go
res, err := getAPIResults()
if err != nil {
    return nil, err
}
```

This is a good idea, but it ends up littering your code with an endless amount
of junk boilerplate. Besides making code hard to read and slow to write, it
also makes bugs more likely. For example, I've reversed comparisons by accident
a few times (i.e. `==` instead of `!=`):

``` go
res, err := getAPIResults()
if err == nil {
    return nil, err
}
```

The result is buggy code, and the compiler can't help you find a problem.

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
```

Rust also makes other helpers like `unwrap()` available to developers so that
the error system is always helping you, but never getting in your way.

### Eliminating Null (#null)

### Compiler Help (#compiler-help)

### Pattern Matching (#pattern matching)

### Safety (#safety)

### Immutability (#immutability)

Amazing how little this gets in your way. Makes programs inherently safer.

### FFI (#ffi)

### Destructors (#destructors)

### Macros (#macros)

The potential for abuse on these is almost unlimited, but within reason they
can be _very_ useful.

## Neutral (#neutral)

### Rustfmt (#rustfmt)

### Pragmas (#pragmas)

### Module System (#modules)

## Bad (#bad)

### Batteries Not Included (#batteries-not-included)

Too much has been moved out of the standard library. For example:

*

It's true that Cargo still makes these modules easy to pull back in, but not
standardizing on basic functions will lead to the inevitable fracturing of
library usage. Having ten different "pretty standard" HTTP libraries like in
Ruby or twenty different package managers like in Go is an anti-pattern of the
highest order.

### Implicit Copies (#implicit-copies)

### Concurrency (#concurrency)

While I have no doubt that this will be fast, I believe that it's the wrong
concurreny model for humans. Besides being hard to reason about, they also
litter your codebase with the abstractions desiged to work with them.

I'd much prefer to see a green threading module in the style of Go that's fully
idiomatic and easy to pull in from the standard library. It's true that would
force some additional runtime overhead, but it's easily worth the trade-off,
especially for the services that would be most like to use it (i.e. Rust code
that's running standalone and not embedded in another process).

### HTTP (#http)

### Lifetimes (#lifetimes)

### Error Verbosity (#error-verbosity)

Defining a basic error struct hard. Defining an error enum hard.

## Summary (#summary)

Systems but maybe too much friction for big services.

Haskell for everyone (too esoteric and too unapproachable given too much accumulated garbage).
