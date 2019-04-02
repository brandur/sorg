---
title: Building Safe Systems With Schemas and Strong Constraints
published_at: 2019-01-14T16:25:00Z
location: San Francisco
hook: TODO
tags: ["postgres"]
---

I've talked before about how relational databases like
Postgres provide a powerful platform to build applications
on by offering primitives like [ACID transactions](/acid)
and rolling out [small but important
optimizations](/sortsupport) with new versions.

But of equal importance to these features in a relational
database is the relational part. Originally based on the
[relational model][relationalmodel] of predicate logic,
this design encourages declaratively defining the shape of
entity types in *relations* (A.K.A. tables), how they
relate to each other with keys, and constraints on them
through the definition of a domain.

Almost every developer has run a `CREATE TABLE ...` before
and is familiar with the basics of this concept, but what
isn't so obvious is how having the discipline to build and
maintain a well-defined schema and strong data constraints
on it is a hugely powerful tool for creating robust and
reliable software.

## Establishing guaranteed expectations (#guaranteed-expectations)

When using a "NoSQL" data store (or even a poorly designed
schema in a relational database), accessing any field on
any object is a potential error. For example, if a
developer sees an `email` attribute on a `users`
collection, they might make the reasonable assumption that
it's always present, and write code in such a way that
assumes it's there:

``` ruby
send_password_reset(user.email) # potential bug!
```

Maybe every user does have an email, but it's easy to
imagine a situation where this isn't the case. What if it
turned out that in the early days of the system emails
hadn't been collected? So while the code would work for the
bulk of current users who do have emails on record and
probably in the test suite which creates synthetic users
with modern specifications, it would fail in production for
a certain class of legacy user.

!fig src="/assets/constraints/nulls.svg" caption="Nulls present in data that's mostly non-null."

Even when applications are using application-level
definitions to define a pseudo-schema (e.g. MongoEngine or
Mongoose), it's still possible for data to be present in
the data store that was put there before the current schema
came in.

Even if this example seems a little contrived, this class
of problem can appear in hundreds of different shapes and
sizes. Having worked on a large-scale Mongo-based
application for a few years now, this is *easily* one of
the most frequently seen bugs, and has manifested in every
way from minor annoyance to major production incident.

It's also exactly the class of problem that relational
schemas solve. Schemas are enforced at the database-level
so application developers get an *absolute guarantee* that
data is in the right shape. Not only can they guarantee
that fields like `email` are present through the
application of `NOT NULL` constraints, they provide the
building blocks for even tighter constraints. I could use
them to dictate that all emails should be within a certain
length, or that all prices should be positive. Data that
doesn't meet these requirements is firmly rejected.

## Referential integrity (#referential-integrity)

Now that we've covered that strong expectations are a good
thing, let's look at some tools to establish them. A key
concept in relational databases is the idea of
[normalization][normalization], which encourages relations
to reference the canonical sources of data instead of
duplicating it to reduce redundancy and ensure correctness.

These references can be strongly defined through the use of
a **foreign key**. For example, a pair of tables for a blog
could indicate that its tags always reference an article:

``` sql
CREATE TABLE article (
    id bigserial PRIMARY KEY,
    ...
);

CREATE TABLE tag (
    id bigserial PRIMARY KEY,
    article_id bigint
        REFERENCES article (id),
    ...
);
```

**[Referential integrity][integrity]** is the property that
these relational references are always valid in that they
reference a valid key in another table, or are `NULL`.
Adding a foreign key ensures referential integrity.

Now if some code accidentally deletes an article without
first doing something about the tags that reference it, the
database will block the action [1]:

``` sql
ERROR:  update or delete on table "article" violates foreign
    key constraint "tag_article_id_fkey" on table "tag"

DETAIL:  Key (id)=(1) is still referenced from table "tag".
```

This might initially seem annoying, but like with any
constraint, some *convenience* is being sacrificed for
*correctness*. Articles can still be deleted, but the
developer must indicate explicitly what to do about tags
before they are. The more mature your application, the more
appreciative you'll be of this trade off.

Foreign key constraints mesh very well with `NOT NULL`
constraints. Foreign keys ensure that all references are
valid, and adding `NOT NULL` provides the additional
guarantee that all entities have a reference. In practice
we'd raise our tags with one because it never makes sense
to have a dangling tag that doesn't reference an article:

``` sql
CREATE TABLE tag (
    id bigserial PRIMARY KEY,
    article_id bigint NOT NULL -- <— the new NOT NULL
        REFERENCES article (id),
    ...
);
```

### Deleting Europe (#europe)

A favorite horror story from my time at Heroku is when we
deleted Europe. Heroku allows apps to be deployed in one of
multiple regions including the Americas and Europe. One day
someone was messing around in a production console and
deleted the row which represented that European region and
which every European app referenced. The API instantly
starting throwing errors for every Europe-related request.

At the time we didn't have good discipline around foreign
keys, and although one table referenced values in another,
there wasn't a formal key between the two. Having one would
have saved us *a lot* of trouble as the system would have
blocked the deletion completely [2].

## Domain integrity (#domain-integrity)

### NOT NULL (#not-null)

`NOT NULL` is the simplest possible constraint, but also
one of the most powerful. It's the opposable thumb of the
database world, and separates man from beast.

A null represents the absence of a value, and constraining
a field to be _not null_ forces it to be present. Seeing a
`NOT NULL` constraint in a database table tells you something
about every row in that table -- namely that you can expect
values to be present for that field universally. That makes
coding against it easier in that no conditional check on
the value is necessary, and safer in case such a check is
forgotten.

One of the most egregious mistakes of the SQL standard is
that nullable columns are the default rather than vice
versa. A real-world application should aim for the
strongest expectations by making as many fields
non-nullable as possible (which has become easier more
recently as raising a new non-nullable column is [a
non-blocking operation in Postgres 11](/postgres-default).
If in doubt, start with non-nullable and relax the
constraint if necessary.

### Type system (#type-system)

Type systems are all the rage right now (once again)
through growing popularity of languages like Go, Scala, and
even typed versions of JavaScript like TypeScript.

Booleans, integers, floating-point numbers, byte arrays, date/times

Select appropriate type for what you're doing.

Add types with constraints. e.g. `VARCHAR`

### CHECK (#check)

### Unique partial indexes (#unique-partial-indexes)

## Conclusion (#conclusion)

Migrations.

A schema should be a living thing. It's not only evolving
with the addition of new fields, but old fields are being
pruned and existing fields adjusted for accuracy.

[1] A foreign key without an explicit `ON DELETE`
instruction will default to `ON DELETE NO ACTION` which
raises an error when a referenced row is deleted.

[2] This is also a good reason to prefer `ON DELETE NO
ACTION` or `ON DELETE RESTRICT` in production systems (as
opposed to `CASCADE`, `SET DEFAULT`, `SET NULL`).

[integrity]: https://en.wikipedia.org/wiki/Referential_integrity
[normalization]: https://en.wikipedia.org/wiki/Database_normalization
[relationalmodel]: https://en.wikipedia.org/wiki/Relational_model
