+++
hook = "Using `DialFunc` to return a minimal stub for `net.Conn` that can simulate hard-to-reproduce conditions like an error on `Close`."
published_at = 2024-03-02T15:36:03-08:00
title = "Stubbing degenerate network conditions in Go with `DialFunc` and `net.Conn`"
+++

While fixing a [bug in River's pgx driver's listener](https://github.com/riverqueue/river/pull/246) the other day, I got to a point where I was trying to write a regression test, but found it difficult because getting the `Close` function on a pgxpool connection to return an error is practically impossible through normal usage. I thought I'd be able to do it with a prematurely cancelled context, but it didn't do the trick.

Reading pgx's source, errors on `Close` are generally avoided in favor of quietly cleaning up disposed of resources, and more or less the only time an error will ever be returned is if there's a problem deep at the network level in the underlying [`net.Conn`](https://pkg.go.dev/net#Conn).

Even there, I was having difficulty because even after reading Go source, it wasn't readily apparent how I could prompt `net.Conn`'s `Close` to return an error, even if I could inject my own through three layers worth of pgx abstractions.

But I did [find a way that's reasonably clean](https://github.com/riverqueue/river/pull/250). Like many other well-behaved Go packages involving networking, pgxpool provides a [configurable `DialFunc`](https://pkg.go.dev/github.com/jackc/pgx/v5/pgconn#Config) that's implemented by a function that returns `(net.Conn, error)`.

``` go
type DialFunc func(
    ctx context.Context,
    network, addr string,
) (net.Conn, error)
```

## Minimal viable stub (#stub)

`net.Conn` is an interface, so we can combine `DialFunc` with a lightweight stub that embeds a real `net.Conn` (so functions on the interface that we don't need to stub are inherited), but lets individual functions be overridden as needed:

``` go
type connStub struct {
    net.Conn

    closeFunc func() error
}

func newConnStub(conn net.Conn) *connStub {
    return &connStub{
        Conn: conn,

        closeFunc: conn.Close,
    }
}

func (c *connStub) Close() error {
    return c.closeFunc()
}
```

Wrap a connection from `net.Dailer` with a stub, and then we're well positioned to easily have `Close` return an error of choice:

``` go
var config *pgxpool.Config = testPoolConfig()
config.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
    // Dialer settings come from pgx's default internal one (not exported).
    conn, err := (&net.Dialer{KeepAlive: 5 * time.Minute}).Dial(network, addr)
    if err != nil {
        return nil, err
    }

    connStub = newConnStub(conn)
    return connStub, nil
}

listener := &Listener{dbPool: testPool(ctx, t, config)}

expectedErr := errors.New("conn close error")
connStub.closeFunc = func() error {
    return expectedErr
}

require.ErrorIs(t, listener.Close(ctx), expectedErr)
```
