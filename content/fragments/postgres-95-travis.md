---
title: Postgres 9.5 on Travis CI
published_at: 2016-10-01T18:38:29Z
hook: A recipe for getting Postgres 9.5 up and running on Travis CI.
---

**Update:** As of December 1, 2016, [Travis has made both 9.5 and 9.6 available
on Trusty containers][travis-update] so most of what's said in this article is
no longer relevant. Please refer to their standard documentation for
information on installing Postgres.

~~The recent release of Postgres 9.6 has brought into sharp relief that Travis
still doesn't even have widely available 9.5 support quite yet. I was recently
trying to get a 9.5 database running there, and was somewhat confused by the
current state of affairs, so I'm writing this. There's a [big GitHub
thread][mega-thread] on the subject, but it'll take some reading to gather
context.~~

~~So for your convenience: Postgres 9.5 is supported, but only for non-container
builds [1]. Here are the `.travis.yml` incantations necessary to get it
running:~~

``` yaml
dist: trusty
sudo: required
addons:
  postgresql: "9.5"
```

~~A few Travis employees have indicate that they're trying to get 9.5 support on
containers, but their comments were from ~1.5 months ago (as of this writing)
and there doesn't seem to have been any progress yet.~~

~~I'll try to keep this post up-to-date as I'm made aware of new developments.~~

[1] Meaning that you'll have to trade some build speed to get 9.5 support.

[mega-thread]: https://github.com/travis-ci/travis-ci/issues/4264
[travis-update]: https://github.com/travis-ci/travis-ci/issues/4264#issuecomment-263550556
