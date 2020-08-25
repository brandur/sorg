+++
image_alt = "A steep San Francisco street"
image_url = "/assets/images/nanoglyphs/013-remote/slope@2x.jpg"
published_at = 2020-08-11T17:15:39Z
title = "Remote; the Origins of Postgres"
+++

Readers -- it's been a while. I don't know about you, but I've been finding lockdown to be more antithetical for creative work than expected. The prospect of knocking two hours a day off commute, reduced meeting load, fewer in-person obligations, and a less frantic lifestyle seemed attractive in that it might give a disciplined person more time for side projects and hobbies. Maybe it does, but I wouldn't know, because I'm not that disciplined person. I've realized the hard way that simple acts of talking to new people and seeing new places are a big piece of a successful creative/productive brew, and those are in short supply in our cloistered new bubble worlds.

I'm trying some fixes -- get up earlier, go for a lot of walks, write in the morning, write outside (I'm writing this from a now-decommissioned tram station in Dolores Park). An iPad + wireless keyboard has been a winning combination for minimizing the cognitive overhead of multi-tasking, stepping out of the office (figuratively speaking -- my laptop is my office these days), and because its screen is one of the few I own that's legible in daylight.

It's been so long that I'll reintroduce myself. I'm Brandur. This is _Nanoglyph_, a newsletter on software and adjacent practices which you hopefully signed up for, but maybe months ago. If you're dubious of that, you can [unsubscribe in one click](%unsubscribe_url%).

---

