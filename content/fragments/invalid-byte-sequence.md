+++
hook = "Handling a common programming language/database asymmetry around tolerance of zero bytes."
# image = ""
published_at = 2024-12-19T14:58:05-07:00
title = "ERROR: invalid byte sequence for encoding UTF8: 0x00 (and what to do about it)"
+++

One of the oldest errors I ever remember seeing in an error tracker:

> ERROR: invalid byte sequence for encoding "UTF8": `0x00`

Through my time at Heroku it was like a distant friend. Not one that you'd see every day, but one who'd appear to be surprise you a few dozen times a year. Since it didn't seem to be causing any major fallout and I never heard a user complain about it, I'm somewhat embarrassed to say that in four years neither myself nor anyone else ever bothered to look into it.

These days, on a Go stack and with much better control and insight into any changes we make, we're pretty aggressive about trying to prune Sentry errors down to zero. Over a few months I'd see the `0x00` error come and go, and finally decided to look into it.

The problem comes from Postgres raising an error when a caller tries to insert a text/varchar value containing a value of `0x00`, or zero byte. The same value that's used to terminate a string in plain old C. Postgres [explicitly disallows it](https://www.postgresql.org/docs/current/sql-syntax-lexical.html#SQL-SYNTAX-STRINGS-ESCAPE):

> The character with the code zero cannot be in a string constant.

The tricky part is that although Postgres won't take a zero byte, almost every programming language ever created _will_, thereby creating a natural asymmetry between database and language stack.

As far as I know, there aren't any legitimate uses for sending a zero byte to an API or web app. Looking back through our logs, the main places I've seen it are from bots out on the internet, presumably using common attack patterns to probe for weaknesses, or from pentest teams that we paid to do the same.

## Validating at the edges (#edges)

We're using the [validate framework for Go](https://github.com/go-playground/validator) to check that API inputs are sound, like that they're present, below a max length, or within bounds. In a language known for its verbosity, validate annotations are succinct and quick to write.

The custom validations `apistring200`, `apistrong2000`, `apistring20000`, etc. are assigned to API string parameters in [order of magnitude tiers](/text#varchars). Their implementation denies `\x00`s that come in with request payloads:

``` go
// API strings are meant to provide a reasonable default validation
// for strings that come in via the API that aren't already
// validated more strictly. The main idea is to make sure that
// we're not getting long, unbounded input that'll either store a
// very invalid value to the database or be rejected by a DB-level
// constraint (which would bubble up as a 500 with little context).
//
// They also validate that strings contain no invalid unicode
// sequences, and that no `\x00` zero bytes are present, both of
// which Postgres will reject.
must(registerAPIString("apistring200", 200))
must(registerAPIString("apistring2000", 2_000))
must(registerAPIString("apistring20000", 20_000))
must(registerAPIString("apistring200000", 200_000))

const (
    apiStringErrorMessage = "`{0}` should be a non-empty string with a maximum length of %d characters, and contain no invalid unicode sequences or zero bytes"
)

func registerAPIString(tag string, maxLength int) error {
    if err := validate.RegisterValidation(tag, func(fl validator.FieldLevel) bool {
        val := fl.Field().String()

        if len(val) == 0 || len(val) > maxLength {
            return false
        }

        if !utf8.ValidString(val) {
            return false
        }

        // A zero (0x00) rune is valid UTF-8 and won't be caught
        // by the unicode check above, but Postgres will refuse
        // to insert it.
        if strings.Contains(val, "\x00") {
            return false
        }

        return true
    }); err != nil {
        return err
    }

    return registerTranslation(tag, fmt.Sprintf(apiStringErrorMessage, maxLength))
}
```

Notably, it also denies invalid UTF-8 byte sequences (`\x00` is not desirable, but it is valid UTF-8), another common malformed input that internet bots like to send, and which will cause its own Postgres error.

Struct fields are tagged with validations, making use easy and concise:

``` go
// Request for creating a new account.
type AccountCreateRequest struct {
    // Full name for the new account.
    Name *string `json:"name" validate:"apistring200"`
    
    ...
```

## Storing raw request properties (#raw-request-properties)

That takes care of input forms, but another place we'd see the problem is when trying to insert [canonical API lines](/canonical-log-lines) to the database for operational visibility. Even where we denied a request with invalid input with a 400, we record a canonical line for it, invalid input and all.

For this case, we take anything invalid in the input and replace it with a placeholder token that's safely storable to Postgres:

``` go
// TrimInvalidUTF8 replaces any invalid UTF-8 or \x00 bytes with
// symbolic stand-in tokens. This lets strings that contain invalid
// UTF-8 be stored to Postgres, which normally won't tolerate
// invalid UTF-8 in string-like fields.
func TrimInvalidUTF8(s string) string {
    if !utf8.ValidString(s) {
        s = strings.ToValidUTF8(s, "[invalid UTF-8]")
    }

    // A zero (0x00) rune is valid UTF-8 and won't be caught by the
    // check above, but Postgres will refuse to insert it. Replace
    // all instances with a marker that Postgres can tolerate and
    // which is indicative of what happened. This should only ever
    // happen because of random probing from malicious internet
    // actors sending garbage into HTTP paths and what not.
    if strings.Contains(s, "\x00") {
        s = strings.ReplaceAll(s, "\x00", "[0x00 UTF-8 rune]")
    }

    return s
}
```

This is combined with another helper to that samples inputs longer than we're willing to store:

``` go
// Returns a string that's been truncated the given max length and
// stripped of any invalid UTF-8 that Postgres might balk at.
// Returns an empty string on `nil` for purposes of the batch
// insert will treat empty strings as NULL.
validTruncatedStringOrEmpty := func(sPtr *string, maxLength int) string {
    if sPtr == nil {
        return ""
    }

    return stringutil.SampleLongN(stringutil.TrimInvalidUTF8(*sPtr), maxLength)
}
```

When inserting a canonical line for a request, inputs are sanitized and truncated. This happens for obvious fields where an invalid input can be sent like a query string or form body, but for less obvious ones as well. Invalid input can come in almost anywhere, including headers like `Content-Type` or `User-Agent`:

``` go
insertParams.ContentType[i] =
    validTruncatedStringOrEmpty(logData.ContentType, 200)
insertParams.HTTPPath[i] =
    validTruncatedStringOrEmpty(&logData.HTTPPath, 200)
insertParams.QueryString[i] =
    validTruncatedStringOrEmpty(logData.QueryString, 2000)
insertParams.UserAgent[i] =
    validTruncatedStringOrEmpty(logData.UserAgent, 200)
```

## 0x01 down (#one-down)

This is one of those little housekeeping tasks that may not be that important, but is quite gratifying. With the steps above we've eradicated "invalid byte sequence" errors, taking us a step closer to our target steady state of zero Sentry issues.
