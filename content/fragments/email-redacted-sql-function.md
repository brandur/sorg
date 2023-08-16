+++
hook = "Neither here nor there, but a PL/pgSQL function for redacting emails in Postgres to derisk leaking PII."
published_at = 2016-08-24T13:36:55Z
title = "SQL email redaction function for Postgres"
+++

We recently shipped a [saved queries](https://docs.crunchybridge.com/concepts/saved-queries) feature that shows the results of a SQL query in browser and which can produce shareable links.

When sharing data, it's crucial that PII (personally identifiable information) be redacted. PII comes in many forms, with one possible kind being email. Emails should never be shown anywhere wholesale, but even a partial email can add a lot of context, like showing the domain of an account that owns a resource.

I wrote an `email_redacted` function:

``` sql
CREATE OR REPLACE FUNCTION email_redacted(email text) RETURNS text
    LANGUAGE plpgsql
    IMMUTABLE
    PARALLEL SAFE
    AS $_$
DECLARE
    parts                text[] = string_to_array(email, '@');
    domain               text;
    parts_without_domain text[];
    name_parts           text[];
    i                    int;
BEGIN
    IF array_length(parts, 1) IS NULL OR array_length(parts, 1) < 2 THEN
        RETURN email;
    END IF;

    domain               = parts[array_upper(parts, 1)];
    parts_without_domain = trim_array(parts, 1);
    name_parts           = string_to_array(array_to_string(parts_without_domain, '@'), '.');

    IF array_upper(name_parts, 1) IS NULL THEN
        RETURN email;
    END IF;

    FOR i IN 1 .. array_upper(name_parts, 1) LOOP
        name_parts[i] = substr(name_parts[i], 1, 2) || '***';
    END LOOP;

    IF domain IS NULL THEN
        RETURN array_to_string(name_parts, '.');
    END IF;

    RETURN array_to_string(name_parts, '.') || '@' || domain;
END
$_$;
```

Sample run:

```
=# SELECT email_redacted('steve.zissou@crunchydata.com');
       email_redacted
-----------------------------
 st***.zi***@crunchydata.com
```

Some caveats:

* When sharing data publicly, this is still too much information about an account. The function's meant to add context and derisk sharing for internal use.

* My PL/pgSQL sucks, and I'm sure it could be written better. Most inputs work with the function dramatically shortened by removing all the `IF` checks, but those are needed to protect against degenerate inputs.

## Appendix: Test suite (#appendix-test-suite)

A Go test case to check inputs (uses many internal abstractions so only meant to demontrate the coarse shape of such a thing):

``` go
func TestEmailRedacted(t *testing.T) {
	t.Parallel()

	ctx := ptesting.Context(t)

	invokeFuncNullable := func(dbtx DBTX, email *string) *string {
		t.Helper()
		row := dbtx.QueryRow(ctx, "SELECT email_redacted($1)", email)
		var s *string
		err := row.Scan(&s)
		require.NoError(t, err)
		return s
	}

	invokeFunc := func(dbtx DBTX, email string) *string {
		return invokeFuncNullable(dbtx, &email)
	}

	tx := ptesting.TestTx(ctx, t)

	require.Equal(t, "fo***@example.com", *invokeFunc(tx, "foo@example.com"))
	require.Equal(t, "fo***.ba***@example.com", *invokeFunc(tx, "foo.bar@example.com"))
	require.Equal(t, "fo***@example.com", *invokeFunc(tx, "foo@bar@example.com"))
	require.Equal(t, "fo***.ba***.ba***@example.com", *invokeFunc(tx, "foo.bar.baz@example.com"))
	require.Equal(t, "fo***@example.com", *invokeFunc(tx, "foo@bar@baz@example.com"))

	// Degenerate cases
	require.Equal(t, "@example.com", *invokeFunc(tx, "@example.com"))
	require.Equal(t, "example.com", *invokeFunc(tx, "example.com"))
	require.Nil(t, invokeFuncNullable(tx, nil))
}
```