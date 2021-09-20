+++
image_alt = "A lizard in San Bruno park"
# image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/028-cool-tools/lizard@2x.jpg"
published_at = 2021-09-18T15:18:57Z
title = "Lowbeer's Tipstaff; Cool Tools"
+++

Readers --

I recently enjoyed [Why William Gibson Is a Literary Genius](https://thewalrus.ca/why-william-gibson-is-a-literary-genius/) from The Walrus.

While I can't purport to be _totally_ crazy about Gibson's books, his cultural significance is unquestionable, having coined the term _cyperspace_ and more or less singled-handedly launched the entire genre of cyberpunk. You also have to appreciate just how out novel so much of his work is -- _The Peripheral_ revolves around a future version of Earth networking to a past version of Earth by way of a never-explained server mechanism in China. The entirely of _Zero History_ is the story of its characters searching for the fashion designer behind a "secret brand" of clothing called Gabriel Hounds.

One of the most distinguishing characteristics of Gibson's prose is how challenging it is. He doesn't spell anything out, instead relying on the reader to understand what happened in specific scenes by way of inference, and to extrapolate the overall narrative by way of clues planted along the way. I liked _The Peripheral_, but I was just re-reading my own review of it, and remembered that I'd felt that there was a little _too much_ inference required:

> Between invented terminology not explained until hundreds of pages later (the record by my casual read being "homes", defined on page 350), dense prose, and dialog that seems to suggest that the characters are deliberately trying to confuse each other, this book is difficult to read. The first 100 pages are practically incomprehensible, although it gets easier. [...]

As an example, the article's author describes a scene in _The Peripheral_ where Ainsley Lowbeer, an inspector in 22nd century London has a shape-shifting weapon called a "tipstaff" which can summon a drone strike. None of this is said explicitly (the following a quote from the book):

> [...] morph again, becoming a baroque, long-barreled gilt pistol, with fluted ivory grips, which Lowbeer lifted, aimed, and fired. There was an explosion, painfully loud, but from somewhere across the lower level, the pistol having made no sound at all. Then a ringing silence, in which could be heard an apparent rain of small objects, striking walls and flagstones. Someone began to scream.
>
> “Bloody hell,” said Lowbeer, her tone one of concerned surprise, the pistol having become the tipstaff again.

The article makes the case of how Gibson manages to say so much in so few words:

> The succession of double-take-inducing details is exquisitely managed. (It’s as if someone called in an airstrike on a rotary phone.) Gibson doesn’t explain how the tipstaff works or why it assumes the look of a “baroque” pistol; alert readers will get that tipstaffs are the products of nanotech and nostalgia, of advanced societies that have aestheticized how they do harm. A lesser writer, of course, would’ve insisted that the pistol do the firing, but Lowbeer’s ornamental weapon disgorges nothing, not a peep. The explosion is elsewhere, and Gibson is mindful that explosions have epilogues, the follow-up sound of raining objects “striking walls and flagstones. Someone began to scream.”

Like the title claims, genius.

---

Welcome to _Nanoglyph_, a newsletter about cyberpunk and data warehousing. This week: how Stripe taught its employees to use SQL, and why this is a very, very good idea, along with in the same theme as Lowbeer's tipstaff, a laundry list of cool internal tools for inspiration.

If you're wondering why I write so much about an organization which I no longer work for, the answer involves the fallibility of memory -- I want to get words to paper before it all fades into a foggy haze (some of which is already happening). Don't worry, I won't subject you to too much more of this -- my plan right now is a very complimentary issue this week, a less-complimentary one next week, and then into new territory.

---

## Data, lakes of it (#data)

Perhaps one of the more interesting systems we had in place at Stripe was an incredibly comprehensive data warehouse. The company prides itself in data-driven decision-making and design, and the warehouse was the backbone that made that philosophy possible.

Unlike what you might've seen in the past with more traditional businesses, the warehouse was with Stripe since quite early on -- it was well-concreted by the time I started in 2015. Likely the major impetus for the early adoption is how the company was powered by Mongo. With a relational database you can get away for a long time using it in hybrid form to power both transaction-processing and analytics, but with Mongo, you really have no plausible way of querying data beyond the most primitive operations -- you can ask your cluster to execute some custom JavaScript, but nothing like SQL exists.

Nightly ETL jobs would dredge tracts of Mongo pages, transform them into a more agnostic format, and bulk load them into a warehouse so that we could get at the information more easily. Originally, this was _a lot_ of extra infrastructure for the dubious privilege of getting to use Mongo as a datastore, but it did come with the benefit of establishing a warehouse and a well-worn path of getting data into that warehouse early on, and that to even more adapters until it became the one stop shop for everything as all kinds of data sources were connected -- Kafka queues, JIRA, GitHub Enterprise, Workday, etc. -- if you could think of it, it was probably in there.

Some examples of things that a person in my position might ask the warehouse about:

* Get the distribution of the major Ruby versions of users making requests with our Ruby API library and the number of unique users on each so that we could measure the fall out of dropping support for an old one [1].

* Get the contact email for all distinct merchants who've made a request with SSLv3 in the last six months so that we can reach out to tell them it's being phased out.

* Get an aggregate shape of a single user's API request profile over multiple days including average and maximum RPS, along with the API endpoints they use most. We'd then determine if we should bump them to a higher rate limiting tier or request that they make changes to a grossly inefficient integration.

Originally, this all operated on what was perhaps the world's most tortured Redshift cluster. It was impressive what this thing could do considering the sheer volume of data and load that we were throwing at it, but using it was rocky -- queries would run for minutes at a time if there was any other load in the cluster, and often just time out completely. An "observatory" tab was added to our custom UI so that users could go in and kill other peoples' stuck queries so that theirs might succeed.

Eventually, an alternative implementation on top of [Presto](https://prestodb.io/) was introduced, and after a period of hybrid support for both systems, we eventually transitioned entirely to that as it proved to be far more reliable and friendly for wide simultaneous use.

### SQL at scale (#sql)

A Stripe UI gave its internal users an interface not dissimilar from [Heroku Data Clips](https://blog.heroku.com/simple_data_sharing_with_data_clips). We could put an SQL query into an editor in the browser, and have it dispatched to the cluster. After results came back, it could be tabulated and plotted, then annotated with a title and description (so you could find it again later), shared via link, or forked for refinement. The interface was very raw early on, but eventually given a makeover and some Stripe-style spazaz.

It was a solid set up, but I'd hazard a guess that many major tech companies have something similar, especially with the glut of analytical products that are on the market these days. But something those other companies don't have, and what Stripe got _really_ right wasn't technical, it was organizational.

At most companies (and this applies even to tech companies in the valley), data is a human service as much as a technical one. People in sales, marketing, and sometimes even engineering hand off analytical requests to a team which then figures out how to get a result. They'll also be using data warehousing tools, but with a team of specialists running them. This works fine, and has the advantage that most of the company never has to get its hands dirty, but comes with all the obvious downsides -- slow turnarounds, resourcing problems as data teams can only take on so much, and information simply being used _less_ -- when it's painful to get and you can't do it yourself, after a while you're only going to be sending over the most important requests.

What Stripe got right: making analytics a self-serve process. Every employee could access the data warehouse and run their own SQL, and while it was expected that people would collaborate or ask for help while doing so, it was _not_ expected for them to dump the work on someone else.

SQL is a high bar, even for engineers, and especially for non-technical people. The basic case of pulling data from a table is pretty easy, but once you get into CTEs, inner versus left versus cross joins, complex aggregates, window functions, etc., it can warp the mind. Positing that non-technical people should be able to learn it is an aggressive position, but incredibly, it worked. I'd attribute a lot of that success to the power of the example -- although many people wouldn't be up for writing their own complex, deeply nested query from scratch, they were up for using someone else's as prior art, and manipulating that to get the desired outcome.

This "teach a man to fish" philosophy made a _huge_ difference. Being able to do something for yourself versus asking someone else to do it is a night and day difference in terms of whether it's likely to happen, and the common case on the ground was that everyone was running data all the time. Every company in Silicon Valley claims to be data driven, but Stripe _really is_ a data-driven organization if there is such a thing.

(I'll caveat by saying some of what I've said here got less true as the organization got bigger. Access to parts of the warehouse were request-only or restricted depending on the sensitivity of the data. We also did achieve final retrograde form with hands off middle management who would delegate rather than do.)

<img src="/photographs/nanoglyphs/028-cool-tools/ridge@2x.jpg" alt="A ridgeline in San Bruno park" class="wide" loading="lazy">

## Other neat things at Stripe (#stripe)

Stripe's data warehouse is great, but was only one amongst many powerful internal tools. Here's a few examples of others for inspirational purposes:

* **True SSO:** Stripe's internal authentication system and security story is the best I've ever seen, including even from professional IDP providers. Except in some very rare instances (Gmail, logging onto their laptop), no Stripe employee enters a password for any services, ever. All services are configured to use an internal single sign on (SSO) portal for authentication, and that portal identifies users via an OpenSSL client certificate sent by the browser and pre-provisioned to each laptop, which can be revoked by an admin in seconds. The portal requires entry of a second factor (Yubikey) once per day and all services require VPN access which requires another certificate, but again, no password. Most employees have no idea what's even happening, but even for the least tech savvy, the process is incredibly fluid. Paradoxically, no passwords means _better_ security, and the model that Stripe uses internally might in an alternate reality be the security model in use by the whole internet -- with client certs obviating the need for password managers -- if only OpenSSL wasn't so f*ing impossible to use.

    On my last day as access was being revoked at 3 PM sharp, I suddenly got about nineteen error modals stacked on top of each other as deeply nested layers worth of security protocols all failed simultaneously.

* **Home:** An internal home portal featuring links to many useful things along with an internal directory showing every employee, the team they're on, who they report to, and who reports to them. Each team had its own landing page that described what it did, where to find it on Slack, and how to open a ticket with them. Each employee had some key vitals listed along with time zone, location, and a calendar showing availability (available / partially available / off). It was better designed than any professional product in this space that I've seen on the market.

### Operations (#operations)

* **Deploy UI:** I like shells and terminals more than almost anybody else, but they're not a good fit for deploy procedures. Stripe's deploy UI walked an operator through each step of a progressive rollout process (deploy to canaries first, deploy to boxes taking less important traffic first), pausing at each step to prompt them to continue or cancel, and telling them where to find relevant dashboards. In case a revert is required, they're helpfully walked through the rollback process. This is a huge improvement over scrambling to find the right terminal commands to run during an emergency where you're not thinking straight.

* **Red/green deploys:** In short, keeping two sets of servers and alternating deploys between them. Requires running twice the server infrastructure, but allows you to roll back really, really fast. In a situation like Stripe's where being down for even seconds is an expensive affair, the cost/benefit calculus makes sense.

* **Big red button:** In the beginning, there was a Slack channel called `#warroom` where everyone would take to in the event of an incident -- fine for a smaller company but not scalable. Its replacement was a system where anyone could click one button to start a new incident response. It would get a randomly assigned human-memorable identifier, its own special purpose Slack channel, and incident doc. An on-call incident manager would get pulled in automatically, along with anyone else deemed necessary with an easy-to-use `/page` Slack command that could target a team or a specific person. Employees were instructed to bias towards starting incident responses and paging when in doubt, although this might not have worked so well if half the company wasn't made up of 23 year olds without families or major responsibilities, and willing to swallow infinite pain in the company's name.

* **Feature-flags-as-a-service:** A powerful flag system with a lot of configurability. Each flag gets a description for what it is, and can be activated for just single users, a percentage of traffic, a percentage of merchants, or a combination of all the above. Every non-trivial change gets rolled out under a flag so that in case it breaks, the blast zone is minimized, and it can be rolled back quickly.

### Development (#development)

* **Custom interface to CI:** A custom CI sitting on top of Jenkins. Instead of scrolling through terminal output looking for failures that occurred, they got a pretty list of specific test cases, each of which could be expanded for details, and which had a button to copy an invocation to clipboard that would run the failing test from terminal.

* **Review shorthand:** It was easy to request a review from either a team or specific user by including magic strings in GitHub comments like `r? @brandur`. This got less impressive over time as GitHub added its own review system that could be invoked quickly with `hub` (GitHub's command line client), but was still useful because it was smart in other ways too -- like waiting until after a successful build before requesting a review, or randomly assigning a reviewer on a target team.

* **Slack notifications:** Developers would get a Slack notification when one of their builds failed, succeeded, or if someone had requested their review, with the goal of tightening development loops and review turnarounds. I'm including this because it's interesting and it was a good implementation, but it was also kind of an anti-feature -- the only reason build notifications were needed is because builds took so long, and the idea of fast reviews is nice, but carried a definite price tag as developers were wedded all day to Slack, with distracting notifications firing constantly.

* **Livegrep:** This one's [a public project](https://github.com/livegrep/livegrep) originally started by a Stripe engineer, and in essence, a search engine for code that stays fast even where there's millions of lines worth of it, so it can index every repository in use at your organization. Tools like [the silver searcher (`ag`)](https://github.com/ggreer/the_silver_searcher) are fine for single repos, but it's useful having a place you can go where you're trying to search for an error message produced by an obscure project that you've never even heard of.

I should caveat that Stripe has a lot of uncool tools too -- at one point we went full Atlassian for reasons that I still don't understand (I swear that in aggregate the use of JIRA in the country's biggest enterprises must be reducing GDP by a full single digit), but the company has a lot of great stuff. It was very good at periodically observing pain points that were costing productivity, and tasking people to reduce them, thereby producing more compound leverage across the whole organization.
 
 ---
 
## Go ORMs and text (#orms-and-text)
 
 Speaking of cool tools, Postgres is a cool tool. Last week I published two articles on it:
 
 * [How we went all in on sqlc/pgx for Postgres + Go](/sqlc): Where I argue that if you're using Postgres with Go, [sqlc](https://github.com/kyleconroy/sqlc) is the ORM-adjacent tool that you want.

 * [Postgres: Boundless `text` and back again](/text): Where I make the case that as great as Postgres' `TEXT` type is, you might still want to use a `VARCHAR` instead.

---

## California streamin' (#california-streamin)

With Apple's annual iPhone announcement event having come and gone already ("California Streaming" -- give whoever came up with that one a raise -- perfect), I was just reflecting on how far tooling in the form of hardware's come in just a decade or so.

The first laptop I ever bought myself was Apple's plastic MacBook in my last year of university. It was okay. Good for the time, but weird issues around the plastic discolouring and peeling back, and although you'd get a few classes worth of power out of it, you didn't leave home without a charging brick. I would've been using a Motorola KRZR at the time, with which I'd occasionally text people at three words per minute.

Fast forward to today, the iPhone 13 pushes the 12'a already-tremendous battery life another 2.5 hours, and is a few orders of magnitude more powerful than my old MacBook. I'm still using a five-year old iPad Pro, but it's still _perfect_, with almost edge-to-edge screen and charged only once a week. Laptops were the final frontier in decent on-the-go battery, but Apple's M1 finally cracked that one too -- usable all day as long as you remember to plug the computer in overnight.

The world is more messed up than ever, and the very fabric of our civilization might be coming apart, but g'damn do we ever have some cool tools.

Until next week.

[1] We'd often offer tacit support for old language versions considered deprecated by core because so many users have trouble upgrading in a timely manner.
