---
hook: |-
  The case for a concerted effort to build a powerful, but streamlined, platform
  on AWS.
location: San Francisco
published_at: 2016-06-15T15:24:53Z
title: AWS Islands
---

By commoditizing the management of servers and other resources, AWS is
indisputably an incredible tool that provides an inordinate amount of leverage
to organizations that use it. Using AWS, most of us can can avoid ever setting
up server hardware, changing a corrupt disk, or troubleshooting a faulty
router; thus avoiding the need to bring entire technical disciplines in-house.

But AWS focuses a little _too much_ on infrastructure. Its APIs, SDKs, and even
web console are feature-rich and powerful, but don't incorporate the workflows
that its users are following as they go about their work day-to-day. It's so
low level that every company using it ends up enriching their experience by
building custom tooling to help with commonly needed tasks like getting
fleet-wide visibility, running deployments, or managing configuration. Generally
this tooling is built using common foundational tools like Puppet and AWS SDKs,
but it's still very much custom.

This effect has a name, the [Galápagos syndrome][galapagos]; originally coined
to refer to the isolated branch of development that is the Japanese mobile
phone industry. Each organization's tooling is its own branch of evolutionary
development, and like the disappearing mobile phones of Japan [1], probably a
dead end of development at that.

The problem with internal tooling is that it invariably sucks. Usually
appearing in the fast-moving early stages of a company and not being exposed to
the same rigor as open-source, it's often self-inconsistent, undocumented,
logically unsound, slow, and with no test suite in sight. That isn't to say
that it isn't still useful, especially relative to the purely utilitarian base
layer that is AWS, but it represents a huge cost of development and maintenance
for software that's generally mediocre at best.

As resource-conscientious organizations, we should be trying to pool these
discrete points of considerable expenditure towards the development of a rich
middle tier from which we can all reap the benefits. This tier would meet these
criteria:

1. **Flexible** enough to meet the requirements of organizations that will have
   a wide spectrum of sizes and use cases.
2. **Simple** enough that smaller shops with fewer resources can use it to
   quickly and easily deploy just a few basic apps to jumpstart their business
   (ideally as easy as `git push heroku master`).
3. **General** enough so as not to lock into features that are forever going to
   be specific to AWS so that it can be run on a number of different cloud
   providers.

The potential upside here is a significantly improved operator experience, the
potential to avoid vast swaths of internal only bugs, and hugely streamlining
operations effort by cutting out unnecessary software development and
maintenance. The downside is having to learn someone else's system, but this is
no worse than what we already do for AWS.

[Kubernetes][kubernetes], with its strong engineering team and with Google
backing its finances and notoriety, is my favorite candidate so far, but it's
still too early to tell whether it's going to be a clear winner.

[1] While maintaining a rich domestic phone culture for a long time, almost the
    entirety of Japan's population [is now on iOS and Android devices][share].

[galapagos]: https://en.wikipedia.org/wiki/Galápagos_syndrome
[kubernetes]: http://kubernetes.io/
[share]: http://www.statista.com/statistics/260415/market-share-held-by-smartphone-operating-systems-in-japan/
