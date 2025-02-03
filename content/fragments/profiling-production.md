+++
hook = "Using Go's `runtime.MemStats` and canonical log lines to isolate huge memory allocations to a specific endpoint."
# image = ""
published_at = 2025-02-02T18:02:27-07:00
title = "Profiling production for memory overruns + canonical log stats"
+++

You're only lucky for so long. After four years of running our Go API in production with no memory trouble whatsoever, last week we started seeing instantaneous bursts of ~1.5 GB suddenly allocated, enough to cause Heroku to kill the dyno for being "vastly over quota" (our steady state memory use sits around ~50 MB, so we run on 512 MB dynos).

This was of course, concerning. We were only experiencing a few of these a day, but with no idea what was causing them, and having appeared very suddenly, we had to assume that they might get more frequent. Not only is the API suddenly being taken offline at any moment is a bad place to be UX-wise, and even with our careful use of transactions, makes resource leaks between components possible.

## Alloc delta (#alloc-delta)

To localize the problem, I used Go's [`runtime.MemStats`](https://pkg.go.dev/runtime#MemStats) in conjunction with our [canonical API lines](/nanoglyphs/025-logs), making a new `total_alloc_delta` property available to see how many allocations took place during the period of an API request:

``` go
func (m *CanonicalAPILineMiddleware) Wrapper(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var (
            memStats      runtime.MemStats
            memStatsBegin = m.TimeNow()
        )
        runtime.ReadMemStats(&memStats)
        var (
            memStatsBeginDuration = m.TimeNow().Sub(memStatsBegin)

            // TotalAlloc doesn't decrement on heap frees, so it gives
            // us useful info even if the GC runs during the request.
            totalAllocBegin = memStats.TotalAlloc
        )

        // API request served here
        next.ServeHTTP(w, r)
    
        // Middleware continues ...
        memStatsEnd := m.TimeNow()
        // Since we're only interested in one field, reuse the same
        // struct so we don't need to allocate a second one.
        runtime.ReadMemStats(&memStats)
        var (
            memStatsEndDuration = m.TimeNow().Sub(memStatsEnd)
            totalAllocDelta     = memStats.TotalAlloc - totalAllocBegin
        )
        
        logData := &CanonicalAPILineData{
            ID:                   m.ULID.New(),
            HTTPMethod:           r.Method,
            HTTPPath:             r.URL.Path,
            ...
            ReadMemStatsDuration: timeutil.PrettyDuration(memStatsBeginDuration + memStatsEndDuration),
            TotalAllocDelta:      totalAllocDelta,
        }

        plog.Logger(ctx).WithFields(structToFields(logData)).
            Infof(
                "canonical_api_line %s %s -> %v %s(%s)",
                r.Method,
                routeOrPath,
                logData.Status,
                idempotencyReplayStr,
                duration,
            )
```

`MemStats` provides a large bucket of properties to pick from, but `TotalAlloc`'s a useful one because it represents bytes allocated to the heap, but unlike similar stats like `HeapAlloc`, it's monotonically increasing. It's not decremented as objects are freed:

``` go
// TotalAlloc is cumulative bytes allocated for heap objects.
//
// TotalAlloc increases as heap objects are allocated, but
// unlike Alloc and HeapAlloc, it does not decrease when
// objects are freed.
TotalAlloc uint64
```

This is good because it means that all API requests will end up with the same memory heuristic, and made roughly comparable. Garbage collection may or may not occur during a request. Using `TotalAlloc` makes it irrelevant whether it did or not.

With that deployed, I can search logs for outliers (`:>500000000` means greater than 5 MB):

``` txt
source:platform app:app[web] canonical_api_line (-http_route:/health-checks/{name})
    total_alloc_delta:>500000000
```

And voila, we turn up the bad ones. Here, an API request that spiked memory a full 5 GB!

``` txt
Jan 29 10:18:33 platform app[web] info canonical_api_line
    POST /queries -> 503 (2.53252138s)
total_alloc_delta=5008335944
```

## Parallel allocations (#parallel-allocations)

The use of `TotalAlloc` is imperfect because it not only tracks allocations of the current API request, but allocations across the current API request _and_ all parallel requests.

We can see this effect through false positives:

``` txt
Feb 1 23:07:18 platform app[web] info canonical_api_line
    GET /clusters/{cluster_id}/databases -> 504 (2m57.322010348s)
total_alloc_delta=743772480
```

It looks like this API request allocated 744 MB, but what actually happened is that it was a bad timeout that executed for a full three minutes [1]. During that time, other API requests served in the interim allocated the majority of that memory.

