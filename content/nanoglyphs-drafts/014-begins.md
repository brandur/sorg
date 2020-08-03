+++
image_alt = "TODO"
image_url = "/assets/images/nanoglyphs/014-begins/slope@2x.jpg"
published_at = 2020-04-21T05:11:14Z
title = "Remote; the Origins of Postgres"
+++

Readers -- it's been a while. I've been finding lockdown to be more antithetical for creative work than expected -- you'd think knocking two hours a day off commute alone would be an incredible productivity boon just by itself, not even to mention all the time saved through reduced meetings, fewer social obligations, and a less frantic lifestyle, but so it is. Despite some improvements, there's a lot counterbalancing the equation -- the existential contemplation that we're all doing these days being an obvious element, but the simple acts of talking to new people and seeing new places are a big piece of a successful creative/productive brew, and those are in short supply in our cloistered new bubble worlds.

I'm trying new techniques -- get up earlier, go for a lot of walks, write in the morning, write outside (I'm writing this from a now-decommissioned tram station in Dolores Park). An iPad + wireless keyboard has been a winning combination for minimizing the cognitive overhead of multi-tasking, stepping out of the office (figuratively speaking -- my laptop is my office these days), and because its screen is one of the few I own that's legible in daylight.

It's been so long that I'll reintroduce myself. I'm Brandur. This is _Nanoglyph_, a newsletter on software and adjacent practices which you hopefully signed up for, but maybe months ago. If you're dubious of that, you can always [unsubscribe in one click](%unsubscribe_url%).

---

My company's now been 100% remote since March, a total 180 from being ~100% non-remote previously. If nothing else, that's been an experience so unprecedented that I can viscerally feel that I'm living history. If we've ever spoken in person about remote, I probably came off as undecided-and-leaning-slightly-bearish on the idea in general. My last company was split down the middle between local and remote, and it wasn't a clear win -- it seemed to me at the time that local employees were more engaged and pulled heavier loads. There's also some practical truths (often discounted for the purposes of blog posts) like that major overlap in time zones is important for collaboration, and that despite advancements in the technology over the years, IRL meetings are still much more fluid and more functional than VC.

So naturally, I was a little worried about what might happen in an all remote world. And while there was some thrashing during those first few weeks of transition, we've slingshotted back to the point where there's little productivity difference compared to when we were all in an office together. People are online and engaged, still very much shipping code, and if anything, the company is getting _more_ hours out of them than they did before (maybe not a good thing). VC is still worse than an IRL meeting, but it is much better when _everyone_ is on VC.

Like everyone, I'm very curious to see where the world of knowledge working lands with remote in the coming years. Notably, some very big tech companies like Shopify, Twitter, and Square have formally adopted remote policies, but some allowance must be made for PR opportunism in such grand pronouncements. We'll have a better idea when we can see what these companies actually do, compared to what they said, when there's a real alternative again.

---

## Postgres begins

Last week I read the excellent paper [Looking Back at Postgres](https://arxiv.org/abs/1901.01973), which touches on the early ideas that drove its creation. Enthusiasts may know already that the project was sparked by Michael Stonebreaker while he was UC Berkeley in the 80s. The name is a derivation of “Post-Ingres”, referring to [Ingres](https://en.wikipedia.org/wiki/Ingres_(database)), an earlier database that also came out of UC Berkeley, and was later commercialized.

Object-oriented programming was the hottest of hot ideas at the time, and Postgres’ _object-relational_ model was to a large extent a reflection of that into a database, with a big part of that being the invention of ADTs.

### ADTs

ADTs (abstract data types) were an important feature that was introduced early and evolved over the years.  One of the original intentions was to be able to use them to store complex user-space objects, once again in that vein of OO. Nowadays this is more or less a dead idea, but has re-emerged over the years in new forms like document-oriented databases, and is still done in Postgres, but much more commonly with built-in data types like `jsonb`.

Still, ADTs turned out to be hugely important for implementing new built-in types -- everything from the various string classes to `jsonb` to `inet`/`cidr`. The extensibility in Postgres' index types helped here as well. The B-tree by itself can already be applied to practically anything, but where it can't (say a two-dimensional lookup), the GiST (generalized search tree) steps in to fill the gap.

### Active databases and rule systems

Early development effort was poured into "rule" systems, which executed code that broadly involved executing logic inside the database. Possibilities at the time were rewriting a query to perform an additional step, or monitoring for predefined conditions and responding with an action as they appear.

This work, along with similar initiatives from IBM and MCC, eventually led to the creation of triggers in the SQL standard, and are a common feature in many systems today.

All these years in retrospect, like OO, the idea in general isn't an unambiguous win. Although still in common use, practical use of triggers tends to lead to opaque and unexpected side effects, especially when chained together. Many of us tend to use them very minimally.

### Log-centric storage and recovery

This one is fascinating. Although today we're all used to data in journaled logs (like the <acronym title="Write Ahead Log">WAL</acronym>), Postgres was originally aiming for a much more ambitious moonshot in that data would be stored _solely_ in a log. Along with the benefits of a WAL, it would also unlock the potential of arbitrary time travel back to any database state by having the database pick a previous state and apply changes forward until the desired moment.

Not-too-surprising performance concerns eventually led Postgres to adopt a more convention model of storage manager combined with WAL, but it's fascinating to think about just how sticky this idea is, and continues to be. It's very similar "event-sourcing", which will crop up at any software shop given enough time, and I have to admit that [I've even written about it](/redis-streams) before.

---

This tiny summary does little justice to the detailed content of the full paper, so again, it's certainly worth a read. With a history of only ~50 years, software is newer than just about any other field, but it's amazing how rich its origins are already. Even more notable, many of the progenitors of the most important projects in the field are still with us today, which is a unique privilege for those of us active in it today.

---

In an effort to do something at least minimally constructive, I booted off a new [Sequences project](/sequences/2020-light) consisting of local shots taken around San Francisco in 2020. It's all starting to blend into the background because I'm seeing it again and again, but some of what you can see around here right now is practically unprecedented -- for example seeing the Great Highway closed and Ocean Beach's sands slowly reclaiming it. I've dropped the pretense that I'm posting once a day, but I do as often as I remember to do so.

By keeping my little photography projects cordoned to a tiny orbiting satellite like mine on the edge of that vast expanse that is the web, I'm reducing the impressions it'll get probably by 99%, but in this day and age that strikes me as the right decision. I enjoy Facebook and Twitter as much as anyone, but we as a society need to start having some serious conversations about them, and our own susceptibility to their emergent effects.

![Apple's iMac Pro](/assets/images/nanoglyphs/014-begins/great-highway@2x.jpg)

The solution? I wish it were something as simple as just "the indy web", but it's pretty clear that it takes the dead simplicity and network effects of a centralized platform to really bring the masses, and I'm not sure what to do about that.

Have a great week.

