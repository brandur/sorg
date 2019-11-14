+++
published_at = 2019-11-05T23:00:26Z
title = "12 Factors and Entropy"
+++

![Folding keyboard](/assets/images/nanoglyphs/003-12-factors/folding-keyboard@2x.jpg)

Well, my weekly newsletter has become more like a monthly, but I’m not ready to give up on it quite yet. As a reminder, this is _Nanoglyph_, a (supposedly) weekly newsletter on software.

My writing habits have been somewhat frozen recently -- not so much a writer's block, but a problem whereby every attempt at a directed task on a computer morphs into something else -- email, code review, YouTube. To thaw them, I’ve gone radical. I’m writing this from an iPhone and folding Bluetooth keyboard (pictured above) under the theory that the glacial interaction speed of iOS will be conducive for some multitasking-free focus. When slow input (touch) and long animation makes transitioning between apps is a 1-2 second operation, the platform, intentionally or not, inherently discourages the habit.

The keyboard has a solid feel and the world's most satisfying folding mechanism, but between cramped keys, a random 10% a depressed button does nothing, and occasional repeats, every second word needs retyping. Navigating text takes so long that I stopped fixing typos. Write fast, fix later -- I’ll get them on an edit pass, which is a good habit anyway. It has a few problems, and using it in public looks completely ridiculous, but the improved focus is working.

## 12 factor ascendant (#12-factor)

Google put together a guide for how to apply [12-factor to its cloud platform](https://cloud.google.com/solutions/twelve-factor-app-development-on-gcp), showing that despite its age (ancient by the standards of the fast moving discipline of computing), it’s proven itself to be one of the most sticky ideas in the age of cloud technology.

The guide suggests the use of many Google-specific products as answers to each factor, but that's okay given that the glue keeping everything together is open. For example, although a Kubernetes ConfigMap is recommended for configuration, it still injects environmental variables, and that means that reuse or migration between Google’s managed Kubernetes and something else is not only possible, but quite approachable.

It’s easy to undervalue that. At Stripe we’re close to the opposite end of the spectrum in terms of generality, and everything from configuration to logging to booting a server requires infrastructure that’s been so heavily customized that it'd be useless outside of our unique [Galapagos environment](/aws-islands). It works, but any significant change _starts_ at thousands of hours of engineering work, and goes up by orders of magnitude from there.

## Spooky effects at a distance (#getrandom)

[Fixing `getrandom()`](https://lwn.net/Articles/800509/) is a great piece that opens a tiny window onto the world of kernel development, and what's inside is fascinating. The saga starts with the addition of `getrandom()` to the kernel after the LibreSSL project successfully advocated that it was too difficult to reliably get secure entropy (when good security practice is difficult, insecurity is likely).

It worked well until a bug was reported whereby computers hang on boot which turned out to be caused by the X Window System blocking on a call to `getrandom()` because not enough entropy was available. That reduction turned out to be caused by an [optimization in ext4](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/commit/?id=b03755ad6f33b7b8cd7312a3596a2dbf496de6e7) that minimized disk I/O on startup. Breaking changes don’t get much more roundabout than this — a bug-free optimization to a file system had the downstream effect of breaking boot for a graphical UI. You can’t help admire the intrepid bug hunter who figured that out.

## The jitter fix (#getrandom-jitter)

A [follow up of the last article](https://lwn.net/Articles/802360/) ("_Really_ fixing `getrandom()`") goes into how the problem was eventually patched. A number of solutions were put forward, like making `getrandom()` block until it was able to return sufficient entropy, and/or having it return an error in case it couldn’t, but there were major concerns with each approach. Linus felt that blocking by default is wrong, and returning an error was akin to punting the problem downstream, where it wasn't likely to be handled well.

The fix that was eventually committed falls back to "jitter" entropy in cases where not enough “normal” entropy is available. Jitter entropy relies on the fact that executing instructions on a modern CPU is sufficiently complex that the precise time takes to do so is inherently unpredictable. Although not considered particularly good entropy, it doesn't have any known flaws either, so it was enough as a fallback option for the kernel. It has a major advantage of being quite fast, so even in case a program has to block while jitter entropy is being generated, it won’t have to do so for long (~1s for 128 bits on a system with a 100 Hz tick).

---

I don’t review many things, but this week I wrote one for Apple’s new [AirPods Pro](/fragments/airpods-pro). Usually, I can’t decide whether Apple’s direction these days is good or bad. The Touch Bar spells a clear "bad", but some decisions, like their return to a scissor keyboard mechanism [1] in the new 16-inch MacBook Pro, are more positive. Most of what they do is grey: a welcome-until-you-see-the-price MacBook Air refresh, discontinuation of the single-port-but-kind-of-great 12” MacBook, a [new pro computer and monitor](https://www.apple.com/mac-pro/), but priced so far into the stratosphere [2] that you want eight figures of net worth before thinking about one.

AirPods are the one product line that’s been nothing but a shining light. Like the original iPhone, they were technology so refined that it wasn’t clear it was even within human reach until the day they were released. The first iteration was close to perfect but for poor sound isolation (useless on the street, public transport, etc.) and underwhelming "tap" interface. The Pros fix both those problems, and for my money, there’s never been a better pair of earphones.

---

I haven’t advertised this newsletter broadly, and while I’m experimenting with it, it has a circulation of roughly one person (me), but if you like prototypes, you can sign up for it early [here](http://nanoglyph-signup.herokuapp.com). If I keep going, you’ll keep getting them. If I don't, there shouldn't be any ill effects.

See you next week.

[1] And physical escape key! And T-shaped arrow keys! Apple's been spoiling us recently with the reintroduction of 50 year old keyboard technology.

[2] Starting at $6k and $5k for computer and monitor respectively, but for the bottom of the range configurations that no one is actually supposed to order, so expect to spend far more.
