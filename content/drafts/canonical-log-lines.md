---
title: Canonical Log Lines
published_at: 2016-08-24T14:38:47Z
hook: A lightweight and technology agnostic analytical pattern for easy
  visibility into production systems.
location: San Francisco
---

_Canonical log lines_ are a lightweight pattern used within Stripe for improved
service observability. They act as a middle ground between other types of
analytics in that they present a good trade-off between ease of access and
flexibility.

For some background, if we look at standard practices in production systems, we
could say that they emit tiers of operational information. For example:

* Metrics (i.e. emitted to services like statsd, Librato, Datadog, etc.).
* Log traces.

While **metrics** provide fast feedback on specific and commonly-used system
measurements, they're not well suited for allowing information to be queried
arbitrarily and ad-hoc. They're perfect for answering a question that's known
ahead of time like, _"how many requests per second is my stack doing?"_,
but less useful for questions that I think up on the spot and need immediate
answers for like, _"how many requests per second is userXYZ making to my stack
via TLS 1.0?"_

**Log traces** can answer the latter question, but often make such analysis
difficult because they tend to have poor signal-to-noise ratio when looking for
specific information. They can be difficult to navigate even with sophisticated
logging systems like Splunk.

Canonical log lines aim to be a middle tier of analytics to help bridge that
gap:

!fig src="/assets/canonical-log-lines/observability-tiers.svg" caption="The tiers of observability, showing the trade-off between query flexibility and ease of reference."

## What Are They? (#what-are-they)

The concept is simple: canonical lines are one big log line (probably in
[logfmt](/logfmt)\) style that is emitted at the end of a request [1] and which
defines fields that contain all of that request's key information.

!fig src="/assets/canonical-log-lines/canonical-log-lines.svg" caption="Canonical log lines being emitted (and ingested) for each request."

For example, we might include:

* HTTP verb, path
* Source IP and user agent
* [Request IDs](/request-ids)
* Response status
* Error ID and message (for a failed request)
* ID of the API key used, OAuth application and scope
* ID and email of an authenticated user

We could also include other internal information:

* Name of the service, `HEAD` Git revision, release number
* Timing information (e.g. duration of the request, time spent in database)
* Remaining and total rate limits

Here's an example in [logfmt](/logfmt) style (line breaks added for clarity):

!fig src="/assets/canonical-log-lines/example-line.svg" caption="What a canonical log line looks like in raw form as it's being emitted."

## Storage (#storage)

After emitting canonical lines, the next step is to put them somewhere useful.
At Stripe we use a two prong approach combining Splunk and Redshift.

### Splunk (#splunk)

Splunk is a powerful shorter-term store that's great for getting fast insight
into online systems. It's great for:

1. Powerful querying syntax that's quite terse and fast to write.
2. Very fast: queries come back quickly, especially if they're scoped down to a
   small slice of the total data volume (i.e. like the last hour worth of logs).
3. Ingested directly from syslog so that data is available almost as soon as
   it's emitted.
4. By tagging every line emitted to Splunk with [request IDs](/request-ids), we
   can easily cross-reference canonical lines with any other part of the raw
   trace that came from a request.

These traits make Splunk ideal for operational work where we might be in the
midst of an incident and need information immediately.

The downside of Splunk is that it's expensive to license and run, and your
retention is generally limited by your total capacity being eaten up by the
reams of raw information being pushed into the system. It's not an unusual
sight to see our operations teams at Stripe trying to prune the traces of our
highest traffic systems to keep Splunk running under quota [2].

### Redshift (#redshift)

The other system that we use to ingest canonical lines is Redshift. Any other
data warehousing system would do just as well.

Some advantages of a data warehouse:

1. Scalability: tables can be arbitrarily large and still queryable without
   trouble.
2. Low cost: especially compared to systems like Splunk, data can be archived
   for extended periods without price becoming an issue. We prune canonical
   lines after 90 days, but they could conceivably be stored for much longer.
3. By importing other non-request data into the warehouse, like say information
   on your users from the core data store, it allows greater levels of deep
   insight by stitching together data generated from different data sources.

We import data by emitting lines over a fast queueing system (NSQ), archiving
batches of it to S3, and then periodically running a `COPY` pointing to the
bucket from Redshift.

## Examples

### Example: HTTP 500s By Breakage

One of the hazards of any software stack is that unexpected breakages will
happen. For a typical web service, this often takes the form of an exception
raised internally and then converted to an HTTP 500 and returned to the user.

While it's reasonably likely that a web service already has a dashboard in
place for the frequency of 500s that are occurring, what if I wanted, for
example, to drill into the errors from one particular API endpoint. By scoping
canonical log lines down to 500 responses coming from a single API method, that
information is easily available:

!fig src="/assets/canonical-log-lines/error-timechart.png" caption="Error counts for the last week on the \"list events\" endpoint."

