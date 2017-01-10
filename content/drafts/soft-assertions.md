---
title: Soft Assertions
published_at: 2016-11-26T04:31:32Z
hook: Detect problems and inconsistencies in production without affecting
  users.
location: San Francisco
---

In the next few months, I'm going to write about some of my
favorite operational tricks that I've seen inside Stripe
over the year that I've worked here. I'll specifically aim
to cover techniques that are lightweight to implement and
technology agnostic, so that they're still interesting even
if your stack doesn't have a thing in common with ours.
Today we'll cover _soft assertions_.

Many readers will probably be familiar with the concept of
an assertion from C:

``` c
assert(expression);
```

`assert` checks to make sure that its expression evalutes
to true, and terminates the program if it doesn't. Even
readers unfamiliar with C will probably recognize it from
their favorite testing framework, where the assertion is a
staple in practically every language.

An assertion works just as well in an HTTP stack, but
rather than simply terminate the program, we'd likely want
to throw an exception and have a middleware translate it
into a 500 to show to the end user. That way our worker
process can stay alive to fulfill more requests despite the
failure.

As an example of where an assertion might be useful,
imagine a web stack where we've successfully authenticated
an incoming request, but find that there is no user record
corresponding to the key that we just authenticated. Its
absence is a direct indication of an integrity problem in
our database that a human should take a closer look at.
Sending a 500 back to the user making the request is
appropriate because we've encountered an internal error
that is completely unexpected.

For the purposes of this article, we'll call this type of
assertion a "hard assert" because it must strictly evaluate
to true for an operation to continue.

## The Soft Assertion (#soft-assertion)

Soft assertions are similar to hard assertions, but instead
of crashing the program, we have it notify someone and keep
running. The service's operator gets a detailed report with
information on what went wrong and can go investigate.
Ideally, the end user whose request is being serviced is
none the wiser than anything went wrong.

In a testing environment soft assertions behave identically
to strong assertions and fail a test run. With sufficient
coverage, we'll detect a soft assertion that would ever
fire under normal conditions without leaving the test
suite. Any that fire in production will be exactly the type
ofexceptional circumstance that we want to know about.

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

### Example: IP Rate Limiting (#example-ip)

A project I was working on recently involved putting a new
rate limiter in the Stripe API stack. One of the dimensions
that it was limiting on was by the origin IP of an incoming
request.

Now I know that I should always be able to pull an IP out
of a request's environment, but given that we fulfill _a
lot_ of different requests, I also know that there's some
possibility of a corner case that I haven't foreseen
failing to return an IP.

If that does happen, my code shouldn't crash. To hedge
against the possibility, I'll add some code to detect that
case, fail a soft assertion, and recover:

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
assertion fire even once in three weeks of production
traffic. I can now safely conclude that I was being overly
cautious, and simplify the code by removing the recovery
code and the assertion itself.

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

The author of a new feature is obviously a good candidate
for who should get notified, at least at first.

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

Along with other key metrics like exception count and
availability, the rate of soft assertions failures is also
a great metric to measure on while considering the
integrity of a new release. If there's a sudden spike in
failures, the deployment should be cancelled, and an
automated system should take the canary out of rotation to
further reduce the possibility of user impact, and to give
operators a chance to investigate their exact cause.

## Summary (#summary)

As we've shown, soft assertions are a useful technique for
routing out unexpected edge cases, and are particularly
nice because they're light weight enough to make a good fit
into any tech stack regardless of its implementation
language.

By institutionalizing the use of soft assertions and
recovery code over throwing exceptions, we help foster an
overarching value of _biasing to success_. Especially for a
product like ours that deals with money, a failure on any
request can be costly to our users, and by trying to handle
those tricky edge cases, we err on the side of failing open
by fulfilling requests, even if they didn't go quite the
way we expected them to. Our aspirational goal is to never
drop any request, even after accounting for the unexpected
difficulties of the real world.

If this sort of production hardening sounds interesting to
you, we're hiring! Check out our [jobs page][jobs] to see
our open positions.

[bystander-effect]: https://en.wikipedia.org/wiki/Bystander_effect
[jobs]: https://stripe.com/jobs
