+++
image_alt = "Yunomine Onsen"
image_url = "/photographs/nanoglyphs/022-entropy/yunomine@2x.jpg"
published_at = 2021-03-07T00:09:32Z
title = "Time and Entropy"
+++

A few weeks back, looking at some photos of old vacations and feeling nostalgic, I started looking up some blogs written by friends of mine years ago as they were making their own journeys to various places around the world.

I was disappointed to find every one of them, without exception, now gone. The vanishings were for different reasons -- a defunct blogging service, a deleted account, or just an expired domain -- but the result was the same. Entropy had taken its toll, and these little highlights of history were scrubbed from time.

The Internet Archive's [Wayback Machine](https://archive.org/web/) had saved some of them from total extinction, but only for those whose URL I could remember, and even for those, there were many holes -- pages and images that their crawler hadn't indexed for whatever reason.

There's an argument to be made that maybe this is okay. Content is created, gets a brief moment to shine, and is forgotten, leaving a vacuum for the newly created. I buy this to some degree. There's something special about the ephemeral -- in some ways we value something more knowing that it's not going to last, but I still find it unsatisfying. In very few of these cases was the ephemeralism a conscious choice, and moreover, and many of these blogs are well representative of the earliest days of the internet, and historically interesting, especially as we continue our seemingly unstoppable trajectory towards the entire world's content slurped up into a few major platforms.

As usual, I'm Brandur, this is _Nanoglyph_. If you're tired of pseudo-philosophical ramblings about software, you can always unsubscribe in [one simple click](%unsubscribe_url%).

---

## Erosion resistance (#erosion-resistance)

