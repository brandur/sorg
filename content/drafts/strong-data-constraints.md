---
title: Building Safe Systems With Strong Data Constraints
published_at: 2019-01-14T16:25:00Z
location: San Francisco
hook: TODO
tags: ["postgres"]
---

A common source of bugs in data-backed software is when an
assumption is made about the shape of some data that turns
out to not be true. For example, when seeing an `email`
field on a `user` relation, it's a pretty reasonable
assumption that users have email available, as they do in
many systems where you'd see such a field, and it might be
true -- however, it could just as easily turn out not to be
as well. What if it turned out that emails hadn't been
collected in the system's early days for example? Most
users would have emails and accessing the field would work
most of the time, but once in a while it would crash the
program with a null reference error.

Relational databases help mitigate this problem by
encouraging the design of concrete schemas that are always
consistent according to constraints that have been
specifically designed. Those constraints might be anything
from a check that the field is present, to whether
referential integrity is intact. Data that doesn't meet
requirements is firmly rejected.

## Constraints (#constraints)

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

### Foreign keys (#foreign-keys)

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

### Type system (#type-system)

Type systems are all the rage right now (once again)
through growing popularity of languages like Go, Scala, and
even typed versions of JavaScript like TypeScript.

Booleans, integers, floating-point numbers, byte arrays, date/times

Select appropriate type for what you're doing.

Add types with constraints. e.g. `VARCHAR`

### CHECK (#check)

### Unique partial indexes (#unique-partial-indexes)

### The schema (#schema)

Possibly the most important constraint is the schema
itself. In a relational database, you define tables, and
the fields within them. Compare that to a document-oriented
database where there are no requirements. The equivalent of
tables are loose collections of keys and values that may or
may not be present.

Schemas in relational databases should be treated like
living things. Not only are new fields being added to them,
but fields that are no longer needed should be dropped, and
existing fields should be adjusted as appropriate to
maximize the schema's fidelity. The more accurate the
schema, the easier it is to write programs that access
them, and the more bug free those programs are likely to
be.

## Flowing safety into programming languages (#programming-languages)

## Conclusion (#conclusion)

A schema should be a living thing. It's not only evolving
with the addition of new fields, but old fields are being
pruned and existing fields adjusted for accuracy.

[integrity]: https://en.wikipedia.org/wiki/Referential_integrity
[normalization]: https://en.wikipedia.org/wiki/Database_normalization
