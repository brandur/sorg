+++
hook = "A safe, succinct data fixtures pattern using sqlc and validator."
# image = ""
published_at = 2025-03-20T08:56:52-07:00
title = "The right way to do test fixtures in Go"
+++

Every test suite should start early in building a strong convention to generate data fixtures. If it doesn't, data fixtures will still emerge (they're that necessary), but in a way that's poorly designed, with no API (or a poorly designed one), and not standardized.

Other languages tend to have common libraries for fixture generation. As if often does, Go goes its own way and doesn't have a ubiquitous fixtures package, but especially when combining sqlc and [validator](https://github.com/go-playground/validator), it does well without one.

Here's one of our project's 130 fixtures:

``` go
package dbfactory

type MultiFactorOpts struct {
    ID          *uuid.UUID              `validate:"-"`
    AccountID   uuid.UUID               `validate:"required"`
    ActivatedAt *time.Time              `validate:"-"`
    ExpiresAt   *time.Time              `validate:"-"`
    Kind        *dbsqlc.MultiFactorKind `validate:"-"`
}

func MultiFactor(ctx context.Context, t *testing.T, e db.Executor, opts *MultiFactorOpts) *dbsqlc.MultiFactor {
    t.Helper()

    validateOpts(t, opts)

    var (
        num          = nextNumSeq()
        numFormatted = formatNumSeq(num)
    )

    multiFactor, err := dbsqlc.New().MultiFactorInsert(ctx, e, dbsqlc.MultiFactorInsertParams{
        ID:          ptrutil.ValOrDefaultFunc(opts.ID, func() uuid.UUID { return ptesting.ULID(ctx).New() }),
        AccountID:   opts.AccountID,
        ActivatedAt: ptrutil.TimeSQLNull(opts.ActivatedAt),
        ExpiresAt:   ptrutil.TimeSQLNull(opts.ExpiresAt),
        Kind:        string(ptrutil.ValOrDefault(opts.Kind, dbsqlc.MultiFactorKindTOTP)),
        Name:        fmt.Sprintf("%s no. %s", ptrutil.ValOrDefault(opts.Kind, dbsqlc.MultiFactorKindTOTP), numFormatted),
    })
    require.NoError(t, err)

    return multiFactor
}
```

The minimum viable use of the fixture needs only `AccountID`:

``` go
mf := dbfactory.MultiFactor(ctx, t, tx, &dbfactory.MultiFactorOpts{
    AccountID: account.ID,
})
```

But all salient properties are settable, so a more elaborate use just involves sending more overrides:

``` go
expiredMF := dbfactory.MultiFactor(ctx, t, bundle.tx, &dbfactory.MultiFactorOpts{
    AccountID: account.ID,
    ExpiresAt: ptrutil.Ptr(time.Now().Add(-5 * time.Minute)),
    Kind:      ptrutil.Ptr(dbsqlc.MultiFactorKindWebAuthn),
})
```

## Observations (#observations)

A few aspects worth calling out:

* Under the principle of not mocking the database, fixtures are real live data records. They're queryable using the full expressiveness of SQL, are valid according to the schema's data types/checks/triggers, and satisfy foreign keys.

* Fixtures never return an error, instead failing their input `t` so that generating a fixture is a one liner for the caller and doesn't need an `if err != nil { ... }` check.

* Inputs are annotated with [the Go validate framework](https://github.com/go-playground/validator) to demarcate required versus non-required or more complex validations as needed. This is a godsend because it keeps validations short (zero additional lines instead of a minimum of three for an `if` statement) and fast/easy to write.

* As few properties are made `validate:"required"` as possible, with non nullable fields given defaults instead of marked mandatory for the caller to fill. This makes fixtures easier to use and reduces boilerplate at call sites. e.g. `name` is a required property on `multi_factor` above, but the fixture generates a sane default.

* Insert statements are generated with [sqlc](/sqlc).

``` sql
-- name: MultiFactorInsert :one
INSERT INTO multi_factor (
    id,
    account_id,
    activated_at,
    expires_at,
    kind,
    name
) VALUES (
    @id,
    @account_id,
    @activated_at,
    @expires_at,
    @kind,
    @name
) RETURNING *;
```

* We use of a lot of custom pointer helpers like `ptrutil.TimeSQLNull` (changes a pointer to a `sql.NullTime`) and `ptrutil.ValOrDefault`. Each one of these changes a ~4 line local variable declaration and `if` block to one LOC that it's inlined into the insert. True Go dogmatists won't like this, but it saves dozens of lines per test fixture, and given hundreds of test fixtures, this adds up to thousands of lines saved overall.

* Each test case gets its own lazily marshaled monotonic ULID generated based on `t`. Separate generators guarantee monotonicity even if some test cases rewind their generators to generate ULIDs at particular times.

## Organizing with var blocks (#var-blocks)

Typically, fixtures are generated together in a `var ( ... )` block, keeping tests looking nice and tidy:

``` go
t.Run("SetNameSSOJoinSCIMError", func(t *testing.T) {
    t.Parallel()

    bundle, ctx := setup(t)

    var (
        org  = dbfactory.Organization(ctx, t, bundle.tx, &dbfactory.OrganizationOpts{SCIMEnabled: true})
        team = dbfactory.Team(ctx, t, bundle.tx, &dbfactory.TeamOpts{OrganizationID: &org.ID})
        _    = dbfactory.AccessGroupAccount_Admin(ctx, t, bundle.tx, team.ID, bundle.account.ID)
    )

    _, err := pservicetest.InvokeHandler(bundle.svc.Update, ctx, &TeamUpdateRequest{
        Name:   ptrutil.Ptr("new name"),
        TeamID: eid.EID(team.ID),
    })
    prequire.APIErrorWithMessage(t, &apierror.BadRequestError{}, fmt.Sprintf(errMessageTeamUpdateSCIM, "name"), err)
})
```

## Standardize conventions, even the small ones (#standardize-conventions)

We have a few helpers that are used in almost every test fixture. These are so trivial that they almost don't need to be extracted into their own functions, but we've done so to prevent implementations from drifting and keep code maximally succinct.

``` go
// Formats a number like "000007". Typically used in conjunction
// with nextNumSeq to make identifiers prettier and so they align
// better.
func formatNumSeq(num int64) string {
    return fmt.Sprintf("%06d", num)
}

var numSeq int64

// Gets a unique number that can be used in names, etc. and which
// is more friendly to look at than a UUID.
func nextNumSeq() int64 {
    return atomic.AddInt64(&numSeq, 1)
}

func validateOpts(t *testing.T, opts any) {
    t.Helper()

    err := validate.Struct(opts)
    require.NoError(t, err)
}
```