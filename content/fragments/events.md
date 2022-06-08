+++
hook = "If you're going to have an events table, delete old stuff, and do so in batch queries with a `LIMIT`. Consider partitioning."
published_at = 2022-05-02T00:33:09Z
title = "There's always an events table"
+++

Seemingly, an invariant of SaaS products: there's _always_ an events table.
	
Reading a blog post [from Retool](https://retool.com/blog/how-we-upgraded-postgresql-database/) recently, I was entertained to read about their events (AKA audit log) table:

> The larger 2 TB table, an append-only table of audit events in the app, was easy to transfer: we waited until after the cutover to migrate the contents, as the Retool product functions just fine even if that table is empty.

Which naturally, was the biggest table in their database. It's always the biggest table in the database.

## Event manager (#event-manager)

At Heroku, we put one in early in an ad-hoc way without really understanding the ramifications of doing so. It wasn't used for anything for the longest time, but in the meantime caused its fair share of operational pain. I distinctly remember Xmas 2014 where I spent a good number of my nights looking into why Postgres `DELETE` queries trying to evict old events were timing out, and leading to other knock on problems like [an unstable job queue](/postgres-queues). Postgres isn't particularly good at making large changes on large tables, and this was compounded by the `DELETE` operation being unbounded.

Back around then it was about 1 TB in size. Not exactly web scale, but 10x the size of the rest of the database combined, and all data that no one ever looked at or used.

Besides being wasteful, a reason that's bad is that if (heaven forbid) you ever had to restore your database from backup, it would take 10x longer because of non-critical data which no one would even miss if it wasn't there.

## Endless notifications (#endless-notifications)

Stripe exposes its events by way of [public API](https://stripe.com/docs/api/events), and which is the same abstraction that used to produce the payloads for its webhooks.

The API docs have claimed a retention of only 30 days since the day it was put in, but that was never true until very recently, and the fact that no one bothered adding a cleaner in those early days was a mistake that would end up costing absolutely embarrassing amounts of time and resources. It was used a little more often than our Heroku event log, but something like 99% of accesses were on just the very latest data, leaving untold terabytes to sit there untouched in hot storage on costly Mongo servers.

## Do it as right as you can (#as-right-as-you-can)

So I'm embarrassed to admit that one of the things I've done over the last year is add our own [event log](https://docs.crunchybridge.com/api/event/).

Here's how I justified it: it's a useful product feature in that it powers a user-visible audit log in our Dashboard where users can view things like their recent logins or who created a new database cluster on their team. We tried to keep it as efficient as possible by using [ULIDs](https://github.com/ulid/spec) as primary ID, meaning that insertions of new objects are [colocated and fast](/nanoglyphs/026-ids#uuids) like they would be for a serially incremented column [1]. We make use of appropriate data types keep the tuples as small as possible -- e.g. `uuid` over `text`, `jsonb` over `hstore`/`json`/`text`. Nothing emitted at high-frequency is stored.

Most importantly, _we delete stuff_. Events are removed automatically after 90 days, and there's a clamp on the API side that makes sure that events older than that horizon aren't returned even if the cleaner were to get behind. Making sure there's a sold cleanup story is the aspect I'd stress the most -- "we'll figure it out later" can be a costly and painful decision, whereas just doing it from the beginning takes so little time that it's practically free.

### Batch + limit (#batch-limit)

There's a certain art to implementing the cleaning process. Delete in batches (as opposed to one-by-one), and without moving the event objects out of the database server -- try to avoid loading them in the cleaning process before removing them, and don't bother sending removed items back over the wire [2].

Limit the size of each batch of deletions, so just in case the worker dies off and there's a big backlog to work through by the time it comes back, it'll always make some incremental progress instead of timing out by trying to do too much at once. The limit also keeps individual queries shorter, avoiding long-running transactions that produce feedback in the rest of the system.

Here's our SQL, which looks a tad convoluted because `DELETE` doesn't take a `LIMIT` so we use one by way of an inner `SELECT`:

``` sql
DELETE FROM event
WHERE id IN (
    SELECT id
    FROM event
    WHERE event.created_at < @created_at_horizon
    LIMIT @max
);
```

We're operating at smaller scale, but if we expected larger data sizes we'd use [table partitioning](https://www.postgresql.org/docs/current/ddl-partitioning.html). A `DELETE` statement could be slow on a huge dataset and a long-running query could snowball into trouble elsewhere, but running `DROP TABLE` on a partition that's no longer needed is extremely fast.

[1] Although admittedly, at our scale, this isn't a huge problem yet.

[2] Deleting in batch and without sending anything over the wire might sound obvious, but in my industry experience, it's not.
