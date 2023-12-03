+++
hook = "Introducing River, a Postgres-based job queue designed for resilience and correctness through strong transactional guarantees."
#image = "/assets/images/river/bishop-pass.jpg"
#image = "/assets/images/river/bishop-pass-bw.jpg"
#image = "/assets/images/river/evolution-valley.jpg"
#image = "/assets/images/river/kumano.jpg"
#image = "/assets/images/river/seldon-pass.jpg"
image = "/assets/images/river/shrine.jpg"
#image = "/assets/images/river/ten-waterfalls.jpg"
location = "San Francisco"
published_at = 2023-11-20T06:18:48-08:00
title = "River: a Fast, Robust Job Queue for Go + Postgres"
hn_link = "https://news.ycombinator.com/item?id=38349716"
+++

Years ago I wrote about [my trouble with a job queue in Postgres](/postgres-queues), in which table bloat caused by long-running queries slowed down the workers' capacity to lock jobs as they hunted across millions of dead tuples trying to find a live one.

A job queue in a database can have sharp edges, but I'd understated in that writeup the benefits that came with it. When used well, transactions and background jobs are a match made in heaven and completely sidestep a whole host of distributed systems problems that otherwise don't have easy remediations.

Consider:

* In a transaction, a job is emitted to a Redis-based queue and picked up for work, but the transaction that emitted it isn't yet committed, so none of the data it needs is available. The job fails and will need to be retried later.

<div class="ml-auto w-10/12"><img src="/assets/images/river/data-not-visible.svg" alt="Job failure because data is not yet visible"></div>

---

* A job is emitted from a transaction which then rolls back. The job fails and will also fail every subsequent retry, pointlessly eating resources despite never being able to succeed, eventually landing the dead letter queue.

<div class="ml-auto w-10/12"><img src="/assets/images/river/data-roll-back.svg" alt="Job failure because data rolled back"></div>

---

