---
title: Soft Assertions and the Search Six 9s
published_at: 2016-11-26T04:31:32Z
hook: Detect problems and inconsistencies in production without affecting
  users.
location: San Francisco
---

We measure the availability of the API by how many requests
to it fail due to an internal problem. These blips of
downtime can have a number of different causes; for
example, a bug that was accidentally deployed, timeouts
caused by an overloaded database shard, or the untimely
death of one of our API worker processes.

Possibly the cause of failure that's both the most common
and most preventable is the accidental deployment of buggy
code. It's a very easy thing to do because even if the new
code has tests written for it that were thought to be
nearly exhaustive, because it tends to be only in
production where unexpected inconsistencies and other
thorny corner cases present themselves.

Bugs can also be hugely painful for users. One that makes
it out into the wild for even a few minutes before it's
reverted could still mean the failure of tens of thousands
of requests, which is especially bad if those requests were
related to money.

At Stripe we use a number of defensive operational
techniques to automatically reduce the scope of a deployed
bug and to mitigate the effects of one that does interact
with production requests. The goal is to of course hit 100%
availability despite the inconvenient conditions of
reality, but in the absence of that we want to hit as many
9s we possibly can.

Today we'll cover a pattern called a _soft assertion_ which
we find to be effective in the pursuit of this goal, while
also helping to reduce user pain, being lightweight to
implement, and agnostic enough to technology that it could
plausibly be used in a production stack written in any
language or framework.

## Remembering assert.h (#remembering-assert-h)

Many readers will probably be familiar with the concept of
an assertion from C:

``` c
assert(expression);
```

`assert` is a preprocessor macro that checks to make sure
that its expression evalutes to true. If it's not, some
debugging information is printed to `stderr` and the
program terminates with `abort()`.

Even readers unfamiliar with C may recognize its legacy in
their favorite testing framework, where assertions are a
common staple in many languages, for example `assert_equal`
in Ruby's unit tests, or Rust's `assert_eq!` macro.

An assertion works just as well in an HTTP stack, but
rather than aborting the program, we'd likely want to throw
an exception and have a middleware translate it into a 500
to show to the end user. That way our worker process can
stay alive to fulfill more requests despite the failure.

As an example of where an assertion might be useful,
imagine a web stack where we've successfully authenticated
an incoming request, but find that there is no user record
corresponding to the key that we just authenticated. Its
absence is a direct indication of a data integrity problem
in that a human should take a closer look at. Sending a 500
back to the user making the request is appropriate because
we've encountered an internal error that is unexpected and
which cannot be reconciled in an automated way.

For the purposes of this article, we'll call this type of
assertion a "hard assert" because it must strictly evaluate
to true for an operation to continue.

## The Soft Assertion (#soft-assertion)

Soft assertions are similar to hard assertions, but instead
of crashing the program, we notify someone, recover, and
keep running. The service's operator gets a detailed report
with information on what went wrong and can go investigate.
If everything is working, the end user whose request is
being serviced never even knows that anything went wrong.

The soft assertion's other key behavior is that in a
testing environment they behave identically to strong
assertions in that they cause an uncompromising failure
which will be revealed during a test run. With sufficient
coverage, we'll detect a soft assertion that would ever
fire under normal conditions without leaving the test
suite. Any that fire in production will be exactly the type
of exceptional circumstance that we want to know about.

``` ruby
def soft_assert(expression, message)
  unless expression
    if Utils.environment == "test"
      # If we're running in a test, fail the test case so
      # this gets fixed immediately.
      raise SoftAssertionFailure, message
    else
      # Otherwise, report the failure to our exception
      # tracking service.
      report_assertion_failure(message)
    end
  end
end
```

### Predicting Failure in Rate Limiting (#predicting-failure)

A project I was working on recently involved putting a new
rate limiter in the Stripe API stack. One of the dimensions
that it was limiting on was by the origin IP of an incoming
request.

Now I know that I should always be able to pull an IP out
of a request's environment, but given that we fulfill _a
lot_ of different requests, I also know that there's some
possibility of a corner case that I haven't foreseen
failing to return an IP. If that does happen, my code
shouldn't crash. To hedge against the possibility, I'll add
some code to detect that case, fail a soft assertion, and
recover:

