+++
image_alt = "Stainless backyard"
# image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/047-stainless/stainless-yard@2x.jpg"
published_at = 2026-01-19T12:50:11-08:00
title = "Stainless"
hook = "On joining Stainless, six months at Snowflake, and Komodo dragons."
+++

Almost five years ago I wrote [New Horizons](/nanoglyphs/024-new-horizons) as I left Stripe and went to work on databases at Crunchy.

I can hardly believe it! Those years passed in the blink of an eye, eventually culminating with our acquisition by Snowflake last June.

Welcome to _Nanoglyph_. Five years is a long time, and well, it happened again.

---

Let me start by way of reminiscence. I know a guy (Alex) whom I worked with at Stripe. Another member of the Developer Platform team that owned API infrastructure, API docs, and SDKs.

Even amongst a company of relative high performers as found at Stripe, work output is far from uniform. There was a moment a few years in where Stripe's infamous API reference docs (which our team owned) got so big that it had a hard time loading. You'd open it in your browser and it would hard freeze your browser for 5-10 seconds as it tried to progress. Couldn't click on anything. Couldn't scroll. Depending on the last time you upgraded your hardware you might've briefly gotten a beach ball of death. CPUs churning away millions of cycles on the most inefficient JavaScript and Ruby ever written. Was it mining Bitcoin somewhere in that loading process? An early gambit farming out AI training? No one could really say for sure.

We all knew something had to be done about it. The question was, what. Under the hood this product was a terrifying beast of hydra-esque proportion. Tens (hundreds?) of thousands of lines of the most marginal code ever written. You guys write spaghetti code? Cute. You could call ours spaghetti, if the spaghetti was as fine as thread and the ball of it as wide as Earth. There were no easy incremental fixes here. Only a slog of the magnitude of picking every piece of garbage out of the Pacific Ocean piece by piece out of the side of a one-man dinghy.

One person stepped forward and volunteered for this suicide mission. We all gulped. We knew we'd never see this man again. No one who goes that deep into the Stripe monorepo comes back. Not in one piece.

But he did it anyway. It was Alex.

---

Alex still seems uncomfortably young to a Cretaceous-era fossil like me, but was _very_ young then. My main memory from the time is how hard he worked on this project. All day, he was at his desk locked in. When I went for a post-work run, showered, and went upstairs afterwards, Alex was there. When I came back to the office after a night out at the bar, Alex was there. When I stopped in at 5 AM in the middle of Christmas week, Alex was there. (This last one made up, but you get the point.)

Working is one factor. Delivery is another. Stripe subtly encouraged young engineers to work 60+ hour weeks (these guys were into 9-9-6 _way_ before it was cool), so it wasn't uncommon. But most of them spent the extra time oscillating side to side like a rocking horse, the dullness of the company's institutional rails requiring the energy of a collapsing star to push the unwieldy monorepo forward one inch.

Getting something shipped necessitated a perfect trifecta of extreme aptitude, unparalleled application of effort, and most importantly, _grit_. The new Stripe API reference docs shipped to great fanfare, the majority of it built by one person. I'd maintain to this day that this was either a project for a ten-person team over five years, or a six-month project for Alex.

<img src="/photographs/nanoglyphs/047-stainless/brooklyn-townhouses@2x.jpg" alt="Brooklyn townhouses in snow" class="wide" loading="lazy">

---

## The API company (#api-company)

