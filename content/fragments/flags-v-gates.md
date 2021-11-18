+++
hook = "The different uses of feature flags, and building a new feature flags system."
published_at = 2021-11-18T19:50:38Z
title = "Flags v. gates"
+++

Most developers will be familiar with the concept of a feature flag, which is a system to allow a new feature to be rolled out and rolled back to/from a select subset of users or traffic, and without a deploy or rollback. They're a critical part of avoiding downtime and bugs in a system with a lot of users and still under active development.

Beyond features, flags can conveniently be reused to implement various operational knobs for the day-to-day running of the production stack, to be used during any emergencies that come about. For example, you may have a flag that starts aggressively shedding non-critical traffic, or a flag that temporarily blocks an abusive user.

And adjacent to both those uses, it may also be useful for other things like feature retirement or internal user settings. If you wanted to tweak a certain API behavior that's already in use by many existing users, you could flag all of them into a deprecation flag, expose the new behavior as the default, and worry about fully retiring it later.

At Stripe, we made a distinction between the different types of use that was a somewhat accidental result of past tech decisions. A flag for rollout or operational knob was called a "feature flag", while a flag for deprecated behaviors or internal settings was called a "gate". Feature flags were also given a subclassification of "feature" or "circuit breaker" so that we could more easily determine whether it was meant to be long-lived or not.

So:

* Feature rollout: **feature flag** of type _feature_.
* Operational knob: **feature flag** of type _circuit breaker_.
* Deprecated behavior: **gate**.
* Internal setting (or similar misc. uses): **gate**.

## Convergence (#convergence)

Feature flags got progressively more advanced over the years as people wanted better granularity in rolling out features. By the time I left they could:

* Be entirely off.
* Be entirely on.
* Be on for a specific list of IDs (usually, account IDs).
* Be on by a random percentage.
* Be on by a random percentage _based on ID_ (so passing the same ID always yields the same on/off result -- particularly useful for rolling out to random subset of users).
* Be on by a random percentage _plus_ a specific list of IDs.

Obviously quite useful, but a conceptual problem we eventually had is that feature flags became a complete superset of the functionality of gates. This was a problem especially for newer engineers, who would get (unsurprisingly) confused as to whether they should be using one or the other, and even senior engineers would regularly forget that the flag they're looking for is actually a gate (happened to me five times a week). Ideally, these would be just be one concept.

But it was a tall order to ever bring them back together -- and not for a conceptual reason, but for a technical one. In the implementation, each feature flag was stored as a single Mongo record that contained its metadata along with current state. If the flag contained a list of IDs, those were colocated right inside the record. Each API process loaded all flags in at startup (including all IDs), and periodically synced state from the DB. This made checking a flag very fast, at the cost of more memory overhead.

This was generally fine for feature rollouts where ID sets were small-ish, but wasn't workable when a flag needed to be activated for tens of thousands of users or more -- all of those would exist as a single giant array within one row, and a giant set in every running API process. Gates conversely were stored on an account, and therefore distributed amongst the tens of thousands of records on which they were activate.

All to say that the distinction between a flag and a gate was confusing and we all wished that we could make them a single unified concept, but for technical (and planning ranking) reasons, it's not likely to happen.

## Unification (#unification)

The reason I've been thinking about this lately is that I've been designing a feature flag system at Crunchy, and wanted to build something similarly powerful to what I had before but to avoid the flag/gate dichotomy.

I ended up with this database schema:

``` sql
CREATE TABLE flag (
    id uuid PRIMARY KEY DEFAULT gen_ulid(),
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    description varchar(20000) NOT NULL,
    fraction_enabled double precision
	    CHECK (fraction_enabled > 0 AND fraction_enabled <= 1),
    name varchar(200) NOT NULL UNIQUE,
    owner varchar(200) NOT NULL,
    kind varchar(200) NOT NULL,
    state varchar(200) NOT NULL,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE flag_token (
    id uuid PRIMARY KEY DEFAULT gen_ulid(),
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    flag_id uuid NOT NULL REFERENCES flag (id),
    token_uuid uuid NOT NULL,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL
);
```

Two tables:

* `flag` stores metadata about a feature flag like its name and description, along with a current `state`.
* `flag_token` allows specific IDs to be activated for a flag.

For now, a flag's state can be `on`, `off`, `random`, or `list`, and if it's `list`, we respect the set of IDs found in `flag_token`. I'll probably add other states like `random_by_token` and `random_and_list` later.

Like at Stripe, flags are loaded into API processes on startup and synced periodically. This still produces the potential problem of needing a lot of memory to store giant sets of tokens, but I expect it to take us a really long way because we're optimized memory-wise about as much as possible:

* Tokens are all UUIDs, and our UUIDs are packed byte arrays everywhere -- literally just a type alias for `[16]byte` in Go.
* Token lookups are done with a `map[uuid.UUID]struct{}`. Again, efficient because the UUIDs are byte arrays, and Go's special `struct{}` type has no memory overhead.
* We're using real in-process parallelism so a single Go process serving hundreds or thousands of requests can share the same set of flags and tokens.

We also have many orders of magnitude fewer users, so all told, it's unlikely to ever be a problem.
