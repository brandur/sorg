+++
hook = "Stripe's API got a new major version, and no one noticed."
# image = ""
published_at = 2024-12-28T23:56:24-07:00
title = "Stripe V2"
+++

I happened to notice by way of a Slack bot today that Stripe released a [V2 version of their API](https://docs.stripe.com/api-v2-overview). I thought this must've been a soft launch right before the holidays, surely to be followed up by a more formal blog post, but the Way Back Machine clocked the page in [early October](https://web.archive.org/web/20241004013621/https://docs.stripe.com/api-v2-overview), making it three months old. It's been there all along, I just hadn't seen it before.

The V1 and V2 APIs are separate namespaces and what's available in V2 is currently very minimal (only events and event destinations), so integrations will still use V1 for almost everything, but the overview page tells us about its aspirational design intentions.

## JSON, with a sprinkling of HATEOAS (#json-hateoas)

A few highlights:

* By far the best and biggest change is that request bodies are sent as JSON instead of `application/x-www-form-urlencoded`. Form encoding isn't the worst thing in the world, but it falls flat on its face when encoding complex data types like arrays and maps (or worse, *nested* arrays and maps). It's also just weird and out of place in 2024. This change should've happened ten years ago.

* Pagination has picked up a hypermedia-esque veneer (see [HATEAOS](https://en.wikipedia.org/wiki/HATEOAS)), returning a `next_page_url` that's requested directly instead of a cursor and having the caller build the next URL themselves.

* The new API is trying to move away from a model where subobjects in an API resource are expanded by default, to one where they need to be requested with an `include` parameter. We had plenty of discussions about this before I left. The purpose of the change is to make API requests faster (Stripe's API is quite slow) by rendering less for most requests. I counted only two places where this is actually used so far though, so time will tell whether the gambit actually succeeds or not.

* Endpoints will try for "real" idempotency where callers can converge failed operations to either success or definitive failure:

    > * When you provide the same idempotency key for two requests:
    >     * API v1 always returns the previously-saved response of the first API request, even if it was an error.
    >     * API v2 attempts to retry any failed requests without producing side effects (any extraneous change or observable behavior that occurs as a result of an API call) and provide an updated response.

    Previously (and still for most endpoints), failures from an intermittent blip or bug were a big problem. The idempotency layer dumbly returned whatever canned response had been recorded on the initial go around, so users wouldn't get closure on what exactly happened. Their best hope would that be a Stripe engineer would eventually repair their charge manually at some later time, and send a webhook about it.

## REST-ish v4-ever (#rest-ish)

Lots of positive progress there, but a new API version also presents an opportunity to clear out blemishes, and I expected to see more of that. A few points that are less good:

* I was hoping they'd fix their verbs to play more nicely with modern REST conventions. Instead of using `POST` everywhere, use `POST` for endpoints that are knowingly not idempotent (without an idempotency key), `PUT` for mutation endpoints that are, and `PATCH` for mutation endpoints that aren't. I admit it's pedantic, but it's so absolutely trivial to implement, and the use of a good verb signals more information than a reader would otherwise have with a cursory glance at API structure.

* They're still doing the RPC-style calls like:

    ```
    POST /v2/core/event_destinations/:id/enable
    ```

    Also pedantic, but `enable` here should theoretically be reserved for a nested resource. I think it's cleaner to model actions as IDs under a shared "actions" subresource:
    
    ```
    POST /v2/core/event_destinations/:id/actions/enable
    ```

## Nouveau DX (#nouveau-dx)

Frankly, I was a bit shocked by how little attention this got. There was a time not too long ago when Stripe cutting a new API version would've been a major event in the tech world, but in three months I didn't come across a single person who mentioned it.

A major part of this is that Stripe is no longer a great technical leader in the same sense that it used to be. But also, as [Colin points out](https://x.com/tweetsbycolin/status/1873241754784411656):

> This is an undeniable sign that "a great REST API" is no longer the benchmark for great DX

That's got to be true too. Few of us want to be making manual HTTP calls out to APIs anymore. These days a great SDK, not a great API, is a hallmark, and maybe even a necessity, of a world class development experience.