I delayed even longer in sending this issue in an effort to convince Google that I'm not spam. Instead of using one of the major newsletter platforms I've jury-rigged my own made up of Go bolts duct taped to Mailgun. It's fun and cheap, but was probably mistake -- any sender is already on a knife's edge with the Great Spam Filter, and best practice is to never step out of line. I wrote about [enabling DMARC](/fragments/dmarc) (a standard to help validate an email's origin) in the hopes of assuaging Google, which also touches on related concepts like SPF and DKIM. Like a lot of technology, email standards are a veritable acronym soup, but not hard to understand once through the tough exterior.

---

## On being remote (#being-remote)

My company's now been 100% remote since March, a 180 degree turn from being ~99% non-remote previously. The feeling of living history is visceral. If we've ever spoken about remote before now, I might have come off as undecided-and-leaning-slightly-bearish on the concept in general. My last company was split down the middle between local and remote, and it wasn't a clear win. Opinions were very mixed, but it seemed to me at the time that local employees were more engaged and pulled heavier loads. There's also some practical truths (often discounted for the purposes of blog posts) like that major overlap in time zones is important for collaboration, and that despite advancements in the technology over the years, VC still has lightyears to go before it reaches the fluidity of a real life meeting.

So naturally, I was worried about what might happen in an all remote world, but that fear has turned out to be largely unfounded. After a little thrashing during those first few weeks of transition, we've sling shotted back to the point where there seems to be little productivity difference compared to when we were all in an office together. People are online and engaged, still very much shipping code, and if anything, the company is getting _more_ hours out of them than they did before (maybe not a good thing). VC is still worse than a real life meeting, but it is much better when _everyone_ is on VC. The biggest downside is the social aspect -- really getting to know people and making new friends is much harder over text conversations, or inside scheduled blocks.

(Like everyone else on Earth) I'm very curious to see where knowledge working lands with respect to remote in the coming years. Notably, some very big tech companies like Shopify, Twitter, and Square have formally adopted remote policies, but some allowance must be made for PR opportunism in those sorts of grand pronouncements. We'll have a better idea when we can see what these companies actually do, compared to what they've said, when there's a real alternative again.

---

## Postgres begins

A few weeks back I read the excellent paper [Looking Back at Postgres](https://arxiv.org/abs/1901.01973), which touches on the early ideas that drove its creation. Enthusiasts may know already that the project was sparked by Michael Stonebreaker while he was UC Berkeley in the 80s. The name is a derivation of “Post-Ingres”, referring to [Ingres](https://en.wikipedia.org/wiki/Ingres_(database)), an earlier database that also came out of UC Berkeley, and was later commercialized.

Object-oriented (OO) programming was the hottest of hot ideas at the time, and Postgres’ _object-relational_ model was to a large extent a reflection of that into a database, with a big part of that being the invention of ADTs.

### Pluggable types

ADTs (abstract data types) were an important feature that was introduced early and evolved over the years.  One of the original intentions was to be able to use them to store complex user-space objects, once again in that vein of OO. Nowadays this is more or less a dead idea, but has re-emerged over the years in new forms like document-oriented databases, and is still done in Postgres, but much more commonly by serializing from application space to built-in data types like `jsonb`.

Still, ADTs turned out to be hugely important for implementing new built-in types -- everything from the various string classes to `timestamp` to `jsonb`. The extensibility in Postgres' index types helped here as well. The B-tree by itself can already be applied to practically anything, but where it can't (say a two-dimensional lookup that you'd need for PostGIS' spatial indexing), the GiST (generalized search tree) steps in to fill the gap.

### Active databases and rule systems

Early development effort was poured into "rule" systems, which executed code that broadly involved executing logic inside the database. Possibilities at the time were rewriting a query to perform an additional step, or monitoring for predefined conditions and responding with an action as they appear.

This work, along with similar initiatives from IBM and MCC, eventually led to the creation of triggers in the SQL standard, and are a common feature today.

All these years in retrospect, like OO, the idea in general isn't an unambiguous win. Although still in common use, practical use of triggers tends to lead to opaque and unexpected side effects. That applies doubly once they start being chained together, and you have triggers firing triggers firing triggers. Many of us tend to use them very conservatively.

### Log-centric storage and recovery

This one is fascinating. Although today we're all used to data in journaled logs (like the <acronym title="Write Ahead Log">WAL</acronym>), Postgres was originally aiming for a much more ambitious moonshot in that data would be stored _solely_ in a log (think, no fully materialized tables on disk). Along with the benefits of a WAL, it would also unlock the potential of arbitrary time travel back to any database state by having the database pick a previous state and apply changes forward until the desired moment.

Not-too-surprising performance concerns eventually led Postgres to adopt a more convention model of storage manager combined with WAL, but this was a sticky idea, and continues to be. It's very similar "event-sourcing", which will crop up at any software shop given enough time, and I fully admit to [having previously written about that](/redis-streams).

---

This tiny summary does little justice to the detailed content of the paper, so again, it's certainly worth a read. With a history of only ~50 years, software is newer than just about any other field, but it's amazing how rich its origins are already. Even more notable, many of the progenitors of the most important projects in the field are still with us today, which is a unique privilege for those of us active in it today.

---

In an effort to do something at least minimally constructive, I booted off a new [Sequences project](/sequences/2020-light) consisting of local shots taken around San Francisco in 2020. It's all starting to blend into the background because I'm seeing it again and again, but some of what you can see around here right now is practically unprecedented -- for example seeing the Great Highway closed and Ocean Beach's sands slowly reclaiming it. I've dropped the pretense that I'm posting once a day, but I do as often as I remember to.

By keeping my little photography projects cordoned to a tiny orbiting satellite like mine on the edge of that vast expanse that is the web, I'm reducing the impressions it'll get probably by 99%, but ... I'm okay with that. I enjoy some dabbling in Facebook and Twitter as much as anyone, and can't help but feel that little rush when someone likes my post, but we as a society need to start having some serious conversations about them, or at least our susceptibility to their emergent effects. On my optimistic days, I hope that in ten years we'll see social media as junk food for the mind -- viscerally satisfying on a primitive level thanks to how our minds and bodies have evolved over thousands of years, but indulgent and unhealthy.

![The Great Highway, car free](/assets/images/nanoglyphs/013-remote/great-highway@2x.jpg)

The problem with _not_ being on social media is that we still don't have great mechanisms to discover and connect content outside of it. Links work, but linking from the same sites you visit every day doesn't add a lot of fresh content to the mix. RSS and newsletter are good, but also could use much better discoverability. Link aggregators like Hacker News are still my most oft-used solution, but need incredible amounts moderation as their size makes them attractive targets for spam and abuse.

The network effects of centralized platforms gives them a tremendous advantage over the web's original decentralized model, not to mention their sheer convenience, and it's hard for the independent web to compete. Maybe that's okay, and we should be interacting in smaller circles rather than broadcasting to the entire world. Something to think about anyway.

Have a great week.

