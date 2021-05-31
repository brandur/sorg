+++
image_alt = "Redwoods in Heroes Grove, Golden Gate Park"
image_url = "/photographs/nanoglyphs/025-logs/heroes-grove@2x.jpg"
published_at = 2021-05-22T21:13:01Z
title = "Canonical Log Lines 2.0"
+++

I'm about a month into my new job -- still at a point where everything feels fresh and exciting, and not long enough to have accumulated even an ounce of jadedness. So far it's exactly the change I was looking for. Life is less frantic at a smaller company, and yet the feeling of relative calm is misleading. It seems intuitively like a more relaxed environment would result in commensurately less output, but it's actually the opposite, with individual people able to do much more. My time's flipped from a previous ratio of 90/10 busy work/development to an almost perfectly inverted 10/90, with the 90 spent building things. It's been fantastic.

There's a lot to talk about -- self-correcting transaction-based systems in Postgres, organizing moderately large Go projects, best practices for HTTP APIs in 2021, and more. I'd originally started baking it all into one very long post before realizing that it would need its own table of contents and glossary. Instead, I'm breaking it up, with this issue containing some highlights from the first month and some reflections on canonical log lines.

As usual, I'm Brandur, welcome to _Nanoglyph_, a newsletter about service toolkits and vintage bike messenger photography.

---

Back at Old Heroku, we were some of the original purveyors of microservices. Knowing what we know today, that's something of a dubious credential, with the industry as a whole having gone through an entire expansion and contraction phase since, initially buying into the idea wholesale before overcompensating with a hard turn into skepticism as the downsides of a fragmented data model and major coordination overhead became more evident, before finally adopting a more neutral outlook that many of us hold today.

We practiced what we preached, with separate systems for routing, dyno lifecycle management, Postgres, user-facing dashboard, billing, and orchestration, all chattering to each other at hyper frequency to keep the cloud healthy and alive. We invented names of endearment for services and their interfaces boundaries alike -- Hermes, Shushu, Lockstep, KPI (kernel platform API), Midgard -- and they all spoke the lingua franca of the internet -- HTTP.

I worked on a team called "API" which managed a component called `api` which acted as, you guessed it, an API, specifically the HTTP interface used by Heroku's dashboard and CLI. A command line invocation of `heroku create myapp` translated into a `curl -H "Accept: application/vnd.heroku+json; version=3" -X POST https://api.heroku.com/apps` as it left your terminal.

Some of our closest collaborators were the DoD (not the real one [1]), who ran a component called "Shogun", powering Heroku Postgres, and later Heroku Redis and Kafka. Shogun could be thought of as a state machine with an HTTP API, designed to create databases and manage their lifecycles through provisioning, maintenance, failover, recredentialing, and eventual retirement. Just like a 3rd party addon provider, when a database needed provisioning, the API would reach out to Shogun, ask for a new resource, and inject the connection string it responded with into the environment of an app. It probably seems like Shogun would have been managing _big_ state even if not _enormous_ state, and that probably would've been the case if not for Heroku's generous free tier, which along with free dynos, gave a free database to anyone asked for one. A few years into the Cedar stack's lifespan, free tier databases were as plentiful as distant stars in the night sky, winking in and out of existence every second of every day, and guaranteeing Shogun permanent job security.

<img src="/photographs/nanoglyphs/025-logs/heroku-sign@2x.jpg" alt="Heroku sign at 650 7th" class="wide">

---

Joining Crunchy, the setup I've inherited is eerily similar to what I used to work in back then, and even involves some of the same people.

With Crunchy Bridge, a frontend written in TypeScript and React is itself stateless, and delegates operations to a backend API written in Go on Postgres, and responsible for keeping track of everything from users to provisioned clusters to subscriptions. And similarly to Heroku, the platform in turn delegates cluster management to yet another backend service, called Owl. Like Shogun, it's primarily a state machine, initiating operations, then tracking them through to completion. The service responsible for provisioning Postgres databases is of course itself built on a Postgres database, and it leverages transactional guarantees to make sure that no state is ever lost and no action ever skipped.

---

My job's to help run and develop the middle platform tier. It's a mature codebase that runs a production service, but everything exists on a continuum, and after leaving Stripe where over the years we wrote code to handle every conceivable edge (and that poor code showed it, with millions of if/else conditionals and complications to handle countless special cases), it feels young.

Certainly there was plenty to do in my first couple weeks, with most of my work put toward getting us up and running with a basic service toolkit, taking the platform from a running program to an _observable_ running program that's conducive to debugging production and tells us about problems when they happen. A few features that I got installed:

