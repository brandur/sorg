---
title: Safe Schemas, Strong Constraints
published_at: 2019-04-26T18:45:56Z
location: San Francisco
hook: TODO
tags: ["postgres"]
---

Relational databases like Postgres provide a powerful
platform on which to build applications by offering the
powerful primitive of [ACID transactions](/acid), useful
features like upsert or logical replication, and rolling
out [small but important optimizations](/sortsupport) with
new versions.

But they also do something more -- an old, but invaluable,
feature that's suggested right in their name. Their
relational nature was originally premised on the
[relational model][relationalmodel] of predicate logic, and
encourages declaratively defining the shape of entity types
in **relations** (better known to most of us as tables),
how those relations relate to each other with keys, and the
constraints on them through the declaration of **domains**.

Almost every developer is familiar with at least the basics
of the concept, having run DDL like `CREATE TABLE ...`, but
it's not obvious how the relational model can be a major
boon for building better software. A well-defined schema
and strong data constraints are hugely powerful tools for
creating robust software that's safe by default, even in
the presence of bugs or mistakes made elsewhere in the
system. Maintaining them takes time and discipline, but
it's work that pays off in correctness and reliability.

## Guaranteed expectations (#guaranteed-expectations)

When using a poorly designed schema (or any "NoSQL" data
store), accessing any field on any object is a potential
error. For example, if a developer sees an `email`
attribute on a `users` collection, they might make the
reasonable assumption that it's always present, and write
code in such a way that assumes it's there:

``` ruby
send_password_reset(user.email) # potential bug!
```

And it's quite possible that every user *does* have an
email, and that works fine, but it's also easy to imagine a
situation where it isn't. What if it turned out that in the
early days of the system emails weren't collected when new
accounts were created? Accessing `user.email` would be fine
for the bulk of current users who have emails on record,
and also fine in the test suite which creates idyllic
synthetic users with modern specifications, but it'd fail
hard in production for a certain class of legacy user.

!fig src="/assets/constraints/nulls.svg" caption="Nulls present in data that's mostly non-null."

In the world of NoSQL, it's common to use application-level
tooling to define pseudo-schemas for data (e.g. MongoEngine
or Mongoose), but that provides no guarantee that all data
conforms to it. Data that would now be invalid could have
been created before the current schema came into effect.
The underlying data store makes no guarantees.

This class of **schema underspecification** problem can
appear in hundreds of shapes and sizes. Having worked on a
large-scale Mongo-based application for years now, this is
*easily* one of the most frequent cause of bugs, and has
manifested in every way from minor annoyance to major
production incident.

It's also exactly the class of problem that relational
databases excel at solving. Schemas are enforced at the
database level so application developers get an *absolute
guarantee* that data is in the right shape, regardless of
its age or circumstances of creation. Not only can they
guarantee that fields like `email` are present through the
application of `NOT NULL` constraints, but they also
provide the building blocks for even tighter constraints:
they could be used to dictate that all emails should be
within a certain length and comply to a certain format.
Data that doesn't meet these requirements is firmly
rejected.

## Referential integrity (#referential-integrity)

A key concept in relational databases is the idea of
[normalization][normalization], which encourages relations
to reference the canonical sources of data instead of
duplicating it. This helps to reduce redundancy, but more
importantly ensures correctness.

As a supremely simple example, it'd be convention to make
sure that a blog article's slug is stored in a single
canonical `article` table rather than distributed out to
every table that references an article. If that slug
subsequently needs to be updated, it's easy to do so in
exactly one place instead of going through every table.

TOOD: Article slug schema example

References are strongly defined through the use of a
**foreign key**. For example, `tag.article_id` references
its canonical sequence in `article.id`:

``` sql
CREATE TABLE article (
    id   bigserial    PRIMARY KEY,
    slug varchar(100) NOT NULL,
    ...
);

CREATE TABLE tag (
    id         bigserial PRIMARY KEY,
    article_id bigint
        REFERENCES article (id),
    ...
);
```

**[Referential integrity][integrity]** is the property that
these relational references are always valid in that they
reference a valid key in another table, or are `NULL`.
Foreign keys ensure referential integrity.

If some code accidentally deletes an article without first
doing something about the tags that reference it, the
database will block the action [1]:

``` sql
ERROR:  update or delete on table "article" violates foreign
    key constraint "tag_article_id_fkey" on table "tag"

DETAIL:  Key (id)=(1) is still referenced from table "tag".
```

As with all of the constraints we'll look at, this might
initially seem annoying, but the key with all of them is
that we're sacrificing some *convenience* to gain
*correctness*. Articles can still be deleted, but the
developer must indicate explicitly what to do about tags
before they are. The more mature an application, the more
this trade off makes sense.

