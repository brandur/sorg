+++
image_alt = "The mountains of Canmore"
image_url = "/photographs/nanoglyphs/017-twenty/canmore@2x.jpg"
published_at = 2021-01-06T18:21:00Z
title = "Twenty Years"
+++

There was a time when a personal newsletter was a pretty unusual beast. I might have to contextualize the concept by describing a push versus pull model, the rise and fall of RSS, and how although email is one of the oldest internet technologies, it may very well still be the most important. But with [Substack's meteoric rise](https://www.newyorker.com/magazine/2021/01/04/is-substack-the-media-future-we-want) in 2020 we now have newsletters coming out of the woodwork, and most of you will be well-versed with the idea by now.

I'm Brandur. I mostly talk about APIs, databases, and bad software stacks. This _is_ a newsletter, but _not_ a Substack. I send it on what was supposed to be a weekly cadence, but it's turned into something more like quarterly. I recently published some articles on [**feature casualties of large databases**](/large-database-casualties) and [**offset pagination**](/fragments/offset-pagination), and you may have discovered it through one of those.

As always, if you don't want to see it anymore, you can unsubscribe in [one, easy click](%unsubscribe_url%).

Welcome to 2021.

---

## 69 characters (#69-characters)

Those of you who've been on this list for a while will notice that I'm playing with _Nanoglyph's_ design/layout a bit. It started when I was reviewing some previous issues and realized that my fonts were way too small on every device that mattered, which was quickly rectified.

That snowballed, and the most obvious change this time around is the removal the `max-width` constraint newsletter's email version. I've long subscribed to the idea that [narrow text widths are more optimal for readability](https://ux.stackexchange.com/questions/108801/what-is-the-best-number-of-paragraph-width-for-readability), and HTML aside, for the longest time I sent text-based email to everyone I knew manually wrapped at exactly 69 characters.

I stopped doing that because even though it was a better reading experience on wide screens, it degraded poorly. On smaller screens especially, the manual breaks made messages very difficult to read. (If you haven't seen this before, try reading PG Hackers on your smartphone to see what I mean.) Not wrapping text creates a wide ocean of endlessly long text edge-to-edge on large screens, but hands control back to the user who can pick their ideal window size and have text wrap accordingly.

I'm applying the same principle back to _Nanoglyph's_ design. Emails get no `max-width` so you can pick your own. The [web versions](/nanoglyphs/017-twenty) still have one, widened slightly compared to before. I also have a new scheme for storing photos outside Git and am running them all through [a MozJPEG size optimization pass](/fragments/libjpeg-mozjpeg), so I'm including higher resolution versions of them.

---

## Post-secondary, post-crash (#post)

I'll save you some platitudes by skipping the usual repartee about 2020. For me, its most distinctive characteristic was that technology -- already indisputably one of the most powerful forces in world -- doubled down in its sheer dominance over every facet of our lives. Already integral to every one of our most important industries like communication, transport, and entertainment, this year it entwined itself in the very way we live, how we conduct business, and the most personal aspects of human connection, as it became the sole portal to our closest friends and family.

Like for the boiling frog, we forget that the water wasn't always this hot. Computers have been around a long time, but the total ascendency of tech wasn't always a given, and hasn't been one for long.

Rewind to the mid-aughts, and me and my geek friends working our way through university. Through our entire childhood we always knew that the _really_ successful members of our cohort entered med school, dentistry, and law. That wasn't us. We'd blown our shot long ago by failing to maintain 100.0 grade point averages, [IB accreditation](https://en.wikipedia.org/wiki/International_Baccalaureate), good standing on the debate team, and membership in a half dozen other recommended extracurriculars.

Right behind those truly elite degrees were the "respectable" engineering disciplines. _Ideally_ you'd make the grades in first year to enter oil & gas or mechanical engineering. If not, majoring in electrical or chemical wasn't the worst thing that could happen to you -- good enough that your family wouldn't disown you at least. Just make sure that whatever you do, don't -- and I mean _don't_ -- slip below the line of polite society and into those -- _yuck_ -- computing majors. Only the tiniest of sidesteps away from those nerds in computer science, and _really_, not even real engineering. Why even bother with university at that point? Go to SAIT, and develop the same skills in half the time and a quarter of the cost.

Computers were important -- We were all glued to the padded blue walls of Facebook and busy fragging each other in Counterstrike -- but they weren't a career. Everyone knew that sure, they run things, but their place in business was closer to primitive infrastructure like electricity, or the building's plumbing system. They _certainly_ weren't a salable product unto themselves -- the unforgivable hubris of pets dot-com and the early 2000s was definitive proof of that. The economy's true profit centers were, and always would be, the known staples of industry -- piping oil, building cars, mining coal. Computers would be there of course, to handle payroll and accounting, in the back room where they belong.

Even if against all odds you did land a job (probably in IT), _everyone_ knew that your days were numbered. All computer-related divisions at any company of note were already halfway out the door to being outsourced in India or China. The ones that weren't were just walking dinosaurs, gone just as soon as an enterprising manager found the time to drown them offshore.

Ignoring common sense and all sage advice, my friends and I, who spent the lion share of our free time coding D-grade video games, running unencrypted FTP servers, and writing table-based HTML layouts, went into computing anyway, ready to throw our lives away in pursuit of doing something we liked. The good news is that getting in was easy -- computer engineering had so many vacant positions that I doubt anyone even bothered to check my marks on the way in.

---

So, a few years later, we got our worthless degrees, accepted jobs that were salaried but paid only a little better than minimum wage, and spent the next few decades saving up to buy small houses in the deep exaburbs, admiring our high-powered petroleum engineering ex-brethren from afar as they drove from skyscraper to pied-à-terre penthouse in new-off-the-lot M-series beamers, all the while crying softly during our 90-minute commutes. The end.

---

Or that's how it was supposed to go anyway.

Instead, the last few decades have shown a turnaround of colossal proportion. At the turn of the century, the S&P 500's top slots were filled by GE, Exxon, Pfizer, Citigroup, and Cisco. By 2020: Apple, Microsoft, Amazon, Facebook, and Alphabet. The S&P 500 had a great year -- 16% up -- but if you'd skipped the index in favor of holding FAANG long instead, you'd have netted somewhere in the range of 30 to 80% returns, just in 2020. The unicorn tech IPO train, and Tesla, have absolutely dominated as the most important market news in years. Facebook, Google, and Twitter eclipse our national discourse, and are on the lips of everyone from school children to comp-sci PhDs to congresspeople.

I work in the valley, and new employees stem disproportionately from the America's top universities. People who would have been doctors and petroleum engineers in my generation, but who instead chose to enter the less traditional, but possibly more lucrative, technology field in Silicon Valley. Undoubtedly, their marks and list of credentials are so high and long that they'd make my eyes bleed, but somehow, I've snuck my way into this bastion of academic prodigies and overachievers.

Meanwhile, my "elite" friends back home are hurting. I'm from a petrol city, and between fracking and failed pipelines, oil and gas, once the pinnacle of the profit-driven world, has been crumbling for years. Friends who went into medicine often didn't make their specialties, and are disillusioned by the idea of becoming GPs. Friends who went into law hate it. Everyone wants to move to Vancouver.

But work and finance aside, what's new last year in particular is how deeply technology infiltrated our lives. We were on the path already, but the virus acted as a catalyst that accelerated the process by decades.

In 2020 I took practically every one of my work meetings from my bedroom. I switched to getting 90%+ of my groceries online (previously, 0%). I spent Fridays and Saturdays on Twitch, keeping company with favorite DJs as they played to virtual dance floors and ran emoji hype trains in exchange for "subs" and "bits". I watched as a combination of technology and salt-the-earth lockdown policies razed brick and mortar to the ground, probably permanently, and made already-ginormous Amazon the most important company in the world, and Jeff Bezos its wealthiest billionaire. I spent my first ever Xmas morning on Zoom.

<img src="/photographs/nanoglyphs/017-twenty/da-vinci@2x.jpg" alt="Da Vinci" class="wide">

There's a scene in 1995's excellent _Hackers_ where a computer virus ("Da Vinci") takes control of an oil tanker, threatening to capsize it unless ransom is paid. An executive insists that the ship ballasts be put back under manual control to neutralize the threat. Eugene "The Plague" responds:

> There's no such thing anymore, Duke. These ships are totally computerized. They rely on satellite navigation, which links them to our network, and the virus, wherever they are in the world.

The first time I watched the movie, that idea was utterly ridiculous. Computerization, sure, but computerization to the exclusion of all else? No way. Technology can only penetrate so far into the existence of companies or people before they recognize the potential for abuse and fragility, and start to backpedal. These days, I'm not so sure. Da Vinci gets a little more believable every year.

---

I've benefited enormously from the rise of tech, but I don't want to see it total ascendent to the exclusion of everything else. Aside from the more obvious threat of monopolies controlling our minds and markets, I'm a traditionalist who still thinks that close human contact matters, and is essential for health and wellness.

In Japan there's a word for social recluses who isolate themselves from society, choosing instead of work and love a solo life of indulgence in internet, TV, and video games -- [_hikikomori_](https://en.wikipedia.org/wiki/Hikikomori). The west loves stories of Asia's eccentricities, and the typical reaction goes something like, "lolol, those crazy Japanese! what'll they do next. that'd never fly here." Every one of us being _hikikomori_ for all intents and purposes for the last year is an aberration, gone the second vaccines are distributed. 2022 will be akin the roaring 20s as we all get catharsis for two years worth of missed connections.

I'm not so sure. One of the reasons that lockdowns have been so amenable to so many isn't only the ostensible community care, it's that very often, their most ardent supporters haven't had to give much up. Well before Covid, even the young had been shifting away from bars, sports, and nightclubs to a more sedentary life of Snapchat, Netflix, and Overwatch. Busier working professionals with families have rarely ventured beyond the living room for entertainment in years. Now, everyone's at home (so no need to worry about FOMO), and it's a habit. This might just be the spark that was needed for the rise of a generation of _hikikomori_ in the west. Japan's culture isn't fundamentally different -- it's just ahead of the curve.

<img src="/photographs/nanoglyphs/017-twenty/river-2@2x.jpg" alt="The Bow River" class="wide">

---

## The new year (#new-year)

2020 was a bad year for a lot of people. I was one of the luckier ones who got through largely unscathed.

In retrospect, my biggest regret is not making the best of a bad situation. Normally, our most important, scarcest resource is time. The reason you can't pick up that new hobby, learn a language, or write that book is because there isn't enough of it. Last year, time was one of the few things we had left (well, in California at least) as commutes disappeared and the world shut down.

I made some progress on learning Japanese, but all in all, I spent too much time on video games and moping. (Although, _Final Fantasy VII Remake_ was pretty great. No regrets.) I've set a [few rough aspirational goals](/fragments/2021) for 2021, but the most important one: _make good use of time_.

There's a reasonable chance that five years from now we look back on 2020, and as awful it was, find that there are parts of it that we miss. The constant ferrying back and forth from home to office to gym to school might be back by then, and seem like abject wastefulness after seeing a world without it.

A few other points for 2021:

* **Do one thing:** Be distracted less from notifications and busywork and focus on one thing at a time. I've found iA Writer's full screen focus mode to be helpful with this so far.

* **Less screen time:** Spend less time with glowing rectangles, and _get them out of the bedroom_. So far, this one is not going well as I'm on-call and literally sleeping with a phone in bed with me, but there's still time to turn it around.

* **Write/tweet:** Write more, and yes, even tweet more. Make creative moments and content production a habit.

* **Reignite work:** I still fondly remember my early days in San Francisco back in 2012. Computing wasn't new, but modern SV-style cloud-first computing _was_, and we were still discovering the boundaries of this new world. It was incredibly fun, and there'd never be a day that went by where you didn't learn something new. By comparison, computing these days is refined, stable, and kind of boring, not all of which are bad things. Even so, find ways of bringing some of that old interestingness back to the art.

_Mongo delenda est_.

Until next time.
