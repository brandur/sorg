+++
published_at = 2019-11-05T23:00:26Z
title = "12 Factors and Entropy"
+++

![Folding keyboard](/assets/images/nanoglyphs/003-12-factors/folding-keyboard@2x.jpg)

Well, my weekly newsletter has become more like a monthly, but I’m not ready to give up on it quite yet.

My writing habits have been frozen recently, and I’m experimenting with a few radical techniques to thaw them. I’m writing this text on an iPhone and a folding Bluetooth keyboard (pictured above) under the theory that the glacial interaction speed of iOS will actually be an advantage when trying to achieve some multitasking-free focus.

That part’s turning out to be true, but the keyboard needs work. It’s got a solid feel and the world's slickest folding mechanism, but between cramped keys, a random 10% chance a depressed key doesn’t do anything, and occasional repeats, I’m retyping every second word. Navigating text takes so long that I’ve stopped fixing typos — write fast and fix later. I’ll get them on an edit pass. I look completely ridiculous, but it’s working.

TODO: Pictured above.

## 12 factor ascendant (#12-factor)

Google put together a guide for how to apply [12-factor to its cloud platform](https://cloud.google.com/solutions/twelve-factor-app-development-on-gcp), showing that despite its age (ancient by the standards of fast moving computing), it’s proven itself to be one of the most sticky ideas in the age of cloud computing.

The guide suggests the use of many Google-specific technoologies as answers to each factor, but that's okay given that the glue keeping everything together is open. For example, although a Kubernetes ConfigMap is recommended for configuration, it still injects environnmental variables, and that means that reuse or migration between Google’s managed Kubernetes and something else is still possible.

It’s easy to undervalue that. At Stripe we’re close to the opposite end of the spectrum in terms of generality, and everything from configuration to logging to booting a server requires infrastructure that’s been customized to the maximum possible degree. It works, but any kind of significant change turns into thousands of hours of engineering work.

## Spooky effects at a distance (#getrandom)

[Fixing `getrandom()`](https://lwn.net/Articles/800509/) is a great piece that opens a tiny window onto the world of kernel development, and what's inside is fascinating. The saga starts with the addition of `getrandom()` to the kernel after the LibreSSL project successfully advocated that it was too difficult to reliably get secure entropy (when good security practice is difficult, insecurity is likely).

It worked well until a bug was reported whereby computers hang on boot which turned out to be caused by the X Window System blocking on a call to `getrandom()` because not enough entropy was available. That reduction turned out to be caused by an [optimization in ext4](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/commit/?id=b03755ad6f33b7b8cd7312a3596a2dbf496de6e7) that minimized disk I/O on startup. Breaking changes don’t get much more roundabout than this — a bug-free optimization to a file system had the downstream effect of breaking boot for a graphical UI. You can’t help admire the intrepid bug hunter who figured that out.

## The jitter fix (#getrandom-jitter)

A [follow up of the last article](https://lwn.net/Articles/802360/) ("_Really_ fixing `getrandom()`") goes into how the problem was eventually patched. A number of solutions were put forward, like making `getrandom()` block until it was able to return sufficient entropy, and/or having it return an error in case it couldn’t. There were major concerns about both approaches though — Linus felt that blocking by default is wrong, and returning an error was akin to punting the problem downstream, where it wasn't likely to be handled well.

The fix that was eventually committed falls back to "jitter" entropy in cases where not enough “normal” entropy is available. Jitter entropy relies on the fact that executing instructions on a modern CPU is sufficiently complex that the precise time takes to do so is inherently unpredictable. Although not considered a particularly good form of entropy, its assumptions don’t have any known flaws either, so it was enough as a fallback option for the kernel. It has a major advantage of staying quite fast, so even in case a program has to block while jitter entropy is being generated, it won’t have to block for long (~1s for 128 bits on a system with a 100 Hz tick).

---

I don’t review many things, but this week I wrote one for Apple’s new [AirPods Pro](/fragments/airpods-pro). I can’t decide these days whether Apple’s direction is good or bad. The Touch Bar seems to indicate a pretty clear no, but some decisions, like their apparent reversal in position on the failure-prone butterfly key mechanism, are more positive. Most of what they do is in the grey: a welcome-until-you-see-the-price MacBook Air refresh, discontinuation of the single-port-but-ultra-streamlined 12” MacBook, a new pro computer and monitor, but priced so far into the stratosphere that you want eight figures of net worth before considering one.

AirPods are their one product line that’s nothing but a shining light. Like the original iPhone, they were technology so refined that it wasn’t clear that it was even within human reach until the day they were released. The originals were close to perfect except for their poor sound isolation (useless on the street, public transport, etc.) and “tap” interface. The Pros fix both those problems, and for my money, there’s never been a better pair of earphones.

---

I still haven’t advertised this newsletter broadly, and while I’m experimenting with it, it has a circulation of roughly one person (me), but if you like prototypes, you can sign up for it early [here](http://nanoglyph-signup.herokuapp.com). If I keep going you’ll keep getting them, there shouldn't be any other ill effects.

See you next week.
