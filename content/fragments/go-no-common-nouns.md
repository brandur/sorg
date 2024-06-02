+++
hook = "Avoiding naming Go packages after common nouns like `rate` or `server` so that they don't clash with variable names, and how to find a more fitting name for them instead."
published_at = 2024-06-02T09:50:45-07:00
title = "Go: Don't name packages common nouns"
+++

If there's one single, overriding request that I could make to Go package authors, it's this: **Don't name packages after common nouns.**

Let me pick on [`time/rate`](https://pkg.go.dev/golang.org/x/time/rate) as an example. Great little package, but absolutely awful name.

I see why they did it. In Go, exported names in a package are referenced with their package name, and `rate.Limiter` sure has a nice ring to it.

But it's also bad. Why? Because `rate` is simple, obvious name that we might want to use for a variable name, but doing would make the variable supersede the package in its defined context, so below its definition the package can no longer be referenced:

``` go
import "golang.org/x/time/rate"

// variable name now takes precedent over package
rate := PerSecond(100)

// referencing the package is no longer possible
rate.NewLimiter(...)
```

Yes, package users can give the package a different alias from `import`, but given this will be such a common problem for all a package's users, why make them do this?

## Try adding words (#add-words)

The fix is so easy, so simple, and costs nothing. Come up with an alternative name that's still descriptive, but no longer a simple noun. A strategy that usually works great is to add another word.

How about this? `rate` -> `ratelimit`.

Variable names in Go are camel case, so `ratelimit` will never clobber a variable name. Even if it that wasn't the case, having variables called `rate`, `limit`, or `rateLimiter` is plausible and even likely, but `ratelimit`? Not a thing.

`ratelimit.Limiter` doesn't roll off the tongue in quite the same way, but the tiny cosmetic loss is vastly outweighed by the usability benefit.

Here's some other real world offenders that've I've seen added to projects in earnest over the years (most of which I've long sinced renamed):

``` sh
./client
./db
./limit
./lock
./log
./server
./service
./test
```

Can you imagine trying to figure out what to call a variable representing a server when the name `server` is used by a package? People will use `svr` or some other contrived hack, but let me make the simple point: _why_? Is the package name `server` really that much better than `apiserver` or `httpserver` or `acmeserver`? Why not give the package a better, more descriptive name, and save everybody a lot of trouble.

## Internal packages: Throw a letter on it (#letter-prefix)

Lastly, let me introduce you to my own highly controversial in house technique. If you were about to call a package by a simple noun like "client" or "log" and are having trouble coming up with a good alternative, stop what you were about to do and just add a single letter prefix in front that's keyed to your project name (e.g. "p" = "Crunchy Platform"):

``` sh
./pclient
./pdb
./plimit
./plock
./plog
./pserver
./pservice
./ptest
```

Go purists will absolutely hate this, but it's infinitely better than ending up with a package named after a simple noun. If you're feeling uncreative and are having trouble with names, add a letter and call it a day. They're internal packages, so they're easy to rename without breaking anybody.

You don't want to do this for public packages, but for internal ones that are easy to rename/refactor, it's fine.