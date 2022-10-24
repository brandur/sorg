+++
hook = "Tracking improvements to Postgres partitioning since 2017. In just five years it's gone from a non-feature to a very good one."
published_at = 2022-10-05T17:41:49Z
title = "Partitioning in Postgres, 2022 edition"
+++

Partitioned tables aren't an everyday go to, but are invaluable in some cases, particularly when you have a high volume table that's expected to keep growing.

In Postgres, trying to remove old rows from a large, hot table is flirting with disaster. A long running query must iterate through and mark each one as dead, and even then nothing is reclaimed until an equally expensive vacuum runs through and frees space, and only when it's allowed to after rows are no longer visible to any other query in the system, whether they're making use of the large table or not. Each row removal land in the WAL, resulting in significant amplification.

But with partitions, deletion becomes a simple `DROP TABLE`. It executes instantly, and with negligible costs (partitioning [has other benefits](https://www.postgresql.org/docs/current/ddl-partitioning.html) too). The trade-off is maintenance. Not long ago there was no formal partitioning at all in Postgres -- it was entirely a user-level construct which needed all kinds of manual plumbing to run. And even once Postgres got support for native partitioning, things were always harder -- routing inserts/updates, adding/removing partitions, adding indexes, support for features like foreign keys and triggers, etc.

But, a lot of work has gone into improving the operator experience since partitioning was introduced. Here's a sprinkling of new features that have come into Postgres over the last five years:

* **Postgres 10:** Brings in the original [`CREATE TABLE ... PARTITION BY ...`](https://www.postgresql.org/message-id/flat/55D3093C.5010800@lab.ntt.co.jp) declarative partitioning commands.
* **Postgres 11:** Support for `PRIMARY KEY`, `FOREIGN KEY`, indexes, and triggers on partitioned tables.
* **Postgres 11:** `INSERT` on the parent partitioned table routes rows to their appropriate partition.
* **Postgres 11:** `UPDATE` statements can move rows between partitions.
* **Postgres 12:** Foreign keys can reference partitioned tables
* **Postgres 12:** Improved `INSERT` performance, `ALTER TABLE ATTACH PARTITION` no longer blocks queries.
* **Postgres 13:** Support for row-level `BEFORE` triggers on partitioned tables.
* **Postgres 13:** Logical replication on partitioned tables (previously, partitions would have to be replicated individually).
* **Postgres 14:** Partitions can be detached in a non-blocking way with `ALTER TABLE ... DETACH PARTITION ... CONCURRENTLY`.

It's largely been a story of getting partitioned tables up to feature parity with non-partitioned tables, and they're now very, very close.

## Day-to-day partition management (#partition-management)

We added our first partitioned table recently, and things have gotten so good that we skipped bringing in a partition-managing extension (e.g. `pgpartman`), opting instead to just add a small background job to our existing worker framework. It wakes every ten minutes, and run through these steps:

* Lists existing partitions.

* Calculates which new partitions should be created (if any) and creates them, bringing future partitions up three days in advance to give us some buffer in case something goes wrong. Creating new partitions is easy:

    ``` sql
    CREATE TABLE widget_20221005 PARTITION OF widget
        FOR VALUES FROM ('2022-10-05') TO ('2022-10-06');
    ```

* Partitions outside the retention window are detached:

    ``` sql
    ALTER TABLE widget
        DETACH partition widget_20221005;
    ```

* Detached partitions are kept around for three days in case an operator wants to inspect them, then dropped with the standard:

    ``` sql
    DROP TABLE widget_20221005;
    ```

`INSERT`s and `UPDATE`s all happen on the parent table, so partitioning is completely abstracted away from normal application logic.

The entirety of this scheme took me a few hours to write, and has been running for weeks, entirely problem-free. Recall that Postgres' DDL, including all the partitioning commands above is transactional, which let me to write very thorough tests easily, so the implementation is well-vetted. Each test case runs in a test transaction which cleanly rolls back altered altered state including new partitions after finishing, and is isolated from other tests cases running in parallel.

## Drawbacks, but few (#drawbacks)

I was amazed at how far I got without even noticing a single downside with partitioning aside from the modest activation overhead, but I'd be irresponsible not to note that there are still has a couple outstanding deficiencies.

The biggest one you're likely to run into is that indexes cannot be created or dropped concurrently. As any seasoned Postgres operator surely knows, use of `CONCURRENTLY` in a hot system is an absolute must to avoid service disruption, and not having access to it is a definite problem.

The workaround is to raise indexes concurrently on each individual partition, then raise it non-concurrently on the parent table so that future partitions get the index too. The parent will detect that each of its children has an index already that covers the requisite columns, and realize no additional work needs doing. There's [an outstanding patch](https://commitfest.postgresql.org/35/2815/) for creating an index on a partitioned table concurrently which would be a godsend if it landed. Fingers crossed for Postgres 16.

Partitioned tables also can't support `UNIQUE` indexes, which reduces the range of their potential uses. This one will be harder to fix because a solution would require an index that could span multiple tables, which would be new territory for Postgres.

(There are [a few other limitations too](https://www.postgresql.org/docs/current/ddl-partitioning.html#DDL-PARTITIONING-DECLARATIVE-LIMITATIONS), but no others are particularly notable.)

## 2022: Highly plausible (#highly-plausible)

Still, this leaves things in a very good place. I remember looking at a partitioning scheme that a colleague had implemented back around 2015, which was entirely manual because there was practically nothing built into Postgres to help at the time. It involved hundreds of lines of gnarly SQL functions and triggers just to get the basics working, and didn't cover schema or index management at all. It was well above my level of comfort for something I'd want to put into prod.

These days I wouldn't hesitate [1]. Partitioning massively improves Postgres' viability for storing high volume data, and it's getting better every year.

[1] Although remember, most tables probably don't need to be partitioned. Don't use this feature for no reason.
