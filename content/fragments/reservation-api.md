+++
hook = "A nicer API for conditional rate limiters."
published_at = 2022-03-27T17:33:40Z
title = "Reservation-style rate limiting APIs"
+++

In many cases, the basic process for rate limiting is straightforward. For example to rate limit the number of requests allowed for an account, we get pseudocode like this:

``` go
rateOK := rateLimit(accountID)
if !rateOK {
    return ErrTooManyRequests
}
serveRequest()
```

There is another class of what I'll refer to as "conditional" rate limiting that makes things a little more complex. The canonical example is rate limiting invalid authentication attempts by IP. A naive version of that looks like:

``` go
rateOK := rateLimit(ip)
if !rateOK {
    return ErrTooManyRequests
}

tx := connPool.Begin()
authOK := authenticate(tx)
```

But this code has a problem: by rate limiting before the authentication check, it conflicts with legitimate traffic. It's a common pattern to be aggressive limiting invalid authentication attempts (e.g. 30/hour) while allowing a much greater rate for valid traffic (e.g. 100/second), which isn't allowed for here.

## Post-authenticate (#post-authenticate)

A simple-but-wrong fix for invalid requests interfering with legitimate traffic is to move the rate limit check after the authentication step:

``` go
tx := connPool.Begin()
authOK := authenticate(tx)

if !authOK {
    rateOK := rateLimit(ip)
    if !rateOK {
        return ErrTooManyRequests
    }

    return ErrUnauthorized
}
```

But this code is still buggy. Although it'll eventually start returning `ErrTooManyRequests`, it still allows an attacker to iterate because rate limiting only occurs after authentication is checked -- the attacker just needs to know to ignore any `ErrTooManyRequests` results.

## Multi-part (#multi-part)

The real fix is to rate limit in two parts. First, with a non-blocking check on remaining limit (`tryRateLimit`) that would fail with no limit left, but not use rate so as not interfere with valid requests. Rate is then consumed further down, only after failed authentication:

``` go
if !tryRateLimit(ip) {
    return ErrTooManyRequests
}

tx := connPool.Begin()
authOK := authenticate(tx)

if !authOK {
    rateLimit(ip)
    return ErrUnauthorized
}
```

This is how most conditional limiters were written at Stripe. It works, but is kind of ugly, and prone to misuse if either call is forgotten or accidentally refactored out.

## Reservations (#reservations)

Rate limiting with a reservation API is the same thing, but nicer, and less prone to error. The initial non-consuming check is replaced with `Reserve` which immediately consumes rate. Later, if the authentication is deemed _valid_, the procured reservation is cancelled, which cedes the consumed rate back to the usable pool:

``` go
rateOK, _, reservation := mw.limiterInvalid.Reserve(ip)
if !rateOK {
    return ErrTooManyRequests
}

tx := connPool.Begin()
authOK := authenticate(tx)
if authOK {
    reservation.Cancel()
}
```

Aside from being prettier, another benefit is that languages like Go will help protect against refactoring regressions -- if `reservation.Cancel` was to be accidentally removed, the compiler would complain that `reservation` isn't in use and die.

We're building a [single dependency stack](/fragments/single-dependency-stacks) and are using the in-memory [`golang.org/x` rate limiting package](https://pkg.go.dev/golang.org/x/time/rate) with some augmentations to support multiple keys. It has a reservation-based API, which is where the idea came from. The same API could be supported with a Redis-based package without much trouble.
