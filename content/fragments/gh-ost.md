---
title: Gh-ost
published_at: 2016-08-03T07:59:34Z
---

Yesterday, GitHub broadcasted an indomitable sense of self-satisfaction far and
wide with the release of [Gh-ost][gh-ost], a tool to help makes changes to an
online MySQL database.

In MySQL, changes to a database's schema or indexes somewhat infamously require
a full table lock which prevents other operations from running simultaneously.
Because any change to a non-trivial table could take minutes or more, that
means in practice you don't do them.

Companies address the problem in different ways. At iStockphoto, we would
simply use the easiest possible route: just don't introduce new indexes or drop
old columns. You can imagine how well that worked. GitHub suggests the use of
Gh-ost, a rather complex tool that copies a table's schema to a new "ghost"
table, makes required schema changes, backfills, and then cuts over after the
replacement is ready.

An ex-colleague of mine posted this glib reaction to Gh-ost's release:

[![MySQL vs. Postgres](/assets/fragments/gh-ost/vs.jpg)](/assets/fragments/gh-ost/vs@2x.jpg)

That might look like a simplification, but the truly offensive part about all
of this is that it's not. Postgres uses clever mechanics that allow the
addition or removal of either a column or an index a non-blocking operation so
that a live database can stay online through these changes without
interruption. There are people out there making tweaks to their Postgres
schemas that contain data sets as large as GitHub's _every day_ without any
assistance from specialized tools. I know this because I was one of them.

These impressive features allowing non-blocking operations weren't always
there, but rather added incrementally throughout Postgres' development by
hard-working members of the community. For example, it wasn't possible to drop
an index concurrently in 9.1, but it was by 9.2. The same advancements never
happened in MySQL, and hugely complex and operationally expensive systems like
Gh-ost are required to compensate.

I understand that sometimes you're stuck on a legacy system, and sometimes that
may require the development and application of solutions that would be
inadvisable under normal conditions. But what kills me here is that there isn't
one place throughout GitHub's write-up on Gh-ost where they mention that the
entire project was built to address deficiencies in MySQL, and that there might
be a better way if you're not already locked into a big MySQL cluster. In fact,
they'll even go out of their way [to champion MySQL over better
alternatives][vmg] despite being fully aware how problematic its use can be.
There is a level of willful misdirection here that I just can't wrap my head
around.

[gh-ost]: https://github.com/github/gh-ost
[vmg]: https://twitter.com/vmg/status/757987482478776320
