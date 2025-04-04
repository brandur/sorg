<div class="my-16
            ml-[calc(-1rem)] mr-[calc(-1rem)] w-[calc(100%+2rem)]
            lg:mr-[calc(-1rem-75px)] lg:ml-[calc(-1rem-75px)] lg:w-[calc(100%+2rem+2*75px)]
            xl:mr-[calc(-1rem-150px)] xl:ml-[calc(-1rem-150px)] xl:w-[calc(100%+2rem+2*150px)]
            ">
    <img src="/photographs{{DownloadedImage .Ctx "/now/2025-apr-04" "https://www.dropbox.com/scl/fi/t8clk35k7peh3jo49255c/L1000718.heic?rlkey=jdbez5jzln6ju6jl5ukfbhoop&dl=1" 1300}}" alt="Sky blue" class="lg:rounded-lg w-full" loading="lazy">
</div>

I'm back in San Francisco after a series of trips to Seattle, Kelowna, and Reno.

Some haphazard notes:

* After running into the first honest-to-god transaction anomaly that I remember ever seeing in production, I'm building locking (`SELECT ... FOR UPDATE`) support into the lightweight, home grown data loading framework that we use in Crunchy's API.

* I'm been journaling daily for 500 consecutive days now. 128,000 words so far in 2025.

* We recently wrote and released [OpenTelemetry middleware for River](https://github.com/riverqueue/rivercontrib). Most major providers have OpenTelemetry support these days, so it's usable with DataDog or Sentry for example.

* After noticing how inefficiently packed the JPGs coming out of my cameras were (and how camera sensors are getting denser but laptop hard drives are staying the same size), I wrote a script to [archive photography artifacts in more compact form](/fragments/optimizing-jpegs-for-archival). I'm debating whether I should make the jump to storing everything in HEIC since it's fewer moving parts to get to optimal storage and I'd be able to avoid messing around with esoteric tooling like MozJPEG.

* I finally took the step to black hole Reddit on all my computers. Focus on making cool stuff instead of wasting precious hours of mortality wrestling down in this filthy, hyper-partisan sludge. This company's stock cannot go to zero fast enough (and at -48% YTD, it's on its way).

<div class="mb-16 not-prose">
    <p class="font-serif italic my-1 leading-normal not-prose text-sm tracking-tight">This page was last updated on
        <span class="font-bold">Apr 4, 2025</span>.
    </p>
</div>

<details>
    <summary class="font-semibold font-serif italic list-outside pl-1 text-sm">Dec 19, 2024</summary>
    <div class="my-10">

<div class="my-16
            ml-[calc(-1rem)] mr-[calc(-1rem)] w-[calc(100%+2rem)]
            lg:mr-[calc(-1rem-75px)] lg:ml-[calc(-1rem-75px)] lg:w-[calc(100%+2rem+2*75px)]
            xl:mr-[calc(-1rem-150px)] xl:ml-[calc(-1rem-150px)] xl:w-[calc(100%+2rem+2*150px)]
            ">
    <img src="/photographs{{DownloadedImage .Ctx "/now/7th-ave" "https://www.dropbox.com/scl/fi/l3kduxtlns6pzjlmo4zns/L1000513.JPG?rlkey=2dzclgpezw7ei2px864afiky2&dl=1" 1300}}" alt="7th Ave" class="lg:rounded-lg w-full" loading="lazy">
</div>

I'm in Calgary for the holidays. Let's see:

* Upgraded my M2 MacBook Air to the new M4 MacBook Pro with nanotexture. These computers don't get major changes very often, so right after release is a good time to buy. I got the non-Pro non-Max chip variant because it has the best battery life, advertised at 24 hours, which is crazy.

* Tied up my last couple major projects for the year at work:

    * Revamped our ACL layer to put in support for a flexible way to organize access to a team's resources by its members called an _access group_. We're still finishing up the frontend work on this project, but the API's finished.

    * Using S3's event streaming and inventory APIs, built a way to bill for S3 usage in a shared bucket for Crunchy's new warehouse product. This is more complicated than it sounds because S3 objects aren't just billed by size, but rather size _over time_. The billable unit is a GiB/mo rather than GiB.

* Next up will be to improve the versability of our saved queries product by adding UI configurable input parameters to queries.

* We're still plugging away [on River](https://github.com/riverqueue/river), the best Go/Postgres queue out there.

<details>
    <summary class="font-semibold font-serif italic list-outside pl-1 text-sm">Oct 19, 2024</summary>
    <div class="my-10">

