+++
hook = "The `text` type in Postgres, why it's awesome, and why you might want to use `varchar` anyway. Also, a story about trying to get string parameters bounded at Stripe."
location = "San Francisco"
published_at = 2021-09-10T15:55:19Z
tags = ["postgres"]
title = "Postgres: Boundless `text` and Back Again"
+++

One of the major revelations for almost every new user to Postgres is that there's no technical advantage of specifying columns as `varchar(n)` compared to just using bound-less `text`. Not only is the `text` type provided as a convenience (it's not in the SQL standard), but using it compared to constrained character types like `char` and `varchar` carries no performance penalty. From the Postgres [docs on character type](https://www.postgresql.org/docs/current/datatype-character.html) (and note that `character varying` is the same thing as `varchar`):

> There is no performance difference among these three types, apart from increased storage space when using the blank-padded type, and a few extra CPU cycles to check the length when storing into a length-constrained column. While `character(n)` has performance advantages in some other database systems, there is no such advantage in PostgreSQL; in fact `character(n)` is usually the slowest of the three because of its additional storage costs. In most situations `text` or `character varying` should be used instead.

For many of us this is a huge unburdening, as we're used to micromanaging length limits in other systems. Having worked in large MySQL and Oracle systems, I was in the habit of not just figuring out what column to add, but also how long it needed to be -- should this be a `varchar(50)` or `varchar(100)`? `500`? (Or none of the above?) With Postgres, you just stop worrying and slap `text` on everything. It's freeing.

I've since changed my position on that somewhat, and to explain why, I'll have to take you back to Stripe circa ~2018.

## Stripe (#stripe)

One day we came to a rude awakening that we weren't checking length limits on text fields in Stripe's API. It wasn't just that a few of them weren't checked -- it was that practically none of them were. While the API framework did allow for a maximum length, no one had ever thought to assign it a reasonable default, and as a matter of course the vast majority of parameters (of which there were thousands by this point) didn't set one. As long as senders didn't break any limits around size of request payload, they could send us whatever they wanted in any field they wanted. The API would happily pass it through and persist it to Mongo forever.

I don't remember how exactly we noticed, but sufficed to say we only did when it became a problem. Some user was sending us truly ginormous payloads and it was crashing HTTP workers, tying up database resources, or something equally bad.

As far as problems in computing go, checking string lengths isn't one that's considered to be particularly hard, so we set to work putting in a fix. But not so fast -- these weren't the early days of the company anymore. We already had countless users, were processing millions of requests, and that meant by extension that we could expect many of those to include large-ish strings. We'd never had rules around lengths before, and without a hard constraint, given enough users and enough time, someone (or many someones as it were) eventually starts sending long strings. Suddenly introducing maximums would break those integrations and create a lot of unhappy users. Stripe takes backwards compatibility very seriously, and would never do something like that on purpose.

Already fearing what I was about to find, I went ahead and put a probe in production that would generate statistics around text field lengths, including upper bounds and distribution, and waited a day to gather data.

It was even worse than we'd thought -- we had at least hundreds of users (and maybe thousands, my memory is bad) who were sending huge text payloads. Worse yet, these were all legitimate users -- legitimate users who for one reason or another had decided over the years to build unconventional integration patterns. They'd be doing something like sending us their whole product catalog, or a big JSON blob to store, and as part of their normal integration flows.

We'd occasionally engage in active outreach campaigns to get users to change something, but it's a massive amount of work, and we have to offer generous deprecation timelines when we do. Given the nature of this problem and the number of users involved, it wasn't worth the effort. My dream of constraining most fields like customer or plan name to something reasonable like "only" 200 characters was a total non-starter.

Instead, we ran the numbers, and came up with a best fit compromise that would leave the maximum numbers of users unaffected while still bounding fields text fields to something not completely crazy (the chosen number was 5000, as viewable in the [public OpenAPI spec](https://github.com/stripe/openapi)). And even the new very liberal limit was too long for a few users sending us giant payloads, so we gated them into an exemption.

Let me briefly restate Hyrum's law:

> With a sufficient number of users of an API, it does not matter what you promise in the contract: all observable behaviors of your system will be depended on by somebody.

Truer words have rarely been spoken.

## varchars considered ~harm~helpful

Starting my [new position back in April](/nanoglyphs/024-new-horizons), one thing I checked early on is whether we were checking the length of strings that we were passing on through to the database. Nope. It turns out that this is a _very_ easy mistake to make.

This is a downside to the common Postgres wisdom of "just use `text`". It's generally fine, but there are ramifications at the edges that are harder to see.

I've gone back to the habit of making most text fields `varchar` again. But I still don't like micromanaging character lengths, or how after a while every `varchar` column has a different length seemingly picked at random, so I've pushed that we adopt some common order of magnitude "tiers". For example:

* `varchar(200)` for shorter-length strings like names, addresses, email addresses, etc.
* `varchar(2000)` for longer text blocks like descriptions.
* `varchar(20000)` for really long text blocks.

The idea is to pick _liberal_ numbers that are easily long enough to hold any even semi-valid data. Hopefully you never actually reach any of these maximums -- they're just there as a back stop to protect against data that's wildly wrong. I wouldn't even go so far as to encourage the use of the numbers I pitched above -- if you try this, go with your own based on what works for you.

Having a constraint in the database doesn't mean that you shouldn't _also_ check limits in code. Most programs aren't written to gracefully handle database constraint failures, so for the sake of your users, put in a standard error-handling framework and descriptive error messages in the event this ever happens. Once again, the database is the back stop -- there as a last layer of protection when the others fail.

### Coercible types and operations (#coercible-types)

Back in the old days, there was a decent argument to avoid `varchar` for operational resilience if nothing else. Changing a column's data type is often an expensive process involving full table scans and rewrites that can put a hot database at major risk. Is the potential agony really worth it just to use a `varchar` that's later found to be too short?

Luckily, when it comes to _relaxing_ constraints, this isn't too much of a problem anymore. From the [Postgres docs on `ALTER TABLE`](https://www.postgresql.org/docs/current/sql-altertable.html):

> Adding a column with a volatile `DEFAULT` or changing the type of an existing column will require the entire table and its indexes to be rewritten. As an exception, when changing the type of an existing column, if the `USING` clause does not change the column contents and the old type is either binary coercible to the new type or an unconstrained domain over the new type, a table rewrite is not needed; but any indexes on the affected columns must still be rebuilt.

Note the wording of "unconstrained domain". A `varchar(200)` is an unconstrained domain over a `varchar(100)` because it's strictly longer. Postgres can relax the constraint without needing to lock the table for a scan. Going back the other way isn't as easy, but you shouldn't need to do that.

### SQL domains (#sql-domains)

Another idea I've been been experimenting with is encoding a standard set of text tiers as [domains](https://www.postgresql.org/docs/current/sql-createdomain.html), which defines a new data type with more constraints:

``` sql
CREATE DOMAIN text_standard AS varchar(200) COLLATE "C";
CREATE DOMAIN text_long AS varchar(2000) COLLATE "C";
CREATE DOMAIN text_huge AS varchar(20000) COLLATE "C";
```

The domains can then be used by convention in table definitions:

``` sql
# CREATE TABLE mytext (standard text_standard, long text_long, huge text_huge);

# \d+ mytext
                                       Table "public.mytext"
  Column  |     Type      | Collation | Nullable | Default | Storage  | Stats target | Description
----------+---------------+-----------+----------+---------+----------+--------------+-------------
 standard | text_standard |           |          |         | extended |              |
 long     | text_long     |           |          |         | extended |              |
 huge     | text_huge     |           |          |         | extended |              |
```

The only thing I don't like about this set up is that it somewhat obfuscates what those columns are because they're no longer a common type. It is quite easy to get Postgres to hand you back domain definitions with `\dD`:

``` sql
# \dD
                                      List of domains
 Schema |     Name      |           Type           | Collation | Nullable | Default | Check
--------+---------------+--------------------------+-----------+----------+---------+-------
 public | text_huge     | character varying(20000) | C         |          |         |
 public | text_long     | character varying(2000)  | C         |          |         |
 public | text_standard | character varying(200)   | C         |          |         |
```

But ... almost nobody will know how to do that off the top of their head.

## Integrity in depth (#integrity)

Constraints on text fields are a very small part of a broader story in how relational databases are built to help you. In the beginning, all their pedantry around data types, foreign keys, check constraints, ACID, and insert triggers may seem unnecessarily obscure and inflexible, but in the long run these features serve as strong enforcers of data integrity. You don't have to wonder whether your data is valid -- you know it is.
