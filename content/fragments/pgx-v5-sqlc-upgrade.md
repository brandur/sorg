+++
hook = "Upgrading a Go project from pgx v4 to v5, and more specifically, from sqlc's `pgx/v4` to `pgx/v5` driver (the hard part)."
published_at = 2023-05-20T18:50:34+01:00
title = "PGX + sqlc v4 to v5 upgrade notes"
+++

Last week, we upgraded from pgx v4 to v5. PGX v5 brings in some [nice incremental improvements](https://github.com/jackc/pgx/blob/master/CHANGELOG.md#v500-september-17-2022), with the major one being that packages have been reorganized so that `pgtype`, `pgconn`, and `pgproto3` all move into the `pgx` Go module instead of living on their own, and that has the major advantage of making the versioning between them much less confusing.

Upgrading pgx was the easy part. The hard part was upgrading [sqlc](/sqlc) to use the new pgx v5 driver, which had been available for a few months, but was only considered stable in the latest sqlc 1.18 release.

Notes from the upgrade:

* For nullable fields, pgx types like `pgtype.Text` are now preferred over `sql` types like `sql.NullString`.
* Some types like `jsonb` got easier to work with by becoming `[]byte` instead of `pgtype.JSONB`. `pgtype.JSONB` was particularly annoying because its zero value was `Undefined`, meaning that for any query that contained one you'd have to make sure to explicitly initialize a `pgtype.JSONB` regardless of whether it was used or not because pgx would refuse to write an `Undefined` value.
* In general, PGX v5 did away with its "tri-state" `Status` system of `Undefined`/`Present`/`Null` in favor of an `sql`-like `Valid` boolean, so types like `[]byte` that are already nullable now fit better in the paradigm.
* `cidr` and `inet` become [`netip.Prefix`](https://pkg.go.dev/net/netip#Prefix). A little annoying to have to make the change, but unlike `net.IP`, `netip` structs are immutable and support `==` comparisons.
* Range types use the new `pgtype.Range[]` generic type instead of specific ones like `pgtype.Tstzrange` (and also drop the tri-state `Status` field in favor of the simpler `Valid` boolean). This is a great use of Go's generics and makes everything easier and more predictable.

## Our custom configuration (#custom-configuration)

We were making such heavy use of `sql.Null*` types that in order to not produce a patch with a million-line diff, I manually mapped back to `sql.Null*` in sqlc's configuration (for the nullable variants only -- sqlc does the right thing and maps to Go built-ins for non-nullable fields):

``` yaml
overrides:
  - db_type: "pg_catalog.bool"
    go_type: "database/sql.NullBool"
    nullable: true
  - db_type: "pg_catalog.float8"
    go_type: "database/sql.NullFloat64"
    nullable: true
  - db_type: "pg_catalog.int4"
    go_type: "database/sql.NullInt32"
    nullable: true
  - db_type: "pg_catalog.int8"
    go_type: "database/sql.NullInt64"
    nullable: true
  - db_type: "pg_catalog.varchar"
    go_type: "database/sql.NullString"
    nullable: true
  - db_type: "text"
    go_type: "database/sql.NullString"
    nullable: true
```

The v5 driver changes `uuid`s to use `pgtype.UUID` which wasn't interoperable with anything else we already had, so I mapped those back to Google's package:

``` yaml
overrides:
  - db_type: "uuid"
    go_type: "github.com/google/uuid.UUID"
  - db_type: "uuid"
    go_type: "github.com/google/uuid.NullUUID"
    nullable: true
```

I found `pgtype.Interval` really hard to work with because it doesn't give you any way to convert to and from `time.Duration`, so we mapped to `time.Duration` for more ergonomic use:

``` yaml
overrides:
  - db_type: "pg_catalog.interval"
    go_type: "time.Duration"
  - db_type: "pg_catalog.interval"
    # It seems like this could be the simpler `go_type: "*time.Duration"`, but
    # that outputs double points like `**time.Duration` for reasons that are
    # beyond me (bug?). The expanded version of `go_type` usage below works.
    go_type:
      import: "time"
      type: "Duration"
      pointer: true
    nullable: true
```

The other one that's awful to work with is `pgtype.Numeric` because there aren't any built-ins for interchanging it with more common Go number types like `float64`/`int64`/`uint64`. I didn't find any clean workarounds, so I implemented utility functions for conversions (although am on the look out for something better).

## If your program hangs on exit ... (#hang-on-exit)

The hardest bug I had to resolve during the upgrade was related [the `WaitForNotification` helper](https://pkg.go.dev/github.com/jackc/pgx/v5#hdr-Listen_and_Notify) for receiving a `pg_notify` message. Previously, if its connection was released, it'd send an error back to the caller that'd force it to die as well. That doesn't happen anymore, and my program would hang on shutdown because I'd deferred a `Close()` on the connection pool, which won't return until all resources are released by the pool, and resources would never be released because `WaitForNotification` would never be interrupted, causing my program to hang on exit (luckily the problem was discovered in CI thanks to our [program start check](/fragments/program-start-check)).

The fix was to make sure to use a cancellable context which is cancelled on shutdown:

``` go
ctx, cancel := context.WithCancel(ctx)

// Waiting on a notification is only cancellable by canceling its context,
// so here, cancel on shutdown so that we stop listening and exit cleanly.
go func() {
    <-shutdown
    cancel()
}()

if err := startListening(ctx, listenConn, waitGroup); err != nil {
    // Manifests from cancellation as the worker is shutting down.
    if errors.Is(err, context.Canceled) {
        return
    }

    ...
```

The total diff was 114 files changed and ~800 lines of code over a focused session of ~6 hours. It was somewhat painful, but now that it's over, I'm glad to have tackled it because usage of sqlc was only moving in one direction, so the project was only getting harder. Another few months and it might not have been possible to do in a single change, and figuring out how to have multiple pgx versions coexist while upgrading incrementally would've added a whole novel layer to the ordeal.
