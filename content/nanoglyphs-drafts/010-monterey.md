+++
image_alt = "San Carlos Beach in Monterey at sunrise."
image_url = "/assets/images/nanoglyphs/010-monterey/san-carlos@2x.jpg"
published_at = 2020-02-01T02:56:57Z
title = "Monterey; the Darker Side"
+++

I’m recently back from a dive weekend in Monterey where I was doing some instructor-led dives to finish my PADI Advanced Open Water certification. The photo above is of the Cost Guard Pier from San Carlos Beach, taken a little before sunrise.

Monterey’s best known for its famous aquarium (as depicted in _Star Trek IV_) and its multitude of cute sea otters that you can regularly spot from shore, but it’s also a well-known dive spot. The Bay Area especially is somewhat hostile in terms of dive conditions, so it’s common for dive shops around here and even as far north as Sacramento to send groups all the way to Monterey for field tests. The dive shop I went with was located in Sausalito for example, and given just how coastal that is, you’d intuitively think they’d dive closer to home (but they don’t).

That said, just because it’s not inhospitable for diving doesn’t mean that it’s hospitable either. I was using a 7 mm wetsuit (for reference, you can dive in many popular Caribbean and Pacific spots with a 3 mm no problem) with a hood, and two layers worth on the torso. It was so much neoprene that unable to easily bend in most places, I walked around land with a stiff terminator stride. I felt like the Michelin Man, and had the suit been white instead of black, would have looked like him too. And even with all the suit’s trouble, that water was _chilly_. Getting in initially was okay thanks to all the excitement of pushing past waves and avoiding submarine rocks, but at 12C, the water saps the heat out of you. After being in there only 30 minutes, I found myself fantasizing about the 12C air temperature on land, which looks like the same temperature but feels _far_ warmer.

--

The ocean’s conditions were no less forgiving than its temperature. We got lucky on day one with relatively calm currents and good visibility. Day two wasn’t so good. We gathered on the beach at 6 AM to do a pre-dawn dive. On our first try getting in, the waves were especially rough, and the weight and surface area of our dive gear didn’t make it easy getting past them. In the first ten minutes of the dive our group ran into 1 lost mask, 1 broken fin (mine, as a wave knocked me back into a rock and snapped it in half), 1 air tank escaped from its strap, and 1000 PSI lost to a bad free-flowing reg. The lost mask ended the expedition, and we swam back to the comfort of land. With strong currents and visibility measured in feet, we never did see that mask again.

![Storefront](/assets/images/nanoglyphs/010-monterey/storefront@2x.jpg)

Trying to resolve my new fin shortage, a few minutes later I walked into one of the local dive shops in my soaking wetsuit, plunked my credit card down on the counter, and said, “I need to rent some fins for the day.”

Moments later, our group’s other dive instructor walked into the same local dive shop in his soaking wetsuit, plunked his credit card down on the counter, and said, “I need to rent some fins for the day.”

I thought I was being mocked, but no. The shop’s other dive team had faired no better than ours -- within minutes of starting their dive, he’d lost a fin to a bad wave, and like our group’s mask, it was never seen again.

We did eventually make it out, although we had to prolong the next dive to make up for the miss on the first. The best part of the day was a submarine knot-tying exercise -- it would have been entertaining in any conditions, but it was especially good on this day because the extra energy it took to stabilize the weight belt we were trying to rescue from the ocean floor in chaotic ocean currents was perfect for staying warm.

---

So what’s all that got to do with software? Admittedly, not much, but to fit the somber feeling of the mood photo above, I’ve selected some amazing pieces that showcase some of software’s darker aspects.

A sin that the computing blogosphere could fairly be accused of is too much cheerleading. In sharp contrast to the internet at large, it’s quite common to present new technologies and new practices in an overly positive light to build hype around them, and less common to later talk about their problems as those become known. The generally positive spin is mostly a good thing, but it often presents a distorted view to potential new adopters. In this edition we’re briefly going to put a finger on the other side of the scale.

---

## Asyncastrophe (#asyncastrophe)

Asynchronous programming models sure seem to be a popular candidate for the future of concurrency, with a big uptake in them across C#, JavaScript, Python, and Rust in recent years. [“I’m not feeling the async pressure”](https://lucumr.pocoo.org/2020/1/1/async-pressure/) goes into one of it’s major pitfalls.

Despite the model being willing by default to try for nearly limitless concurrency (e.g. an asynchronous socket connect will happily spin off as many tasks as there are new connections), there’s almost certainly a bottleneck somewhere in the system, and that will eventually cause it to degrade in ways that are unexpected and possibly catastrophic.

