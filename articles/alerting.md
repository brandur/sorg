---
hook: A set of general guidelines to consider when designing a alerts for a production
  system.
location: Leipzig (finished in San Francisco)
published_at: 2015-08-18T11:28:48Z
title: Designing Alerts
---

Adding alerts to systems has become a widespread standard practice that helps
engineers keep their production systems up. The basic idea of such a setup is
that systems should be designed in such a way that they should be able to
compensate for common failure cases, but if something happens that's beyond the
boundaries of the system's ability to handle, it should tell a human about it
so that they can come in and fix the problem manually.

The on-call model for software was adopted from other industries, the surgeon
who may get called in to perform an emergency surgery for example. But unlike
the surgeon and quite a few of our other pager-bearing counterparts, it's often
possible to achieve a considerable amount of automation in software industry
with the brunt of failures being handled by internal systems automatically and
relatively few dropping through the cracks to a human. Receiving a page is of
course a less-than-desirable outcome because it might lead to someone waking up
in the middle of the night to fix a problem at work, making automation
attractive (but not wholly sufficient in itself).

Adopting a system of alerting certainly isn't a problem-free endeavor though;
like many other areas in technology, there are a plethora of pitfalls to run
into.  Lack of a appropriate discipline while designing them can lead to bad
weeks on-call and operator fatigue. Specifically, here are a few problems that
are easy to run into:

* Alarms that page too aggressively; they're responded to and it turns out that
  nothing is wrong.
* Alarms that aren't specific enough; they're responded to and significantly
  more analysis is needed to figure out what's going on.
* Alarms that need to be passed down the line because they only represent an
  indirect failure in the observed service.
* Alarms that go off time and again, without the root problem ever getting
  fixed.

There's no universal playbook that'll ensure that these kinds of things never
surface, but following a few general guidelines can help reduce the number of
them significantly. Here's a few that I've assembled over the years while
putting alerts into production for various services.

## Guidelines (#guidelines)

### Design for Granularity (#granularity)

There's nothing worse than waking up in the middle of the night and discovering
that an alert has gone off that doesn't have an obvious remediation because it
could mean that any number of things have gone wrong. This inevitably leads to
a drawn out investigation that's further slowed by the operator being
half-asleep.

This one may seem obvious, but there are quite a few types of alerts that seem
like a good idea until a closer inspection reveals that they're breaking the
granularity rule. For example, an alert on something like a service's HTTP
`/health` endpoint is a very widespread pattern, but one where a failure can
ambiguously mean anything from a thread deadlock to its database going down. A
much more powerful alternative pattern is to have a background process
constantly logging fine-grain health information on a wide range of system
telemetry, and using that to implement alarms that try to detect each type of
failure condition individually. This will allow an operator to identity the
root of the failure faster, and execute a quick resolution.

A goal to shoot for here is to make sure that every alarm in your system has a
1:1 ratio with a possible causation. If receiving an alert could mean that more
than one thing has gone wrong, then there's probably room to make that alert
more granular.

### Alert at the Root Cause (#root-cause)

Alerts should be designed to measure against the metric being emitted by the
system which is the most directly relevant to them. This is another one that
seems like obvious advice, but even an alert designed by an advanced user can
often be found to have a number of unnecessary abstractions layered on top of
it when scrutinized closely. A system operator's goal here should be to slice
through these abstractions until only the most basic level is left.

As an example, I previously wrote about how [long-lived transactions degraded
the performance of our Postgres-based job queue](/postgres-queues). We'd
originally been alerting based on the number of jobs in our background queue
because that was the most obvious symptom of the problem. Upon closer study,
we realized that the job queue only bloated because the time to lock a job was
increasing, so that lock time became a more obvious candidate for an alert. But
going in even further, we realized that the reason lock time degraded was
almost always due to an old transaction somewhere in the system, so we started
alerting on that. After even more time passed, we noticed lock degradation that
was unrelated to oldest transaction, so we started alerting on the number of
dead tuples in the table, which is directly correlated to lock time and an
early warning for when the system starts degrading for any reason.

### Minimize External Services (#external-services)

In all cases except maybe your most mission critical system, it's not worth
waking up your operators when a third party service goes down that one of your
components happens to depend on. Keep your alerts inward-facing so that if they
trigger, there's always meaningful action that can be taken by your team rather
than just passing that page on to someone else.

By extension, wherever you have any measure of control (with other teams
internally for example), try to encourage the operators of services that you
depend on to maintain appropriate visibility into their own stacks. Your goal
here is certainly to make sure that the system as a whole stays up, but that
the team receiving the page are the ones with the best ability to influence the
situation.

