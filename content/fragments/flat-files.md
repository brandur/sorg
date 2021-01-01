+++
hook = "Reduction in moving parts: going to more portable flat files, and as a side benefit, cutting CI build times by 30-40%."
published_at = 2021-01-01T18:09:29Z
title = "Development log: Moving to flat files"
+++

I have a few days off for holidays, and naturally, I've been spending it messing around with the home-grown blogging software that builds this site. I'd recommend using your time more wisely than I do, but it's amazing how fun it is to hack on little Go projects.

Somewhere around a decade ago, I added some "quantified self" pages like [`/twitter`](/twitter) and [`/reading`](/reading). They were driven by a separate project called ["Blackswan"](https://github.com/brandur/blackswan) (I was still on a Ruby-style obfuscating naming binge at the time) which ran a Heroku scheduled job and stored the output to a tiny hobby Postgres database. During its build step, this blogging engine connected directly to that database and constructed the pages according to what it found.

I've been wanting to move away from this system for a while now. Mainly because:

* Installing a Postgres service in GitHub Actions to run the test suite adds enough overhead that it has a noticeable impact on build times.
* "Blackswan" hasn't been maintained in years, and just _barely_ works these days.
* Fewer dependencies is better, and eliminating Postgres makes the software faster and easier to get running.
* I don't expect the free offerings on Heroku to last all that much longer, and it's better to migrate off before it turns into an emergency.

## qself (#qself)

I ended up extracting the parts of Blackswan that were still good to a new project called [qself](https://github.com/brandur/qself). Instead of upserting to Postgres, it writes the data to TOML flat files which are easy to move around and commit to projects. A second repository, [qself-brandur](https://github.com/brandur/qself-brandur), runs the qself executable in CI, and [commits the results](/fragments/self-updating-github-readme) as part of the build process.

When the blog's build process runs, it grabs the latest version of the flat files via cURL, which lets it build up-to-date Goodreads and Twitter pages:

``` sh
curl --compressed --output data/goodreads.toml https://raw.githubusercontent.com/brandur/qself-brandur/master/data/goodreads.toml
curl --compressed --output data/twitter.toml https://raw.githubusercontent.com/brandur/qself-brandur/master/data/twitter.toml
```

That let me drop all references to Postgres in my code and the build. Not only does this produce a more self-sufficient program that's easier to bootstrap, but dropping the Postgres service from GitHub Actions trimmed 30-40% off my build times (down to ~90 seconds for a `master` build, roughly half of which is syncing to S3).

Breaking Postgres off mirrors the motivation that led me to switch from a dynamic to static site in the first place. In case all my software stops working, I'm still left with byproducts that can easily be archived and which are almost universally interoperable (HTML, or TOML data files in this newer case). Compare that to say a dynamic Ruby project, which without constant maintenance gets a little less likely to run every year. It's also easy to reuse them throughout different projects without sharing around a connection string that might need to be rotated at some point.

Anyway, more like a step sideways than a step forward, but it was a fun little project, and leaves code in a state I'm a little happier with.

Happy new year!
