---
title: A Glance at Kubernetes
published_at: 2016-04-12T04:06:32Z
---

After unintentionally putting it off for years, I recently took a look at
Google's [Kubernetes][kubernetes] for the first time, and was impressed by the
sophistication of the project. [This ACM article][acm] does a good job of
explaining its premise along with some of the history of its predecessors at
Google.

I find Kubernetes especially fascinating because instead of just shipping a new
cloud product (in the spirit of Amazon for example), Google's instead decided
to build an open platform on top of the infrastructure. They've also gone out
of their way to prove very early on that it's broadly applicable outside of the
GCE's specialized ecosystem.

Some of the highlights that stood out to me:

* It's multi-provider. AWS, GCE, and Azure available as turnkey solutions. In
  my experience, you get this working early or you don't get it working at all.
* The entire application state being driven purely by etcd is a pretty
  interesting feature. Theoretically you can avoid a single point of failure
  throughout your entire control plane. Back at Heroku, database failures were
  the most common reason for development outages.
* Co-located pods on nodes are awesome. You don't need these for the most basic
  web app deployment, but being able to expand your system with the use of
  agents for log ingestion and the like at a logic layer higher than Puppet is
  a very useful abstraction.
* Let's also not discount the backing of a company like Google is a major asset
  to confidence in its longevity (compare that to Mesos for example).

And some more tepid commentary:

* It's an open question to how difficult it is to operate the system. i.e.
  Theoretically it's trivial, but given 500 deployed nodes, how may full-time
  engineers will it take in practice to babysit it?
* Its documentation is nicely complete, but it and the project's branding feel
  very rough around the edges. This might seem like an inconsequential point,
  but given how sophisticated Kubernetes is already, confusion might be one of
  the biggest blockers to the project's success right now.
* GCE still needs its own RDS. I want to have at least the option for easy
  access to a ready-made database.

There's a lot of companies out there that find PaaS like Heroku too
restricting, but by going to IaaS have ended up building and maintaining their
own platform to make it palatable; a _very_ steep price to pay. The industry
needs a platform that's easy to use, powerful enough to be universally
applicable, and open enough that many organizations across many disciplines are
willing to use it.

The next step for me is to actually try this thing out.

[acm]: https://queue.acm.org/detail.cfm?id=2898444
[kubernetes]: http://kubernetes.io/
