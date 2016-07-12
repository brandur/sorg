---
hook: How we build minimal, platform deployable, Rack service stubs to take the pain
  out of developing applications that depend on an extensive service-oriented architecture.
location: San Francisco
published_at: 2013-06-03T17:22:11Z
title: SOA and Service Stubs
---

[Service-oriented architecture](http://en.wikipedia.org/wiki/Service-oriented_architecture) (SOA) is a popular technique at a number of software development shops these days, each trying to avoid the age-old problem of the monolithic codebase. This includes us at Heroku, where the underlying kernel that supports the Heroku platform is fundamentally decomposed into a number of backend services, each of which having a well-defined set of responsibilities, and a team to operate it.

Despite us being happily well-invested in this architectural approach, SOA has a number of downsides, one of the foremost being that running a single application in isolation becomes difficult because it relies on so many external services.

A traditional solution for us has been to stub these services at a high-level to avoid making remote calls during development and testing. Here are some real examples from our codebase of what these service stubs looked like:

``` ruby
# Logplex stub
Logplex.stub(:create_token)

# Addons stub
Addons::Client.mock! if ENV["RACK_ENV"] == "test"

# Domain management service stub
Maestro::ResourceHandler.mock!

# Billing management service stub
def detect_handler
  return RealHandler.new if ENV.has_key?("CORE_SHUSHU_URL")
  MockHandler.new
end

# Process management service stub
return ServiceApi::MockHandler.new("https://psmgr.heroku-\#{name}.com") unless psmgr_url

# Resource provisioning
def api_calls_enabled?
  Utils.in_cloud? || RAILS_ENV == 'test' || ENV['API_CALLS_ENABLED'] # OMG HACK
end
response = handler.install(app, user, options) if api_calls_enabled?
```

## Progression (#progression)

Implementations varied as more services were added over time, and generally became more sophisticated as we learned the downsides of particular approaches and iterated on them. This progression eventually led to the _Rack service stubs_ we use today, and which are detailed in the next section.

The starting point was to simply use a stubbing framework for testing to stub out any methods that make a call to a remote service:

``` ruby
Logplex.stub(:create_token)
```

This approach will get tests passing, but couples your tests to the interface of the service handler, and prevents the handler itself from being exercised. What if handlers could be written that knew how to mock themselves?

``` ruby
Addons::Client.mock! if ENV["RACK_ENV"] == "test"
```

This works pretty well for testing, but leaves it up to the developer to find themselves a working staging version of the remote service that won't balk at being consumed by their development installation. This can work given a company-wide culture of well-documented and well-maintained staging environments, but even then, development can be slowed or blocked if a staging installation owned by another team breaks.

A possible next step is to build service handlers that will respond correctly in both development and testing environments:

``` ruby
def detect_handler
  return RealHandler.new if ENV.has_key?("CORE_SHUSHU_URL")
  MockHandler.new
end
```

Now we're getting somewhere! Our mocks should behave reasonably during development and testing, and switch over to the real handler when the code hits staging or production.

While this kind of stub generally works pretty well, it still leaves us with a large disparity between development and production in that two different handlers, and therefore two completely different code paths are run in the two environments. A problem caused by this gap would hopefully be caught in a high-fidelity staging environment before making it to production, but even in staging, debugging is harder and slower compared to a local production copy.

## Rack Service Stubs (#rack)

While re-approaching the application code for our API, we started experimenting with doing away with the variety of stub handlers, and tried replacing them with actual implementations of the foreign stubs written with Rack-compliant libraries. These Rack stubs are designed to implement only the subset of the foreign API required by the calling app, and are greatly simplified to provide the bare minimum of the requirements for a correct response (and do little else).

Here's an example Rack stub written for one of the API's backend services in Sinatra:

``` ruby
class IonStub < Sinatra::Base
  post "/endpoints" do
    status 201
    content_type :json
    MultiJson.encode({
      id:           123,
      cname:        "tokyo-1234.herokussl.com",
      elb_dns_name: "elb016353-1923944129.us-east-1.elb.amazonaws.com",
    })
  end
end
```

Because the stub is a fully functional application in its own right, it becomes immediately useful in both development and testing. A platform that trivializes deployment extends this use to cloud-hosted development and staging environments as well (i.e. you can `git push heroku master` this stub and to make it available for other apps to talk to).

## Testing (#testing)

Use of service stubs in tests is made simple by using [Webmock's](https://github.com/bblimke/webmock) excellent Rack support to intercept requests made to a particular URL and send them off to be processed by the stub. Here are some examples of simple helper methods that we use in the API:

``` ruby
# generic helper for use with any service
def stub_service(uri, stub, &block)
  uri = URI.parse(uri)
  port = uri.port != uri.default_port ? ":\#{uri.port}" : ""
  stub = block ? Sinatra.new(stub, &block) : stub
  stub_request(:any, /^\#{uri.scheme}:\/\/(.*:.*@)?\#{uri.host}\#{port}\/.*$/).
    to_rack(stub)
end

# One-liners specifically for a specific stubs, pointing to configured
# locations of each remote service. A configuration value might look like:
#
#    ADDONS_URL=https://api-user:api-pass@addons.heroku.com
#

def stub_addons
  stub_service(ENV["ADDONS_URL"], AddonsStub, &block)
end

def stub_ion(&block)
  stub_service(ENV["ION_URL"], IonStub, &block)
end
```

Now a stub can be initialized in a test and a remote service call made:

``` ruby
it "should make a call to ion" do
  stub_ion
  endpoint = IonAPI.create_endpoint!
end
```

This is particularly useful for tests that aim to exercise as many levels of application code as possible by stubbing at the level of HTTP calls rather than at a local service library, and ensuring that we're running as much production code as we can.

Error conditions from the remote service can be tested by extending stubs with Sinatra's widely-known DSL for particular test cases:

``` ruby
it "should raise an error on a bad ion response" do
  stub_ion do
    post("/endpoints") { 422 }
  end
  lambda do
    IonAPI.create_endpoint!
  end.should raise_error(IonAPI::Error)
end
```

## Development (#development)

By including a small snippet of conditional run code along with each stub, we ensure that each can be booted as an application in its own right:

``` ruby
class IonStub < Sinatra::Base
  ..
end

if __FILE__ == $0
  $stdout.sync = $stderr.sync = true
  IonStub.run! port: 5100
end
```

Sinatra will boot such a stub by simply invoking its filename:

``` bash
$ ruby test/test_support/service_stubs/ion_stub.rb
>> Listening on 0.0.0.0:5100, CTRL+C to stop
```

Locally exporting an environmental variable ensures that the main app points to a booted stub in its development context:

```
ION_URL=http://localhost:5100
```

With the stub running and the proper configuration in place, the main app can be booted (in this example, the Heroku API):

``` bash
$ bundle exec puma --quiet --threads 8:32 --port 5000 config.ru
listening on addr=0.0.0.0:5000 fd=13
```

Requests to the app will call into the running service stub, allowing us to go a long way towards simulating a working service cloud with minimal setup required. In the example above, we've stubbed the API's infrastructure service, so we can pretend to provision an `ssl:endpoint`:

``` bash
$ export HEROKU_HOST=http://localhost:5000
$ heroku addons:add ssl:endpoint
$ heroku certs:add secure.example.org.pem secure.example.org.key
Adding SSL Endpoint to great-cloud... done
WARNING: ssl_cert provides no domain(s) that are configured for this Heroku app
great-cloud now served by tokyo-1234.herokussl.com
Certificate details:
Common Name(s): alt1.example.org
                alt2.example.org
                secure.example.org

Expires At:     2031-05-05 19:05 UTC
Issuer:         /C=US/ST=California/L=San Francisco/O=Heroku/CN=secure.example.org
Starts At:      2011-05-10 19:05 UTC
Subject:        /C=US/ST=California/L=San Francisco/O=Heroku/CN=secure.example.org
SSL certificate is self signed.
```

### Foreman (#foreman)

The process can be further streamlined by using [Foreman](https://github.com/ddollar/foreman) and adding stubs to the list of processes that should be started:

```
web:     bundle exec puma --quiet --threads 8:32 --port 5000 config.ru

# stubs
ionstub: bundle exec ruby service_stubs/ion_stub.rb
```

Put configuration in a place where Foreman can find it, like a local `.env`:

```
ION_URL=http://localhost:5100
```

With stub processes in place, issuing a `foreman start` results in something like the following:

```
18:18:22 web.1              | listening on addr=0.0.0.0:5000 fd=13
18:18:22 ionstub.1          | == Sinatra/1.3.5 has taken the stage on 5100 for development with backup from WEBrick
```

That convenience further compounds with the more service stubs that are added. Here's a sample of the boot process of the Heroku API:

```
18:20:38 web.1              | listening on addr=0.0.0.0:5000 fd=13
18:20:38 addonsstub.1       | == Sinatra/1.3.5 has taken the stage on 4101 for development with backup from WEBrick
18:20:38 apollostub.1       | == Sinatra/1.3.5 has taken the stage on 4111 for development with backup from WEBrick
18:20:38 eventmanagerstub.1 | == Sinatra/1.3.5 has taken the stage on 4102 for development with backup from WEBrick
18:20:38 ionstub.1          | == Sinatra/1.3.5 has taken the stage on 4103 for development with backup from WEBrick
18:20:38 paymentsstub.1     | == Sinatra/1.3.5 has taken the stage on 4109 for development with backup from WEBrick
18:20:38 psmgrstub.1        | == Sinatra/1.3.5 has taken the stage on 4107 for development with backup from WEBrick
18:20:38 maestrostub.1      | == Sinatra/1.3.5 has taken the stage on 4105 for development with backup from WEBrick
18:20:38 yobukostub.1       | == Sinatra/1.3.5 has taken the stage on 4114 for development with backup from WEBrick
18:20:38 logplexstub.1      | == Sinatra/1.3.5 has taken the stage on 4104 for development with backup from WEBrick
18:20:38 vaultstub.1        | == Sinatra/1.3.5 has taken the stage on 4112 for development with backup from WEBrick
18:20:38 vaultusagestub.1   | == Sinatra/1.3.5 has taken the stage on 4113 for development with backup from WEBrick
18:20:38 zendeskssostub.1   | == Sinatra/1.3.5 has taken the stage on 4115 for development with backup from WEBrick
````

A working example of this project is available at [brandur/service-stub-example](https://github.com/brandur/service-stub-example).
