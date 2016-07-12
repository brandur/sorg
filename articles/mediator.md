---
hook: Interactors by a different name.
location: San Francisco
published_at: 2014-03-11T15:25:07Z
title: The Mediator Pattern
---

Grouper published a post last week [about how they use interactors](http://eng.joingrouper.com/blog/2014/03/03/rails-the-missing-parts-interactors) in their Rails app to help keep their ActiveRecord models as lean as possible. Somewhat amusingly, while doing a major refactor of the Heroku API, we'd independently arrived at a nearly identical pattern after learning the hard way that callbacks and large models are the inviting pool whose frothy water conceals treacherous rocks.

The main difference was in appellation: we called the resulting PORO's "mediators", a [design pattern](http://en.wikipedia.org/wiki/Mediator_pattern) that defines how a set of objects interact. I'm not one to quarrel over naming, but I'll use the term "mediator" throughout this article because that's how I'm used to thinking about this pattern.

The intent of this article is to build on what Grouper wrote by talking about some other nice patterns that we've built around the use of mediators/interactors.

## Lean Endpoints (#lean-endpoints)

One goal of our usage of mediators is to consolidate all the business logic that might otherwise have to reside in a combination of an API endpoint's body and methods on models. Ideally what remains in the endpoint should be a set of request checks like authentication, ACL, and parameters; a single call down to a mediator; and response logic like serialization and status.

Here's a small (and slightly beautified) excerpt from the API endpoint for creating an SSL Endpoint:

``` ruby
module API::Endpoints::APIV3
  class SSLEndpoints < Base
    ...

    namespace "/apps/:id/ssl-endpoints" do
      before do
        authorized!
        @ap = get_any_app!
        check_permissions!(:manage_domains, @ap)
        check_params!
      end

      post do
        @endpoint = API::Mediators::SSLEndpoints::Creator.run(
          auditor: self,
          app:     @ap,
          key:     v3_body_params[:private_key],
          pem:     v3_body_params[:certificate_chain],
          user:    current_user
        )
        respond serialize(@endpoint), status: 201
      end
    end

    ...
  end
end
```

This pattern produces a convention that helps keep important logic out of endpoints and in the more readily accessible mediator classes. It also keeps unit tests for the endpoints focused on what those endpoints are responsible for: authentication, parameter and permission checks, serialization, and the like. For success cases, we can mock out the mediator's call and response and focus on doing more comprehensive tests on the business logic in the mediator's own unit tests. The entire stack still gets exercised at the integration test level, but we don't have to get into the same level of exhaustive testing there.

A mocked endpoint unit test might look like the following (note that the specs are using the [rr](https://github.com/rr/rr) mocking syntax):

``` ruby
# endpoint unit tests
describe API::Endpoints::APIV3::SSLEndpoints do
  ...

  describe "POST /apps/:id/ssl-endpoints" do
    it "calls into the mediator" do
      mock(API::Endpoints::APIV3::SSLEndpoints).run(hash_including({
        app:  @app,
        key:  "my-private-key",
        pem:  "my-pem",
        user: @user,
      })
      authorize "", @user.api_key
      header "Content-Type", "application/json"
      post "/apps/#{@app.name}/ssl-endpoints", MultiJson.encode({
        private_key:       "my-private-key",
        certificate_chain: "my-pem",
      })
    end
  end

  ...
end
```

The mediator unit tests will go into far greater detail and look something like this:

``` ruby
# mediator units tests
describe API::Mediators::SSLEndpoints::Creator do
  ...

  it "produces an SSL Endpoint" do
    endpoint = run
    assert_kind_of API::Models::SSLEndpoint, endpoint
  end

  it "makes a call to the Ion API to create the endpoint" do
    mock(IonAPI).create_endpoint
    run
  end

  ...

  private

  def run(options = {})
    API::Mediators::SSLEndpoints::Creator.run({
      app:  @app,
      key:  @key_contents,
      pem:  @pem_contents,
      user: @app.owner,
    }.merge(options))
  end
end
```

## Lean Jobs (#lean-jobs)

Much in the same way that mediators keep our endpoints lean, they do the same for our async jobs. By encapsulating all business logic into a mediator, we leave jobs to focus only one two things:

1. **Model materialization:** Async jobs are passed through some kind of backchannel like a database table or a redis queue, and have to marshaled on the other side. It's up to the job to figure out how to find and instantiate the models that it needs to inject into its mediator. This logic may change job to job: if we have a job to create a logging channel for an app, but that app has already been deleted by the time it runs, then we should fall through the job without an error; but if we have an async job to a destroy an app, and its record is no longer avaiable, then something unexpected happen and we should raise an error.
2. **Error handling:** A job's second responsibility is to rescue errors and figure out what to do with them. If we're trying to provision an SSL Endpoint and got a connection error to our downstream endpoints service, then we might want to send the job back into the work queue; but if something like a configuration error occurred, we might want to notify our error service and fail the job permanently.

Let's look at what an async job might look like for the hypothetical SSL Endpoint creation mediator from above:

``` ruby
module API::Jobs::SSLEndpoints
  class Creator < API::Jobs::Base
    def initialize(args = {})
      super
      require_args!(
        :app_id,
        :key,
        :pem,
        :user_id
        )
    end

    def call
      # If the app is no longer present, then it's been deleted since the job
      # was dequeued; succeed without doing anything.
      return unless @app = App.find_by_id(args[:app_id])

      # If the user is no longer present, then they may have deleted their
      # account isince the job was dequeued; succeed without doing anything.
      return unless @user = User.find_by_id(args[:user_id])

      API::Mediators::SSLEndpoints::Creator.run(
        auditor: self,
        app:     @app,
        key:     args[:key],
        pem:     args[:pem],
        user:    @user
      )

    # Something is wrong which will prevent the job from ever succeeding. Fail
    # the job permamently and notify operators of the error.
    rescue API::Error::ConfigurationMissing => e
      raise API::Error::JobFailed.new(e)

    # Something has caused a temporary disruption in service. Queue the job
    # again for retry.
    rescue Excon::Errors::Error
      raise API::Error::JobRetry
    end
  end
end
```

(Note that the above is a simplified example. If you were going to send a sensitive secret like an SSL key through an insecure channel, we'd want to encrypt it.)

## Strong Preconditions (#strong-preconditions)

From within any mediator, we assume that a few preconditions have already been met:

* **Parameters:** All parameters are present in their expected form.
* **Models:** Rather than passing around abstract identifiers, parameters are materialized models so that no look-up logic needs to be included.
* **Security:** Security checks like authentication and access control have already been made.

Making these strong assumptions has a number of advantages:

* The complexity of the resulting code is reduced dramatically. We don't have to spend LOCs checking that objects are present or whether they're in their expected form (almost like working in a strongly typed language!).
* It eases testing as the boilerplate for checking parameter validation and the like can be consolidated elsewhere.
* Allows mediators to be called more easily from outside their normal context like from a debugging/operations console session.

## Mediators All the Way Down (#nesting-mediators)

One way to think about mediators is that they encapsulate a discrete piece of work that involves interaction between a set of objects; a piece of work that otherwise might have ended up in an unwieldy method on a model. Because units of work are often composable, just like those model methods would have been, it's a common pattern for mediators to make calls to other mediators.

Here's a small example of an app mediator that also deprovisions the app's installed add-ons:

``` ruby
module API::Mediators::Apps
  class Destroy < API::Mediators::Base
    ...

    def destroy_addons
      @app.addons.each do |addon|
        API::Mediators::Addons::Destroyer.run(
          addon:   addon,
          auditor: @auditor,
        )
      end
    end

    ...
  end
```

Of course it's important that your mediators have a clear call hierarchy so as not to develop any circular dependencies, but as long as developers don't get too overzealous with mediator creation, this is pretty safe.

## Patterns Through Convention (#convention)

While establishing mediators as the default unit of work, it's also a convenient time to start building other useful conventions into them. For example, we build in an auditing pattern so that we're still able to produce a trail of audit events even the mediator's work is performed from unexpected places like a console:

``` ruby
module API::Mediators::Apps
  class Destroy < API::Mediators::Base
    ...

    def call
      audit do
        ...
      end
    end

    private

    def audit(&block)
      @auditor.audit("destroy-app", target_app: @app, &block)
    end
  end
end
```

Another example of an established convention is to try and build out call bodies composed of a series of one-line calls to helpers that produces a very readable set of operations that any given mediator will perform:

``` ruby
module API::Mediators::Apps
  class Destroy < API::Mediators::Base
    ...

    def call
      audit do
        App.transaction do
          destroy_addons
          destroy_domains
          destroy_ssl_endpoints
          close_payment_method_history
          close_resource_histories
          delete_logplex_channel
          @app.destroy
        end
      end
    end

    ...
  end
```

A few years into working with the mediator pattern now, and I'd never go back. Although mediator calls are a little more verbose than they might have been as a model methods, they've allowed us to lean out the majority of our models to contain only basics like assocations, validations, and accessors. This has the added advantage of leaving us more decoupled from our ORM (ActiveRecord in this case) than ever before.

Eliminating callbacks has also been a hugely important step forward in that it reduces production incidents caused by running innocent-looking code that results in major side effects, and leaves us with more transparent test code.
