+++
image_alt = "San Carlos Beach in Monterey at sunrise."
image_url = "/assets/images/nanoglyphs/010-monterey/san-carlos@2x.jpg"
published_at = 2020-02-01T02:56:57Z
title = "Monterey; the Darker Side"
+++

In January I did a dive weekend in Monterey doing some educational dives to finish my PADI Advanced Open Water certification. The photo above is of the Coast Guard Pier from the perspective of San Carlos Beach, taken a little before sunrise.

Monterey’s best known for its famous aquarium (as depicted in _Star Trek IV_ [1]) and multitude of adorable sea otters who are regularly visible from shore, but it’s also a dive spot well known to enthusiasts along the west coast. The Bay Area turns out to be quite hostile in terms of dive conditions, so it’s common for dive shops around here, and even as far north as Sacramento, to range all the way to Monterey to test their students in open water. I dove with a shop in Sausalito, and given how coastal the town is, you'd intuitively think they’d dive closer to home -- maybe even right off their back doorstep. But it's a case where intuition would be wrong, all their field tests happen during weekend trips to Monterey.

That said, just because Monterey isn't inhospitable for diving doesn’t make it hospitable either. I was using a 7 mm wetsuit (for reference, you can dive in many popular Caribbean and Pacific spots with a 3 mm no problem) with a hood, and two layers of it padding the torso. So much neoprene that, unable to bend in any of the normal places, I walked around all day with the stiff stride of Schwarzenegger's Terminator. I felt more like the Michelin Man though, and had the suit been white instead of black, would have looked like him too.

Even with all the suit’s trouble, that water was _cold_. Like sap-the-life-out-of-you cold. Normally the worst part about getting into the ocean is the shock of the water in those first few moments, but here it was reversed. Carrying heavy equipment to the beach and the excitement of pushing out past waves and avoiding submarine rocks tended to keep you pleasantly warm, but over a dive's next 30 to 45 minutes of being relatively inactive in 12C (54F) water, that heat drains away. It wasn't long before I found myself fantasizing about the Monterey shore -- also about 12C midday on that Sunday, but which felt like a sunbaked beach in Cabo by comparison.

---

The ocean’s conditions were no less forgiving than its temperature. We got lucky on day one with relatively calm currents and good visibility, but day two wasn’t as good. We gathered on the beach at 6 AM to do a pre-dawn dive. On our first entry, the waves were especially rough, and the weight and surface area of our dive gear made getting out past them out to open water difficult. In the first ten minutes of the dive our group ran into 1 lost mask, 1 broken fin (mine, as a wave knocked me back into a rock and snapped it in half), 1 air tank escaped from its strap, and 1000 PSI lost to a bad free-flowing reg. The lost mask ended the expedition, and we swam back to the comfort of land. With constant tumultuous currents and visibility measured in feet, we never did see that mask again.

![Storefront](/assets/images/nanoglyphs/010-monterey/storefront@2x.jpg)

Trying to fix my newfound fin problem, a few minutes later I walked into one of the local dive shops in my soaking wetsuit, plunked my credit card down on the counter, and said, “I need to rent some fins for the day.”

Moments later, our group’s other dive instructor walked into the same local dive shop in his soaking wetsuit, plunked his credit card down on the counter, and said, “I need to rent some fins for the day.”

For three seconds I thought he was mocking me, but no. The other dive team had faired no better than ours -- within minutes of starting their dive, he’d lost a fin to a bad wave, and like our group’s mask, it was never seen again.

We did eventually make it out to sea, but had to prolong the next dive to make up for the miss on the first. The best part of the day was a submarine knot-tying exercise -- it would have been entertaining in any conditions, but it was especially good on this day because the energy it took to stabilize a distressed weight belt on the ocean floor in chaotic ocean currents was perfect for getting warm.

So, all in all, awesome trip.

---

What’s all that got to do with software? Not much, but to fit the somber feeling evoked by the Pacific's icy depths, I’ve selected some amazing pieces that showcase some of software’s darker aspects.

A sin that the computing blogosphere could fairly be accused of is cheerleading to a fault. In sharp contrast to the internet at large, it’s quite common to present new technologies and new practices in an unequivocally positive light to build hype around them, but less common to later talk about their problems as those become known. The generally positive spin is mostly a good thing as generates interest and catalyzes discovery, but presents a distorted view to potential adopters. In this edition we’re briefly going to put a finger on the other side of the scale.

---

## Asyncastrophe (#asyncastrophe)

The asynchronous programming model sure seems to be a strong candidate for the future of most concurrency, with a big uptake in recent years across C#, JavaScript, Python, and Rust. [I'm not feeling the async pressure](https://lucumr.pocoo.org/2020/1/1/async-pressure/) goes into one of its major pitfalls.

It describes how despite the model being willing by default to try for nearly limitless concurrency (e.g. an asynchronous socket connect will happily spin off as many tasks as there are new connections), there’s almost certainly a bottleneck somewhere in the system, and that bottleneck will eventually cause it to degrade in ways that are unexpected, and possibly catastrophic.

For example, an asynchronous HTTP server may allow an unlimited number of clients to connect, but if the program is backed by a database and connection pool, that will put a hard cap on the number of those clients that can actually be served. Clients unable to acquire a database connection get parked in an async task until something eventually times out. Given a severely overloaded system, this could easily degenerate to a point where a majority of requests are failing messily.

