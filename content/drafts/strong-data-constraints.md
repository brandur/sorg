---
title: Building Safe Systems With Strong Data Constraints
published_at: 2019-01-14T16:25:00Z
location: San Francisco
hook: TODO
---

Stable expectations result in more bug-free systems.

By offering a wide variety of constraints and types,
relational databases excel at providing stable
expectations.

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

RDMSes provide an easy way to do this with foreign keys. A
foreign key dictates that a column references values a in
column in another table, and can also define what should
happen if any of those target values are removed.

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

Select appropriate type for what you're doing.

Add types with constraints. e.g. `VARCHAR`

### CHECK (#check)

### Unique partial indexes (#unique-partial-indexes)

### The schema (#schema)

## Flowing safety into programming languages (#programming-languages)

## Conclusion (#conclusion)

A schema should be a living thing. It's not only evolving
with the addition of new fields, but old fields are being
pruned and existing fields adjusted for accuracy.

[integrity]: https://en.wikipedia.org/wiki/Referential_integrity
[normalization]: https://en.wikipedia.org/wiki/Database_normalization