Foreign key constraints mesh very well with `NOT NULL`
constraints. Foreign keys ensure that all references are
valid, and adding `NOT NULL` provides the additional
guarantee that all entities in a relation have a reference.
We'd raise the `tag` relation like this because it never
makes sense to have a dangling tag that doesn't reference
an article:

``` sql
CREATE TABLE tag (
    id         bigserial PRIMARY KEY,
    article_id bigint    NOT NULL     -- <— added: NOT NULL
        REFERENCES article (id),
    ...
);
```

### When we deleted Europe (#europe)

One day at Heroku, we deleted Europe. Heroku allows apps to
be deployed in one of multiple regions including the
Americas and Europe. On the fateful day, someone was
messing around in a production console and accidentally
deleted the row which represented that European region and
which every European app referenced. The API instantly
starting throwing errors for every Europe-related request.

Although the proximate cause of the incident was user
error, more fundamentally, the system should have made it
much more difficult for this to happen. At the time we
didn't have good discipline around foreign keys, and
although one table referenced values in another, there
wasn't a formal contract between the two in the form of a
foreign key. Having one would have saved us *a lot* of
trouble as the deletion would have been blocked
automatically [2].

## Domain integrity (#domain-integrity)

While referential integrity covers the connectedness of
relations, **domain integrity** deals with the constraints
within them. A **domain** is the possible set of values that
a field can contain, and databases provide various
mechanisms to constrain them.

### NOT NULL (#not-null)

`NOT NULL` is the simplest possible constraint, but also
one of the most powerful. It's the opposable thumb of the
database world, and a key property that separates those
that are worth using from those that aren't.

A null represents the absence of a value, and constraining
a field to be _not null_ forces it to be present. Seeing a
`NOT NULL` constraint in a database table tells you
something about every row in that table -- namely that you
can expect values to be present for that field universally.
That makes code written that uses it cleaner (fewer null
checks), and safer to write.

Best practice is to make every `NOT NULL` unless you're
sure that it shouldn't be. It's easy to drop a `NOT NULL`
constraint, but it's harder to add one back later (although
this is now [easier as of Postgres 11](/postgres-default)).
One of the most egregious mistakes of the SQL standard is
that nullable columns are the default rather than vice
versa.

### Types: also good in databases (#types)

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
CREATE UNIQUE INDEX index_user_email ON user (email)
    WHERE deleted_at IS NULL;
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
INSERT INTO password_reset_link (user_id, reset_nonce)
    VALUES (123, ‘abc57DefZ1t’)
    ON CONFLICT (user_id) DO NOTHING RETURNING reset_nonce;
```

### CHECK: the most flexible constraint (#check)

`CHECK` is the Swiss Army Knife of constraints, and allows
the specification of rules that aren't easily represented
by any other kind. It allows an arbitrary expression for
validation on a field or table which must pass for new data
to be allowed.

As a really simple example, we could use one to validate
that prices are greater than zero:

``` sql
CREATE TABLE products (
    id    bigserial,
    name  varchar(500),
    price numeric       CHECK (price > 0)
);
```

And it's usefulness goes far beyond that. A few other way
that it could be applied:

* Check that a string is not empty.

TODO: Expand list.

## Consistent effort, major reward (#effort-reward)

A schema isn’t just an upfront design exercise. New ones
should be created according to the best specifications
available, but will certainly never be sufficient as its
companion application continues to evolve.

Schemas should be in a constant state of refinement to
reflect new developments in an application’s data model and
new constraints to which it’s beholden. This isn’t always
easy. Migrating the schema itself isn't that easy —
migration frameworks like ActiveRecord’s are needed to make
sure that the right migrations are applied and changes are
kept in sync between systems. Migrating **data** to fit new
constraints is even harder, especially where large amounts
of it have been accumulated, but the effort pays off.

Some developers felt the rigidity of progressive schema
evolution so confining that it led to the 00s’ wave of
NoSQL. There is no schema — only documents — so
applications can write new records however they want, with
no restrictions.

Week one on such a system really is faster. Changes are
deployed quickly and efficiently, and there aren’t enough
legacy versions to cause much trouble. But that initial
speed is eventually replaced by a pervasive uncertainty.
Even once a data mapping tooling is used to define a
“current” canonical schema, it’s anybody’s guess as to the
shapes of data that are already in there. Without
exception, the results are that code is more difficult to
write, and more difficult to read as it's littered with
null checks. Bugs become more common, and impossible to
suppress completely.

Like doing the dishes after every meal instead of waiting
for them to pile up at the end of the week (by which time
they smell and are attracting flies), keeping schemas up to
date takes more consistent effort, but helps avoid entire
classes of problems completely. Leverage the powerful tools
offered by relational databases like foreign keys, types,
and `NOT NULL` to build detailed schemas with strong
constraints, and the result will be safer, more correct,
more reliable software.

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
