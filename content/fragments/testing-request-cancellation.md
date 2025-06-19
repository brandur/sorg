+++
hook = "Using built-in `net/http` facilities to make sure that canceled requests are abandoned immediately to save time and resources."
# image = ""
published_at = 2025-06-20T00:16:09+02:00
title = "Testing the graceful handling of request cancellation in Go, 499s"
+++

We had a situation a few days ago where a lazy loading problem in our Ruby code led to long running requests that our Dashboard, with an optimistic five second deadline on backend requests, was timing out. This raised a question in Slack: if our frontend does time out a backend request, does the request keep running? Or does the API know how to save resources by abandoning it midway through?

If the API stack's being bombarded by expensive requests that are largely being canceled early, it's a huge optimization to make sure that they only use the resources that they absolutely to. Requests discarded early stop executing immediately and no further effort is put toward servicing them.

In most code I've ever worked in, I could quite confidently answer the question above with a definitive and resounding "no". Doing a good job of request cancellation requires it be baked quite deeply into language and low level libraries, which isn't common. And even when those handle it well, userland code usually doesn't. Also, cancelling a request midway in services that don't use transactions would be unacceptably dangerous -- [mutated state would be left mutated](/acid#atomicity), and that'd cause untold trouble later on.

## Cancellation in Go (#go-cancellation)

But in a Go stack, the built-in HTTP server [should handle cancellations using context](https://pkg.go.dev/net/http#Request.Context):

> For incoming server requests, the context is canceled when the client's connection closes, the request is canceled (with HTTP/2), or when the ServeHTTP method returns.

And with our code being widely safeguarded by transactions, the feature should even be safe to use!

## Now prove it (#prove-it)

Theory is one thing, but reality is another. If request cancellations indeed work, we should be able to prove it, so I set up a little bootstrap in pursuit of that. To make testing easy, add an artificial API endpoint waiting on sleep or context finished:

``` go
select {
case <-time.After(5 * time.Second):
case <-ctx.Done():
        return nil, ctx.Err()
}
```

Start the API server. Then from another terminal, run cURL and interrupt it after a few seconds:

``` sh
$ curl -i http://localhost:5222/sleep
^C
```

I found that we were handling canceled requests reasonably well, but that the error we were logging wasn't right. The code was checking context cancellation, but getting confused between context that was canceled from the HTTP server versus one canceled by our built-in timeout middleware, improperly sending a `408 Request timeout` to logs.

## Local vs. request context (#local-vs-request)

After a little refactoring, I ended up with this code:

``` go
func (e *APIEndpoint[TReq, TResp]) Execute(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Add a default timeout for all API requests to ensure there's
    // always a backstop in case of degenerate behavior. Rescued
    // below and turned into a more user-friendly error.
    ctx, cancel := context.WithTimeout(ctx, RequestTimeout)
    defer cancel()
    
    ...
    
    ret, err := e.serviceHandler(ctx, req)
    if err != nil {
        if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
            // Distinct error message when the request itself was
            // canceled above the API stack versus we had a
            // cancellation/timeout occur within the API endpoint.
            if r.Context().Err() != nil {
                // This is a non-standard status code (499), but
                // fairly widespread because Nginx defined it.
                err = apierror.NewClientClosedRequestError(ctx, errMessageRequestCanceled).WithSpecifics(ctx, err)
            } else {
                err = apierror.NewRequestTimeoutError(ctx, errMessageRequestTimeout).WithSpecifics(ctx, err)
            }
        }

        WriteError(ctx, w, err)
        return
    }
```

Should a context error occur, we return a `408 Request timeout` in case of a timeout on local `ctx`, but a `499 Client closed request` if context was canceled upstream by the HTTP server canceling `r.Context()`.

`499` isn't real status code, but rather one invented by Nginx which happens to be useful here. It doesn't really matter what status code we use because the end user (who canceled the request before the status code returned) will never see it. It's purely for our own logging and telemetry.

Looking at local logs running the sleep/cancel routine, I now see this:

``` txt
canonical_api_line GET /sleep -> 499 (4.162702459s)
    api_error_cause="context canceled"
    api_error_internal_code=client_closed_request
    api_error_message="Context of incoming request canceled; API endpoint stopped executing."
```

### Generalizing cancellation handling (#generalizing-cancellation)

Although our demo uses an artificial sleep statement, importantly this still works for any normal requests. Our code isn't littered with `<-ctx.Done()` checks all over the place, but it does have a great many database operations like this one:

``` go
account, err := dbsqlc.New().AccountTouchLastSeenAt(ctx, e, apiKey.AccountID)
if err != nil {
    return nil, xerrors.Errorf("error looking up account: %w", err)
}
```

These call into Sqlc which call into Pgx, and Pgx detects a canceled context and sends back an error. In the event of a canceled request, the first database operation will come back with an error that'll bubble back up the stack to our API endpoint infrastructure. There it'll be turned it into a `499`. Subsequent database operations won't run, saving time and resources.

```go
// API service handler error handling. Repeated from above.
ret, err := e.serviceHandler(ctx, req)
if err != nil {
    if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
        // Distinct error message when the request itself was
        // canceled above the API stack versus we had a
        // cancellation/timeout occur within the API endpoint.
        if r.Context().Err() != nil {
            // This is a non-standard status code (499), but
            // fairly widespread because Nginx defined it.
            err = apierror.NewClientClosedRequestError(ctx, errMessageRequestCanceled).WithSpecifics(ctx, err)
        } else {
            err = apierror.NewRequestTimeoutError(ctx, errMessageRequestTimeout).WithSpecifics(ctx, err)
        }
    }

    WriteError(ctx, w, err)
    return
}
```

Pgx is one example of a library that'll check context cancellation, but it'll generally occur in any low level library that's doing I/O. As another example, SDKs like AWS or Stripe will usually go through `net/http`, which will catch them.

With code exercised (and adequate new testing in place), I was confident returning to Slack and declaring that "yes", request cancellation is handled smoothly. I can't say the same about our Ruby code, but that's an adventure for another day.
