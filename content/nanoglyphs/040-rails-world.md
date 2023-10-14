+++
image_alt = "The Crunchy booth at the expo hall"
image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/040-rails-world/expo-hall@2x.jpg"
published_at = 2023-10-14T12:30:18-07:00
title = "Rails World, 7.1, Amsterdam"
+++

Last week I traveled to Amsterdam for the first ever Rails World. It was a great event -- well-organized, great location, and interesting, passionate people in every direction. In 2023, you'd be forgiven for assuming that the world's moved on from Rails, but the framework's done an admirable job of keeping itself current, despite the quantum leaps happening in JavaScript and the browser.

Tickets were gone in 45 minutes the hour they went on sale. I tried to buy one the first _minute_ they went on sale to get early bird pricing, but all the early bird tickets were already spoken for. I was twenty seconds too late.

DHH's keynote opened the conference, and was bursting at the seams with food for thought and novel ideas, reminding me of the old world circa Rails 3 where it felt like the whole web was a new frontier that was being discovered. There should be a video of it up soon in case you want to watch it for yourself, but a few highlights that I wrote down:

* Some shots were jokingly taken at AirBnB and Twitter's propensity for extreme microservice-ification, which DHH (in jest) explained as a ZIRP [1] phenomenon. Luxury pontification that's affordable only when companies are awash in cheap money.

* With belts tightening, we'll be transitioning away from extreme developer specialization to one of the _renaissance developer_, one who ships features across every layer of the stack (frontend, backend, infrastructure, etc.).