One of the most inspirational ideas to come out of the old Heroku days was [erosion resistance](https://blog.heroku.com/the_new_heroku_4_erosion_resistance_explicit_contracts).

At the time, in a pre-container pre-serverless world, it was incredibly common to lose applications as something failed in their underlying OS:

* Log files filled up the disk.
* A system upgrade changed something and made the application unstartable.
* The server itself crashed and had no mechanism for coming back up cleanly or redundancy.

Heroku suggested solving the problem by minimizing the surface area of the contract between application and host, and making it explicit. This allowed the underlying system to be changed and upgraded while keeping it compatible with a running application. OSes would stay up-to-date and applications would stay running even if left unchanged for many years.

---

Time would show that the original idea of erosion resistance a little optimistic. Heroku's original `cedar` stack would eventually require an upgrade to `cedar-14`, and that was subsequently replaced by more regular stack upgrades in the form of `heroku-16`, `heroku-18`, and `heroku-20` (the numerals are the major version of Ubuntu each is tied to). Upgrades generally require user intervention as slugs were tied to system dependencies in subtle ways.

Still, the broad idea of erosion resistance is a useful one, and a principle that we can apply when building anything. How can I choose a programming language, tech stack, and database (or lack thereof) that maximizes the chances that what I'm building will be around ten years from now?

Here's a few ideas.

---

{{NanoglyphSignup .InEmail}}

### Markdown and Git (#markdown-and-git)

I [argued in a blog post a few weeks ago](/fragments/graceful-degradation-time) that whatever you choose as your blogging frontend, make sure that the sources are available in a simple, widely interoperable format like Markdown, and stored portably in a system like Git. Even after the domain hosting this newsletter is long gone, its sources should still [be available on GitHub](https://github.com/brandur/sorg/tree/master/content/nanoglyphs), and even if GitHub is gone by then, the source repository could easily be moved to somewhere else. I hope that it'll be readable decades from now.

Contrast this to a dynamite site which stores sources in a database which needs frequent upgrades to stay online, and renders with a dynamic language with many dependencies and whose  shifting ecosystem requires full time attention. A decade on, it's likely that archive.org is the only place where any of its content is still accessible.

### Think on business models (#business-models)

About ten years ago, a novel new writing platform called Medium experienced a sudden surge in popularity. With beautiful content layouts and a power editor, it was especially popular amongst the tech community in those early years, who abandoned their own blogs to rush to it in droves.

And for many years, it was quite respectable to write on Medium, but there was a problem. This was an investor-bootstrapped company employing many highly-priced software engineers and operating in the world's most expensive city. And even with occasional surges in traffic, blogging isn't traditionally known as a highly lucrative business. These two forces were fundamentally at odds to each other, and over the years Medium degraded in pursuit of a business model -- first with little nag messages and ads, and later with full-screen modals and paywalls. They'd taken their users' content and held it hostage, but they owned the platform, and it was well within their right to do so.

These days, to most of us in tech circles, it seems like the height ill-advised hubris to put new content on Medium, but it was different in those early days. It was a great platform at the time, and people were optimistic that it would only get better.

One way to be entropy resistant is to _think on business models_. The exact extent to which Medium went bad may not have been foreseeable, but the fact that there was a strong chance it'd get worse absolutely was. The same was true for a lot of the traditional platforms like Tumblr, Flickr, Blogger, and even as far back as Geocities. Although they've all survived in some form through to today, your content on those services was always at risk.

This is one reason that I'm not quite as enthusiastic as most people with the internet's current obsession with Substack. As I [wrote about last week](/nanoglyphs/021-ides#push-model), I wish only the best for them, and that they put a dent in our highly polarized media institutions, but at the end of the day, sending email is expensive. And least Substack has a business model, and there's no question that they're doing well off their top writers, but there _is_ a question as to how the platform treats its less popular writers over the long term.

### Posterous and Posthaven (#posterous-and-posthaven)

Years ago, some old friends from Calgary were about to quit their jobs and go travel the world. They wanted to keep a record of it, and asked me what blogging platform they should use.

To my everlasting chagrin, my younger self of the time told them to try Posterous, the product du jour, whose differentiating feature was being able to write posts write in your inbox, and publish them by email (which ironically, is an idea [that's just come back](https://hey.com/world/) for round two). Of course, a few years later, Posterous was acquired by Twitter and shut down shortly thereafter.

But I got lucky. The founders of Posterous felt strongly about keeping their users' content online, so created a _new_ platform, aptly named Posthaven. From [their pledge](https://posthaven.com/our_pledge):

> Need a blog? So did we. This one is made to last forever.
> 
> We need a good solid blog platform for ourselves, and we know our friends and colleagues need one too. That's why we started Posthaven.
>
> We're not going to show ads. We're not looking for investors. We're going to make money the best way we know how: charging for it.
>
> It is a simple, clean, well-lit place to keep your stuff.

Posthaven charges $5/month, which is more money than most people are used to paying for hosting content, but their payment model comes with an interesting quirk: after you've been paying for the service for a year, they'll keep your existing content around forever, even if you stop paying.

They provided an easy route for migrating content off the sinking Posterous, and because of that, my friends' blog is still online to this day, unlike many of its contemporaries. (See [Beyond the Maple Tree](https://beyondthemapletree.com/).)

Posthaven is still a very small team, so it remains to be seen how long their promise of "forever" will actually last, but I hope they make it decades further.

### One TLD, as generic as possible (#generic-tld)

A common reason for old sites to fall offline is that even if content still has a home, their owner stops paying for the domain name. At around $10/yr [1], domains aren't _that_ expensive, but they're still a non-trivial cost over many years, especially when people own more than one, and at some point they're not worth the upkeep.

Here's an idea for erosion resistance: don't buy a TLD for every new project. Get one top-level domain that's generic sounding and which you're reasonably happy with, and then nest future projects as subdomains under it. Subdomains are free, so instead of paying for five domains forever, you're just paying for one.

I blog under my own name at brandur.org, which works well enough for personal use, but putting subdomains under it that are not directly related to me would be strange. For contrast, I have an ex-colleague that blogged at xph.us, a domain that's short and not particularly suggestive of anything. You could nest anything under that and no one would blink an eye.

### Newsletters: "Dark" persistence (#dark-persistence)

A common thesis is that distributed systems are more robust. If you want information to last, put it on a blockchain. (I'm not so sure about that one myself, but you hear it a lot.)

Here's one last idea for a "dark" form of erosion resistance: write a newsletter. Even if the original content disappears from the web, there's a reasonable chance that it's recoverable as someone out there still has a copy of it in their mailbox.

I'm obsessive about correctly annotating and archiving the newsletters I get (each is sorted by rule into its own dedicated label), so if my friends' disappeared blogs had been newsletters instead of blogs, I'd still be able to read every word of them.

This is a dark form of persistence because it brings the content out of the public sphere, but it's not bad as a last resort.

---

<img src="/photographs/nanoglyphs/022-entropy/yunomine-2@2x.jpg" alt="Yunomine Onsen 2" class="wide">

## Yunomine (#yunomine)

Today's photos come from a trip to Japan a few years back where one of the places we visited along the Kumano Kodo was [Yunomine Onsen](https://www.tb-kumano.jp/en/places/yunomine/), a tiny hot spring village.

I've already written about it in [an old edition of _Passages & Glass_](/passages/003-koya#yunomine), but it's been a good year to revisit old memories. Nestled around a steaming creek where locals go to slow cook onsen eggs, the village features Japan's oldest hot spring (discovered 1,800 years ago), and a tiny two-person bathhouse called Tsuboyu. Everything was covered with thick mineral buildups, subtly beautiful in a unique way. Easily one of the best trips of my life.

Until next week.

[1] Although fluctuating much higher for specialty domains.
