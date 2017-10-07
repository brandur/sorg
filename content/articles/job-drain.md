---
title: Transactionally Staged Job Drains in Postgres
published_at: 2017-09-20T14:58:14Z
location: Calgary
hook: Building a robust background worker system that
  leverages transactional isolation to never work a job too
  early, and transactional durability to never let one drop.
hn_link: https://news.ycombinator.com/item?id=15294722
---

Background jobs are one of the most common patterns in web
programming, and for good reason. Slow API calls and other
heavy lifting is deferred to out-of-band workers so that a
user's request is executed as quickly as possible. In web
services, fast is a feature.

But when it comes to working with background jobs in
conjunction with ACID transactions of the likes you'd find
in Postgres, MySQL, or SQL Server, there are a few sharp
edges that aren't immediately obvious. To demonstrate,
let's take a simple workflow that starts a transaction,
executes a few DB operations, and queues a job somewhere in
the middle:

``` ruby
DB.transaction do |t|
  db_op1(t)
  queue_job()
  db_op2(t)
end
```

It's not easy to spot, but if your queue is fast, the job
enqueued by `queue_job()` is likely to fail. A worker
starts running it before its enclosing transaction is
committed, and it fails to access data that it expected to
be available.

As an easy example, imagine `db_op1()` inserts a user
record. `queue_job()` puts a job in the queue to retrieve
that record, and add that user's email address (along with
a unique internal ID) to an email whitelist managed by
another service. A background worker dequeues the job, but
finds that the user record it's looking for is nowhere to
be found in the database.

!fig src="/assets/job-drain/job-failure.svg" caption="A job failing because the data it relies on is not yet committed."

A related problem are transaction rollbacks. In these cases
data is discarded completely, and jobs inserted into the
queue will _never_ succeed no matter how many times they're
retried.

## For every complex problem ... (#complex-problem)

Sidekiq has [a FAQ on this exact subject][sidekiq]:

> _Why am I seeing a lot of "Can't find ModelName with
> ID=12345" errors with Sidekiq?_
>
> Your client is creating the Model instance within a
> transaction and pushing a job to Sidekiq. Sidekiq is
> trying to execute your job before the transaction has
> actually committed. Use Rails's `after_commit :on =>
> :create` hook or move the job creation outside of the
> transaction block.

Not to pick on Sidekiq in particular (you can find similar
answers and implementations all over the web), but this
solution solves one problem only to introduce another.

If you queue a job _after_ a transaction is committed, you
run the risk of your program crashing after the commit, but
before the job makes it to the queue. Data is persisted, but
the background work doesn't get done. It's a problem that's
less common than the one Sidekiq is addressing, but one
that's far more nefarious; you almost certainly won't
notice when it happens.

Other common solutions are equally as bad. For example,
another well-worn pattern is to allow the job's first few
tries to fail, and rely on the queue's retry scheme to
eventually push the work through at some point after the
transaction has committed. The downsides of this
implementation is that it thrashes needlessly (lots of
wasted work is done) and throws a lot of unnecessary
errors.

## Transactions as gates (#transactions-as-gates)

We can dequeue jobs gracefully by using a
_transactionally-staged job drain_.

With this pattern, jobs aren't immediately sent to the job
queue. Instead, they're staged in a table within the
relational database itself, and the ACID properties of the
running transaction keep them invisible until they're ready
to be worked. A secondary ***enqueuer*** process reads the
table and sends any jobs it finds to the job queue before
removing their rows.

Here's some sample DDL for what a `staged_jobs` table might
look like:

``` sql
CREATE TABLE staged_jobs (
    id       BIGSERIAL PRIMARY KEY,
    job_name TEXT      NOT NULL,
    job_args JSONB     NOT NULL
);
```

And here's what a simple enqueuer implementation that sends
jobs through to Sidekiq:

``` ruby
# Only one enqueuer should be running at any given time.
acquire_lock(:enqueuer) do

  loop do
    # Need at least repeatable read isolation level so that our DELETE after
    # enqueueing will see the same jobs as the original SELECT.
    DB.transaction(isolation_level: :repeatable_read) do
      jobs = StagedJob.order(:id).limit(BATCH_SIZE)

      unless jobs.empty?
        jobs.each do |job|
          Sidekiq.enqueue(job.job_name, *job.job_args)
        end

        StagedJob.where(Sequel.lit("id <= ?", jobs.last.id)).delete
      end
    end

    # If `staged_jobs` was empty, sleep for some time so
    # we're not continuously hammering the database with
    # no-ops.
    sleep_with_exponential_backoff
  end

end
```

Transactional isolation means that the enqueuer is unable
to see jobs that aren't yet commmitted (even if they've
been inserted into `staged_jobs` by an uncommitted
transaction), so jobs are never worked too early.

!fig src="/assets/job-drain/transaction-isolation.svg" caption="Jobs are invisible to the enqueuer until their transaction is committed."

It's similarly protected against rollbacks. If a job is
inserted within a transaction that's subsequently
discarded, the job is discarded with it.

The enqueuer is also totally resistant to job loss. Jobs
are only removed _after_ they're successfully transmitted
to the queue, so even if the worker dies partway through,
it will pick back up again and send along any jobs that it
missed. _At least once_ delivery semantics are guaranteed.

!fig src="/assets/job-drain/job-drain.svg" caption="Jobs being sequestered in a staging table and enqueued when they're ready to be worked."

## Advantages over in-database queues (#in-database-queues)

[Delayed_job][delayedjob], [que][que], and
[queue_classic][queueclassic] use a similar transactional
mechanic to keep jobs hidden, and take it even a step
further by having workers dequeue jobs directly from within
the database.

This is workable at modest to medium scale, but the frantic
pace at which workers try to lock jobs doesn't scale very
well for a database that's experiencing considerable load.
For Postgres in particular, [long-running
transactions](/postgres-queues) greatly increase the amount
of time it takes for workers to find a job that they can
lock, and this can lead to the job queue spiraling out of
control.

The transactionally-staged job drain avoids this problem by
selecting primed jobs in bulk and feeding them into another
store like Redis that's better-suited for distributing jobs
to competing workers.

[delayedjob]: https://github.com/collectiveidea/delayed_job
[que]: https://github.com/chanks/que
[queueclassic]: https://github.com/QueueClassic/queue_classic
[sidekiq]: https://github.com/mperham/sidekiq/wiki/FAQ#why-am-i-seeing-a-lot-of-cant-find-modelname-with-id12345-errors-with-sidekiq
