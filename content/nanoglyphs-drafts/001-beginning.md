+++
published_at = 2019-07-05T23:08:03Z
title = "The Beginning is the Beginning is the Beginning"
+++

![Steps from Baker Beach](/assets/images/nanoglyphs/001-beginning/steps@2x.jpg)

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
to happen, but we'll see how it goes. I'm not intending to
send the first five issues just to make sure that getting
into a weekly cadence is semi-realistic. Writing [_Passages
& Glass_](/newsletter) was one of my more rewarding
personal projects of the last couple of years, so I figure
that the worst case scenario is another semi-defunct
writing project.

Welcome to _Nanoglyph_.

---

My plan right now is to send out roughly three links a
week, but because _just_ links are boring, they'll be
flavored with opinionated editorial. Topics will be largely
software-focused and include my usual areas of interest
like databases, cloud services, and programming languages,
and with a philosophical skew towards building a better
software future that's simpler, safer, and more
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

## Boring technology (#boring-technology)

A link to the excellent talk [Choose Boring
Technology][boring] resurfaced this week. I found myself
nodding throughout as I read it. New technology is
expensive -- in the short-term to get up and successfully
running on it, and in the longer term experiencing its
bumpy road to maturity.

More boring technology choices would've saved my current
company countless dollars and engineering hours that have
been sunk into working around the deficiencies of our
non-boring technology. Maybe there's something to be said
for living an adrenaline-fueled life in the fast lane --
non-boring technology advocates are the BASE jumpers of the
software world -- but these days I'm a boring technology
engineer through and through.

## Rust ascending (#vector)

This week Timber released [Vector][vector], a router for
observability data like metrics and logs. It's common in
industry to have a daemon (and often more than one) running
on each server node that's responsible for forwarding logs
and metrics to central aggregators for processing. At
Stripe for example, we have locals daemons that forwards
logs off to Splunk, and our custom [Veneur][veneur] that
collects and forwards metrics to SignalFx. Vector tries to
be a consolidated solution by handling metrics and logs,
supporting a number of backends that can be forwarded to,
and offering a number of configurable transformations which
allow basic filtering all the way up to custom Lua scripts.

Notably, it's written in [Rust][rust] which conveys a
number of benefits -- easy deployment by just shipping a
single static binary out to servers, more efficient and
more predictable use of memory as there's no garbage
collector involved, and likely fewer bugs as the they're
ferreted out by the language's compiler and type system
well before release.

Normally, I'd say that the less software in the world the
better, and that it'd be better to use the multitude of
existing products with similar features, but between
Vector's promise to consolidate many services into a
single, more flexible daemon, and the leverage it's going
to pick up by being written in Rust, I'm really hoping to
see the project succeed.

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
[rust]: https://www.rust-lang.org/
[vector]: https://github.com/timberio/vector
[veneur]: https://github.com/stripe/veneur
