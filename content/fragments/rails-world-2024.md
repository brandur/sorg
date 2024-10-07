+++
hook = "A few reflections on this year's event. (It was great.)"
image_url = "https://www.dropbox.com/scl/fi/4cquusdmhqxg6zef7e0jp/L1000080.JPG?rlkey=77n8qrbsbzeea212rrv83kzu2&dl=1"
published_at = 2024-10-06T13:17:03-07:00
title = "Rails World 2024"
+++

I attended Rails World again this year, this time in Toronto. A quick recap while it's still fresh.

What a great event. Both this year and last the organizers went out of their way to pick some of the most incredible venues I've ever seen. Many places are adequate to the task of containing a conference for a few days, but few make your mouth go wide with a "wow" as you walk into the place.

This year's was held at Evergeen Brick Works, an old factory that lapsed into a state of disrepair for many years, and later converted to event venue. Its renovators decided to keep some aspects of the previous abandoned wreck. Its roof that'd fallen in wasn't replaced, leaving the evergreens that'd grown in the interim stretching through up into the sky (unclear what would've happened if it'd rained). Derelict machinery and the more tasteful graffiti had been left in place to add to the character. Meanwhile, ultra-modern acoustics and AV equipment made for excellent talks, and clashed nicely with the exposed brick.

<img src="/photographs{{DownloadedImage $.Ctx "/fragments/rails-world-2024/evergreen" "https://www.dropbox.com/scl/fi/d1de5k9n0mun8oq1tdujp/L1000128.JPG?rlkey=m2dz02vqevm4ekt04iykfyvnm&dl=1" 600}}" loading="lazy" class="rounded-md">

<img src="/photographs{{DownloadedImage $.Ctx "/fragments/rails-world-2024/graffiti" "https://www.dropbox.com/scl/fi/4vz36r3ign3r7e4m6g2hp/L1000131-cropped.jpg?rlkey=l8uxbn6lsq5555cy4524oeqet&dl=1" 600}}" loading="lazy" class="rounded-md">

