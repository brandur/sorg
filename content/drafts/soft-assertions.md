---
title: Soft Assertions
published_at: 2016-11-26T04:31:32Z
hook: Detect problems and inconsistencies in production without affecting
  users.
location: San Francisco
---

Last week I started writing about some of my favorite operational tricks at
Stripe, beginning with [canonical log lines](/canonical-log-lines). Here we'll
cover _soft assertions_, another very lightweight technique that's similarly
useful.

Some readers will probably be familiar with the concept of an assertion from C:

``` c
assert(expression);
```

If not, practically everyone will be familiar with the basic idea from their
unit testing framework. The basic function of an assertion is to check that an
expression evaluates to true, and to terminate the program, throw an exception,
or fail a test if it's not.

An assertion can be deployed to production HTTP stack, but we probably wouldn't
want to just terminate the program like we would in a C program. Instead, we'd
likely throw an exception and have a middleware translate it into a 500 to show
to the end user.
