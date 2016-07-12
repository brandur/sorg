---
hook: When the use of an anti-pattern is considered beneficial.
location: San Francisco
published_at: 2014-02-06T15:31:37Z
title: Healthy Anti-patterns
---

In [Tracing Request IDs](/request-ids), I briefly talked about the possibility of making a request ID easily available from anywhere in an app through a pattern called the [_request store_](https://github.com/steveklabnik/request_store). It's a simple construct that stores data into Ruby's thread-local context:

``` ruby
# request store that keys a hash to the current thread
module RequestStore
  def self.store
    Thread.current[:request_store] ||= {}
  end
end
```

Middleware is then used in an app which makes sure that all context that was added to the store is cleared between requests:

``` ruby
class Middleware::RequestStore
  ...

  def call(env)
    ::RequestStore.store.clear
    @app.call(env)
  end
end
```

I usually take it a bit further in a larger application, where it's been my habit to extend the original pattern so that we inventory exactly what it's supposed to contain, making it more difficult to accidentally create opaque dependencies by mixing data in randomly:

``` ruby
module RequestStore
  def log_context ; store[:log_context] ; end
  def request_id  ; store[:request_id]  ; end

  def log_context=(val) ; store[:log_context] = val ; end
  def request_id =(val) ; store[:request_id]  = val ; end

  private

  def self.store
    Thread.current[:request_store] ||= {}
  end
end
```

## The Anti-pattern (#the-antipattern)

Much like the infamous [singleton pattern](http://en.wikipedia.org/wiki/Singleton_pattern), the request store introduces global state into its application, which in turn makes it more difficult to reason about the dependencies of any given piece of code. Global state can have other side effects too, like making testing more difficult; globals that initialize themselves implicitly can be hard to set without a great stubbing framework, and will keep their value across multiple test cases, which is surprising behavior for anyone not expecting it.

This sort of technique is slightly less controversial in the world of dynamic languages (where you often have something like a [GIL](http://en.wikipedia.org/wiki/Global_Interpreter_Lock) to save you from race conditions across threads) , but I think it's safe to say that my highly pattern-oriented colleagues back in the enterprise world would have chastised me for considering the use of global state of any kind. Instead, they'd strongly prefer the use of a dependency injection framework to make certain information accessible from anywhere in an app.

Despite all this, from an engineering perspective the side effects of using the request store over time have been minimal. By staying vigilant in making sure that it doesn't creep beyond its originally intended use, the request store becomes a very convenient way to store a few pieces of global state that would otherwise be very awkward to access. We keep its use in check by coming to consensus on what can get added to it through discussion in pull requests.

The request store isn't an isolated case either. Projects like Rails and Sinatra have been using global patterns in places like [managing their database connection](https://github.com/rails/rails/blob/4-0-stable/activerecord/lib/active_record/core.rb#L86-L88) or [delegating DSL methods from the main module](https://github.com/sinatra/sinatra/blob/184fe58ca5879d04fce82fcb190c10f72e1f63bc/lib/sinatra/base.rb#L1988) for as long as they've existed. These uses may have caused some grief for somebody over the years, but lasting as long as they have is a testament to their success at least on a practical level.

As long as anti-patterns can continue to show positive productivity results and to cause minimal harm, I'll keep using them.
