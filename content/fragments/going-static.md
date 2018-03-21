---
title: Going static
published_at: 2016-07-14T00:31:55Z
hook: UNWRITTEN. This should not appear on the front page.
---

As of yesterday, I completed a rewrite of this site and moved it over from a
[Ruby/Sinatra stack][org] to a [static site built by a custom Go
executable][sorg], and deployed based on the ideas that I wrote about in [the
Intrinsic Static Site](/aws-intrinsic-static).

Overall, it was a good opportunity to do general housecleaning, and to write
tests for modules that didn't previously have any, but the major impetus for
the move was for bigger reasons. I've touched on a few of them below.

## Static Is fast (#fast)

The site is now stored in S3 after it's built from Travis, and then served from
a CloudFront distribution at one of their [~40 edge locations][cloudfront] from
around the world. Every page is static HTML so there's no dynamic rendering or
data store access to slow anything down.

The old stack, which was dynamically rendered but buffered by rack-cache and
using CloudFlare as a CDN, was pretty fast already. But I like that the new
stack is all hosted by one company (AWS) rather than being straddled across
CloudFlare and Heroku, and that I now have fine-grain control over the behavior
of my CDN.

## Static is resilient (#resilient)

The previous site had a few dependencies in the form of a Postgres database
used to procure tweets, books, runs, etc., and depended on Flickr in some cases
to serve images. On the whole both were very dependable, and rack-cache helped
mitigate the damage caused by any service hiccups, but the former would
occasionally result in a database connection error which would 500 the site,
and the latter would occasionally break connection while an image was loading.

The new system still uses Postgres and Flickr as sources, but they're only
referenced during the site's build step. Any trouble will error the build and
cancel deployment, so no user will ever see a problem.

## Static is future-proof (#futureproof)

I've recently been concerned with just how many minor tweaks I had to make to
old the Ruby codebase just to keep it modern and alive. For example, it's good
practice to follow the currently released version of Ruby at all times lest you
be caught with a major disparity gap when a major backwards incompatible change
comes down the pipe in the future.

Ruby's worst maintenance problem is that every gem that needs even reasonable
levels of performance is written in C. These native extensions are built as
bundler installs gems, and are prone to breakage as libraries that they're
linked against on the system are updated. I can't even count how many hours
I've sunk into fixing eventmachine builds because something changed in OpenSSL.

From a longevity perspective, Go is Ruby's polar opposite. The language and its
core libraries are incredibly stable, and no release in recent memory has been
backwards incompatible. By convention dependencies are vendored so that they
don't rely on a centralized repository system, and are almost without exception
written in Go itself so that build problems are few and far between.

There's a reasonable chance that this site's new code will build on a
contemporary Go compiler from ten years in the future with no changes. You
couldn't say the same about a Ruby codebase even when thinking about only two
years down the road.

And even in a future where all compilers have failed, I'm still left with the
build artifact itself in S3 which is a simple collection of HTML, CSS, and
image files that will run anywhere, including a browser pointed at localhost.
It's trivially simple to move them around or archive them permanently.

## There and back again (#there-and-back-again)

Considering that I originally started publishing on the original static site
generators like Movable Type, then went dynamic after the static site craze of
the early 2010s and joining Heroku, it's mildly amusing to me that I'm going
back to my roots. That said, I'm happy to be back on a platform that lets me
make big changes more safely and eats less of my time in ongoing maintenance.

[cloudfront]: https://aws.amazon.com/cloudfront/details/
[org]: https://github.com/brandur/org
[sorg]: https://github.com/brandur/sorg
