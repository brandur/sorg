+++
image_alt = "Sunrise over Twin Peaks"
image_url = "/photographs/nanoglyphs/024-new-horizons/sunrise@2x.jpg"
published_at = 2021-04-14T04:15:09Z
title = "New Horizons"
+++

Last week, I quit my job. Even knowing it was the right move, I was still a little frazzled at the point of doing so -- it was the end of a long 5+ years of stability at a company that many would consider a rocket ship -- firing off to the stars draped in elegant green, white, and orange, perceived with superior business acumen and [moral compass](https://stripe.com/climate) than almost anyone else in modern cutthroat capitalism. Still, the time had come, and I felt good doing it.

I'm going back to my roots and joining Crunchy Data to work on Postgres and Postgres-adjacent product offerings. I wrote [more about that here](/fragments/crunchy), but the gist of it is that I want to get back to an industry that is personally of more interest to me (databases, and developer tooling more broadly), and Crunchy's a company that I've long admired for its industry leaders and supportive approach with respect to open source.

---

Before I get too far -- in case this is your first time reading _Nanoglyph_, I'm [Brandur](https://twitter.com/brandur). This newsletter is ostensibly themed around software simplicity and sustainability, although often isn't. I'll be talking more about Postgres and corporate tooling for the next few months.

Don't want it anymore? You can always quit in [one easy click](%unsubscribe_url%).

---

## Castles of sand (#castles-of-sand)

One of the big reasons that I'm so interested in tools like Postgres is that -- boiled right down to the core kernel of brutal truth -- software sucks, and we as an industry desperately need to find ways of making it better. Whenever I'm doing anything non-trivial in a web app like transferring money, configuring cloud resources, or even something as simple as responding to a comment, problems of varying sorts are an everyday occurrence. Most of them are minor -- having to refresh a page that's worked its way into a bad state, resubmitting a form that didn't quite go through, or clicking a link twice to _really_ go there the second time -- but most services don't work quite how they're supposed to at least some of the time.

There's a few main themes from which these problems stem. One is just the sheer number of moving parts in modern computing architectures. A single request might be routed through an edge CDN, terminated by a regional load balancer, dispatched to a worker node via Envoy, and locally reverse proxied through Nginx before finally being handled by an app that internalizes many layers of its own. Throw microservices into the mix, and you get a multiplicative effect on top of that already substantial setup. Every component, although added with the best of intentions, is another small opportunity for failure, and that failure is happening at low ambient levels around the clock.

But _another_ big reason that software isn't reliable is that we largely haven't figured out how to build it sustainably, yet. Your freshest faced intern knows in an academic sense to build abstractions in layers, and we do a good job of this in some places -- operating system I/O APIs, programming languages and their standard libraries, and the kernel interface, to name a few, but once we climb up into user space, it's vastly more common for programs to become monolithic balls of amorphous direction and quality. This is partly because of demands on resources and time, partly because most programming languages do a poor job of encouraging separation, and partly because we don't know what we're doing as well as we think we do.

I like to use the metaphor of a sand castle. Picture one made of sand moistened by the morning tide, but which has been drying under the sun for hours so that the sand sticks but is no longer _sticky_. Different sections of it are slowly crumbling apart as they dry out and lose cohesiveness. A couple builders, frantically trying to save their creation, rush around the periphery dampening select towers and ramparts -- carefully reforming them to keep the integrity of the whole structure. This is modern software development. The builders are engineers, the loosening parts are bugs, and the work never ends -- left too long without attention and the edifice regresses into a vaguely shaped mound in no time at all.

I think there's an opportunity to change this. By building carefully -- strong foundational layer on top of strong foundational layer -- it's possible for higher level layers to operate in confidence knowing that lower layers just _work_. Again, we already do a good job of this in some places -- I can almost always just assume that the CPU is executing my code correctly, that memory allocation is practically never going to fail, or that I/O operations will succeed modulo known error conditions, and yet most often, all bets are off once we hit user space.

I think that Postgres can be one of those strong foundations, one that most developers can build on to make higher quality software. I've written about this [somewhat previously](/acid), and intend to dig deeper into the subject.

---

<img src="/photographs/nanoglyphs/024-new-horizons/sunrise-tree@2x.jpg" alt="Sunrise towards downtown, through a tree" class="wide">

## Digital communication and Slack ascendency (#communication)

Okay, old man pontifications incoming. I've have the dubious privilege of working long enough in industry to have seen practically every iteration of the relationship between businesses and digital communication. It's only been a little over a decade, but quite a wild ride.

During my very first internship, email was the indisputable gold standard. Some of us had Google Hangouts, but its use at work wasn't formalized in any way, shape, or form. The preferred method for fast communication was to walk over and knock on someone's cubicle, although this was BlackBerry [1], so we weren't half bad at using email as a form of proto-IM via our shiny prototype [BB 8800s](https://www.gsmarena.com/blackberry_8800-1911.php) (these may as well have been made from chipped rocks and sharpened sticks by today's standards, but they were the best phone in the world at the time). We'd set up an unauthorized internal IRC server, but nothing of consequence happened there -- 90% of message activity was coordinating coffee runs to Tim Hortons.

