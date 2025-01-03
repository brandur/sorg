+++
hook = "Taking five minutes to write a benchmark so I know which of these I should be reaching for first."
# image = ""
published_at = 2025-01-02T22:34:32-07:00
title = "Go's bytes.Buffer vs. strings.Builder"
+++

I was writing some Go code today that generated other Go code. Writing it line by line, mostly in a loop, but with pre- and post-matter.

My usual go to for this type of thing is [`bytes.Buffer`](https://pkg.go.dev/bytes#Buffer), but after I'd finished the implementation, given that I was working entirely with strings, I started to wonder if I should've used [`strings.Builder`](https://pkg.go.dev/strings#Builder) instead.

I realized that I had no idea whether one was faster than the other, so I wrote a quick benchmark to check:

``` go
package main

import (
    "bytes"
    "strings"
    "testing"
)

var fragments = []string{
    "This",
    "is a series of",
    "string fragments",
    "that will be concatenated together",
    "into a single larger string",
    "so that we can",
    "determine which of Go's various",
    "tools for doing this",
    "is most efficient.",
    "I found a few articles",
    "online",
    "but most were poorly cited",
    "or",
    "behind a Medium login wall",
    "or otherwise",
    "not of admirable quality.",
}

func BenchmarkBytesBuffer(b *testing.B) {
    for range b.N {
        var buf bytes.Buffer

        for _, fragment := range fragments {
            buf.WriteString(fragment)
            buf.WriteString(" ")
        }

        _ = buf.String()
    }
}

func BenchmarkConcatenateStrings(b *testing.B) {
    for range b.N {
        var str string

        for _, fragment := range fragments {
            str += fragment
            str += " "
        }
    }
}

func BenchmarkStringBuilder(b *testing.B) {
    for range b.N {
        var sb strings.Builder

        for _, fragment := range fragments {
            sb.WriteString(fragment)
            sb.WriteString(" ")
        }

        _ = sb.String()
    }
}
```

``` sh
$ go test -bench=. -benchmem
goos: darwin
goarch: arm64
pkg: github.com/brandur/go-builder-vs-buffer
cpu: Apple M4
BenchmarkBytesBuffer-10           5013081    217.3 ns/op    1280 B/op    5 allocs/op
BenchmarkConcatenateStrings-10    1603748    753.5 ns/op    5557 B/op    31 allocs/op
BenchmarkStringBuilder-10         6916813    146.9 ns/op    752 B/op     6 allocs/op
PASS
ok      github.com/brandur/go-builder-vs-buffer 4.724s

```

So there you have it. At least when it comes to concatenating only strings at relatively modest sizes, `strings.Builder` is about 33% faster, and 80% faster than [1] than concatenating strings. Given that the DX is identical between the two, I'll make it my new default go to.

[1] Put otherwise, `bytes.Buffer` is 50% slower than `strings.Builder` and concatenating strings is 500% slower.
