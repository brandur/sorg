+++
hook = "Connecting to Postgres requires only a connection string, but there are other important options that should usually be included."
published_at = 2021-09-09T16:21:16Z
title = "Important (but not required) parameters to include when connecting to Postgres"
+++

It's pretty obvious how to get a vanilla Postgres connection up and running. What's _not_ so obvious is that you should do a little beyond the basics when constructing one -- not so much because you need them today, but because they might save you from an operational accident tomorrow.

Here are some of the most useful settings that we keep an eye on:

* **Max connections:** Especially in languages where concurrency allows for potentially unlimited connection use, it's a good idea to cap maximum connections per process so to not exhaust total allocation or accidentally slow the database. Aim for connection economy by architecting program so that threads/coroutines only check a connection out of a pool around the critical section of work where they need one, thereby allowing many parallel workers to share relatively few connections. Discussed more thoroughly in [How to Manage Connections Efficiently in Postgres](/postgres-connections).

* `application_name`: A purely informational tag, but one that should always be set. When looking at active connections in `pg_stat_activity`, the application name of the connected client is shown so you can easily tell where it's coming from. Very useful for debugging long running transactions.

* `idle_in_transaction_session_timeout`: Time after which an inactive transaction is terminated. Useful for cases where a bug might have left a process idle in transaction, or (for example) where an operator opened a psql session, ran `BEGIN`, and forgot about it. Long running transactions lead to dead rows not being vacuumed and maybe to [operational failure](/postgres-queues). Defaults to no timeout, which is dangerous.

* `statement_timeout`: Time after which an an individual query is allowed to run before being terminated. Queries that run too long in an OLTP system are usually indicative of a problem (e.g. they're stuck, you have a missing index), and almost never desirable. Defaults to no timeout, which is dangerous.

## pgx (#pgx)

We have a simple wrapper around [pgx](https://github.com/jackc/pgx)'s pool connect that makes us send required config, and assigns sensible defaults for important config that isn't:

``` go
type ConnectConfig struct {
	IdleInTransactionSessionTimeout time.Duration
	MaxConns                        int32
	StatementTimeout                time.Duration
}

func Connect(ctx context.Context,
	databaseURL, applicationName string, config *ConnectConfig) (*Conn, error) {

	if config.IdleInTransactionSessionTimeout == time.Duration(0) {
		config.IdleInTransactionSessionTimeout = 10 * time.Second
	}
	if config.MaxConns == 0 {
		config.MaxConns = 20
	}
	if config.StatementTimeout == time.Duration(0) {
		config.StatementTimeout = 10 * time.Second
	}

	pgxConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	pgxConfig.MaxConns = config.MaxConns
	runtimeParams := pgxConfig.ConnConfig.RuntimeParams
	runtimeParams["application_name"] = applicationName
	runtimeParams["idle_in_transaction_session_timeout"] =
		strconv.Itoa(int(config.IdleInTransactionSessionTimeout.Milliseconds()))
	runtimeParams["statement_timeout"] =
		strconv.Itoa(int(config.StatementTimeout.Milliseconds()))

	pool, err := pgxpool.ConnectConfig(ctx, pgxConfig)
	if err != nil {
		return nil, err
	}

	return &Conn{
		pool: pool,
	}, nil
}
```
