+++
hook = "A pair of helper functions `ValOrDefault` and `ValOrDefaultFunc` that can help significantly to clean up Go code around assigning default values."
published_at = 2024-05-15T13:02:29+02:00
title = "ValOrDefault"
+++

**Update:** It turns out that everything I wrote here has been superseded as of Go 1.22, which added [`cmp.Or`](https://pkg.go.dev/cmp#Or), which works almost identically. It can be used to easily assign defaults:

``` go
config = &Config{
    CancelledJobRetentionPeriod: cmp.Or(config.CancelledJobRetentionPeriod, maintenance.CancelledJobRetentionPeriodDefault),
    CompletedJobRetentionPeriod: cmp.Or(config.CompletedJobRetentionPeriod, maintenance.CompletedJobRetentionPeriodDefault),
    DiscardedJobRetentionPeriod: cmp.Or(config.DiscardedJobRetentionPeriod, maintenance.DiscardedJobRetentionPeriodDefault),
}
```

Thanks [_@earthboundkid_](https://github.com/earthboundkid) for the correction.

---

Especially with respect to certain major pain points, Go's traditionally lived by an accidental butchering of Larry Wall's old quote of, "Make easy things easy, and hard things possible" to the much more unfortunate variant of, "Make easy things hard, and hard things impossible."

Historically, an example of this has been initializing default values, like you might do in a configuration struct:

``` go
type Config struct {
    CancelledJobRetentionPeriod time.Duration
    CompletedJobRetentionPeriod time.Duration
    DiscardedJobRetentionPeriod time.Duration
}
```

The caller may have set some properties on an incoming instance, but the others would default to their zero values. Until quite recently, setting defaults on any properties without a value required separate conditionals for each:

``` go
if config.CancelledJobRetentionPeriod == 0 {
    config.CancelledJobRetentionPeriod = maintenance.CancelledJobRetentionPeriodDefault
}

if config.CompletedJobRetentionPeriod == 0 {
    config.CompletedJobRetentionPeriod = maintenance.CompletedJobRetentionPeriodDefault
}

if config.DiscardedJobRetentionPeriod == 0 {
    config.DiscardedJobRetentionPeriod = maintenance.DiscardedJobRetentionPeriodDefault
}
```

The zealots would cite this as perfectly fine and a-feature-not-a-bug because Go prefers more verbosity to make code more readable, while ignoring the fact that every other programming language on Earth has a way to do this that's not only legible, but almost certainly more so than the Go version.

e.g. Ruby:

``` ruby
config.cancelled_job_retention_period ||= CANCELLED_JOB_RETENTION_PERIOD_DEFAULT
config.completed_job_retention_period ||= COMPLETED_JOB_RETENTION_PERIOD_DEFAULT
config.discarded_job_retention_period ||= DISCARDED_JOB_RETENTION_PERIOD_DEFAULT
```

Agree or not, Go's necessary `if` conditions sure made for some ugly ladders of code. In something like a test data factory, which aims to make test objects easy to build with minimum options, but also to allow every separate property to be overridden if necessary, you could have hundreds of lines worth of `if`s for a large object, each of which is a potential bug liability in case of a copy/pasta error.

## Rescued by generics (#rescued-by-generics)

Thankfully, this is an area where Go 1.18's generics really bailed the language out. Most of my projects now have a `valutil.ValOrDefault` [1], the implementation of which is trivial:

``` go
// ValOrDefault returns the given value if it's non-zero, and otherwise returns
// the default.
func ValOrDefault[T comparable](val, defaultVal T) T {
    var zero T
    if val != zero {
        return val
    }
    return defaultVal
}
```

**Update:** See the top of the article, but Go 1.22's built-in `cmp.Or` behaves identically to `ValOrDefault`, and should be preferred over this custom implementation.

With it, we can tighten up the code above considerably:

``` go
config = &Config{
    CancelledJobRetentionPeriod: ptrutil.ValOrDefault(config.CancelledJobRetentionPeriod, maintenance.CancelledJobRetentionPeriodDefault),
    CompletedJobRetentionPeriod: ptrutil.ValOrDefault(config.CompletedJobRetentionPeriod, maintenance.CompletedJobRetentionPeriodDefault),
    DiscardedJobRetentionPeriod: ptrutil.ValOrDefault(config.DiscardedJobRetentionPeriod, maintenance.DiscardedJobRetentionPeriodDefault),
}
```

There's also a `ValOrDefaultFunc` variant that lazily marshals a default by invoking a function, but only if necessary.

``` go
// ValOrDefault returns the given value if it's non-zero, and otherwise invokes
// defaultFunc to produce a default value.
func ValOrDefaultFunc[T comparable](val T, defaultFunc func() T) T {
    var zero T
    if val != zero {
        return val
    }
    return defaultFunc()
}
```

``` go
ClientID: valutil.ValOrDefaultFunc(config.ID, func() string { return defaultClientID(time.Now().UTC()) }),
```

`ValOrDefault` is a very small improvement, but one that could probably help nicen up most Go projects, at least a little. See [River's `valutil` implementation](https://github.com/riverqueue/river/tree/master/internal/util/valutil) for reference.

[1] The helper's package is named `valutil` due to my [policy on util packages](/fragments/policy-on-util-packages):
