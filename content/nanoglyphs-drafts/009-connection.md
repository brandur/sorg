+++
image_alt = "A snow drift and blue sky in Calgary."
image_url = "/assets/images/nanoglyphs/009-connection/hoarfrost@2x.jpg"
published_at = 2020-01-02T07:13:58Z
title = "Connection"
+++

I’m back in San Francisco after the holidays. As soon as I left Calgary, the weather there took a turn for the worse with lows around -35C. My narrow escape was extra fortunate as my body, now acclimated to moderate Mediterranean-esque climate conditions, would’ve buckled under the stress. But I do miss the snow. Depicted above is a small sample of some of the beautiful hoarfrost that the region sees regularly.

If you’re wondering what in the name of all that is holy you’re reading right now, it’s a weekly newsletter called _Nanoglyph_, loosely written around the themes of simple and sustainable software. You may have signed up recently after reading the last edition on [Actix Web](/nanoglyphs/008-actix) and the optimizations it’s put in place to lead the TechEmpower benchmarks. If you want to eject and never see it again, you can [unsubscribe right here](%unsubscribe_url%).

---

The 008 edition of this newsletter experienced a bit of a traffic surge, and a larger than normal number of people ended up clicking through to [its signup page](https://nanoglyph-signup.brandur.org/). An hour or so later a reader (thanks Thomas!) emailed me to say that trying to use the app was producing that detestable error message:

```
FATAL: remaining connection slots are reserved for
non-replication superuser connections
```

How embarrassing. I’ve spent a lot of time talking about how connection use is the single largest operation caltrop Postgres users are likely to find lodged in their foot, and gone so far as to write a detailed article on [connection management](/postgres-connections). I learnt my lesson ten years ago, and should know how to avoid this problem by now.

The signup page is a tiny Go app that talks to the Mailgun API. It was originally stateless itself (containing just a Mailgun API key), but I eventually realized that even with easy to find unsubscribe links in dispatched emails, it was pretty negligent not to have a double-opt in [1] process for signups to curb abuse. The database temporarily tracks the state of a partial signup as a user receives a confirmation email and clicks through to a unique URL to complete it.

It’s an ultimate NIH setup which I’d normally strongly advocate against, but it does save a hefty newsletter service subscription fee, gives me perfect creative control over the final look and layout of emails and over what kind of creepy tracking they embed (or in the case of _Nanoglyph_, _do not_ embed).

But back to the problem at hand, the reason my connections were depleted was a confluence of a few things:

* The database I’m running is a free hobby-dev on Heroku, with a connection limit of 20.
* Although I’m on the cheapest of the cheap as far as Postgres plans go, overall connection limits on even much more expensive hosted plans tend to be low.
* Go’s connection pool defaults are very aggressive — by default it’ll open unlimited connections and reuse opened connections in perpetuity.
* I’d recently moved the app to Google Cloud Run, and unlike my old setup on Heroku, it might spawn an arbitrary number of running containers depending on incoming traffic (more on this an upcoming edition).

A side effect of connection exhaustion on a managed database is that you yourself can’t open a psql session to look at `pg_stat_activity` [2] and understand what’s going on, and because of that I couldn’t find the smoking gun, but it was probably Go. My reading of `database/sql`’s docs would seem to suggest that although it’ll open an unlimited number of connections, once connections and no longer being used and the maximum idle threshold is surpassed (default: 2), some of those connections should have been released, but apparently were not.

Configuring Go’s connection pool with a maximum number of allowed connections and a maximum lifetime fixed the problem:

```
db.SetConnMaxLifetime(60 * time.Second)
db.SetMaxOpenConns(5)
```

The side effect is that during busy periods a request may have to wait for a connection to become available from the pool, but that’s a good trade off for better operational stability. All database transactions in the app are short, so even in the worst case scenarios, requests will never wait long.

---

Bad defaults

Turns out most connections are idle, but managed Postgres providers set their configurations with the worst case scenario in mind.

PgBouncer

---

[1] “Double opt in” is email marketing jargon that means a user opted in once by entering an email, then opted in again by confirming it via link sent to it. Helps confirm that entered emails are valid and to curb abuse.

[2] `pg_stat_activity` is a system table in Postgres that can be queried to show information on active Postgres backends (i.e. connections) including current query, time spent in current query, or whether they’re idle.
