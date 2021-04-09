+++
hook = "I expect to soon have a lot more Postgres in my life."
published_at = 2021-04-09T15:36:59Z
title = "Joining Crunchy Data"
+++

I woke up a few months ago, and realized that I've spent more years professionally as a Mongo user than as a Postgres user. Sufficed to say, it wasn't a good moment, so I'm changing things up. This month, I'm joining [Crunchy Data](https://www.crunchydata.com/) to do database-related work full time.

It's always a little fraught to name names, because of course I'm going to miss somebody, but I'm going to name a few anyway (with apologies to anyone I've forgotten). Once again, I'll be working adjacent to old colleagues [Craig](https://www.craigkerstiens.com/), [Farina](https://github.com/fdr), and [Will](https://github.com/will), who are currently hard at work implementing [Crunchy Bridge](https://www.crunchydata.com/products/crunchy-bridge/), one of the first cloud-agnostic hosted Postgres solutions. I worked with them originally at Heroku, where they made up a key core of the Heroku Postgres team (affectionately known as the "DoD", or Department of Data), and admired their work from afar as they subsequently helped build [Citus Cloud](https://www.citusdata.com/product/cloud) [1].

For quite some time, Crunchy's always been one of the biggest supporters of open-source Postgres development, employing key committers like [Peter Geoghegan](https://twitter.com/petervgeoghegan) and [Tom Lane](https://en.wikipedia.org/wiki/Tom_Lane_(computer_scientist\)), thereby having directly funded major recent improvements like [B-tree deduplication](https://commitfest.postgresql.org/27/2202/) and [bottom-up index deletion](https://commitfest.postgresql.org/30/2757/), along with about [half of the total mailing list activity](https://twitter.com/regardstomlane). Other committers hold active positions at the company including [Stephen Frost](https://twitter.com/net_snow) and [Joe Conway](https://www.joeconway.com/), along with Postgres core team and community members like [Jonathan Katz](https://twitter.com/jkatz05). (Again, I don't know the company's topology yet, so sorry to all the people I missed.)

I'm excited. There's an oft-cited trope in technology circles that products are _only_ about execution -- that your choice of database or programming language doesn't matter. Having seen reams of evidence to the contrary first hand, count me as a firm disbeliever. Instead of working with strong constraints, ACID, and rich data types, I've spent the last few years building expertise on how to not break things in a schemaless world, how to build applications without transactions (or put otherwise, how to mitigate collateral damage), and how to repair non-relational models that [should just have been relational in the first place](https://stripe.com/blog/online-migrations).

With enough time, effort, and discipline, it's possible to succeed regardless of your platform choices -- it's theoretically possible to write the next Unicorn in Cobol -- but there are technologies out there that _encourage_ success rather than just permitting it, and Postgres is one of the best of them.

Of course, it's been an interesting run at Stripe, and I'll miss many aspects of the company. In the coming weeks, I'll post a few reflections to [Nanoglyph](https://nanoglyph-signup.brandur.org/).

[1] Later acquired by Microsoft, and now officially known as ["Hyperscale (Citus) on Azure Database for PostgreSQL"](https://www.citusdata.com/product/hyperscale-citus/).

