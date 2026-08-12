+++
hook = "Almost every notable managed Postgres provider bundles a connection pooler, giving us a strong hint that connection management is a product problem rather than a user problem."
published_at = 2026-08-12T11:37:03-05:00
title = "Does anyone run Postgres without PgBouncer?"
+++

I got a nice shout-out from [Ben Dicken over the weekend](https://x.com/BenjDicken/status/2086550652776595655) on an old article I'd written on [managing database connections](/postgres-connections). (This guy is apparently the Mick Jagger of databases, because I can't remember having gotten so many inbound LinkedIn invitations in one day before.)

Something that hit hard is that I wrote this article almost _ten years ago_.

Just as striking is that as I was reading back through it, I realized that despite being a decade old, it's still pretty much up to date. Postgres is still, shall we say, not _great_ at managing lots of connections, so you want to use local connection pools, short-term checkouts, and a pooler like PgBouncer.

It got me wondering: _how_ standard is it to use a pooler, exactly? To answer that question, I made a table of all managed Postgres providers with household notoriety and whether they support PgBouncer, something close to PgBouncer, or no connection pooling at all.

| Provider                 | Pooler?  | Implementation                  | Availability / caveat                  |
| ------------------------ | :------: | ------------------------------- | -------------------------------------- |
| **Aiven**                |    ✅    | **PgBouncer**                   | Startup plans and above                |
| **Alibaba RDS**          |    ✅    | **PgBouncer**                   |                                        |
| **AWS RDS / Aurora**     |    ✅    | **RDS Proxy**                   | Separate managed proxy service         |
| **Azure PG**             |    ✅    | **PgBouncer**                   |                                        |
| **Crunchy Bridge**       |    ✅    | **PgBouncer**                   |                                        |
| **DigitalOcean**         |    ✅    | **PgBouncer**                   |                                        |
| **EDB Postgres AI**      |    ✅    | **PgBouncer**                   |                                        |
| **Fly.io <acronym title="Managed Postgres">MPG</acronym>** |    ✅    | **PgBouncer**                   |                                        |
| **Google Cloud SQL**     |    ✅    | **PgBouncer / managed pooling** | Requires Enterprise Plus               |
| **Heroku**               |    ✅    | **PgBouncer**                   | Some plans only                        |
| **IBM Cloud**            |    ❌    | —                               | Self-managed only                      |
| **Neon**                 |    ✅    | **PgBouncer**                   |                                        |
| **OCI (Oracle)**         |    ❌    | —                               | No managed pooler                      |
| **PlanetScale**          |    ✅    | **PgBouncer**                   |                                        |
| **Railway**              |    ✅    | **PgBouncer**                   | Added as separate service              |
| **Render**               |    ✅    | **PgBouncer**                   | On paid databases                      |
| **Supabase**             |    ✅    | **PgBouncer** or **Supavisor**  | PgBouncer or Supavisor (proprietary pooler) for serverless |
| **Tiger Cloud**          |    ✅    | **PgBouncer**                   |                                        |

Not only is PgBouncer support widespread, but we see above that the overwhelming majority of providers bundle it out of the box. I'd go a step further -- since neither IBM nor Oracle is a service that any self-respecting person not part of an enterprise sales cycle would actually use, _one hundred percent_ of plausible managed Postgres providers bundle a pooler.

## If everyone needs it, is it really a non-core function? (#non-core)

In some ways, it could be argued that this status quo is okay. Users that need a connection pooler have access to one, and can use it to keep prod stable.

But there's undoubtedly a lot of wasted effort here. Every provider has had to come up with their own homegrown mechanism for getting multiple components set up and configured and establish a convention for where to find Postgres versus its bouncer. Every user needs to [reference a guide](https://planetscale.com/docs/postgres/connecting/pgbouncer#when-to-not-use-pgbouncer) explaining PgBouncer's limitations (e.g. don't listen/notify) and read about its [pooling modes and tradeoffs](https://www.pgbouncer.org/config.html#pool_mode).

Imagine if you went to your local car dealership and they sold you a car without a windshield. On the way over you'd noticed that 100% of vehicles on the road did in fact have windshields, and for good reason because it turns out to be pretty dangerous to drive without one. Since you were the one that bought the car, it'd be hard to argue that it's not your responsibility now to outfit it with a windshield before it's roadworthy, but it'd _also_ be fair to later be pissed off at the dealer for selling a vehicle that can't just be driven off the lot.

## Reintegration (#reintegration)

What if there was a world where you went to your favorite Postgres provider and you got one database URL, one port, and no extra configuration or caveats to worry about? Your managed provider doesn't need to add an aftermarket windshield because one came with the car already. We know a place like this can exist because that's already how things work in MySQL and Mongo-land.

There are reasons it doesn't happen, like reviving the age-old processes versus threads debate, which very few contributors are venerated enough to push for progress on, but given the developer-years' worth of effort in working around Postgres' lack of connection pooling, it's hard to argue this wouldn't be one of the highest-impact operational improvements possible.