For example, an asynchronous HTTP server may allow an unlimited number of clients to connect, but if the program is backed by a database and connection pool, that will put a hard cap on the number of those clients that it can actually server. Clients unable to acquire a database connection get parked in an async task until they eventually timeout. Given an overloaded system, this could easily degenerate to a point where a majority of requests are failing.

The article suggests building flow control into programs and into the APIs of underlying async libraries. From the example above, a program might track its own busyness based off the number of clients currently being served, and respond to new connections with a 503 if deemed to be too high instead of assuming the best by blindly handling it.

## My code’s compiling (#codes-compiling)

Pingcap writes about [Rust’s compilation model “calamity”](https://pingcap.com/blog/rust-compilation-model-calamity/). They develop a major Rust program containing roughly 2 million lines of code and find that compilation times can be a major hit to productivity as it takes 15 minutes to compile in development and 30 minutes in release mode. The author goes on to outline some of Rust’s features that lead to longer compile cycles.

A commenter on HN point out that these times aren’t actually _that_ bad in that compilation would be similar for a C++ program of the same size. Someone else crunched the numbers to estimate that a Go program of that size would compile quite a bit faster in about 2 minutes, but it has to be considered that this is mostly possible because of language’s shunning of generics. What would be a major benefit to developer productivity through safer data structures and less repeated code would also a detractor in longer compile times.

It’s hard not to be in favor of generics, but a fast edit-compile-run loop is of paramount importance. Having to repeat code and write with `interface{}` in Go is annoying, but in a very real way it feels _much_ more productive than any other language because everything is just so fast -- `go build`? It’s already done. `go test`? Finished 50 ms after you hit the return key. The computer waits on the human instead of the human waiting on the computer, which is a very good (and unfortunately very unique) feeling.

## Wirth’s eternal law (#wirths)

[Software Disenchantment](https://tonsky.me/blog/disenchantment/) aggrieves the modern state of software and how we’ve been steadily moving towards a world of much faster hardware, but make software that never seems to speed up regardless of the horsepower under the hood. Loading web pages can take 10s of seconds, and modern browsers running on modern PCs can’t scroll them at 60 FPS. There are web pages larger than the entirety of the Windows 95 operating system (30 MB). Despite today’s apps and mobile operating systems being pretty much the same as they ever were, phones more than a few years old can’t run them because the software has expanded in size and resource requirements, seemingly without gaining anything of value.

A core cause of this is that like liquid filling a container, software tools and techniques seem to cause it to expand in size and deepen its layers of abstraction until the best hardware of the day is saturated. From the article:

> What it means basically is that we’re wasting computers at an unprecedented scale. Would you buy a car if it eats 100 liters per 100 kilometers? How about 1000 liters? With computers, we do that all the time.

If you tackle this one, make sure to read all the way through to the end, which gets a little less nihilistic around the headline “It’s not all bad”. The author feels that the first step in X is acknowledging that we have a problem, then backpedaling until reaching a point where tools and techniques allow fast, quality software to be delivered reliably.

---

## Challenging the deep (#challenging-deep)

Back to the ocean: I found [this article](https://www.theatlantic.com/magazine/archive/2020/01/20000-feet-under-the-sea/603040/) from The Atlantic to be a very good read. Nominally about ocean mining and its potential environmental impact, it goes far beyond that, talking about the hadal zone and the history of its exploration.

The Hadal zone is even further down than the abyssal zone at 20,000 to 36,000 feet, and appropriately named for Hades, the Greek underworld. It’s been visited by humanity only a handful of times, most notably with expeditions to Challenger Deep in the Mariana Trench starting with one in 1960 by Jacques Piccard and Don Walsh in the bathyscaphe _Trieste_. As they descended the enormous pressure cracked their viewing window -- you just have to imagine how scary of a moment that would have been, and appreciate the sheer gusto that it would’ve taken to ignore it and keep going.

![The Mariana Trench](/assets/images/nanoglyphs/010-monterey/mariana-trench@2x.jpg)

Aside from the US Navy (which owned the _Trieste_), some of the Deep’s main prospective explorers have been billionaires (or at least 100s of millionaires), with James Cameron making in 2010 on an expedition that by all accounts was extremely technically adept despite his background as a filmmaker. As a foil to Cameron, Richard Branson blustered about making a descent of his own in a science-fiction submersible reminiscent of a fighter jet, but didn’t make any headway on the project.

Reading about the difficulties of these expeditions sure threw my troubles with some cold water and a couple waves into sharp contrast.

Until next week.
