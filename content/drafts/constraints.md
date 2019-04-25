---
title: Safe Schemas and Strong Constraints
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

### The curious case of deleting Europe (#europe)

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

While referential integrity covers the connectedness of
relations, **domain integrity** deals with the constraints
within them. A "domain" is the possible set of values that
a field can contain, and databases provide various ways to
constrain them.

### NOT NULL (#not-null)

`NOT NULL` is the simplest possible constraint, but also
one of the most powerful. It's the opposable thumb of the
database world, and separates man from beast.

A null represents the absence of a value, and constraining
a field to be _not null_ forces it to be present. Seeing a
`NOT NULL` constraint in a database table tells you something
about every row in that table -- namely that you can expect
values to be present for that field universally, which
makes code written for it cleaner and safer to write.

Best practice is to make every `NOT NULL` unless you're
sure that it shouldn't be. It's easy to drop a `NOT NULL`
constraint, but it's harder to add one back later (although
this is now [easier as of Postgres 11](/postgres-default)).
One of the most egregious mistakes of the SQL standard is
that nullable columns are the default rather than vice
versa.

### Type system (#type-system)

Strong type systems in programming languages are a popular
idea right now the growing popularity of languages like Go,
Rust, and Scala, and with the augmentation of dynamically
typed languages with type annotations like TypeScript and
Mypy in Python, Just like in programming languages, types
are useful at the database level.

Selecting the right type vastly shrinks the possible domain
of a field, and that acts as an additional form of
validation when storing values. String-like fields of
`text` and `varchar` are common, but it pays to type a
field as `boolean`, `integer`, `bytea` or `timestamptz`
where appropriate.

Some types take an extra parameter that further constrains
their values. For example, a `varchar` specified as
`varchar(500)` will accept a maximum of 500 characters.
Like with `NOT NULL`, it's almost always a good idea to put
these constraints in place. They save a lot of potentially
bad user input, and it's easy to relax them later but
difficult to tighten them up [3].

#### Boundless strings (#boundless-strings)

For a very long time it was possible to put strings of any
length into a user-modifiable string fields in Stripe's
API. Mongo doesn't provide an easy way to constrain data
types like with a `VARCHAR`, and no one had bothered to add
checks in the application-space.

Eventually we locked them down, but because we take
forwards compatibility very seriously, we had to do so in
such a way that no one would be broken. We chose lengths of
fields that were far longer than ideal because they
represented the upper bound of what people were storing to
them.

A simple default length constraint on fields from the
beginning would've saved weeks worth of effort, and left
the final product with a far more desirable design.

### Unique partial indexes (#unique-partial-indexes)

A common requirement is to enforce uniqueness on values in
a particular field. Like many databases, Postgres offers a
`UNIQUE` keyword when creating columns, but far more
interesting (and practically useful) is the ability to
enforce uniqueness on a **partial index**, which is one
that doesn’t cover the entirety of the table.

TODO: Partial index diagram.

For example, we may have a users table that contains a flag
indicating when a record in it was deleted, and which
contains `NULL` in the case of all live rows. We want to
enforce uniqueness on email addresses, but only for records
that are still live:

``` sql
CREATE UNIQUE INDEX index_user_email ON user (email) WHERE deleted_at IS NULL;
```

In practice this turns out to be a powerful tool that’s
useful in a surprisingly wide range of situations.

These indexes can also be used in conjunction with “upsert”
to have an application instruct the database on what to in
the case of a conflict. For example, when generating a
password reset link we can have an application
optimistically generate a random reset nonce, but in the
case a previous record existed (another password reset on
the same account was tried recently), we can instruct
Postgres to do nothing and return whatever nonce was
already there to use instead:

``` sql
INSERT INTO password_reset_link (user_id, reset_nonce) VALUES (123, ‘abc57DefZ1t’) ON CONFLICT (user_id) DO NOTHING RETURNING reset_nonce;
```

### CHECK (#check)

`CHECK` is the most flexible constraint that Postgres
allows, enabling an arbitrary expression for validation on
a field or table. For example, we could use one to validate
that prices are greater than zero:

``` sql
CREATE TABLE products (
    id bigserial,
    name varchar(500),
    price numeric CHECK (price > 0)
);
```

`CHECK` is a great choice for domain checks that aren't
easily covered by one of the other types of constraints.

## Conclusion (#conclusion)

A strong schema isn’t solely an upfront design exercise.
New ones should be created according to the best
specifications available, but that first pass probably wont
be without errors, and will certainly never be sufficient
as its companion application continues to evolve.

Schemas should be in a constant state of refinement to
reflect new developments in an application’s data model and
new constraints to which it’s beholden. This isn’t always
easy. Migrating the schema itself isn’t trivial as is —
often migrations like ActiveRecord’s are needed to make
sure that the right migrations are applied and changes are
kept in sync between systems — but migrating data to fit
new constraints is even harder, especially where a
non-trivial amount of it has accumulated.

Some developers felt the rigidity of progressive schema
evolution so confining that it led to the 00s’ wave of
NoSQL databases. There is no schema — only documents — so
applications can write new records however they want, with
no restrictions.

And week one on such a system really is faster. Changes are
deployed quickly and efficiently, and there aren’t enough
legacy versions to cause much trouble. But that initial
speed is eventually replaced by a pervasive uncertainty.
Even once data mapping tooling is used to define a
“current” canonical schema, it’s anybody’s guess as to the
shapes of data that are already in there. Without
exception, the result is that code is more difficult too
write; bugs become more common, and practically impossible
to suppress completely.

Like doing the dishes after every meal instead of waiting
for them to pile up at the end of the week, by which time
they smell and are attracting flies, consistent effort in
keeping schemas up to date doesn’t just off by making
things easier to every day, but by avoiding entire classes
of problems completely. Leverage the powerful tools offered
by relational databases like foreign keys, types, and `NOT
NULL` to build more correct, more reliable software.

[1] A foreign key without an explicit `ON DELETE`
instruction will default to `ON DELETE NO ACTION` which
raises an error when a referenced row is deleted.

[2] This is also a good reason to prefer `ON DELETE NO
ACTION` or `ON DELETE RESTRICT` in production systems (as
opposed to `CASCADE`, `SET DEFAULT`, `SET NULL`).

[3] When adding a new constraint to an existing table,
databases like Postgres must lock while existing values
are checked to be within the new allowed bounds. This can
have a profound impact on production operations.

[integrity]: https://en.wikipedia.org/wiki/Referential_integrity
[normalization]: https://en.wikipedia.org/wiki/Database_normalization
[relationalmodel]: https://en.wikipedia.org/wiki/Relational_model
