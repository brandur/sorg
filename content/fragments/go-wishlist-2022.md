+++
hook = "Go 1.19 has shipped. A few things I'd love to see next."
published_at = 2022-08-11T17:00:25Z
title = "Go wishlist (2022)"
+++

[Go 1.19 was released last week](https://tip.golang.org/doc/go1.19), and although it's mostly made up of smaller incremental improvements, there were a couple nice surprises like that Go doc comments now support Markdown-esque links, lists, and headings which will help prettify documentation across the whole ecosystem.

Especially with the releases of Go modules in 1.11 and generics in Go 1.18, Go's made some huge steps forward that've _obliterated_ the worst aspects of working with the language [1], and as a daily user, I don't have much left to complain about.

That said, it's kind of what I do, so here's my 2022 wishlist of major features Go needs the most:

* **Errors that capture stack traces:** 1.13 brought new helpers into core that'd been proposed in the [x/xerrors](https://pkg.go.dev/golang.org/x/xerrors) package. Unfortunately, one aspect that wasn't brought in was xerror's capability to capture a stack trace at the point an error is generated, and to this day the language's recommended practice is to awkwardly trace errors back to their origin using grep (and hope to god that someone wrapped it so its origin isn't ambiguous). Those of us running prod services where this isn't acceptable still use xerrors, which was a [hair's breadth away from deprecation](https://go-review.googlesource.com/c/xerrors/+/410314).

* **Test assertions:** Recommended practice is to assert in tests using `if` statements, which is laborious and noisy. Almost everyone ends up using [testify's asserts](https://github.com/stretchr/testify), but Go would hugely benefit from a couple built-in assert helpers like Rust's [`assert!`, `assert_eq!`, and `assert_ne!`](https://doc.rust-lang.org/rust-by-example/testing/unit_testing.html).

* **Formatted output:** The output produced by Go's toolchain on something like a failed test looks like someone tried to prove out the [infinite monkey theorem](https://en.wikipedia.org/wiki/Infinite_monkey_theorem), gave up after a few days, and donated the result to Google. Even after having seen tens of thousands of tests fail over the years, it still takes my brain long seconds to find the actual error message amidst the sea of colorless noise. Better formatting and colorization would help immensely in streamlining common workflows.

* **Streaming API:** With the introduction of generics, packages like [`x/maps`](https://pkg.go.dev/golang.org/x/exp/maps) and [`x/slices`](https://pkg.go.dev/golang.org/x/exp/slices) that provide common generic operations on maps and slices are finally possible. That's great, but once you start using them, you feel the goal posts moving, and you start to think how it'd sure be nice if you could chain these things together. [Java 8 introducing a streaming API](https://docs.oracle.com/javase/8/docs/api/java/util/stream/Stream.html) and it did wonders for cutting down on boilerplate in the language (see also C#'s [LINQ](https://docs.microsoft.com/en-us/dotnet/csharp/programming-guide/concepts/linq/)). Go could use the same.

    ``` java
    int sum = widgets.stream()
                     .filter(w -> w.getColor() == RED)
                     .mapToInt(w -> w.getWeight())
                     .sum();
    ```

    This would probably necessitate short lambda syntax like Java's stabby operator, which would also be pretty freaking great.

* **Error handling:** The best for last. Go's `if err != nil {}` boilerplate that's needed after practically every statement is infamous far and wide across the programming world, with its champions citing how it does wonders for readability, but the rest of us finding that it makes code noisy and unnecessarily slow to write. Back when Rust introduced its `try!` macro (later becoming the `?` operator) proponents of endless extraneous conditional expressions screamed that the sky was falling. Actually, nothing bad happened, and code became easier to read and faster to write with no downsides to speak of.

[1] Package management and generics being the big ones, but also `go:embed`, built-in error wrapping with `%w`, `strings.Cut`, `T.Setenv`, etc.
