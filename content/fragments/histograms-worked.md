+++
hook = "Following up on the use of histograms to generate metric aggregates that can be used to generate charts over much wider time ranges like a week or a month."
published_at = 2024-05-04T07:45:56+02:00
title = "Histograms worked"
+++

I wrote a few months back about how I was looking into [using histograms for metrics rollups](/fragments/discovering-histograms) to power Crunchy Bridge's metrics dashboards. Previously we'd been aggregating over raw data points, which worked surprisingly well, but was too slow for generating series for longer time ranges like a week or a month.

I'm writing this follow up because in the original post, I said I would. In short, this approach worked unreasonable well. Our raw metrics are partitioned into a series of tables by day like:

* `metric_point_20240502`
* `metric_point_20240503`
* `metric_point_20240504`

Where each partition is about 10 GB (although this keeps growing as new clusters come online).

I wrote an aggregator that iterates through the raw data, divides it up by period, and generates a series of aggregates of varying coarseness like 5 minutes, 1 hour, 6 hours, etc., each of which contains a histogram approximating the data for that period. These in turn get stored to their own partitioned table like:

* `metric_point_aggregate_20240502`
* `metric_point_aggregate_20240503`
* `metric_point_aggregate_20240504`

Each of these is about 1 GB in size, but contains multiple rows per series for each coarsenesses, so although they're nominally 1/10th the size of the raw data, the average amount of data that needs to be iterated is much less than that.

When rendering metric charts in Dashboard, the service decides which aggregate coarseness to use based on requested interval and resolution:

``` go
var metricViewAggregates = []*metricViewAggregate{
	{Name: "5m", Interval: 5 * time.Minute},
	{Name: "1h", Interval: 1 * time.Hour},
	{Name: "6h", Interval: 6 * time.Hour},
	{Name: "1d", Interval: 24 * time.Hour},
}

func chooseBestFitAggregate(period time.Duration, numPoints int) *metricViewAggregate {
	// Start with the most coarse aggregate (so we load the least amount of
	// data), and then our way down from there.
	for i := len(metricViewAggregates) - 1; i >= 0; i-- {
		metricViewAggregate := metricViewAggregates[i]
		if int(period/metricViewAggregate.Interval) >= numPoints {
			return metricViewAggregate
		}
	}
	return nil
}
```

So for example, when rendering a 30 day view, at default resolution we generate 30 intervals for the series that'll go in the chart. The algorithm above will pick 1 day aggregates, so instead of iterating 10s of thousands of raw metric points, it'll select 30 aggregate rows, and use the histograms contained therein to approximate percentiles. Thousands of times less data needs to be retrieved from the Postgres heap, and it's returned in a tiny fraction of time.

And recall that histograms have the amazing property that they're mergeable. So if we wanted 2x resolution for that 30 day view at 60 points instead of 30, the 1 day aggregates are no longer sufficient, so the algorithm will select 30 * 4 = 120x 6 hour aggregates instead, and merge them in pairs down to 60 final points.

This is all fiddly enough that after having wrote it, I was very concerned about the presence of bugs, even with thousands of lines worth of automated unit tests. I rolled the new system out behind a feature flag so that I could pull two browsers up side by side and compare the chart renderings with and without aggregates. To my everlasting surprise, they were practically identical on my first try (histograms produce estimates, so they were never expected to be an identcal match). Code rarely survives first contact with production, but sometimes it does. Especially where a very thorough test suite is involved.

I rolled the project forward, and was able to make one week and one month views for our charts available to customers with reasonable haste, and relatively little trouble.

## Choice of package (#choice-of-package)

There were a couple different histogram implementations to choose from. The [Go port of HdrHistogram](https://github.com/HdrHistogram/hdrhistogram-go) had been mentioned to me, but turned out to be unsuitable for our use because it required that each histogram specify its minimum and maximum bounds in advance. This limitation has been corrected in HdrHistogram implementations for other languages, but the Go port doesn't see as much development activity, and hasn't been updated for three years. I went with the [OpenHistogram Go port](https://github.com/openhistogram/circonusllhist), which although hasn't accumulated as many GitHub stars, has seen more recent commit activity.