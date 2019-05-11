+++
hook = "On superior operational visibility and data rationing."
published_at = 2016-08-28T18:51:50Z
title = "Splunk"
+++

I've been using Splunk through my last two jobs, and consider it an absolutely
invaluable tool. Its capacity to run ad-hoc queries against huge volumes of
loosely structured log data and come back with results in seconds is impressive
to say the least. While statsd-style dashboards have made up a big portion of
our operational visibility at both Stripe and Heroku, in both cases Splunk has
been the _only_ practical tool at our disposal for piercing into our stack and
pulling out specific information, or diagnosing previously unseen problems.

All of that said, I'm not sure whether I could recommend it. Splunk's amazing
utility is somewhat offset by its burdensome pricing model, which is largely
based around the number of bytes that it ingests; with no option for
"unlimited". The tiers are big-ish (on the scale of terabytes a day), but after
you're handling significant traffic and emitting a proportional volume of logs,
they're not enough.

The incremental cost to Splunk for additional data is $0 -- we run all the
infrastructure [1] and are already paying more for that as our volume of data
grows. The only model I can think to compare it to are the "old world"
databases like Oracle that have schemes like [per-core
pricing][oracle-pricing].

To slow our creep up through Splunk's pricing tiers, we've started to apply a
few techniques for data rationing. Logs have been sifted through for unneeded
lines, metrics that could go to statsd or over another bus instead, and any
other repetitive information. New additions are vetted for their impact on
emitted volume before they're merged.

Inevitably, some of that information which was removed or not added would have
been useful for getting insight into a problem in production or resolving a
difficult bug. And while some logging hygiene is certainly laudable, I'm not
sure that in this day and age terabytes should be treated as if they were
scarce pearls painstakingly collected from the bottom of the Pacific.

[1] Which currently sit at a considerable ~80 nodes worth of AWS instances.

[oracle-pricing]: https://en.wikipedia.org/wiki/Oracle_Database#Pricing
