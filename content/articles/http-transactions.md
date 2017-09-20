---
title: Using Atomic Transactions to Power an Idempotent API
published_at: 2017-09-06T16:00:14Z
location: San Francisco
hook: Part one of a series on getting easy data correctness
  by building APIs on the transactional machinery of
  Postgres.
---

The software industry as a whole contains a lot of people
doing a lot of different things, but for every developer
working on new embedded firmware, there's about ten
building the linchpin of modern software -- CRUD apps that
serve requests over HTTP. A lot of these apps are backed by
MVC frameworks like Ruby on Rails or ASP.NET, and backed by
ACID-compliant relational databases like Postgres or SQL
Server.

Sharp edges in production can lead to all kinds of
unexpected cases during the execution of an HTTP request --
client disconnects, application bugs that fail a request
midway through, and timeouts are all extraordinary
conditions that will occur regularly given enough request
volume. Databases can protect applications against
integrity problems with their transactions, and it's worth
taking a little time to think about how to make best use of
them.

There's a surprising symmetry between an HTTP request and a
database's transaction. Just like the transaction, an HTTP
request is a transactional unit of work -- it's got a
clear beginning, end, and result. The client generally
expects a request to execute atomically and will behave as
if it will (although that of course varies based on
implementation). Here we'll look at an example service to
see how HTTP requests and transactions apply nicely to one
another.

## The 1:1 Model (#one-to-one)

I'm going to make the case that for a common idempotent
HTTP request, requests should map to backend transactions
at 1:1. For every request, all operations are committed or
aborted as part of a single transaction within it.

!fig src="/assets/http-transactions/http-transactions.svg" caption="Transactions (tx1, tx2, tx3) mapped to HTTP requests at a 1:1 ratio."

At first glance requiring idempotency may sound like a
sizeable caveat, but in many APIs operations can be made to
be idempotent by massaging endpoint verbs and behavior, and
moving non-idempotent operations like network calls to
background jobs.

Some APIs can't be made idempotent and those will need a
little extra consideration. We'll look at what to do about
them in more detail later as a follow up to this article.

## A simple user creation service (#create-user)

Let's build a simple test service with a single "create
user" endpoint. A client hits it with an `email` parameter,
and the endpoint responds with status `201 Created` to
signal that the user's been created. The endpoint is also
idempotent so that if a client hits the endpoint again with
the same parameter, it responds with status `200 OK` to
signal that everything is still fine.

```
PUT /users?email=jane@example.com
```

On the backend, we're going to do three things:

1. Check if the user already exists, and if so, break and
   do nothing.
2. Insert a new record for the user.
3. Insert a new "user action" record. It'll serve as an
   audit log which comes with a reference to a user's ID,
   an action name, and a timestamp.

We'll build our implementation with Postgres, Ruby, and an
ORM in the style of ActiveRecord or Sequel, but these
concepts apply beyond any specific technology.

### Database schema (#database-schema)

The service defines a simple Postgres schema containing
tables for its users and user actions [1]:

``` sql
CREATE TABLE users (
    id    BIGSERIAL PRIMARY KEY,
    email TEXT      NOT NULL CHECK (char_length(email) <= 255)
);

-- our "user action" audit log
CREATE TABLE user_actions (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users (id),
    action      TEXT        NOT NULL CHECK (char_length(action) < 100),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Backend implementation (#implementation)

The server route checks to see if the user exists. If so,
it returns immediately. If not, it creates the user and
user action, and returns. In both cases, the transaction
commits successfully.

``` ruby
put "/users/:email" do |email|
  DB.transaction(isolation: :serializable) do
    user = User.find(email)
    halt(200, 'User exists') unless user.nil?

    # create the user
    user = User.create(email: email)

    # create the user action
    UserAction.create(user_id: user.id, action: 'created')

    # pass back a successful response
    [201, 'User created']
  end
end
```

The SQL that's generated in the case of a successful
insertion looks roughly like:

``` sql
START TRANSACTION
    ISOLATION LEVEL SERIALIZABLE;

SELECT * FROM users
    WHERE email = 'jane@example.com';

INSERT INTO users (email)
    VALUES ('jane@example.com');

INSERT INTO user_actions (user_id, action)
    VALUES (1, 'created');

COMMIT;
```

## Concurrency protection (#concurrency-protection)

Readers with sharp eyes may have noticed a potential
problem: our `users` table doesn't have a `UNIQUE`
constraint on its `email` column. The lack of one could
potentially allow two interleaved transactions to run their
`SELECT` phase one concurrently and get empty results.
They'd both follow up with an `INSERT`, leaving a
duplicated row.

!fig src="/assets/http-transactions/concurrent-race.svg" caption="A data race causing two concurrent HTTP requests to insert the same row."

Luckily, in this example we've used an even more powerful
mechanism than `UNIQUE` to protect our data's correctness.
Invoking our transaction with `DB.transaction(isolation:
:serializable)` starts it in `SERIALIZABLE`; an isolation
level so powerful that its guarantees might seem
practically magical.  It emulates serial transaction
execution as if each outstanding transaction had been
executed one after the other, rather than concurrently. In
cases like the above where a race condition would have
caused one transaction to taint the results of another, one
of the two will fail to commit with a message like this
one:

```
ERROR:  could not serialize access due to read/write dependencies among transactions
DETAIL:  Reason code: Canceled on identification as a pivot, during commit attempt.
HINT:  The transaction might succeed if retried.
```

We're not going to look into how `SERIALIZABLE` works, but
sufficed to say it may detect a number of different data
races for us, and if it does it'll abort a transaction when
it tries to commit.

### Retrying an abort (#abort-retry)

Even though in our example a race should be rare, we'd
prefer to handle it correctly in our application code so
that it doesn't bubble up as a 500 to a client. This is
possible by wrapping the request's core operations in a
loop:

``` ruby
MAX_ATTEMPTS = 2

