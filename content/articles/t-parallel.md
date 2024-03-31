+++
hook = "How and why we started annotating all our Go tests with `t.Parallel()`, and why you might want to consider doing so too."
location = "San Francisco"
published_at = 2023-08-26T13:48:45-07:00
title = "On Using Go's `t.Parallel()`"
+++

One of Go's best features is not only that it does parallelism well, but that it's deeply baked in. It's best exemplified by primitives like goroutines and their dead simple ease of use, but extends all the way up the chain to the built-in tooling. When running tests for many packages with `go test ./...`, packages automatically run in parallel up to a maximum equal to the number of CPUs on the machine. Between that and the language's famously fast compilation, test suites are fast _by default_ instead of something that needs to be painstakingly optimized later on.

Within any specific package, tests run sequentially, and as long as packages aren't too mismatched in test suite size, that's generally good enough.

But having uniformly sized package test suites isn't always a given, and some packages can grow to be quite large. We have a `./server/api` package that contains the majority of our product's API and ~200 tests to exercise it, and it's measurably slower than most packages in the project.

For cases like this, Go has another useful parallel facility: [`t.Parallel()`](https://pkg.go.dev/testing#T.Parallel), which lets specific tests _within a package_ be flagged to run in parallel with each other. When applied to our large package, it reduced the time needed for a single run by 30-40% or by 2-3x for ten consecutive runs.

Before `t.Parallel()`:

``` sh
$ go test ./server/api -count=1
ok      github.com/crunchydata/priv-all-platform/server/api     1.486s
$ go test ./server/api -count=10
ok      github.com/crunchydata/priv-all-platform/server/api     11.786s
```

After `t.Parallel()`:

``` sh
$ go test ./server/api -count=1
ok      github.com/crunchydata/priv-all-platform/server/api     0.966s
$ go test ./server/api -count=10
ok      github.com/crunchydata/priv-all-platform/server/api     3.959s
```

These tests were already pretty fast (to beat a dead horse again: running _every_ API test for this project is 3-5x+ faster than it took to run _a single test case_ during my time at Stripe; language choice and infrastructure design makes a big difference), but this is one of the packages that we run tests on most frequently, so a 30-40% speed up makes a noticeable difference in DX when iterating.

After adding `t.Parallel()` to this one package, we then went through and added it to every test in every package, and then put in a ratchet with [the `paralleltest` linter](https://golangci-lint.run/usage/linters/#paralleltest) to mandate it for future additions.

Should you bother adding `t.Parallel()` like we did? Maybe. It's a pretty easy standard to adhere to when starting from scratch, and for existing ones it'll be easier to add it today than at any point later on, so it's worth considering.

## Is `t.Parallel()` broadly recommended practice?

As far as I can tell, no.

I like to use the Go language's own source code to glean convention, and by my rough measurement only about 1/10th of its test suite uses `t.Parallel()`:

``` sh
# total number of tests
$ ag --no-filename --nobreak 'func Test' | wc -l
    7786
    
# total number of uses of `t.Parallel()`
$ ag --no-filename --nobreak 't\.Parallel\(\)' | wc -l
     620
```

This isn't too surprising. As discussed above, parallelism across packages is usually good enough, and when iterating tests in one specific package, Go's already pretty fast. For smaller packages adding parallelism is probably a wash, and for very small ones the extra overhead probably makes them slower (although trivially so).

Still, it might not be a bad idea. As some packages grow to be large, parallel testing will keep them fast, and annotating tests with `t.Parallel()` from the beginning is a lot easier than going back to add it to every test case and fix parallelism problems later on.

## Sharp edges (#sharp-edges)

### Sharing a database with test transactions (#test-tx)

The biggest difficulty for many projects will be to have a strategy for the test database that can support parallelism. It's easy to build a system where multiple tests target the same test database and insert data that conflicts with each other.

We use [test transactions](/fragments/go-test-tx-using-t-cleanup) to avoid this. Each test opens a transaction, runs everything inside it, and rolls the transaction back as it finishes up. A simplified test helper looks like:

``` go
func TestTx(ctx context.Context, t *testing.T) pgx.Tx {
    tx, err := getPool().Begin(ctx)
    require.NoError(t, err)

    t.Cleanup(func() {
        err := tx.Rollback(ctx)
        if !errors.Is(err, pgx.ErrTxClosed) {
            require.NoError(t, err)
        }
    })

    return tx
}
```

Invocations of the helper share a package-level pgx pool that's automatically parallel-safe (but still has a mutex to make sure that only one test case initializes it):

``` go
var (
    dbPool   *pgxpool.Pool
    dbPoolMu sync.RWMutex
)
```

Usage is succinct and idiot-proof thanks to Go's test `Cleanup` hook:

``` go
tx := TestTx(ctx, t)
```

### Deadlocks across transactions (#deadlocks)

The trickiest problem I had to fix while enabling `t.Parallel()` involved Postgres upsert. We have a number of places where we seed data with an upsert to guarantee that it's always in the database regardless of whether the program has run before or is starting for the first time. In the test suite, individual test cases would upsert a "known" resource:

``` go
plan := dbfactory.Plan_AWS_Hobby2(ctx, t, tx)
```

Implemented as:

``` go
func Plan(ctx context.Context, t *testing.T, e db.Executor, opts *PlanOpts) *dbsqlc.Plan {
    validateOpts(t, opts)

    configPlan := providers.Default.MustGet(opts.ProviderID).MustGetPlan(opts.PlanID, true)

    plan, err := dbsqlc.New(e).PlanUpsert(ctx, dbsqlc.PlanUpsertParams{
        CPU:         int32(configPlan.CPU),
        Disabled:    configPlan.Disabled,
        DisplayName: configPlan.DisplayName,
        Instance:    configPlan.Instance,
        Memory:      configPlan.Memory,
        ProviderID:  opts.ProviderID,
        PlanID:      configPlan.ID,
        Rate:        int32(configPlan.Rate),
    })
    require.NoError(t, err)
    return &plan
}
```

To my surprise, adding `t.Parallel()` would fail many tests at these invocations. Despite every test case running in its own transaction, it's still possible for them to deadlock against other as they tried to upsert exactly the same data.

We resolved the problem by moving to a fixture seeding model, so when the test database is being created, in addition to loading a schema and running migrations, we also load a common set of test data in it that all tests will share (test transactions ensure that any changes to it are rolled back):

``` make
.PHONY: db/test
db/test:
    psql --echo-errors --quiet -c '\timing off' -c "DROP DATABASE IF EXISTS platform_main_test WITH (FORCE);"
    psql --echo-errors --quiet -c '\timing off' -c "CREATE DATABASE platform_main_test;"
    psql --echo-errors --quiet -c '\timing off' -f sql/main_schema.sql
    go run ./apps/pmigrate
    go run ./tools/src/seed-test-database/main.go
            
```

So the implementation becomes a lookup instead:

``` go
func Plan(ctx context.Context, t *testing.T, e db.Executor, opts *PlanOpts) *dbsqlc.Plan {
    validateOpts(t, opts)

    _ = providers.Default.MustGet(opts.ProviderID).MustGetPlan(opts.PlanID, true)

    // Requires test data is seeded.
    provider, err := dbsqlc.New(e).PlanGetByID(ctx, dbsqlc.PlanGetByIDParams{
        PlanID:     opts.PlanID,
        ProviderID: opts.ProviderID,
    })
    require.NoError(t, err)

    return &provider
}
```

### Logging and `t.Log` (#logging)

We make fairly extensive use of logging, and previously we'd just log to everything in tests to stdout. This is fine because Go automatically suppresses output to stdout without an additional `-test.v` verbose flag, and because tests ran sequentially, even when testing verbosely the output looked fine, with logs for each test case correctly appearing within their begin/end banners.

But with `t.Parallel()`, everything became mixed together into a big log soup:

```
=== RUN   TestClusterCreateRequest/StorageTooSmall
--- PASS: TestClusterCreateRequest (0.00s)
    --- PASS: TestClusterCreateRequest/StorageTooSmall (0.00s)
=== CONT  TestMultiFactorServiceList
=== RUN   TestMultiFactorServiceList/Success
=== RUN   TestMultiFactorServiceUpdate/SuccessWebAuthn
time="2023-08-20T22:26:28Z" level=info msg="password_hash_line: Match result: success [account: eee5c815-b7c6-4f19-8e1d-92428eed32ab] [hash time: 0.000496s]" account_id=eee5c815-b7c6-4f19-8e1d-92428eed32ab hash_duration=0.000496s hash_match=true
=== RUN   TestClusterServiceDelete/Owl410Gone
=== RUN   TestMultiFactorServiceList/Pagination
time="2023-08-20T22:26:28Z" level=info msg="sessionService: password_hash_upgrade_line: Upgraded password from \"argon2id\" to \"argon2id\" [account: eee5c815-b7c6-4f19-8e1d-92428eed32ab] [hash time: 0.000435s]" account_id=eee5c815-b7c6-4f19-8e1d-92428eed32ab new_algorithm=argon2id new_argon2id_memory=1024 new_argon2id_parallelism=4 new_argon2id_time=1 new_hash_duration=0.000435s old_algorithm=argon2id old_hash_iterations=0
=== RUN   TestClusterUpgradeServiceCreate/HobbyMaximum100GB
=== RUN   TestClusterServiceCreate/WithPostgresVersionID
=== RUN   TestMultiFactorServiceUpdate/WrongAccountNotFoundError
=== RUN   TestClusterServiceForkCreate/WithTargetTime
--- PASS: TestMultiFactorServiceList (0.01s)
    --- PASS: TestMultiFactorServiceList/Success (0.00s)
    --- PASS: TestMultiFactorServiceList/Pagination (0.00s)
=== CONT  TestClusterServiceActionTailscaleDisconnect
=== RUN   TestClusterServiceActionTailscaleDisconnect/Success
time="2023-08-20T22:26:28Z" level=info msg="password_hash_line: Match result: success [account: eee5c815-b7c6-4f19-8e1d-92428eed32ab] [hash time: 0.000828s]" account_id=eee5c815-b7c6-4f19-8e1d-92428eed32ab hash_duration=0.000828s hash_match=true
```

This isn't usually a problem because you're not reading the logs anyway, but quickly becomes one if you get a test failure, and only have senseless noise around it to help you debug.

The fix for this is [`t.Logf`](https://pkg.go.dev/testing?#T.Logf), which makes sure to collate log output for to the particular test case that emitted it. This will generally require a shim to use with a logging library like:

``` go
// tlogWriter is an adapter between Logrus and Go's testing package,
// which lets us send all output to `t.Log` so that it's correctly
// collated with the test that emitted it. This helps especially when
// using parallel testing where output would otherwise be interleaved
// and make debugging extremely difficult.
type tlogWriter struct {
    tb testing.TB
}

func (lw *tlogWriter) Write(p []byte) (n int, err error) {
    // Unfortunately, even with this call to `t.Helper()` there's no
    // way to correctly attribute the log location to where it's
    // actually emitted in our code (everything shows up under
    // `entry.go`). A good explanation of this problem and possible
    // future solutions here:
    //
    // https://github.com/neilotoole/slogt#deficiency
    lw.tb.Helper()

    lw.tb.Logf((string)(p))
    return len(p), nil
}
```

Then with Logrus for example:

``` go
func Logger(tb testing.TB) *logrus.Entry {
    logger := logrus.New()
    logger.SetOutput(&tlogWriter{tb})
    return logrus.NewEntry(logger)
}
```

Now when a test fails, any logs it produced are grouped correctly:

``` go
--- FAIL: TestSessionServiceCreate (0.05s)
    --- FAIL: TestSessionServiceCreate/PasswordHashAlgorithmUpgrade (0.05s)
        entry.go:294: time="2023-08-20T22:34:15Z" level=info msg="password_hash_line: Match result: success [account: 81b967f7-4f5c-4ab4-b1d7-3c455db35767] [hash time: 0.000694s]" account_id=81b967f7-4f5c-4ab4-b1d7-3c455db35767 hash_duration=0.000694s hash_match=true
        entry.go:294: time="2023-08-20T22:34:15Z" level=info msg="sessionService: password_hash_upgrade_line: Upgraded password from \"argon2id\" to \"argon2id\" [account: 81b967f7-4f5c-4ab4-b1d7-3c455db35767] [hash time: 0.011716s]" account_id=81b967f7-4f5c-4ab4-b1d7-3c455db35767 new_algorithm=argon2id new_argon2id_memory=19456 new_argon2id_parallelism=4 new_argon2id_time=2 new_hash_duration=0.011716s old_algorithm=argon2id old_hash_iterations=0
        session_service_test.go:197:
                Error Trace:    /Users/brandur/Documents/crunchy/platform/server/api/session_service_test.go:197
                                                        /Users/brandur/Documents/crunchy/platform/server/api/session_service_test.go:158
                Error:          artificial failure
                Test:           TestSessionServiceCreate/PasswordHashAlgorithmUpgrade
```

Bridges for common loggers like slog are usually available as public packages. [Slogt](https://github.com/neilotoole/slogt), for example.

### goleak (#goleak)

Our tests use [goleak](https://github.com/uber-go/goleak) to detect any accidentally leaked goroutines, a practice that I'd recommend since leaking goroutines without realizing it is easily one of Go's top footguns.

Previously, we had a pattern in which every test case would check itself for goroutine leaks, but adding `t.Parallel()` broke the pattern because test cases running in parallel would detect each other's goroutines as leaks.

The fix was to use goleak's built-in `TestMain` wrapper:

``` go
func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}
```

Leaked goroutines are only detected at package-level granularity, but as long as you're starting off from a baseline of no leaks, that's good enough to detect regressions.

## Other notes (#other-notes)

### Requiring `t.Parallel()` in tests, but not subtests (#tests-not-subtests)

By default the `paralleltest` lint will not only require that every test case define `t.Parallel()`, but that every subtest (i.e. `t.Run("Subtest", func(t *testing.T) { ... })`) define it as well. This is generally the right thing to do because it means that parallelism has better granularity and therefor more likely to produce more optimal throughput and lower the total runtime.

Due to a historical tech decision made long ago, we were ubiquitously using a testing convention within test cases where we had plenty of subtests, but subtests were not parallel safe because they were all sharing a single `var` block.

Refactoring to total parallel-safety would've taken dozens of hours and wasn't a good use of time, so we declared `t.Parallel()` at the granularity of test cases but _not_ subtests to be "good enough". I added an [`ignoremissingsubtests` option to `paralleltest`](https://github.com/kunwardeep/paralleltest/pull/32) to support that, and if your set up is anything like ours, maybe that'll help you:

``` yaml
linters-settings:
  paralleltest:
    # Ignore missing calls to `t.Parallel()` in subtests. Top-level
    # tests are still required to have `t.Parallel`, but subtests are
    # allowed to skip it.
    #
    # Default: false
    ignore-missing-subtests: true
```

## Takeaways (#takeaways)

As noted above, it's not exactly Go convention to make ubiquitous use of `t.Parallel()`. That said, it's reduced our test iteration time for large packages by 30-40%, and that's enough of a development win that I personally intend to use it for future Go projects.

And although increased test speed is its main benefit, when combined with `go test . -race` it's actually managed to help suss out some tricky parallel safety bugs that weren't being caught with sequential-only test runs. That's a big advantage because that whole class of bug is _very_ difficult to debug in production.

Activating `t.Parallel()` everywhere for an existing project could be a big deal, but integrating it from the beginning has very little ongoing cost, and might yield substantials benefits later on.
