+++
hook = "A privacy and GDPR friendly approach to getting basic site analytics, and one which is more accurate in an age when adblockers reign supreme."
location = "San Francisco"
published_at = 2021-02-15T23:45:23Z
title = "Minimally Invasive Analytics for Hackers: GoAccess and Athena/SQL"
+++

For years, like half the websites on the internet, I've been using Google Analytics to measure traffic to this site. I've never used it as anything other than a glorified hit counter, and as the product's become more complex over the years, I'm regularly confused by its expanding interface.

More recently, concerned with Google's insatiable appetite to track us to every of the internet, I've started looking into more privacy-friendly alternatives. I landed on [Goatcounter](https://www.goatcounter.com/), which purposely does not set cookies and makes no attempt to track users, and is GDPR-compliant without a notice. It's also got a great, simple interface that has just the information I actually care about.

But something I noticed after installing it is that although Goatcounter is an independent analytics product with good intentions (track as little as possible), that doesn't keep it safe from being included in uBlock Origin's block list. Indeed my adblocker was blocking analytics for my own website:

{{Figure "uBlock Origin blocking analytics sites." (ImgSrcAndAltAndClass "/photographs/articles/hacker-analytics/ublock-origin.png" "uBlock Origin blocking analytics sites." "overflowing")}}

This got me thinking: if a large segment of my visitors are using adblockers, are my analytics even accurate? If not, how far are they off? It turns out that especially for a technical site like this one, they are absolutely not accurate, and off by _a lot_. It turns out that if there's any demographic of person who has an adblocker installed -- it's you, dear reader.

I had a post from this website briefly spike to #1 on Hacker News last week, and decided to try and find out the difference between the traffic my analytics were showing, and the traffic I was actually receiving.

My estimate is that I got **~2.5x** more unique, non-robot visitors than reported by analytics (~38k versus ~13k), meaning that roughly **60% of users are using an adblocker**. Read to the end to see how I got these numbers.

If analytics products are being blocked at that level, there's a strong argument that it's not worth using them at all -- they're really only giving you a sample of traffic rather than an even semi-complete view. So what do you do instead? Well, how about analytic'ing like its 1999 by reading log files. This is much easier than it sounds because believe it or not, there's some amazing tooling to support it.

## GoAccess (#goaccess)

[GoAccess](https://goaccess.io/) is a very well-written log parsing and visualizing utility, featuring both a curses-based terminal interface and a web server that will produce HTML.

{{Figure "GoAccess' command line interface." (ImgSrcAndAltAndClass "/photographs/articles/hacker-analytics/goaccess.png" "GoAccess' command line interface." "overflowing")}}

It supports all common web logging formats including from Apache, Nginx, ELBs, and CloudFront. This site is [hosted on S3 and served via CloudFront], so I'm using the latter. Logging from CloudFront is easily configurable under the main settings panel for a distribution:

{{Figure "Configuring CloudFront logging." (ImgSrcAndAltAndClass "/photographs/articles/hacker-analytics/cloudfront-logging.png" "Configuring CloudFront logging." "overflowing")}}

Consider using a log prefix as well so that you can log from multiple sites to the same bucket, and save yourself from configuring the same thing over and over again.

Another nice augmentation is to configure the target S3 bucket with an expiration policy. This allows you to say, have logs pruned automatically after 30 days. so that you can further protector your visitors privacy, and so that they don't stick around forever eating into your storage costs.

{{Figure "Creating an S3 lifecycle rule for expiration." (ImgSrcAndAltAndClass "/photographs/articles/hacker-analytics/s3-lifecycle-rules.png" "Creating an S3 lifecycle rule for expiration." "overflowing")}}

(Create a new "lifecycle rule" under the `Management` section of a bucket. The settings are all straightforward.)

With logging now set up, your ready to sync your logs down and start using GoAccess.

### Some ergonomics (#ergonomics)

I have a [Git repository](https://github.com/brandur/logs) that acts as a little analytical test bed for logs. I don't commit any actual logs, but it contains a variety of scripts that provide easy shortcuts for frequent tasks.

Here's one that uses awscli to sync my logging bucket down locally:

``` sh
#!/bin/bash

aws s3 sync s3://<logs_bucket> logs-brandur/ --delete
```

So I can easily run:

``` sh
bin/sync
```

Here's another that starts GoAccess uses my standard logs location, with Gzipped logs streamed into it and filtered through a list of exclusions that I don't care about seeing:

``` sh
#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <site>"
    exit 1
fi

NUM_DAYS=30

files=(logs-brandur/$1/*)

# protects against degenerately large numbers of files in the directory
last_files=${files[@]: -3000}

gunzip -c $last_files | grep --line-buffered -v -E -f exclude_list.txt | goaccess - -p conf/goaccess.conf --keep-last $NUM_DAYS
```

Now, instead of that convoluted and impossible-to-remember invocation, I just run:

``` sh
bin/terminal
```

## Deeper introspection with SQL and Athena (#athena)

GoAccess is great, but it an be a little slow to sync logs down locally and boot it up. And while it gives us most of the information that we'd care about, we're still stuck on its rails. We can expand our use of analytics-via-logs using [AWS Athena](https://aws.amazon.com/athena/), which gives us the ability to analyze our logs with arbitrary SQL at relatively low cost.

Athena is built on [Presto](https://prestodb.io/), an SQL engine specializing in large, distributed data. Unlike a traditional data warehouse, Presto doesn't need an online component where data is stored centrally -- it's more than happy to spin itself up ad-hoc and read data as needed out of a collection of files stored on S3.

Start, by creating a new Athena database:

``` sql
CREATE DATABASE brandur_logs;
```

(By the way, don't try to use hyphens when naming things, or you will run into some really awful error messages.)

And then creating a new table within it that has the same structure as the Cloudfront logging format. Note that `LOCATION` statement at the end which specifies that the table's source is an S3 path.

``` sql
CREATE EXTERNAL TABLE IF NOT EXISTS logs_brandur.brandur_org (
  `date` DATE,
  time STRING,
  location STRING,
  bytes BIGINT,
  request_ip STRING,
  method STRING,
  host STRING,
  uri STRING,
  status INT,
  referrer STRING,
  user_agent STRING,
  query_string STRING,
  cookie STRING,
  result_type STRING,
  request_id STRING,
  host_header STRING,
  request_protocol STRING,
  request_bytes BIGINT,
  time_taken FLOAT,
  xforwarded_for STRING,
  ssl_protocol STRING,
  ssl_cipher STRING,
  response_result_type STRING,
  http_version STRING,
  fle_status STRING,
  fle_encrypted_fields INT,
  c_port INT,
  time_to_first_byte FLOAT,
  x_edge_detailed_result_type STRING,
  sc_content_type STRING,
  sc_content_len BIGINT,
  sc_range_start BIGINT,
  sc_range_end BIGINT
)
ROW FORMAT DELIMITED 
FIELDS TERMINATED BY '\t'
LOCATION 's3://logs-brandur/brandur.org/'
TBLPROPERTIES ( 'skip.header.line.count'='2' );
```

One downside is that the Athena interface is rough even by Amazon's low standards, but the fact that someone else will run a Presto cluster so that you don't have to, is a godsend.

{{Figure "Querying via Athena's UI." (ImgSrcAndAltAndClass "/photographs/articles/hacker-analytics/athena-query.png" "Querying via Athena's UI." "overflowing")}}

One of AWS' best features is that it has a complete API for every service, and this API is reflected into commands in awscli, which makes it very easy for us to access and use. I have less-than-zero desire to touch Athena's web UI, so I wrote another [little script](https://github.com/brandur/logs/blob/master/bin/query) that creates an Athena query, polls the API until it's finished, then shows the results in a simple tabulated form. The script takes an `.*sql` file as input, so I can write SQL with nice syntax highlighting and completion in Vim, and have them version controlled in Git -- two features not available if using the vanilla Athena product.

``` sh
$ bin/query queries/brandur.org/unique_last_month.sql
```

As an example, here's a query that maps over my Cloudfront logs to give me unique visitors per day:

``` sql
SELECT
    date_trunc('day', date) AS day,
    count(distinct(request_ip)) AS unique_visitors

FROM brandur_logs.brandur_org
WHERE status = 200
  AND date > now() - interval '30' day

  -- filter out static files
  AND uri NOT LIKE '%.%'

  -- filter known robots (list copied from Goaccess and truncated for brevity)
  AND user_agent NOT LIKE '%AdsBot-Google%'
  AND user_agent NOT LIKE '%Googlebot%'
  AND user_agent NOT LIKE '%bingbot%'

GROUP BY 1
ORDER BY 1;

```

For a "tiny" data set like mine (on the order of 100 MB to GBs), Athena replies in seconds:

``` sh
$ bin/query queries/brandur.org/unique_last_month.sql
query execution id: 65df1113-b206-4fc0-b1d2-8ac8017cbc35

 + ---------- + --------------- +
 | day        | unique_visitors |
 + ---------- + --------------- +
 | 2021-01-17 | 624             |
 | 2021-01-18 | 801             |
 | 2021-01-19 | 820             |
 | 2021-01-20 | 806             |
 | 2021-01-21 | 824             |
 | 2021-01-22 | 866             |
 | 2021-01-23 | 743             |
 | 2021-01-24 | 692             |
 | 2021-01-25 | 947             |
 | 2021-01-26 | 808             |
 | 2021-01-27 | 894             |
 | 2021-01-28 | 860             |
 | 2021-01-29 | 781             |
 | 2021-01-30 | 599             |
 | 2021-01-31 | 627             |
 | 2021-02-01 | 817             |
 | 2021-02-02 | 879             |
 | 2021-02-03 | 835             |
 | 2021-02-04 | 886             |
 | 2021-02-05 | 1232            |
 | 2021-02-06 | 540             |
 | 2021-02-07 | 530             |
 | 2021-02-08 | 19599           |
 | 2021-02-09 | 14626           |
 | 2021-02-10 | 1934            |
 | 2021-02-11 | 1341            |
 | 2021-02-12 | 1148            |
 | 2021-02-13 | 809             |
 | 2021-02-14 | 888             |
 | 2021-02-15 | 901             |
 + ---------- + --------------- +
```

Here's a different query that shows me my most popular articles this month:

``` sql
SELECT
    uri,
    count(distinct(request_ip)) AS unique_visitors

FROM brandur_logs.brandur_org
WHERE status = 200
  AND date > now() - interval '30' day

  -- filter out static files
  AND uri NOT LIKE '%.%'

  -- filter known robots (list copied from Goaccess and truncated for brevity)
  AND user_agent NOT LIKE '%AdsBot-Google%'
  AND user_agent NOT LIKE '%Googlebot%'
  AND user_agent NOT LIKE '%bingbot%'

GROUP BY 1
ORDER BY 2 DESC
LIMIT 20;
```

And again, it executes in seconds:

``` sh
$ bin/query queries/brandur.org/top_articles_last_month.sql
query execution id: 1830fea4-725d-4e73-ab53-0ffcff3a189f

 + ------------------------------------ + --------------- +
 | uri                                  | unique_visitors |
 + ------------------------------------ + --------------- +
 | /fragments/graceful-degradation-time | 32802           |
 | /                                    | 4854            |
 | /articles                            | 2090            |
 | /logfmt                              | 2016            |
 | /large-database-casualties           | 1726            |
 | /fragments                           | 1448            |
 | /postgres-connections                | 1393            |
 | /about                               | 1227            |
 | /photos                              | 942             |
 | /fragments/rss-abandon               | 821             |
 | /newsletter                          | 811             |
 | /idempotency-keys                    | 797             |
 | /go-worker-pool                      | 724             |
 | /twitter                             | 690             |
 | /fragments/homebrew-m1               | 645             |
 | /now                                 | 633             |
 | /fragments/test-kafka                | 611             |
 | /fragments/ffmpeg-h265               | 598             |
 | /elegant-apis                        | 575             |
 | /postgres-atomicity                  | 412             |
 + ------------------------------------ + --------------- +
```

Note that Athena is currently priced at $5 per TB of data scanned. That makes it quite cheap for a site like mine that generates on the order of 100 MB of logs per month, but it's worth thinking about if you're running something much larger. It also means that it's cheaper if you retain data for shorter periods of time, thereby running analytics over less of it (and making your site more privacy-friendly).

([Thanks to Mark](https://markmcgranaghan.com/cloudfront-analytics) for inspiring this section of the post.)

## How many adblockers? (#adblockers)

By comparing the results from my online analytic tools and those from these logging-based solutions, I can get a rough idea of how many of my visitors are using adblockers, and therefore invisible to analytics.

I'm using my HN spike from last week as a good slice of time to measure across. Note that this analysis isn't perfectly scientific and certainly has some error bars, but I've done my best to filter out robots, static files, and duplicate visits, so the magnitude should be roughly right.

{{Figure "GoatCounter's measurement of an HN traffic peak." (ImgSrcAndAltAndClass "/photographs/articles/hacker-analytics/goatcounter.png" "GoatCounter's measurement of an HN traffic peak." "overflowing")}}

Both Google Analytics and Goatcounter agreed that I got **~13k unique visitors** across the couple days where it spiked. GoAccess and my own custom Athena queries agreed that it was more like **~33k unique visitors**, giving me a rough proportion of **2.5x**.

So while analytics tools are still useful for measuring across a  sample of visitors, they're not giving you the whole story, and that in itself is a good reason that you might want to drop them, privacy concerns aside.

Personally, I think it's still fine to use the ones that are making an effort to be privacy-aware like Goatcounter, and they certainly yield benefits over analytics-by-logging like giving you JavaScript-only information like operating system and screen size.
