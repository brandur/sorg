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
unit testing framework.



From `man assert`:

```
The assert() macro tests the given expression and if it is
false, the calling process is terminated. A diagnostic
message is written to stderr and the abort(3) function is
called, effectively terminating the program.

If expression is true, the assert() macro does nothing.
```