> canonical-api-line status=500 api_method=AllEventsMethod earliest=-7d | timechart count

By putting this query into a Splunk subsearch, I can trivially join it with
other emitted log lines. For example, by joining on a "breakage line" (one
where we log an exception), I can look at these errors grouped by class:

!fig src="/assets/canonical-log-lines/top-errors.png" caption="The names of the Ruby exception classes emitted for each error, and their relative count."

> [search canonical-api-line status=500 api_method=AllEventsMethod sourcetype=bapi-srv earliest=-7d | fields action_id] BREAKAGE-SPLUNKLINE | stats count by error_class | sort -count limit 10

I can also invert this to pull information _out_ of the canonical lines. Here
are counts of timeout errors over the last week by API version:

!fig src="/assets/canonical-log-lines/top-api-versions.png" caption="An inverted search. API versions pulled from the canonical log line and fetched by class of error."

> [search breakage-splunkline error_class=Chalk::ODM::OperationTimeout sourcetype=bapi-srv earliest=-7d | fields action_id] canonical-api-line | stats count by stripe_version | sort -count limit 10

### Example: TLS Deprecation

One project that I'm working on right now is helping Stripe users [migrate to
TLS 1.2 from older secure protocols][upgrading-tls]. TLS 1.2 will eventually be
required for PCI compliance, so we're trying to identify users who are on TLS
1.0 and 1.1 and give them some warning that an upgrade will be required.

While responding to support questions on the topic, it dawned on me that I was
running the same queries in Splunk over and over. I stopped and spent a couple
minutes creating a Dashboard in Splunk that's powered purely by canonical log
lines:

!fig src="/assets/canonical-log-lines/tls-requests-by-merchant.png" caption="Splunk dashboard for requests from a user by TLS version."

Each panel performs a search for canonical log lines matching a certain
user, excludes lines generated by calls from internal systems, then pulls
some metrics and tabulates or draws a plot of the results.

The same thing is also possible from Redshift. Here's a query to ask for any
users that have made requests to our API services using pre-1.2 versions of
TLS in the last week [3]:

```
SELECT distinct(user_id)
FROM canonical_lines.api
WHERE created > GETDATE() - '7 days'::interval
  AND tls_version < 'TLSv1.2'
ORDER BY 1;
```

## Implementation (#implementation)

Middleware makes a good home for implementing canonical log lines. This
middleware will generally be installed close to the top of the stack and inject
an object through context (i.e. `env` in Rack) which is populated by downstream
components before being finalized and emitted by the logging middleware.

For example, the logging middleware might inject an object with a `request_id`
field which will then be populated downstream by a "request ID" middleware
after it extracts one from an incoming request's headers.

Below is a basic implementation. Although Ruby code is provided here, the same
basic concept can easily be applied to any technology stack.

``` ruby
# A type containing fields that we'd like to populate for the final canonical
# log line and which can encode itself in logfmt format.
class CanonicalLogLine
  # service information
  attr_accessor :service
  attr_accessor :release
  attr_accessor :git_head

  # request identification
  attr_accessor :request_id

  ...

  def to_logfmt
    ...
  end
end

# A middleware that injects a canonical log line object into a request's #
# context and emits it to the log trace as the rest of the stack has finished
# satisfying the request.
class CanonicalLogLineEmitter < Middleware
  attr_accessor :app

  def initialize(app)
    self.app = app
  end

  def call(env)
    line = CanonicalLogLine.new
    env["app.canonical_log_line"] = line
    ...

    app.call(env)

    # Emit to logs.
    log.info(line.to_logfmt)
  end
end
```

And an example of a more complete middleware stack that we install it into:

``` ruby
App = Rack::Builder.new do
  # Top of the middleware stack.
  use CanonicalLogLineEmitter

  # Other middleware.
  use Cache
  use Deflater
  use ErrorHandler
  use RequestID
  use SSL

  run Main
end
```

## Summary

Canonical log lines are a simple middle tier of analytics that can be readily
implemented in any production stack. They don't make as convenient of a
reference as prebaked statsd-style metrics dashboards, but are infinitely more
flexible. Likewise, they don't have the sheer information density of raw log
traces, but make access to information far more convenient.

[1] I refer to canonical lines being tied to "requests" because web services
    are a ubiquitous type of app that many of us are working on these days, but
    they can just as easily be applied to other uses as well. For example,
    producing one after a background job runs or after completing a map/reduce
    operation.

[2] At 1000 requests per second, emitting one extra 200 byte line per request
    will account for 15-20 GB of additional data over a 24 hour period. With
    Splunk quotas often only in the TBs, that'll easily start eating into them.

[3] Note that I'm using a string comparison trick here when comparing TLS
    versions in that the versions `TLSv1`, `TLSv1.1`, and `TLSv1.2` happen to
    sort lexically according to their relative age.

[upgrading-tls]: https://stripe.com/blog/upgrading-tls
