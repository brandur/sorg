---
title: Building Language Toolchains with Batteries Included
published_at: 2017-12-29T21:36:25Z
location: Calgary
hook: TODO
---

Imaginary language.

Gems.

Bundler.

Rbenv.

Ruby. Python looks remarkably similar.

Rust is a good example. Compiler, Cargo, 

`rustfmt` and `cargo` are different executables, but
packages are so native to the language that far and away
the more idiomatic way to build anything is to just use
`cargo`.

How are updates done? `rustup`:

``` sh
rustup ...
```

Getting and installing nightly sounds painful. Nope:

`rustup` is also core to Rust, so much so that Cargo
understands how to invoke itself on alternate toolchains:

```
cargo +nightly build
```

`rustfmt`

## Best practice by default (#best-practice)

While the idea of a package and modularity is fairly
comprehensible, newer developers won't understand why they
should have something like a `Gemfile` and `Gemfile.lock`
because the problems they're designed aren't intuitive
until you run into them. Even intermediate developers won't
understand why something like `rbenv` should exist because
you don't see its value until you're working between
multiple projects and need to start running Ruby upgrades.

Building language toolchains to include batteries is a
great way to not only lower the bar to getting started (in
the case of something like `rustup`), but also to ensure
that everyone is following best practices from day one,
whether they know it or not. Even if it doesn't make sense
in the beginning, they'll come to appreciate it over time.
