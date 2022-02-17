+++
hook = "Using a custom sampling function to filter out Sentry spans that are never going to be used by anyone."
published_at = 2022-02-17T22:57:39Z
title = "Filtering Sentry spans no one looks at"
+++

Yesterday, I [tweeted about](https://twitter.com/brandur/status/1494025314342637568) how we're finding Sentry spans to be pretty useful for analyzing performance problems. Here's a little addition I put into prod today -- nothing mind-blowing, but useful.

Sentry has a built-in sampling concept, with the idea is that when you're using spans for aggregate analysis, you only need a small subset for them to be useful. We've had our sampling rate cranked to 1.0 for a while, which is somewhat wasteful, but allows us to emit a `sentry_trace_url` to our [canonical log line](/nanoglyphs/025-logs#canonical-log-lines-2) so that you can click through to visualize any request that goes to the API.

{{FigureSingleWithClass "Visualization of a Sentry" "/photographs/fragments/filtering-sentry-spans/span-visualization.png" "overflowing"}}

The downside is that as our API traffic's increased over the last few months, we've had to bump our limits with Sentry to prevent it from discarding requests. A few extra bugs their way isn't a big deal, but we find that request tracing to be a technology that's really useful once in a while, with no one looking at it 99% of the time. Logs are used overwhelmingly more commonly used for observing production.

In addition to a sampling rate, Sentry's SDKs also provide a sampling _function_ to specify custom sampling logic. I made the minor logical step that most of our API traffic is bot-related stuff that no one is _ever_ going to look at. In particular for us, we have a lot of traffic going to our health endpoint as an uptime service monitors us.

Plugging in a function to ignore that health endpoint took almost no time at all:

``` go
{
    ...

    if err := sentry.Init(sentry.ClientOptions{
        Dsn:         dsn,
        Environment: environment,
        TracesSampler: sentry.TracesSamplerFunc(func(samplingContext sentry.SamplingContext) sentry.Sampled {
            return tracesSampler(samplingContext.Span.Context())
        }),

    ...
}

func tracesSampler(ctx context.Context) sentry.Sampled {
    // Always sample all requests unless it's a path that we
    // explicitly skip.
    if requestInfo, ok := requestinfo.FromContext(ctx); ok {
        if !isRequestSampled(requestInfo.Method, requestInfo.URL.Path) {
            return sentry.SampledFalse
        }

        return sentry.SampledTrue
    }

    // Skip any orphan spans that occurred outside of a request.
    return sentry.SampledFalse
}

func isRequestSampled(method, path string) bool {
    if method == http.MethodGet && path == "/healthz" {
        return false
    }

    return true
}
```

Alternatively, we could also have randomized it so as to still sample this endpoint occasionally, but at a much lower rate than most of the API. This would keep Sentry's aggregate analysis features for the health endpoint still useful.

We're currently sending ~350k spans to Sentry a month, and I expect this simple change will wipe out about half of them. For now, we only exclude the health endpoint, but I expect we'll expand the list in the future. Other exclusion ideas:

* `404` spam from probes looking for open WordPress installations and the like.
* Misauthentications (`401`s) caused by bots.
* Rate limited requests.
* It might actually be a good idea to omit all `4xx` requests because they'll tell you far less about an endpoint's performance characteristics than successful ones.