* **Sentry reporting:** Rewiring endpoints to pass errors back to a central location where they can be contextualized and reported to Sentry, including a feature I've wanted for years: a clickable Sentry link generated right in the log line that takes you right to the error's UI. (A decade later and millions of dollars spent on Observability at `$lastJob`, you still can't easily get from a log line to a Sentry backtrace.)

* **Request-level transactions:** Existing code was leveraging transactions in a few places, but after so many years under the yoke of Mongo I'm _very_ keen on getting as much of any particular API request wrapped into transactions as possible to avoid accidentally invalid state. So far I've introduced a basic framework for winding a top level transaction down into all the DB operations of an endpoint so that it can execute as one atomic unit. (Did I mention how good it feels to be working with a real database again?)

* **API errors:** A framework for returning errors with rich flavor that helps the caller understand what went wrong, is succinct to code in the implementation, and makes a strong distinction between errors that should be user-facing versus those that shouldn't.

* **Distributed tracing:** We got Platform and Owl's communication set up on Sentry's distributed tracing platform. We can get a lot of the same information out of logs, but it's still pretty to be able to visualize it. Because both traces and errors are in Sentry, if an error occurs it appears right in the Sentry trace, and an operator can easily link straight to the backtrace.

<img src="/photographs/nanoglyphs/025-logs/sentry-trace@2x.png" alt="Distributed Sentry trace" class="img_constrained">

* **Structured logging:** Added structured logging and a canonical log line to the project, and tried JSON instead of logfmt for the first time. It's great. More on this below.

* **GitHub Actions:** Amazingly, the tests were in good shape despite no CI process. I've been a bit of a GitHub Actions fanatic over the last couple years, and adding some for the project was one my first moves. Linting, testing, executable start tests, and migration checks all run in separate parallel jobs. The entire workflow takes one minute to run, meaning that between the time you push a branch and navigate to the pull request in your browser, it's almost always already done.

    It is possible to write a thorough test suite which is also fast. Don't let anyone tell you differently. (At `$lastJob` again, we had many _single_ test cases took more than a minute to run, but worse yet, "staff" engineers who insisted that a test suite which would take a week to run if not for massive parallelization was an inevitable byproduct of success, a sentiment from which I'm still experiencing PTSD.)

* **Graceful restarts:** We're running on Heroku, where dynos are restarted once a day, or sooner if the application is deployed, so it's important that programs are able to handle restarts gracefully, waiting for existing connections to finish up before exiting, similar to what Puma or Unicorn does for you in Ruby. Luckily, recent augmentations to Go's HTTP server make this [absolutely trivial](https://golang.org/pkg/net/http/#example_Server_Shutdown) to implement.

