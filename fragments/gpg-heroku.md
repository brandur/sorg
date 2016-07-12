---
title: GPG + the Heroku CLI
published_at: 2014-12-11T21:53:11Z
---

Luckily, this one is easy, but I figured I'd put a placeholder here just in
case anywhere is curious as to how this works.

The Heroku CLI reads and writes to the same `.netrc` format that Curl uses with
the [netrc gem][netrc]. Support for GPG has been contributed to it in that if a
`.netrc.gpg` is present, it will shell out to GPG to read from it.

[netrc]: https://github.com/heroku/netrc
