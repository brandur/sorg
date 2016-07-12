---
hook: Data warehouses aren't just for the enterprise! Let's examine a few of their
  basic characteristics, and build some examples with Go/Ruby and Postgres.
location: San Francisco
published_at: 2014-09-05T16:25:04Z
title: The Humble Data Warehouse
---

Data warehouse. The term immediately brings heavy enterprise into mind, complete with business intelligence, heavy XML-based ETL instructions, and "N-dimensional" data processing technology like the [OLAP hypercube](http://en.wikipedia.org/wiki/OLAP_cube).

A few months back, a colleague of mine [built an app](https://github.com/mfine/prism) that pulls commit data from within the Heroku organization down from the GitHub API. Over the course of a few days (pesky rate limiting), it backfilled the entire commit history of every project that the company ever created into a Postgres database, then left a scheduled process running to periodically update itself with new data.

The project had a specific goal. While web APIs are powerful tools, their RESTful paradigm often doesn't organize data in quite the way that you'd like to access it, preferring instead to present data as resources with highly consistent interfaces.

    GET /repos/:owner/:repo/commits/:sha

GitHub's API couldn't be used directly for any kind of advanced querying or analysis, but it could be leveraged to build a local reflection of that data, and that was exactly what the app was designed to do:

```
=> select org, repo, sha, date
   from commits
   where email = 'brandur@mutelight.org' or email = 'brandur@heroku.com'
   order by date desc
   limit 10;
  org   | repo |                   sha                    |          date
--------+------+------------------------------------------+------------------------
 heroku | api  | fad1459005ad049fdbb2c854731f14fafb3a77d4 | 2014-02-12 16:28:08+00
 heroku | api  | 7c479b3b8d7bd983f8eba2a8e8e0e3cc2678e23a | 2014-02-12 16:27:56+00
 heroku | api  | 73c7cbef78f68699b3c8847a016b075725d34d3b | 2014-02-12 16:27:29+00
 heroku | api  | 409bcd4251acfab459856951e3e4e4b76c5c91bb | 2014-02-12 16:25:39+00
 heroku | api  | c319c0f405afc6d06b8c40af54875511fa5a0961 | 2014-02-12 05:04:54+00
 heroku | api  | 33b82dae3cce7e1304f391e5420c8560b0acf600 | 2014-02-12 05:02:28+00
 heroku | api  | 81c2a6f5c3c5338cf4f27e14fcb738288953ed1f | 2014-02-12 04:51:48+00
 heroku | api  | 569619257571af50e7bc291afd1bc73818bfa030 | 2014-02-12 04:44:28+00
 heroku | api  | fb1e1e47048353c1bbddff1164ca87bde8d876d9 | 2014-02-12 00:49:39+00
 heroku | api  | 4cae708c0d3de1842c9bd9fd3e0ea8481f2fd946 | 2014-02-12 00:46:45+00
(10 rows)
```

With the data now fully housed in an RDMS, the powerful querying features of SQL are available to filter, map, transform, join, compare, and aggregate this data in any way imaginable. Better yet, although the volume of commit data on GitHub's servers is undoubtedly very large and would take non-trival time and system resources to process, we've succeeded in boiling it down to just the subset that we're interested in --- to the extent that nearly any kind of number crunching takes negligible time even on a tiny database.

``` sql
--
-- most commits in the last six months
--
select email, count(*) from commits
where date > now() - '6 months'::interval
group by email order by count desc limit 10;

--
-- most frequent weekend committers
--
select email, count(*) from commits
where extract(dow FROM date) in (0,6)
  and date > now() - '6 months'::interval
group by email order by count desc limit 5;

--
-- longest commit messages
--
select email, avg(char_length(msg)) from commits
where date > now() - '6 months'::interval
group by email order by avg desc limit 5;

-- and much more!
```

This is an example of a data warehouse (DWH) on a scale small enough to be agile, and by extension free of the negative connotations of heavy software and big enterprise, but still very useful for analysis and reporting. In the modern world, Postgres databases are cheap and primitive software building blocks needed to extract data from foreign sources (i.e. API SDKs, HTTP clients, RSS readers, etc.), are readily available in the form of gems or NPM packages, allowing simple DWHs like this one to be built from scratch with amazing rapidity. No XML or 200k LOC frameworks required --- only your language and libraries of choice and your favorite database.

## Tweets (#tweets)

As another example, I've been using a similar technique for years to [archive my tweets](https://github.com/brandur/blackswan). Compare this query to slowly manually paging back through your list of tweets looking for that link you posted six months ago:

```
=> select occurred_at, substr(content, 1, 50)
   from events where type = 'twitter'
     and content ilike '%iceland%'
     and metadata -> 'reply' = 'false'
   order by occurred_at desc
   limit 10;
     occurred_at     |                       substr
---------------------+----------------------------------------------------
 2013-07-22 04:24:57 | Half of Iceland now wants the old centre-right par
 2012-11-04 23:18:24 | The British, with a help of a Canadian (!) task fo
 2012-10-05 12:22:18 | What’s happening in Iceland’s metal scene? http://
 2011-12-28 20:55:46 | Have Icelandic lineage/ancestors? Then check out:
 2011-10-02 00:34:33 | Bar tending at the annual Icelandic fall feast. Dr
 2011-07-11 14:48:56 | Awesome. My brother just pointed out that there's
 2011-01-03 02:41:43 | Beautiful "Icelandic Dragon Sword" calligraphy cou
 2011-01-02 18:34:41 | "Icelandic Dragon Sword", since my actual name can
 2010-06-11 02:45:49 | #CCP is inspired by #Iceland: http://vimeo.com/122
 2010-06-07 22:49:13 | When your country's primary industry sinks, do thi
(10 rows)
```

Just like it's more expensive enterprise cousins, this warehouse has [its own ETL process](https://github.com/brandur/blackswan/blob/master/lib/black_swan/spiders/twitter.rb) for pulling down these tweets from Twitter's API and storing them. It's written in Ruby and leverages community gems to stay concise and DRY.

## A File Warehouse (#file-warehouse)

As a final practical example, let's build a small Postgres data warehouse containing the contents of our home directories. I find myself consistently running into the problem where my disk is near full, but my operating system does a poor job of helping me to identify the best candidates for removal.

First create a database (you'll need Postgres and Ruby installed to follow along):

```
$ createdb home-warehouse
```

Now we'll create a table to hold all our files:

```
$ psql home-warehouse -c \
  'create table files (size bigint, name text, dir boolean)'
```

And finally, let's run our simple ETL that uses `du` to dump file information into the table that we've created (note the multiplication by 512 to convert blocks into bytes):

```
$ du -a . | \
  ruby -n -a -e 'puts "#{$F[0].to_i * 512}\t#{$F[1]}\t#{File.directory?($F[1])}"' | \
  psql home-warehouse -c '\COPY files FROM STDIN'
```

I've indexed my `~/Downloads` directory, and it's seriously bloated in a bad way. As you can see, I've got a serious amount of garbage in there:

```
=> select count(*) from files where dir = false;
 count
-------
 14819
(1 row)

=> select pg_size_pretty(sum(size)) from files where dir = false;
 pg_size_pretty
----------------
 18 GB
(1 row)
```

I don't want to just remove everything though; once in a while this oversized directory comes in handy because it allows me to dig up some file that I've had in the ancient past and which I want to look at again. Instead, let's just find the top candidates for deletion:

```
=> select name, pg_size_pretty(size) from files where dir = false order by size desc limit 30;

                             name                             | pg_size_pretty
--------------------------------------------------------------+----------------
 ./dchha39_Death_Throes_of_the_Republic_VI.mp3                | 300 MB
 ./GCC-10.7-v2.pkg                                            | 273 MB
 ./eclipse-standard-kepler-R-macosx-cocoa-x86_64.tar.gz       | 198 MB
 ./jdk-7u45-macosx-x64.dmg                                    | 184 MB
 ./andean.zip                                                 | 155 MB
 ./command_line_tools_for_xcode_june_2012.dmg                 | 147 MB
 ./dads_gift/musical_evenings_with_the_captain_1996.zip       | 126 MB
 ./ideaIC-12.1.4.dmg                                          | 117 MB
 ./complete/nzbget.log                                        | 116 MB
 ./dads_gift/musical_evenings_with_the_captain_ii_1997.zip    | 108 MB
```

Those look okay to delete by me, so let's axe them:

```
$ psql home-warehouse -c \
  '\COPY (select name from files where dir = false order by size desc limit 30) TO STDOUT' | \
  xargs rm
```

While the example above may be slightly contrived -- it doesn't compute anything that couldn't be determined using simple UNIX tooling -- its purpose is mainly to demonstrate that data can be queried in arbitrary ways using the full power of the SQL:2011 language standard after it's loaded into an RDMS.

The example above also demonstrates two other important concept (albeit in a trivial way):

1. The blocks to byte conversion in the Ruby ETL script shows how data can be transformed as it enters the warehouse to make it easier for for us to work with it.
2. Notice how we've also just sliced off a tiny piece of the filesystem (my `~/Downloads` directory). A massive dataset that would otherwise by expensive to query can be boiled down to just the pieces that we care about for fast analysis.

Truly a data warehouse for everyone! I'd highly recommend downloading [Postgres](http://www.postgresql.org/download/) and trying this out for yourself today.
