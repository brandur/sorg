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

## The Canonical Line (#canonical-line)

Canonical log lines are compilation of important information that's aggregate
across the lifetime of a single request (or background job run, etc.) and
emitted all onto a single line of logs. Here are some examples of information
that we might like to include:

* Basic request information like HTTP verb, host, path, source IP, and user
  agent.
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
* Rate limiting information wuch as whether rate limiting occurred, what their
  total limit is, and how much of it is remaining.

In [logfmt](/logfmt) style, this might look something like (note that I've
added newlines and comments for clarity, but this would be all one big long
line in a real system):

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

Middleware makes a pretty good home for implementation, where it generally
lives close to the top of the middleware stack and injects an object through
context (i.e. `env` in Rack) which then has its fields populated to the maximum
possible extent by downstream components in the request stack.

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

## Logging Tiers (#log-levels)

* L2: Standard log traces (probably info-level in production).
* L1: Canonical log lines.
* L0: Exceptions.

## Splunk Tricks (#splunk)

(Note this also uses the concept of [request IDs](/request-ids) that get tagged
onto every line of a single request's log trace.)

The inverse also works.

## Warehousing (#warehousing)

Cheap to store long-term.

We emit over NSQ for permanent storage in S3. A scheduled task goes through and
puts S3 segments into Redshift. Redshift allows easy analytical work. Another
scheduled task prunes the horizon of the canonical lines so that they're
removed after 90 days to keep
