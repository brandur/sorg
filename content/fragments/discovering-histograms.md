+++
hook = "A good blog post on why aggregating aggregates doesn't work, and using histograms to do the job instead."
published_at = 2024-01-13T12:35:38-08:00
title = "Discovering histograms, kicking off metric rollups"
+++

A [great post from Circonus](https://www.circonus.com/2018/11/the-problem-with-percentiles-aggregation-brings-aggravation/) on why aggregating percentiles doesn't work, and explaining the use of histograms to get it right.

I'm just breaking ground on a project to produce rollups for our metrics. Until now we've been using only the raw metric data and Postgres' `percentile_cont` to aggregate it, which considering the bluntness of the instrument, has been working amazingly well. But as you might expect, it's slow to load metrics for longer timeframes like a week or a month because there's so much data to crunch through. We always knew that we'd have to do something about that, but our naive approach worked for the time being and let us kick the can forward a ways.

I've had a few people encourage me to pick up a specialized time series database or Postgres extension to resolve the problem, but I'm trying to avoid doing so. Our set up right now is uses vanilla Postgres only with no specialized extensions, which keeps it easy to operate, easy to upgrade, and portable.

By writing code to do our own time series aggregation, I know I'm walking a fine line with <acronym title="Not Invented Here">NIH</acronym> territory, but similar decisions we've made have so far shown a good track record (e.g. writing a ~100 line partition manager [instead of pulling in `pg_partman`](/fragments/postgres-partitioning-2022#partition-management)). A little more code needs to be written, but our development environment remains trivially easy to set up, and our operational environment trivially easy to understand. All the hard math will be outsourced to a project like [HdrHistogram](https://github.com/HdrHistogram/HdrHistogram) (Richard's suggestion), which has a Go library and seems to be well maintained.

Along with a much faster metrics product, I'm looking forward to seeing what kind of retention the more storage efficient histograms will enable. We're ingesting ~10 GB worth of raw metrics a day, which means that even if we could practically query a data set that large, it'd be expensive to store, so we let older metrics roll off quite quickly. By having histograms for the granularity of say, a day, we could conceivably provide metrics lookbacks of a year or more quite cheaply, which even if not all that useful for day to day use, would be a cool feature.

Not much more to say for now, but I'll do a postmortem when the project's done.