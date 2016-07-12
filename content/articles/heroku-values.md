---
hook: Some of my favorite practices and ideas from almost four years at Heroku.
image: "/assets/heroku-values/heroku-values.jpg"
location: San Francisco
published_at: 2015-11-05T06:20:16Z
title: My Heroku Values
---

In the spirit of [Adam Wiggins' inspiring list of Heroku
values][wiggins-values] which was published when he left the company that he
co-founded, I wanted to publish a list of my own now that I've transitioned
away.

My time at Heroku was easily the most valuable learning experience of my life,
and I'll always remember my time there very fondly. I remember upon first
joining just how disfunctional and inefficient that it made the jobs that I'd
held previously seem in contrast, and I'm hoping that by putting some of these
concepts down on paper I'll be able to reference and re-use them in my future
work.

I suspect that at least some of these ideas might be interesting to even those
with no relation to the company. Heroku was a place founded and formed by
people who came from outside the traditional corporate structure, and what
resulted was a mostly divergent branch of structure compared to mostly anywhere
else. Even the bad ideas should be novel enough to be intriguing in a mild way
(not to suggest that there any here of course!).

I should add the caveat that this is a compendium of values from the entire
duration of my stay at the company; not all had been established when I got
there, and not all were still in place when I left.

## Technology (#technology)

### The Platform (#platform)

One of the greatest pleasures during work at Heroku was the Heroku product
itself. Apps could be created and deployed in seconds, which encouraged
innovation by making prototyping easy, and allowed incredibly fast iteration on
production products. Every company should have a Heroku-like interface for
their developers to use.

I wouldn't go so far to say that companies should definitively use Heroku, but
it is a good way to have one without a major investment in infrastructure. As a
company scales, it might be worth putting a self-hosted one in place like
Remind has done with [Empire][empire] or Soundcloud has done with
[Bazooka][bazooka] (PDF warning). GitHub's model of deploying experiments and
small apps to Heroku and eventually promoting them to more dedicated
infrastructure (if necessary) is also worthy of note as a pretty nice
compromise between development agility and performance.

### Dogfooding (#dogfooding)

Continuing from above, we used our own products wherever possible. Every
production app at the company was deployed on the Heroku platform except for a
small set of core services that couldn't be. More internal Salesforce apps were
making their way over every year as well, demonstrating that the idea was
valuable enough to be organically making its way out and into the much larger
parent company.

Every internal app that required a login (e.g. the Heroku Dashboard, the Help
system, the add-ons SSO portal) used the same Heroku OAuth provider that's
available to third parties, leaving services loosely coupled and easy to build.

Still one of my favorite accomplishments is that Dashboard (the service that
allows customers to log into a web interface and manage their apps) runs off of
the same [public V3 API][heroku-api] available to customers. I can't even
describe the number of bugs uncovered by this technique; bugs that would have
otherwise been encountered by frustrated customers or third-party developers.

### Twelve Factor (#twelve-factor)

[Twelve-factor][twelve-factor] methodology provided a very nice set of guiding
principles for internal apps so that an engineer could reason about them more
easily. Every app got its configuration from the environment. Every app had a
Procfile. Every app emitted logs to standard out.

I've previously read criticism on twelve-factor which postulates that it's an
artificial set of principles to work around limitations in the platform. I
don't buy this for a second, but I'll let [Randall Degges cover this
position][degges-12factor] because he puts it far more succinctly than I ever
could.

