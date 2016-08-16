---
title: Canonical Log Lines
hook: Using canonical log lines for powerful and succinct introspection into an online
  system.
location: San Francisco
published_at: 2016-07-20T14:31:33Z
---

_Canonical log lines_ are a lightweight pattern used within Stripe for improved
service observability. Their greatest strength is that 

If we look at a standard production system, we could say that it emits multiple
"tiers" of analytical information:

1. Metrics (i.e. emitted to services like statsd, Librato, Datadog, etc.).
2. Log traces.

While metrics provide very fast feedback on specific and commonly-used system
measurements, they're not very good at allowing information to be queried
arbitrarily and ad-hoc. While perfect for answering a question that's known
ahead of time like, _"how many requests per second is my stack doing?"_,
they're less useful for questions that I think up on the spot and need
immediate answers for like, _"how many requests per second is userXYZ making to
my stack via TLS 1.0?"_

Log traces can answer the latter question, but often make such analysis
difficult because they tend to be noisy and have information is spread out
across many lines, which can be difficult to correlate even with sophisticated
logging systems like Splunk.

Canonical log lines are "middle tier" of analytics that help to bridge the gap
between metrics and raw traces:

1. Metrics (i.e. emitted to services like statsd, Librato, Datadog, etc.).
2. **Canonical log lines.**
3. Log traces.

!fig src="/assets/canonical-log-lines/observability-tiers.svg" caption="The tiers of observability, showing the trade-off between query flexibility and ease of reference."

## What Is It? (#what-is-it)

The canonical log line's implementation is dead simple: it's one big log line
(probably in [logfmt](/logfmt)\) style that is emitted at the end of a request
[1] and which contains a set of information tuples describing its aspects.

!fig src="/assets/canonical-log-lines/canonical-log-lines.svg" caption="Canonical log lines being emitted (and ingested) for reach request."

Some examples of the type o information that a canonical line might include:

* Basic request vitals like HTTP verb, host, path, source IP, and user agent.
* Any [request IDs](/request-ids) that the request may have been tagged with.
* Response information like status and content type.
* Any error information (if the request errored) like error ID and message.
* Authentication information like the ID of the API key used, or the OAuth
  application and scope in use (if applicable).
* Information on the authentication user like their human-friendly label (say
  an email) or internal identifier for quick and stable reference.
* General information about the running app like its name, HEAD Git revision,
  and current release.
* Aggregate timing information like the total duration of the request, or the
  total amount of time spent in database queries.
* Rate limiting information: whether rate limiting occurred, what a user's
  total limit is, and how much of it is remaining.

In [logfmt](/logfmt) style, this might look like (note that I've added newlines
and comments for clarity, but this would be all one big long line in a real
system):

```
canonical-api-line

  # service information
  service=api release=v752 git_head=6bb8ec015a21ff7f6c3f34f6fe99357352692a72

  # request identification
  request_id=55f10e07-ec6c-486d-8131-80f846fbe465

  # request information
  request_content_type=application/json request_ip=1.2.3.4
  request_method=GET request_path=/users request_user_agent="Curl 1.2.3"

  # response information
  response_content_type=application/json response_status=200
  error_id= error_message=

  # authentication information
  user_email=nameless@example.comkuser_id=1234
  auth_oauth_app_id=2345 auth_oauth_scope=identity,users

  # timing
  timing_request_total=0.099 timing_database_total=0.085

  # rate limiting
  rate_limiting_enforced=false rate_limiting_limit=100 rate_limit_remaining=99
```

## Implementation (#implementation)

Middleware makes a good home for implementation. This middleware will generally
live close to the top of the stack and inject an object through context (i.e.
`env` in Rack) which is populated by downstream components before being
finalized and emitted by the logging middleware.

For example, the logging middleware might inject an object with a `request_id`
field which will then be populated downstream by a "request ID" middleware
after it extracts one from an incoming request's headers.

An example of a basic implementation:

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

## Storage (#storage)

After emitting canonical lines, but the next step is to make use of them by
ingesting them into a system that allows them to be queried in some meaningful
way.

