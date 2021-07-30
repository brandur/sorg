+++
hook = "Form encoding does have a few niceties that make it amenable to human use, so what's wrong with it? Here we explore why to use JSON instead."
published_at = 2021-07-30T16:56:44Z
title = "Why form encoding in web APIs is an anti-pattern"
+++

A longstanding property of Stripe's API is that it takes its request payloads as `application/x-www-form-urlencoded`, the same content type used by your browser to send HTML forms to servers. It's also the same encoding as what your see in a query string in a URL, and looks like this:

```
a=1&b=2
```

It does have some advantages, the big one being that it's quite easy for humans to write. Compare this to JSON, which has far more special characters, and where the omission of any one is a parse error:

``` json
{"a":1,"b":2}
```

Another upside is that cURL supports it nicely, even having the smarts to submit this request as a `POST` because it includes a `-d` parameter:

``` sh
curl https://api.stripe.com/v1/charges \
  -d amount=2000 \
  -d currency=usd
```

The JSON equivalent is still legible, but not as easy to write:

``` sh
curl https://api.stripe.com/v1/charges \
  -d '{"currency":2000,"currency":"usd"}'
```

So if form-encoding is convenient, why did web APIs broadly towards JSON instead? Part of it was surely JSON's growing popularity, but the other is that although form-encoding does have a couple niceties, it doesn't enhance gracefully beyond the very basics.

I probably need to dissuade you from using it because the world's been moving to JSON for a long time, but for posterity's sake, here's a little rationalization for why it had to go.

## Strings all the way down (#strings)

Form-encoding gives you access to any type that you want as long as it's a string. When you see a payload like `amount=100&capture=true` it's tempting to think you see an integer and boolean, but you don't. If an application wants either of those, it has to do parsing and coercion server-side.

JSON doesn't go overboard with advanced types either, but just the inclusions of all your basics -- string, number, boolean, along with objects and arrays -- gets you a _long_ way.

## Data structures (#data-structures)

It's maps and arrays where form-encoding really breaks down. There's no native support for them, but many years ago Rack (or I believe it was Rack, good information on this is hard to find), brought in these conventions:

```
# array with two elements
arr[]=1&arr[]=2

# map with two keys
map[a]=1&map[b]=2
```

Not too bad right? Well, not too bad yet. Let's look at how to nest two maps:

```
outer[inner][a]=1&outer[inner][b]=2&outer[other][c]=3
```

That's a single outer map with two inner maps at `inner` and other, the first with keys `a` and `b` and the second with key `c`. That's getting a little more awkward, but still not too bad -- it's mostly readable, which is more than you can say for a lot of encoding formats.

How about an array of arrays?

```
arr[][]=1&arr[][]=2&arr[][]=3
```

Nope, that's a no go. You just can't do it.

At Stripe, our our most infamous and least favorite encoding case was an array of maps:

```
arr[][a]=1&arr[][b]=2&arr[][a]=1
```

Can you tell what that's supposed to be? Well according to Rack rules which aren't formally described anywhere, it's a two-element array with each element containing a map, the first with keys `a` and `b` and the second with key `a`. Here it is in JSON:

``` json
arr: [
  {"a":1,"b":2},
  {"a":1}
]
```

How do we know it's an array of two elements instead of one with one element and a repeated key? Well, once again due to underspecified rules, the way this works is that as a decoder is parsing a map inside an array element, as soon as it detects a _duplicate_ map key, it starts a new array item instead of continuing the old one.

Sufficed to say, one needs to be _very_ careful when implementing an encoder that handles these types of structures. If you're encoding two maps and accidentally put your keys in the second one in the wrong order, keys will migrate between array elements unexpectedly.

### The integer-indexed hammer (#integer-indexed-arrays)

After being frustrated for about the hundredth time with subtle encoding/decoding bugs in our implementation, we eventually ditched the Rack's convention and just started assigning integer indexes to array elements:

```
arr[0][a]=1&arr[0][b]=2&arr[1][a]=1
```

This allowed us to greatly simplify our encoding/decoding implementations, even if does introduces it's own possible edges (e.g. skipped indexes). It's better than the alternative, but you should still avoid it if you can-- JSON is easier.

## Why Stripe never switched (#stripe)

The original decision to use form encoding at Stripe was likely an artifact of the time and building on a Rack/Rails-esque stack, but it's notable that JSON never came later either.

Circa 2016 a couple of us tried to push it forward. By then, we'd never be able to realistically drop support for form encoding, but by adding support for JSON we'd be able to simplify our client library implementations (e.g. I wrote [a whole form-encoding module for stripe-go](https://github.com/stripe/stripe-go/tree/master/form)) and reduce the emergent bugs from such.

But even by then, it was already hard. I could've written a middleware for our Ruby stack within a few hours, but we already had a microservice layered on top of the API, and one which we didn't have the necessary privilege to contribute to (PCI stuff). We checked to see if someone else on its team could do it, but were told in no uncertain terms that although it was fine to ask, the project wouldn't be prioritized.

I'm sure that going to JSON is still be possible, but it's one of those logistical problems that gets harder with every passing year at a software company that's constantly growing in scope and complexity -- these days there's at least two more services on the pile that would need to be updated, and probably more. Given that client library implementations are already mature after years of production use in the wild, and with the invention of "simplified" integer-indexed form array encoding, it's not likely to happen.

## JSON is the way (#json)

I know a few ex-Stripe engineers who have either built or are building web APIs at their new companies (including myself), but I don't know any who would go with form-encoding again. JSON isn't problem-free itself, but it works well and is so widely recognized that it's the right choice almost all the time.
