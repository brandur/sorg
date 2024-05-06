+++
hook = "Using Go's `DisallowUnknownFields` option to improve an API's integration experience by making paramter naming mistakes faster to resolve."
location = "Berlin"
published_at = 2024-05-05T08:21:27+02:00
title = "Web APIs: Enriched DX By Disallowing Unknown Fields"
+++

Go's JSON library provides the [decoder option `DisallowUnknownFields`](https://pkg.go.dev/encoding/json#Decoder.DisallowUnknownFields) which even if not intuitively obvious, is a handy option fo adding a layer of improved DX to web APIs. As the name would suggest, it causes a decoder to error when encountering a property in a JSON object being decoded that's not present in the struct being decoded to.

``` go
type Request struct {
    Message string `json:"message"`
}

data := `{"message":"Hello.","unknown":"Not a field on the struct."}`

decoder := json.NewDecoder(bytes.NewReader([]byte(data)))
decoder.DisallowUnknownFields()

var req Request
if err := decoder.Decode(&req); err != nil {
    log.Fatal(err) // json: unknown field "unknown"
}
```

---

When a user is integrating a web API, especially in the beginning, it's common for the initial prototype to be written by a human, and humans are prone to making mistakes. Say you're trying to programmatically procure an access token against `POST /access-tokens`. The endpoint takes an optional parameter called `expires_in` which is a number of seconds after which the new access token will expire automatically. By virtue of reading the documentation slightly wrong, you're accidentally sending `expires: 3600` instead of `expires_in: 3600`. The result is that your requested expiry time is silently ignored, not only producing the wrong result, but possibly even a security leak as your account accidentally amasses access tokens that never expire.

`DisallowUnknownFields` widely fixes this class of mistake for all an API's users. Some code extracted from our API:

```go
decoder := json.NewDecoder(bytes.NewReader(rawPayload))

// Balks if finding fields in the request payload that don't map to anything
// in the target request structure. Acts as a small DX aid for users who may
// have accidentally misnamed a field.
//
// Specific API endpoints can invert this behavior through and option while
// defining the endpoint.
if !allowUnknownJSONFields {
        decoder.DisallowUnknownFields()
}

if err := decoder.Decode(v); err != nil {
    apierror.NewBadRequestError(
        r.Context(),
        fmt.Sprintf("Invalid JSON in request body: %s.", err),
    ).Write(r.Context(), w)
    return nil, false
}
```

Now, sending `expires` instead of `expires_in` is an error that tells the user exactly what's wrong:

```sh
$ curl -i -H "Authorization: Bearer $CRUNCHY_API_KEY" \
    -H "Content-Type: application/json"
    -X POST $CRUNCHY_API_URL/access-tokens -d '{"expires":3600}'

HTTP/2 400
{
    "message":"Invalid JSON in request body: json: unknown field \"expires\".",
    "request_id":"5d2078fe-6ea5-4f41-816e-4717cf6c22b7"
}
```

It's a feature that's not needed every day, but it's easy to implement, and the day it is, it'll save hours worth of time and frustration.

## Caveats and edges (#caveats-and-edges)

There are a few sharp edges to be aware of. They're easy to avoid once you know about them, but aren't totally apparent for those integrating the pattern for the first time.

### Turning it on safely (#safely-on)

If you have an existing API with existing users, `DisallowUnknownFields` isn't universally safe to turn on because there may be integrations out there that have been sending invalid JSON fields for years, but which was never a problem before. Those previously happy users become unhappy when disallowing unknown fields suddenly breaks all their requests.

You can still turn it on, but doing so takes a few more steps:

* Start by organizing the API by pre and post `DisallowUnknownFields`. New API endpoints get the check automatically while existing ones default to it off.

* Add logging probes to existing endpoints that fire when they encounter an unknown parameter. Search your logs for these later to see what unknown parameters are present, if any, and how many.

``` go
if err := decoder.Decode(v); err != nil {
    if strings.Contains(err.Error(), "unknown field") {
        plog.Logger(ctx).WithFields(logrus.Fields{
            "api_endpoint_method": r.Method,
            "api_endpoint_path":   r.URL.Path,
        }).Warnf("Unknown field error: %s.", err)

        decoderAllowingUnknown := json.NewDecoder(bytes.NewReader(rawPayload))
        err = decoderAllowingUnknown.Decode(v)
    }

    if err != nil {
        apierror.NewBadRequestError(
            r.Context(),
            fmt.Sprintf("Invalid JSON in request body: %s.", err),
        ).Write(r.Context(), w)
        return nil, false
    }
}
```

