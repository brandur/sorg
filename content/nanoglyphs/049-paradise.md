+++
image_alt = "Marker just off the beach in Raja Ampat"
# image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/049-paradise/marker@2x.jpg"
published_at = 2026-02-22T13:22:49-08:00
title = "The Death Knell, Paradise"
hook = "One more quick thought on LLMs, the end of Heroku (?), things we should've done, and a glimpse of paradise in Raja Ampat."
+++

I'd intended for last week's [edition on LLMs](/nanoglyphs/048-llms) to be largely a one off. It mostly will be, but I've had a few amusing conversations on the subject since.

An insight from an old colleague:

> **Mark’s Law: any proponent of AI labor replacement is replaceable by AI.**

Words to live by? One might even consider it a moral imperative to help AI labor replacement proponents out of their backbreaking labor sooner rather than later.

That's a joke, but it seeded the thought in me that even if we assume that AI labor replacement is coming, it's not obvious _which_ labor is going to be replaced. I've seen exec/middle management on LinkedIn imply in excited tones how they can thin out their reports, but if an executive/middle manager's primary responsibilities are communication, coordination, and prioritization, are we so sure that organizationally-minded developers with time freed up from LLM use can't shoulder these tasks? (Amazon [seems to think so](https://finance.yahoo.com/news/amazon-ceo-says-cutting-middle-170605849.html).) Product people rave jubilantly on how they can now build products directly, but are we so sure that developers with a taste for good product don't expand their purview into product work? SaaS founders play with decimating their staff in favour of agentic models doing the same job, but if purchasers can build near simulacrums in days, do they need what that founder is selling?

I'm not making any claims, merely saying that the future isn't obvious. Certainly not so cut and dry as LinkedIn's legions of AI influencers would have you believe.

But enough about that. Thanks for reading Nanoglyph. This week: Fargate, the end of Heroku (?), and some photos from the island of Yeben in Raja Ampat, a prime candidate for paradise on Earth. As always, this edition [written by human](/human).

<img src="/photographs/nanoglyphs/049-paradise/beach@2x.jpg" alt="Beach in Raja Ampat" class="wide" loading="lazy">

## The partial product (#partial-product)

Stainless: all of us being very familiar with the AWS stack, we've been deploying everything so far to Amazon.

AWS is the most dominant cloud there is, but use it long enough and it's hard not to develop some misgivings about it. The sheer complexity of its product surface is the 8th wonder of the world. Billions [1] of human hours sunk into building a catalog so expansive that you need to use the console's search feature to find the service you're looking for.

A few weeks ago I spent an hour figuring out how to give one part of my account (Athena) access to another part of my account (S3 Tables). I had five separate docs pages pulled up. Not one accurately explained the process, even from a 30,000 foot level. Not even Claude could figure it out! I eventually got it by using docs to triangulate to within a few football fields of the right place and closed the last mile with trial and error. These sorts of sharp edges are a semi-regular occurrence.

We built our program to a container image and deployed it to ECS Fargate. Of course to get the image over there we need an ECR (container registry). Running containers don't have a static hostname, so you need a $20/month load balancer sitting in front of your single container app. To make all of this reproducible, you'll want to write a 200 LOC CloudFormation YAML description. Deployment is user-initiated, so you'll need a custom deploy script to automate it. And remember, Fargate is the "hands free" version of ECS that doesn't need you to run your own EC2 containers. This is the easy version of things!

The good news is that LLMs have made all of this much easier. A colleague had a deployment suite generated within a few hours, and unlike humans, LLMs tend to get CloudFormation files and shell scripts right, so they worked without having to iterate on them 50 times fixing dumb mistakes.

Still, we now have a bread basket of shell scripts and YAML files that no human knows that well, and yet are ours to maintain. The bill for a couple non-production prototypes is climbing towards $100/month. One time I tried to get the public URL for the service I'd just deployed, eventually finding it 5 minutes later after following about 18 AWS Console links. Maybe I have unrealistic expectations, but I found myself thinking: so we pay Amazon to host our software. It's a plug-and-play product that's not plug-and-play, nothing is integrated well, the console is confusing and not full-featured, the docs are awful, and it's pretty expensive. It's 2026. This is the best we can do?

