---
title: Not Hosted Here
published_at: 2016-04-27T15:23:51Z
---

Everyone of course knows of the widspread anti-practice of [_not invented
here_][not-invented-here], whereby many software companies overzealously
rewrite software internally when an existing package would have worked fine.
The net effect is normally a project that doesn't work as well as the existing
alternative would have, and which becomes a maintenance burden for engineering
and operations teams forever.

I'd like to throw another in the same vein into the mix: **not hosted here**
(NHH). Conformance to this idea involves the insistence of installing and
maintaining a self-hosted system where leaning on an existing service would
have been faster, cheaper in the short run, and _much_ cheaper in the long run.
It's usually done in the name of compliance, security, or perceived need for
customization or access to specific features, and can take many forms:

* Far and away the most common case, running a self-hosted service where a
  hosted equivalent is as good or better. Some examples:
    * HipChat instead of Slack.
    * GitHub Enterprise instead of GitHub.
    * Sentry instead of Rollbar.
    * Phabricator instead of GitHub.
    * HackPad over your favorite hosted wiki.
* If you ever find yourself running your own mail server in this day and age,
  you're experiencing very severe NHH.
* Running Kafka boxes where an equivalent AWS or Google Cloud service would
  work just as well.
* Maintaining custom server builds instead of using a platform like Heroku
  (before your company has on the order of 100 engineers).
* Running bare metal servers over cloud equivalents (before your company has on
  the order of 10k engineers).

The most common consequence of NHH is the long term expense that goes into
maintaining the hosted service. Where a cluster of Kafka servers might require
a team of 1-5 engineers to operate in near perpetuity, using Kinesis requires
none. Services like HipChat can usually operate fairly autonomously for longer
periods, but without a caretaker will fall severely out-of-date with the
current version, and may have difficult and time-consuming upgrade paths. The
worst possible consequence may be related to security in that outstanding
vulnerabilities won't get patched as quickly as a hosted version. Other
problems might include chronic unreliability, degraded performance as services
are slow to be scaled up with usage, and frequent service outages for offline
maintenance.

[not-invented-here]: https://en.wikipedia.org/wiki/Not_invented_here
