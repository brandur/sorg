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

For some background, if we look at a standard production system, we could say
that it emits multiple "tiers" of analytical information. For example:

* Metrics (i.e. emitted to services like statsd, Librato, Datadog, etc.).
* Log traces.

While **metrics** provide fast feedback on specific and commonly-used system
measurements, they're not very good at allowing information to be queried
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

The concept is dead simple: canonical lines are one big log line (probably in
[logfmt](/logfmt)\) style that is emitted at the end of a request [1] and which
defines fields that contain all of that request's key information.

!fig src="/assets/canonical-log-lines/canonical-log-lines.svg" caption="Canonical log lines being emitted (and ingested) for reach request."

Here are some examples of what might be included:

* Basic request vitals like HTTP verb, host, path, source IP, and user agent.
* Any [request IDs](/request-ids) that the request may have been tagged with.
* Response information like status and content type.
* Any error information (if the request errored) like error ID and message.
* Authentication information like the ID of the API key used, or the OAuth
  application and scope in use (if applicable).
* Information on the authenticating user like their human-friendly label (say
  an email) or internal identifier for quick and stable reference.
* General information about the running app like its name, HEAD Git revision,
  and current release number.
* Aggregate timing information like the total duration of the request, or the
  total amount of time spent in database queries.
* Rate limiting: whether rate limiting occurred, what a user's total limit is,
  and how much of it is remaining.

Here's an example in [logfmt](/logfmt) style (line breaks added for clarity):

!fig src="/assets/canonical-log-lines/example-line.svg" caption="What a canonical log line looks like in raw form as it's being emitted."

## Storage (#storage)

After emitting canonical lines, the next step is to put them somewhere useful.
At Stripe we use a two prong approach combining Splunk and Redshift.

### Splunk (#splunk)

Splunk is a powerful shorter-term store that's great for getting fast insight
into online systems. It's great for:

1. Powerful querying syntax that also happens to be quite terse and fast to
   write.
2. Very fast: queries come back quickly, especially if they're scoped down to a
   small slice of the total data volume (i.e. like the last hour worth of logs).
3. Ingested directly from syslog so that data is available almost as soon as
   it's emitted.
4. By tagging every line emitted to Splunk with [request IDs](/request-ids), we
   can easily cross-reference canonical lines with any other part of the raw
   trace that came from a request.
5. Available immediately: ljgs are fed directly into Splunk and can be queried
   in near real-time.

These traits make Splunk ideal for operational work where we might be in the
midst of an incident and need information immediately.

The downside of Splunk is that it's expensive to license and run, and your
retention is generally limited by your total capacity being heavily eaten into
by the sheer amount of raw information being pushed into the system. It's not
an unusual sight to see our operations teams trying to prune the traces of our
highest traffic systems to keep Splunk running under quota [2].

### Redshift (#redshift)

The other system that we use to ingest canonical lines is Redshift. What's
written here applies to any data warehousing system, so feel free to replace
"Redshift" with your software of choice.

Some advantages of a warehouse:

1. Scalability: tables can be arbitrarily large and still queryable without
   trouble.
2. Low cost: especially compared to systems like Splunk, data can be archived
   for extended periods without price becoming an issue. We prune canonical
   lines after 90 days, but they could conceivably be stored for much longer.
3. By importing other non-request data into the warehouse, like say information
   on your users from a core data store, it allows greater levels of deep
   insight by stitching together data generated from different data sources.

We import data by emitting lines over a fast queueing system (NSQ), archiving
batches of it to S3, and then periodically running a `COPY` pointing to the
bucket from Redshift. The implementation here doesn't matter very much, but
it's worth nothing Redshift does best when data is ingested using bulk ETL, and
by extension won't be updated in near real-time like a system ingesting
directly from syslog (e.g. Splunk).

## Examples

### Example: HTTP 500s By Breakage

One of the hazards of any software stack is that unexpected breakages will
happen. For a typical web service, this often takes the form of an exception
raised internally and then converted to an HTTP 500 and returned to the user.