## Pprof to S3 (#pprof-to-s3)

Getting our memory overruns localized to a particular endpoint was good, but even having done that, I'd need a little more help to figure out where the rogue memory was going. To that end, I put in one more clause in the middleware so that in case of a huge overrun, the process dumps a [pprof](https://github.com/google/pprof) heap profile to S3:

``` go
    ...

    // If we used a particularly huge amount of memory during the
    // request, upload a profile to S3 for analysis. Buckets have a
    // configured life cycle so objects will expire out after some
    // time.
    if err := m.maybeUploadPprof(ctx, logData.RequestID, totalAllocDelta); err != nil {
        plog.Logger(ctx).Errorf(m.Name+": Error uploading pprof profile: %s", err)
}

...

const pprofTotalAllocDeltaThreshold = 1_000_000_000

func (m *CanonicalAPILineMiddleware) maybeUploadPprof(ctx context.Context, requestID uuid.UUID, totalAllocDelta uint64) error {
    if !m.pprofEnable || totalAllocDelta < m.pprofTotalAllocDeltaThreshold {
        return nil
    }

    profKey := fmt.Sprintf("%s/pprof/%s.prof", m.EnvName, requestID)

    var buf bytes.Buffer
    if err := pprof.WriteHeapProfile(&buf); err != nil {
        return xerrors.Errorf("error writing heap profile: %w", err)
    }

    if _, err := m.aws.S3_PutObject(ctx, &s3.PutObjectInput{
        Body:   &buf,
        Bucket: ptrutil.Ptr(awsclient.S3Bucket),
        Key:    &profKey,
    }); err != nil {
        return xerrors.Errorf("error putting heap profile to S3 at path %q: %w", profKey, err)
    }

    plog.Logger(ctx).Infof(m.Name+": pprof_profile_generated_line: TotalAlloc delta %d exceeded %d; generated pprof profile to S3 key %q",
        totalAllocDelta, m.pprofTotalAllocDeltaThreshold, profKey)

    return nil
}
```

Our memory problem ended up being a queries endpoint that was overly willing to read giant result sets into memory, then serialize the whole thing into a big JSON buffer for response, which was also pretty indented (and in Go's `encoding/json`, indenting a JSON response requires a _second_ giant buffer 2x the size of the first one). I fixed it by reducing the maximum number of rows we were willing to read into the response.

I'm not expecting to run into new memory overruns or leaks anytime soon, but I left the pprof code in place for the time being. It only does work in case of huge memory increases so there's no performance penalty most of the time, and it might come in handy again.

## Stop the world (#stop-the-world)

A token glance at the implementation of `runtime.ReadMemStats` looks a little concerning:

``` go
// ReadMemStats populates m with memory allocator statistics.
//
// The returned memory allocator statistics are up to date as of the
// call to ReadMemStats. This is in contrast with a heap profile,
// which is a snapshot as of the most recently completed garbage
// collection cycle.
func ReadMemStats(m *MemStats) {
    _ = m.Alloc // nil check test before we switch stacks, see issue 61158
    stw := stopTheWorld(stwReadMemStats)

    systemstack(func() {
        readmemstats_m(m)
    })

    startTheWorld(stw)
}
```

To produce accurate stats, the runtime needs to "stop the world", meaning that all active goroutines are paused, a sample taken, and resumed.

Intuitively, that seems like it could be pretty slow, and some initial googling seemed to confirm that. However, I later found a [patch from 2017](https://go-review.googlesource.com/c/go/+/34937) that'd improved the situation considerably by doing cumulative tracking of relevant stats so only a very brief stop the world was required. It indicated a reduction in timing down to 25µs, even at 100 concurrent goroutines.

I added a separate log stat to see how long my two `ReadStatMems` calls were taking, and found they were averaging ~100µs for both:

``` txt
read_mem_stats_duration=0.000098s
read_mem_stats_duration=0.000110s
read_mem_stats_duration=0.000113s
read_mem_stats_duration=0.000126s
read_mem_stats_duration=0.000123s
read_mem_stats_duration=0.000084s
read_mem_stats_duration=0.000091s
read_mem_stats_duration=0.000092s
read_mem_stats_duration=0.000090s
read_mem_stats_duration=0.000083s
```

That's 50µs per invocation instead of 25µs, but given that a single DB query takes an order or two of magnitude longer at 1-10ms, a little delay to get memory stats is acceptable. If our stack was hyper performance sensitive or saturated with huge request volume, I'd take it out.

[1] We're supposed to time out faster than that. I'll have to look into why this request took so log.
