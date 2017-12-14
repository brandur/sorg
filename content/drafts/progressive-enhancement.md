---
title: Progressive Enhancement in Product Design
published_at: 2017-12-13T16:37:21Z
location: San Francisco
hook: TODO
---

At re:Invent this year, Amazon announced the release of
[Fargate][fargate], a product that aims for the panacea of
container-based deployments: easily running containers
without the overhead of managing the underlying
infrastructure. Its promise is simplification. A shallow
stack is a manageable one; there's fewer moving parts that
a user needs to setup and understand.

Done right, this sort of product will eat the ops world.
From the solo developer all the way up to the largest
organizations; only the very few really need deployments
that allow fine-grain control over every aspect of the
infrastructure, but until now there's been little choice in
the matter.

I deployed my first container to Fargate the other day. It
worked fine, but I couldn't shake the feeling that it could
have been so much better. Let's take a quick glimpse at the
steps involved in getting an application up and running on
Fargate.

!fig src="/assets/progressive-enhancement/step1.png" caption="Fargate setup step 1: define a task."
!fig src="/assets/progressive-enhancement/step2.png" caption="Fargate setup step 2: define a service."
!fig src="/assets/progressive-enhancement/step3.png" caption="Fargate setup step 3: configure a cluster."

## The looming mountain ahead (#looming-mountain)

Let's count the number of concepts that a new user needs to
understand to walk through Fargate setup:

* "Tasks" (and it doesn't help that it's one of the most
  over overloaded words in the English language when
  applied to computer science).
* VPCs.
* vCPUs (their relative capability compared to normal
  "non-v" CPUs).
* Execution roles.
* "Services" (and how it's distinct from a task --
  non-trivial).
* Security groups.
* Load balancers.
* Clusters.
* Subnets.

Amazon's been kind of enough to provide defaults for pretty
much everything, .

## Naive education (#naive-education)

The argument could be made that it's better to have all the
aspects of a product visible up front. If a user later
needs to fix a problem, they'll be totally stuck if they
don't understand how anything works.

Unfortunately, even after familiarizing myself with all the
basic Fargate primitives, I still found myself unable to
answer any non-trivial questions. How do I get operational
introspection (e.g. logs)? How do I connect a load balancer
to a running application? What happens if a task's host
dies? (Is a new one booted?)

I'd put forward a good effort towards self-education, but I
still didn't really know anything. Just like with a simpler
product, true familiarity wouldn't come until I started
working with it really closely with it or running into
problems.

## Progressive enhancement (#progressive-enhancement)

[fargate]: https://aws.amazon.com/fargate/
