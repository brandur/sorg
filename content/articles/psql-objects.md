---
hook: Using backslash commands in psql to navigate and describe object hierarchy in
  Postgres and Redshift.
location: San Francisco
published_at: 2015-10-26T16:08:30Z
title: Exploring Object Hierarchies in Psql
---

Postgres (or Redshift if you're so inclined) arranges itself through a set of
nested objects with familiar constructs like a table or an index stored in the
highest logical layer. The complete hierarchy looks like this:

* **Cluster:** The top-level Postgres installation. Users and groups are shared
  between all databases in a cluster, but other objects are contained within a
  database.
    * **Database:** Groups together a set of schemas. May have a default schema
      so that relations can be created within it directly.
        * **Schema:** Groups together a set of relations. Duplicate relation
          names are not allowed in the same schema, but are allowed in
          different schemas.
            * **Relation:** Any other type of named Postgres object like a
              table, view, index, or function.

``` monodraw
                       DATABASE CLUSTER                         
                                                                
┌─────────────────────────────────────────────────────────────┐ 
│                                                             │█
│                                                             │█
│                                                             │█
│      DATABASE_A                                             │█
│     ┌────────────────────────────────────────────────┐      │█
│     │                                                │░     │█
│     │    Schema_1              Schema_2              │░     │█
│     │   ╔══════════════════╗  ╔══════════════════╗   │░     │█
│     │   ║ ┌──────────────┐ ║  ║                  ║   │░     │█
│     │   ║ │   Table_A    │ ║  ║                  ║   │░     │█
│     │   ║ └──────────────┘ ║  ║                  ║   │░     │█
│     │   ║ ┌──────────────┐ ║  ║                  ║   │░     │█
│     │   ║ │   Table_B    │ ║  ║                  ║   │░     │█
│     │   ║ └──────────────┘ ║  ║                  ║   │░     │█
│     │   ╚══════════════════╝  ╚══════════════════╝   │░     │█
│     │                                                │░     │█
│     └────────────────────────────────────────────────┘░     │█
│      ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░     │█
│                                                             │█
│                                                             │█
│                                                             │█
└─────────────────────────────────────────────────────────────┘█
 ███████████████████████████████████████████████████████████████
```

See [the Postgres documentation][schemas-docs] for more information about each
of these layers.

Between complex nesting and name collisions, using Postgres tools to navigate
unfamiliar arrangements of these objects can be a little tricky. Luckily `psql`
exposes some handy shortcuts for working with them.

## Databases

We can use `\list` (`\l`) to list all the databases in a cluster:

    \l

A different database in the cluster can be accessed with `\connect` (`\c`):

    \c postgres

## Schemas

We can use `\dn` to list schemas within a database:

    \dn

## Relations

Relations can be listed with `\d`:

    \d

A trailing character can be used in the form of `\d{E,i,m,s,t,v}` to specify a
type of relation (see [the full documentation for more
information][psql-docs]). This is often combined with a "pattern" which we can
set to the name of a schema we want to query. For example, this will list all
tables and views inside of `my_schema`:

    \dtv my_schema.

And a particular relation can be described with (and try using a `\d+` instead
to get even more information):

    \d my_schema.my_table

## Quick Reference

Now we've introduced a lot of esoteric commands here that might be hard to
remember, but luckily there's an easy trick. Psql has a few different internal
help mechanisms, and we can use `\?` to invoke a very succinct help menu that
contains a quick reference for every "backslash command":

    \?

If you're going to take one thing away from this article, make it `\?`.

## Search Path

One other important piece to mention is the [`search_path`
setting][search-path-docs], which has a few important functions:

1. When trying to resolve a symbol without a schema prefix, each schema in the
   `search_path` will be tried in turn. Schemas outside of the `search_path`
   will not be tried, and a schema-less symbol that belongs to one of them will
   not be resolved.
2. When listing relations with `\d` and without providing a search pattern,
   only schemas in `search_path` will be shown.
3. May subtly shift behavior in a few other places. For example, in Redshift
   `pg_table_def` will only show information for tables contained in schemas
   that are present in `search_path`.

`search_path` can be set as follows (and note that this can also be placed in
your `~/.psqlrc`):

    set search_path to '$user', infra, logs, public;

[psql-docs]: http://www.postgresql.org/docs/current/static/app-psql.html
[schemas-docs]: http://www.postgresql.org/docs/current/static/ddl-schemas.html
[search-path-docs]: http://www.postgresql.org/docs/8.1/static/ddl-schemas.html#DDL-SCHEMAS-PATH