During my arc at Crunchy, a parallel story has been playing out. A solo founder (Alex) bootstrapped what should have been [the greatest unborn unicorn of all time](https://narwallmask.com/), failed fast on it, wasn't discouraged, and went on to start a [new company, Stainless](https://www.stainless.com/).

One of the main differentiators of Stripe's API was its SDKs. Instead of sending random cURL calls with loose parameters and no type checking in Stripe's general direction, you'd have a (mostly) conventional SDK built and ready in the language of your choice. It was a huge advantage, but not cheap to maintain -- I'd imagine the SDKs team at Stripe is up to a few dozen engineers by now, at a cost of easily multiple millions per year.

Stainless' main product is something similar, but made generic so that it functions for anybody with an OpenAPI spec. Ship your OpenAPI to Stainless, SDKs ready for your users to plug straight into their software come out the other side.

If you've integrated against [OpenAI](https://platform.openai.com/docs/api-reference/introduction), you've used Stainless generated SDKs. Or CloudFlare. Or Anthropic. I still don't quite understand how he managed to close logos of this level of notoriety with a company that was only a few years old, but he did it.

Like any good founder, talk to Alex and you quickly learn he's not ready to stop there. The company's charter isn't SDKs, it's developer experience, and he's got entire new spheres of product development that he's intending to get into post-haste. If he has his way, SDKs are just the beginning.

I was pleasantly surprised to learn too that he's already managed to hire a supermajority of the heavy hitters we knew from previous jobs, both from Heroku and Stripe. How? Another question mark. Possibly due to Alex' persuasive powers which I can attest to first hand. In the space of a month I went from a "lmfao startup" to a "see you at the office in a few weeks".

So yes, that's the news. Last December I joined Stainless along with some [old](https://github.com/bgentry) [buddies](https://github.com/pedro) in a product engineering capacity. Alex pitched me early on the product he wanted us to build. It seemed like a good idea, and for about a day afterward I kept thinking: yes, it _is_ a good idea. In fact, it's more than just a good idea. How has no one built that yet?

We've been pushing hard on it since returning to work in early January. I can't wait to write more about it.

<img src="/photographs/nanoglyphs/047-stainless/christmas-trees@2x.jpg" alt="Christmas trees in Brooklyn" class="wide" loading="lazy">

---

## Under the dome (#dome)

All in all, I ended up doing a little under six months at Snowflake. I'm still turning the acquisition over in my mind, and I wonder how I'll think about it ten years from now when it's more soundly in the rearview mirror.

So far, I've gone back and forth in a flip-flop-FLIP pattern of thinking about it:

1. Acquisition? Liquid stock, ESPP, Jones Soda on tap in the lunch room? Great!

2. Wait, most of us aren't going to make much of anything off of this and we're about to be fed into the meat grinder of a mega-corp. 4x/year perf cycles. Bad!

3. With a little more time and lengthy contemplation: All good things must come to an end and chapters in life rarely close this cleanly. Good?

At Crunchy we'd built a niche environment where we'd set a high technical bar, built a team of people we liked, and had a low-touch culture of shipping that kept things moving while maintaining good work/life balances all around. It was comfortable. Maybe a little _too_ comfortable. The bull case is that I was living the dream -- paid well, remote, and doing intellectually stimulating work. The bear case is that I was in stasis, watching from the corner window as the world accelerated and life passed me by. If all goes well, I have a few good working years left in me, and an argument could be made that they should be spent on the frontier trying to strike gold rather than taking it easy from the sidelines. With the acquisition, this world was ending whether I liked it or not. The question was no longer whether I should take a next step, but rather what's the next step to be?

We didn't waste a lot of time in the six months I was there, and got [Snowflake Postgres to public preview](https://www.snowflake.com/en/engineering-blog/postgres-public-preview/). I wouldn't have minded staying a little longer to see it to full GA, but time is the ultimate zero sum equation.

My brief glimpse back into the black box of a large tech company was useful in that it reinforced some of my longheld beliefs:

* Years ago I wrote about [Oracle's edit-compile-run loop](/nanoglyphs/029-path-of-madness), a multi-day affair that from the sounds of it, is positively hellish to work with. This is *not just an Oracle problem*. I can't reiterate this hard or often enough: **the CI loop is king**. It should be lavished with attention like a gentle deity and kept lovingly in pristine shape. Those who fail to do so should expect fire and brimstone.

* Monorepos? Yep, still bad. An observation I made along the way is that every team that _could_ opt out of using the monorepo _did_ opt out of the monorepo. While monorepos get all the lip service in the world, there's also a chronic attitude amongst senior engineering of monorepo for thee, but not for me. Everyone likes them in theory. Everyone hates them in practice.

* When given unlimited leeway and latitude for self-aggrandizement, the shenanigans that security teams will get up to knows no bounds. If a large organization has a 60% productivity tax in place through crappy tooling, there's another 15% security tax on top of that as you have security IC5s pummeling concessions out of dev teams all day long, then turning around and doing things that a learned undergrad would know is a bad idea, like hand-rolling their own half-baked crypto projects.

* Private Go Modules (i.e. any use of Go in a non-public context) are a really subpar experience. This isn't a big company thing, it's a Go thing.

* Jira. Confluence. Workday. How do these products still exist?? Their presence in an organization makes said organization _less_ productive, not more, and they ... cost money?! What the hell. There is no bull case.

* On a positive note, I can only admire how well Gmail and Slack scale to even huge organizations while staying fast. An ambition in life is to build a product that works so well that you don't even notice as it silently scales up to thousands of heavy users.

---

<img src="/photographs/nanoglyphs/047-stainless/rinca-boat@2x.jpg" alt="Boat to Rinca" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/047-stainless/rinca-boat-2@2x.jpg" alt="Boat to Rinca 2" class="wide" loading="lazy">

## The dragons of Rinca Island (#rinca)

I dropped the ball on my Indonesia vacation log, so here's a short addendum that I'd written up while sitting on a dive boat last year.

Heading out of the Komodo region, naturally we stopped to go see the dragons on the way out. Komodo is of course the best known island in Komodo National Park, but Rinca is another island in the same park that hosts almost as many Komodo dragons and is less well traveled.

It's not that easy to get to. These aren't like parks in the west with giant parking lots and huge highways leading into them that support tens of thousands of daily visitors. It's at least an hour's trip by boat from anywhere. The facility is fairly sizable once you're there, featuring some nicely built boardwalks that keep you above the flood/dragon plain, but feels practically abandoned. We only saw one other tour group all day.

The small crowds allow for a more private experience. After paying our way in and following the boardwalk some ways, we arrived at a mixed group of vendors and guides. One man peeled himself off the rest of the group, pulled down the conventional anti-dragon armament from the weapons rack (a forked stick), introduced himself, and took us through a gate. I was expecting to have to walk around for a while to find a dragon, but they've strategically built a watering hole next to the main building, so immediately there were three dragons right in front of us.

The best part of the Rinca Island experience is that it's as raw as you could ask for. Few other tour groups. No fences between you and the dragons. Just one Indonesian and his stick.

<!-- <img src="/photographs/nanoglyphs/047-stainless/komodo-walking@2x.jpg" alt="Komodo dragon walking" class="wide" loading="lazy"> -->

<img src="/photographs/nanoglyphs/047-stainless/komodo-walking-crop@2x.jpg" alt="Komodo dragon walking crop" class="wide" loading="lazy">

It's not quite as dangerous as it sounds though. We arrived around 10 AM, and by this time the dragons aren't very active. We had a few walk short distances to the water, but once there, they'd fall into a lazy-man pose with arms tucked back behind them that I instantly recognized as my preferred couch position for gaming on PS5.

<img src="/photographs/nanoglyphs/047-stainless/komodo-lazy@2x.jpg" alt="Komodo dragon lazy" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/047-stainless/komodo-lazy-2@2x.jpg" alt="Komodo dragon lazy 2" class="wide" loading="lazy">

It's hard to blame them because by this time it's already *hot*, at 30C/86F and little breeze. Later in the trip we'd hike up a hill to an overlook and during the 15 minute stint I sweated out about half my bodyweight. With its island chain spanning an area wider than the continental United States, Indonesia's a vast country with a lot of varied ecologies. Ambon, where we'd go after Rinca, is a jungle. Komodo where the dragons live is hot, arid territory made up mainly of long grasses and other dry weather plants. It reminds me of California.

The dragons are capable of short, powerful bursts of activity that have famously taken down animals as large as water buffalo, but they're rare enough that the water buffalo have little fear of them. It was a common sight to see dragons, water buffalo, and macaques cohabiting a single watering hole just feet from each other. With that said, as we hiked further around the island we'd come across an occasional "tombstone" marking where a dragon-on-water-buffalo encounter had not ended well for the latter.

<img src="/photographs/nanoglyphs/047-stainless/water-buffalo-skull@2x.jpg" alt="Water buffalo skull" class="wide" loading="lazy">

I asked our guide what it was like to live on the island (they work a 20 days on, 10 days off model not dissimilar to a rig worker back home) and he told me that he's friends with enough other people on the island that they find ways to keep it fun. The dragons are most active around sunrise and sunset, and a few of them have taken to livestreaming and making TikTok videos around that time. I encouraged him on the hustle -- in a country where most of the population spends most of their time staring vapidly into the algorithmic abyss of TikTok, _publishing_ a TikTok video is a major creative act.

In the interest of full disclosure, something I learned from my real life visit is that Komodo dragons are smaller than I expected. At 79 to 91 kg (174 to 201 lbs.), an average adult male weighs about the same as I do. Huge for a lizard, but smaller than one might think by searching "komodo dragon" on Google and looking at human/dragon images, where the dragon invariably gets a close up with a wide angle lens while a human sits in the optically distorted background, looking tiny by comparison. As a species we've developed a challenging relationship with the truth in all things.

A very cool experience. I'm not sure I'd travel to the area *just* for the Komodo dragons as the whole thing was over in a few hours, but it's a perfect pairing for a dive trip.

<img src="/photographs/nanoglyphs/047-stainless/komodo-sticks@2x.jpg" alt="Komodo sticks" class="wide_portrait" loading="lazy">
