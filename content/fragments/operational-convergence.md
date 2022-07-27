+++
hook = "Life without a REPL, and how to still be able to manipulate production which even has quite a few benefits over more one-off REPL-driven operations.."
published_at = 2022-07-25T15:14:36Z
title = "Operational convergence, for REPL-less languages like Go"
+++

First off, file this one under "maybe good idea, maybe horrible idea", but it's better than what we had before, so it might be useful to you too.

I've been primarily supporting a Go stack for about a year now. There's a lot of things I love about the language: compile speed, runtime speed, statically checked syntax that's heavy, but not too heavy, no breakages across language versions, excellent, lightweight toolchain.

But especially where web services are concerned that have many domain objects, it has a deficiency, and a major one: no REPL.

This makes operations in deployed environments difficult. A good production REPL isn't just something you use once every few months, it's something you use every day -- it's good for examining data, good for fixing data, good for poking at the internal state of online components, good for one-off operations, and good for debugging, amongst many more. I'll make a bold claim that our logging practices are _meticulous_ -- everything that runs [leaves a detailed paper trail](/nanoglyphs/025-logs) -- and yet it's often not enough, and I still find myself wishing often that I could more easily introspect a production object.

In the Go ecosystem, the apparent widely-agreed upon alternative to not having a REPL is as best as I can tell ... nothing. When I first started, I'd ask colleagues how do to something that seemed pretty run-of-the-mill, like how to delete an account. The answer: you can't. (Or rather, you can't without dozens of risky manual steps and bespoke one-off scripts.) It's not a productized feature. You could try and modify data in a `psql` shell, but can't perform a full teardown and cleanup that'd require application code to run. And these caveats also applied to any similar operation, and apply far beyond our own code to Go at large. The fact that this omission isn't something more Go developers talk about is somewhat astonishing -- I can't help but wonder what most people are doing to operate production.

In all fairness, even if Go did have a REPL (and it sort of does in the form of third party projects like Gore), it might still be a problem. Go code isn't conducive to being written from the hip -- practically every expression needs to be followed by an `if err != nil { ... }` check, syntax is generally verbose, and is rarely chainable (nothing approaches the compactness of something like `arr.map { |x| ... }.filter { |x| ... }.uniq` like you'd find in Ruby/Python/etc.). A REPL would make things more possible, but would still be a chore.

I don't necessarily have the world's best solution, but I do have _a_ solution that's been working reasonably well for us.

## First stab (#first-stab)

We initially started down the path of writing an admin API, with the idea in conjunction with an admin CLI or dashboard to accomplish what you need. It was somewhat workable, but major weaknesses were visible early on:

* It's a lot of code. In Go, basic CRUD scaffolding is potentially thousands of lines of code including tests, and the combination of parameters and endpoints you need to support your multitude of possible admin operations are endless. More code is more code to maintain -- it makes refactoring harder, and especially where code is used infrequently (like an admin API), is likely to contain undiscovered bugs.

* It's inflexible. It's more common than not to go and see if something is supported in the admin tools, only to find that it's not.

* When you finally try to do something with it, you won't remember how to do it. Minimal documentation (the normal state for internal tools at most companies) and no code completion means you're going to have an awkward time.

I tried another alternative briefly of writing a minimal amount of Ruby code _for my Go code_ to be able to run operations from a Ruby REPL (Ruby ORMs can provide most common operations given little code), and it kind of worked, but frankly, was not sustainable.

## Operational convergence (#operation-convergence)

But that still left me with a question in need of an answer.

One of Go's best features is excellent concurrency -- starting a couple extra goroutines is so negligible in cost that it's indistinguishable from free. One way we take advantage of this is to zip all our various types of background workers up into a single process -- a major advantage because it means spinning up a new one is just a few extra lines of code -- no need to configure a new infrastructure component and figure out how it's going to be deployed. Every time we deploy, the API and all its background workers go down together, gracefully finishing off any work that was in progress, then come back up, all in the span of a second or two. Any new workers come alive for the first time on the other side.

This got me thinking -- what if we had a series of small, well-defined background services whose job it was to notice inconsistencies or flags put down by operators, and upon noticing any, push state in the right direction to correct it.