* Reaching out to individual users and asking them to correct bad parameters is possible, but probably more trouble than it's worth. A cheaper solution is to grandfather in existing errors by adding hidden fields to JSON structs that'll let `DisallowUnknownFields` be enabled for the endpoint, but keep existing integrations compatible.

``` go
// Request parameters for creating a new access token.
type AccessTokenCreateRequest struct {
    ...

    // When activating strict JSON parameter validation we found that Customer X
    // was accidentally sending `expires` instead of `expires_in`. We've asked
    // them to stop, but in the meantime we allow this parameter so we don't
    // break them.
    Expires int `json:"expires" openapi:"hide" validate:"-"`
}
```

There's a point where doing this for too many unknown fields becomes impractical, but for all but the largest APIs, unknown fields will be an edge that with a little luck, isn't that common.

### Deprecating fields carefully (#deprecating-fields)

When removing an old field from the API it might be tempting to strip it out request structs completely. It just makes sense right? If it's ignored anyway and not used anywhere then why should it be in there.

`DisallowUnknownFields` will require more care in deprecating fields. Even if the parameter hasn't been doing anything useful in years, it may still be sent by users, and if it's removed, those existing integrations break.

The workaround is to keep deprecated parameters passed their expiration date, but mark them as such in a way that bubbles up to public documentation and generated bindings that makes it clear that they're not useful and should no longer be used.

``` go
// Request parameters for creating a new access token.
type AccessTokenCreateRequest struct {
    ...

    // Client ID is the unique identifier of the API key that the new access
    // token should be associated with.
    //
    // Deprecated: This field used to be required, but an associated access
    // token is now inferred automatically using the secret included as part of
    // the `Authorization` header. This parameter is now ignored.
    ClientID *eid.EID `json:"client_id" validate:"-"`
}
```

Once again, logging probes come in handy here. Add a unique string like `access_token_client_id_received` that's easily searchable in logs, and some time later once it hasn't been seen in a long time, do a clean up pass and strip the old parameter out.

### Prepare an escape hatch (#escape-hatch)

Use of `DisallowUnknownFields` is suitable for most API endpoints, but an escape hatch _will_ be required, so prepare for it.

A common place where `DisallowUnknownFields` should not be applied are webhook receive endpoints. Although in a fashion they're technically part of your API's surface area, they're really more like the _push_ API of another vendor, and because adding a new field to an API is widely considered to not be a breaking change, that vendor may add new parameters to their webhook pushes anytime.

The problem can be especially insidious because the webhook APIs of many large vendors are quite stable, so your receiver will be working fine with `DisallowUnknownFields` for many months or years, before suddenly every request starts failing overnight as a new parameter is added.

Our in house API endpoint framework takes the option `AllowUnknownJSONFields` to indicate that JSON requests should not ban unknown fields:

``` go
// Webhook endpoint where Stripe broadcasts asynchronous message about customer
// payment information.
type StripeWebhookEndpoint struct{}

func (e *StripeWebhookEndpoint) Materialize() apiendpoint.APIEndpointer {
    return &apiendpoint.APIEndpoint[StripeWebhookRequest, StripeWebhookResponse]{
        Extras: apiendpoint.APIEndpointExtras{
            AllowUnknownJSONFields: true, // <-- unknown fields allowed
        },
        Method: http.MethodPost,
        Route:  "/webhook",
        ServiceHandler: func(svc any) func(ctx context.Context, req *StripeWebhookRequest) (*StripeWebhookResponse, error) {
            return svc.(StripeService).Webhook
        },
        SuccessStatusCode: http.StatusOK,
        Title:             "Stripe webhook receiver",
    }
}
```

## Use outside Go (#outside-go)

`DisallowUnknownFields` is obviously an option specific to Go, but this pattern is widely reusable in other languages, and easy to implement yourself if it's not built into the ecosystem's dominant JSON package.

## Augmentation with Levenshtein distance (#levenshtein)

An obvious next augmentation is not only to indicate that a parameter name doesn't exist, but to use the [Levenshtein distance
](https://en.wikipedia.org/wiki/Levenshtein_distance) to known parameter names to suggest one. So a user who sends `expires` is told that they probably meant `expires_in`, giving them a path to resolution that takes seconds instead of minutes.

``` sh
Invalid JSON in request body: unknown field "expires". Did you mean "expires_in"?"
```