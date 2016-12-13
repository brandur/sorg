---
title: Soft Assertions
published_at: 2016-11-26T04:31:32Z
hook: Detect problems and inconsistencies in production without affecting
  users.
location: San Francisco
---

Last week I started writing about some of my favorite operational tricks at
Stripe, beginning with [canonical log lines](/canonical-log-lines). Continuing
in the same vein of lightweight, technology agnostic techniques, here we'll
cover the idea of _soft assertions_.

Many readers will probably be familiar with the concept of an assertion from C:

``` c
assert(expression);
```

Its basic function is to check that an expression evaluates to true and to
terminate the program if it's not. Even those who haven't seen one in C will
almost certainly be familiar with the idea from testing frameworks in
practically every language.

An assertion works just as well in an HTTP stack, but rather than simply
terminate the program, we'd likely want to throw an exception and have a
middleware translate it into a 500 to show to the end user.

As an example, imagine if we have an authentication system that marshals loads
an access token passed in with an HTTP request into an `AccessKey` model
retrieved from a database. From there, we use the model's many-to-one
association to load a `User` model before we continue to process the request.
Say we successfully load an `AccessKey` model, but find that it has no
associated `User` in the database. This would be an ideal time to assert on the
presence of a `User` because its absence would be a direct indication that
something is amiss with our data's integrity. Sending a 500 back to the user is
perfectly appropriate because this is an internal application error that should
never have occurred.

We might call this type of assertion a "hard assertion" because it must
strictly evaluate to true for processing to continue.

## Introducing the Soft Assertion (#soft-assertion)

Exception in tests and development. Swallowed silently in production.

Catches problems in edge cases that may have been forgotten in the test cycle.

## Implementation (#implementation)

Give it a team and even an individual name.

Stick a failure in a queue and have it mail out as well as go to Sentry or your favorite exception tracking service..
