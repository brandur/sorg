---
title: Implicit behavior
published_at: 2016-08-04T19:37:47Z
hook: UNWRITTEN. This should not appear on the front page.
---

``` ruby
app.metadata.migrated = true
app.save!
```

Cuts a new release!

``` ruby
merchant.metadata.migrated = true
merchant.save!
```

Sends a webhook!

These are both just mines waiting for a novice developer to come through and
detonate them as they try to perform what should have been a relatively benign
operation, like say running a data migration.

One thing that makes Ruby one of the most dangerous languages out there is that
its meta-programming conventions invite abuse. Take ActiveRecord's `after_save`
and `before_save` hooks for example. There are valid use cases for both, but
there's going to be an irresistable background force that will make these
likely to be exploited over time.
