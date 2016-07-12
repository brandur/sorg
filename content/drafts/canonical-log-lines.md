---
hook: Using canonical log lines for powerful and succinct introspection into an online
  system.
location: San Francisco
published_at: 2016-04-14T17:03:10Z
title: Canonical Log Lines
---



## Noisy Logging & Operational Introspection (#noisy)

Very useful when debugging.

Although great for debugging certain problems, makes day-to-day checks a little more challenging.

Also, makes running certain types of analytical queries require some extreme Splunk-fu.

## The Canonical Line (#canonical-line)

Some interesting things about a request that can be logged:

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

The Ruby code for that might look like:

``` ruby
log.info "CANONICAL-LOG-LINE",

  # basic request information
  request.content_type: request.content_type,
  request_ip:           extract_ip(request),
  request_method:       request.method,
  request_path:         request.path_info,
  request_user_agent:   request.user_agent,

  # request IDs
  request_id:        request_id,
  request_parent_id: request_parent_id,

  # response
  response_content_type: content_type,
  response_status:       status,

  # errors
  error_id:      error.id,
  error_message: error.message,

  # user information
  user_email: user.email,
  user_id:    user.id,

  # authentication
  auth_oauth_app_id: auth.oauth_app_id,
  auth_oauth_scope:  auth.oauth_scope,

  # app information
  app_git_head: config.app_git_head,
  app_name:     config.app_name,
  app_version:  config.app_version,

  # timing information
  timing_database_total: timing.database_total,
  timing_request_total:  timing.request_total,

  # rate limiting information
  rate_limit_enforced:  rate_limit.enforced?,
  rate_limit_limit:     rate_limit.limit,
  rate_limit_remaining: rate_limit.remaining,
  rate_limit_reset:     rate_limit.reset
```

Middleware makes a pretty good home for this pattern, where a log line is
emitted after calling into `app.call(env)`.

Need to pass information out through the stack.

## Warehousing (#warehousing)

Cheap to store long-term.

We emit over NSQ for permanent storage in S3. A scheduled task goes through and
puts S3 segments into Redshift. Redshift allows easy analytical work. Another
scheduled task prunes the horizon of the canonical lines so that they're
removed after 90 days to keep
