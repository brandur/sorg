+++
hook = "How we can expect a new wave of APIs, driven by user demand to open products and services for use through LLMs."
location = "Seattle"
published_at = 2026-03-27T08:11:08-07:00
title = "The Second Wave of the API-first Economy"
+++

Over a decade ago when some colleagues and I were building out Heroku's V3 API, we set an ambitious goal for ourselves. The new _public_ API should be so expansive, so flexible that it'd be suitable to power our own user-facing dashboard. No private APIs or other escape hatches necessary. If a user was so inclined, they could use the API to build their own version of a dashboard.

It worked. Not only was the company's new dashboard built on V3, but an outside user would go on to build Heroku's first ever iOS app on it, with no additional feature requests sent our way.

---

## The first wave (#first-wave)

Our dashboard-on-public-APIs-only seems needlessly idealistic nowadays, but it was an objective born of the time. The year was 2011, and the optimism around the power of APIs was palpable. A new world was opening up. One of openness, interconnectivity, unbounded possibility.

And we weren't the only ones thinking about it:

* Only a year before (2010) Facebook released its original Open Graph API, providing immensely powerful insights into its platform data.

* Twitter's API at the time was almost completely open. You didn't even need an OAuth token --- just authenticate on API endpoints with your username/password and get access to just about anything.

* GitHub was doing really impressive API design work, providing an expansive, feature complete API providing access to anything developers could need, and playing with forward-thinking ideas like hypermedia APIs/HATEOAS.

You can still find traces of this bygone era, standing like some cyclopean ruins from a previous age. Hit the root GitHub API and you'll find an artifact over a decade old --- a list of links that were intended to be followed as [hypermedia](https://en.wikipedia.org/wiki/HATEOAS):

```sh
$ curl https://api.github.com | jq

{
  "current_user_url": "https://api.github.com/user",
  "current_user_authorizations_html_url": "https://github.com/settings/connections/applications{/client_id}",
  "authorizations_url": "https://api.github.com/authorizations",
  "code_search_url": "https://api.github.com/search/code?q={query}{&page,per_page,sort,order}",
  "commit_search_url": "https://api.github.com/search/commits?q={query}{&page,per_page,sort,order}",
  "emails_url": "https://api.github.com/user/emails",
  "emojis_url": "https://api.github.com/emojis",
  "events_url": "https://api.github.com/events",
  ...
```

This wasn't a pre-planned, stack-ranked feature that a product team spent half a year putting together. It was one or two early engineers who got really excited about an API idea, and shipped it, probably without even asking for permission.

---

Part of the push for open APIs was simple good will towards the rest of the world. The engineers building them were brought up in the earliest days of the internet, steeped in its original counterculture, and had an innate bias towards radical openness.

There was also a feeling from the companies involved that the APIs would be beneficial for their bottom lines. Users and third parties would use APIs to supplement the core product with add-ons and extensions that'd drive growth and increase product retention and satisfaction.

Sites like the now defunct ProgrammableWeb popped up to discuss and catalog the newly appearing APIs, and the "programmable web" wasn't only a website, it was a principle.

In the near future, _all_ platforms would be API-first, providing full programmatic access and opening a new wave of interoperability across the web that'd let any service talk to any other service and massively accelerate the scope and reach of the internet. APIs would help expand everything from freedom to communication to commerce. An overwhelming force for good in the world.

---

## The frontier closes (#frontier-closes)

Of course, it didn't last. The programmable web went through a phase of expansion, reached its maximum extent, and began to contract.

* Twitter's famous API, which used to be an API tinkerer's dream, leveled off and began to dip as the company struggled to find ways to generate revenue. New features no longer got first-class API treatment. Access to the firehose was closed. Third-party Twitter clients were restricted and eventually locked out.

* The power of Facebook's Graph API was hugely constricted post-Cambridge Analytica where a single rogue app was able to suck up data on millions of users and put it up for sale. Strict app review procedures were implemented. The API went from open access to a walled garden.

* Even more extreme, Instagram's previously public API was deprecated totally. Realizing they had a real money maker on their hands, they saw no reason to share ad revenue with anyone else. Use Instagram through the first-party app or not at all.