* He was quite excited for [Bun](https://bun.sh/), a new JavaScript runtime (yes, the JS world is reinventing itself again) and one-stop toolchain for bundling, testing, and dependency management. It's faster than what's out there right now, and far better integrated.

* But although faster builds are better than slow builds, they're not as good as _no_ build, and DHH talked a lot about building apps like Hey that don't bundle their JavaScript or ship source maps, preferring to leave it in the same JS files it's developed in. [Propshaft](https://github.com/rails/propshaft) is an ultra-minimal bundler that sets up a load path and provides digest stamping for caching and cache expiration.

* Turbo is adding _turbo morph_, a new action that can act as a broad replacement for the existing `<body>` replacement default and which uses DOM diffing to replace the UI to the minimal possible extent. It's easy to integrate and provides interaction niceties like making sure scroll position isn't lost.

* The introduction of [Solid Cache](https://github.com/rails/solid_cache). Redis has traditionally been the cache of choice for Rails, but DHH notes that we're in a different world than we were ten years ago, and quite notably, disks are much faster than they used to be thanks to huge advances in SSDs. Solid Cache is backed with MySQL and in-memory cache look up is still used for hot keys, but the use of a full database allows the cache to overflow to disk, where it becomes slower to access but far larger at minimal cost. Rolling out Solid Cache at Basecamp they'd incurred a slight performance loss on cache lookups at P50, but a massive 50% gain at P95 as their available cache size became effectively limitless.

* He talked about the idea of **Solid Queue**, which would bring similar persistence ideas to Active Job. A fair bit of airtime was devoted to Sidekiq, which in his estimation was a great product, but which should have an open-source equivalent, something that might emerge with the help of projects like [good_job](https://github.com/bensheldon/good_job) and _Mission Control_, a GUI for Active Job.

* In a similar vein Heroku also got an entire slide devoted to it. Again, a "great product for its time", but one which should have an open-source alternative that works just as well. His proposal is [Kamal](https://kamal-deploy.org/), which seems like an interesting project, but in my estimation glosses over the hard parts of operating a service.

<img src="/photographs/nanoglyphs/040-rails-world/keynote-solid-cache@2x.jpg" alt="Keynote on the subject of Solid Cache" class="wide" loading="lazy">

## Rails 7.1 (#rails-7-1)

[Rails 7.1](https://rubyonrails.org/2023/10/5/Rails-7-1-0-has-been-released) released the same day, which now produces `Dockerfile`s for new projects by default, supports batch enqueueing with Active Job, and unlocks Bun as an option for a new app's JS engine.

It didn't make the headline features, but 7.1 also brings in support for CTEs [using the `.with` keyword on models](https://github.com/rails/rails/pull/37944). e.g.

``` ruby
Post.with(
  posts_with_comments: Post.where("comments_count > ?", 0),
  posts_with_tags: Post.where("tags_count > ?", 0)
)
```

With translates to:

``` ruby
WITH posts_with_comments AS (
    SELECT * FROM posts WHERE (comments_count > 0)
), posts_with_tags AS (
    SELECT * FROM posts WHERE (tags_count > 0)
)
SELECT * FROM posts;
```

I'll caveat by saying that the idea of expressing complicated SQL expressions as Ruby DSL seems a little crazy to me from a maintainability perspective, but it's a good that it's possible. Being able to build up large queries incrementally without resorting to concatenating strings together is occasionally an indispensable feature.

<a href="/photographs/nanoglyphs/040-rails-world/version-wall@2x.jpg">
    <img src="/photographs/nanoglyphs/040-rails-world/version-wall@2x.jpg" alt="Version wall" class="wide" loading="lazy">
</a>

## 20 years (#20-years)

A clever setup from Rails World in the opening hallway -- a Rails version wall. Every notably release gets a block with some of its important features and is slotted in chronologically by its release. Sharpies are provided for people to write in their names according to when they started using the framework, with the first block listing only one solitary name.

<img src="/photographs/nanoglyphs/040-rails-world/version-wall-2003@2x.jpg" alt="Version wall (2003)" class="wide" loading="lazy">

I started using Rails during the 2.x series when it was feature complete and gaining in popularity, but before some of the really elegant syntax that came down the line in 3.0. I couldn't remember which minor version I onboarded into. I wrote my name under 2.3.

<img src="/photographs/nanoglyphs/040-rails-world/version-wall-2009@2x.jpg" alt="Version wall @ 2.3 (2009)" class="wide" loading="lazy">

## A shot to the heart (#short-to-the-heart)

The burst of vitality injected by Rails World might've been exactly what the Ruby community needed. I attended [RailsConf this year](/nanoglyphs/036-queues#railsconf), and there was so little vitality there that I'd almost written off the whole ecosystem.

You won't find a keynote from DHH at RailsConf post-2021 because he was uninvited and made persona non grata for _daring_ to suggest that people should spend their time at work working, instead of bludgeoning each other to death over politics (this guy is a real monster). An editorial decision was made to minimize technical content in favor of talks on soft skills and unionization because under the framework of its dominant ideology, technical content at a technical conference is now considered (straining credulity) ... bad? To cut costs, luxuries like printed name badges and coffee weren't available, but this sort of frugality is the price you pay to engage in high-cost signaling exercises, like giving up a $500k deposit to move a planned event from Texas to California to telegraph contempt for the state of Texas and the 30 million people who live there.

But even if you're not a critic of the status quo, having a few plausible options around can't be anything but a good thing. Next year Rails World will be shifting continents to my home country, to be held in Toronto September 2024.

<img src="/photographs/nanoglyphs/040-rails-world/noord-1@2x.jpg" alt="Amsterdam Noord (train station)" class="wide" loading="lazy">

## Amsterdam Noord (#amsterdam-noord)

Excepting an accidental half-hour stopover a few months back, I haven't been to Amsterdam since 2011.

Once of the beautiful parts about the Netherlands is that the best parts of its reputation are actually true. Visit San Francisco expecting to ride up and down hills on cable cars, only to realize they're kept in operation exclusively for use by tourists and a local would be caught dead stepping foot on one. Visit Paris expected to bathe in the glory of the Mona Lisa, only to find a postage-stamp sized painting on a distant wall, barely visible behind the mob of would-be TikTok influencers with selfie sticks, apparently assigned to guard duty. Holland is known for its extensive bike culture, with infrastructure that's so good and so extensive that it's an overwhelmingly popular way to get around, and it's true.

I didn't stay in the central part of Amsterdam on this trip because the hotel prices were obscene and the core area is a zoo, more reminiscent of Disneyworld than a city where people live. Instead, I took a bit of a potshot and stayed up in Amsterdam Noord just north of the channel (called ["IJ"](https://en.wikipedia.org/wiki/IJ_(Amsterdam\))). Having traveled around central Amsterdam and Amsterdam Noord fairly extensively now, I've seen for myself that the bike paths aren't just an optical illusion in the middle part of the city to deceive visitors -- they really do go everywhere, snaking around parks, lakes, and canals, and even up into the countryside. It's glorious.

For prospective visitors to Amsterdam, in this humble traveler's opinion, staying in Amsterdam Noord is a life hack. It's cheaper, and only marginally further away because there's easy access via (free, frequent) water taxi, and the fast, clean M52 rail line that runs down to Amsterdam Centraal in 10 minutes flat. Similar to London, fare gates take Apple/Google Pay without trouble, so there's never any mussing with ticket machines.

<img src="/photographs/nanoglyphs/040-rails-world/noord-2@2x.jpg" alt="Amsterdam Noord" class="wide" loading="lazy">

On the way home, after spending two hours on the tarmac, my flight was cancelled. That was bad, but after some initial consternation, it was the most civilized flight disruption I've ever been a part of. It was a full 777 so rebooking everyone on alternate flights would've been impossible, but we were informed that a new flight had been chartered for the next day, and that we'd all been booked to stay at a Steigenberger hotel near the airport.

In most cities the area immediately around an airport is an exclusion zone of roads and concrete -- as inhospitable to life as the surface of Mars -- but in Amsterdam it's actually kind of nice. Craig had gotten stuck with me so we explored a few hotel bars, and the next morning I had a great run through a nearby park called Het Amsterdamse Bos -- a stone's throw across the river, and just as nice as anything you find near Amsterdam central.

Until next week.

<img src="/photographs/nanoglyphs/040-rails-world/coffee-cart@2x.jpg" alt="Coffee cart" class="wide" loading="lazy">

[1] ZIRP = Zero-interest Rate Policy.