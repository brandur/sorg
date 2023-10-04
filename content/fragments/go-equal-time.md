+++
hook = "A time assertion for Go that ignores `time.Time`'s monotonic component and stops at microsecond-level precision."
published_at = 2023-10-04T11:56:57+02:00
title = "A Postgres-friendly time comparison assertion for Go"
+++

Three major problems with trying to compare time values in Go:

* Times in Go carry a monotonic component, so attempting to compare them directly will more often than not produce the wrong result (e.g. using testify's `require.Equal` helper).

* Times in Go are precise down to the nanosecond, whereas Postgres stores times to microsecond-level precision, so comparing a time to one round-tripped from Postgres will fail (say with the use of the built-in `time.Equal`).

* Times in Go are structs, so comparison failures that print a difference will produce an eyesore wall of text that's totally incomprehensible to an average person not already deeply familiar with `time` package internals.

My workaround was usually to pre-format times to RFC3339, which was verbose and ugly:

``` go
require.Equal(t,
    endBillingHour.Format(time.RFC3339),
    updatedEntitlement1.LastUsageSubmissionAt.Time.Format(time.RFC3339),
)
```

## The `EqualTime` assertion (#equal-time)

I finally got around to adding a helper to make time comparisons easier and more reliable. It's not much code:

``` go
// EqualTime compares two times in a way that's safer and with better fail
// output than a call to `require.Equal` would produce.
//
// It takes care to:
//
//   - Strip off monotonic portions of timestamps so they aren't considered for
//     purposes of comparison.
//
//   - Truncate nanoseconds in a functionally equivalent way to how pgx would do
//     it so that times that have round-tripped from Postgres can still be
//     compared. Postgres only stores times to the microsecond level.
//
//   - Use formatted, human-friendly time outputs so that in case of a failure,
//     the discrepancy is easier to pick out.
func EqualTime(t testing.TB, t1, t2 time.Time) {
	// Note that leaving off the nanosecond portion will have the effect of
	// truncating it rather than rounding to the nearest microsecond, which
	// functionally matches pgx's behavior while persisting.
	const rfc3339Micro = "2006-01-02T15:04:05.999999Z07:00"

	require.Equal(t,
		t1.Format(rfc3339Micro),
		t2.Format(rfc3339Micro),
	)
}
```

Notably, although nanoseconds are truncated, we still compare all the way to the microsecond level. This helps to root out tests that are accidentally not using a stable clock value, and would otherwise occasionally produce an off-by-one second assertion failure. It's rare to accidentally get the same value to the microsecond with two calls to `time.Now()`, so those tests will tend to fail every time instead of intermittently.

## Appendix: Test suite (#test-suite)

And a test case to show that it works and exercise edges:

``` go
func TestEqualTime(t *testing.T) {
	t.Parallel()

	t1 := time.Now()

	// Strip off milli/micro/nanosecond portion and add back our own test
	// version. This guarantees that adding some nanoseconds (which is done
	// below) won't ever accidentally roll time over to the next microsecond and
	// also helps adding some better stability to the tests.
	//
	// We don't use Round/Truncate because we want to keep the monotonic portion
	// of the time to show that it's ignored during comparison (Round/Truncate
	// would otherwise strip it off).
	t1 = t1.Add(-time.Duration(t1.Nanosecond()))
	t1 = t1.Add(123_456_789)

	// Exactly equivalent time, but Truncate has the effect of stripping the
	// monotonic component off. This lets us verify that it's not considered for
	// purposes of comparison.
	t2 := t1.Truncate(time.Nanosecond)

	// Log times to help visual what's happening. Using `-test.v` you can see
	// the monotonic portion of t1 remains, but has been stripped from t2.
	t.Logf("t1 = %+v", t1)
	t.Logf("t2 = %+v", t2)
	t.Logf("")

	// These will look identical.
	t.Logf("As time.RFC3339Nano:")
	t.Logf("t1 = %+v", t1.Format(time.RFC3339Nano))
	t.Logf("t2 = %+v", t2.Format(time.RFC3339Nano))

	prequire.EqualTime(t, t1, t2)

	// Comparison only happens to the microsecond level to better allow for
	// round trips back from Postgres.
	prequire.EqualTime(t, t1, t2.Truncate(time.Microsecond))

	// Basically the same thing again. Comparison passes even with some extra
	// nanosecond padding.
	prequire.EqualTime(t, t1, t2.Add(100*time.Nanosecond))

	// Fail: difference at second level.
	expectFailure(t, func(t testutil.TestingT) {
		prequire.EqualTime(t, t1, t2.Add(1*time.Second))
	})

	// Fail: difference at millisecond level.
	expectFailure(t, func(t testutil.TestingT) {
		prequire.EqualTime(t, t1, t2.Truncate(time.Millisecond))
	})

	// Fail: difference at microsecond level.
	expectFailure(t, func(t testutil.TestingT) {
		prequire.EqualTime(t, t1, t2.Add(1*time.Microsecond))
	})
}
```