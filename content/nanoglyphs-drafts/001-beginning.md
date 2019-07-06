+++
published_at = 2019-07-05T23:08:03Z
title = "The Beginning is the Beginning is the Beginning"
+++

![Tanks and eucalyptus](/assets/images/nanoglyphs/001-beginning/midtown@2x.jpg)

You've heard the old aphorism that perfect is the enemy of
good. I (and many others) get good mileage out of it
because it's so widely appropriate, and true for so many
things in life. When I'm writing, it's _really_ true. Every
bone in my body wants to find the perfect word for every
position and be so exhaustive that four paragraphs swells
to become a multi-chapter behemoth. Everything takes a long
time. Worse yet, the product is often worse for it -- too
long, too dry.

I've been thinking about ways to fight the effect, and one
that I wanted to try is a weekly newsletter. A time
constraint can be a powerful thing -- as many of us know
from software planning it's easy to produce some that are
too optimistic, or even punitively ambitious (see the games
industry), but deployed realistically, they can force good
results in a timely manner. I've always been impressed by
Eiichiro Oda's work on [One Piece][onepiece] (a manga about
pirates) -- he's been writing/drawing a chapter a week
since 1997 and only missed a few dozen weeks since. He
didn't miss a week until year four, and even then, only one [1].

For me, an immaculate track record like that is never going
to happen, but we'll see how it goes. I'm not even send the
first ten issues just to make sure that getting into a
weekly cadence is semi-realistic. Writing [_Passages &
Glass_](/newsletter) was one of my more rewarding personal
projects of the last couple of years, so I figure that the
worst case scenario is another semi-defunct writing
project.

Welcome to _Nanoglyph_.

---

My plan right now is to send out roughly three links a
week, but because _just_ links are boring, they'll be
flavored with my own opinionated editorial. Topics will be
largely software-focused and include my usual areas of
interest like databases, cloud services, and programming
languages, and with a philosophical skew towards building a
better software future that's simpler, safer, and more
sustainable.

I'd love to eventually include a "mail" column with content
from replies to previous newsletters that's curated in the
spirit of _The Economist_'s with replies that are civil,
edited for brevity, and which don't necessarily agree with
the original material. Strong opinions should be met with
strong opinions after all. This of course depends on (1)
the project getting off the ground, and (2) having enough
readers that there are replies to include.

Okay, with all that said, onto this week's content.

---

GitLab [announced that they were ending support for
MySQL][gitlabmysql], meaning where before it was possible
to run GitLab installations on either MySQL or Postgres,
now only Postgres will be supported. Supporting multiple
databases is a game of lowest common denominator --
juggling the features common to both, papering over
inconsistencies, and never being able to leverage the
strengths of either.

Distant observers like myself were surprised to find that
they supported MySQL in the first place, because that meant
they'd been getting away all these years without features
like [partial indexes][partial], which are useful all the
time, but _especially_ useful for massive database
installations where they can vastly reduce the required
space for indexes on large tables. GitLab is a massive
database installation.

The announcement routed a horde of MySQL advocates from the
woodwork who protested that GitLab's complaints about the
technology weren't real limitations, and that what the move
really showed was a lack of MySQL expertise on the GitLab
team. But reason [seemed to win out in the
end][mysqlcomment] as people _with_ MySQL expertise cited a
history of underdesign that still dogs it today. It's
finally possible to drop columns in production [2] and
finally possible to get [real unicode
support][mysqlunicode] (just make sure to use `utf8mb4`
instead of `utf8`), but still faces a bug where [triggers
sometimes don't fire][mysqltriggers] that's been open for
so long that it'll be old enough to drive in the next few
years, and still has some surprising transaction semantics.

## Boring technology (#boring technology)

A link to the excellent talk [Choose Boring
Technology][boring] resurfaced this week. I hadn't fully
read this one before, but found myself nodding throughout.
New technology is expensive -- in the short-term to get up
and successfully running on it

I often wonder about how history would've been different if
boring technology had been more en vogue at the time the
company I work at today had been making tech stack
decisions. 

[1] Someone on Reddit made [a chart of One Piece chapters
by week][onepiecechart]. It's an impressive sight.

[2] For a long time it wasn't possible to drop a column
without taking an exclusive lock on the table for the
duration of the rewrite. Our DBA-approved solution for this
limitation at iStock was to never drop columns.

[boring]: http://boringtechnology.club/
[gitlabmysql]: https://about.gitlab.com/2019/06/27/removing-mysql-support/
[mysqlcomment]: https://news.ycombinator.com/item?id=20345204
[mysqltriggers]: https://bugs.mysql.com/bug.php?id=11472
[mysqlunicode]: https://medium.com/@adamhooper/in-mysql-never-use-utf8-use-utf8mb4-11761243e434
[onepiece]: https://en.wikipedia.org/wiki/One_Piece
[onepiecechart]: https://i.redd.it/l7leyqae5hy01.png
[partial]: https://www.postgresql.org/docs/current/indexes-partial.html
