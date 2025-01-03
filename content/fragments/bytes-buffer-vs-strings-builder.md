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
BenchmarkBytesBuffer-10          4858627               215.4 ns/op          1280 B/op          5 allocs/op
BenchmarkStringBuilder-10        8368794               144.3 ns/op           752 B/op          6 allocs/op
PASS
ok      github.com/brandur/go-builder-vs-buffer 2.967s
```

So there you have it. At least when it comes to concatenating only strings at relatively modest sizes, `strings.Builder` is about 25% faster. Given that the DX is identical between the two, I'll make it my new default go to.