* Even APIs like GitHub's that stayed quite open had to crack down to a degree. Endpoints became authenticated by necessity and aggressive rate limiting was put in to curb abuse and reduce operational toil. Even when APIs were still quite open, using them to build a full-scale third party app became more difficult as limiters flattened heavy (even if legitimate) use.

The rationale for why APIs were being declawed or disappearing completely varied---abuse, monetization pressure, competitive risk, privacy, etc.---but the pattern was clear. Walls were going up across the world.

APIs didn't disappear, but the expectation of an API became more limited to developer-focused platforms whose users paid them --- Stripe, Twilio, Slack, etc. When new consumer products appeared on the market (e.g. TikTok), no one expected them to have much in the way of an API.

---

## The coming second wave (#second-wave)

For many years this was the status quo. If you were using Twitter, you'd use it from Twitter.com. Facebook, from Facebook.com. Instagram or TikTok, from their respective iOS/Android apps. Developer products like GitHub and Stripe continued strong, but elsewhere, APIs weren't enough of a competitive advantage for anyone who didn't have one to suffer.

But around mid-2025, the world changed. The last half year especially has been distinguished by the rise of indescribably powerful LLMs, which now dominate discourse as the most useful new tool in the world.

They're already useful enough as incredible trivia machines or code generators, but they really start to shine when they integrate with things. For example, it's pretty neat having one generate a valid Kubernetes configuration for your new app, but it's _really_ neat watching it provision an <acronym title="Amazon Elastic Kubernetes Service">EKS</acronym> cluster via `awscli` and send out its first production deploy on your behalf.

Suddenly, an API is no longer liability, but a major saleable vector to give users what they want: a way into the services they use and pay for so that an agent can carry out work on their behalf. Especially given a field of relatively undifferentiated products, in the near future the availability of an API might just be the crucial deciding factor that leads to one choice winning over another.

### Picking my future bank (#my-future-bank)

Let's think about banks. I have a couple bank accounts, each offering a standard set of features largely unchanged since the 60s. If I call them, they'll send me some checks. I can request a transfer between two internal accounts and they will transfer the money ... in 1-5 business days. Nowadays, they even offer ultra-modern features (from 2010) like *gasp*, MFA, just as long as it's through a provider that's paid them off (Symantec VIP). Suffice it to say, they're comfortable in the status quo. My banks do not have good APIs.

So far this has worked out okay for them. People aren't known to migrate banks often, and even if they did, regulatory moats make new incumbents rare.

But in the modern age, can it last? When I want to move $100 from one bank to another, my banks put me through a humiliating ritual of logging into both accounts, and bypassing multiple security checks and captchas before I can perform any operation. All this despite me having just logged into both accounts from this exact location and biometrically-secured computer the day before.

The world I _want_ is to instruct an LLM: "move $100 from Wells Fargo checking to Charles Schwab brokerage" and it will just _happen_. And to be fair, LLMs are already so absurdly good at reverse engineering things that this might already work today. But you know what'd work better? If both banks shipped with APIs, LLM-friendly usage instructions (through MCP or the like), and a strong auth layer to give me confidence that the whole process is secure.

If I were choosing a bank today, some considerations would be the same as they've always been---competent security, free checking, no foreign transaction fees---but I'd also futureproof the choice by picking one that's established technical bona fides by providing an API. Even if I'm not quite ready to trust all my credentials to an agent quite yet, I assume that this day is coming.

### APIs, ubiquitous again (#ubiquitous-again)

Now apply the same principle to every service you use during the course of a week, or ever:

* **Online marketplaces:** Robot, schedule my normal Amazon Fresh order for the first available slot tomorrow morning.

* **Office co-working:** Robot, book me a desk at Embarcadero Center today.

* **Ski resorts:** Robot, buy me a day pass for tomorrow and load it to my resort card. Confirm the price with me first.

* **Restaurants:** Robot, put in my usual lunch order at Musubi Kai. Get me the unadon!

Where _wouldn't_ you want an API?

Forecasting the future is infamously hazardous, but based on the adoption patterns of myself and the people around me, I expect the demand to interact with services through LLMs is going to be overwhelming, and services aiming to provide a good product experience or which have competitive pressure (i.e. someone else could provide the good product experience instead) will offer APIs.

