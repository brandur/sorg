---
hook: When building an app against a web API, do you pull in their SDK or just make
  raw HTTP calls? Here are a few reasons that I avoid SDKs when I can.
location: San Francisco
published_at: 2014-02-03T14:46:52Z
title: Why I Don't Want Your SDK in Production
hn_link: https://news.ycombinator.com/item?id=7177887
---

Recently the topic of whether it was better practice to use an SDK or basic HTTP calls when interacting with the API of a foreign service came up on [Programmable Web](http://blog.programmableweb.com/2013/10/04/when-sdks-are-better-than-apis/) and on [Traffic and Weather](http://trafficandweather.io/posts/2013/12/27/episode-20-im-going-to-punch-a-wall). The question mostly depends on the premise that many services provide modern RESTful APIs whose endpoints do a good job of encapsuling basic actions that users are interested in, and very little specialized logic needs to be implemented client-side, keeping even direct HTTP calls relatively simple. The trade-off then ends up being taking on an extra dependency in return for an abstraction layer in the language of your choice.

Heroku's internal Ruby culture is extensive enough that over the years, as a group was shipping a new service, it was fairly common to also ship a client gem to help talk to it. Those gems would act as lightweight SDKs for the convenience of developers, and get bundled into any projects that needed to interact with the services that they were abstracting. Although shipping these was a very thoughtful and appreciated good will gesture, I largely started stripping them out of apps that I was working on after observing the side effects of their use.

Here's why I don't want your SDK in production.

## Instrumentation (#instrumentation)

We end up investigating enough tricky problems that it's pretty important for anything that happens inside our apps to produce detailed log trails that will later empower us to analyze exactly what went on. When an error occurs, dumping absolutely everything about it might save you from having to try and expensively reproduce it by hand later on. These traces contain standard log information like the the request's resulting response code and elapsed time, but should ideally also include app-specific information like the current [request ID](/request-ids) and follow the same format conventions that are used elsewhere in the app.

We could wrap any SDKs to include the extra logging, but by making our own HTTP calls, we can [build a single Excon instrumentor](https://github.com/geemus/excon#instrumentation) and re-use it for every service that we call.

## Metrics (#metrics)

Following the same idea as logging, we also want to emit metrics around any calls to foreign services so that we can track the quality of their operation. What's the average service time? How often does it respond with a 503? Does it ever return internal server errors? If the service is critical enough to the operation of our own app, we might even want to put alarms in place to help us keep an eye on it. Or better yet, we could have those alarms trigger for the team who manages that service in case they're not generating great metrics themselves.

## Performance and Persistent Connections (#performance)

The performance of calls to foreign services that are in hot paths is concerning enough that we might want to try to optimize them by keeping pools of persistent connections around.

Does an SDK handle this? Maybe. Does it handle it correctly for any given app's concurrency model? Again, maybe; every SDK has to be examined on a case-by-case basis. We can bypass the uncertainly completely by standardizing on a common pattern for connection re-use against all services that we interact with.

## Error Handling, Edge Cases, and Idempotency (#error-handling)

There's huge variability in the way that SDKs handle errors and other types of less common edge cases. I've seen everything from allowing the exceptions produced by the SDK's internal HTTP library bubble back up to our code to swallowing problems completely and passing invalid data back to us. Even in the best case scenario where an SDK has identified and documented every failure scenario, we still have to consider every error and figure out what to with it. By making basic HTTP calls, we can re-use patterns across services. For example, in most cases when we get a 503 back, we'll bubble that 503 back up to the consumer of our app.

Retries are also worth considering here. If it look like we just hit a basic network problem, and we know of an endpoint to be idempotent, we might want to retry the call a few times. An SDK could do this too, but we can't know its exact behavior without digging into it, and even then it might not be doing the right thing.

## The Grep Test (#grep)

Especially when connection problems and errors bubble up, it's often useful to be able to identify what segment of code was trying to make a call to some HTTP endpoint. When working with basic HTTP calls, working that out is [one grep away](http://jamie-wong.com/2013/07/12/grep-test/), but with SDKs this often has to be reasoned out by comparing the host in a URL to the name of a library.

You can get around this by just vendoring in all your dependencies, but that's terrible.

## Just Freedom Patch! (#freedom-patch)

It's true that the points above are still possible with SDKs be it through designing pluggable SDKs, monkey patching, or building sophisticated wrappers. The problem is that these SDKs are being shipped from by different companies and different people and will have wildly different conventions and capabilities throughout; all of which need to be examined and learned on a case-by-case basis, and the options above probably won't represent a big time saving compared to just wrapping the HTTP calls yourself and re-using the patterns that you already have.

In summary, thank-you for shipping me your SDK &mdash; I appreciate the gesture immeasurably &mdash; but I'd much rather use your well-designed web API.

## Addendum (#addendum)

_(Added February 5th, 2014)_

Judging by some of the feedback I've received from this article, it seems that I've somewhat miscommunicated my position on SDKs. The key to understand this opinion piece is the _in production_ part of the title, meaning that for my own high traffic service I'd like to be as close to the wire as possible and have the option of opting out of SDKs (and a great API makes that easy).

I certainly acknowledge that SDKs are useful for many purposes, including getting users bootstrapped on a service as quickly as possible and helping to tease bugs out of your API as you spend time building the SDK. I wouldn't ship an API without them.