put "/users/:email" do |email|
  MAX_ATTEMPTS.times do
    begin
      DB.transaction(isolation: :serializable) do
        ...
      end

      # Success! Leave the loop.
      break

    rescue Sequel::SerializationFailure
      log.error "Failed to commit serially: #{$!}"
      # Failure: fall through to the next loop.
    end
  end
end
```

In this case, we might have more than one of the same
transaction mapped to the HTTP request like so:

!fig src="/assets/http-transactions/transaction-retry.svg" caption="An aborted transaction being retried within the same request."

These loops will be more expensive than usual, but again,
we're protecting ourselves against an unusual race. In
practice, unless callers are particularly contentious,
they'll rarely occur.

Gems like [Sequel][sequel] can handle this for you
automatically (this code will behave similarly to the loop
above):

``` ruby
DB.transaction(isolation: :serializable,
    retry_on: [Sequel::SerializationFailure]) do
  ...
end
```

### Data protection in layers (#layers)

I've taken the opportunity to demonstrate the power of a
serializable transaction, but in real life you'd want to
put in a `UNIQUE` constraint on `email` even if you
intended to use the serializable isolation level. Although
`SERIALIZABLE` will protect you from a duplicate insert, an
added `UNIQUE` will act as one more check to protect your
application against incorrectly invoked transactions or
buggy code. It's worth having it in there.

## Background jobs (#background-jobs)

It's a common pattern to add jobs to a background queue
during an HTTP request so that they can be worked
out-of-band and a waiting client doesn't have to block on
an expensive operation.

Let's add one more step to our user service above. In
addition to creating user and user action records, we'll
also make an API request to an external support service to
tell it that a new account's been created. We'll do that by
queuing a background job because there's no reason that it
has to happen in-band with the request.

``` ruby
put "/users/:email" do |email|
  DB.transaction(isolation: :serializable) do
    ...

    # enqueue a job to tell an external support service
    # that a new user's been created
    enqueue(:create_user_in_support_service, email: email)

    ...
  end
end
```

If we used a common job queue like Sidekiq to do this work,
then in the case of a transaction rollback (like we talked
about above where two transactions conflict), we could end
up with an invalid job in the queue. It's referencing data
that no longer exists, so no matter how many times job
workers retried it, it can never succeed.

### Transaction-staged jobs (#staged-jobs)

A way around this is to create a job staging table into our
database. Instead of sending jobs to the queue directly,
they're sent to a staging table first, and an ***enqueuer***
pulls them out in batches and puts them to the job queue.

``` sql
CREATE TABLE staged_jobs (
    id       BIGSERIAL PRIMARY KEY,
    job_name TEXT      NOT NULL,
    job_args JSONB     NOT NULL
);
```

The enqueuer selects jobs, enqueues them, and then removes
them from the staging table [2]. Here's a rough
implementation:

``` ruby
loop do
  DB.transaction do
    # pull jobs in large batches
    job_batch = StagedJobs.order('id').limit(1000)

    if job_batch.count > 0
      # insert each one into the real job queue
      job_batch.each do |job|
        Sidekiq.enqueue(job.job_name, *job.job_args)
      end

      # and in the same transaction remove these records
      StagedJobs.where('id <= ?', job_batch.last).delete
    end
  end
end
```

Because jobs are inserted into the staging table from
within a transaction, its _isolation_ property (ACID's "I")
guarantees that they're not visible to any other
transaction until after the inserting transaction commits.
A staged job that's rolled back is never seen by the
enqueuer, and doesn't make it to the job queue.

I call this pattern a [_transactionally-staged job
drain_](/job-drain).

It's also possible to just put the job queue directly in
the database itself with a library like [Que], but [because
bloat can be potentially dangerous in systems like
Postgres][queues], this probably isn't as good of an idea.

## Non-idempotent requests (#non-idempotent-requests)

What we've covered here works nicely for HTTP requests that
are idempotent. That's probably a healthy majority given a
well-designed API, but there are always going to be some
endpoints that are not idempotent. Examples include calling
out to an external payment gateway with a credit card,
requesting a server to be provisioned, or anything else
that needs to make a synchronous network request.

For these types of requests we're going to need to build
something a little more sophisticated, but just like in
this simpler case, our database has us covered. In part two
of this series we'll look at how to implement [idempotency
keys][idempotency] on top of multi-stage transactions.

[1] Note that for the purposes of this simple example we
could probably make this SQL more succinct, but for good
hygiene, we use length check, `NOT NULL`, and foreign key
constraints on our fields even if it's a little more noisy
visually.

[2] Recall that like many job queues, the "enqueuer" system
shown guarantees "at least once" rather than "exactly once"
semantics, so the job themselves must be idempotent.

[idempotency]: https://stripe.com/blog/idempotency
[que]: https://github.com/chanks/que
[queues]: /postgres-queues
[sequel]: https://github.com/jeremyevans/sequel
