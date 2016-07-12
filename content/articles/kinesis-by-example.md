---
hook: Splitting and merging in action.
location: San Francisco
published_at: 2015-03-19T07:27:45Z
title: Kinesis Shard Splitting & Merging by Example
---

The Kinesis developer guide covers shard [splitting and merging from a high-level](http://docs.aws.amazon.com/kinesis/latest/dev/kinesis-using-sdk-java-resharding.html), but I find that it's occasionally helpful to help solidify these types of advanced topics with examples. Here we'll walk through what the most basic splitting and merging operations look like on a Kinesis stream to get a better feel for the concepts.

First of all, I start out with a stream called `split-merge-test` that has a single shard. It's come online and is in an `ACTIVE` state:

``` sh
$ aws kinesis describe-stream --stream-name split-merge-test
{
    "StreamDescription": {
        "StreamStatus": "ACTIVE",
        "StreamName": "split-merge-test",
        "StreamARN": "arn:aws:kinesis:us-east-1:551639669466:stream/split-merge-test",
        "Shards": [
            {
                "ShardId": "shardId-000000000000",
                "HashKeyRange": {
                    "EndingHashKey": "340282366920938463463374607431768211455",
                    "StartingHashKey": "0"
                },
                "SequenceNumberRange": {
                    "StartingSequenceNumber": "49548859072970256769156879668947671610661756289899560962"
                }
            }
        ]
    }
}
```

Note above that the shard has a `HashKeyRange` assigned to it that starts at zero and ends at `340282366920938463463374607431768211455`. When a record is sent into a Kinesis stream, a basic hash function is applied to its partition key. The result of this function maps the record to one of the stream's shards based on the hash key range that each shard handles. A stream's total capacity is increased by subdividing the hash key range on an existing shard so that it maps to more shards than it did before.

## Splitting (#splitting)

Splitting a shard is a manual process in that an operator must decide how to divide up its total hash space between the new shards that will be created. I've decided to split mine evenly between two new shards, so I perform some basic arithmetic on my `EndingHashKey` to find the halfway point between it and zero:

``` ruby
$ irb
irb(main):001:0> 340282366920938463463374607431768211455 / 2
=> 170141183460469231731687303715884105727
```

Now that we have our answer, let's proceed to perform the split:

``` sh
$ aws kinesis split-shard --stream-name split-merge-test \
  --shard-to-split shardId-000000000000 \
  --new-starting-hash-key 170141183460469231731687303715884105727
```

The stream goes into `UPDATING` status. The shards look as they did before because the work to change their topology is still in-progress:

``` sh
$ aws kinesis describe-stream --stream-name split-merge-test
{
    "StreamDescription": {
        "StreamStatus": "UPDATING",
        "StreamName": "split-merge-test",
        "StreamARN": "arn:aws:kinesis:us-east-1:551639669466:stream/split-merge-test",
        "Shards": [
            {
                "ShardId": "shardId-000000000000",
                "HashKeyRange": {
                    "EndingHashKey": "340282366920938463463374607431768211455",
                    "StartingHashKey": "0"
                },
                "SequenceNumberRange": {
                    "StartingSequenceNumber": "49548859072970256769156879668947671610661756289899560962"
                }
            }
        ]
    }
}
```

A few seconds later, we can see the results of our changes. There are a few key things to note below:

* The hash key range of shards is **immutable**. When we split a shard, our "parent" is still available but has now entered a state called `CLOSED` (`shardId-000000000000` in this example). Its entire hash key range has been taken over by its children, `shardId-000000000001` and `shardId-000000000002`. A `CLOSED` shard is easily identifiable by the presence of an `EndingSequenceNumber`.
* The stream is once again `ACTIVE` now that updates are finished.
* Child shards have a pointer called `ParentShardId` back to the parent that they split from so that some history is maintained.
* The stream's sequence number jumped quite a bit during the split, by about 10^48 in fact. This is slightly less impressive when you take into account that the sequence jumps by about 10^24 between two normal record insertions, but this is quite a bit bigger than even that.

``` sh
$ aws kinesis describe-stream --stream-name split-merge-test
{
    "StreamDescription": {
        "StreamStatus": "ACTIVE",
        "StreamName": "split-merge-test",
        "StreamARN": "arn:aws:kinesis:us-east-1:551639669466:stream/split-merge-test",
        "Shards": [
            {
                "ShardId": "shardId-000000000000",
                "HashKeyRange": {
                    "EndingHashKey": "340282366920938463463374607431768211455",
                    "StartingHashKey": "0"
                },
                "SequenceNumberRange": {
                    "EndingSequenceNumber": "49548859072981407141756144980517230543978492779512725506",
                    "StartingSequenceNumber": "49548859072970256769156879668947671610661756289899560962"
                }
            },
            {
                "ShardId": "shardId-000000000001",
                "HashKeyRange": {
                    "EndingHashKey": "170141183460469231731687303715884105726",
                    "StartingHashKey": "0"
                },
                "ParentShardId": "shardId-000000000000",
                "SequenceNumberRange": {
                    "StartingSequenceNumber": "49548859213219643322715968606065803827347328807764754450"
                }
            },
            {
                "ShardId": "shardId-000000000002",
                "HashKeyRange": {
                    "EndingHashKey": "340282366920938463463374607431768211455",
                    "StartingHashKey": "170141183460469231731687303715884105727"
                },
                "ParentShardId": "shardId-000000000000",
                "SequenceNumberRange": {
                    "StartingSequenceNumber": "49548859213241944067914499229207339545619977169270734882"
                }
            }
        ]
    }
}
```

It may also be worth pointing out that although `shardId-000000000000` is considered to be `CLOSED` now, as the last records that it contains leave Kinesis' retention window it will transition from `CLOSED` to `EXPIRED`. When it does, no further records can ever be retrieved from the shard.

## Merging (#merging)

Now let's see what happens when we merge the two shards back together. A merge operation takes two shards as parameters: (1) the main shard to merge, and (2) the adjacent shard that will be mixed into it. Note that the use of the word "adjacent" here is not an accident; because of the way that Kinesis shards handle hash key ranges, only two shards that handle ranges that are contiguous can be merged back together.

``` sh
$ aws kinesis merge-shards --stream-name split-merge-test \
  --shard-to-merge shardId-000000000001 \
  --adjacent-shard-to-merge shardId-000000000002
```

As before, our stream enters `UPDATING`, but does not yet reflect our changes:

``` sh
$ aws kinesis describe-stream --stream-name split-merge-test
{
    "StreamDescription": {
        "StreamStatus": "UPDATING",
        "StreamName": "split-merge-test",
        "StreamARN": "arn:aws:kinesis:us-east-1:551639669466:stream/split-merge-test",
        "Shards": [
            {
                "ShardId": "shardId-000000000000",
                "HashKeyRange": {
                    "EndingHashKey": "340282366920938463463374607431768211455",
                    "StartingHashKey": "0"
                },
                "SequenceNumberRange": {
                    "EndingSequenceNumber": "49548859072981407141756144980517230543978492779512725506",
                    "StartingSequenceNumber": "49548859072970256769156879668947671610661756289899560962"
                }
            },
            {
                "ShardId": "shardId-000000000001",
                "HashKeyRange": {
                    "EndingHashKey": "170141183460469231731687303715884105726",
                    "StartingHashKey": "0"
                },
                "ParentShardId": "shardId-000000000000",
                "SequenceNumberRange": {
                    "StartingSequenceNumber": "49548859213219643322715968606065803827347328807764754450"
                }
            },
            {
                "ShardId": "shardId-000000000002",
                "HashKeyRange": {
                    "EndingHashKey": "340282366920938463463374607431768211455",
                    "StartingHashKey": "170141183460469231731687303715884105727"
                },
                "ParentShardId": "shardId-000000000000",
                "SequenceNumberRange": {
                    "StartingSequenceNumber": "49548859213241944067914499229207339545619977169270734882"
                }
            }
        ]
    }
}
```

And finally the stream re-enters its `ACTIVE` state with our new merged shard. It's worth pointing out that:

* Like before with our split, closed shards `shardId-000000000001` and `shardId-000000000002` are still around, but now have an `EndingSequenceNumber` to indicate that they are closed.
* The new shard `shardId-000000000003` remembers its history. It points back to its `ParentShardId`, as well as the `AdjacentParentShardID` that also helped to derive it.

``` sh
$ aws kinesis describe-stream --stream-name split-merge-test
{
    "StreamDescription": {
        "StreamStatus": "ACTIVE",
        "StreamName": "split-merge-test",
        "StreamARN": "arn:aws:kinesis:us-east-1:551639669466:stream/split-merge-test",
        "Shards": [
            {
                "ShardId": "shardId-000000000000",
                "HashKeyRange": {
                    "EndingHashKey": "340282366920938463463374607431768211455",
                    "StartingHashKey": "0"
                },
                "SequenceNumberRange": {
                    "EndingSequenceNumber": "49548859072981407141756144980517230543978492779512725506",
                    "StartingSequenceNumber": "49548859072970256769156879668947671610661756289899560962"
                }
            },
            {
                "ShardId": "shardId-000000000001",
                "HashKeyRange": {
                    "EndingHashKey": "170141183460469231731687303715884105726",
                    "StartingHashKey": "0"
                },
                "ParentShardId": "shardId-000000000000",
                "SequenceNumberRange": {
                    "EndingSequenceNumber": "49548859213230793695315233917635362760664090379986927634",
                    "StartingSequenceNumber": "49548859213219643322715968606065803827347328807764754450"
                }
            },
            {
                "ShardId": "shardId-000000000002",
                "HashKeyRange": {
                    "EndingHashKey": "340282366920938463463374607431768211455",
                    "StartingHashKey": "170141183460469231731687303715884105727"
                },
                "ParentShardId": "shardId-000000000000",
                "SequenceNumberRange": {
                    "EndingSequenceNumber": "49548859213253094440513764540776898478936738741492908066",
                    "StartingSequenceNumber": "49548859213241944067914499229207339545619977169270734882"
                }
            },
            {
                "ShardId": "shardId-000000000003",
                "HashKeyRange": {
                    "EndingHashKey": "340282366920938463463374607431768211455",
                    "StartingHashKey": "0"
                },
                "ParentShardId": "shardId-000000000001",
                "AdjacentParentShardId": "shardId-000000000002",
                "SequenceNumberRange": {
                    "StartingSequenceNumber": "49548859483727682580892427312894066474572005964670566450"
                }
            }
        ]
    }
}
```

Further splits and merges will all follow same pattern, leaving behind a trail of dead shards that act as a historical record to follow the lifecycle of the stream. The reason behind this design might not be immediately obvious, but sufficed to say that the immutable property of a shard's hash range is important in helping to guarantee that records can be consumed in-order even across a merge or split. We'll leave a more detailed explanation on this topic to a future article.
