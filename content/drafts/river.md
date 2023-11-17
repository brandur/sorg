+++
hook = "TODO"
#image = "/assets/images/river/bishop-pass.jpg"
#image = "/assets/images/river/bishop-pass-bw.jpg"
#image = "/assets/images/river/evolution-valley.jpg"
#image = "/assets/images/river/kumano.jpg"
#image = "/assets/images/river/seldon-pass.jpg"
image = "/assets/images/river/shrine.jpg"
#image = "/assets/images/river/ten-waterfalls.jpg"
location = "San Francisco"
published_at = 2023-11-14T09:47:16-08:00
title = "River: a Fast, Robust Job Queue for Go + Postgres"
+++

Years ago I wrote about [my woes with a Postgres job queue](/postgres-queues), in which table bloat caused by long-running queries slowed down the workers' capacity to lock jobs as they hunted across millions of dead tuples before they found a live one.

A job queue in a database can have sharp edges, but what I'd left out of that piece are all the benefits that came with it. Transactions and background jobs are a match made in heaven and sidestep completely a whole host of queueing problems that most systems take for granted.

Consider:

* In a transaction, a job is emitted to a Redis-based queue and picked up for work, but the transaction that emitted it isn't yet committed so none of the data it needs is available. The job fails and will need to be retried later.

* A job is emitted from a transaction which then rolls back. The job fails and will also fail every subsequent retry, pointlessly eating resources despite never being able to succeed.

* To work around the data visibility problem, a job is emitted to Redis _after_ the transaction commits.  But there's a brief moment between the commit and job emit where if the process crashes, the job is gone.

* If both queue and store are non-transactional, all of these problems get even worse. Instead of data not being visible, it may be that it's in a partially ready state at which point all bets are off.

Work in a transaction has other nominal benefits too. Postgres' [`NOTIFY`](https://www.postgresql.org/docs/current/sql-notify.html) respects transactions, so the moment a job is ready to work a job queue can wake a worker to work it, bringing the delay before work happens down to the sub-millisecond level.

Despite its problems, we never did replace our database job queue at Heroku. The price of switching would've been high, and despite its blemishes, the benefits outweighed the costs. I then spent the next six years staring into a maelstrom of pure chaos as I worked on a non-transactional data store. No standard for data consistency was too low. Code was a morass of conditional statements to protect against a million possible (and probable) edges where actual state didn't line up with expected state. Job queues "worked" by brute force, bludgeoning jobs over and over until they could reach a point that could by tacitly considered "successful".

I also picked up a Go habit to the point where it's now been my language of choice for years now. Working with it professionally during that time, there's been more than a few moments where I wished I had a good framework for background jobs, but looked around, and found the existing options lacking.

## River is born (#river-is-born)

