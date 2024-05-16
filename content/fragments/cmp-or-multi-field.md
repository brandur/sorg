+++
hook = "A more elegant way to sort on multiple fields using Go 1.22's `cmp.Or` helper."
published_at = 2024-05-16T08:55:29+02:00
title = "Use of Go's `cmp.Or` for multi-field sorting"
+++

The [fragment I wrote on a value helper](/fragments/val-or-default) yesterday was upstaged entirely by the fact that I'd missed a new helper that came in with Go 1.22, [`cmp.Or`](https://pkg.go.dev/cmp#Or). (These minor embarassments happen occasionally, and is part of the reason I write these pieces. It's better than be corrected than wrong!)

While I'd expect `cmp.Or` to be mainly useful in cases like setting defaults as a one-liner, its docs suggest that another place where it slots in nicely is for use with `slices.SortFunc` when sorting on multiple fields.

The func-less `slices.Sort` works fine more many cases, and where something less standard is needed, `SortFunc` steps in:

``` go
slices.SortFunc(tags, func(a, b marketplacetypes.Tag) int {
    return strings.Compare(*a.Key, *b.Key)
})
```

When sorting a common internal type that's not `cmp.Ordered`, a nice pattern is to extract a sort function to a helper package so its sorts become clean one-liners elsewhere:

``` go
package uuidutil

import (
    "bytes"
    "slices"

    "github.com/google/uuid"
)

func Compare(u1, u2 uuid.UUID) int {
    u1Bytes := [16]byte(u1)
    u2Bytes := [16]byte(u2)
    return bytes.Compare(u1Bytes[:], u2Bytes[:])
}
```

``` go
slices.SortFunc(metricIDs, uuidutil.Compare)
```

## Multi-field sorting made succinct (#multi-field-succinct)

Newly equipped with the knowledge of `cmp.Or`, I spelunked our project looking for code to clean up, and found this example of a multi-field sort. It's dense, but trying to achieve something very simple: sort on name first, then Postgres server ID, then discriminator string. Put charitably, it's quite awkward, and another one of those little blemishes in Go that we didn't like to talk about:

``` go
slices.SortFunc(metricView.Series, func(a, b *MetricsSeries) int {
    nameCmp := strings.Compare(a.Name, b.Name)
    if nameCmp != 0 {
        return nameCmp
    }

    postgresServerCmp := uuidutil.Compare(a.postgresServerID, b.postgresServerID)
    if postgresServerCmp != 0 {
        return postgresServerCmp
    }

    return strings.Compare(a.discriminator, b.discriminator)
})
```

And its replacement using `cmp.Or`:

``` go
slices.SortFunc(metricView.Series, func(a, b *MetricsSeries) int {
    return cmp.Or(
        cmp.Compare(a.Name, b.Name),
        uuidutil.Compare(a.postgresServerID, b.postgresServerID),
        cmp.Compare(a.discriminator, b.discriminator),
    )
})
```

_Much_ better. `cmp.Compare` returns zero on equality, which also happens to be the zero value for the `int` type. So as each comparison runs and returns equality, `cmp.Or` advances to the next argument until a non-zero result is found.