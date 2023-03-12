+++
hook = "Keeping the good parts of util while avoiding lazy dumping grounds."
published_at = 2023-03-12T16:24:16-07:00
title = "Policy on util packages"
+++

Ah `util`, the ubiquitous fixture of practically every project since the invention of code.

I spent my early years programming _expecting_ a util package in every project, oftentimes being the one to create it myself. It wasn't until my third job out of school that I ran into someone who had an allergic reaction to the addition of one and argued strongly against it.

I later came to realize that they were right. The problem with `util` is that it encourages sloppy organization by becoming the dumping ground for everything under the sun. It starts life with limited use, but inevitably grows to house a sprawling, monstrous patchwork of dozens of functions with only the most tenuous relation to each other. Programmers skew lazy, and tend to take the path of least resistance by dumping functions into `util` instead of thinking carefully about what a more appropriate home might be.

That said, sometimes you really just want something util-shaped. A place for a couple static functions that really don't need a more elaborately designed package built around them.

We've come to a compromise of sorts, with two simple rules:

* A general `util` package is not allowed.

* Individual util packages targeted around specific, well-defined concepts _are_ allowed. They're suffixed with `*util` like `emailutil` or `randutil`.

This buys us the convenient ergonomics of util packages, while making sure that dumping grounds do not appear. The addition of a `*util` suffix works particularly nicely in Go because the alternative of giving packages overly simplistic names like `email` or `password` means that these symbols become unusable as variable names in other packages that import them (package names are used to reference construct within them like--e.g. `emailutil.Redact`--and are ambiguous with variable names).

Here's what our package hierarchy looks like right now:

```
util/
    assetutil/
    cookieutil/
    cryptoutil/
    emailutil/
    jsonutil/
    maputil/
    passwordutil/
    ptrutil/
    randutil/
    signingutil/
    sliceutil/
    stringutil/
    testutil/
    timeutil/
    uuidutil/
```

Simple, yes. Obvious, yes. But also an organizational pattern I'd recommend again and again because it takes zero effort and yields good results.