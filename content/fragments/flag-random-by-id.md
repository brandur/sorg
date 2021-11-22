+++
hook = "Useful for rolling features out to a subset of users so that it stays enabled or disabled for any specific one deterministically."
published_at = 2021-11-22T17:58:46Z
title = "Implementing a random by ID flag type"
+++

Last week I wrote about [feature flags](/fragments/flags-v-gates), and specifically how I was implementing a light framework in our code. I initially brought in flags that could be fully off, fully on, enabled randomly, or enabled for specific tokens/IDs. This week I put a new type for "random by ID", which I'd always intended to add, but which made good distinct project.

Random by ID differs from "pure" random in that the flag is enabled randomly by ID (instead of purely randomly), and that for any given flag and ID combination, a flag check will always return the same result for its enabled/disabled state.

This is useful from a product perspective because it means that when the ID in question is an account ID, users don't observe non-deterministic behavior as a feature is being rolled out -- once they're in, they're always in, and progressing the rollout just has the effect of sending it out to ever more accounts.

Here's my code:

``` go
// for example
flag := "my_new_feature"
fractionEnabled := 0.5 // = 50% enabled
token := uuid.New()

// map flag + token into hash space
tokenBytes := ([16]byte(token))
dataBytes := append([]byte(flag), tokenBytes[:]...)
hashBytes := sha3.Sum224(dataBytes)

// make result usable as a fraction between 0 to 1
var hashInt big.Int
hashInt.SetBytes(hashBytes[:])
hashInt.Mod(&hashInt, big.NewInt(1000))
fraction := float64(hashInt.Int64()) / 1000.0

isEnabled := fraction < fractionEnabled
```

It's nothing special, and not being an algorithms guy, I'm sure the algorithm police will be out to arrest me for some suboptimality. The broad strokes:

* Start by hashing the flag + token. Recall that a hash function is one which takes an input of arbitrary size and maps it into a result of _fixed_ size. A good hash function (like SHA3) maps any arbitrary values as uniformly as possible into the output so that there's a good distribution.

* Treat the bytes as an int, `mod 1000` it, and turn that into a fraction between 0 and 1 that we can compare against.  We `mod 1000` instead of `mod 100` so that the flag can be enabled by a fraction of a percent like 0.1%.

That's it! A few notable features:

* The reason that we use flag name in addition to token as input is so that any given token doesn't always fall into the same fraction across flags. With just a token, for any new feature we enabled with a flag, the same IDs would always fall into the tier of rollout. Combining with flag name randomizes that while still returning consistent results within any given flag.

* Specific IDs stay enabled as the flag is rolled out. So if we moved a flag from 10% to 20%, the accounts that had been in that original 10% still have the feature enabled at 20%.

## Hashing algorithm (#hash)

I used SHA-3 as a hash algorithm, but for such a simple purpose, it doesn't matter much what gets used. Ever-useful Wikipedia has a great [SHA function comparison chart](https://en.wikipedia.org/wiki/SHA-3#Comparison_of_SHA_functions) showing that SHA-3 is a little slower than SHA-2, and about twice as slow as SHA-1 or MD5.

I used it anyway because it's (1) in the Go stdlib, (2) it's the latest SHA standard, and (3) hashing is so fast that it's the least of my performance problems.

I also experimented with the built-in [`maphash` package](https://pkg.go.dev/hash/maphash), which has the nice properties that (1) it returns a `uint64` sum so you don't have to dip into `bigint`, and (2) it's probably faster than SHA-3. Unfortunately for me, this built-in hasher uses a random seed as input, and the seed is not exportable outside of the current process by design, meaning that I couldn't guarantee consistent flag results outside of a single process. I imagine this choice was made specifically to stop misuses like I was about to make so that they're still free to vary the hashing implementation without breaking external users.

[`KangarooTwelve`](https://en.wikipedia.org/wiki/SHA-3#Later_developments) is known to be faster than SHA-3, but it's not in the standard library, so not worth it. [`crc64`](https://pkg.go.dev/hash/crc64) is also going to be faster than SHA-3, but at this point I just stopped micro-optimizing. Once again, most hash functions are pretty fast.
