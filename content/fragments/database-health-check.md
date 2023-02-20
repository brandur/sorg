+++
hook = "Health checks suitable for status pages that show down when the application is down."
published_at = 2023-02-19T13:17:21-08:00
title = "Honest health checks that hit the database"
+++

In web services, a common pattern is a health check, often used for something like a reverse proxy (e.g. [ELB](https://aws.amazon.com/elasticloadbalancing/)) to know whether a node is online. Our API's health check is used as a target [for our status page](https://status.crunchybridge.com/s/crunchy/963934) to show uptime.

We're running an HA-ready `premium-0` database on Heroku [1], and a few weeks ago it has some brief downtime as its leader was lost and HA standby rotated into place. The downtime wasn't well represented on our status page, and although more green might make us look better short term, it's not honest to people who might be experiencing the outage and trying to confirm it.

Downtime wasn't represented because the health check endpoint was a no-op HTTP handler that ran perfectly fine even when the database was down. I followed up by adding a separate health check at `GET /health-checks/complete` which makes an extra effort to exercise the stack more thoroughly by opening connections to both our databases (the second being our [ephemeral DB](/fragments/ephemeral-db)) and running a `SELECT 1` to make sure they work:

``` go
{
    errGroup, ctx := errgroup.WithContext(ctx)

    errGroup.Go(func() error { return checkDatabase(ctx, svc.Begin) })
    errGroup.Go(func() error { return checkDatabase(ctx, svc.BeginEphemeral) })

    if err := errGroup.Wait(); err != nil {
        return nil, apierror.NewServiceUnavailableErrorf(ctx, "Health check error: %v.", err)
    }
}

// Does a `SELECT 1` against a database as a basic check that it's online.
func checkDatabase(ctx context.Context, begin func(context.Context) (db.Txer, error)) error {
    tx, err := begin(ctx)
    if err != nil {
        return xerrors.Errorf("error starting transaction: %w", err)
    }
    defer func() { tx.RollbackLogged(ctx) }()

    _, err = tx.Exec(ctx, "SELECT 1")
    if err != nil {
        return xerrors.Errorf("error pinging database: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return xerrors.Errorf("error committing transaction: %w", err)
    }

    return nil
}
```

It's mounted as a separate health check from the more basic one in case we want to put our service being an ELB, in Kubernetes, etc. A health check for purposes of a reverse proxy should still return a `200 OK` even when the database is down so that the reverse proxy doesn't accidentally take all its nodes out out of rotation. Instead, each node should be able to detect the down database and return an appropriate error.

It's not strictly necessary to run the two database checks in parallel, but like in many other circumstances, Go's [`errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup) makes this so easy and problem-free that I do it anyway.

[1] We'd originally put this on Heroku to avoid any bootstrapping problem, but given our present day architecture in which the API isn't a dependency for database uptime (only the backend state manager below it), we'll dogfood this by moving to one of our own databases soon.
