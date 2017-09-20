---
title: Transactionally Staged Job Drains in Postgres
published_at: 2017-09-20T01:10:26Z
location: Calgary
hook: Building a robust background worker system that
  leverages transactional isolation to never work a job too
  early, and transactional durabiity to never let one drop.
---

Background jobs are one of the most common patterns in web
programming, and for good reason. Heavy lifting can is
deferred to an out-of-band worker so that a user's request
is executed as quickly as possible. When it comes to web
services, fast is a feature.

But when it comes to background jobs that are queued within
an ACID transaction of the likes you'd find in Postgres,
MySQL, or SQL Server, there are a few sharp edges that
aren't immediately obvious.

To demonstrate, let's take a simple workflow that starts a
transaction, executes a few DB operations, and queues a job
somewhere in the middle:

``` ruby
DB.transaction do |t|
  db_op1(t)
  queue_job()
  db_op2(t)
end
```

It's not easy to spot, but if you're running a reasonably
fast job queue, that background job is likely to fail. A
worker starts running it before its enclosing transaction
is committed, and it fails to access data that it assumed
would be available.

As an easy example, imagine `db_op1` inserts a user record.
`queue_job` puts a job in the queue to add that user's
email address (along with a unique internal ID) to an email
whitelist managed by an external service. A background
worker tries to do that work, but finds that the user
record is nowhere to be found in the database.

!fig src="/assets/job-drain/job-failure.svg" caption="A job failing because the data it relies on is not yet committed."

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

Not to pick on Sidekiq in particular (your can find similar
answers and implementations all over the web), but this
isn't a very good solution to the problem.

If you create a job _after_ a transaction is committed, you
run the risk of your program crashing between the commit
and when the job is enqueued. Data is persisted, but the
background work doesn't get done. It won't happen very
often, but you won't notice when it does.

Other common solutions are equally as bad. For example,
it's also common to allow the job's first few tries to
fail, and rely on a retry that happens sometime after the
transaction is committed to see the work through to
completion. This implementation thrashes needlessly and
throws a lot of unnecessary errors.

## Transactions as gates (#transactions-as-gates)

We can solve this problem elegantly using a
_transactionally-staged job drain_.

With this pattern, jobs aren't immediately sent to the job
queue. Instead, they're staged in a table within the
relational database itself, and the ACID properties of the
running transaction keep them invisible until they're ready
to be worked. A secondary ***enqueuer*** process reads the
table and sends any jobs it finds to the job queue before
removing their rows.

Here's some sample DDL for what a job staging table might
look like:

``` sql
CREATE TABLE staged_jobs (
    id       BIGSERIAL PRIMARY KEY,
    job_name TEXT      NOT NULL,
    job_args JSONB     NOT NULL
);
```

And here's what a simple enqueuer implementation:

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

Transactional isolation means that the enqueuer is unable
to see jobs that aren't yet commmitted (even if they've
been inserted), so jobs are never worked too early.

!fig src="/assets/job-drain/transaction-isolation.svg" caption="Jobs are invisible to the enqueuer until their transaction is committed."

The enqueuer is also totally resistant to job loss. Jobs
are only removed _after_ they're successfully transmitted
to the queue, so even if the worker dies partway through,
it will pick back up again and send any jobs that it
missed. _At least once_ delivery semantics are guaranteed.

!fig src="/assets/job-drain/job-drain.svg" caption="Jobs being sequestered in a staging table until and enqueued when they're ready to be worked."

## Advantages over in-database queues (#in-database-queues)

[Delayed_job][delayedjob], [que][que], and
[queue_classic][queueclassic] use a similar transaction
mechanism, but take it a step further by having workers
dequeue jobs directly from within the database.

This is workable at modest scale, but the frantic pace at
which workers will try to lock jobs doesn't scale very well
for a database that's experiencing diverse sets of load.
For Postgres in particular, [long-running
transactions](/postgres-queues) can greatly increase the
amount of time it takes for workers to find a visible job
that they can lock, and this can lead to an extremely
degenerate job queue.

The transactionally-staged job drain avoids this problem by
selecting primed jobs in bulk and feeding them into another
store like Redis that's more suited for distributing jobs
to competing workers.

[delayedjob]: https://github.com/collectiveidea/delayed_job
[que]: https://github.com/chanks/que
[queueclassic]: https://github.com/QueueClassic/queue_classic
[sidekiq]: https://github.com/mperham/sidekiq/wiki/FAQ#why-am-i-seeing-a-lot-of-cant-find-modelname-with-id12345-errors-with-sidekiq
