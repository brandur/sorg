+++
hook = "Using `goleak` to find leaks early and avoid production pain."
published_at = 2022-12-18T20:56:58Z
title = "Habitually testing for goroutine leaks"
+++

The hardest bug I've ever squashed in Go wasn't a memory leak (the garbage collector's generally got your back there), but a goroutine leak, in which I was regularly starting goroutines that were getting parked on a wait somewhere rather than exiting cleanly. The overhead of any given one of them was small, but in aggregate eventually enough take down my program.

I recently took to shoring up our codebase by introducing the use of Uber's [goleak](https://github.com/uber-go/goleak) in packages at risk of accidentally leaving something running. It runs after a test (or tests of an entire package) and errors if extra goroutines are found.

Luckily, no serious leaks came up, although there were some minor ones. I was impressed that we didn't find more given tens of thousands of lines of preexisting code and the use of hundreds of third-party modules, which is a testament of the Go ecosystem's high average quality.

In the cases they are found, leaky modules are hard to fix because they're in external code and may only be a transitive dependency. We ended up wrapping goleak's helpers to automatically ignore known offenders:

``` go
package prequire

var knownGoroutineLeaks = []goleak.Option{
    // `ptesting` starts up some pools that run for the duration of
    //  the test suite. Ignore pgxpool's background goroutine.
    goleak.IgnoreTopFunction("github.com/jackc/pgx/v4/pgxpool.(*Pool).backgroundHealthCheck"),

    // OpenCensus boots goroutines on `init`, and even though we only
    //  have it asa transitive dependency and it's not used, even that's
    //  enough to have itboot its workers.
    //
    // There's an open GitHub issue:
    //
    // https://github.com/census-instrumentation/opencensus-go/issues/1191
    //
    // I opened a fix, but it may or may not make it in:
    //
    // https://github.com/census-instrumentation/opencensus-go/pull/1287
    goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
}

// VerifyNoGoroutineLeak verifies that no goroutines are still around
//  after atest case runs, allowing for some extraneous goroutines
//  that we can't do muchabout.
func VerifyNoGoroutineLeak(t *testing.T, options ...goleak.Option) {
    t.Helper()
    goleak.VerifyNone(t, append(knownGoroutineLeaks, options...)...)
}

// VerifyNoGoroutineLeakMain is similar to the above, but runs in a
// special GoTestMain function. It's purpose is to check no leaks once
// at the end of the test runs for an entire package.
func VerifyNoGoroutineLeakMain(m *testing.M, options ...goleak.Option) {
    goleak.VerifyTestMain(m, append(knownGoroutineLeaks, options...)...)
}
```

It works, but it's something to be careful about -- it'll be far easier and more tempting to add new leaks to the ignored list rather than dig in and fix a root problem.

## Goroutine hygiene (#goroutine-hygiene)

If you're writing a package, a couple guidelines for goroutines that'll benefit users:

* Don't start package-level goroutines.
* Definitely don't start goroutines in `init`.

Package-level goroutines are a problem because it's not clear that they have to be stopped, and whose responsibility it is to do so. The superior alternative are goroutines local to a `struct` instance, and which has a `Stop` or `Close` function.

Goroutines in `init` are a problem because as noted above, they'll start even if the package wasn't directly used. This makes it _really_ non-obvious where they should be shut back down, if there's even a way of doing so.
