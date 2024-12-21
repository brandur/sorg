+++
hook = "Avoiding overflows with Go's `time.Duration` in the presence of exponential algorithms."
# image = ""
published_at = 2024-12-21T10:30:01-07:00
title = "Go's maximum time.Duration"
+++

While working on a River bug related to retry policy, I came across a case where it was actually plausible to overflow Go's built-in `time.Duration` and wrap back around to negative number.

A duration has a much simpler representation than a timestamp. It's an `int64` counted in nanoseconds:

``` go
// A Duration represents the elapsed time between two instants
// as an int64 nanosecond count. The representation limits the
// largest representable duration to approximately 290 years.
type Duration int64
```

As the comment states, the maximum duration is about 290 years. More precisely, 292 (non-leap) years, 171 days, and 23 hours:

``` go
func main() {
    const (
        maxDuration time.Duration = 1<<63 - 1

        day  = 24 * time.Hour
        year = 365 * day
    )

    var (
        years        = maxDuration / year
        withoutYears = maxDuration % year

        days        = withoutYears / day
        withoutDays = withoutYears % day
    )

    fmt.Printf("max duration: %dy%dd%s\n", years, days, withoutDays)
}
```

``` sh
$ go run main.go
max duration: 292y171d23h47m16.854775807s
```

292 years is a long time, and it's not likely most programs will need more than that, but our retry algorithm is exponential, and crosses that threshold after 310 retries.

## Compile v. runtime overflow (#compile-v-runtime-overflow)

When performing a direct calculation on a constant, the compiler will detect the overflow:

``` go
func main() {
    const maxDuration time.Duration = 1<<63 - 1
    var maxDurationSeconds = float64(maxDuration / time.Second)

    notOverflowed := time.Duration(maxDurationSeconds) * time.Second
    fmt.Printf("not overflowed: %+v\n", notOverflowed)

    overflowed := time.Duration(int64(maxDuration)+1) * time.Second
    fmt.Printf("overflowed: %+v\n", overflowed)
}
```

``` sh
$ go run main.go
./main.go:15:30: int64(maxDuration) + 1 (constant 9223372036854775808 of type int64) overflows int64
```

But performing the same operation on a variable will happily wrap around:

``` go
overflowed := time.Duration(maxDurationSeconds+1) * time.Second
fmt.Printf("overflowed: %+v\n", overflowed)
```

``` sh
$ go run main.go
not overflowed: 2562047h47m16s
overflowed: -2562047h47m16.709551616s
```

## Little practical use, but well defined (#well-defined)

I [fixed River's back offs at large attempt counts](https://github.com/riverqueue/river/pull/698) by using Go 1.21's `min` function combined with the maximum known number of seconds that'll fit in a `time.Duration`:

``` go
// The maximum value of a duration before it overflows. About 292 years.
const maxDuration time.Duration = 1<<63 - 1

// Same as the above, but changed to a float represented in seconds.
var maxDurationSeconds = maxDuration.Seconds()

func (p *DefaultClientRetryPolicy) NextRetry(job *rivertype.JobRow) time.Time {
    return time.Now().Add(timeutil.SecondsAsDuration(
        p.retrySeconds(len(job.Errors) + 1),
    ))
}

func (p *DefaultClientRetryPolicy) retrySeconds(attempt int) float64 {
    retrySeconds := math.Pow(float64(attempt), 4)
    return min(retrySeconds, maxDurationSeconds)
}
```

After hitting retry attempt 310, the algorithm backs off 292 years at a time. This behavior will never be of any real use to anybody, but I changed it to be _well defined_ behavior of no real use to anybody, with no risk of odd bugs that might otherwise result from an overflow.