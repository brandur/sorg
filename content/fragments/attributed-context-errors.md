+++
hook = "Making Go's generic context deadline exceeded errors traceable with small helpers that attribute the operation that timed out and indicate how long the timeout was."
# image = ""
published_at = 2026-06-10T07:09:13-05:00
title = "Rich, fully attributed context timeout errors in Go"
+++

In Go, it's generally considered good practice to add [context timeouts](https://pkg.go.dev/context) to network-shaped requests like a database call. Although it doesn't happen often, without a backup context timeout, it's possible for a network request to get gummed up and never come back. Postgres' default `statement_timeout` is zero, or no timeout, so an accidentally long-lived database operation can bring a program to a grinding halt.

Common practice is to add a `context.WithTimeout` scoped into a function where an operation occurs:

``` go
func queryWithoutCauseInError(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
    defer cancel()

    return executeDatabaseOperation(ctx, "SELECT id, email FROM users WHERE id = $1")
}
```

This works fine, except that in case of failure, the error produced is _incredibly_ generic:

``` txt
context deadline exceeded
```

It's a problem in two ways:

* In case of multiple operations with timeouts, you can't tell which failed.
* In case of nested context timeouts, you can't tell which failed.

Both cases are _very_ common, leaving production vulnerable to unattributed context errors whose causes are hard to track down.

## `context.WithTimeoutCause` + context unraveling (#with-timeout-cause)

In Go 1.21, functions like [`context.WithTimeoutCause`](https://pkg.go.dev/context#WithTimeoutCause) were added with the aim of addressing this problem, but because the cause itself lives in context, you have to make sure to check it before it goes out of scope:

``` go
func queryWithCauseInError(ctx context.Context) error {
    ctx, cancel := context.WithTimeoutCause(ctx, 100*time.Millisecond, errQueryTimeout)
    defer cancel()

    err := executeDatabaseOperation(ctx,
        "SELECT id, email FROM users WHERE id = $1")
    if err == nil {
        return nil
    }

    if cause := context.Cause(ctx); errors.Is(cause, errQueryTimeout) {
        return fmt.Errorf("queryWithCauseInError: %w: %w", cause, err)
    }
    return err
}
```

Produces:

``` txt
queryWithCauseInError: timeout after 100ms: context deadline exceeded
```

When used right, it lets you return a considerably better error like `queryWithCauseInError: timeout after 100ms: context deadline exceeded` with attribution to the operation that went awry and how long the timeout was, but it's very manual. Things get ugly and less reliable if you have multiple places an error might be returned:

``` go
func queryMultipleWithCauseInError(ctx context.Context) error {
    ctx, cancel := context.WithTimeoutCause(ctx, 100*time.Millisecond, errQueryTimeout)
    defer cancel()

    err := executeDatabaseOperation(ctx,
        "SELECT id, email FROM users WHERE id = $1")
    if err != nil {
        return withContextCause(ctx, "queryMultipleWithCauseInError",
            fmt.Errorf("load user: %w", err))
    }

    err = executeDatabaseOperation(ctx,
        "SELECT timezone FROM settings WHERE user_id = $1")
    if err != nil {
        return withContextCause(ctx, "queryMultipleWithCauseInError",
            fmt.Errorf("load settings: %w", err))
    }

    err = executeDatabaseOperation(ctx,
        "INSERT INTO audit_log (user_id, event) VALUES ($1, $2)")
    if err != nil {
        return withContextCause(ctx, "queryMultipleWithCauseInError",
            fmt.Errorf("record audit log: %w", err))
    }

    return nil
}

func withContextCause(ctx context.Context, functionName string, err error) error {
    if cause := context.Cause(ctx); errors.Is(cause, errQueryTimeout) {
        return fmt.Errorf("%s: %w: %w", functionName, cause, err)
    }
    return err
}
```

The code above still returns "good" errors, but we have to remember to call our `withContextCause` helper before returning an error on every operation. A forgotten call (which is easy when copy/pasting) reverts to a generic `context deadline exceeded`.

## `AttributedTimeout` — good attribution, no extra work (#attributed-timeout)

I'm experimenting `AttributedTimeout`/`AttributedTimeoutV` helpers that wrap this up into something a little more ergonomic and safer to use:

``` go
// AttributedTimeout runs innerFunc with a timeout.
//
// If innerFunc returns context.DeadlineExceeded because this helper's local
// timeout fired, AttributedTimeout returns an error that includes the operation and wraps
// context.DeadlineExceeded. This makes timeout errors easier to trace back to
// the specific operation that introduced the timeout instead of surfacing only
// the generic "context deadline exceeded" message.
func AttributedTimeout(ctx context.Context, timeout time.Duration, operation string, innerFunc func(ctx context.Context) error) error {
    _, err := AttributedTimeoutV(ctx, timeout, operation, func(ctx context.Context) (struct{}, error) {
        return struct{}{}, innerFunc(ctx)
    })
    return err
}

// AttributedTimeoutV runs innerFunc with a timeout and returns its value. It's the
// same as AttributedTimeout, but also returns a generic value for convenience.
func AttributedTimeoutV[T any](ctx context.Context, timeout time.Duration, operation string, innerFunc func(ctx context.Context) (T, error)) (T, error) {
    // need a specific, local error that we can recognize in case multiple
    // levels of these helpers are wrapped within one another
    timeoutErr := fmt.Errorf("timeoututil.AttributedTimeout: %w", context.DeadlineExceeded)

    ctx, cancel := context.WithTimeoutCause(ctx, timeout, timeoutErr)
    defer cancel()

    ret, err := innerFunc(ctx)
    if err != nil && errors.Is(err, context.DeadlineExceeded) && errors.Is(context.Cause(ctx), timeoutErr) {
        var zero T
        return zero, fmt.Errorf("%s timed out after %s: %w", operation, timeout, err)
    }
    return ret, err
}
```

Now, even with multiple error return sites, we can safely forget about any cause-checking helpers because the outer `AttributedTimeout` will remember to extract a cause and build a friendly error. When multiple `AttributedTimeout`s are nested within one another, it still works, returning the most appropriate operation name.

``` go
func queryMultipleAttributedTimeoutUtil(ctx context.Context) error {
    // note the addition of AttributedTimeout and a work closure
    return timeoututil.AttributedTimeout(ctx,
        100 * time.Millisecond,
        "queryMultipleAttributedTimeoutUtil", // name to attribute in case of error
        func(ctx context.Context) error {
            err := executeDatabaseOperation(ctx,
                "SELECT id, email FROM users WHERE id = $1")
            if err != nil {
                return fmt.Errorf("load user: %w", err)
            }

            err = executeDatabaseOperation(ctx,
                "SELECT timezone FROM settings WHERE user_id = $1")
            if err != nil {
                return fmt.Errorf("load settings: %w", err)
            }

            err = executeDatabaseOperation(ctx,
                "INSERT INTO audit_log (user_id, event) VALUES ($1, $2)")
            if err != nil {
                return fmt.Errorf("record audit log: %w", err)
            }

            return nil
        })
}
```

``` txt
queryMultipleAttributedTimeoutUtil timed out after 100ms:
    load user: context deadline exceeded
```

I know the Go community frowns somewhat on APIs that take closures the way `AttributedTimeout` does, but after a lot of experimentation, I found that it's the cleanest way to make this happen. The clearly-attributed, full-information context timeout errors are well worth a little added syntax.

If you found this interesting, remember to check out my Go job queuing library [River](https://github.com/riverqueue/river), which makes extensive use of multi-level timeouts across many operations.
