+++
hook = "Touring the ORM and Postgres landscape in Go, and why sqlc is today's top pick."
location = "San Francisco"
published_at = 2021-09-07T17:10:05Z
tags = ["postgres"]
title = "Why We're All in on sqlc/pgx for Postgres + Go"
+++

After a few months of research and experimentation with running a heavily DB-dependent Go app, we've arrived at the conclusion that [sqlc](https://github.com/kyleconroy/sqlc) is the figurative Correct Answer when it comes to using Postgres (and probably other databases too) in Go code beyond trivial uses. Let me walk you through how I got there.

First, let's take a broad tour of popular options in Go's ecosystem:

* `database/sql`: Go's built-in database package. Most people agree -- best to avoid it. It's database agnostic, which is kind of nice, but by extension that means it conforms to the lowest common denominator. No support for Postgres-specific features.

* [`lib/pq`](https://github.com/lib/pq): An early Postgres frontrunner in the Go ecosystem. It was good for its time and place, but has fallen behind, and is no longer actively maintained.

* [`pgx`](https://github.com/jackc/pgx): A very well-written and very thorough package for full-featured, performant connections to Postgres. However, it's opinionated about not offering any ORM-like features, and gets you very little beyond a basic query interface. Like with `database/sql`, hydrating database results into structs is painful -- not only do you have to list target fields off ad nauseam in a `SELECT` statement, but you also have to `Scan` them into a struct manually.
	* [`scany`](https://github.com/georgysavva/scany): Scany adds some quality-of-life improvement on top of pgx by eliminating the need to scan into every field of a struct. However, the desired field names must still be listed out in a `SELECT ...` statement, so it only reduces boilerplate by half.

* [`go-pg`](https://github.com/go-pg/pg): I've used this on projects before, and it's a pretty good little Postgres-specific ORM. A little more below on why ORMs in Go aren't particularly satisfying, but another downside with go-pg is that it implements its own driver, and isn't compatible with pgx.
	* [Bun](https://bun.uptrace.dev/guide/pg-migration.html#new-features): go-pg has also been put in maintenance mode in favor of Bun, which is a go-pg rewrite that works with non-Postgres databases.

* [`gorm`](https://gorm.io/): Similar to go-pg except not Postgres specific. It can use pgx as a driver, but misses a lot of Postgres features.

## Queries as strings (#strings)

A big downside of vanilla `database/sql` or pgx is that SQL queries are strings:

``` go
var name string
var weight int64
err := conn.QueryRow(ctx, "SELECT name, weight FROM widgets WHERE id = $1", 42).
	Scan(&name, &weight)
if err != nil {
	...
}
fmt.Println(name, weight)
```

This is fine for simple queries, but provides little in the way of confidence that queries actually work. The compiler just sees a string, so you need to write exhaustive test coverage to verify them.

And it gets worse. When you're writing a larger application that's trying to hydrate models, in an effort to reduce code duplication, you might start slicing and dicing those query strings -- gluing little pieces together to share code. e.g.

``` go
err := conn.QueryRow(ctx, `SELECT ` + scanTeamFields + ` ...)
```

You can make it work, and still verify what you have is right by way of tests, but it gets messy fast.

## ORMs (#orms)

ORMs like go-pg make this a little better by bringing some typing into the mix, which has some benefit for reducing mistakes:

``` go
story := new(Story)
err = db.Model(story).
    Relation("Author").
    Where("story.id = ?", story1.Id).
    Select()
if err != nil {
    panic(err)
}
```

However, without generics, Go's type system can only offer so much, and in practice, the compiler can't catch all that much more than when we were concatenating strings together. In the code above, `Model()` returns a `*Query` object. `Relation()` also returns a `*Query` object, and so does `Where()`. go-pq can  do some intelligent shuffling (e.g. putting a `LIMIT` before a `WHERE` wouldn't work in SQL, but go-pg will make it work because it's constructing the query lazily), but like with strings, there's a plethora of mistakes that will only be caught on runtime.

ORMs also have the problem of being an impedance mismatch compared to the raw SQL most people are used to, meaning you've got the reference documentation open all day looking up how to do accomplish things when the equivalent SQL would've been automatic. Easier queries are pretty straightforward, but imagine if you want to add an upsert or a [CTE](https://www.postgresql.org/docs/current/queries-with.html).

## sqlc (#sqlc)

And that's where [sqlc](https://github.com/kyleconroy/sqlc) comes in. With sqlc, you write `*.sql` files that contain table definitions along with queries annotated with a name and return type in a magic comment:

``` sql
CREATE TABLE authors (
  id   BIGSERIAL PRIMARY KEY,
  name text      NOT NULL,
  bio  text
);

-- name: CreateAuthor :one
INSERT INTO authors (
  name, bio
) VALUES (
  $1, $2
)
RETURNING *;
```

After running `sqlc generate` [1] (which generates Go code from your SQL definitions), you're now able to run this:

``` go
author, err = dbsqlc.New(tx).CreateAuthor(ctx, dbsqlc.CreateAuthor)
    Name: "Haruki Murakami",
    Bio:  "Author of _Killing Commendatore_. Running and jazz enthusiast.",
    ...
})

if err != nil {
    return nil, xerrors.Errorf("error creating author: %w", err)
}

fmt.Printf("Author name: %s\n", author.Name)
```

sqlc isn't an ORM, but it implements one of the most useful features of one -- mapping a query back into a struct without the need for boilerplate. If you have query with a `SELECT *` or `RETURNING *`, it knows which fields a table is supposed to have, and emits the result to a standard struct representing its records. All queries for a particular table that return its complete set of fields get to share the same output struct.

Rather than implement its own partially-complete SQL parser, sqlc uses PGAnalyze's [excellent `pg_query_go`](https://github.com/pganalyze/pg_query_go), which bakes in the same query parser that Postgres really uses. It's never given me trouble so far -- even complex queries with unusual Postgres embellishments work.

This query parsing also gives you some additional pre-runtime code verification. It won't protect you against logical bugs, but it won't compile invalid SQL queries, which is a far shot better than the guarantees you get with SQL-in-Go-strings. And thanks to SQL's declarative nature, it tends to produce fewer bugs than comparable procedural code. You'll still want to write tests, but you don't need to test every query and corner case as exhaustively.

### Codegen (#codegen)

I'm slightly allergic to the idea of codegen on a philosophical level, and that made me reluctant to look too deeply into sqlc, but after finally getting into it, it's won me over.

Go makes programs like sqlc easily installable in one command (`go get github.com/kyleconroy/sqlc/cmd/sqlc`), and quickly with minimal fuss. Go's lightning fast startup and runtime speed means that your codegen loop runs in the blink of an eye. Our project is sitting around 100 queries broken up across a dozen input files and its codegen runs in (much) less than a second on commodity hardware:

``` go
$ time sqlc generate

real    0.07s
user    0.08s
sys     0.01s
```

Even if we expand our number of queries by 100x to 10,000, I think we'll still be comfortable with the timing on that development loop.

A GitHub Action verifies generated output, and between checkout, pulling down an sqlc binary, and running it, the whole job takes a grand total of 4 seconds to run.

### pgx support appears (#pgx)

Previously, a major reason not to use sqlc is that it didn't support pgx, which we were already bought into pretty deeply. A recent [pull request](https://github.com/kyleconroy/sqlc/pull/1037) has addressed this problem by giving sqlc support for multiple drivers, and the feature's now available in the sqlc's latest release.

The authors also managed to write it in such a way that it's coupled very loosely -- our mature codebase was making heavy use of pgx already and had a number of custom abstractions built on top of it, and yet I was able to get sqlc slotted in alongside them and fully operational in less than an hour. We could even weave sqlc invocations in amongst raw pgx invocations as part of the same transaction, giving us an easy way to migrate over to it incrementally.

### Caveats and workarounds (#caveats)

A few things in sqlc are less convenient compared to a more traditional ORM, but there are workarounds that land pretty well. For example, a noticeable one is that sqlc queries can't take an arbitrary number of parameters, so doing a multi-row insert doesn't work as easily as you'd expect it to. However, you can get around this by sending batches as arrays which are unnested into distinct tuples in the SQL:

``` sql
-- Upsert many marketplaces, inserting or replacing data as necessary.
INSERT INTO marketplace (
    name,
    display_name
)
SELECT unnest(@names::text[]) AS name,
    unnest(@display_names::text[]) AS display_names
ON CONFLICT (name)
    DO UPDATE SET display_name = EXCLUDED.display_name
RETURNING *;
```

Another one is `UPDATE` where with a normal ORM you'd just add as many target fields and values (i.e. `UPDATE foo SET a = 1, b = 2, c = 3, ...`) through the query builder as you wanted. Queries in sqlc must be fully structured in advance, so this doesn't work. What you can do is something like this where each field is conditionally updated based on the presence of an associated boolean:

``` sql
-- Update a team.
-- name: TeamUpdate :one
UPDATE team
SET
    customer_id = CASE WHEN @customer_id_do_update::boolean
        THEN @customer_id::VARCHAR(200) ELSE customer_id END,

    has_payment_method = CASE WHEN @has_payment_method_do_update::boolean
        THEN @has_payment_method::bool ELSE has_payment_method END,

    name = CASE WHEN @name_do_update::boolean
        THEN @name::text ELSE name END
WHERE
    id = @id
RETURNING *;
```

The Go code to update a field ends up looking like this:

``` go
team, err = queries.TeamUpdate(ctx, dbsqlc.TeamUpdateParams{
    NameDoUpdate: true,
    Name:         req.Name,
})
```

sqlc doesn't have any built-in conventions around how queries are named or organized, so you'll want to make sure to come up with your own so that you can find things.

## Summary and future (#summary)

I've largely covered sqlc's objective benefits and features, but more subjectively, it just _feels_ good and fast to work with. Like Go itself, the tool's working for you instead of against you, and giving you an easy way to get work done without wrestling with the computer all day.

I won't go as far as to say that its the best answer across all ecosystems -- the feats that Rust's SQL drivers can achieve with its type system are borderline wizardly -- but sqlc's far and away my preferred solution when working in Go.

Lastly, generics are coming to Go, possibly [in beta form by the end of the year](https://go.dev/blog/generics-proposal), and that could change the landscape. I could imagine a world where they power a new generation of Go ORMs that can do better query checking and give you even better type completion. However, it's safe to say that's a good year or two out. Until then, we're happy with sqlc.

[1] There's a little configuration involved as well -- see [this quickstart](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html) in the official documentation.
