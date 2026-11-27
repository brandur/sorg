+++
hook = "It sounds like heresy to say out loud, but there's room for a new database."
published_at = 2026-11-27T11:09:35-06:00
title = "A `fork(2)` in the road: Regimented stability versus relentless progress"
+++

After the revert fiasco of Postgres 19, I was thinking about what the future holds for databases. Is inertia enough to carry the existing databases that we know so well into the future? How important is stability and pervasiveness of use compared to dynamicism and agility?

As unassailable as it is in history, pedigree, and performance, Postgres has major structural problems that'll act against it as it transitions the LLM age:

- Once per year release cycle. Reverted features don't have another chance until September 2027 at the earliest.

- Decision by consensus. New features cannot happen quickly. Large changes are fundamentally difficult to make and nearly impossible to land from anyone except known major contributors.

Postgres' weaknesses here _should_ leave SQLite standing golden: it's always had a tiny development team and doesn't accept outside contributions. The project owners could be making unilateral decisions on what new features to do, and have swarms implement them quickly and meticulously, complying to the project's famous ultra-thorough test suite with little trouble.

But there's nothing to suggest that this is happening. The SQLite team has made a couple small, solid improvements over the last year (e.g. `ALTER TABLE` can add/drop `NOT NULL` constraints), but development isn't observably faster compared to the year before. New features and improvements are decidedly not a relentless torrent.

MySQL continues its zombie existence. Technically alive under Oracle's stewardship, but spiritually dead.

Oracle, by far the most popular database on planet Earth according to extremely trustworthy and transparent db-engines dot com, continues to be perplexingly difficult in finding a single firsthand user of post-2010 or thereabouts. (Many people "know a guy".)

It sounds like heresy to say out loud, but there's room for a new database.

It's still possible that Postgres and SQLite were already far enough past the post of "good enough" that they'll be the winners even if they never change again, but imagine for a moment a new age DB with these properties:

- Core team of 1–2 people capable of taking fast, decisive action.

- Only supports archs people actually use: `arm64`, `amd64`.

- Rolling releases. Features appear in weeks not years. LLMs and LLM-generated testing is used to vet changes instead of manual QA testing + time.

- Thoroughly tested and regression protected. (SQLite does well in this area but Postgres doesn't.)

- Longstanding bugbears slain: ~unlimited idle connections. Thread-based concurrency model. For SQLite, more than one writer.

- Built-in HA and/or multi-master.

- Real, earnest work towards resolving major flaws like bloat, even if it requires significant internal rearchitecture.

- "Daemonless" mode making it operate like an embedded database like SQLite. It's an unexamined assumption that the Postgres postmaster model and the SQLite file model have to be separate databases — they could be alternate modes on the same project!

- Built-ins for reasonable OLAP performance without reaching for an external product.

- Continuous performance and stability improvements driven by LLM refinement.

Trust me, I know: intuitively, the idea of displacing Postgres feels _crazy_.

But then I think back on how much change we've seen in just the last year and it makes me think: imagine another 10 years at a similar rate. The idea that Postgres would _not_ be displaced is crazy.

But nothing's certain — another likely possibility is that the two widespread institutional options, Postgres and SQLite, become even more widespread and more cemented as LLMs cargo cult them and LLM steerers don't revisit the decision. Postgres and SQLite passed the baseline of "good enough" a long time ago, and there's an argument to be made that no new feature added to a database is sufficiently worth it to be messing around with experimental systems.

I don't know what will happen, but I go forward with an open mind and open eyes.
