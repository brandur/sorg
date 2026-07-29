+++
hook = "Postgres database templates make pgtestdb's per-test database isolation remarkably quick, around 100 ms of setup time, and within spitting distance of River's schema-based approach."
published_at = 2026-07-29T01:15:00-05:00
title = "pgtestdb's template cloning approach to testing is fast"
+++

I was reminded by [Cup o' Go](https://cupogo.dev/episodes/proposals-proposals-proposals-and-faster-postgresql-tests-with-peter-downs) yesterday of the existence of Peter Downs' [pgtestdb](https://github.com/peterldowns/pgtestdb), a Go/Postgres testing package.

pgtestdb is built around Postgres [template databases](https://www.postgresql.org/docs/current/manage-ag-templatedbs.html), a built-in feature that you can try right from a vanilla psql shell:

``` sql
CREATE DATABASE dbname TEMPLATE template_to_copy;
```

Copying a template is very fast, more so than migrating a test database from scratch, and _much_ more so than some of the heavyweight Docker-based techniques some projects are using these days. At a low level, Postgres enumerates the template's relations and copies their materialized heap, index, and catalog files in 8 kB page chunks.

I remember reading about this feature years ago, but to be honest I'd forgotten it existed, and I was curious how it performed compared to other testing approaches, so I had Codex splice pgtestdb into River's test suite to see how it'd fare.

I like to think that River's testing methodology is more or less a gold standard for speed and reliability. It uses a custom set of test helpers that isolate test cases based on schema, an approach that's slower than [test transactions](/fragments/go-test-tx-using-t-cleanup), but which has some advantages:

* Leaves test state in place to examine in case of a failing test.
* Enables testing of database-wide features like listen/notify.
* Enables testing edges around multiple transactions interacting and rollbacks.

## Results (#results)

In Postgres a schema is lighter weight than a database, so the schema-based approach has that edge. However, you can't clone a schema, so the schema-based approach has to run migrations every time, giving pgtestdb a distinct advantage in that respect.

That should give us an interesting comparison. Here are the results I got:

| Method | Count | Mean | p90 | p95 | Max |
|---|---:|---:|---:|---:|---:|
| pgtestdb clone | 466 | 98.4ms | 247.4ms | 299.5ms | 465.1ms |
| Create + migrate schema | 81 | 99.4ms | 152.1ms | 209.0ms | 327.0ms |

What we find is that the timing of both approaches is remarkably similar, right around 100 ms of setup time.

I'd always internalized that anything involving creating new databases would be relatively slow, so I was surprised at how fast pgtestdb's approach turned out to be here.

I'm going to leave River's tests on its existing schema-based method given it's already fast and testing schema isolation is incidentally useful in verifying that River's [schema-based configuration](https://riverqueue.com/docs/alternate-schema) works as advertised, but I'm going to add a recommendation in our docs for pgtestdb, particularly for users aiming to test end-to-end (i.e. job inserted by client → fully worked by worker).

## Optimizing via reuse (#optimizing-via-reuse)

I was sandbagging a little above. Although _setup time_ for the schema-based approach is similar to pgtestdb's full databases, overall the test suite runs ~3.5x faster on the former:

| Method                  | Wall time |
|-------------------------|----------:|
| pgtestdb clone          | 51.07s    |
| Create + migrate schema | 14.54s    |

But it's not because schemas are that much faster. River's test helpers have a useful optimization in that they'll create as many test schemas as Go's instantaneous parallelization requires, but keep them pooled as test cases finish. If an unclaimed schema is ready, a test case will clean and reuse it instead of generating a new one from scratch [1].

This is a little easier said than done because you need to think about details like schema version -- i.e. when testing across schema versions, each test case must only reuse a schema on the same version it expects. This is very doable, of course, but takes a little thought. I wrote River's implementation pre-LLM, and it took me a few days to squeeze all the bugs out.

I mention reuse because it could be done with pgtestdb as well, potentially as part of the package, or as an augmentation in projects that call into it. 100 ms to bootstrap a test database is pretty fast, but if you're building a full application that's going to have 10,000 tests, ideally you want a test setup on the order of 10x faster than that (10-20 ms, the lower the better).

[1] If a test case fails, its schema isn't reused, leaving the state available for inspection/debugging.
