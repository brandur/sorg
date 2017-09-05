---
title: How to Map an Atomic Transaction onto an Idempotent
  HTTP Request
published_at: 2017-09-05T20:29:07Z
location: San Francisco
hook: TODO
---

A lot of us are building applications that server requests
over the lingua franca of today's internet -- HTTP. Many of
those same applications are backed by a relational database
like Postgres, and requests perform a series of operations
against it to complete.

There's a surprising symmetry between an HTTP request and a
database's transaction. Just like the transaction, an HTTP
request is a transactional piece of work -- it's got a
clear beginning, end, and result. The client generally
expects one to execute atomically (although whether it
actually does may vary based on implementation).

The sharp edges encountered in the real world can lead to
all kinds of unexpected cases during the execution of an
HTTP request -- client disconnects, application bugs that
fail a request midway through, and various sorts of
timeouts will occur regularly given enough request volume.
In these cases a relational DB's transaction can be a
powerful tool for making sure that data stays correct even
given these sorts of adverse conditions.

I'm going to suggest that for a common idempotent HTTP
request, requests map to backend transactions at 1:1. At
first glance requiring idempotency may sound like a
sizeable caveat, but in a well-designed API, many
operations can be tweaked so that they're idempotent.

TODO: Diagram of HTTP request and transaction.

## Let's create a user (#create-user)

Lets come up with a simple test scenario where we're
building an idempotent "create user" endpoint. A client
hits it with a single `email` parameter, and the endpoint
responds with status `201` if the user already existed, and
status `200` otherwise.

```
PUT /users?email=jane@example.com
```

On the backend, we want to insert a record for a new user,
and we also want to insert a "user action", a record that
will serve as an audit log and which comes with a reference
to a user's ID, an action name, and a timestamp.

In SQL [1]:

``` sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL CHECK (char_length(email) <= 255)
);

-- our "user action" audit log
CREATE TABLE user_actions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id),
    action TEXT NOT NULL CHECK (char_length(action) < 100),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

The server route checks to see if the user exists. If so,
it returns immediately. If not, it creates the user and
user action, and returns. In both cases the transaction
commits successfully:

``` ruby
put "/users/:email" do |email|
  DB.transaction(isolation: :serializable) do
    user = User.find(email)
    halt(200, 'User exists') unless user.nil?

    user = User.create(email: email)
    UserAction.create(user_id: user.id, action: 'created')
    [201, 'User created']
  end
end
```

The first time we invoke the endpoint we get a `200 OK`:

``` sh
$ curl -i http://localhost/users/jane@foo.com
HTTP/1.1 200 OK
...

User created
```

And the next time, a `201 Created`:

``` sh
$ curl -i http://localhost/users/jane@foo.com
HTTP/1.1 201 Created
...

User exists
```

The SQL that's generated on the successful insertions looks
roughly like:

``` sql
START TRANSACTION ISOLATION LEVEL SERIALIZABLE;
SELECT * FROM users WHERE email = 'jane@example.com';
INSERT INTO users (email) VALUES ('jane@example.com');
INSERT INTO user_actions (user_id, action) VALUES (1, 'created');
COMMIT;
```

## Concurrency protection (#concurrency-protection)

Readers with sharp eyes may have noticed a potential
problem: our user's `email` column doesn't have a `UNIQUE`
constraint. The lack of one could potentially allow two
interleaved transactions to run their `SELECT` phase one
concurrently and get empty results. They'd both follow up
with an `INSERT`, thus leaving a duplicated row.

Luckily, the magic of the `SERIALIZABLE` isolation level
protects us from harm here. It guarantees that all
outstanding transactions would have been committable in
serial order, no matter which one committed first in
practice. In the case of such a race condition, one of the
two outstanding transactions would fail on commit with this
message:

```
# COMMIT;
ERROR:  could not serialize access due to read/write dependencies among transactions
DETAIL:  Reason code: Canceled on identification as a pivot, during commit attempt.
HINT:  The transaction might succeed if retried.
Time: 0.291 ms
```

Even though this race condition is rare, we'd prefer to
handle it correctly in our application code. This is
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
      break
    rescue Sequel::SerializationFailure
      log.error "Failed to commit serially: #{$!}"
      # fall through to the next loop
    end
  end
end
```

In this case, we might have more than one of the same
transaction mapped to the HTTP request like so:

TODO: Diagram of mapping.

These loops will be more expensive than usual, but keep in
mind that we're only protecting against an unusual
concurrency access race. In practice, unless callers are
particularly contentious, it should rarely occur.

Gems like Sequel can handle this for you automatically:

``` ruby
DB.transaction(isolation: :serializable,
    retry_on: [Sequel::SerializationFailure]) do
  ...
end
```

And just as short addendum, although I've taken this
opportunity to demonstrate the power of a serializable
transaction, in real life you'd want to put in a `UNIQUE`
constraint on `email` even if you intended to use
`SERIALIZABLE`. It'll be one additional check to help
protect against incorrectly invoked transactions or buggy
code.

## Background jobs and job staging (#background-jobs)

It's a common pattern to add jobs to a background queue
during an HTTP request so that they can be worked
out-of-band and a waiting client doesn't have to block on
an expensive operation. Say for example that in addition to
creating our user record, we're going to tell an external
support service that an account's been created:

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
then in the case of a transaction rollback, we could end up
with an invalid job in the queue. No matter how many times
our job workers retried it, it would never succeed.

A way around this is to create a job staging table into our
database. Instead of sending jobs to the queue directly,
they're send to staging table first, and an ***enqueuer***
pulls them out in batches and sends them to the job queue.

``` sql
CREATE TABLE staged_jobs (
    id BIGSERIAL PRIMARY KEY,
    job_name TEXT NOT NULL,
    job_args jsonb NOT NULL
);
```

The enqueuer selects jobs, enqueues them, and then removes
them from staging [2]. Its implementation looks roughly
like this:

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

Because jobs are staged from within a transaction, its
_isolation_ (ACID's "I") property guarantees that they're
not visible until after the transaction commits. A staged
job that's inserted from within a transaction that rolls
back is never seen by the enqueuer, and doesn't make it to
the job queue.

It's also possible to just put the job queue directly in
the database itself with a library like [Que], but [because
bloat can be potentially dangerous in systems like
Postgres][queues], I'd generally recommend against that.

## Non-idempotent requests (#non-idempotent-requests)

What we've covered here works nicely for HTTP requests that
are idempotent. That's probably a healthy majority given a
well-designed API, but there are always going to be some
endpoints that are not idempotent. For example making a
call out to an external payment gateway with credit card
information, requesting a server to be provisioned, or
anything that needs to make a fallible synchronous network
request.

For these types of requests we're going to need to build
something like an [idempotency key][idempotency] on top of
multi-stage transactions. This gets a little more involved
though, so I'm going to save it for a part two of this
article.

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
