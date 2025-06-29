+++
hook = "Using a coalescable parameter to stub time as necessary in tests, but otherwise use the shared database clock across all operations."
# image = ""
published_at = 2025-06-29T10:48:16-07:00
title = "Occasionally injected clocks in Postgres"
+++

In a standard app deployment that's scaled horizontally across many nodes, we can expect the clocks to be a little askew across the fleet. It's generally not a huge problem these days because our [use of NTP](https://en.wikipedia.org/wiki/Network_Time_Protocol) is so good and so widespread, but minor drift is still present.

Where a single source of time authority is desired, a nice trick is to use the database. A single database is shared across all deployed nodes, so by using the database's `now()` function instead of `time.Now()` in code, we can expect perfect consistency across all created records.

But a downside of this approach is that it makes time hard to stub because Postgres' time is hard to stub. Stubbing time is often a necessity in tests and not being able to do so is a deal breaker.

We've been using a hybrid approach with some success. A call to `coalesce` prefers an injected timestamp if there is one, but falls back on `now()` most of the time (including in production) to share a clock.

## Step 1: SQL + sqlc (#sql-sqlc)

Here's a sample query showing the `coalesce` in action. `sqlc.narg` defines a parameter as nullable.

``` sql
-- name: QueuePause :execrows
UPDATE queue
SET paused_at = CASE
                WHEN paused_at IS NULL THEN coalesce(
                    sqlc.narg('now')::timestamptz,
                    now()
                )
                ELSE paused_at
                END
WHERE name = @name;
```

In `sqlc.yaml`, tell sqlc to emit nullable timestamps as `*time.Time` pointers:

``` yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: ...
    schema: ...
    gen:
      go:
        overrides:
          - db_type: "timestamptz"
            go_type:
              type: "time.Time"
              pointer: true
            nullable: true
```

Which generates this code:

``` go
const queuePause = `-- name: QueuePause :execrows
UPDATE queue
SET
    paused_at = CASE WHEN paused_at IS NULL THEN coalesce($1::timestamptz, now()) ELSE paused_at END
WHERE CASE WHEN $2::text = '*' THEN true ELSE name = $2 END
`

type QueuePauseParams struct {
    Now  *time.Time
    Name string
}

func (q *Queries) QueuePause(ctx context.Context, db DBTX, arg *QueuePauseParams) (int64, error) {
    result, err := db.Exec(ctx, queuePause, arg.Now, arg.Name)
    if err != nil {
        return 0, err
    }
    return result.RowsAffected(), nil
}
```

## Step 2: Stubabble time generator (#stubbable-time-generator)

Working in Go, define a `TimeGenerator` interface:

* When unstubbed, it returns the current time from `NowUTC()` or `nil` from `NowUTCOrNil()`.
* When stubbed, it returns the stubbed time from `NowUTC()` or a pointer version of the same from `NowUTCOrNil()`.

``` go
// TimeGenerator generates a current time in UTC. In test
// environments it's implemented by TimeStub which lets the
// current time be stubbed. Otherwise, it's implemented as
// UnstubbableTimeGenerator which doesn't allow stubbing.
type TimeGenerator interface {
    // NowUTC returns the current time. This may be a stubbed
    // time if the time has been actively stubbed in a test.
    NowUTC() time.Time

    // NowUTCOrNil returns if the currently stubbed time _if_
    // the current time is stubbed, and returns nil otherwise.
    // This is generally useful in cases where a component may
    // want to use a stubbed time if the time is stubbed, but
    // to fall back to a database time default otherwise.
    NowUTCOrNil() *time.Time
}
```

A stubbable implementation for tests:

``` go
type TimeStub struct {
    nowUTC *time.Time
}

func (t *TimeStub) NowUTC() time.Time {
    if t.nowUTC == nil {
        return time.Now().UTC()
    }

    return *t.nowUTC
}

func (t *TimeStub) NowUTCOrNil() *time.Time {
    return t.nowUTC
}

func (t *TimeStub) StubNowUTC(nowUTC time.Time) time.Time {
    t.nowUTC = &nowUTC
    return nowUTC
}
```

An unstubbable time generator for production:

``` go
type UnstubbableTimeGenerator struct{}

func (g *UnstubbableTimeGenerator) NowUTC() time.Time       { return time.Now() }
func (g *UnstubbableTimeGenerator) NowUTCOrNil() *time.Time { return nil }

func (g *UnstubbableTimeGenerator) StubNowUTC(nowUTC time.Time) time.Time {
    panic("time not stubbable outside tests")
}
```

### Step 3: Distributing a shared time generator (#shared-time-generator)

The next key aspect is that all code needs to share a single instance of `TimeGenerator` so that when it's stubbed from a test, all services and subservices get the same stubbed value.

We put a `TimeGenerator` on a base service archetype that's automatically injected from top-level services to subservices:

``` go
func (c *Client[TTx]) QueuePauseTx(ctx context.Context, tx TTx, name string, opts *QueuePauseOpts) error {
    executorTx := c.driver.UnwrapExecutor(tx)

    if err := executorTx.QueuePause(ctx, &QueuePauseParams{
        Name:   name,
        Now:    c.baseService.Time.NowUTCOrNil(), // <-- accessed here
        Schema: c.config.Schema,
    }); err != nil {
        return err
    }
```

By default, it's instantiated as `UnstubbableTimeGenerator`. From tests, it's a `TimeStub`:

``` go
func BaseServiceArchetype(tb testing.TB) *baseservice.Archetype {
    tb.Helper()

    return &baseservice.Archetype{
        Logger: Logger(tb),
        Time:   &TimeStub{},
    }
}
```

In a test, time is stubbed like:

``` go
stubbedNow := client.baseService.Time.StubNowUTC(time.Now().UTC())
```

## Loose conviction (#loose-conviction)

Consider this one a loose recommendation. It's useful in some situations where timestamp consistency is critically important, but not in others where it isn't. Server clocks tend to be pretty good nowadays, and it's a lot of code to avoid a few tens of microseconds worth of drift.

Also, consider that there might be a downside to using the database clock. In SQL, `CURRENT_TIMESTAMP` and `now()` in Postgres represent the current time _at the start of the current transaction_ rather than the current time. This might be a benefit as all records created during a transaction are assigned the same created time, but it's just as often undesirable because depending on the duration of the transaction, timestamps can be wildly unrepresentative of when things actually happened.