So a few months ago, [Blake]() and I did what one should generally never do, and started writing a new queueing project built specifically around Postgres, Go, and our favorite Go driver, [pgx](https://github.com/jackc/pgx). And finally, after many long discussions about API shapes and implementation approaches, it's ready for alpha use. [Introducing River](https://github.com/riverqueue/river), a job queue for building fast, airtight applications.

One of the relatively new features in Go (since 1.18) that we really wanted to bake in was the use of generics. A river workers looks like this, taking a `river.Job[JobArgs]` parameter that provides strongly typed access to the arguments within:

```go
type SortWorker struct {
    river.WorkerDefaults[SortArgs]
}

func (w *SortWorker) Work(ctx context.Context, job *river.Job[SortArgs]) error {
    sort.Strings(job.Args.Strings)
    fmt.Printf("Sorted strings: %+v\n", job.Args.Strings)
    return nil
}
```

No raw JSON blobs. No `json.Unmarshal` boilerplate in every job. No type conversions. Reflect-free.

Jobs are raw Go structs with no embeds, magic, or shenanigans. Only a `Kind` implementation that provides a unique, stable string to identify the job as it round trips to and from the database:

```go
type SortArgs struct {
    // Strings is a slice of strings to sort.
    Strings []string `json:"strings"`
}

func (SortArgs) Kind() string { return "sort" }
```

Beyond the basics, River supports batch insertion, error and panic handlers, periodic jobs, subscription hooks for telemetry,  unique jobs, and a host of other features.

Job queues are never really done, but we're pretty proud of the API design and initial feature set. Check out [the project's README](https://github.com/riverqueue/river) and [getting started guide](https://riverqueue.com/docs).

## What's different now? (#whats-different)

You might be thinking: Brandur, you've had trouble with job queues in databases before. Now you're promoting one. Why?

A few reasons. The first is, as described above, transactions are really _just a really good idea_. Maybe _the_ most important idea in robust service design. For the last few years I've been putting my money where my mouth is and building a service modeled entirely around transactions and strong data constraints. Data inconsistencies are still possible, but especially in a relative sense, they functionally don't exist. The amounted of time this saves operators from having to manually mess around in consoles fixing things cannot be overstated. It's the difference between night and day.

### Single dependency stacks (#single-dependency-stacks)

Another reason is that dependency minimization is great. I've written previously about how at work [we run a single dependency stack](/fragments/single-dependency-stacks). No ElastiCache, no Redis, no bespoke queueing components, just Postgres. If there's a problem with Postgres, we can fix it. No need to develop expertise in how to operate rarely used, black box systems.

Interestingly, this idea isn't unique. An interesting development in Ruby on Rails 7.1 is the addition of [Solid Cache](https://github.com/rails/solid_cache), which 37 Signals uses to cache in the same database that they use for the rest of their data (same database, but different instances of it of course). Ten years ago this would've made little sense because you'd want a hot cache that'd serve content from memory only, but advancements in disks (SSDs) has been so great that they measured a real world difference in the double digits (25-50%) moving their cache from Redis to MySQL, but with a huge increase in cache hits because a disk-space system allows the cache to widen expansively.

### Ruby non-parallelism (#ruby-non-parallelism)

A big part of our queue problem at Heroku was queue design and Ruby deployment. Because Ruby doesn't support real parallelism, it's commonly deployed with a [process forking model](/nanoglyphs/027-15-minutes) to maximize performance, and this was the case for our job queue. Every worker was an independent Ruby process.

This produced a lot of contention and unnecessary work. Running independently, every worker was separately competing to lock every new job. So for every new job to work, every worker contending with every other worker and iterating millions of dead job rows every time. That's a lot of inefficiency.

A River cluster may run with many processes, but there's orders of magnitude more parallel capacity within each as individual jobs are run on goroutines. A producer inside each process locks jobs for all its internal executors, saving a lot of grief. Processes may still contend with each other, but far fewer of them are needed.

### Improvements in Postgres (#postgres-improvements)

During my last queue problems we would've been using Postgres 9.4. We have the benefits of nine new major versions since then, which have brought a lot of optimizations around performance and indexes.

* Probably the most important for a queue was the addition of [`SKIP LOCKED`](https://www.2ndquadrant.com/en/blog/what-is-select-skip-locked-for-in-postgresql-9-5/) in 9.5, which lets transactions find rows to lock with less effort by skipping rows that are already locked. This feature is actually old now, but we didn't have it at the time.

* Postgres 12 brought in `REINDEX CONCURRENTLY`, allowing queue indexes to be rebuilt periodically to remove detritus and bloat.

* Postgres 13 added [B-tree deduplication](https://www.postgresql.org/docs/13/btree-implementation.html#BTREE-DEDUPLICATION), letting indexes with low cardinality (of which a job queue has multiple of) be stored much more efficiently.

* Postgres 14 brought in an optimization to [skip B-tree splits](https://www.postgresql.org/docs/14/btree-implementation.html#BTREE-DELETION) by removing expired entries as new ones are added. Very helpful for indexes with a lot of churn like a job queue's.

And I'm sure there's many I've forgotten.

Also exciting is the [potential addition of a transaction timeout setting](https://www.postgresql.org/message-id/CAAhFRxiQsRs2Eq5kCo9nXE3HTugsAAJdSQSmxncivebAxdmBjQ@mail.gmail.com). Postgres has timeouts for individual statements and being idle in a transaction, but not for the total duration of a transaction. Like with many OLTP operations, long-lived transactions are hazardous, and it'll be a big improvement to be able to put an upper bound on these things.

## Sum (#sum)

Anyway, [check out River](https://github.com/riverqueue/river) and help kick the tires a bit. We prioritized getting the API as polished as possible (we're _really_ trying to avoid a `v2`), but are still doing a lot of active development as we refactor internals and nicen things up.