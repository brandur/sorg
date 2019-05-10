+++
hook = "TODO"
published_at = 2017-05-11T18:46:36Z
title = "On managed environments"
+++

At Stripe we distribute Puppeted laptops to every new
developer that lets them get a running start with Ruby,
Rbenv (and its various plugins), Bundler, GPG, and many
other tools pre-installed. This has a few nice
characteristics in that it's often enough to get almost
every new employee to be able to a successful deployment on
their first or second day.

There are similar ideas with Docker: instead of setting up
a working local environment, write a `docker-compose.yml`
and let the pieces will slide together.

## Self-help not endorsed (#self-help)

In my experience, these managed systems can be fine for
days or weeks, but eventually the user will hit a problem,
and when they do, no self-help is possible. The combination
of not having had a hand in building their own environment,
and the complexity introduced by the managed configuration,
means that the only route to a solution is to get direct
help from a maintainer of the scheme. This will happen
again and again, because with the crutch to lean on, at no
point does the user come to understand it any better.

## The zen of self-assembly (#self-assembly)

Contrast this to self-assembly. It's painful for a day or
two, but the user sees every layer go in. Even if they
don't have a perfect understanding as they're doing it,
they'll have some memory that will come in handy for
troubleshooting.

That's not to say that you should go out and build your own
operating system or car. There is a point where the
complexity is so considerable, and the walls of the black
box sufficienty opaque, that it's appropriate to seal it
off and accept the cost of going to see the mechanic a few
times a year.

Development environments don't fit in this category.
They're lighter weight for one, but also so dynamic
(changes to code, runtimes, and libraries are happening on
a daily basis) that they're more prone to problems. In
their case, the extra pain is well worth the insight that
it buys.