This continued to be largely true for my next few jobs. By the time I was at iStockphoto, the IRC server was a little more formalized in that it was recommended that most people join it, but it was still by and large engineering only, with little participation from management or exec. You were often on it, but could get away without being in there too. By this time the cubicles were gone, but tapping people on the shoulder remained a preferred method for fast discussion.

Heroku was my first place that chat-as-a-service really made an appearance in a big way. We ran the whole gamut, starting with Basecamp before jumping ship to HipChat, before jumping ship yet again to Slack, which at the time was making market headway that few would have thought possible in the even-then crowded market of chat apps.

Slack was particularly notable for being truly delightful to use -- people loved it, and that showed through the amount of time they'd spend with the product. I remember it also as the first time that I saw the dark side of group messaging. Although fast and efficient, it was no longer an unambiguously good thing.

### Slack culture (#slack-culture)

One of the most notable things coming into Stripe was that the company didn't just use Slack, it _lived and breathed_ it. Almost every Stripe spends almost every moment of every day on it, and it's usually possible to use it to get a reply from almost every person in the company more or less immediately -- many of whom will reply even if they're in a meeting. Team "runners" respond to inbound asks, ideally within minutes, with traditional JIRA tickets only created for more complex issues. Bots augment the situation further by telling you when you need to a review, when a build finishes, or when it's your turn to deploy. Long, synchronous conversations are being held in Slack all day long, and it's the default place where decisions are made.

By comparison, email, the undisputed future of communication only two decades ago, with few exceptions is used only for company or department-wide blasts, and as a receptacle for botspam from code change or document change warnings, JIRA, and failed automated checks.

But Slack culture doesn't obsolete the traditional meeting either. When text conversations are progressing slowly or with difficulty, no one's afraid to jump up to a high-bandwidth face-to-face a meeting. A one-word Slack shortcut (`/zoom`) enables it instantly. There's less of a culture of meeting grandstanding as ridiculed by Dilbert in the 90s, but if anything, meetings are more frequent than before, especially with Zoom and universal WFH having reduced activation energy to zero.

### Inboxes with inboxes (#inboxes-with-inboxes)

Slack culture has some heavy upsides. Namely, instant access to _anyone_ on Slack is tremendously convenient, and can people expediently unblocked from sticky situations. And with so many conversations happening in public, search becomes a powerful tool, and you can often fix your own problems by searching Slack and using the same resolution as someone who's already had them.

But it's not all good. Under the yoke of Slack, a long, uninterrupted block of work is measured in the range of about 10-15 minutes, and it shows. Code changes tend to come in bite-sized patches, and improvements are made very incrementally. No one can afford to be focused on any one thing for long. Those who do ship large changes tend to develop them outside of business hours when there's less distracting Slack action.

People are adapting to their environment. Good multi-tasking was always a strength in business, but triple-A Slack-era engineers are the Olympian gold medalist _prodigies_ of multi-tasking, able to keep ten unrelated balls in the air while quickly moving between them with minimal context switching overhead. Self-sufficiency, on the other hand, is on fast decline -- it's easy to lose the ability to work out a problem for yourself when it takes seconds to ask someone else to solve it for you.

Most days I have a hard time telling whether it's a net good or net bad, but I tend to come down closer to the latter side. I'll be watching old Seinfeld episodes portraying offices in the early 90s and fantasize about the idea of a physical inbox that's checked a few times a day. Compare that to an email inbox, JIRA ask board, dozens of Slack channels and DMs, a multitude of potentially active Slack threads (["your inbox has inboxes"](https://twitter.com/triskweline/status/1035073193550249984)), and even one for tracking reactjis. At least the desk phone is gone.

---

## Pop v. funk (#pop-v-funk)

A correction from last week: in a [writeup on Japanese city pop](/nanoglyphs/023-enhancement#japanese-city-pop), I incorrectly linked [this video](https://www.youtube.com/watch?v=qXC4AyjRikg&t=3246s) as a city pop example, but it's actually a distinct genre called "future funk":

> Electroswing is to "swing from the 1920s" as future funk is to citypop. I love both genres but they're very different! There is also a modern citypop revival, with artists like [Junk Fujiyama](https://soundcloud.com/oases-ong/junk-fujiyama), producing straight citypop records in 2021.

This was a reading comprehension problem on my part. Thanks to Nate for the correction.

---

And now, just one last shout out to all the nice people who were reacted very encouragingly to me switching jobs. Finding a hole of negativity to burrow into is infinitely easy on the internet, and it's incredibly encouraging being washed over by a wave of the polar opposite. Thank you.

See you next week.

[1] Technically, Research In Motion at the time.
