+++
image_alt = "Herald Square in New York"
image_url = "/photographs/nanoglyphs/033-heroku/herald-square@2x.jpg"
published_at = 2022-05-02T15:01:18Z
title = "Heroku Redux"
+++

Last week one my colleagues [sent out a tweet](https://twitter.com/craigkerstiens/status/1519483444861935616) with the not-so-simple question of, "was Heroku a success or failure?"

It got me thinking. I remember a few years back having a conversation with one of my ex-colleagues from the company where I said something (probably unwisely) that implied that Heroku had been a failure, and which was met with open astonishment as the other party had internalized precisely the opposite impression -- to him, it as an unmitigated success. Reaction to the Twitter question was similarly split, with lots of engagement in both directions.

Selling to Salesforce for $212 million was an obvious win, but that's balanced by a product that's broadly been on cruise control since 2012, and which hasn't seeded much in terms of product longevity or lasting industry technology.

## The success argument (#success-argument)

Most obviously, with an acquisition of that size, some people got rich, and even for those who didn't make anything off the sale (yours truly included), it still gave us a couple of great years working in an environment that was startup-like in terms of feel and innovation, but with the benefits of big tech salaries and largesse.

There's also the matter of Heroku being surprising sticky. Given a product that's gone largely unchanged for years now, along with a host of new entrants in the market, I would've assumed that Heroku would've been resigned to obscurity by broader cloud competition years ago, but to this day, it's still plausible as a platform. At my company, we made the decision to host on it as little as a year ago -- we know the product well, there's minimal lock-in in case we need to evacuate, and it keeps us generally hands-free on operations/infrastructure not core to the services that we're selling. Although every major cloud provider's introduced new services aimed at satisfying that PaaS tier (and in cases like AWS, more than one), so far few have been able to match Heroku's streamlined workflows and simplicity.

Beyond that, a lot of great things fell out of Heroku:

* **Outsourcing infrastructure:** For the longest time, deploying programs to the internet was hard. Then PHP came along, and won the world not just through concise syntax and a simple deployment story, but one with gaping holes. Deploying generic stacks was hard -- Rails at the time involved setting up a load balancer, per-server reverse proxy, CGI process, and whatever else it took to keep everything monitored and alive. Heroku vastly simplified this story, letting developers focus on building programs instead of configuring infrastructure. An obvious good in today's container world, but one which wasn't at the time.

* **Postgres:** The rise in prominence of Postgres over the last decade is owed to a number of factors including great core advances and a relative languishing in its competitors, but by making it a core piece of the platform offering and giving it a lot of high profile publicity, Heroku was an important part of the equation.

* **Containers:** Few remember it, but Heroku ran containers before they were cool, using [LXC](https://linuxcontainers.org/) as a central technology in its Cedar stack.

* **DX and CLIs:** The Heroku CLI was a hugely important part of the product right alongside Git itself. The two tools in conjunction let a user manage every aspect of an app from the command line, which was very novel for the time. The CLI planted some of the seeds that would eventually grow into DX, which is now a dedicated branch of the tech industry.

* **Buildpacks:** Generic formulae for how to deploy an application written in a particular language, buildpacks were a precursor to the `Dockerfile`, and arguably, a more appropriate level of abstraction. Custom buildpacks were supported from the earliest days of the Cedar stack, letting users run any technology they wanted to bring.

* **Review apps:** Ephemeral applications that let users easily interact with development versions of their products. Certainly not a unique idea, but Heroku was well-placed to make these available to anyone, and easily.

That's a pretty impressive list -- even one or two of these would be more of a mark on the world than most tech companies _ever_ make. But there's also a common undercurrent to most these items -- although they're all great ideas and will make a lasting impression into the future in how services are deployed, none of them resulted in lasting residuals for the Heroku product itself -- other platforms captured the concepts and took the proceeds, and even with commercial aspect aside, no specific technologies will be attributed back to Heroku. Even though Docker as a company might be doomed to failure, it'll be remembered as the progenitor of container-based deployment for decades. Future history about the 2010s will talk about the evolution of Docker to OCI, but Heroku will be a footnote at best.

Heroku was the ultimate ideas factory -- concepts like 12-factor, erosion resistance, and DX will stand the test of time, but few of their beneficiaries will recognize their lineage back to Heroku.

---

## Imagination v. reality (#imagination-reality)

Not much lasting product or technological impact is one side of the coin, the other being disappointment around an epic vision with so much potential that never materialized.

The Cedar stack was a work of true genius (I had no hand in its creation, so this isn't as self-serving than it sounds). The previous Aspen and Bamboo stacks were far more limited, supporting only specific versions of specific stacks, and with a lot of special-casing necessary. Cedar made Heroku the platform that could run everything -- users could bring their own stacks with buildpacks and `Procfile`s, and its sophisticated internal state machine and routing layer made apps running on it impressively robust. Even after learning how the sausage was made, I never stopped being impressed by how well it worked, or how the platform could be pushed to do things that its creators never imagined thanks to incredible flexibility.

Back in 2012 the momentum coming off shipping Cedar was so great that despite its success, it was only considered the first step of a much more ambitious project. Before long, it'd be extended to handle programs of all shapes and sizes, with the current 512 MB container just an incidental first option. Even the biggest data-crunching apps would be deployable on containers with 10s of GBs of memory, and all the way down to tiniest one-off cloud `grep` runs needing only just a couple MBs of memory.

It'd become modular. The shared router fleet was an adequate option for most uses, but large users might want to bring their own routing implementations to avoid being in a shared cloud, or to provide their own highly customized routing configurations. Even in the Heroku "kernel" would be swappable so that you could still have Heroku building, orchestrating, and monitoring your apps, but those apps would be running on your own dedicated, single-tenant servers.

### The Self-hosting Singularity (#self-hosting-singularity)

The Heroku cloud would become so extensible and so robust that just like a self-bootstrapping language compiler, it'd be able to host itself. Core components like the API, state machine, and router would run as Heroku apps and gain all the DX ergonomics and robustness of such. This optimistic and ambitious vision was coined, _The Self-hosting Singularity_.

It'd be the anti-AWS. Whereas AWS exposes every possible primitive to a new user the first time they log in -- thousands of confusing and overlapping concepts -- the Heroku vision would've been to expose none of it to a new user. They'd be started with a basic `git push heroku master` and single dyno app, but as their software developed and their requirements became more complicated, new primitives would've been progressively exposed as they were needed -- VPCs with ingress/egress rules, configurable hosts with alternative base images or architectures, SSH access, static IPs, etc. An onion that could be peeled back layer by layer.

There were some other things too. [12-factor factor no. IV](https://12factor.net/backing-services) ("backing services") describes "attachable resources", persistent services like databases that live as isolated resources which can be attached and detached to and from more ephemeral apps at will. It took us years to ship this feature, and although it was pretty neat when we did, by then Heroku's golden days of product leadership were behind it, and we didn't make much in the way of inroads into convincing anyone else as to why it was a good idea.

Pricing was another elusive beast. The cost of jumping up from the free tier to a paid app was a big step up which users had complained about it since the product's inception. Eventually, a new pricing model did ship, but didn't do much to make the product less expensive or users more happy.

---

## Why we fail (#why-we-fail)

So what happened? All the foundations for success were in place, so not achieving its ambitious vision wasn't an inevitability.

* **Operation mire:** After Cedar went in, frequent product outages caused by both factors out of our control (`us-east-1` was particularly horrible in those days) and those within it (for a while there we seemed to have a bad deploy go out every other day) were escalating to a point of becoming an existential liability. Product work was deprioritized in favor of shoring up out operational story -- putting in metrics, alerts, safe deploy processes, and just generally building operational muscle.

* **Product cadence:** Especially in those earlier days, there was no institutional framework for shipping new features. It was possible, but generally involved sending pull requests yourself or pitching specific people to help make changes. Even where incentive to push a new feature was strong, that incentive tended to fizzle out over organizational boundaries -- we had some infamously degenerate cases like the organizations feature being built on top of the core API as a separate microservice because there was no mechanism to make it happen in a more integrated way.

* **Docker tunnel vision:** The first iteration of Docker shipped with such fanfare and widespread interest that a lot of us developed what I'd say in retrospect was an unhealthy fixation on it. We internalized a fairly defeatist attitude that Docker containers were the future, and what we were doing was the past.

    To some degree this was correct, but what we should've seen at the time was `Dockerfile`s were still a very low level of abstraction -- one so low as to be somewhat undesirable. What we see a lot today is container technology making up the basis of a lot of deployment stacks, but acting more as a primitive, with a lot of technology to improve their ergonomics layered on top. In many ways buildpacks were a better abstraction layer for application developers -- instead of writing `Dockerfile`s for everything, they can just use the tooling common to their stack like `Gemfile`, `Cargo.toml` or `go.mod` and have the build process figure out how to bake that into a deployable image. From there, if say the base layer needs to be updated or the minor/patch-level of a programming language needs to be updated, it can be done so broadly without having to tweak every project's `Dockerfile`.
		
    Heroku later adopted more general container technology, which was the right move, but the whole time we should've been working towards doing that while not losing sight of our own strengths.

* **Next stack fixation:** Heroku stacks were named after trees: Aspen, Bamboo, Cedar. Cedar was a quantum leap over Bamboo, and most of us took it for a given that our next goal was to build a stack that was as much better than Cedar as Cedar had been compared to Bamboo. Lofty ambitions are generally laudable, but in this case they planted a subliminal seed that Cedar was the past, which disincentivized major investment in it.

    In retrospect, we see based on the convergence of available technology today that there _probably isn't_ a stack that's as much better than Cedar as Cedar was to Bamboo. It would've been better to concentrate on incremental improvements to Cedar rather than a panacea that was always somewhere over the horizon.

* **Ideator/operator divide:** Made possible by the dynamics of being a well-funded small company inside of a big company, for a while we had a fairly unique situation of employing a number of people who spent their time experimenting, prototyping, and coming up with ideas, almost like a tiny Bell Labs or Xerox PARC within the company. On the other side of the fence we had a bunch of hardened operators, who generally spent their time so deep in operational concerns that they rarely had their heads above water. Generally speaking, the ideators had no ability to put anything into production and make their ideas manifest, while simultaneously the operators had no spare effort or time to make substantive product improvements. This led to a lot of cool internal demos, but also a predictable bias towards product flatline.

---

## The cloud graph (#cloud-graph)

Three very different things that I'd recommend you check out this week:

* [Why the past 10 years of life have been uniquely stupid](https://www.theatlantic.com/magazine/archive/2022/05/social-media-democracy-trust-babel/629369/) ([archive](https://archive.ph/j2U17)) — In which Jonathan Haidt argues that the virality of social media is incompatible with a functional democracy, with the most extreme voices in society given a platform, and the wide middle cut out entirely.

* [6 months of Go](https://typesanitizer.com/blog/go-experience-report.html) — Extensive treatise on Go from an engineer who previously worked on the Swift compiler. Starts with positives before going into a long list of entirely fair language critiques. Some articles like this come off as smear pieces. This one doesn't.

* [What is a major chord](https://www.jefftk.com/p/what-is-a-major-chord) — I once read literally _Music Theory for Dummies_ and despite taking extensive notes at the time, could not give you a single takeaway from that book all these years later. This extremely succinct article introduces the note as a frequency before walking through scales, major scales, chords, and major chords. I've never gotten so close to actually understanding these concepts.

<img src="/photographs/nanoglyphs/033-heroku/greenwich-st@2x.jpg" alt="Greenwich St intersection" class="wide" loading="lazy">

## On the resilience of cities (#resilience-cities)

I was in New York a few weeks ago. The first time I visited this city 

<img src="/photographs/nanoglyphs/033-heroku/brooklyn-cafe@2x.jpg" alt="A Brooklyn cafe" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/033-heroku/highline@2x.jpg" alt="The Highline" class="wide" loading="lazy">