The article suggests building flow control into programs and into the APIs of underlying async libraries. From the example above, a program might track its own busyness based off the number of clients currently being served, and respond to new connections with a 503 if deemed to be too high instead of assuming the best by blindly handling it.

``` python
request_handler = RequestHandlerService()
if not request_handler.is_ready:
    response = Response(status_code=503)
else:
    response = await request_handler.handle(request)
```

## My code’s compiling (#my-codes-compiling)

Pingcap writes about [Rust’s compilation model “calamity”](https://pingcap.com/blog/rust-compilation-model-calamity/). They develop a major Rust program containing roughly 2 million lines of code (including dependencies) and find that compilation times can be a major hit to productivity as it takes 15 minutes to compile in development and 30 minutes for release. The author goes on to outline some of Rust’s features that lead to longer compile cycles, many of which are designed to increase safety or runtime speed, but with the price of more heavy lifting during compilation.

A commenter on HN point out that these times aren’t actually _that_ bad in that they'd be similar for a C++ program of the same size. Someone else crunched the numbers to estimate that a Go program of this magnitude would compile quite a bit faster in about 2 minutes, a speed up that's probably mostly attributable to language’s infamous shunning of generics. A major boon to productivity through safer data structures and less repeated code also turns out to be a detractor in longer compile times.

It’s hard not to be in favor of generics, but a fast edit-compile-run loop really is a game changer in boosting productivity. Having to repeat code and write with `interface{}` in Go is annoying, but in a very real way the language feels _much_ more productive than anything else because everything is just _so_ fast. `go build`? It’s done in less time than it took you to type it. `go test`? Finished milliseconds after you hit the return key. The computer waits on the human instead of the human waiting on the computer, which is a very good (and practically unique) feeling.

## Wirth’s eternal law (#wirth)

[Software Disenchantment](https://tonsky.me/blog/disenchantment/) bemoans the state of modern software development and how we’ve steadily been moving towards a world of infinitely faster hardware, but make software that never seems to get faster regardless of the horsepower under the hood. Loading web pages can take 10s of seconds, and bleeding edge browsers running on the latest PCs can’t scroll them at 60 FPS. There are web pages larger than the entirety of the Windows 95 operating system (30 MB). Despite today’s apps and mobile operating systems being pretty much the same as they ever were, phones more than a few years old can’t run them because the software has expanded in size and resource requirements, seemingly without gaining anything of value.

A core cause of this is that like liquid filling a container, the tools and techniques we use to develop software seem to cause it to expand in size and deepen in abstraction until the best hardware of the day is saturated (which is [not a new idea](https://en.wikipedia.org/wiki/Wirth%27s_law)). From the article:

> What it means basically is that we’re wasting computers at an unprecedented scale. Would you buy a car if it eats 100 liters per 100 kilometers? How about 1000 liters? With computers, we do that all the time.

If you tackle this one, make sure to read all the way through to the end, which gets less nihilistic around the headline “It’s not all bad”. The author feels that the first step in recovery is acknowledging that we have a problem, then backpedaling until reaching a point where tools and techniques allow fast, quality software to be delivered reliably.

---

## Challenging the deep (#challenging-deep)

Back to the ocean: I found [this article](https://www.theatlantic.com/magazine/archive/2020/01/20000-feet-under-the-sea/603040/) from The Atlantic to be a very good read. Nominally about ocean mining and its potential environmental impact, it goes far beyond that, talking about the unforgiving _hadal zone_ and history of its exploration.

The hadal zone is even deeper than the abyssal zone at 20,000 to 36,000 feet, and appropriately named for Hades, the Greek underworld (could marine naming be any more bad ass?). It’s been visited by humanity only a handful of times, most notably with expeditions to the Challenger Deep in the Mariana Trench like Jacques Piccard and Don Walsh's 1960 descent in the bathyscaphe _Trieste_. 30,000 feet down, the enormous pressure at those depths tore a crack in their viewing window, shaking the entire ship -- you just have to imagine how scary of a moment that would have been, and appreciate the sheer gusto that it would’ve taken to ignore it and keep going down to their final depth of 35,814 feet.

![The Mariana Trench](/assets/images/nanoglyphs/010-monterey/mariana-trench@2x.jpg)

Aside from the US Navy (owners of the _Trieste_), some of the Deep’s main prospective explorers over the years have been billionaires (or at least 100s of millionaires), with James Cameron making in 2010 on an expedition that by all accounts excelled in a technical sense despite his background as a filmmaker. Cameron made it down, but had to ascend earlier than planned and cancelled his follow up dives. Narrating the adventure from his superyacht's helicopter overhead was Paul Allen. As a foil to Cameron, Richard Branson blustered about making a descent of his own in a science-fiction submersible reminiscent of a fighter jet, but didn’t make any headway on the project.

Reading about the difficulties of these deep-diving projects made me feel a little better about my comparatively tame troubles with a few waves and cold water.

Until next week.

[1] _Star Trek IV_ asserts that the aquarium visited by Kirk and Mr. Spock is across the Golden Gate in Sausalito, but it's actually the one in Monterey, which at a ~3 hour drive from San Francisco, wouldn't have been quite so plot-convenient to get to.