While it's reasonably likely that a web service already has a dashboard in
place for the frequency of 500s that are occurring, what if I wanted, for
example, to drill into the errors from one particular API endpoint. By scoping
canonical log lines down to 500 responses coming from a single API method, I
can easily get that information from Splunk:

!fig src="/assets/canonical-log-lines/error-timechart.png" caption="Error counts for the last week on the \"list events\" endpoint."

> canonical-api-line status=500 api_method=AllEventsMethod earliest=-7d | timechart count

Now say I want to cross-reference that result with some other information that
was emitted into the log trace for any request. By putting my original query
into a Splunk subsearch, and "joining" log traces on a request ID, it's easy.

Continuing from the example above, I want to look for the precise class of
error that was emitted for each breakage, so I join on another specialized
"breakage line" that's emitted for each one:

!fig src="/assets/canonical-log-lines/top-errors.png" caption="The names of the Ruby exception classes emitted for each error, and their relative count."

> [search canonical-api-line status=500 api_method=AllEventsMethod sourcetype=bapi-srv earliest=-7d | fields action_id] BREAKAGE-SPLUNKLINE | stats count by error_class | sort -count limit 10

I can invert this to pull information _out_ of the canonical lines as well.
Here are counts of timeout errors over the last week by API version:

!fig src="/assets/canonical-log-lines/top-api-versions.png" caption="An inverted search. API versions pulled from the canonical log line and fetched by class of error."

> [search breakage-splunkline error_class=Chalk::ODM::OperationTimeout sourcetype=bapi-srv earliest=-7d | fields action_id] canonical-api-line | stats count by stripe_version | sort -count limit 10

### Example: TLS Deprecation

One project that I'm working on right now is helping Stripe merchants [migrate
to TLS 1.2 from older secure protocols][upgrading-tls]. TLS 1.2 will eventually
be required for PCI compliance, so we're trying to identify merchants who are
on TLS 1.0 and 1.1 and given them some warning that an upgrade will be
required.

While responding to support questions on the topic, I realized it would be
useful to be able to quickly reference the key TLS metrics for any given
account. I spent 10 minutes learning how to make dashboard in Splunk, and then
another 10 to create a dashboard that's powered purely be canonical log lines:

!fig src="/assets/canonical-log-lines/tls-requests-by-merchant.png" caption="Splunk dashboard for requests from a merchant by TLS version."

Each panel performs a search for canonical log lines matching a certain
merchant, excludes lines generated by calls from internal systems, then pulls
some metrics and tabulates or draws a plot of the results.

The same thing is also possible from Redshift. Here's a query to ask for any
merchants that have made requests to our API services using pre-1.2 versions of
TLS in the last week [3]:

```
SELECT distinct(merchant_id)
FROM canonical_lines.api
WHERE created > GETDATE() - '7 days'::interval
  AND tls_version < 'TLSv1.2'
ORDER BY 1;
```

We can take it a step further by joining against other information in our
warehouse. To track who's upgraded their TLS implementation and who hasn't,
we've introduced a field on every merchant that keeps track of their minimum
TLS version (`merchants.minimum_tls_version`). To prevent merchants from
regressing to an old TLS version after they've started using a new one, we lock
them into making TLS requests that only use their minimum version or newer.

I'd like to see which merchants flagged into a pre-1.2 TLS version have since
upgraded their integrations to help track our progress towards total
deprecation. By joining my table of canonical lines with one containing
merchant information, I can easily get this from Redshift by asking for all
merchants on old TLS versions who have not made a request using a pre-1.2
version of TLS in the past week:

``` sql
SELECT id
FROM merchants m
WHERE minimum_tls_version < 'TLSv1.2'
  AND NOT EXISTS (
    SELECT 1
    FROM canonical_lines.api
    WHERE merchant_id = m.id
      AND created > GETDATE() - '7 days'::interval
      AND tls_version < 'TLSv1.2'
  )
ORDER BY 1;
```

And as easily as that, I have a data set of all merchants who are probably
candidates to have their `minimum_tls_version` raised.

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
