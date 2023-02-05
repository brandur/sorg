+++
hook = "A test helper in Go to avoid long, unsightly assertion lists."
published_at = 2023-02-04T13:22:03-08:00
title = "PartialEqual"
+++

File this under "maybe a good idea, maybe a bad idea", but it's something I did.

In our Go codebase, we often find ourselves with long, unsightly laundry lists of assertions like this (using [testify/require](https://pkg.go.dev/github.com/stretchr/testify/require)):

``` go
resp, err := apitest.InvokeHandler(svc.Create, ctx, req)
require.NoError(t, err)
require.Equal(t, dbsqlc.EnvironmentProduction, *resp.Environment)
require.Equal(t, req.IsHA, *resp.IsHA)
require.Equal(t, req.MajorVersion, resp.MajorVersion)
require.Equal(t, req.PlanID, resp.PlanID)
require.Equal(t, eid.EID(version.ID), resp.PostgresVersionID)
require.Equal(t, req.RegionID, resp.RegionID)
require.Equal(t, req.Storage, resp.Storage)
require.Equal(t, req.TeamID, resp.TeamID)
```

They're not only ugly, but they have a bad tendency to quickly fall out of date as new fields are added, and before too long it's not clear whether they were even meant to be exhaustive or not.

I recently introduced a new alternative called `PartialEqual`. Here's the code above rewritten to use it:

``` go
resp, err := apitest.InvokeHandler(svc.Create, ctx, req)
require.NoError(t, err)
prequire.PartialEqual(t, &apiresourcekind.Cluster{
		Environment:  ptrutil.Ptr(dbsqlc.EnvironmentProduction),
		IsHA:         ptrutil.Ptr(req.IsHA),
		MajorVersion: req.PostgresVersionID.Int32,
		PlanID:       req.PlanID,
		RegionID:     req.RegionID,
		Storage:      req.Storage,
		TeamID:       req.TeamID,
}, resp)
```

Why not just use `Equal`? More often than not, the structs I'm comparing contain something like a timestamp (think `CreatedAt` or `UpdatedAt`) or other volatile field that doesn't have predictable value to test against (think like `api_key_secret`). Also, `PartialEqual` gives us the option to compare just a subset of fields that we care about.

`PartialEqual` works by looking at which fields are non-zero on the expected struct (left side) and then comparing only those values against the actual struct (right side). Anything that's zero (say a `CreatedAt` containing only a default `time.Time` value) is skipped.

Advantages:

* A lot more human-readable. It's much easier to see a missing field and know to update it.

* Go's formatted lines everything up nicely, contributing even more to readability.

But, it does have a downside brought on by Go's annoying design around zero values. It's not possible even through reflection to know whether a value in a struct explicitly set to a zero value like `false`, `0`, `""`, or whether it was left out, so `PartialEqual` may accidentally ignore values that happen to be zero, thereby creating a sizable footgun.

The workaround is to use `PartialEqual` for non-zero values, and standard assertions for zero ones. Here's a degenerate case from our code where we care inordinately more than usual about zeroes and had to put in a lot of extras:

``` go
updatedQuery, err := queriesEphemeral.QueryGetByID(ctx, query.ID)
require.NoError(t, err)
prequire.PartialEqual(t, dbsqlcephemeral.Query{
		NumFailures:  1,
		ResultFields: pgtype.JSONB{Status: pgtype.Null},
		Status:       string(dbsqlcephemeral.QueryStatusFailed),
}, updatedQuery)
require.NotZero(t, updatedQuery.FinishedAt)
require.Zero(t, updatedQuery.NextRunAt)
require.Greater(t, *updatedQuery.LastRunDuration, time.Duration(0))
require.Zero(t, updatedQuery.ResultS3Key)
require.Zero(t, updatedQuery.RunningAt)
```

## Alternatives (#alternatives)

I sent this out to a private Go channel for reactions, and `PartialEq` didn't exactly get a stellar reception, although no one was doing anything that much better.

The most convincing alternative was the use of something like [go-cmp](https://pkg.go.dev/github.com/google/go-cmp/cmp) with its option for ignoring fields:

``` go
var ignoreTimestamps = cmpopts.IgnoreFields(recordio.Phase{},
    "CreatedAt",
    "UpdatedAt",
)
```

This is okay, but has a few problems of its own:

* Having to refer to property names with strings is bad. Makes refactoring harder and breaks IDE symbol lookups. (This isn't specific to go-cmp. There's no way except strings to refer to a field in Go.)

* It assumes that comparing every field in an object is always what you're trying to do, which is often wrong. I find that it's often better for refactoring agility to have one test case that's exhaustive, but then for other to only look at subsets of fields that are interesting for the particular test case. It saves a lot of updating when a new field is added or an old one removed.

All in all, this pattern's not a strong recommendation, but we've had `PartialEqual` in for a few months now and it's a tool I find myself reaching for frequently, and is much better than what we were doing before.

## Prototype implementation (#implementation)

I put [my implementation of `PartialEqual` into a Gist](https://gist.github.com/brandur/7b459a1ed81bfd041fabf05dc34265e3) that you can clone down, but mainly to act as reference. The code's not tremendous by any means, and its got a dependency on testify/require which isn't optimal, but if you're interesting in trying the pattern, it'll give you a start.