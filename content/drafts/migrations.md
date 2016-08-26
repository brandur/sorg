---
title: Migrations
hook: While simple without much data, migrations can become an increasingly dangerous part of your stack. Details a few techniques on how to keep migrations safe.
location: San Francisco
published_at: 2016-08-05T23:35:17Z
---

Dangers:

* Database load.
* Callout load on other systems (say sending mail to Mailgun and hit your rate limit).

* Dry run
* `variable_throttle`

* `perform(:action)`
    * `dryrun_action`
    * `real_action`

Write tests!

Use a relational database with a transaction, then verify the results.

Running:

* Use tmux!