A misstep that we made internally is that the component that handled [Heroku
Dropbox Sync](https://devcenter.heroku.com/articles/dropbox-sync) ended up
being built on top of a rickety component whose job it was to stream platform
events and which had a very poor track record for reliability. It was
ostensibly owned by my own team, and we only had bare bones alerting on it.
Dutifully though, they put an alarm in place around an end-to-end integration
test that injected a release into a Heroku app and waited for it to come out of
the other end. When the audit steamer failed, they got paged, and they
re-raised those pages to us, resulting in a bad situation for everyone
involved.

### Safe at Rest (#safe-at-rest)

One tempting mistake in a well-trafficked production environment is to build an
alarm off of the ambient load in the system. For example, given a service
designed to persist auditing events into a long-term archive we might alert on
the fact that an event was persisted in the last five minutes. This often won't
show a problem for a long time, but is undesirable because these kinds of
alarms can trigger false positives in certain situations like a lull in traffic
or a system-wide maintenance state, and also map poorly to development
environments where there is no consistently reliable load.

Whenever possible, design alerts that don't rely on any ongoing traffic at all,
but if that can't be avoided, then make sure that there's a built-in
multi-environment mechanism for stimulating it artificially.

### Avoid Hypotheticals (#avoid-hypotheticals)

An overly enthusiastic engineering spinning up a new service might fall into
the trap of guessing where all the alarms on it should be. Well-understood
failure cases should be planned for and designed against, but some care should
be taken to not roam too far out into the realms of the hypothetical. If in the
future these alarms do end up going off, they'll more often than not take an
operator by surprise and course to resolution unclear.

Stay conservative when it comes to adding new alerts; it's okay to add alerts
that are expected proactively, but for most others it might be better to wait
until more concrete information is available. It's always possible to add new
alerts when new problems occur or unexpected situations are observed.

### Throttle On Slowly (#throttle-slowly)

Being on the wrong end of a pager after a new product goes into production
might lead to a harrowing week. Luckily, no product goes into production
overnight. Take advantage of the relatively long product lifestyle by putting
in alerts during the alpha and beta phases that produce a notification that
somebody will receive eventually (like an e-mail), but which won't page outside
of business hours. One those warning-style alerts are vetted and stable,
promote them to production.

### Don't Allow Flappy Alarms to Enter the Shared Consciousness
(#flappy-alarms)

As busy engineers, one bad habit that we're particularly susceptible to is
applying the fastest possible fix to a problem and moving on without
considering whether there may also be an only incrementally more expensive
solution that could buy a much more permanent fix. In the case of an alarm,
this often looks like responding to it and doing some basic investigation to
make sure that nothing is seriously wrong, but without considering that the
alarm may be very broken and badly in need of attention. Over time, it's easy
to become desensitized to these types of flappy alarms to the point where they
become built into the team's shared consciousness and where no one will
consider them to any real depth.

Newer employees might be especially susceptible to this problem because as far
as they're concerned, some alert might have been going off for the entire
length of their contemporary career. They'll also make implicit assumptions
that their senior colleagues would have looked into the problem already if
there was anything that could be done about it.

My advice for these types of situations is (of course) to try to spend a bit of
time trying to tweak or change the alarm so that it's less prone to produce
false positives. _However,_ if nothing can easily be done to improve it, it's
far better to eliminate the alarm completely than leave it in-place in its
degraded state. Given a bad alarm, responders are already unlikely to be doing
much of anything useful when it goes off, so it's better to save them the pain
in the first place.

An example of this that we had was to put an alert on 500 status codes coming
from backend services when after we had an incident that involved a service
going down that we would have been easily able to detect. The alert was added,
but at a level that would trigger based on occasional ambient spikes in backend
errors, which caused it to go off randomly every day or two. Every time it did,
an operator would have to go in, find the service that was causing the trouble,
and compare its current error levels to historical records before deciding how
to proceed. It didn't take long before operators were ignoring these alarms
completely, making them noisy and worthless.

### Treat Alarms as an Evolving System (#evolve)

As an extension to the previous point, it's a good idea to always think about
your current set of alarms as a evolving system that you can and should be
constantly tweaking and improving. Obviously this applies to adding new alarms
as exotic new failure cases are discovered, but even if what already have works
pretty well, there may still be a more optimal configuration or a different
alarm that could be put in place that does a better job compared to what's
already in there.

This also applies in the reverse: try to never get yourself in a position where
you're cargo culting by keeping alarms around just because they've always been
there. Even if you're not the original author of a particular component, take
control of its stack and keep it sane.

### Empower Recipients to Improve the Situation (#empower-recipients)

When I first started working at Heroku, we had a global pager rotation where
for one day every few weeks, one on-call engineer would respond to any problem
that occurred across the entire platform. For reasons that are hopefully mostly
intuitive, this situation was utterly depraved; engineers would wake up,
acknowledge pages, follow playbooks by rote, and hope against all odds that
this would be the last page of the night. Everyone had strong incentive to fix
the problems that were interrupting their sleep and lives half a dozen times a
day, but for the most part these problems were in foreign codebases where the
cost to patch them would be astronomically high.

We eventually did away with this catastrophe by moving to team-based pager
rotations and inventing the "ops week", which generally meant that the on-call
engineer wasn't working on anything else besides being on-call. This would give
them a bit of free capacity to go in and address the root problems of any pages
that had woken them up, thus empowering them to reduce their own level of pain.

### Observe Ownership (#ownership)

It may be tempting for an enthusiastic operator to put alarms into place that
indeed do represent failure cases for a service that they own, but which upon
closer inspection may have a root cause that lies outside of its area of
responsibility. As mentioned in the ["Alert at the Root Cause"](#root-cause)
above, it's important to make sure that the most direct source of a problem is
traced, and in some cases that source may lie across team boundaries. A poorly
placed alarm may result in an operator waking up only to pass the page onto
another team who can address the problem more directly. Given this situation,
it's far better to have that other team wake up first; hopefully they can
address the problem before it bubbles up to everyone else.

Once again, this one might seem obvious, but there are a number of situations
where this problem is easy to encounter. For example, given a situation where
an alarm belongs in the realm of a mostly non-technical team without a good
operational situation, it might be easier to keep the alarm on your side rather
than fire up the various bureaucratic mechanisms that would get them to analyze
the situation and eventually take ownership. But even if easier in the short
term, it's likely to cause trouble over time in that your team is unlikely to
ever be able to address the underlying problem (see ["Empower Recipients to
Improve the Situation"](#empower-recipients) above).

## Summary (#summary)

There's a common theme to all the guidelines mentioned above: most of them are
intuitive at first sight, but can still represent dangers for even experienced
teams. A successful process tries to use whatever guidelines are helpful to put
together an initial set of alarms for a service, and then makes sure to iterate
until that set has been optimized to maximize uptime and speed to resolution,
all the while decreasing operator pain.
