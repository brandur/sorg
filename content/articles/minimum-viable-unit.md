+++
hook = "LLMs made software cheaper to build, but not free. A look at the new buy-vs-build math of the LLM age, and the pricing zone where small software businesses can still survive."
image = "/assets/images/minimum-viable-unit/vista.jpg"
location = "Berlin"
published_at = 2026-05-31T12:41:38+02:00
title = "The Minimum Viable Unit of Saleable Software"
+++

Last week I wrote about [leaving Stainless](/nanoglyphs/051-that-was-fast) and my intention to work on building my side project [River](https://riverqueue.com) into a small, sustainable business. When I sent that letter, a few people asked about my thought process in trying to run a software company in the age of AI: "Are you crazy?! Anything you ship can be instantly displaced by an internal package built by an LLM!" Having become as much of an LLM convert as anyone at this point, I acknowledge that it's a very fair question. Indeed I might be crazy, but I'll talk through my thought process, and you can decide.

Let me start with an anecdote. This morning I was browsing the internet's most wretched hive of engagement farmers and master solicitors of fake information and fictional anecdotes, LinkedIn. One user there posted about how his company had been spending $400/mo on Atlassian's Jira. He'd felt personally slighted by this outrageous bill, so he'd had his team build a new internal task tracker using Claude. Gone was Jira and the $400/mo spend, replaced by a custom package that could be tooled out in any way they needed via continued refinement by an LLM.

We've been talking about buy vs. build in software circles for years, but last year the calculus changed. It used to be that build was a _very_ expensive proposition, especially given the state of engineering salaries and scarcity of great people. One could expect huge upfront cost, schedule overruns, and an infinitely deep rabbit hole to slide down. The general wisdom had always been to build only inside your core domain and avoid getting sidetracked by peripheral projects. Once your company reached enormous size, and the cost of those distractions disappeared comfortably into its margins, then maybe they'd be worth doing.

But LLMs changed all of that. Suddenly it was quite possible to produce substantial pieces of software by getting models to do the work.

---

## Cheap != zero (#cheap-ne-zero)

While LLMs have made software considerably cheaper to build, they haven't brought it to zero. Good LLM-built systems still involve a feedback loop, where an operator has the model work for a while, makes adjustments based on results, asks for another pass, refines further, and so on, taking dozens of loops to get to a satisfactory result that's an optimal compromise between time spent and quality.

And like before, maintenance will be an ongoing cost. Especially for more complex packages, there's always going to be a feature to add or bug to fix. LLMs will make those changes easier to make, but don't make them free, with the most expensive element being the part-time labor of the human in the equation who oversees and verifies results.

Back to our $400/mo Atlassian anecdote above: after considering the initial build effort, including refinement passes, and the ongoing LLM-driven maintenance, does it pass the smell test, like at all? A task tracker's still a complex piece of software, and even with gratuitous use of LLMs, you'd expect to spend at a minimum a few weeks on the initial push (charitably). From there, its internal owner will switch to bug fixes and feature development.

Let's try to come up with some rough numbers to quantify the situation. Let's say we have an engineer making $200k/year and working 40 hours a week (pretend for a second 9/9/6 was blessedly never conceived). That's $16.7k/mo, $3,850/week, or $96/hour:

``` ruby
salary = 200_000.0

{
  month: salary / 12,
  week:  salary / 52,
  hour:  salary / 52 / 40,
}.each { |k, v| puts "%-6s $%0.2f" % ["#{k}:", v] }
```
``` txt
month: $16666.67
week:  $3846.15
hour:  $96.15
```

To counterbalance the $400/mo that would've been paid to Atlassian, the engineer can spend _no more_ than 4 hours a month (400 / 96) prompting features/fixes on their homegrown Jira clone, or looking after its database, or whatever, not including context switching overhead. Even with LLM help, that's completely unrealistic already, but let's be charitable and say they can get it down to 2 hours a month. It'd still take _37 months_ to break even after those initial 2 weeks of effort (number of months to make back Atlassian's $400/mo minus 2 hours/mo maintenance effort = 2 * 3846.15 / (400 - 2 * 96.15)).

Don't get me wrong, I hate Jira just as much as anyone who's ever used it and have a nearly uncontrollable urge to want to rebuild it too, but the math here doesn't pencil out [1].

### The build threshold (#build-threshold)

But does that always hold true? Let's take the other side for a second by examining a much higher-priced SaaS product. Gemini reports that the price of a fully loaded Salesforce seat is ~$500/mo. Say you need 50 seats, that's $25k/mo!

For that price you could have 1.5x engineering resources (25 / 16.7) working on your clone full time. Once again, a CRM's a reasonably complex piece of software and a rebuild wouldn't be trivial, but no matter how you construe it, this is closer to a "build" decision, even for a smaller company. (And with Salesforce down 30% YTD, the markets seem to believe it too.)

---

## The zone of viability (#zone-of-viability)

I'm contending (and/or hoping) that for a software package of arbitrary complexity, there's a **zone of viability** in which when priced within reason, it'll make sense to buy over build, even given the existence of the powerful LLMs that've become our daily companions:

<img src="/assets/images/minimum-viable-unit/zone-of-viability.svg" alt="Zone of viability in a sweetspot between cost and complexity">

Software in the zone of viability satisfies two conditions:

* There's sufficient novelty as to make a rebuild-by-LLM non-trivial, and with some ongoing maintenance burden.

* Pricing is not so exorbitant as to strongly encourage rebuild-by-LLM.

As long as continued pricing within reason keeps software within the zone of viability, the total paid in licensing is less than the cumulative expense of prompting its initial push and sustaining its continued existence.

Somewhere along the zone of viability is the **minimum viable unit of saleable software**, below which a rebuild is the same or less effort compared to going through the purchasing process for a third party and not cost-effective over the long run.

| | Ongoing price | Ongoing spend | Engineer equivalent hours/mo | Equivalent engineering resources | Buy | Build |
| --- | ---: | ---: | ---: | ---: | :---: | :---: |
| **Jira** | $400/mo | $400/mo | 4.2 hours | 0.02 engineers | ✔ | |
| **Salesforce** | $500/seat/mo | $25k/mo | 260 hours | 1.5 engineers | | ✔ |

---

## River as a plausible business (#river)

For the last few years Blake's worked on a small business based on our [open-source project River](https://riverqueue.com), a job queue for Go and Postgres, and for at least the next few months, I'll be taking over full-time. This self-serving blog post is a long way of saying that I hope that despite the world having crossed the LLM horizon, River comes in over the minimum viable unit of saleable software and is still a plausible company in the modern age.

In terms of novelty, River's an open-source project that makes almost all job-related features (periodic jobs, scheduled jobs, unique jobs, web UI, ...) available for free, but reserves some advanced features (workflows, sequential jobs, concurrently-limited jobs, ...) and billing capability (billing by invoice) for a [Pro version](https://riverqueue.com/pro) that we charge for. An LLM could reproduce the latter features, but we've put enough thought into their API design and performance properties that it'd take some work to get back to something of similar fidelity.

In terms of price, we used a sublinearly scaling pricing model based on team size rather than headcount, starting at $125/mo for up to 20 developers, and scaling up to a multiple of that for an unlimited site license. So for a small-to-medium development team, $125/mo is the all-in cost across everyone.

So back to the question at the top: did I get this right? Who knows. For now I'm betting my livelihood on it, and the coming months will tell.

---

_A note on the photo at the top:_ This is a natural feature called "Zlatnite Mostove" ("The Golden Bridges") in the Vitosha mountains near Sofia, Bulgaria where I hiked recently after attending [Balkan Ruby](https://balkanruby.com/). The field of rocks is called a "bridge" because it covers an active river underneath it. This post is partly about River and that's a river, so I'm banking on enough of a connection to be justifiable.

[1] It does, however, pencil out to use a different product instead. In this particular case, it's easy: use Linear instead of Jira.
