---
hook: Stubbing distributed services with Sinatra-based services stubs and enabling
  stronger constraints with JSON Schema.
location: San Francisco
published_at: 2014-04-30T16:23:22Z
title: Testing Distributed Services with JSON Schema
---

A while back I wrote about [how we use service stubs](/service-stubs) to make a distributed architecture less painful to work with, and more recently we wrote about [how we're using JSON Schema](https://blog.heroku.com/archives/2014/1/8/json_schema_for_heroku_platform_api) to describe the new Heroku platform API (actually it's a JSON Hyper-Schema, but I'll use the two terms interchangeably in this document).

More recently, we've started to adapt the interfaces of other internal components to be fronted by a JSON Schema definition as well. This document describes a simple technique for augmenting the accuracy of service stubs by taking advantage of the JSON Schemas of other components, allowing errors to be caught more quickly and more easily than would otherwise be possible.

As discussed previously, the pre-Schema stubs are simply small Sinatra-based apps that do a minimal amount of work in order to approximate a correct response for some remote service:

``` ruby
post "/apps" do
  content_type :json
  status 201
  id = SecureRandom.uuid
  JSON.pretty_generate({
    id: id,
    name: "app-#{id}",
  })
end
```

## Towards Stronger Constraints (#strong-constraints)

The basic stubbing technique is pretty effective just by itself because it allows an app's stack to be exercised all the way out to the HTTP calls it's making without a mess of inconsistent stubbing sprinkled throughout the codebase. The data that's returned from the stubs is low fidelity compared to what would be returned by the actual service, but in practice it's not a huge problem when running in isolation or exercising a test suite.

A JSON Schema for the remote service allows this situation to be improved further by strengthening the constraints around these local stubs in such a way that they'll respond to input as if they were a production system. We use [Committee](https://github.com/heroku/committee) to perform this function, a small library that can be mixed into an app as a simple piece of middleware:

```
# will validate input parameters
use Committee::Middleware::RequestValidation,
  schema: File.read("schema.json")

post "/apps" do
  ...
end
```

Given a JSON Schema that describes how that endpoint takes parameters, Committee will respond to requests appropriately. For example, take the small fragment of JSON Schema below that describes how to POST data to an apps resource ([see the full schema for this example](https://github.com/heroku/committee/blob/03347b7e46fa4499aa6f789098039e7b91597086/examples/schema.json)) where we define `name` to be a string of a specific format:

``` json
"definitions": {
  "name": {
    "description": "unique name of app",
    "example": "example",
    "pattern": "^[a-z][a-z0-9-]{3,50}$",
    "readOnly": false,
    "type": [
      "string"
    ]
  },
  ...
},
"links": [
  {
    "description": "Create a new app.",
    "href": "/apps",
    "method": "POST",
    "rel": "create",
    "schema": {
      "properties": {
        "name": {
          "$ref": "#/definitions/app/definitions/name"
        }
      },
      "type": [
        "object"
      ]
    },
    "title": "Create"
  },
  ...
]
```

By booting the stub above and testing it with Curl, we can see that an error is thrown for a request that included `name` as an integer instead of the expected string type:

``` bash
$ curl -i http://localhost:5000/apps -X POST \
  -H "Content-Type: application/json" -d '{"name":123}'

HTTP/1.1 422
Content-Type: application/json
X-Content-Type-Options: nosniff
Server: WEBrick/1.3.1 (Ruby/1.9.3/2012-04-20)
Date: Wed, 30 Apr 2014 03:49:34 GMT
Content-Length: 106
Connection: Keep-Alive

{
  "id": "invalid_params",
  "error": "Invalid type for key \"name\": expected 123 to be [\"string\"]."
}
```

Another nice feature here is that while the JSON Schema acts as a consistent way to describe a set of endpoints, the stub itself is still a Sinatra app, so further constraints that can't be described by a JSON Schema can also be added:

``` ruby
use Committee::Middleware::RequestValidation,
  schema: File.read("schema.json")

post "/apps" do
  content_type :json
  if (8..17).include?(Time.now.hour)
    status 422
    JSON.pretty_generate(
      message: "Can't create apps outside of business hours!"
    )
  end
  ...
end
```

## Symmetry (#symmetry)

Committee and other validation tools like it aren't just for use by stubs! By using Committee in the actual implementation of a component, that component's code around parameter validation can be simplified drastically because it can assume that if a request has made it into its handler, then the parameters of that request are present and of the right type.

This also helps to produce a very nice system symmetry in that components are validating their incoming requests with exactly the same interface definition as other components are using to validate their requests to _that_ component. By using this technique, we can ensure that far more errors are caught quickly with simple integration tests before wasting time tracking them down in a more realistic system.

The full example above [is available on GitHub](https://github.com/heroku/committee/tree/master/examples).
