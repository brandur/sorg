---
hook: A simple pattern for tracing requests across a service-oriented architecture
  by injecting a UUID into the events that they produce.
location: San Francisco
published_at: 2013-06-02T17:21:43Z
title: Tracing Request IDs
---

Log into Heroku's [Dashboard](https://dashboard.heroku.com) and you'll hit three different components that work together to usher you in and show you an app list, run a `git push heroku master` and that number is more like six. These kinds of service-oriented patterns produce all kinds of logistical benefits, but debugging such a system at production scale can get messy.

It's key to have powerful techniques at your disposable to gain introspection and track down bugs in your production system. A simple one that's tremendously useful even on its own, is the use of request IDs to trace requests as they thread themselves through a set of composed components.

## Request IDs (#request-ids)

A very lightweight alternative to something like [Twitter's Zipkin](http://engineering.twitter.com/2012/06/distributed-systems-tracing-with-zipkin.html), and based on the same ideas as the [troubleshooting technique that Amazon uses for route 53](http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/ResponseHeader_RequestID.html), request IDs are a way of grouping all the information associated with a given request, even as that request makes its way across a distributed architecture. The benefits are two-fold:

* Provides a tagging mechanism for events that are produced, so that a full report of what occurred and timing in every component touched can be generated.
* Exposes an identifier to users, both internal and external, which can be used to track down specific issues that they're running into.

In practice, the request ID is a UUID that's generated at the beginning of a request and stored for its duration. Here's a simple Rack middleware that does this job:

``` ruby
class Middleware::Instruments
  def initialize(app)
    @app = app
  end

  def call(env)
    env["REQUEST_ID"] = SecureRandom.uuid
    @app.call(env)
  end
end
```

Logging events in the main app include this identifier in their associated data. A sample helper for Sinatra might look like:

``` ruby
def log(action, data={})
  data.merge!(request_id: request.env["REQUEST_ID"])
  ...
end
```

And with everything done right, a request emits a nice logging stream with each event tagged with the generated request ID:

```
app=api authenticate elapsed=0.001 request_id=9d5ccdbe-6a5c-4da7-8762-8fb627a020a4
app=api rate_limit elapsed=0.001 request_id=9d5ccdbe-6a5c-4da7-8762-8fb627a020a4
app=api provision_token elapsed=0.003 request_id=9d5ccdbe-6a5c-4da7-8762-8fb627a020a4
app=api serialize elapsed=0.000 request_id=9d5ccdbe-6a5c-4da7-8762-8fb627a020a4
app=api response status=201 elapsed=0.005 request_id=9d5ccdbe-6a5c-4da7-8762-8fb627a020a4
```

Our apps are all configured to drain their log streams to Splunk, which provides a centralized location that allows us to query for all information associated with a given request ID:

```
9d5ccdbe-6a5c-4da7-8762-8fb627a020a4
```

## Heroku's Request IDs (#heroku)

Heroku's routing layer can [generate a request ID](https://devcenter.heroku.com/articles/http-request-id) automatically, which allows platform-generated logging events to be tagged in as well. Rather than generating them yourself, these IDs can be accessed through an incoming header:

``` ruby
def log(action, data={})
  data.merge!(request_id: request.env["HTTP_HEROKU_REQUEST_ID"])
  ...
end
```

## Composing Request IDs (#composition)

Request IDs provide a convenient mechanism for digging into a single request for any given app, but so far they're not much help when it comes to a number of composed apps that are constantly making calls to each other.

We take the concept a step further by having apps that make calls to other apps inject their own request ID via a request header.

``` ruby
api = Excon.new("https://api.heroku.com", headers: {
  "Request-ID" => request.env["REQUEST_ID"]
})
api.post("/oauth/tokens", expects: 201)
```

The callee in turn accepts a request ID, and if it looks like a valid identifier, tags all its requests with the given request ID _as well as_ one that it generates itself. This way we can make sure that a request across many apps can be tracked as a group, but each app always has a way of tracking every one of its requests invidually.

``` ruby
def call(env)
  env["REQUEST_ID"] = SecureRandom.uuid
  if env["HTTP_REQUEST_ID"] =~ UUID_PATTERN
    env["REQUEST_ID"] += "," + env["HTTP_REQUEST_ID"]
  end
  @app.call(env)
end
```

The event stream emitted by the composed apps is now tagged based on all generated request IDs:

```
app=id session_check elapsed=0.000 request_id=4edef22b...
app=api authenticate elapsed=0.001 request_id=9d5ccdbe...,4edef22b...
app=api rate_limit elapsed=0.001 request_id=9d5ccdbe...,4edef22b...
app=api provision_token elapsed=0.003 request_id=9d5ccdbe...,4edef22b...
app=api serialize elapsed=0.000 request_id=9d5ccdbe...,4edef22b...
app=api response status=201 elapsed=0.005 request_id=9d5ccdbe...,4edef22b...
app=id response status=200 elapsed=0.010 request_id=4edef22b...
```

A Splunk query based on the top-level request ID will yield logging events from all composed apps. Note that although we use Splunk here, alternatives like Papertrail will do the same job.

<div class="attachment"><img src="/assets/request-ids/splunk-search.png"></div>

## Tweaks (#tweaks)

### Inject Any Number of Request IDs (#multiple)

A minor modification to the middleware pattern above will allow any number of request IDs to be injected into a given app, so that a request can be traced across three or more composed services.

``` ruby
def call(env)
  env["REQUEST_ID"] = SecureRandom.uuid
  if env["HTTP_REQUEST_ID"]
    request_ids = env["HTTP_REQUEST_ID"].split(",").
      select { |id| id =~ UUID_PATTERN }
    env["REQUEST_ID"] = (env["REQUEST_ID"] + request_ids).join(",")
  end
  @app.call(env)
end
```

### Respond with Request ID (#response)

The request ID can be returned as a response header to enable easier identification and subsequent debugging of any given request:

``` ruby
def call(env)
  request_id = SecureRandom.uuid
  ...
  status, headers, response = @app.call(env)
  headers["Request-ID"] = request_id
  [status, headers, response]
end
```

```
curl -i https://api.example.com/hello
...
Request-ID: 9d5ccdbe-6a5c-4da7-8762-8fb627a020a4
...
```

Heroku's new [V3 platform API](https://devcenter.heroku.com/articles/platform-api-reference#request-id) includes a request ID in the respones with every request.

### Storing Request ID in a Request Store (#storage)

In a larger application, producing logs from a context-sensitive method like a Sinatra helper may be architecturally difficult. In cases like this, a thread-safe request store pattern can be used instead.

``` ruby
# request store that keys a hash to the current thread
module RequestStore
  def self.store
    Thread.current[:request_store] ||= {}
  end
end

# middleware that initializes a request store and and adds a request ID to it
class Middleware::Instruments
  ...

  def call(env)
    RequestStore.store.clear
    RequestStore.store[:request_id] = SecureRandom.uuid
    @app.call(env)
  end
end

# class method that can extract a request ID and tag logging events with it
module Log
  def self.log(action, data={})
    data.merge!(request_id: RequestStore.store[:request_id])
    ...
  end
end
```