<img src="/photographs/nanoglyphs/049-paradise/marker-dark@2x.jpg" alt="Marker just off the beach in Raja Ampat (in the evening)" class="wide" loading="lazy">

## Layoffs & LEGO (#layoffs-lego)

A few weeks ago we got word of layoffs at Heroku. The entire company's turned over since I was there (so take this with a grain of salt since I don't have any first- or even second-hand information anymore), but supposedly 30-40% of the staff was let go. The corporate jargon in their public announcement is hard to parse, but it sounds like the intention is to functionally sunset the product.

In a way, this is a shame. Compared to the process of deploying an app to Fargate I described above, the Heroku experience is still better despite not noticeably changing in 15 years.

The expansive layoffs were a bit of a surprise because I'd been told numerous times over the years that Heroku as a division inside of Salesforce was doing well. At Rails World last year, we found our booth right next to a large, expensive looking one from Heroku. While we were giving out stickers and coloring books, the swag they'd brought this year were authentic, miniature Heroku-branded LEGO sets, allegedly $50/set to produce. I didn't think much of it at the time, but things at Heroku seemed to be going pretty well.

But in another way, it wasn't a big surprise. The Heroku experience has been largely static for all of recent history. Its best known innovators and most ambitious people left a decade ago, and though I'm sure they had plenty of good people left, I'd hazard to guess that a lot of the staff left were those in remote roles with grandfathered leveling from 2021's remote-mania and LCOL-geo comps that'd be near impossible to pull in 2026's market. And when CRM is down 40+% YoY (for years I'd gotten so used to this stock only ever going up and to the right that it was shocking to see it in such ill health), something's got to give.

---

Soon after the announcement a flurry of LinkedIn posts was hatched, reading something to the effect of "Salesforce should have ...", "Salesforce could have ...", etc. I'm going to break with my ex-colleagues here and say that I don't think Salesforce is responsible for anything. The failure of Heroku is on Heroku.

The Salesforce acquisition was a dream scenario. We kept our own (gorgeous) independent office for another five years. Free lunch continued. New corporate Amexes were issued. Expenses were underwritten. Salesforce bankrolled two of the most lavish conferences ever held. Heroku retained perfect control over our staffing, technical direction, and infrastructure. During those early years, if we'd told Salesforce we were building a rocket ship, they would've funded it.

All the company's brightest minds were together in one place, out of startup mode, and at long last, with a bit of breathing room. It would've been the time to punch the NOS and see if we could hit escape velocity.

But we didn't. We could hardly know it at the time, but Heroku's innovative stretch was already at an end. The product would get a little better, it'd continue to run well, and would hold its own against myriads of new competition over the years, but it'd never make another big splash in the way that Cedar did.

That was 2012, and though it would've been the best opportunity to go big, it wouldn't be the last. It lost momentum, but its resources never went away, and that powerhouse brand still carries through to today, where I'd hazard a guess that technical people would still find the name "Heroku" more recognizable than "Fargate", "Fly", or "Render". Heroku could've still recovered in 2015, and 2018, and 2021, and 2024. It just didn't.

<img src="/photographs/nanoglyphs/049-paradise/pier@2x.jpg" alt="Pier out the boathouse in Raja Ampat" class="wide" loading="lazy">

## The path not taken (#path-not-taken)

I've thought a lot about what we could've done differently over the years. You could ask ten people from that time and get ten different answers, but here's my version of the list:

* **SSH:** We'd gotten so many requests over the years to have a way of connecting to a running build or production dyno.

* **Multi-cloud:** Support Azure, GCP, and hopefully Salesforce's own cloud too. There was no reason not to do this except that a lot of AWS-specific code had already been written.

* **VPCs:** Isolation is important for most large customers, and its absence often a dealbreaker.

* Kill the free tier in favor of **$5 dynos**. The free tier was so obviously unsustainable, and $5 is such an attractive price point.

    Importantly though, these should not be _loss leader_ $5 dynos. Some of the dumb stuff we used to do, like giving away storage for slugs and free bandwidth, meant that we'd lose money even on paying customers sometimes. Meter for those things enough to close the gap, even if the pricing model becomes more complex. If a $5 dyno is still in the red even after cranking up all the dials to improve economy, then make it a $10 dyno instead.

* **Out-Cloudflare Cloudflare:** I remember having conversations about the Heroku routing layer becoming a "rich" proxy for Heroku apps five years before anyone knew what Cloudflare was. We talked about add-ons that'd be able to insert rich features in front of apps with a single install command (we already had the routing layer after all, and it was terminating TLS). Cloudflare executed so well that it would've been hard to compete, but who knows.

* **Let's Encrypt:** Went into public beta December 2015. Heroku would launch support for it with Automated Certificate Management (ACM) in March 2017, 1.5 years later, losing out to Amazon's AWS Certificate Management (also ACM) which launched January 2016. We were all so excited about the Let's Encrypt launch and should've had support in early 2016. Even with an AWS competitor, it would've been a major differentiator at the time.

* **Flexible dyno sizes:** Support 512 MB dynos all the way up to 384 GB dynos and everything in between. Support new EC2 generations as soon as they're available. Never let current offerings dilapidate. Make it clear what each dyno is, with no opaque dyno naming. `Standard-2X`, `Performance-M`, `Performance-L-RAM`, etc. sounds like you're hiding something.

* **Vertical and horizontal auto scaling:** Especially with the existence of the routing layer, this should've been very doable. This is the sort of feature that justifies the cost of a PaaS compared to IaaS.

* Support **Docker**, but don't let its existence make you crazy (it freaked all of us out when it was released). Buildpacks are still good, and have major advantages over requiring `Dockerfile`s for everything.

* **All in on Go:** We had some early proponents of Go in the company (you might say the earli_est_), but only amongst the ivory tower crowd who didn't touch production. It'd come into play eventually, but only after it wasn't very exciting to get into. We should have pulled it in early to replace the routing layer and CLI, and used it to demonstrate how at least in some cases, [language speed really does matter](/nanoglyphs/037-fast).

The critical consideration to tie it all together would be to keep Heroku's trademark simplicity even with added capability. Use the company's signature taste to avoid becoming AWS.

I've always liked the idea of **progressive enhancement**. A user's first interaction with Heroku should be `git push heroku master`, where they get to experience the delight of seeing their app up and running with a single command. But from there they should be able to peel back layers one at a time to achieve as much control as they need, whether it be isolation via VPC, automation via API, or debugging by SSHing into an active production box.

<img src="/photographs/nanoglyphs/049-paradise/boathouse@2x.jpg" alt="Boathouse at Cove" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/049-paradise/boardwalk@2x.jpg" alt="Boardwalk at Cove" class="wide" loading="lazy">

<!--

<img src="/photographs/nanoglyphs/049-paradise/dive-boat@2x.jpg" alt="Dive boat from another nearby resort" class="wide" loading="lazy">

-->

While we're at it, a few don'ts:

* **Don't get stuck in operational quagmire**. We had entire teams disappear into operational debt, never to reemerge. As important as production and operational stability are, products need to move forward.

* **Don't chase the serverless trend**. You want serverless in the Heroku sense where you have running containers without managing a server, but you don't need serverless in the AWS Lambda sense that does one request per container. The latter is a garish level of inefficiency with only the most marginal of benefits.

* **Don't make Dogwood** (Heroku Private Spaces). If it's not good enough to be run by a hobbyist and scaled up to an enormous user (who pays you a lot of money), then it's not good enough for that user either. Do make the alternate history version of Dogwood, a stack so meaningfully better than Cedar that it earns its name.

* **Don't do Cassandra:** Pretty obvious in retrospect. Anyone so big that they have to run Cassandra (very few people should) is also so big that they can afford to run it themselves.

* **Don't rebuild the API:** This one hurts. I spent over a year building out Heroku's new V3 API, taking the existing foundations and refining it for stronger convention, better flexibility, and modern best practices. This project was immensely satisfying (it felt like I should be working for free), but we ate up an enormous opportunity cost working on it. The API was the highest leverage component in the stack and we could've spent the time helping to ship other features instead.

<img src="/photographs/nanoglyphs/049-paradise/fish@2x.jpg" alt="Fish off the dock" class="wide" loading="lazy">

## Raja Ampat (#raja-ampat)

The photos from this issue are from Raja Ampat, an archipelago in Indonesia known for its incredible marine biodiversity.

We stayed at a resort called "Cove" on one of the archipelago's many tiny islands (you can walk across it in about half an hour), accessed via roughly three hours of boat travel from Sorong. Sorong's a busy place, but by the time you get out to Cove, there is nobody.

Cove's small island is called Yeben, and if you close your eyes and imagine what a deserted tropical paradise is supposed to look like, what you're seeing is probably a lot like Yeben. Far from any population centers, there you get perfect beaches, crystal clear water, and gorgeous vegetation. The island's forest is dense enough that you won't be going hiking in an arbitrary direction of your choice, but they'd done some nice trail building so you could follow a path up and over the top of the island to its southern side, then the beach to its eastern tip, then back north and west to get back to where you started.

In the evening hundreds of bats the size of sparrows flew in frantic zig zags overhead. Once in a while, a heavy flying fox would peel its massive body off the top of a tree and glide off.

<img src="/photographs/nanoglyphs/049-paradise/yeben@2x.webp" alt="Yeben Island" class="wide" loading="lazy">

The dozen tourist villas (there's only 20-30 visitors on the island at a time) were lined up along an oceanside boardwalk. A large, open-sided (flaps were drawn over in the event of rain) common building complete with couches and pool table served us for briefing and dining. As with most diving trips, we sunk into a routine schedule: wake early, get breakfast, load the boats, dive, surface interval on a random beach, dive, back to base for lunch, rest, back out for one more dive at a more local site.

At one of the beaches they'd take us for surface interval we came across these beautiful mangrove monitors. Being the first large land animal you've seen in days, it's exciting the first time you spot one, then as you look harder it dawns on you that the first wasn't alone, and there's a fair few scampering through the underbrush.

<img src="/photographs/nanoglyphs/049-paradise/mangrove-monitor-1@2x.jpg" alt="Mangrove monitor" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/049-paradise/mangrove-monitor-2@2x.jpg" alt="Mangrove monitor" class="wide" loading="lazy">

### On sharks and shrimp (#sharks-and-shrimp)

Yeben's shore was surrounded by shallows populated by huge schools of fish, and more unusually, dozens of black-tipped reef sharks. Most of them were tiny little guys, but the larger ones were big enough to occasionally breach their fins out of the water, creating an ominous presence when viewed sideways from shore (still small for sharks though). I could stand on the dock and snap photos as they passed below.

<img src="/photographs/nanoglyphs/049-paradise/sharks-1@2x.jpg" alt="Sharks(s) off the dock in Raja Ampat" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/049-paradise/sharks-2@2x.jpg" alt="Sharks(s) off the dock in Raja Ampat" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/049-paradise/sharks-3@2x.jpg" alt="Sharks(s) off the dock in Raja Ampat" class="wide" loading="lazy">

It might surprise non-divers to see how blasé divers are around sharks. Doing three dives a day I wasn't particularly predisposed to spend even more time in the water, but an Italian couple would finish their three only to immediately jump back in to go snorkeling around the pier for an hour, happily sharing the shallows with these little guys.

I was walking back along the dock one time, looked down, and to my surprise, there was a mantis shrimp looking straight back up at me!

Mantis shrimp build form-fitting circular holes for themselves, and back into them so they're looking out face first, framing big eyes, feelers, and praying mantis-like forearms. They're already strange looking creatures, and when the rest of their body is hidden away leaving only their face visible, seeing them is a little jolting, like seeing a facehugger from _Alien_ out of the corner of your eye.

You can often find mantis shrimp on dives in the area, but it was very lucky to see one through only a few feet of water like this. I wish I'd gotten a photo of this guy, but by the time I got back with a camera he'd evacuated the area, his hole already collapsed. Here's someone else's photo instead [2].

<img src="/photographs/nanoglyphs/049-paradise/mantis-shrimp@2x.jpg" alt="Mantis shrimp (credit Cédric Péneau)" class="wide" loading="lazy">

This Indonesia travel log is almost at the end, and about time too given this was all four months ago already. Next issue I have a few more photos from a nearby star-shaped lagoon that was quite beautiful, but that'll be the end of it.

Until next week.

[1] ChatGPT estimates that 1.4 billion human hours have been spent developing AWS.

[2] Mantis shrimp photo credit [Cédric Péneau](https://en.wikipedia.org/wiki/Odontodactylus_scyllarus#/media/File:Odontodactylus_scyllarus_R%C3%A9union.jpg).