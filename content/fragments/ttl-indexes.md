+++
hook = "An ode to my favorite feature in Mongo."
published_at = 2020-08-26T17:30:16Z
title = "TTL indexes in Mongo"
+++

I want to call out a great feature in Mongo: [TTL indexes](https://docs.mongodb.com/manual/core/index-ttl/). I've been critical of this database in the past, but I've got to give credit where credit is due, and TTL indexes are an extraordinary idea -- easily one of the best new database features in years.

A TTL index isn't used for traditional indexing functions like optimizing lookups or enforcing uniqueness. It's a very specific type of index that exists on a single field and has one purpose: removing documents that have reached the end of their pre-determined time to live (TTL). They're created by passing the special `expireAfterSeconds` option on index builds:

``` js
db.eventlog.createIndex(
    { "createdTime": 1 },
    { expireAfterSeconds: 3600 }
)
```

Once every 60 seconds Mongo runs a background task to remove any documents in the collection whose `createdTime` + `expiresAfterSeconds` is smaller than the current time.

The reason they're so useful is that cleaning out old data is one of the most common patterns in data design there is. Any application that reaches a certain size is eventually going to want to do it somewhere. Some examples of use at Stripe:

* Removing [idempotency keys](/idempotency-keys) after 24 hours.
* Removing records representing webhooks and the HTTP conversations we've had with servers receiving webhooks after a few weeks [1].

Over the years they've proven scalable and reliable. The number of records that they clean up in these large collections is mind-boggling, yet I can't remember a single production problem so far. Prior to TTL indexes we removed old idempotency keys manually (think `DELETE FROM ... WHERE createdTime < X`), and it was awful. Even with a sharding scheme to partition the work, the process was unreliable and constantly behind. We switched to TTL indexes, and haven't thought about it since. They're the kind of feature that just does its job, disappearing into the background to the point where you forget it's even there. The best kind of feature.

## Use `expireAfterSeconds` = 1 (#expire-one)

One adjustment to the manual's recommended usage that I'd advise is to just always use an `expireAfterSeconds` value of 1, then set your timestamps to the future time when you want them removed:

``` js
db.eventlog.createIndex(
    { "createdTime": 1 },
    { expireAfterSeconds: 1 }
)
```

This works better because:

* It's not uncommon to want to reconfigure expiry time (say upping retention from 3 months to 6). A value of 1 lets you make that change in application code, avoiding the need to issue specialized Mongo commands to modify the `expireAfterSeconds` value, or rebuild the index.
* It enables dynamic expiry time. Say for example that high importance documents should be retained for a year, but less important ones can expire out in a month.

My ulterior motive for calling out TTL indexes is to seed the idea in the minds of developers of other databases because this feature should exist _everywhere_. I don't know what the equivalent of a TTL index looks like in an RDMS/SQL, but I want it.

[1] Webhook records are kept around for a while to allow for a retry schedule and to make request/response information available to users to help with debugging.
