+++
hook = "Using the cryptographically secure `crypto/rand` to generate a float between 0 and 1."
published_at = 2022-06-02T22:36:30Z
title = "Generating a random `float64` with `crypto/rand`"
+++

We went through a pen test recently, and one of the low priority items that fell out of it was the use of Go's `math/rand` in a few places instead of the cryptographically secure `crypto/rand`. I shouldn't have been surprised -- security people hate the use of PRNGs, even when their use is deliberate and secure. In this case I'd been using a PRNG to check whether we should enable or disable random flags, and although it wasn't a security problem, we'd recently introduced some rand helper utilities, so I switched them over anyway.

I ran into the problem that [`crypto/rand` provides only a very minimal interface](https://pkg.go.dev/crypto/rand) -- not much more than a reader on top of the system's version of `/dev/urandom`. I'd previously been using `math/rand`'s `Float64` for flag checks and there was no equivalent.

It wasn't too hard to put in, but doing so did make me learn a little bit about how floats are implemented. I'd previously written a helper to supplement the built-in `Int` by providing an easy way of generating a number between 1 an N that doesn't require diving into `math/big`:

``` go
package randutil

import (
	"crypto/rand"
	"math/big"
)

// Intn is a shortcut for generating a random integer between 0 and
// max using crypto/rand.
func Intn(max int64) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		panic(err)
	}
	return nBig.Int64()
}
```

We then build on that to implement `Float64`:

``` go
// Float64 is a shortcut for generating a random float between 0 and
// 1 using crypto/rand.
func Float64() float64 {
	return float64(Intn(1<<53)) / (1 << 53)
}
```

Put simply: generate a number between 0 and 2<sup>53</sup>, then divide that by 2<sup>53</sup> to get a float between 0 and 1.

Why 2<sup>53</sup>? A `float64` in Go consists of three parts:

* 1 bit sign.
* 53 bits "mantissa", otherwise known as a _coefficient_ or _significand_.
* 10 bits signed integer exponent which modifies the magnitude of the mantissa.

So a `float64`'s primary value is within those 53 mantissa bits, which is why we use that number as the bound for our float calculation.

I stole the implementation from `math/rand`'s `Float64` which notes that the `float64(Intn(1<<53)) / (1 << 53)` one-liner _would_ be its implementation, if not for a concern around backwards compatibility dating back to Go 1, which ties it to a slightly more complicated version.