I used to wish that we'd gone down an alternative branch of web technology and adopted a protocol like [Gopher](https://en.wikipedia.org/wiki/Gopher_(protocol)) so we'd have a more standardized web experience instead of every product you use producing its own unique UX, most bad. I think we will see more standardization, just not in the form I expected. The convention of the future will be human language, fed into what looks a lot like a terminal, and fulfilled via API.

### On behalf of people (#on-behalf-of-people)

Notably, this is different than the first wave of APIs that I described above. Instead of APIs being to offer infinitely flexible access for inter-service access, scrape data, or build apps on top of someone else's platform, their primary use will be to fulfill requests on behalf of a primary user. Exactly like what they'd be doing through a first party app, but in a programmatic way.

{{Figure "During the first wave, APIs were largely aimed at third parties who'd use them to extend and augment the underlying platform to provide additional features for users." (ImgSrcAndAltAndClass "/assets/images/second-wave-api-first/first-wave.svg" "During the first wave, APIs were largely aimed at third parties who'd use them to extend and augment the underlying platform to provide additional features for users." "overflowing")}}

{{Figure "In the second wave, APIs map cleanly to normal product capabilities. They provide programmatic access for agents that act on behalf of people." (ImgSrcAndAltAndClass "/assets/images/second-wave-api-first/second-wave.svg" "In the second wave, APIs map cleanly to normal product capabilities. They provide programmatic access for agents that act on behalf of people." "overflowing")}}

It may seem like a subtle difference, but there are considerable differences. The second model better incentivizes APIs to exist:

* APIs aren't for building a product that aims to displace the offerings of the underlying platform, but rather for giving users an alternative way to access it.

* Security models are simplified because they're the same ones used by the product itself. Users have the same visibility that they'd have through a first party app, and no more.

* Aiming to support access patterns for a single person, platforms can rate limit much more aggressively to curb expenses and operational problems associated with offering an API.

APIs should aim to provide a little more leeway than they would for a human, but only nominally so. An agent acting on my behalf should be able to occasionally poll LinkedIn for old colleagues that I should be reconnecting with and send them connect requests, but if someone's set up their ClawBot to scrape the entire social graph on their behalf, platforms should feel more than free to throttle the hell out of them and give them a strike towards a permanent ban.

[Slack's rate limits](https://docs.slack.dev/ai/slack-mcp-server/#rate-limits) are a good example of this, supporting numbers like 50 channel or 100 profile reads per minute.

### Limits of the model (#limits-of-the-model)

While can expect many products and services to offer APIs for good agentic interoperability, it won't be forthcoming everywhere.

Don't expect much out of Instagram, TikTok, or other platforms that power themselves with ads. Neither from monopolies that won't feel any serious pressure to change --- you won't be reliably paying your Xfinity bill via agent anytime soon.

### Hints of the future, today (#future-today)

In this section I figured I'd call out a few services that are already pulling this future forward:

* As I was in the middle of writing this essay, I got a [note from Basecamp](https://world.hey.com/dhh/basecamp-becomes-agent-accessible-3ae6b949) that they'd revamped themselves for LLM accessibility, including [new API](https://github.com/basecamp/bc3-api), [new CLI](https://github.com/basecamp/basecamp-cli), and [bundled skill](https://github.com/basecamp/basecamp-cli/blob/main/skills/basecamp/SKILL.md) to instruct agents on their use.

## The spring thaw (#spring-thaw)

Fifteen years ago, a number of us maximalists thought that APIs were going to eat the world, ushering in a new paradigm of interoperability that was going to vastly expand our capabilities as users and even change the world for the better.

What we got instead was an API winter. As useful as APIs were in some situations, that usefulness was outweighed by concerns around revenue, privacy, and abuse.

But as scary of a thought as it was that this might be the end, it wasn't. We're at the beginning of a new wave of APIs that'll appear to support use by agents acting on behalf of people. As this mode of operation gets more popular, expect the availability of an API to be a competitive edge that differentiates a service from its competitors. The result will be a global proliferation of APIs like never before seen.