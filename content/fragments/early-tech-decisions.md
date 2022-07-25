+++
hook = "On how software that survives first encounter with production tends to live longer than we think it should."
published_at = 2022-07-23T18:59:32Z
title = "The disproportionate influence of early tech decisions"
+++

Spend five years at a hypergrowth startup like Stripe, and you see a lot changes during that time. Organizationally, it's night and day, as a few hundred people scaled to thousands, the structure adapted to teams with charters and responsibilities that were much more fixed, and with a rigid managerial hierarchy to ensure continued product momentum even with many more hands involved. Similarly, culture, process, and planning methodology all shifted enormously.

But notably, a place where change was either much slower, or even non-existent in some cases, was the technical realm. This is unsurprising when it comes to the monolithic foundations of the platform, which are infamously hard to change even for the most productive companies. e.g.

* The underlying database: Mongo originally begot more Mongo, but switching any database anywhere is unusual.

* Programming languages. Although more were added to the mix over time, all the originals were still firmly in place, with the central API still be driven by the same language it'd always been driven by -- Ruby.

* Cloud provider: AWS originally, and with so much custom infrastructure built on its APIs, unlikely to ever change.

But again, everybody knows that it's hard to migrate a database or rewrite code in a new language, so this status quo wouldn't be surprising anywhere you find it.

## Not only the megalithic (#megalithic)

What is more surprising is that it's not only the big stuff that has a tendency to stay fixed. It's the small and medium-sized elements as well. A few examples out of thousands:

* The core code for executing an API method and rendering an API resource was basically the same as what'd be written eight years before, just with many more embellishments added. This might not be so bad except that there was near universal consensus that it was a mess.

* API workers ran on a customized Thin with graceful restarts provided by way of Einhorn. With many advancements in the Ruby ecosystem having occurred since, neither of these would be considered for even one second in a modern Ruby stack.

* An early decision was made to punt on the use of threading. A decade later, it was still never turned on -- no one would say threading was a bad thing per se, but the marginal improvement it would allow in a language like Ruby never ended up being worth the massive effort involved in getting it enabled safely.

* Projects started on a [NIH](https://en.wikipedia.org/wiki/Not_invented_here) Ruby ORM for Mongo, which would prove ~impossible to get off of due to the sheer amount of use and deep customization.

* The system for sending webhooks was largely unchanged since when it was put in. It got a major queue overhaul from a custom stack to Kafka at one point, but requests were still being made by the same single-threaded Ruby processes as they always had been.

* Users could remove their test data by way of a specialized background worker that'd query each collection and empty it row-by-row. It was known to be inefficient and probably unsustainable even the day it was put in, and everyone _knew_ it'd be replaced by something better ASAP. Almost a decade later it was still chugging along, having gotten just enough attention during that time to keep it just functional enough so it could be plausibly claimed to be doing its job.

A common theme is that all of these were known to be very imperfect implementations even when they were originally added, but they were added anyway in the name of velocity, with an implicit assumption baked in that they'd be shored up and improved later when there were more resources and slack time to do so. And although they certainly were shored up, they never improved by all that much. Another common theme is that they all lived _far_ longer than their progenitors would've ever expected, as it turned out that slack time is chronically a vanishingly rare luxury in engineering organizations, and software is _always_ hard to change once it's firmly cemented in place.

## Velocity and maintainability: A balancing act (#balancing-act)

So what's my point? Simply this: software has inertia.

Most of it will eventually die by virtue of rewrites, or the products, companies. or organizations using it dying themselves, but what makes it passed that initial push for survival will likely last longer than expected. The [Lindy effect](https://en.wikipedia.org/wiki/Lindy_effect) states:

> by which the future life expectancy of some non-perishable things, like a technology or an idea, is proportional to their current age. Thus, the Lindy effect proposes the longer a period something has survived to exist or be used in the present, the longer its remaining life expectancy.

Young companies push development aggressively because they're optimizing for their survival -- spending too much time agonizing over writing the perfect code might lead to the product not shipping, which might lead to the company's total failure.

It's not a bad instinct, but quality is more of a sliding scale than it is a good or bad dichotomy, and I'd argue that many small companies optimize too much in favor of speed by trading away too much in terms of maintainability by shipping the first thing that was thrown at the wall.

And this fails the other way too, where major believers in academic-level correctness agonize over details to such a degree that projects never ship, and sometimes never even start.

As with most things, the answer is somewhere in the middle. Spend time thinking and planning, but not to a degenerate extent -- it's also important to _do_. Refactoring is a key part of the equation -- code is never right the first time, it converges on right through many iterations. And ideally the first couple refactors are _significant_, not only small patches that leave the bulk unchanged. More refactoring passes are better, but subsequent ones will produce [diminishing returns](https://en.wikipedia.org/wiki/Diminishing_returns).
