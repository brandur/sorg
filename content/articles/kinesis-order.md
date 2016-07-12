---
hook: On guaranteeing order with the bulk put API of an event stream.
location: San Francisco
published_at: 2015-03-05T01:13:46Z
title: Guaranteeing Order with Kinesis Bulk Puts
---

Playing with Kinesis recently, we came across a problem of how to guarantee order when posting a set of records to its bulk API. This article summarizes it and talks about how although we never directly solved the problem that we thought we had, we were able to use a slightly altered approach to have the system meet some of the characteristics that we wanted to see.

The basic primitive to send a record into the Kinesis API is [`PutRecord`](http://docs.aws.amazon.com/kinesis/latest/APIReference/API_PutRecord.html) which allows a producer to put a single record into the stream. The API has an optional request parameter called `SequenceNumberForOrdering` that allows a consumer to pass in a previously-generated sequence number to guarantee that no matter what sequence number is generated for the record, it will be larger than the one you had before.

You can get a better feel for this idea with this sample dialog between a record producer and the Kinesis API:

```
REQUEST 1 [Producer]
PutRecord record1.

RESPONSE 1 [Kinesis]
OK. SequenceNumber="123".

REQUEST 2 [Producer]
PutRecord record2. SequenceNumberForOrdering="123".

RESPONSE 2 [Kinesis]
OK. SequenceNumber="124".
```

Kinesis also provides a bulk API that allows many records to be injected into the stream at once called [`PutRecords`](http://docs.aws.amazon.com/kinesis/latest/APIReference/API_PutRecords.html). The bulk API has a nice characteristic in that it allows up to 1000 records, or 1 MB, per second to be written. If one of your goals is great throughput, the economy of scale that you get by wrapping all your requests up into a single HTTP envelope and sending them all through at once is considerable.

Use of the bulk API introduces a problem though in that guaranteeing order becomes more difficult. Kinesis will try to order new records according to how they came in with your payload, but any given record in that request can fail, and the semantics around failures dictate that the even if a failure does occur, every non-failed record in the batch will succeed normally. Responsibility falls to the producer to detect such failures and retry them on a subsequent requests.

For example:

```
REQUEST 1 [Producer]
PutRecords
  - record1
  - record2
  - record3

RESPONSE 1 [Kinesis]
  - OK. SequenceNumber="123".
  - FAILED.
  - OK. SequenceNumber="124".

REQUEST 2 (retry failed) [Producer]
PutRecords
  - record2

RESPONSE 2 [Kinesis]
  - OK. SequenceNumber="125".
```

Here we try to post three records in order (`record1`, `record2`, `record3`), but they will end up out-of-order in the stream due to a failure (`record1` and `record3` staying ordered, but the failed `record2` being injected into the stream only on a retry).

So with `PutRecord` we can guarantee order at the cost of throughput, and if we use `PutRecords` we get improved throughput but without order. So the question is: is there a way that we can get both of these desirable characteristics?

## Sharding (#sharing)

Before we get there, let's briefly touch upon the concept of sharding with respect to a Kinesis stream. [As described in this architectural diagram](http://docs.aws.amazon.com/kinesis/latest/dev/key-concepts.html), a Kinesis stream is split into one or more shards for purposes of scalability; each shard has an upper limit on the amount of data that it can be written into it or read out of it (currently these limits at 1 MB/s and 2 MB/s respectively), so as the volume of data in a stream is increased, more shards can be added to achieve a form of horizontal scaling. Records within a shard are ordered according to how records were sent into them, and this order will be maintained when they're streamed out to a consumer. However, when producing to and consuming from multiple shards, no kind of ordering between shards can be guaranteed.

Producers control which shards they're producing to by specifying a [partition key](http://docs.aws.amazon.com/kinesis/latest/APIReference/API_PutRecord.html) along with any record they're sending into the stream; both the `PutRecord` and bulk `PutRecords` APIs support (and require) one. Partition keys are mapped through a hash function which will result in one of the stream's shards being selected in a deterministic fashion. As long as the total number of shards in a stream has not changed, a partition key will consistently map to the same shard no matter how many times it's reused.

So back to our original question: how can we guarantee that all records are consumed in the same order in which they're produced? The answer is that we can't, but that we shouldn't let that unfortunate reality bother us too much. Once we've scaled our stream to multiple shards, there's no mechanism that we can use to guarantee that records are consumed in order across the whole stream; only within a single shard. So instead of focusing on a global guarantee of ordering, we should instead try to to leverage techniques that will get us as much throughput as possible, and fall back to techniques that allow us to control for certain subsets of records where we deem it necessary.

## Sequential Puts per Partition Key; Bulk Otherwise (#per-partition)

To achieve the above, we're using a simple algorithm on our producers:

``` ruby
while records = more_records()
  records_to_post = records.
    group_by { |record| record.partition_key }.
    map { |_, partition_group| partition_group.first }
  kinesis.put_records(records_to_post)
end
```

Most of the time all pending records that need to be sent into the stream will be posted to Kinesis as a single bulk batch. However, if we find multiple records that have the same partition key, we only post the first one that was produced, and wait for the next cycle through to post any other events in the same partition.

Let's solidify this idea a little by using an example. In our production system, we pool records in a Postgres database before streaming them out to Kinesis. That database has a schema that looks like this:

```
=> SELECT partition_key, record_data FROM kinesis_records ORDER BY id;

            partition_key             |              record_data
--------------------------------------+---------------------------------------
 8a9e7a19-9fe1-49b2-9b42-591520784449 | {"resource":"app","action":"create"}
 d0d97986-0c90-404f-bccd-9ac6c27f9235 | {"resource":"app","action":"create"}
 8a9e7a19-9fe1-49b2-9b42-591520784449 | {"resource":"app","action":"update"}
 8a9e7a19-9fe1-49b2-9b42-591520784449 | {"resource":"app","action":"destroy"}
 b20d88bc-ba68-41e3-87cb-3a93cc619833 | {"resource":"app","action":"update"}
(5 rows)
```

When we want to select a batch of records to stream, we'll use an SQL query that partitions our pending records over `partition_key` and selects the first record for each partition:

```
=> SELECT partition_key, record_data FROM kinesis_records
WHERE id IN (
  SELECT MIN(id)
  FROM kinesis_records
  GROUP BY partition_key
)
ORDER BY id;

            partition_key             |              record_data
--------------------------------------+---------------------------------------
 8a9e7a19-9fe1-49b2-9b42-591520784449 | {"resource":"app","action":"create"}
 d0d97986-0c90-404f-bccd-9ac6c27f9235 | {"resource":"app","action":"create"}
 b20d88bc-ba68-41e3-87cb-3a93cc619833 | {"resource":"app","action":"update"}
(3 rows)
```

In the data set above, those first three records would all be posted in a single batch. The query would then run again and fetch the next record in the `8a9e7a19-...` sequence:

```
            partition_key             |              record_data
--------------------------------------+---------------------------------------
 8a9e7a19-9fe1-49b2-9b42-591520784449 | {"resource":"app","action":"update"}
```

That would be posted by itself in a second batch. The worker would then run one more time to fetch the final record for that partition and post it as a third batch:

```
            partition_key             |              record_data
--------------------------------------+---------------------------------------
 8a9e7a19-9fe1-49b2-9b42-591520784449 | {"resource":"app","action":"destroy"}
```

By partitioning over our known key and selecting the first result ordered by the table's sequential ID, we achieve the same effect as the pseudocode algorithm above, resulting in a set of records with unique partition keys that are safe to post in bulk even if a failure ends up ordering it in a way that we didn't intend.

## Partition Key Selection (#partition-key)

A side effect of this approach is that the selection of a partition key that's logical for your records becomes one of the most important concerns when starting to stream a new type of record. The partition key is the only mechanism available for controlling the order in which records are streamed to consumers, and some consideration must be taken when selecting partition keys to ensure that all records in the stream will play nicely together.

As an example, consider a GitHub-like service that would like to stream all repository-related events that occur within it through a Kinesis stream. We'd like to stream three types of events:

1. Create repository.
2. Destroy repository.
3. Commit to repository.

Repositories can be referenced through a combination of account and project name (e.g. `brandur/my-project` like you'd see after `github.com/`) and commits can be referenced by their SHA hash (e.g. `c0ab1e5c...`).

Say we have a consumer on the Kinesis stream that's maintaining some basic state like a cache of all repositories and commits that are known to the service. When a new repository is created, it will set a key for it in a Redis store, and when that repository is destroyed, it will remove that key.

We can see even in this basic example that if our cache consumer receives a create and destroy event for the same repository out of order, it will be left with invalid data. After receiving the destroy event first, it will remove a cache key that was never set, and then after receiving the misordered create event, it will set a cache key for a repository which is no longer valid. To ensure that these events are always received by all consumers in the correct order, we can set the partition key of the create and destroy events to the same value so that they'll end up on the same shard. In this case, the name of the repository that they're related to (`brandur/my-project`) is a good candidate for this shared key. Kinesis will consistently translate the repository name to a single shard, and all consumers listening to that shard will receive the events that are emitted from it in the expected order.

The same principle applies to streamed commits as well. It might be tempting at first glance to partition commits based on their SHA identifier (`c0ab1e5c`), but if we did so, we would open up the possibility of a consumer receiving a commit for a repository that doesn't exist because it's already processed a destroy event for that repository that came out of a different shard. We can solve this problem by assigning the partition key of commit events to the identifier of their parent repository (again, `brandur/my-project` instead of `c0ab1e5c`) so that they'll end up on the same shard as the rest of the repository's events.

With this system, we accept that consumers may not be consuming events in exactly the same order in which we emitted them, but we also know that order will be guaranteed when it matters. The result is much improved scalability in the way that the stream can be split into any number of shards and our consumers will continue to consume the stream correctly.
