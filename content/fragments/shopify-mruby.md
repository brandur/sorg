---
title: "Shopify Scripts, MRuby, and the $500,000 Release"
published_at: 2017-04-07T15:39:30Z
location: San Francisco
hook: A good story about technological optimistic and
  languages that might be a little too permissive.
---

Here's a pretty good story about being a little too
optimistic with technological implementation, and
especially where applied to a language that lets you do
whatever you want.

Last year Shopify released a product called [Shopify
Scripts][shopify-scripts] that lets store owners run
arbitrary Ruby code after one of their customers adds an
item to their cart. This could be used to apply a product
discount or activate a promotion for example. The product
ran on MRuby, a lightweight version of the Ruby language
which is designed to be more embeddable than its
progenitor, and secured in a sandboxing engine of Shopify's
own design called [MRubyEngine][mruby-engine].

Shopify quickly found out the hard way that an extremely
permissive language combined with a nascent C codebase
wasn't a recipe for secure code, and vulnerability reports
started flowing in. For example:

* [Failure to do an array bound checks from MRuby, leading
  to possible remote code execution][vuln-1] (worth $20k).
* [Improper optimization in MRuby can lead to a null
  pointer dereference][vuln-2] (worth $10k).
* [Overriding `#to_s` to return a `nil` return leads to
  MRuby aborting][vuln-3] (worth $10k).

It's worth noting that as far as vulnerabilities go, these
are all pretty run-of-the-mill bugs. Many security
researchers work a lot harder to find bugs that are worth a
lot less.

Shopify did the only thing they could which was to put in
kernel-level sandboxing around MRuby, but by then they must
have already paid in close to $450k (the all time total as
I'm reading it today is $479,300):

> Update Dec 9, 2016: We wish to thank the researchers who
> have submitted vulnerability reports to the Shopify
> Scripts program. As of today, we have implemented
> technical mitigations (seccomp-bpf sandboxing and process
> isolation) on the application servers hosting Shopify
> Scripts. As a result, we expect most vulnerabilities will
> no longer be exploitable without additional bugs in the
> kernel or seccomp itself, and so we are lowering the
> payout amounts for our program to 10% of previous levels.
> Researchers who have submitted issues prior to today will
> receive payouts under the old structure, however any new
> reports received will be eligible for payouts within the
> structure below.

A few takeways come to mind:

* Writing your own compiler from scratch might not be a
  good idea anymore, even if you've got 30 years of
  experience doing it (put another way, use LLVM).
* Overly permissive languages introduce a lot of scary
  possibilities that are difficult to protect against.
  Consider something that's not Ruby.
* It's probably not possible to have enough experience in C
  to be able to write it safely. Looking at some of these
  MRuby patches, the implementation's code is so obscuring
  that it introduces plenty of opportunity for error. When
  you get one, it's often a buffer overrun that leads to
  remote execution.

All of that said, Shopify Scripts is a great idea. I'm glad
to see that an approach involving a more depth-wise defense
seems to have been successful.

[hackerone]: https://hackerone.com/shopify-scripts
[mruby-engine]: https://github.com/Shopify/mruby-engine
[shopify-scripts]: https://help.shopify.com/api/tutorials/shopify-scripts
[vuln-1]: https://hackerone.com/reports/181321
[vuln-2]: https://hackerone.com/reports/181828
[vuln-3]: https://hackerone.com/reports/180977
