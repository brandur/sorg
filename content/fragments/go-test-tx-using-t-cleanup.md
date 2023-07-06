+++
hook = "An elegant way of combining test transactions with Go's built-in test abstractions."
published_at = 2023-07-05T18:56:24-07:00
title = "A `TestTx` helper in Go using `t.Cleanup`"
+++

I'm a big fan of the use of test transactions. They've got some downsides (which I'll dig into elsewhere), but they're extremely fast, allow practically limitless parallelism, and remove the need for complicated and expensive cleanup subsystems.

Previously, I'd been using them with a block-style test helper into which a function is injected to provide the inner body between transaction start and rollback, but I was pointed in the direction of Go's [`t.Cleanup()`](https://pkg.go.dev/testing#T.Cleanup) helper which runs an arbitrary operation after a test finishes.

A simple test transaction helper combining pgx and `t.Cleanup()`:

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

Use of it within a test is a succinct `tx = TestTx(ctx, t)`:

``` go
func TestTestTx(t *testing.T) {
	t.Parallel()

    var (
        ctx = context.Background()
        tx  = TestTx(ctx, t)
    )

	_, err := tx.Exec(ctx, "CREATE TABLE test_tx_table (id bigint)")
	require.NoError(t, err)
}
```

Some may be wondering: why not just use `defer`? Well, there's a subtle advantage to `t.Cleanup()` that may not be intuitive, but obvious when you point it out. Defer only works in the immediate function, whereas `t.Cleanup()` lets a `*testing.T` be injected into a helper, and for that helper to attach a defer-equivalent. Using `defer` every test case would have to handle its own rollback individually:

``` go
tx := TestTx(ctx, t)
defer tx.Rollback(ctx)
```

Two lines isn't that much worse than one, but it adds up over thousands of tests. The abstracted `TestTx` also allows for other niceties like logging and error checking that'd add more lines to the `defer` in every test case.

## Complete (#complete)

Here's a more complete version that:

* Includes a lazily-initialized pool. When `TestTx` is put into a test helper package, other packages can include it and a pool's only initialized if a test transaction is actually used.
* Takes a `testing.TB` as parameter instead of `*testing.T`, making it also compatible with benchmarks and fuzz tests.

``` go
// A pool and mutex to protect it, lazily initialized by TestTx. Once open, this
// pool is never explicitly closed, instead closing implicitly as the package
// tests finish.
var (
	dbPool   *pgxpool.Pool //nolint:gochecknoglobals
	dbPoolMu sync.RWMutex  //nolint:gochecknoglobals
)

func TestTx(ctx context.Context, tb testing.TB) pgx.Tx {
	tryPool := func() *pgxpool.Pool {
		dbPoolMu.RLock()
		defer dbPoolMu.RUnlock()
		return dbPool
	}

	getPool := func() *pgxpool.Pool {
		if dbPool := tryPool(); dbPool != nil {
			return dbPool
		}

		dbPoolMu.Lock()
		defer dbPoolMu.Unlock()

		var err error
		dbPool, err = pgxpool.New(ctx, os.Getenv("TEST_DATABASE_URL"))
		require.NoError(tb, err)

		return dbPool
	}

	tx, err := getPool().Begin(ctx)
	require.NoError(tb, err)

	tb.Cleanup(func() {
		err := tx.Rollback(ctx)

		// Try to look for an error on rollback because it does occasionally
		// reveal a real problem in the way a test is written. However, allow
		// tests to roll back their transaction early if they like, so ignore
		// `ErrTxClosed`.
		if !errors.Is(err, pgx.ErrTxClosed) {
			require.NoError(tb, err)
		}
	})

	return tx
}
```