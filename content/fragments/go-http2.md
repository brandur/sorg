+++
hook = "A lesson in how API contracts can extend beyond function signatures."
published_at = 2018-08-10T13:46:56Z
title = "The subtleties of API contracts, or how enabling HTTP/2 broke Go clients"
+++

A few weeks ago, having finally retired some old crypto
needed to support pre-1.2 TLS, we realized we had an
opportunity to upgrade OpenSSL on the servers terminating
our HTTPS connections, which gave us the option to add
support for HTTP/2 on `api.stripe.com`.

Given that HTTP/2 is pretty obviously the direction that
internet communication is going, we turned it on
enthusiastically. Every client still had the option to
negotiate the old style of HTTP if it didn't support
HTTP/2, so the feature seemed to be strictly an
enhancement, and one with low likelihood of risk.

It took a few days, but we eventually had a strange [bug
report opened on `stripe-go`][bug], the official Go
language library. To make user integrations more robust,
our libraries [will idempotently retry API
requests][idempotency] that look like they failed due to an
intermittent problem like a read/write timeout, and it was
inside this retry mechanism that the problem presented
itself. The first request would execute as expected, but
subsequent retries would send invalid bodies which our
servers would reject.

## HTTP/2 in Go (#http2-in-go)

To understand the problem's true nature, we have to know a
little bit about the history of HTTP/2 in Go. Go picked up
HTTP/2 support impressively quickly with an initial
low-level implementation landing in
[`golang.org/x/net/http2`][gohttp2]. That package was then
leveraged to retrofit HTTP/2 directly into the existing
APIs of the standard library's [`net/http`][gohttp] for
both clients and servers, and shipped with Go 1.6.

This left Go with an interesting property -- one which as
far as I know is still unique among today's programming
languages. HTTP clients in any Go program compiled with
1.6+ can support HTTP/2 automatically as long as a server
can offer it, and with no changes to user code. For over
two years since Go 1.6's release Stripe's API _couldn't_
offer HTTP/2, so every Go program happily fell back to
speaking plain old HTTP(S).

## The upgrade (#upgrade)

But then the day came when we turned on HTTP/2, and exactly
as intended, our Go clients out in the wild automatically
started using it. This is also when the bug made its
appearance: every initial request that `stripe-go` made
continued to work, but subsequent ones made by the
library's retry mechanism were broken.

It took some digging to understand why. The retry code was
written to be memory efficient so that it would initialize
just one `http.Request`, then reuse that request in a loop
until it either got back a successful response, or gave up:

``` go
req, err := s.NewRequest(method, path, key, "application/x-www-form-urlencoded", params)
...

for retry := 0; ; {
    res, err = s.HTTPClient.Do(req)
    ...


    // Break on success or if we've retried too many times
    // already.
    if !s.shouldRetry(err, res, retry) {
        break
    }
}
```

The package's documentation didn't comment on whether it
was safe to reuse an `http.Request` [1], but pre-HTTP/2, it
was. With HTTP/2, it's now only [sometimes
safe][bodyreuse]. When a request's body is `nil` (say for a
`GET` request) the struct can be reused safely, but if not,
there's a danger that a Goroutine elsewhere could still be
using it, and the struct can only be reused by waiting for
that Goroutine to close it.

## Contracts beyond function signatures (#contracts)

Go had added HTTP/2 support in a way that didn't change the
APIs in `net/http`, but that doesn't necessarily mean that
the package's contract wasn't changed. Hyrum's Law was
coined to define this problem:

> With a sufficient number of users of an API, it does not
> matter what you promise in the contract: all observable
> behaviors of your system will be depended on by somebody.

Go had previously not defined whether it was safe to reuse
a request, but it was. Go 1.6 still didn't define whether
it was safe to reuse a request, but it wasn't, and in the
meantime users started to implicitly depend on the
behavior. The slight shift in contract is about as subtle
as things get, but it was a change in contract nonetheless,
and demonstrates how it's possible to introduce a breaking
change even if every function signature stays the same.

In case you're curious, we fixed the problem with a bit of
a hammer by creating a new reader for the same sequence of
bytes with every request and setting its `Body` directly:

``` go
{
    ...

    if body != nil {
        // We can safely reuse the same buffer that we used
        // to encode our body, but return a new reader to
        // it every time so that each read is from the 
        // beginning.
        reader := bytes.NewReader(body.Bytes())

        req.Body = nopReadCloser{reader}
    }

    res, err = s.HTTPClient.Do(req)
    ...
}

//
// And elsewhere ...
//

// nopReadCloser's sole purpose is to give us a way to turn
// an `io.Reader` into an `io.ReadCloser` by adding a no-op
// implementation of the `Closer` interface.
//
// (We need this because `http.Request`'s `Body` takes an
// `io.ReadCloser` instead of a `io.Reader`.)
type nopReadCloser struct {
	io.Reader
}

func (nopReadCloser) Close() error { return nil }
```

Lastly, I'll note that we also got lucky with this one. The
retry mechanic in Go had been added somewhat recently, and
wasn't turned on by default, so few users were broken by
HTTP/2 being enabled.

[1] It technically does today, but it's still [somewhat
difficult to find][reusedocs].

[bodyreuse]: https://github.com/golang/go/issues/19653#issuecomment-341539160
[bug]: https://github.com/stripe/stripe-go/issues/642
[gohttp]: https://godoc.org/net/http
[gohttp2]: https://godoc.org/golang.org/x/net/http2
[idempotency]: https://stripe.com/blog/idempotency
[reusedocs]: https://go-review.googlesource.com/c/go/+/75671
