---
title: The Chatty Pyramid
published_at: 2017-05-30T17:10:46Z
location: San Francisco
hook: TODO
---

My time in university coincided with Facebook's massive
surge in popularity. I'd walk into a lab in the computer
science building. At any given moment, about two thirds of
them were awash with the site's signature blue.

These machines have been purchased by the university with
expectation that students would use them to push outwards
on the frontiers of computer science. Instead, they were
mostly used to post status updates.

Walk around a modern office today and communication is the
majority of what you see. A considerable number of people
are behind a closed door in a meeting. Those that aren't
are more often than not staring at Slack, their email, or
their calendar.

This is an effect that Frederick P. Brooks Jr. warned us
about in _The Mythical Man Month_. As the size of your team
increases, so does the coordination overhead involved in
running it. There are more nodes in the team graph, and the
communication edges between each of them gets exercised
regularly, which produces a quadratic correlation. From the
book:

> Intercommunication is worse. If each part of the task
> must be separately coordinated with each other part, the
> effort increases as n(n - 1)/2. Three workers require
> three times as much pairwise intercommunication as two;
> four require six times as much as two. If, moreover,
> there need to be conferences among three, four, etc.,
> workers to resolve things jointly, matters get worse yet.

**TODO:** Diagram of communication graph.

50 developers give 50 · (50 – 1) / 2 = 1225 channels of communication

## Deep work (#deep work)

## Structuring for modularity (#modularity)

Almost every company in the world has avoided quadratic
communication inefficiency by adopting a structure that I
like to call "the chatty pyramid", which is the safest
corporate organizational scheme known to man. In the
pyramid, every 3-10 workers along the pyramid's floor are
overseen by a manager, and every 3-5 managers are wrapped
up under a director. Layers of senior directors, VPs, EVPs
are built on that all the way up to the executive level.
Communication is still quadratic within teams, but it's
localized _to_ teams, with inter-team affairs delegated
upwards.

**TODO:** Diagram of communication graphs built as a
pyramid.

The chatty pyramid's emphasis on coordination isn't
surprising if you look at how most companies operate.
Developers build code and architecture that gets the job
done, but is minimal in quality and with little supporting
material like documentation, testing, or tooling.
Oftentimes it's not even possible to figure out how to get
a system configured and running without help from the
original author, let alone add a complicated new feature or
squash a subtle bug. Employees start to reach out to each
other because there's no other choice, and this builds to a
chatty culture where each person is spending the lion's
share of their time talking to other people.

## The search for an alternative (#alternatives)

The chatty pyramid is an imperfect structure. It mitigates
the effect of communication explosion, but only by treating
the symptoms while ignoring the cause. Intra-team
communication still eats into big parts of an IC's total
work week which means that per person efficiency stays
relatively low. Keeping teams small also means more
managers, and more managers to manage the managers, and
more managers to manage the managers' managers, so more
often than not the pyramid gets top heavy as an
organization continues to grow.

Lately I've been reading and watching some old favorites
like [Valve's employee handbook][valve] and [GitHub's talks
on structure][github]. It's not obvious that these
structures were wildly successful for either company, but
the key is that they were experimenting! The chatty pyramid
will continue to reign supreme unless a successful
alternative is found, and that alternative won't be
designed in a vacuum.

## A more self-contained future (#self-contained)

It seems to me that we're forgetting the value of
self-service. If an employee can get what they need by
themselves through documentation or other means (within a
reasonable period of time of course), you've just saved a
considerable portion of another person's time whom that
first employee would have had to ask about it. Even better,
the effect becomes multiplicative because people build
their self-help muscles. Just like reps at the gym, every
time they figure something out for themselves, they get a
little better at it. This is unlike the habit of leaning on
others, which yields a temporary fix but nothing in the way
of lasting information.

It's not clear whether a perfectly self-service environment
is possible (or even desirable), but there are relatively
easy habits that we can develop to foster one:

* Write code that's optimized for others rather than
  ourselves. Keep cleverness down and readability high. Use
  abstractions, but don't make them too heavy. Don't just
  write tests, but write tests that are legible and
  maintainable. Don't write too few or too many.

* Write companion documentation, but not too much of it.
  Succinct documentation is easier to read and update.
  Verbose documentation is especially prone to bit rot, and
  even a little unreliable documentation quickly leads to a
  culture where people don't trust it.

* Provide tooling to help bootstrap and operate projects.
  This probably looks like very common operations
  encapsulated into scripts. Recall that there's a trade
  off here between time and maintenance. It's possible to
  script an infrequently performed operation, but there's a
  good chance that it won't work by the next time it's
  needed. Evaluate this trade off, but also try to write
  scripts to be as robust as possible (write tests and
  probably don't use Bash/sh).

* In the style of [README drive development][readme], write
  a page that acts as an easy hub for discovery. It should
  have an explanatory paragraph or two on background and
  architecture, and then links to other documentation
  resources and a brief on available tooling. Tell people
  how to build it and the run the test suite.

* Develop cross-organization conventions on how to find
  information and where things will generally be. For
  example, hubs are in the README, other docs are in the
  wiki, bootstrapping is possible by running `bin/setup`,
  or scripts are always kept in the `scripts/` directory.

[github]: https://example.com
[readme]: https://exmaple.com
[valve]: https://example.com