For example, here's a small service (lets call them "convergers") that looks for recently archived accounts, and for each one, proceeds to archive any related objects:

``` go
var inconsistencies InconsistencyCheckerResult

accounts, err := queries.AccountGetArchivedSinceHorizon(ctx,
    time.Now().Add(-inconsistencyCheckerArchivedAtThreshold))
if err != nil {
    return nil, xerrors.Errorf("error getting recently archived accounts: %w", err)
}

for _, account := range accounts {
    // Runs a big query that archives a whole bunch of resources
    // associated with an account in a single pass.
    res, err := queries.AccountArchiveResourcesByID(ctx, account.ID)
    if err != nil {
        return nil, xerrors.Errorf("error archiving account resources: %w", err)
    }

    var anyChanges bool
    for _, resRow := range res {
        if len(resRow.IDs) > 0 {
            anyChanges = true
            break
        }
    }

    // We expect this this job to enact on archived accounts multiple
    // times while they're in the archived at window. Only log as an
    // inconsistency if anything new was archived so that any given
    // account is only logged as an inconsistency once.
    if anyChanges {
        inconsistency := NewInconsistency(account.ID, true)
        inconsistency.Extra = res

        inconsistencies.AccountsArchivedInPlatform =
            append(inconsistencies.AccountsArchivedInPlatform, inconsistency)
    }
}
```

I can then archive an account via `psql`:

``` sql
begin; -- the operator's most powerful anti-whoopsy tool

update account
set archived_at = now()
where email = 'TARGET'
returning *;

commit;
```

The account archiving service will wake up momentarily, notice the newly archived account, and make sure that other state is correct by archiving any other resources the account owned.

If any database clusters were amongst that account's resources, another converger will wake up and clean up their state by deprovisioning them with our backend state machine (which runs as a separate service).

Another converger wakes up periodically and makes sure that all subscriptions are correct. If a cluster changed size, changed plan, or was created without a subscription, subscriptions are opened and closed until it's in its correct state.

All are written to be idempotent. They're constantly trying to converge state all the time, and after tackling any specific inconsistency once, they won't run on it again because doing so would produce no additional effect. After an account I've archived is handled, convergers will still consider it a valid target, but won't do anything else.

They're also written to be resource efficient. Any particular one will do work in parallel where appropriate, but make sure to bound the number of child goroutines it might spin up using [`x/sync`'s `errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup) so that it's a considerate neighbor for other jobs that might be ongoing in the worker process. Each convergence check runs quickly (and with a context timeout in case it doesn't for some reason), and can therefore share a pool of database connections even where the maximum number of pooled connections is much less than the total number of convergers.

## Maybe apologism? Or, maybe not. (#maybe-apologism)

I'd be lying if I said that I didn't still want a formal REPL in Go paired with more concise syntax to use with it, but this technique does have some advantages compared to REPL-driven operations:

* All convergence code is first-class. It's versioned, tested, and deployed using the same care as anything else. Testing especially gives us very strong certainty that any convergence which is going to be run probably works, as opposed to something one-off you'd write which could easily be wrong, and in the worse case scenario have catastrophic side effects (we once had someone at a previous job use a REPL to accidentally delete the Europe region, which instantly broke everything).

* Compared to an admin interface, the code required to write a new converger is minimal. No API or CLI scaffolding is required -- just the core logic which is often fairly minimal along with some tests to check that it works.

* Once deployed, convergence is fully automated, meaning that there's no tooling that an operator has to figure out how to use. When some manual action is required to start off some process (like deleting an account), it's easily documented because it tends to only involve one step. Here's a screenshot of the docs I wrote for the account deprovisoner mentioned above:

{{Figure "Sample internal docs showing how to kick off a convergence process to offboard an account." (ImgSrcAndAltAndClass "/photographs/fragments/operational-convergence/internal-docs.png" "Internal docs" "overflowing")}}

* Lastly, there's a compounding effect to more convergers. Instead of someone figuring out how to do something in a REPL once and all that effort being lost when they don't document it (or maybe worse, they do document it, but the code involved drifts out of date and is wrong by the time someone else tries it), it's written into core so that other people can take advantage of it. Since it runs through the test suite, it tends to stay working.
