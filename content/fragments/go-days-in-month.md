+++
hook = "How to get the number of days in a month for a given year in Go."
published_at = 2021-08-27T05:09:51Z
title = "Go: Number of days in month"
+++

How do you get the number of days in a given month in Go?

Go's `time` package has an internal `daysIn` function to do this, but it's not exported for general use. Why? "Because simplicity" according to Googlers. The preferred version of "simplicity" is to have ten thousand programs reimplement it separately instead.

This single line function will do the trick:

``` go
import "time"

func daysIn(m time.Month, year int) int {
    return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
```

The reason it works is that we generate a date one month on from the target one (`m+1`), but set the day of month to 0. Days are 1-indexed, so this has the effect of rolling back one day to the last day of the previous month (our target month of `m`). Calling `Day()` then procures the number we want.
