---
title: Soft Assertions
published_at: 2016-11-26T04:31:32Z
hook: Detect problems and inconsistencies in production without affecting
  users.
location: San Francisco
---

A few weeks ago I started writing about some of my favorite
operational tricks at Stripe, beginning with [canonical log
lines](/canonical-log-lines). Continuing in the same vein
of lightweight, technology agnostic techniques, here we'll
cover the idea of _soft assertions_.

Many readers will probably be familiar with the concept of
an assertion from C:

``` c
assert(expression);
```

Its basic function is to check that an expression evaluates
to true and to terminate the program if it's not. Even
those who haven't seen one in C will almost certainly be
familiar with the idea from testing frameworks in
practically every language.

An assertion works just as well in an HTTP stack, but
rather than simply terminate the program, we'd likely want
to throw an exception and have a middleware translate it
into a 500 to show to the end user.

As an example, imagine if we have an authentication system
that marshals loads an access token passed in with an HTTP
request into an `AccessKey` model retrieved from a
database. From there, we use the model's many-to-one
association to load a `User` model before we continue to
process the request. Say we successfully load an
`AccessKey` model, but find that it has no associated
`User` in the database. This would be an ideal time to
assert on the presence of a `User` because its absence
would be a direct indication that there's a data integrity
problem in our database. Sending a 500 back to the user is
perfectly appropriate because this is an internal
application error that should never have occurred.

For the purposes of this article, we'll call this type of
assertion a "hard assert" because it must strictly evaluate
to true for an operation to continue.

## The Soft Assertion (#soft-assertion)

Soft assertions are similar to hard assertions, but instead
of crashing the program, we tell it notify us and keep
running. The service's operator gets a detailed report with
information on what went wrong and can go investigate. The
end user whose request is being serviced never even
realizes that it happened.

In a testing environment soft assertions behave exactly
like strong assertions and fail a test run. With sufficient
test coverage, a soft assertion that would fire on any
standard path. Any that do fire in production should be a
truly exception circumstance that we'd want to know about.

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

### Example: IP Rate Limiting

A project I was working on recently involved putting a new
rate limiter in our API stack. One of the dimensions that
it was rate limiting on was by the origin IP of an incoming
request.

Now I know that I should always be able to pull an IP out
of a request's environment, but given that we fulfill _a
lot_ of different requests, I also know that there's some
possibility of a corner case that I haven't foreseen
happening and failing to return an IP.

If that does happen, I don't want my code to crash. To
hedge against the possibility, I'll add some code to detect
that case, fail a soft assertion, and recover:

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
else will take care of it, and the worst case result is
that _no one_ calls 911. Picking just one person directs
responsibility, and gives the job a better chance of
getting done.

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
has proven itself to be stable, it's usually fine to remove
the assignment to a specific person and lump it back into a
general pool for currently on-call engineers to take a look
at.

## Canary Deploys (#canary-deploys)

At Stripe, we use _canary deployment_ to try and avoid the
case of a newly introduced bug taking down our whole fleet.
The basic idea is very simple: send a new deployment out to
a single node taking real traffic first, and watch for
problems that occur. After it's been confirmed that
everything looks fine, the release gets rolled out to the
whole fleet.

Along with other key metrics like exception count and
availability, the prevalence of soft assertions is also a
great metric to measure on while considering the integrity
of a new release. If there's a sudden spike in assertion
failures, the deployment should be cancelled, and an
automated system should take the canary out of rotation to
further reduce user impact and to given an operator a
chance to investigate the exact cause of the failures.

[bystander-effect]: 
