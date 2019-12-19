+++
published_at = 2019-12-19T20:48:59Z
title = "MOMA & the Rain"
+++

![The San Francisco MOMA](/assets/images/nanoglyphs/006-moma-rain/moma@2x.jpg)

It’s been a catastrophic few weeks of weather in the Bay: almost, kind-of rain. The cloud cover never goes away, and most of the time it feels like it might be so presumptuous as to start to drizzle. A few times it’s even gone as far as to openly rain for a little while, in clear violation of the California weather contract. Winter jackets, toques, and scarves are everywhere (although it’s still 10C+).

Having dinner out Sunday night, the restaurant we were in (along with the entirety of the Cole Valley neighborhood) suddenly went black. The power never came back on, and the staff requested payment in cash before closing up shop. Most cities lose power after feet of snow and sheets of ice push an overexerted system past its limit. In fair-weather San Francisco, we're done in by a few millimeters of rain.

---


It was an indoor type of weekend, so I visited San Francisco’s <acronym title="Museum of Modern Art">MOMA</acronym>. The exhibit I went to see was a display of space-faring art from the golden age of space 70s space optimism (called _Far Out_). Given the quality and quantity of the source material, it had huge potential, and even though it didn't quite live up to it (the number of included pieces was relatively small), it was still a worthwhile visit. You can get a good feel for the project from [NASA's space colony art](https://settlement.arc.nasa.gov/70sArtHiRes/70sArt/art.html).

![Toroidal Colonies by Rick Guidice](/assets/images/nanoglyphs/006-moma-rain/rick-giudice-toroidal-colonies@2x.jpg)

Many of the exhibits at the MOMA fit the stereotype of modern art perfectly: simple geometric shapes and canvases made up of a solid colour. One tip for visitors who aren’t committed to a full ticket: so far it’s held true that the best piece of art at the MOMA is free to see.

There’s a large format exhibit available free to the public through the entrance off Natoma St. These days, it’s the huge moving display of the people of San Francisco pictured above. Maybe the closest thing to a [_Daily Prophet_](https://harrypotter.fandom.com/wiki/Daily_Prophet) you can get in real life.

![Sequences at the MOMA](/assets/images/nanoglyphs/006-moma-rain/sequences@2x.jpg)

Previously, the space was used by _Sequences_ -- a labyrinth of sorts where you’d go in one end following the wall and feel like you’re walking in circles for an unlikely amount of time before getting spit back out the other end, all the while never encountering an intersection or crossing over another tunnel (walking it is really, _really_ cool). It’s since been relocated back to its home at Stanford.

---

## Next generation mainframes (#mainframes)

I _love_ that IBM is still making mainframes. For many of us, the mainframe occupies a special place of veneration in our hearts and minds — having never worked with one, we don’t really understand them, but have heard stories over the decades about their prominent role in critical infrastructure like banking and telecommunications. Almost everyone would be better served by a fleet of commodity servers — better scalability, more flexible, better geographic distribution, cheaper — but sometimes you just want a behemoth in a box.

WikiCheap dives into the [technical details of the Z15](https://fuse.wikichip.org/news/2659/ibm-introduces-next-gen-z-mainframe-the-z15-wider-cores-more-cores-more-cache-still-5-2-ghz/), successor to the Z14. Processors in a mainframe are organized into “drawers” for in-machine redundancy, with each drawer connected to every other drawer by way of an “A-Bus”. Although drawers in the Z15 have two fewer processors compared to the Z14 (from six down to four), they have two more cores (from 10 to 12 cores), and the machine as a whole supports an additional drawer. Clock speed is the same at 5.2 GHz, but L2 and L3 cache capacities have doubled (from 2 to 4 MB and 128 to 256 MB respectively).

## Vectorizing SQL execution (#vectorizing-sql)

Cockroach Labs writes about [builiding a vectorized SQL engine](https://www.cockroachlabs.com/blog/how-we-built-a-vectorized-sql-engine/#) which resulted in a 70x speed up in their microbenchmarks and a 4x speed up in the more general [TPC-H benchmark](http://www.tpc.org/tpch/). In a nutshell, the improvements mainly come from (1) rebuilding their datum representation so that Go type conversions are necessary because data is already available in the required type, and (2) making sure that type operations (a multiplication in this example) occur in a tight loop so that the CPU can make better use of pipelining and necessary data packed densely into low-level caches. The latter is why the new engine is considered to be "[vectorized](https://en.wikipedia.org/wiki/Array_programming)".

The post is incredibly detailed and although nominally about vectorization, it does a great job of demonstrating how to optimize in Go more generally. Write an ideal "speed of light" case to understand the maximum speed up possible, then use a profiler to drill in and remove bottlenecks, starting with the most expensive. This is often possible by using `pprof` to look at lines at Go code, but occasionally requires dropping into assembly, which the tooling makes easy.

There's some metacommentary to be made on Go's adversion to generics. At one point the Cockroach team resorts to using Go templates to generate code suited to each primitive so as to avoid type conversions. It works, but a critical reader will notice that it's a leaky solution to solve a problem which is entirely Go-induced.

## Incentivizing speed (#incentivizing-speed)

The predominance of Google as a search engine and Chrome as a web browser rightfully has a lot of people worried, but longer term ramifications aside, they’ve got a pretty good track record for helping to foster constructive patterns on the web. Websites that didn’t support encryption and those that weren’t mobile friendly we’re given slight penalties in search results, and that acted as major incentive for owners to fix those problems. (As far as I could tell, _brandur.org_ only finally started outranking its dot com analog owned by a famous Faroese pop singer after he refused to update it with either modern encryption practices or to be mobile-friendly.)

More recently, Chrome is planning to add a [slow page indicator](https://blog.chromium.org/2019/11/moving-towards-faster-web.html) which shows a message like “Usually loads slow” to the user on sites which are known to do so. That might just be the fire those site owners need to be lit under them to get the problem fixed.

In the pioneering days of the internet speed was a distant secondary concern compared to feature development, if it even made a list of concerns at all. These days, with internet usage baked into every part of modern life, heavier frontend stacks leading to heavier software bundles, and wider geographic distribution of users, it’s something to get serious about. Luckily for many, it tends to be the type of problem that turns out to return some big wins given even a moderate amount of attention invested.

---

!["Clear Day with a Southern Breeze" (Pink Fuji) by Katsushika Hokusai](/assets/images/nanoglyphs/006-moma-rain/pink-fuji@2x.jpg)

I'll leave you with an article from the 1843 magazine, [Hokusai: old man, crazy to paint](https://www.1843magazine.com/culture/look-closer/hokusai-old-man-crazy-to-paint). It's a short digest of the work of the Japanese artist Katsushika Hokusai, best known for _The Great Wave_ (or more formally, _Under the Wave off Kanagawa_).

Hokusai aimed to improve his art with every passing year, and wanted to live beyond the age of 100 so that he could achieve a dominating level of mastery. He was on record as saying that nothing he'd created before of the age of 70 was worthy of notice (both pieces of art depicted here were produced at 71). Later in his life he started calling himself "Gakyo Rojin", translating to "Old Man Crazy to Paint". Having produced what is perhaps the most iconic piece of Japenese art in history, maybe he was onto something, and maybe his was a philosophy that we could all learn from.

!["Under the Wave off Kanagawa" (The Great Wave) by Katsushika Hokusai](/assets/images/nanoglyphs/006-moma-rain/great-wave@2x.jpg)

See you next week.
