+++
image_alt = "Bunaken"
image_url = "/assets/images/nanoglyphs/014-local-first/bunaken@2x.jpg"
published_at = 2020-09-11T15:12:28Z
title = "Local First"
+++

Last year, my family went on a dive trip to Indonesia. It was magnificent. Picture perfect beauty in any direction you care to look, and doubly so underwater. At a rate of two to three dives per day (I'm the group slacker by usually doing only two, while everyone else opted to triple down), we'd see at least a few of the more exotic sights like octopus, cuttlefish, or mantis shrimp every day. One of the highlights of the trip was when a pod of dolphins swam up alongside our moving dive boat, playing in its wake for long minutes before breaking off -- the kind of thing of that's only supposed to happen in movies [1].

(You're reading Nanoglyph, a newsletter on software and local first principles. If you don't want to get it anymore, [unsubscribe right here](%unsubscribe_url%) to banish it forever.)

---

Speaking of octopuses, I'd highly recommend [this piece](https://www.lrb.co.uk/the-paper/v39/n17/amia-srinivasan/the-sucker-the-sucker) ("_The Sucker, the Sucker!_") from the _London Review of Books_. Not only are these creatures mesmerizing, but the review's prose is immaculate. An excerpt:

> The octopus threatens boundaries. Its body, a boneless mass of soft tissue, has no fixed shape. Even large octopuses – the largest species, the Giant Pacific, has an arm span of more than six metres and weighs a hundred pounds – can fit through an opening an inch wide, or about the size of its eye. This, combined with their considerable strength – a mature male Giant Pacific can lift thirty pounds with each of its 1600 suckers – means that octopuses are difficult to keep in captivity. Many octopuses have escaped their aquarium tanks through small holes; some have been known to lift the lid of their tank, making their way, sometimes across stretches of dry floor, to a neighbouring tank for a snack, or to the nearest drain, and maybe from there back home to the sea.

Octopus spotting during a dive is one of the most reliably consistent forms of delight in life. Another diver grabs your attention and points to a rock. As a rock, it's unimpressive, and your water-addled brain idly wonders why they're wasting your time with it. Suddenly, the rock _shifts_ in a way that, in your experience, is not very characteristic of rocks. You stare harder. Ever so slowly, your mind resolves some patterns that rocks don't normally have. Are those ... eyes? Do rocks have eyes? Not that you recall. Zoom out. More features emerge. An outline of something. Wait, that's no rock, it's a ... ! The other details slam into place in an instant and you finally see the thing. One of the most peculiar-looking creatures on Earth, yet practically invisible.

---

## Local only (#local-only)

One of the Indonesia's downsides that we railed against for a long time before finally becoming accustomed to, was its internet access. The dive resorts we visited claimed to offer it, but it was a web that was life worse than death. We'd frustrate ourselves endlessly trying in vain to visit our usual online fair, watching the progress bar inch across the screen before unceremoniously terminating in failure, and trying the reload button ten times before abandoning the project in frustration. My dad would manage to send a daily email blast to the outside world with a few down-rezed images, but only by waking up around 6 AM and getting ahead of the game. One person might have luck, but two or more brought the system to a grinding halt so complete that no one got through.

We probably could have pre-empted the problem by acquiring some LTE SIM cards in advance -- it's the type of country that skipped hard lines and went straight to wireless infrastructure, but we were staying on tiny islands where big retail amounted to window stands selling potato chips and cigarettes. The better answer was to log off for a few weeks and just ... relax. The first few days were excruciating. We were recent ex-smokers, reaching for a pack for the tenth time that day only to find it wasn't there. But once it was done, we were living a rare from of bliss available to very few in the modern world. Dive. Eat. Read.

---

To my shame, I brought a laptop and would occasionally use it for referencing notes, importing photos, and doing a (very) nominal amount of writing. I was pleased to find that despite the deranged wi-fi, all my common workflows were still available. This was largely thanks to two of my favorite things: a file system, and Dropbox. Practically every note, photo, or media is up there in the cloud, but it's _also_ right there with me on my hard drive. Even with Precambrian internet access, I could still read or reference most of what I cared to.

This was in sharp contrast to my iPhone, where suddenly not a thing was working. No music. No notes. No files. No video. Even something as simple as that article you'd gotten halfway through reading and wanted to finish wasn't reliably available, as mobile Safari made the executive decision to evict it from memory, keeping only a screenshot and the URL, and incorrectly assuming that a functional internet connection for paging it back in would be available. Dropbox, my old reliable friend on the PC, was worthless in app form. Although allowed access to all the same files, it was crippled by an incurable cloud dependency.

![Coral Eye's dive jetty on Bangka in Indonesia](/assets/images/nanoglyphs/014-local-first/coral-eye@2x.jpg)

---

## Local first ideals and CRDTs (#ideals)

About a year ago, the research lab Ink & Switch published an excellent treatise on a principle they've coined as ["local first"](https://www.inkandswitch.com/local-first.html). It's well worth the read in its entirety, but I'll outline their seven ideals for local first software:

1. _No spinners: your work at your fingertips_: Make software faster by not needing to load from the cloud on every action. Data lives locally and is accessed quickly.
2. _Your work is not trapped on one device_: Local first doesn't imply no inter-connection. Software should still have a synchronization mechanism that allows it to be at your fingertips anywhere.
3. _The network is optional_: Software still works without an internet connection. See my story from Indonesia above.
4. _Seamless collaboration with your colleagues_: Cloud-like collaboration of the sort you'd see in a Google Doc should still be possible without irreconcilable conflicts or other ugly side effects.
5. _The long now_: Keeping data locally improves its longevity as cloud services come and go.
6. _Security and privacy by default_: Keeping data locally makes it harder to access by attackers, and is not as prone to privacy violations by rogue employees who have access to your data in their cloud service.
7. _You retain ultimate ownership and control_: Cloud services have been known to lose or lock people out of their own data. When it's local, you own it, although that might come with other responsibilities like maintaining adequate backups.

The piece then goes on to examine how well various existing models comply to these principles. For example Git/GitHub checks most of the boxes like speed (_ideal 1_) and offline access (_ideal 3_), but misses multi-device (_ideal 2_ -- Git repos only exist in specific places), collaboration (_ideal 4_ -- Git collaboration is coarse-grain by way of patches/pull requests, but fine-grain collaboration is more difficult), and privacy (_ideal 6_ -- generally, origin repositories are in a central cloud like GitHub's).

They then go on to suggest <abbr title="Conflict-free Replicated Data Types">CRDTs</abbr> as a possible foundation for future local first technology, and show how they've tried to flesh out what that might look like with various prototypes.

---

## The right abstraction (#abstraction)

This article must have hit the front page of Hacker News half a dozen times in the last year, so it obviously strikes a nerve for the hacker types of the world. And even if you showed Ink & Switch's list of principles to a non-technical member of the general public, they'd probably agree that they'd want most of them. So why is there so little local first software in the world?

My theory is that the big blocker to a smarter local technology is the same as the reason we've largely lost the independent web: a deep-rooted human desire for convenience. People are lazy, and although they like ideas like longevity and privacy in the abstract, we're willing to sacrifice them for systems that make things easy. Big cloud platforms are more than happy to take advantage of this defect in human mentality because it's a win-win from their perspective -- it's better for them if they own your data, so unless users are _demanding_ an alternative at deafening volume, they'd never go about it any other way.

This problem _might_ be addressed by an abstraction that fulfills local first ideals while also being amenable to cloud providers by being easily pluggable, and not overly compromising their services' ease-of-use (or monetization models). Personally, I think CRDTs are too low-level to do this job. The closest we've ever come is the file system, and unfortunately, Apple and Google are on a determined mission to kill those dead (or at least as far as the end user is concerned; they're more happy for the internals of their operating systems to reap their benefits). A replacement wouldn't only have to solve distributed synchronization, but it'd also have to offer better sandboxing and user security to appease the giants.

The next step would be to make sure users understand how to use such abstraction to their own advantage, and understand _why_ they're doing it. Remember how in the early days of personal computing it was common to have training courses in IT? Working with files, Excel, or Word were all considered complicated enough that not everyone would be able to intuit their way to success. Using them would unlock powerful features, but there was a learning curve involved.

Fast forward to today and the programs we have are easy to use, but we've overcorrected. I've made this argument before in places like [use of Markdown](/fragments/slack-bar-raising). We should be teaching the world useful-but-non-trivial skills for self expression instead of dumbing it down to the lowest common denominator, putting it at the mercy of grotesque machinations like <abbr title="What You See Is What You Get">WYSIWYG</abbr> and Confluence. This would be a similar project in educating users why controlling their own data is good for them, and how to go about keeping it secure. There would be an inherent security trade off -- information is less likely to be leaked when it never leaves Google/Facebook/Instagram's servers -- but the benefits outweigh the costs.

Changing course to a world based on local first principles would take immense activation energy. Even with a perfect design, momentum in the direction of the cloud is close to inalterable. Our best hope at this point would probably be to have it show significant success in the microcosm of some popular piece of niche software, which could then export its foundations to a broader audience.

---

![Page from Sandman](/assets/images/nanoglyphs/014-local-first/sandman@2x.jpg)

I spent the last couple weeks reading the entirety of the 75-issue run of Neil Gaiman's original Sandman series. The last time I read it (well over a decade ago) I remember liking it, but this time around it was sublime, with a compelling creativity and narrative design rarely at a level rarely seen in the comic medium. The story follows Dream, king of the Dreaming and one of the Endless along with his siblings: Destiny, Death, Destruction, Desire, Despair, and Delirium. After being accidentally imprisoned by a sorcerer seeking to harness death, he finally escapes, and must set about rebuilding his dilapidated realm.

Just to give you a taste of its overwhelming imagination: in the first few issues Dream visits the inner circles of hell itself, controls Lucifer, and must challenge a demon to a contest of cosmic poetry to recover a stolen item.

The idea for a re-read was sparked after listening to Audible's [original 2020 production of the Sandman](https://www.audible.com/pd/The-Sandman-Audiobook/B086WP794Z), narrated by Gaiman himself. It's also excellent, and faithful to the source material, in case audio is more your speed.

See you around.

[1] Our dolphin encounter might have been particularly good luck as even the dive masters and boat crew, who have taken tourists on literally thousands of expeditions, were into it.
