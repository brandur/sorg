+++
hook = "Life with Postgres only."
published_at = 2022-02-08T18:23:50Z
title = "Single dependency stacks"
+++

As I wrote about in [_Nanoglyph_ 030](/nanoglyphs/030-onionskin), both our major services at Crunchy have only one stateful dependency -- Postgres. No Mongo, Redis, ElasticSearch, or anything else [1].

It wasn't a stated objective, but we decided to see how long we could make it work. Having fewer dependencies has some major benefits:

* Fewer dependencies to fail and take down the service.
* Fewer systems to get proficient in operationally. (A common pitfall is to introduce something like ElasticSearch, only to realize a few months later that no one knows how to run it.)
* Fewer systems to upgrade.
* Fewer systems to monitor.
* Faster and easier development set up -- I could get the projects set up on a brand new machine in less than five minutes.

It does require us to do a few things in a more unusual way. For example, normally I'd push all rate limiting to Redis. Here, we rate limit in memory and assume roughly uniform distribution between server nodes. If we were to need search, we'd use [Postgres' full text search](https://www.postgresql.org/docs/current/textsearch.html) instead of ElasticSearch.

It hasn't required any major abuse in trying to fit square pegs into round holes. If it did, we'd likely drop the single dependency principle -- we knew that it was only going to take us so far. We'll see how far that is.

[1]  Okay fine, S3 too, but that's a different animal.