<div class="my-16
            ml-[calc(-1rem)] mr-[calc(-1rem)] w-[calc(100%+2rem)]
            lg:mr-[calc(-1rem-75px)] lg:ml-[calc(-1rem-75px)] lg:w-[calc(100%+2rem+2*75px)]
            xl:mr-[calc(-1rem-150px)] xl:ml-[calc(-1rem-150px)] xl:w-[calc(100%+2rem+2*150px)]
            ">
    <img src="/photographs{{DownloadedImage .Ctx "/now/skyscrapers" "https://www.dropbox.com/scl/fi/hwt1dm93omj4l2qhuppww/2W4A6069.JPG?rlkey=w3cj9fbkgvk0ohdn6b34x0eod&dl=1" 1300}}" alt="Skyscrapers" class="lg:rounded-lg w-full" loading="lazy">
</div>

I'm in San Francisco, having recently returned from [Rails World in Toronto](/fragments/rails-world-2024), which was a lot of fun. Beyond that:

* I finished a massive access control refactor that builds out a more flexible RBAC (role-based access control) system for Bridge. It doesn't sound much, but the roots of ACL run deep, and it ended up touching every important file in project, in the end taking me multiple months to fully complete.

* Next up is a billing system for metered storage use. Harder than it sounds on the face because you're not just tracking storage, but _storage over time_. If someone stores 10 TB for most of the month but deletes every last byte before invoicing day, they should still be charged for about 10 TB. If someone spikes up to 10 TB for only a single day before deleting it all, they should be charged far less. Rather than charging per GB, you're really charging per GB-hour.

* I'm trying, and mostly failing, to update this site more consistently. The notification came in to update this page on Oct 1st, and yet, *sigh*, I'm doing it on the 19th.

</div>
</details>

</div>
</details>

<details>
    <summary class="font-semibold font-serif italic list-outside pl-1 text-sm">Sep 4, 2024</summary>
    <div class="my-10">

It's September, although in California, it's hard to notice.

I was in Washington DC last week for a company off-site, a once a year event where our team converges in one location. My first time in the city, I toured the length of the National Mall and its supremely iconic landmarks. At almost 40C and that incredible east coast humidity, it was a sweaty affair (but neat!).

Bridge is going well. We shipped a lot this past year, and gotten to a point where it's quite a mature product. As always, our core competency is hosting Postgres, which implies stability, and to some extent being boring and predictable, but our platform and analytics teams are working on big, ambitious projects, so it feels like we've got a couple moonshots in flight.

Years in, our code is as clean and stable as any project I've ever seen of this age, and 100x better than many. I just ran our test suite on my underpowered, commodity laptop, and we're sitting at 4237 tests running in 27.595s, which is great. Last week I pushed through a big test refactor (+41,168 LOC, −29,160) so that not only are all tests marked as pararallelizable, but all _subtests_ of those tests too. We're on Go 1.23 of course, and refactor regularly so that convention is strong and everything is in ship shape.

I'm not working on AI, and indeed may be the last person in San Francisco (and soon, the world?) who's never typed anything into an AI prompt.

</div>
</details>

<details>
    <summary class="font-semibold font-serif italic list-outside pl-1 text-sm">Aug 11, 2024</summary>
    <div class="my-10">

<div class="mb-16 mt-8
            ml-[calc(-1rem)] mr-[calc(-1rem)] w-[calc(100%+2rem)]
            lg:mr-[calc(-1rem-75px)] lg:ml-[calc(-1rem-75px)] lg:w-[calc(100%+2rem+2*75px)]
            xl:mr-[calc(-1rem-150px)] xl:ml-[calc(-1rem-150px)] xl:w-[calc(100%+2rem+2*150px)]
            ">
    <img src="/photographs/now/2024-07-salesforce-tower.jpg" alt="Salesforce Tower" class="lg:rounded-lg w-full" loading="lazy">
</div>

## Pen testing, Q3, DC

It's August. I don't have anything of substance to write about, so here's a couple short points instead:

* We're currently running a pen test with an external contractor. They're the best one we've worked with so far, and a number of legitimate liabilities have fallen out of it. I've spent the last few weeks shoring up the walls.

* My main macro project continues to be rebuilding our roles system. There's been a lot of smaller distractions, and refactoring existing code to make what I'm envisioning possible has taken longer than I expected.

* I've acquired a new toy, a Leica Q3. I finally ordered one the week before prices were set to increase. Leicas are bad deals and I'm old enough now to know very definitively that new consumer products will never make you happy, but what the hell, I hadn't bought anything substantial in a long time.