Eventually some of us would wish for and try to develop even stronger
conventions for building apps (see [service
conventions](#service-conventions)), but the relatively straightforward set of
twelve-factor principles got us started and would always act as a solid
foundation that everyone agreed on.

### The HTTP API Design Guide (#http-api-design)

A fundamental law of the universe is that every engineer will design an HTTP
API slightly differently, even if they're being guided by prior art. This isn't
always a problem, but it's a challenge if you're trying to keep an API cohesive
when it might be contributed to by dozens of different people. I've seen
engineers name their new resource `/resource_with_underscores` even though 78
out of 78 existing resources look like `/resource-with-hyphens`.

We knew that if we wanted a consistent public API, we needed to codify a set of
opinionated conventions, which is why we wrote the [the HTTP API design
guide][api-design-guide] based off of the decisions we'd made building the V3
API. The result is that Heroku's API is one of the most self-consistent HTTP
APIs that you'll find anywhere in the real world.

### Service Conventions (#service-conventions)

Twelve-factor offered some convention when it came to deploying new services,
but we tried to take standardization much further with our service toolkit
[Pliny][pliny], which was designed to offer a powerful out-of-the-box stack
that would be a sane choice for most internal Heroku apps.

The only misstep with regards to Pliny and service conventions is that we
should have pushed them earlier and harder. Even the basic form in which the
project exists today took the company a long way in that not every new service
was a special snowflake of its author's favorite ideas (previously a major
problem), but we could have gone so much further by putting in automatic
distribution of updates, more free services (e.g. built-in rate limiting), and
service discovery and registration. Internal service frameworks are an
important enough problem that most mid-sized I/P/SaaS companies should have
dedicated people building them.

### Postgres (#postgres)

If given the opportunity to start a new stack from a blank slate, I might avoid
some of Heroku's current technological staples (e.g. Ruby). One of the few that
I _would_ use without a doubt though is Postgres. It's powerful, flexible,
incredibly stable, and has consistently been a pleasure to work with over the
years. Having recently had the misfortune to see how other aggrandized database
software operates in production, I feel that I now have an especially sober
view of just how good it really is relative to other products on the market.

It's possible that we missed out on some cutting edge technologies that would
have offered major benefits, but the resources saved by _not_ jumping on every
data store du jour is incalculable. There's probably still room in Heroku's
stack for an HA store, but it was the right thing to do to delay the
introduction of one until a number of mature options were available. In the
meantime, we got really good at operating Postgres and it was fine for almost
everything.

The only thing better than Postgres itself was our Heroku Data team (known
affectionately internally as the DOD, or Department of Data). This team of
hugely talented engineers saved my skin an untold number of times as I dealt
[with pretty messy operational problems][postgres-queues] [1]. I was told a
number of times that as the operator of our largest internal database, I was
their highest-maintenance customer, and it was true.

### Ephmeralization (#ephemeralization)

One powerful idea was that of _ephemeralization_, which can be roughly
described as "doing more with less". But aside from doing more, the act of
reducing the number of moving parts in a system helps to lower its cognitive
burden and made learning it easier. In a similar vein, picking one true way
forward from a collection of similar options helps keep engineers productive as
they move between components.

A few examples:

* Pick and choose single "right" technology stacks from a set of like options.
  For example, prefer Ruby over Python. More generally, try to focus on _just_
  Ruby and Go (and Javascript) over the long run.
* Try to zero in on particular library to perform certain functions. For
  example, preferring Puma for Ruby HTTP stacks by converting existing installs
  of Unicorn, or Thin. Using Sequel instead of ActiveRecord.
* Standardize deployment images so that instead of having individual Chef
  recipes for every component, all would share only one and be configured
  purely at the application level.
* Use a single type of data store consistently. i.e. Postgres.
* Don't create internal forks of libraries (this one should be obvious, but it
  doesn't seem to be).

### Use Services (#use-services)

Whenever you can use hosted services instead of operating them yourself.
Although the cost of infrastructure and bringing a new service online is
usually fairly well-understood, the full personnel costs of maintaining that
service (i.e. who's going to upgrade it and migrate data a year down the road)
and retiring it when the time comes are rarely considered.

## Culture (#culture)

### Leadership & Inspiration (#inspiration)

I've never had the opportunity to work with so many people who inspired me on
such a fundamental level as those who I met at Heroku, especially in my early
days there. The company had everything at one point: great leaders, inspiring
thinkers, and incredibly ambitious engineers. As someone still relatively
inexperienced and new to technological powerhouse is the Bay Area, my first few
months felt like a constant assault of new ideas about everything from
technology to organizational structure. This motivated me to want to build
great things and made work and the learning I did there all around exciting.

### Self-service (#self-service)

Instead of doing work for someone, give them the tools necessary for them to do
it for themselves. For example, Heroku's core API service had a private
administrative branch that employees with a CLI plugin could use to perform
special actions like re-send a sign-up e-mail. This creates a powerful
precedent for people to try to do things out for themselves before leaning on
someone else. If sufficient coverage is reached, this technique helps to
prevent [constant disruption on open communication channels][slack-distractor]
so that people have time to work.

### Cross-team Contribution (#cross-contribution)

Want a new feature or improvement? Send a pull request for it. There is no
better way to demonstrate your commitment to an idea. It also had the side
benefit of giving engineers a wider insight into how the whole machine works by
forcing them to look beyond the narrow confines of the projects that they might
maintain day to day.

This obviously doesn't scale to infinity, but it does scale far further than
many people would have you believe.

### Shipping Cadence (#shipping)

We shipped our services fast and frequently, and had framework of tooling to
make it safe to do so. You'd more often than not see a change go out same day,
which kept endless possibilities open for shipping new products or improving
existing ones.

This was also something that had to be discovered at the organization level.
There was a period in Heroku's history where projects were hard to ship mostly
due to a weak process for getting them across the finish line. This problem was
examined and corrected, and today products make it out the door on a regular
basis.

### Strong Engineers (#engineers)

At its essence, this one is pretty obvious: hire good engineering talent.

But things get a little more murky when examining them in closer detail. You of
course want to look for people who are good at what they do, but it may be even
more important for them to be flexible enough to jump in and fix bugs or modify
almost any project. This requires a degree of being able to learn indepedently
and figure things out for themselves that not everyone is well-suited for, but
if achieved will result in fewer disruptions to the rest of the team and more
work output overall. These ideal candidates may not be able to do a good job of
inverting a binary tree on a whiteboard and may not have a Stanford education
on their CV, and the interview process may have to be adjusted accordingly to
accommodate them.

For quite some time we have a team that would sync up once a week and plow
through huge workloads for the rest of it. Communication happened largely
asynchronously except for the occasional instance where a higher bandwidth
channel was more suitable. It was the most productive environment that I've
ever seen.

### Technical Culture (#technical-culture)

Technical culture was fostered, which (I believe) led to a high degree of
technical excellence in the products that we produced. This mostly manifested
in the way of papers being passed around, general discussions on the
engineering mailing list, and plenty of forward-thinking water cooler
speculation on how to improve products and internal architecture. For a long
time we also held a technical event every Friday called "Workshop" where
engineers could show off some of the interesting projects that they were
working on. It was designed to educate and inspire, and it worked.

### Flexible Environment (#flexible-environment)

<figure>
  <p><img src="https://farm4.staticflickr.com/3685/9549450965_84f27e06b4_z.jpg"></p>
  <figcaption>The Agora Collective in Berlin.</figcaption>
</figure>

<!--
![Agora](https://farm6.staticflickr.com/5538/9549457229_fbd6c7c464_z.jpg)
-->

Traditional organizations generally hold a strong belief that every employee
should physically punch in at 9 AM, leave it at 5 PM, and keep that up for 5
days a week year round. At Heroku people would regularly work at home or out of
the office. It made very little difference to their productivity, but did have
a profoundly positive effect on their overall happiness. For example, I visited
my family back in Calgary for weeks at a time two or three times a year, and
worked from Berlin for roughly three weeks almost every year that I was at the
company.

This is all possible if a company hires well. If you've got the right people on
your team, you don't have to keep an eye on them all day because they'll do the
right things themselves.

### Coffee (#coffee)

Admittedly, this one is a little self-indulgent, but I came to appreciate
coffee for the first time while at Heroku. For the longest time, there wasn't
even a coffee machine in the office; just Chemex pots, a grinder, and paper
filters. The idea was that making coffee would be five to ten minute process,
during which there would be time to interact with colleagues who happened to
drop by the area. The system worked.

I leanrt how to use both Chemex and AeroPress; both of which I continue to use
regularly.

## Process & Organization (#process)

### GitHub (#github)

<figure>
  <p><img src="https://farm8.staticflickr.com/7727/16585790614_1b6a09c72e_z.jpg"></p>
  <figcaption>The OctoTrophy (dodgeball).</figcaption>
</figure>

GitHub has been one of the best pieces of software on the Internet for years,
and is the right way to organize code and projects. Companies should be using
tools that developers can extend to optimize their workflows and maximize their
own their efficiency. With a well-maintained API and healthy ecosystem of
supporting tooling like [hub][hub] and [ghi][ghi], as well as complementing
turn key services like Travis, GitHub is one of those tools.

Time that developers *don't spend* supporting custom infrastructure or fighting
bad tooling is time that can be used to build your product.

### Access to Resources (#resources)

If an engineer needed a new resource for a service being deployed, prototype,
or even one-off experiment, they were at liberty to provision it and keep on
working, even if that resource wasn't free. Resources here might include
everything from dynos to run an app, to a Postgres backend, to a few extra EC2
boxes for deployment to bare metal (relatively speaking). Having Heroku's
considerable catalog of add-on providers and being completely deployed to AWS
helped a lot here in that no internal personnel were ever needed to help with
provisioning.

This practice works because despite a nominal cost to the organization, it
keeps engineer momentum up and the cost of prototypes down. Hopefully it's
becoming fairly standard practice in many newer companies these days, but it's
an easy thing to get wrong. I've previously seen the other side where
provisioning a job queue is a multi-month process involving endless meetings,
territorial ops people, and mountains of paperwork. Although some care needs to
be taken to not shoot from the hip when dropping in new technology, that
approach doesn't help anyone.

### Total Ownership (#total-ownership)

Our own version of "devops", total ownership was meant to convey that a team
responsible for the development of a component was also responsible for its
maintenance and production deployment. This added mechanical sympathy has huge
benefits in that getting features and bug fixes out is faster, manipulating
production is less esoteric, tasks that require otherwise tricky coordination
(like data migrations) are easier, and generally resulting in every person
involved taking more personal responsibility for the product (which leads to
more uptime).

Total ownership was instrumental in helping me to improve my skill in
engineering, but I'm still a little on the fence about it. While I don't miss
the multi-week deployment schedules, I do miss the regular blocks of daily
focus during which I would never have to stop work and deal with an
interruption from production.

### Technical Management (#management)

When I started at Heroku, my manager knew the codebase better than I did, knew
Ruby better than I did, and pushed more commits in a day than I would do in a
week. During our planning sessions we'd sketch in broad strokes on how certain
features or projects should be implemented, and leave it up to the
self-initiative of each engineer on the team to fill in the blanks. There
wasn't the time or the interest for micromanagement.

We eventually moved to a place where a virtuous manager was one who didn't
commit code, wasn't on the pager rotation, and never looked at a support
ticket (i.e. probably the situation that most big organizations have). But
although technical management wasn't an idea that lasted, it was a very good
place to be an engineer while it did.

[api-design-guide]: https://github.com/interagent/http-api-design
[bazooka]: http://gotocon.com/dl/goto-zurich-2013/slides/AlexanderSimmerl_and_MattProud_BuildingAnInHouseHeroku.pdf
[degges-12factor]: http://www.rdegges.com/heroku-isnt-for-idiots/
[empire]: https://github.com/remind101/empire
[ghi]: https://github.com/stephencelis/ghi
[heroku-api]: https://devcenter.heroku.com/articles/platform-api-reference
[hub]: https://github.com/github/hub
[maciek]: https://twitter.com/uhoh_itsmaciek
[pliny]: https://github.com/interagent/pliny
[postgres-queues]: /postgres-queues
[slack-distractor]: http://www.guilded.co/blog/2015/08/29/slack-the-ultimate-distractor.html
[twelve-factor]: http://12factor.net/
[wiggins-values]: https://gist.github.com/adamwiggins/5687294

[1] Thank-you [Maciek][maciek] in particular for stepping in and helping out
    with my Postgres woes way more often than you should have.
