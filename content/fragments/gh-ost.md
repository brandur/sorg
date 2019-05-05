---
title: Gh-ost
published_at: 2016-08-03T07:59:34Z
hook: UNWRITTEN. This should not appear on the front page.
---

Yesterday, released [Gh-ost][gh-ost], a tool to help makes changes to an online
MySQL database.

As some officianados may know, in MySQL, changes to a database's schema or
indexes somewhat infamously require a full table lock, a sizable complication
that prevents any other operation from running simultaneously. Because any
change to a non-trivial table could take minutes or more, in practice those
types of changes are avoided.

Companies address the problem in different ways. At iStockphoto, we took the
easiest possible route: don't introduce new indexes or drop old columns. You
can imagine how well that worked. GitHub suggests the use of Gh-ost, a rather
elaborate tool that copies a table's schema to a new "ghost" table, makes
required schema changes, backfills, and then cuts over after the replacement is
ready.

An ex-colleague of mine posted this glib reaction to Gh-ost's release:

[![MySQL vs. Postgres](/assets/images/fragments/gh-ost/vs.jpg)](/assets/images/fragments/gh-ost/vs@2x.jpg)

That might look like a simplification, and it is, but not by much. Postgres
uses clever mechanics that allow the addition or removal of either a column or
an index a non-blocking operation so that a live database can stay online
through these changes without interruption [1]. There are people out there
making tweaks to their Postgres schemas that contain data sets as large as
GitHub's _every day_ without any assistance from specialized tools. I know this
because I was one of them.

These impressive features allowing non-blocking operations weren't always
there, but rather added incrementally throughout Postgres' development by
hard-working members of the community. For example, it wasn't possible to drop
an index concurrently in 9.1, but it was by 9.2. The same advancements never
happened in MySQL, and complex and operationally expensive systems like Gh-ost
are required to compensate.

I understand that sometimes you're stuck on a legacy system, and sometimes that
may require the development and application of solutions that would be
inadvisable under normal conditions, but GitHub's write-up on Gh-ost doesn't
mention anywhere that the project mainly exists to address deficiencies in
MySQL, and that there might be a better way if you're not already locked into a
big MySQL cluster. They'll even go out of their way [to champion MySQL over
better alternatives][vmg] despite being seemingly aware how problematic its use
can be. Not ackowledging the expensive trade-offs of a MySQL/gh-ost setup
strikes me as somewhat amiss given the number of readers (and by extension
influence) an article like this one will inevitably have.

**Addendum --** After a conversation with the original author of the
article on gh-ost, I feel bad about the aggressive tone that this article
originally conveyed and have tried to correct my hubris by amending it
accordingly. Changes are visible in its [Git history][history].

[1] There are some limitations to online schema changes in Postgres, most
    notably that adding a column with a `DEFAULT` clause or changing a column's
    type will require a rewrite. These are described in detail in the Postgres
    documentation for [ALTER TABLE][alter-table-notes].

[alter-table-notes]: https://www.postgresql.org/docs/9.6/static/sql-altertable.html#AEN75201
[gh-ost]: https://github.com/github/gh-ost
[history]: https://github.com/brandur/sorg/commits/master/content/fragments/gh-ost.md
[vmg]: https://twitter.com/vmg/status/757987482478776320