* I'm headed to Washington DC in a few weeks, the first time I'll have ever visited the nation's capital. I've heard mixed things, and honestly can't say whether I'll love or hate it.

* I've been playing way too much [_Elden Ring_](/fragments/elden-ring), a game I thought I'd hate like other FromSoft games, but which I now acknowledge is a masterpiece of both worldbuilding and gameplay design.

</div>
</details>

<details>
    <summary class="font-semibold font-serif italic list-outside pl-1 text-sm">Jul 8, 2024</summary>
    <div class="my-10">

## RBAC, Python (#rbac)

I'm in San Francisco.

I'm working on rebuilding Bridge's RBAC (role-based access control) system, taking inspiration from [Tailscale's substantial write up on the subject](https://tailscale.com/blog/rbac-like-it-was-meant-to-be), which seems to be the most contemporary thinking on the subject from practioners who _also_ do a good job of it in their own product. Dozens of companies selling enterprise security solutions have strong opinions on the subject, but it doesn't inspire much confidence when their own offerings are of middling quality.

I've been relearning Python to help build a [River's Python client library](https://github.com/riverqueue/riverqueue-python). The language's been a mixed bag overall, but it's been interesting diving into typing, asyncio, and tooling like [Rye](https://github.com/astral-sh/rye), none of which existed the last time I worked in the language.

</div>
</details>

<details>
    <summary class="font-semibold font-serif italic list-outside pl-1 text-sm">May 5, 2024</summary>
    <div class="my-10">

<div class="mb-16 mt-8
            ml-[calc(-1rem)] mr-[calc(-1rem)] w-[calc(100%+2rem)]
            lg:mr-[calc(-1rem-75px)] lg:ml-[calc(-1rem-75px)] lg:w-[calc(100%+2rem+2*75px)]
            xl:mr-[calc(-1rem-150px)] xl:ml-[calc(-1rem-150px)] xl:w-[calc(100%+2rem+2*150px)]
            ">
    <img src="/photographs/now/2024-05-stayery.jpg" alt="Stayery" class="lg:rounded-lg w-full" loading="lazy">
</div>

I'm spending the month in Berlin, where I'm trying to run and write every day, and enjoy time in a place that's less reminiscent of a zombie wasteland than my home city.

The next big project I'm tackling at work is Active Directory. That sounds about as fun as a root canal, but I take it as an interesting challenge. AD is a long in the tooth technology that's still in use by many of the biggest players in the industry (we even used it at Stripe!). How can we integrate it in such a way that it gets big users what they need, produces as little code as possible and as few headaches for us, and maximizes the yield in leverage we get out of the effort. For example, it might involve ignoring the low level AD APIs and integrating [SCIM instead](https://en.wikipedia.org/wiki/System_for_Cross-domain_Identity_Management), thereby buying us compatibility with other non-AD SCIM-based systems.

Blake and I continue work on our open source Postgres job queue, [River](https://github.com/riverqueue/river), which I think is fair to say is the most full-featured in the Go ecosystem by some margin. There's something incredibly satisfying about taking a project scoped to a known, fixed domain, and refinining its code over and over again until it's _perfect_.  Trying to handle every edge, and with attention to detail on every line of code.

I recently published [Ruby gem](https://github.com/riverqueue/riverqueue-ruby) that enables job insertion in Ruby, but for jobs to be worked in Go, which is something that I'd always wanted back at Heroku and Stripe. I wrote about [my experience putting in type checking with Steep, and publishing RBS files](/fragments/ruby-typing-2024) for the project.

</div>
</details>

<details>
    <summary class="font-semibold font-serif italic list-outside pl-1 text-sm">Apr 9, 2024</summary>
    <div class="my-10">

I'm in San Francisco, where inertia keeps me rooted.

Work on Crunchy Bridge continues. As part of filling out a self-evaluation last week I scanned every pull request I've issued over the last year, and I liked what I saw. What we've shipped during that time is above and beyond any org I've worked at before. Small teams, agile tech stacks, and lack of a culture of objection-for-objection's-sake do wonders for productivity.

A few things I sent out the door recently:

* Multi-factor authentication, supporting WebAuthn (Yubikeys and biometric challenges like Touch ID) and TOTP (time-based one-time passwords). ([Notes](/fragments/lean-fast) on getting this shipped quickly.)
* An asynchronous query runner that's built to scale by taking advantage of Go's parallelism, by storing results to S3 instead of Postgres, and which prunes its results regularly.
* [Retired our use of Keycloak](/atoms/gkoxmy2), which involved many smaller tasks like adopting a [fancy new password hashing scheme](/fragments/password-hashing).

Pre-lockdown, I'd gotten into the best shape of my life by baking exercise into my schedule with a daily run commute, fitness which I unfortunately let languish. Newly armed with a WeWork pass and gym membership (for the showers), I'm bringing it back. A straight shot from the mountain down Market St to Embarcadero -- 50km/week if I keep it up.

Next month, Europe.

</div>
</details>

<details>
    <summary class="font-semibold font-serif italic list-outside pl-1 text-sm">Dec 8, 2022</summary>
    <div class="my-10">

I'm back home in Calgary for the holidays, staring into the precipice of 2023 which between money markets, strife, and war is shaping up to be a formidable year.

At work, we're aiming to build the best database-as-a-service in the world. I shipped more features over the last year than the previous five combined, and which were built into a robust stack that to this day has less tech debt than many two-week-old startups (I'm kind of proud of it, you might be able to tell). We have another aggressive roadmap for 2023, and I'll be doing my best to make sure that we don't slip.

I added a couple new sections to the site recently:

* [**Atoms:**](/atoms) Short multimedia particles minus the stress of a social media platform ([atom feed](/atoms.atom)).
* [**Sequences:**](/sequences) Periodic large format photos paired with prose. An older project, but one which I recently revived, flattened, and republished.

A few weeks before that I became somewhat enamored by the idea of [Spring '83](https://www.robinsloan.com/lab/specifying-spring-83/) and ended up [writing a server implementation](https://github.com/brandur/neospring) which is now in prod and [hosts my board](https://neospring.brandur.org/). I don't think Twitter is being displaced anytime soon, but these indy web projects are great.

In 2023: write, move, visit France.

</div>
</details>

<details>
    <summary class="font-semibold font-serif italic list-outside pl-1 text-sm">Apr 5, 2020</summary>
    <div class="my-10">

<div class="mb-16 mt-8
            ml-[calc(-1rem)] mr-[calc(-1rem)] w-[calc(100%+2rem)]
            lg:mr-[calc(-1rem-75px)] lg:ml-[calc(-1rem-75px)] lg:w-[calc(100%+2rem+2*75px)]
            xl:mr-[calc(-1rem-150px)] xl:ml-[calc(-1rem-150px)] xl:w-[calc(100%+2rem+2*150px)]
            ">
    <img src="/assets/images/now/twin-peaks-stairs.jpg" alt="Stairs up to Twin Peaks" class="lg:rounded-lg w-full" loading="lazy">
</div>

It's 2020. Like for almost everyone else on Earth, COVID-19 is top-of-mind. I'm working from home, San Francisco is sheltering in place, and the future is a hugely uncertain time.

As bad as our present day situation is, an indefinite work from home policy has given me more flexibility and more energy in my day-to-day than I've ever had in my adult life, and I'm going to do my best not to waste it.

Some things I’m working on:

* Write every day, and try to so fluidly. Instead of agonizing over every word, get content down, revise, and revise again. Some projects:

    * A (roughly) weekly newsletter called <em>Nanoglyph</em>. I’m challenging myself to send at least 30 editions in 2020, and do so without compromising content quality. You should <a href="https://nanoglyph-signup.brandur.org/">try subscribing</a>.
    * A development log with notes on daily software discoveries and projects. Most entries will be of minor interest, but frequent enough to build writing cadence. <a href="/fragments/google-cloud-run-deploy">For example</a>.

* Meditate every day.

</div>
</details>

<details>
    <summary class="font-semibold font-serif italic list-outside pl-1 text-sm">Dec 31, 2019</summary>
    <div class="my-10">

<div class="mb-16 mt-8
            ml-[calc(-1rem)] mr-[calc(-1rem)] w-[calc(100%+2rem)]
            lg:mr-[calc(-1rem-75px)] lg:ml-[calc(-1rem-75px)] lg:w-[calc(100%+2rem+2*75px)]
            xl:mr-[calc(-1rem-150px)] xl:ml-[calc(-1rem-150px)] xl:w-[calc(100%+2rem+2*150px)]
            ">
    <img src="/assets/images/now/calgary-snow.jpg" alt="Calgary snow" class="lg:rounded-lg w-full" loading="lazy">
</div>

I’m in Calgary for the winter break. It’s the last day of 2019 and we’re on the precipice of a new decade.

Some things I’m working on:

* Writing a (roughly) weekly newsletter called <em>Nanoglyph</em>. I’m challenging myself to send at least 30 editions in 2020, and do so without compromising content quality. You should <a href="https://nanoglyph-signup.brandur.org/">try subscribing</a>.
* Write more, and more fluidly. Instead of agonizing over every word, get content down, revise, and revise again.
* The craft of software has landed on a plateau where big, complicated, fragile backend deployments that support slow, underwhelming, buggy frontend products is the norm. Find techniques and ideas to reverse this trend and boost their signal.
* Ran 1000 miles and did 42k pushups in 2019, do so again in 2020. Keep weight at ~150 lbs. Zero in on ~10% body fat.
* Reboot meditation practice. Did pretty well in 2018, but fell off the wagon completely in 2019. Aim for a couple 30 consecutive day runs.

</div>
</details>

<details>
    <summary class="font-semibold font-serif italic list-outside pl-1 text-sm">Jun 5, 2019</summary>
    <div class="my-10">

<div class="mb-16 mt-8
            ml-[calc(-1rem)] mr-[calc(-1rem)] w-[calc(100%+2rem)]
            lg:mr-[calc(-1rem-75px)] lg:ml-[calc(-1rem-75px)] lg:w-[calc(100%+2rem+2*75px)]
            xl:mr-[calc(-1rem-150px)] xl:ml-[calc(-1rem-150px)] xl:w-[calc(100%+2rem+2*150px)]
            ">
    <img src="/assets/images/now/molecule-man.jpg" alt="Molecule man in Berlin" class="lg:rounded-lg w-full" loading="lazy">
</div>

I'm in Berlin.

A few points of focus:

* A <a href="/sequences-project">photography project called <strong><em>Sequences</em></strong></a> as an experiment to promote the independent web.
* A tiny static site framework that encourages stability through writing build recipes in a compiled language instead of in untyped templates, or having them implied through file organization.
* Writing on topics like <del>WebSockets</del> (<a href="/live-reload">done</a>), operable databases, and stability through data constraints.
* Nutrition and fitness: Leaner diet, run 1000 miles in 2019. Targeting <150 lbs. and ~10% body fat.

</div>
</details>

<details>
    <summary class="font-semibold font-serif italic list-outside pl-1 text-sm">Apr 20, 2018</summary>
    <div class="my-10">

<div class="mb-16 mt-8
            ml-[calc(-1rem)] mr-[calc(-1rem)] w-[calc(100%+2rem)]
            lg:mr-[calc(-1rem-75px)] lg:ml-[calc(-1rem-75px)] lg:w-[calc(100%+2rem+2*75px)]
            xl:mr-[calc(-1rem-150px)] xl:ml-[calc(-1rem-150px)] xl:w-[calc(100%+2rem+2*150px)]
            ">
    <img src="/assets/images/now/sutro-giants.jpg" alt="Sutro giants" class="lg:rounded-lg w-full" loading="lazy">
</div>

I'm in San Francisco, working on technology at Stripe.

A few points of focus:

* Meditating every day.
* Learning <a href="/rust-web">Rust</a>, and using it to proof out resilient services that don't need constant human attention.
* Exercising my <a href="https://twitter.com/brandur/statuses/823588112488013824">attention muscle</a> by putting in more periods of deep thought and intense non-maintenance work. I wake up early, and try to focus on only one thing at a time.

</div>
</details>


<!--

/ ---------------------------------------------------------------------------- 
/ OLD
/ ---------------------------------------------------------------------------- 

## April 20, 2018


## November 26, 2017

/ p I'm in Japan, visiting its unique duality of the most beautiful natural and urban environments in the world, decompressing from the ever-turning treadmill of electronic life, writing, and visiting as many <a href="https://en.wikipedia.org/wiki/Onsen">onsens</a> as I can find.



## September 18, 2017

/ p I'm in Canada enjoying its exceptional natural beauty during the final days of summer, visiting family, and attending the weddings of a few of my oldest friends.



## January 23rd, 2016

/ I'm in San Francisco concentrating on self-discipline, self-improvement, and shipping the next big thing at Stripe.



## Other old stuff

/ li Seeking to deeply understand some of my favorite pieces of technology like Postgres and Rust, and transforming those findings into published material.
/ li Writing an aspirational guide for building software that's simple, robust, and stable without constant human attention.
/ li Some basic voice recording to help me get more articulate and be able to form more cohesive long form thoughts in speech.

-->
