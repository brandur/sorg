---
title: Perfect Replica Reads In Postgres
published_at: 2017-11-11T23:54:10Z
hook: TODO
---

Stale reads.

## The Postgres WAL (#postgres-wal)

## Tracking replica locations (#replica-locations)

LSN = Log sequence number

Earlier location when querying on replica:

```
mydb=# select pg_last_wal_replay_lsn();
 pg_last_wal_replay_lsn
------------------------
 0/15E88D0
(1 row)
```

Later location:

```
mydb=# select pg_last_wal_replay_lsn();
 pg_last_wal_replay_lsn
------------------------
 0/160A580
(1 row)
```

On primary:

```
mydb=# select pg_last_wal_replay_lsn();
 pg_last_wal_replay_lsn
------------------------

(1 row)
```

Get current location on primary:

```
mydb=# select pg_current_wal_lsn();
 pg_current_wal_lsn
--------------------
 0/160A580
(1 row)
```

Get how far ahead or behind one location is compared to another:

```
mydb=# select pg_wal_lsn_diff('0/160A580', '0/160A580');
 pg_wal_lsn_diff
-----------------
               0
(1 row)

mydb=# select pg_wal_lsn_diff('0/160A580', '0/15E88D0');
 pg_wal_lsn_diff
-----------------
          138416
(1 row)

mydb=# select pg_wal_lsn_diff('0/15E88D0', '0/160A580');
 pg_wal_lsn_diff
-----------------
         -138416
(1 row)

```

https://www.postgresql.org/docs/current/static/wal-intro.html
https://www.postgresql.org/docs/current/static/functions-admin.html
http://sequel.jeremyevans.net/rdoc/files/doc/sharding_rdoc.html