``` ruby
def rate_limiting_dimensions(env)
  ip = Utils.ip_from_env(env)

  if ip.nil?
    soft_assert(ip != nil,
      "Expected every request to have an IP.")

    # Throw this request into a random bucket so that we
    # don't inadvertently rate limit someone.
    ip = SecureRandom.uuid
  end

  [user.id, ip]
end
```

If we have any test cases capable of causing this kind of
situation, I'll find out right away. If not, but it happens
in production, I'll get notified with a message and
detailed context that lets me to debug exactly why an IP
was missing from a request and compensate for that
situation in the next version of my code.

In reality, I deployed the code above and didn't see a soft
assertion fire even once in a month of production traffic.
I can now safely conclude that I was being overly cautious,
and simplify the code by removing the recovery code and the
assertion itself.

## Directed Notifications (#directed-notifications)

If you take a course on lifeguarding, one of the tips that
you'll get is that when you're surrounded by a crowd of
people and want someone to call 911, you should pick _one_
particular person out of the crowd by on of their
distinctive features and specifically order them to do it:
"You with the hat, call 911!"

This technique is designed to overcome [the bystander
effect][bystander-effect]. If you'd simply yelled "Call
911!" into the crowd, each person might assume that someone
else will take care of it, and the worst case scenario is
that _no one_ calls 911. Picking just one person directs
responsibility, and makes the job more likely to get done.

In much the same way, it's useful to assign soft assertions
one particular owner so that they're not just lumped in
with the general pool of production exceptions. Like with
the 911 call, this puts the onus on one particular person
to fix any problems that appear:

``` ruby
soft_assert(ip != nil,
  "Expected every request to have an IP.",
  notify: "brandur@stripe.com")
```

This is especially important in a big engineering team
where the dynamics of the larger group tend to erode the
responsibility that any one person feels for the larger
system. The author of a new feature is probably the best
candidate for who should get notified, at least at first.

After the code's been running in production for a while and
has proven itself to be stable, the specific assignment can
be removed, and the failure lumped back into the general
exception pool to be handled by the engineers currently
on-call.

## Canary Deploys (#canary-deploys)

At Stripe, we use _canary deployment_ to try and avoid the
case of a newly introduced bug taking down our production
fleet. The basic idea is simple: send a new deployment out
to a single node first, and watch for problems. After it's
been confirmed that everything looks fine, the release gets
rolled out to the whole fleet.

TODO: image showing canary

Along with other key metrics like exception count and
availability, the rate of soft assertions failures is also
a great metric to measure on while considering the
integrity of a new release. If there's a sudden spike in
soft assertion failures, the deployment should be
cancelled, and an automated system should take the canary
out of rotation to further reduce the possibility of user
impact, and to give operators a chance to investigate their
exact cause.

## Feature Flags (#feature-flags)

Another technique that combines very well with soft
assertions are feature flags for the rollout of new product
capabilities. Supporting code for major new features is
deployed, but disabled by default as it's hidden behind a
flag that can be activated from an administrative
interface:

``` ruby
if FeatureFlag.on?(:new_currencies)
  process_with_new_currences_available
else
  process_normally
end
```

A major advantage of feature flags over canary deployments
is that they can be enabled incrementally. For example, we
might create a system that allows us to enable the new
route for one percent of our traffic before we slowly
throttle it to be fully on:

``` ruby
FeatureFlag.enable(:new_currencies', 'random', 0.01)
```

If a problem appears, the entire system can be disabled
instantly by turning the flag off. If we managed to
manifest the problem as soft assertions instead of hard
exceptions then we've gotten the best of both worlds: we've
identified a bug in our code, and meanwhile not a single
user was affected.

## Summary (#summary)

As we've shown, soft assertions are a useful technique for
routing out unexpected edge cases, and are particularly
nice because they're easy to implement and can fit into
most tech stacks regardless of their implementation
language. By combining them with canary deployments and
feature flags, we can have far greater peace of mind as
risky changes are deployed.

By institutionalizing the use of soft assertions and
recovery code over throwing exceptions, we help foster an
overarching value of _biasing to success_. Failures can be
costly to our users, and by trying to write code that's
robust enough that it can handle edge cases that we don't
know about yet, we err on the side of failing open by
fulfilling requests, even if they didn't go quite the way
we expected them to.

Our aspirational goal is to never drop a request, even
after accounting for inconvenient realities like ever
present human error. Until we can get there, we'll aim to
shore up our practices to get as many 9s of availability as
we can get.

[bystander-effect]: https://en.wikipedia.org/wiki/Bystander_effect