* In an attempt to work around the data visibility problem, a job is emitted to Redis _after_ the transaction commits. But there's a brief moment between the commit and job emit where if the process crashes or there's a bug, the job is gone, requiring manual intervention to resolve (if it's even noticed).

<div class="ml-auto w-10/12"><img src="/assets/images/river/job-emit-failure.svg" alt="Job post-transaction emit failure"></div>

---

* If both queue and store are non-transactional, all of the above and more. Instead of data not being visible, it may be that it's in a partially ready state. If a job runs in the interim, all bets are off.

<div class="ml-auto w-10/12"><img src="/assets/images/river/data-not-complete.svg" alt="Job failure because data is not complete"></div>

---

Work in a transaction has other benefits too. Postgres' [`NOTIFY`](https://www.postgresql.org/docs/current/sql-notify.html) respects transactions, so the moment a job is ready to work a job queue can wake a worker to work it, bringing the mean delay before work happens down to the sub-millisecond level.

Despite our operational trouble, we never did replace our database job queue at Heroku. The price of switching would've been high, and despite blemishes, the benefits still outweighed the costs. I then spent the next six years staring into a maelstrom of pure chaos as I worked on a non-transactional data store. No standard for data consistency was too low. Code was a morass of conditional statements to protect against a million possible (and probable) edges where actual state didn't line up with expected state. Job queues "worked" by brute force, bludgeoning jobs through until they could reach a point that could be tacitly called "successful".

I also picked up a Go habit to the point where it's now been my language of choice for years now. Working with it professionally during that time, there's been more than a few moments where I wished I had a good framework for transactional background jobs, but didn't find any that I particularly loved to use.

## River is born (#river-is-born)

So a few months ago, [Blake](https://github.com/bgentry) and I did what one should generally never do, and started writing a new job queue project built specifically around Postgres, Go, and our favorite Go driver, [pgx](https://github.com/jackc/pgx). And finally, after long discussions and much consternation around API shapes and implementation approaches, it's ready for beta use.

I'd like to introduce River ([GitHub link](https://github.com/riverqueue/river)), a job queue for building fast, airtight applications.

<a href="https://riverqueue.com"><img src="/assets/images/river/river-home.png" srcset="/assets/images/river/river-home@2x.png 2x, /assets/images/river/river-home.png 1x" alt="Screen shot of River home page" class="rounded-3xl"></a>

### Designed for generics (#generics)

One of the relatively new features in Go (since 1.18) that we really wanted to take full advantage of was the use of generics. A river worker takes a `river.Job[JobArgs]` parameter that provides strongly typed access to the arguments within:

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

No raw JSON blobs. No `json.Unmarshal` boilerplate in every job. No type conversions. 100% reflect-free.

Jobs are raw Go structs with no embeds, magic, or shenanigans. Only a `Kind` implementation that provides a unique, stable string to identify the job as it round trips to and from the database:

```go
type SortArgs struct {
    // Strings is a slice of strings to sort.
    Strings []string `json:"strings"`
}

func (SortArgs) Kind() string { return "sort" }
```

Beyond the basics, River supports batch insertion, error and panic handlers, periodic jobs, subscription hooks for telemetry, unique jobs, and a host of other features.

Job queues are never really done, but we're pretty proud of the API design and initial feature set. Check out [the project's README](https://github.com/riverqueue/river) and [getting started guide](https://riverqueue.com/docs).

### With performance in mind (#performance)

One of the reasons we like to write things in Go is that it's fast. We wanted River to be a good citizen of the ecosystem and designed it to use fast techniques where we could:

* It takes advantage of pgx's implementation of Postgres' binary protocol, avoiding a lot marshaling to and parsing from strings.

* It minimizes round trips to the database, performing batch selects and updates to amalgamate work.

* Operations like bulk job insertions make use of [`COPY FROM`](https://www.postgresql.org/docs/current/sql-copy.html) for efficiency.

We haven't even begun to optimize it so I won't be showing any benchmarks (which tend to be misleading anyway), but on my commodity MacBook Air it works ~10k trivial jobs a second. It's not slow.

## What's different now? (#whats-different)

You might be thinking: Brandur, you've had trouble with job queues in databases before. Now you're promoting one. Why?

A few reasons. The first is, as described above, transactions are really _just a really good idea_. Maybe _the best_ idea in robust service design. For the last few years I've been putting my money where my mouth is and building a service modeled entirely around transactions and strong data constraints. Data inconsistencies are still possible, but especially in a relative sense, they functionally don't exist. The amount of time this saves operators from having to manually mess around in consoles fixing things cannot be overstated. It's the difference between night and day.

### Single dependency stacks (#single-dependency-stacks)

Another reason is that dependency minimization is great. I've written previously about how at work [we run a single dependency stack](/fragments/single-dependency-stacks). No ElastiCache, no Redis, no bespoke queueing components, just Postgres. If there's a problem with Postgres, we can fix it. No need to develop expertise in how to operate rarely used, black box systems.

This idea isn't unique. An interesting development in Ruby on Rails 7.1 is the addition of [Solid Cache](https://github.com/rails/solid_cache), which 37 Signals uses to cache in the same database that they use for the rest of their data (same database, but different instances of it of course). Ten years ago this would've made little sense because you'd want a hot cache that'd serve content from memory only, but advancements in disks (SSDs) has been so great that they measured a real world difference in the double digits (25-50%) moving their cache from Redis to MySQL, but with a huge increase in cache hits because a disk-based system allows cache space to widen expansively.

### Ruby non-parallelism (#ruby-non-parallelism)

A big part of our queue problem at Heroku was the design of the specific job system we were using, and Ruby deployment. Because Ruby doesn't support real parallelism, it's commonly deployed with a [process forking model](/nanoglyphs/027-15-minutes) to maximize performance, and this was the case for us. Every worker was its own Ruby process operating independently.

This produced a lot of contention and unnecessary work. Running independently, every worker was separately competing to lock every new job. So for _every_ new job to work, _every_ worker contended with _every other_ worker and iterated millions of dead job rows _every_ time. That's a lot of inefficiency.

A River cluster may run with many processes, but there's orders of magnitude more parallel capacity within each as individual jobs are run on goroutines. A producer inside each process consolidates work and locks jobs for all its internal executors, saving a lot of grief. Separate Go processes may still contend with each other, but many fewer of them are needed thanks to superior intra-process concurrency.

### Improvements in Postgres (#postgres-improvements)

During my last queue problems we would've been using Postgres 9.4. We have the benefits of nine new major versions since then, which have brought a lot of optimizations around performance and indexes.

* The most important for a queue was the addition of [`SKIP LOCKED`](https://www.2ndquadrant.com/en/blog/what-is-select-skip-locked-for-in-postgresql-9-5/) in 9.5, which lets transactions find rows to lock with less effort by skipping rows that are already locked. This feature is old (although no less useful) now, but we didn't have it at the time.

* Postgres 12 brought in `REINDEX CONCURRENTLY`, allowing queue indexes to be rebuilt periodically to remove detritus and bloat.

* Postgres 13 added [B-tree deduplication](https://www.postgresql.org/docs/13/btree-implementation.html#BTREE-DEDUPLICATION), letting indexes with low cardinality (of which a job queue has multiple of) be stored much more efficiently.

* Postgres 14 brought in an optimization to [skip B-tree splits](https://www.postgresql.org/docs/14/btree-implementation.html#BTREE-DELETION) by removing expired entries as new ones are added. Very helpful for indexes with a lot of churn like a job queue's.

And I'm sure there's many I've forgotten. Every new Postgres release brings dozens of small improvements and optimizations, and they add up.

Also exciting is the [potential addition of a transaction timeout setting](https://www.postgresql.org/message-id/CAAhFRxiQsRs2Eq5kCo9nXE3HTugsAAJdSQSmxncivebAxdmBjQ@mail.gmail.com). Postgres has timeouts for individual statements and being idle in a transaction, but not for the total duration of a transaction. Like with many OLTP operations, long-lived transactions are hazardous for job queues, and it'll be a big improvement to be able to put an upper bound them.

## Try it (#try-it)

Anyway, [check out River](https://riverqueue.com/) (see also the [GitHub repo](https://github.com/riverqueue/river) and [docs](https://riverqueue.com/docs)) and we'd appreciate it if you helped kick the tires a bit. We prioritized getting the API as polished as we could (we're _really_ trying to avoid a `/v2`), but are still doing a lot of active development as we refactor internals, optimize, and generally nicen things up.