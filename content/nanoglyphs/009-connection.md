+++
image_alt = "A snow drift and blue sky in Calgary."
image_url = "/assets/images/nanoglyphs/009-connection/hoarfrost@2x.jpg"
published_at = 2020-01-31T20:36:11Z
title = "Connection"
+++

I’m back in San Francisco after the holidays. As soon as I left Calgary, the weather there took a turn for the worse with lows around -35C. My narrow escape was extra fortunate as my body, now acclimated to moderate Mediterranean-esque climate conditions, would’ve buckled under the stress. But I do miss the snow. Depicted above is a small sample of some of the beautiful hoarfrost that the region sees regularly.

If you’re wondering what in the name of all that is holy you’re reading right now, it’s a weekly newsletter called _Nanoglyph_, loosely written around the themes of simple and sustainable software. You may have signed up recently after reading the last edition on [Actix Web](/nanoglyphs/008-actix) and its optimizations a few weeks ago. If you want to eject and never see it again, you can [unsubscribe right here](%unsubscribe_url%).

---

The 008 edition of this newsletter experienced a bit of a traffic surge, and a larger than normal number of people ended up clicking through to [its signup page](https://nanoglyph-signup.brandur.org/). An hour or so later a reader (thanks Thomas!) emailed me to say that trying to use the app was producing that detestable error message familiar to Postgres users around the world:

```
FATAL: remaining connection slots are reserved for
non-replication superuser connections
```

How embarrassing. I’ve spent a lot of time talking about how connection use is the single largest operation caltrop Postgres users are likely to find lodged in their foot, and gone so far as to write a detailed article on [how to manage connections](/postgres-connections). I learnt my lesson ten years ago, and it should've been perfectly crystallized by now, but I still find myself relearning it periodically.

The signup page is a tiny Go app that talks to the Mailgun API. It was originally stateless itself (containing just a Mailgun API key), but it eventually dawned on me that even with easy to find unsubscribe links in dispatched emails, it was pretty negligent not to have a double-opt in [1] process for signups to curb abuse. The database temporarily tracks the state of a partial signup as a user receives a confirmation email and clicks through to a unique URL to complete it.

![Screenshot of Nanoglyph signup app](/assets/images/nanoglyphs/009-connection/nanoglyph-signup@2x.jpg)

It’s an ultimate NIH setup which you should never ever do, but it does save a hefty newsletter service subscription fee, conveys perfect creative control over the final look and layout of emails and over what kind of creepy tracking they embed (or in the case of this newsletter, do not embed).

But back to the problem at hand, the reason my connections were depleted was a confluence of a few things:

* The database I’m running is a free hobby-dev on Heroku, with a connection limit of 20.
* Although I’m on the cheapest of the cheap as far as Postgres plans go, overall connection limits on even much more expensive hosted plans tend to be low.
* Go’s connection pool defaults are very aggressive -- by default it’ll open unlimited connections and reuse opened connections in perpetuity.
* I’d recently moved the app to Google Cloud Run, and unlike my old setup on Heroku, it might spawn an arbitrary number of running containers depending on incoming traffic (more on Cloud Run an upcoming edition).

A side effect of connection exhaustion on a managed database is that you yourself can’t open a psql session to look at `pg_stat_activity` [2] and understand what’s going on, and because of that I couldn’t prove the smoking gun, but it was probably Go. My reading of `database/sql`’s docs would seem to suggest that although it’ll open an unlimited number of connections, once connections and no longer being used and the maximum idle threshold is surpassed (default: 2), some of those connections should have been released, but apparently were not.

Configuring Go’s connection pool with a maximum number of allowed connections and a maximum lifetime fixed the problem:

```
db.SetConnMaxLifetime(60 * time.Second)
db.SetMaxOpenConns(5)
```

The side effect is that during busy periods a request may have to wait for a connection to become available from the pool, but that’s a good trade off for better operational stability. All database transactions in the app are short, so even in the worst case scenarios, requests don't wait long.

---

## First contact (#first-contact)

It’s general wisdom that no software survives first contact with production, but the sheer frequency of the connection problem makes it particularly egregious. It’s one that every user of Postgres is going to run into at some point, probably sooner than later, and most likely the first time they see real traffic, making an underwhelming launch an extraordinarily likely outcome. This seems like a problem that better defaults and better APIs could help with.

A major improvement that we’ve been trending toward over the last decade across many language and projects is making connection pools the default way to talk to databases. In the case of database connections, the pool doesn’t only serve as an optimization to reuse connections, but also to utilize them efficiently and to control their upper bound. Think back to Rails where the default for a long time (and still is where process-based concurrency is used) was to have each worker operate in isolation and open a connection that it never closed. Consider that most workers are idle most of the time, and you can see how inefficiently connections are used.

But Go’s `database/sql` uses a connection pool, and still didn’t save me. Go’s problem is that it has the right options, but they’re easy to miss. They’re not part of the constructor, and if you read the package’s documentation, they’re mentioned well below the fold and only as part of the index list of all functions. Preferring the use of a builder pattern, like the one provided by [`sqlx`](https://github.com/launchbadge/sqlx), helps to remind users that they should look at configuration beyond the connection string:

``` rust
let pool = sqlx::PgPool::builder()
    .max_size(5)
    .build(“postgres://localhost:5432/mydb”)
    .await?;
```

## On the backend (#backend)

There’s potential room for improvement on the Postgres end as well. Managed installations tend to be conservative with connection numbers with an eye for the worst case scenario -- if all those connections really get used, it can create contention in the database and slow its workload. But often in real world applications many connections are idle, and increasing the maximum number allowed may have some impact on memory, but not a serious one on performance.

This is why so many Postgres users have so much success with connection poolers like [PgBouncer](https://github.com/pgbouncer/pgbouncer) and [Odyssey](https://github.com/yandex/odyssey). They let applications hold huge numbers of open connections, and remap them to a relatively fewer number of active connections to the database. Their use is so ubiquitous for serious uses of Postgres that there’s a strong argument that it should provide something equivalent out of the box, and indeed [some effort has been made to that end](https://www.postgresql.org/message-id/ac873432-31cf-d5e4-0b80-b5ac95cfe385@postgrespro.ru), though it’s not yet clear whether the project’s gatekeepers will accept it.

---

## Orchidaceae habits (#orchidaceae-habits)

I’ve been enjoying reading about caring for Phalaenopsis orchids. Even if that name isn’t familiar, it’s the one orchid that everyone has surely seen before as it’s the one type of orchid regularly on sale everywhere from Safeway to IKEA. Here’s one of mine.

![A specimen of Phalaenopsis](/assets/images/nanoglyphs/009-connection/phalaenopsis@2x.jpg)

Most owners (myself included) are guilty of treating them as disposable objects. They look nice for a few months until their flowers fall off, don’t immediately grow more, and so are replaced with fresh plants bought new. Those of us who’ve tried to get them to re-flower don’t have an easy time of it, and generally give up on the project.

An aficionado who grows orchids in his downtown Calgary (purely coincidentally, also my home city) condominium writes some of the most [comprehensive Phalaenopsis care articles](http://herebutnot.com/how-we-grow-orchids-in-calgary-alberta-canada/) there are. These multi-thousand word affairs include, among other things, advice to repot new plants immediately on purchase to avoid root rot, weekly watering regiments with a very specific fertilizer mix, and “leaching” to avoid the build up of hard water minerals that lead to high pH. His results are wonderful, with the plants showing their prosperity through majestic size.

The complete sets for total orchid care are daunting, but I’m starting with a couple big ones this weekend. Sphagnum moss in hand, I’ll be repotting my couple Phalaenopses, then moving onto a complete watering/fertilization and flushing cycle.

---

I’ll give you my usual reminder that if you have feedback or ideas for future material, hit “reply” and send them my way. Until next week, and remember to keep a sharp eye on those connection counts.

[1] “Double opt in” is email marketing jargon that means a user opted in once by entering an email, then opted in again by confirming it via link sent to it. Helps confirm that entered emails are valid and to curb abuse.

[2] `pg_stat_activity` is a system table in Postgres that can be queried to show information on active Postgres backends (i.e. connections) including current query, time spent in current query, or whether they’re idle.