* **API Endpoints:** A downside of Go is that it puts heavy emphasis on HTTP endpoints being nothing more than callable functions (see [`http.Handler`](https://pkg.go.dev/net/http#Handler)), which doesn't make it easy to annotate them with metadata. I'm experimenting with the introduction of a light API framework that gives endpoints enough information to be able to reflect themselves into an OpenAPI description. Go's lack of generics doesn't make this particularly easy to do, so getting it right is still an evolving effort.

In terms of mostly meaningless stats, I've touched about 40k lines so far:

<img src="/photographs/nanoglyphs/025-logs/lines-changed-2@2x.png" alt="Number of lines changed" class="img_constrained" style="width: 350px;">

It's been cathartic. Between overhead, development tools that more closely resembled sharpened sticks than anything built in the 21st century, red tape, and deeply engrained cultural conservatism, at `$lastJob` I was lucky to ship a few hundred LOC on a good week, so this is a much needed 180 degree respite.

Platform and Owl are two of the cleanest codebases I've ever had the pleasure to work in professionally, and contributing to them has reminded me that it _is_ possible to have a job in tech where the technology is actually fun to work with, something that I'd started to forget over the last few years. If you're stuck in a similar position, and have spent your day implementing a tiny feature that should have taken five minutes but took five hours, squashing a bug that only existed in the first place by virtue of overwhelming systemic complexity, or debating people in 37 different Slack threads to get to a consensus that's 3% better compared to if someone had just made a decision themselves, then I don't know, maybe think about that.

---

## Canonical log lines 2.0 (#canonical-log-lines-2)

At Stripe, I described an idea called [canonical log lines](https://stripe.com/blog/canonical-log-lines), which I still stand by is by far the single, easiest method of getting easy insight into production that there is. (I should note that these were not originally my idea, but also that I don't know who's they were.) Starting at Crunchy, I wanted something similar _immediately_, and started coding them on day one.

They're a dead simple concept. In addition to normal logging made during a request, emit one, big, unified log line at the end that glues together all the key vitals together into one place:

> `canonical_log_line http_method=GET http_path=/v1/clusters duration=0.050 db_duration=0.045 db_num_queries=3 user=usr_123 status=200`

When operating production, you're _occasionally_ interested in specific requests, like if one produced a new type of error that needs to be investigated. But much more often you're interested in request _patterns_. Like, are failing more requests now than an hour ago? Does increasing DB latency suggest that Postgres might be degraded? Is one single user hammering us inordinately hard?

This is where canonical log lines come in, and where they come in to perform magically. Along with being a one stop shop for seeing every important statistic of a request in one place, they're also designed to be as friendly to aggregates as logs can possibly be. Combined with a system like Splunk, I can answer all of the questions above with ad-hoc queries like these:

> `canonical_log_line earliest=-2h | timechart count by status`

> `canonical_log_line | timechart p50(db_duration) p95(db_duration) p99(db_duration)`

> `canonical_log_line earliest=-10m | top by user`

Canonical log lines are the one tool that I used reliably every single day of the 5+ years I worked at Stripe. It might seem hard to believe if you've never used something similar, but they really are _that_ good.

This time, since we weren't using Splunk, and since I was responsible for setting up our logging, I decided to try something a little different. Instead of old trusty [logfmt](/logfmt) (the key/value pairs above that look like `http_path=/v1/clusters duration=0.050`), I configured our logs to be emitted as fully structured JSON.

I've always argued that logfmt is a great operational format because while it's machine parseable, it's also relatively friendly to humans, whereas JSON is clearly not. But this matters somewhat less if your logging system is going to be parsing it all for you anyway. For most types of logs, LogDNA (our current Splunk alternative) will just echo them back to you. It'll parse both logfmt and JSON logs when it detects them so that the metadata is available for aggregates, and usually just echo those back too. However, if a JSON log contains the special key `message`, it'll show just that on the display line, with the other information available by expanding it. This leads to a _very_ clean log trace:

<img src="/photographs/nanoglyphs/025-logs/log-clean-trace@2x.png" alt="Clean trace of log lines" class="img_constrained">

We emit `message` as a very terse human-readable explanation of a request's result:

> `canonical_log_line GET /api/v1/teams -> 200 (0.004936s)`

But all the information is still there when we need it, and in a much more human-friendly way:

<img src="/photographs/nanoglyphs/025-logs/log-expanded@2x.png" alt="Log line expanded" class="img_constrained">

Because of their hidden-by-default nature and clean visualization, we can do things that might not otherwise be a good idea. You can see in the image above that I'm dumping a `sentry_trace_url` for every line. Normally that'd be a lot of visual noise in the logs, so I'd avoid it, but since that's not a problem here, it's nice to provide an easily-copyable URL to visualize the request's trace right in Sentry. If the request ended in error, a `sentry_error_url` also gets dropped in so that I can easily go see the full back trace.

Maybe most importantly, all that data is there behind the scenes in a machine-readable format, so aggregates are still possible:

<img src="/photographs/nanoglyphs/025-logs/log-aggregates@2x.png" alt="Aggregating logs" class="img_constrained">

An even more powerful feature unlocked by JSON is the capacity for multi-level nested structures. You probably don't want to go too overboard with this, but as an example, we've found them nice to track every backend service call that occurred an API request along with its interesting vitals (see `owl_requests`):

<img src="/photographs/nanoglyphs/025-logs/log-nested-structures@2x.png" alt="Nested structures in logs" class="img_constrained">

(And while the rich structures are nice, we still include "flattened" information as well like `owl_duration` and `owl_num_requests` for easier/faster use in aggregates.)

In previous environments where I've worked, getting this information probably would've been possible, but it would've been like pulling teeth, and often stressful too because you're usually only trying to figure out how to get it during an operational emergency. Here, we don't just make it _possible_ to find, we make it easy.

## 90s cool (#90s-cool)

Alright, enough about tech. Today, I'll leave you with [this link](https://www.35mmc.com/25/02/2021/street-portraits-of-bike-messengers-standing-by-with-the-rollei-35-se-trevor-hughes) to some incredible photography of bike messengers working in downtown Toronto of the 90s by Trevor Hughes, presented at high resolution and with minimal cruft. Great subjects, a fascinating moment of history, and beautiful photography at at time when beautiful photography was hard.

This isn't a scene that I ever actually experienced, and yet it still hits me straight to the gut with powerful nostalgia. The grunge style of the 90s reaches seemingly unattainable levels of je-ne-sais-quoi cool, there's a certain romance of tight-knit communities of young people living in inner cities, and nary a smartphone in sight. Not only did people still work together in buildings back then, but _physical objects_ needed to be rush couriered between places, something that seems so quaint today amongst our interminable flood of JavaScript and email.

Until next time.

[1] DoD = Department of Data.

<img src="/photographs/nanoglyphs/025-logs/03_93_27_04_BillyJohn@2x.jpg" alt="03_93_27_04_BillyJohn" class="wide">

<img src="/photographs/nanoglyphs/025-logs/04_96_35_20_HandsomeDave@2x.jpg" alt="04_96_35_20_HandsomeDave" class="wide">

<img src="/photographs/nanoglyphs/025-logs/05_93_18_22_Joe_and_messenger@2x.jpg" alt="05_93_18_22_Joe_and_messenger" class="wide">

<img src="/photographs/nanoglyphs/025-logs/12_92_65_02_Chris@2x.jpg" alt="12_92_65_02_Chris" class="wide">

<img src="/photographs/nanoglyphs/025-logs/17_93_48_09_Anita@2x.jpg" alt="17_93_48_09_Anita" class="wide">

<img src="/photographs/nanoglyphs/025-logs/21_95_49_06_Manny@2x.jpg" alt="21_95_49_06_Manny" class="wide">