Attention was paid to every detail. Quality drinks and delicious snacks were always on offer between sessions, and three food trucks operated all day outside (and good choices too: pizza served out of a decommissioned fire truck, beaver tails, and poutine, only Canada's best! [1]). One of my favorite details that was a holdover from the conference's first year is that all breakfast and lunch food is edible standing up, and served out of the same area that made up the convention floor. With few tables available, people mingle organically while eating, preventing a common conference lunch problem of groups self-siloing at tables where they stay immobile for 30+ minutes and meet few new people, if any. Organizers responded dynamically to fix problems as they arose. For example, lunch lines were too long on the first day, so by day two there were double the number of food stations. Pair programming sessions were available all day through Test Double.

This was all a nice change after attending RailsConf a few years back. There you couldn't even get coffee outside a tight 30 minute availability window in the morning. This was understandable because money was tight. Ruby Central was spending it on more important things, like paying out $500k cancellation penalties to send a political "fuck you" to the entire state of Texas, which happily took their money and proceeded to not notice at all. (It may not be a big surprise to hear that 2025 will be the last year of RailsConf.)

DHH is [pretty transparent on numbers](https://world.hey.com/dhh/wonderful-rails-world-vibes-7a6141d2), and was up front that Rails World operates at a loss that's backstopped by the large companies that form Rails Foundation:

> Rails Foundation, the founding core members listed above, as well as the contributing members [...], were willing to happily underwrite a loss of over $100,000 on the conference itself.

I love it. This is one of the best ways for companies getting good leverage out of Ruby/Rails to give back to the community. We're not contributing anywhere near what a colossus like Shopify is, but it felt great to have Crunchy sponsoring the event.

<img src="/photographs{{DownloadedImage $.Ctx "/fragments/rails-world-2024/rails-8" "https://www.dropbox.com/scl/fi/r259fjtqcyfyzmizluilc/L1000098.JPG?rlkey=8mmqlghs9xll0r6txs078bawy&dl=1" 600}}" loading="lazy" class="rounded-md">

## Tech highlights (#tech-highlights)

I spent most of the conference at our booth, so I mostly only got a chance to catch [the keynotes](https://www.youtube.com/watch?v=-cEn_83zRFw), but that was enough to catch the broad themes. A few notable highlights.

### Solid Cache (#solid-cache)

Like last year, David touched upon Solid Cache. This is such a great concept: caches traditionally always needed to be memory bound using a component like memcached or Redis because memory was fast and disks were slow. Now, memory is still fast, but with modern SSDs, disk is _also_ fast, and available in much larger denominations. 37 Signal's products like Hey put their cache in MySQL, where they run it on a 30 TB disk with 60 days retention, and which has a 96% cache hit rate. This especially improves cache hits for the long tail of older keys that would've been long since the evicted given a less spacious in-memory data set.

Solid Cache also dovetails well with the [single dependency stack](/fragments/single-dependency-stacks). Three years later we still run one and exactly one persistence component: Postgres. It's amazing just how plausible this is even for a mature stack, and it makes you realize that even the most fundamental belief systems of the programming world should be reevaluated every once in a while.

37 Signals stubbornly cargo cults Oracle products, but as Andrew covers, [Solid Cache can be made workable on Postgres too](https://andyatkinson.com/solid-cache-rails-postgresql). Although let me caveat that to say I've never done it, and suspect that there might be issues with long-lived deletion expiration queries at the scale of 30 TB of data since Postgres isn't particularly good at efficiently deleting rows (a big reason that recent partitioning improvements are so important).

### Server-phobia (#server-phobia)

For the last few months David's been on an anti-cloud mission. One of keynote slides highlights the size, capacity and cost of a Performance M dyno (1 core/2 threads w/ 2.5GB for $250/mo.), with the next showing a rough equivalent on Hetzner (48 cores/96 threads w/ 256GB for $220/mo.), the clear message being that the Hetzner box is 50-100x more capable, and also cheaper. A big new piece of Rails is [Kamal](https://kamal-deploy.org/), a system that's meant to make deployment to raw metal as simple as it is on Heroku. Kamal bundles the new [Kamal Proxy](https://github.com/basecamp/kamal-proxy), a reverse proxy that coordinates deploys, terminates TLS, and handles graceful restarts.

<img src="/photographs{{DownloadedImage $.Ctx "/fragments/rails-world-2024/performance-m" "https://www.dropbox.com/scl/fi/3w12dz0sw7eqrwhnu8dun/performance-m.jpg?rlkey=eh9r8m6uc4011f7t4hhnctxsh&dl=1" 600}}" loading="lazy" class="rounded-md">

<img src="/photographs{{DownloadedImage $.Ctx "/fragments/rails-world-2024/hetzner" "https://www.dropbox.com/scl/fi/wbm4jcjud1841m02ywral/hetzner.jpg?rlkey=92s8o57of5xmpuy2kb1anujqp&dl=1" 600}}" loading="lazy" class="rounded-md">

He's got a point with this one. For a long time servers represented a huge capital investment and distraction from building an actual product, and in that context AWS and its ancillaries are an attractive idea. But as anyone who's used a lot of AWS could tell you, it may be cheap in the beginning, but it's only a matter of time until that inverts, and AWS bills become a recurring nightmare.

That said, if I were trying to send this message I'd be careful to make it clear that this is a trade off. You're unquestionably going to save money on hardware, but you'll spend more time on management. Someone's also going to be the one carrying the pager for all these boxes, and presumably that's not the 37 Signals CEO or any of its executive team.
		
### Rails 8.1: Et tu search? (#rails-8-1)

Rails 8 was released that day, and he closed the keynote by touching on some expected features for its next major release, 8.1. Next in its sights is the beast that no sane person wants to run: ElasticSearch, with the promise of bringing a sophisticated search engine into Rails itself. Also up for inclusion is "House (MD)", which would make Markdown a more native piece of the Rails stack.

``` ruby
# search on any field
Post.search "announcement"

# by specific fields
Post.search title: "announcement", content: "solid search"
```

---

<img src="/photographs{{DownloadedImage $.Ctx "/fragments/rails-world-2024/conference-hall" "https://www.dropbox.com/scl/fi/f7k4l1f2n9gyjkltxzv5o/L1000198.JPG?rlkey=bffu4mm72i3uha2xqgg0s3r7p&dl=1" 600}}" loading="lazy" class="rounded-md">

---

## Twenty min (#20-min)

Rails World was bigger this year than last, but it's far from a huge conference, as shown by the competitive ticketing process, where tickets were gone 20 minutes after going on sale.

Hard-to-get tickets are bad, but a positive side effect is that everyone at Rails World _really wanted_ to be at Rails World. You don't get there by accident. The result is that every single person you spoke to had something interesting to say. In one case I'd randomly started talking to a couple Dutch guys staying at the same hotel I was, and 15 minutes later we were talking about the trade offs of Aurora versus vanilla Postgres. This will sound self-serving, but I met quite a few people that were already familiar with this website, and they'd ask _me_ about topics I'd written about recently like [Postgres 17 bulk B-tree lookups](https://www.crunchydata.com/blog/real-world-performance-gains-with-postgres-17-btree-bulk-scans) or [generating a couple secure bytes with `gen_random_uuid()`](/fragments/secure-bytes-without-pgcrypto).

I love it. The passion and expertise is the closest I've experienced at any event to what we used to get in the halcyon days of the early 2010s, before tech was so obviously the most important industry in the world, and became ludicrously financialized as every venture firm and Stanford graduate jumped to get a piece of it.

---

## Unpopular opinion: Toronto (#toronto)

Going on its second year now, there's a traditional announcement in the closing keynote of where the next Rails World will be held. In 2025, it'll be back in Amsterdam, and I admit to breathing a sigh of relief (assuming I can even get in).

The Evergreen Brickworks venue is gorgeous, Shopify's Toronto office is fabulous, and I had a good time visiting the city. But. Toronto's downtown is enormous, and it's the kind of place where every street, at every hour day or night, is characterized by the constant roar of total, all-encompassing, gridlock traffic. And like anywhere, when traffic is bad and tempers are heated, roads are never enough space for the pinnacle of human innovation, the automobile, and cars spill over onto every crosswalk and bike lane. With the bike lanes full, bike traffic moves onto the sidewalks, 90%+ of which is also motorized, with few riders even bothering to give lip service to those little foot rest doodads on the bottom of the bike that before the advent of the lithium battery and lightweight motor, used to be for peddling. Stop signs, red lights, and traffic priority all become the loosest of possible suggestions.

I'd be exploring an inner city suburb, with leafy canopy and the most gorgeous, stately houses that positively _ooze_ history in all directions. Amazing! Beautiful! Except, these otherwise quiet streets are filled to the brim with hundreds of bumper-to-bumper SUVs (no self-respecting Canadian drives anything smaller than an SUV, and a family of two or more should ideally upgrade to something a little more size appropriate, like an F-350) inching their way onward at a pace only marginally faster than a brisk walk. I'd cross a bridge over a deep, forested ravine. Look over the edge, expecting to see a peaceful, bubbling brook far below. What do I see instead? A highway of course, which Torontonians have seen fit to plough through each of the city's precious few parks.

After one of the evening parties I found myself talking to a guy who was professing his undying love for the city of Toronto. Me: what exactly do you like about it? Him: the _diversityyyyyy_ man. Me: ... okay, ... anything else?

Sorry, I can't help myself. But also, Amsterdam is the correct answer.

---

To recap, great event, great people. I hope to see many of you there next year.

[1] This is only partially facetious. I love beaver tails and poutine.