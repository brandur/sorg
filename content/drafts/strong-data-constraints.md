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

A key concept in relational database is the idea of
[normalization][normalization], which encourages tables to
reference the canonical sources of data to reduce
redundancy. (So instead of duplicating data, tables
reference where it exists in other tables.)

Assuming the common situation with the existence of many
references in a database, a related concept is maintaining
[referential integrity][integrity], which is the property
that every reference in the database is valid.

Relational databases provide an easy way to do this with
foreign keys. A foreign key dictates that a column
references values a in column in another table, and can
also define what should happen if any of those target
values are removed.

A setting that's appropriate in most cases is `ON DELETE
RESTRICT`. With this configuration, the database won't let
values that are being referenced be deleted unless the
referring rows are deleted first. This is the safest
setting because it might prevent accidental data deletion
through either bugs or human typos.

``` sql
CREATE TABLE article (
    id bigserial PRIMARY KEY,
    ...
);

CREATE TABLE tag (
    id bigserial PRIMARY KEY,
    article_id bigint NOT NULL
        REFERENCES article (id) ON DELETE RESTRICT,
    ...
);
```

In the example above, tags reference their parent articles,
and an article can't be deleted unless all its tags are
first deleted. Note also that we've combined foreign keys
with `NOT NULL`, which means that we've created a strong
requirement that every tag reference an article, and every
article reference is one that exists.

### Deleting Europe (#europe)

At Heroku, 

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

[integrity]: https://en.wikipedia.org/wiki/Referential_integrity
[normalization]: https://en.wikipedia.org/wiki/Database_normalization
[relationalmodel]: https://en.wikipedia.org/wiki/Relational_model