We use a two prong approach where the lines are pushed into:

1. **Splunk**, where they're available almost instantly and can be queried in a
   very free-form way.
2. **Redshift**, where they're stored for much longer and can be more easily
   cross-referenced with other warehouse data.

### Splunk (#splunk)

Splunk is a powerful shorter-term store that's great for getting fast insight
into online systems. It's great for:

1. Powerful querying syntax that also happens to be quite terse and fast to
   write.
2. Very fast. Queries come back quickly, especially if they're scoped down to a
   subset (i.e. like the last hour worth of logs).
3. Ingested directly from syslog so that data is available almost as soon as
   it's emitted.
4. By tagging every line emitted to Splunk with [request IDs](/request-ids), we
   can easily cross-reference canonical lines with any other part of the raw
   trace that came from a request.

All these characteristics make Splunk ideal for operational work where we might
be in the midst of an incident and need information immediately.

The downside of Splunk is that it's very expensive to license and run, and your
retention is generally limited by your total capacity being heavily eaten into
by the sheer amount of raw information being pushed into the system. It's not
an unusual sight to see our operations teams trying to prune the traces of our
highest traffic systems to keep Splunk running under quota [2].

#### Example: 500s By Breakage

Tricks. The inverse also works.

### Redshift (#redshift)

The other system that we use to ingest canonical lines is Redshift. And
although we've found Redshift to be a good technology choice, what I really
mean to talk about is some kind of data warehouse, so feel free to pretend that
this section is talking about your warehousing system of choice.

Some advantages to using a warehouse are:

1. It's scalable. Tables can be arbitrarily large and still queryable without
   trouble.
2. It's cheap. Especially compared to systems like Splunk, data can be archived
   for extended periods without price becoming an issue. We've started pruning
   canonical lines after 90 days, but they could conceivably be stored for much
   longer.
3. By importing other non-request data into the warehouse, like say information
   on your users from a core data store, it allows greater levels of deep
   insight by stitching together data generated from different data sources.

We get data into Redshift by emitting lines over a fast queueing system (NSQ),
archiving batches of it to S3, and then periodically running a `COPY` pointing
to the bucket from Redshift. The implementation here doesn't matter very much,
but it's worth nothing Redshift does best when data is ingested using bulk ETL,
and by extension won't be updated in near real-time like a system ingesting
directly from syslog (e.g. Splunk).

#### Example: TLS Deprecation

!fig src="/assets/canonical-log-lines/tls-requests-by-merchant.png" caption="Splunk dashboard for requests from a merchant by TLS version."

One project that I'm working on right now is helping Stripe merchants who are
using old versions of TLS (i.e 1.0 and 1.1) [upgrade their
integrations][upgrading-tls] before hitting blackout deadlines. Using canonical
lines, I can easily ask Redshift for any merchants that have made requests to
our API services using pre-1.2 versions of TLS in the last week [3]:

```
SELECT distinct(merchant_id)
FROM canonical_lines.api
WHERE created > GETDATE() - '7 days'::interval
  AND tls_version < 'TLSv1.2'
ORDER BY 1;
```

That's pretty handy already, but we can take it a step further. To track who's
upgraded and who hasn't, we've introduced a field on every merchant that keeps
track of their minimum TLS version (`merchants.minimum_tls_version`). To
prevent merchants from regressing to an old TLS version after they've started
using a new one, we lock them into making TLS requests that only use their
minimum version or newer.

I'd like to see which merchants flagged into a pre-1.2 TLS version have since
upgraded their integrations recently so that I can track our progress towards
total deprecation. By joining my table of canonical lines with one containing
merchant information, I can this from Redshift easily by asking for all
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

And thus get a set of all merchants eligible to have their TLS floor updated.

## Summary

We've shown that canonical lines can be a powerful tool for fast and flexible
operational introspection when using them with an appropriate data store. Maybe
their greatest strength is how after putting a few basic ingestion pipelines in
place, they can easily be used to emit information from any language in any
system to produce an easy way to get succinct request information across an
entire internal stack.

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
