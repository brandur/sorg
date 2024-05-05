<div class="my-16
            ml-[calc(-1rem)] mr-[calc(-1rem)] w-[calc(100%+2rem)]
            lg:mr-[calc(-75px)] lg:ml-[calc(-2rem-75px)] lg:w-[calc(100%+2rem+2*75px)]
            xl:mr-[calc(-150px)] xl:ml-[calc(-2rem-150px)] xl:w-[calc(100%+2rem+2*150px)]
            ">
    <img src="/photographs/now/2024-05-stayery.jpg" alt="Stayery, Berlin" class="lg:rounded-lg w-full">
</div>

<!--

 class="max-w-none w-screen lg:-mx-[200px] lg:w-[calc(100%+400px)]">

-->

## Berlin, AD, and queues (#berlin)

I'm spending the month in Berlin, where I'm trying to run and write every day, and enjoy time in a place that's less reminiscent of a zombie wasteland than my home city.

The next big project I'm tackling at work is Active Directory. That sounds about as fun as a root canal, but I take it as an interesting challenge. AD is a long in the tooth technology that's still in use by many of the biggest players in the industry (we even used it at Stripe!). How can we integrate it in such a way that it gets big users what they need, produces as little code as possible and as few headaches for us, and maximizes the yield in leverage we get out of the effort. For example, it might involve ignoring the low level AD APIs and integrating [SCIM instead](https://en.wikipedia.org/wiki/System_for_Cross-domain_Identity_Management), thereby buying us compatibility with other non-AD SCIM-based systems.

Blake and I continue work on our open source Postgres job queue, [River](https://github.com/riverqueue/river), which I think is fair to say is the most full-featured in the Go ecosystem by some margin. There's something incredibly satisfying about taking a project scoped to a known, fixed domain, and refinining its code over and over again until it's _perfect_.  Trying to handle every edge, and with attention to detail on every line of code.

I recently published [Ruby gem](https://github.com/riverqueue/riverqueue-ruby) that enables job insertion in Ruby, but for jobs to be worked in Go, which is something that I'd always wanted back at Heroku and Stripe. I wrote about [my experience putting in type checking with Steep, and publishing RBS files](/fragments/ruby-typing-2024) for the project.

<div class="not-prose">
    <p class="font-serif italic my-1 leading-normal not-prose text-sm tracking-tight">This page was last updated on
        <span class="font-bold">May 5, 2023</span>.
    </p>
</div>

<!--

/ ---------------------------------------------------------------------------- 
/ OLD
/ ---------------------------------------------------------------------------- 

## Apr 9, 2023

I'm in San Francisco, where inertia keeps me rooted.

Work on Crunchy Bridge continues. As part of filling out a self-evaluation last week I scanned every pull request I've issued over the last year, and I liked what I saw. What we've shipped during that time is above and beyond any org I've worked at before. Small teams, agile tech stacks, and lack of a culture of objection-for-objection's-sake do wonders for productivity.

A few things I sent out the door recently:

* Multi-factor authentication, supporting WebAuthn (Yubikeys and biometric challenges like Touch ID) and TOTP (time-based one-time passwords). ([Notes](/fragments/lean-fast) on getting this shipped quickly.)
* An asynchronous query runner that's built to scale by taking advantage of Go's parallelism, by storing results to S3 instead of Postgres, and which prunes its results regularly.
* [Retired our use of Keycloak](/atoms/gkoxmy2), which involved many smaller tasks like adopting a [fancy new password hashing scheme](/fragments/password-hashing).

Pre-lockdown, I'd gotten into the best shape of my life by baking exercise into my schedule with a daily run commute, fitness which I unfortunately let languish. Newly armed with a WeWork pass and gym membership (for the showers), I'm bringing it back. A straight shot from the mountain down Market St to Embarcadero -- 50km/week if I keep it up.

Next month, Europe.

## Dec 28, 2022

I'm back home in Calgary for the holidays, staring into the precipice of 2023 which between money markets, strife, and war is shaping up to be a formidable year.

At work, we're aiming to build the best database-as-a-service in the world. I shipped more features over the last year than the previous five combined, and which were built into a robust stack that to this day has less tech debt than many two-week-old startups (I'm kind of proud of it, you might be able to tell). We have another aggressive roadmap for 2023, and I'll be doing my best to make sure that we don't slip.

I added a couple new sections to the site recently:

* [**Atoms:**](/atoms) Short multimedia particles minus the stress of a social media platform ([atom feed](/atoms.atom)).
* [**Sequences:**](/sequences) Periodic large format photos paired with prose. An older project, but one which I recently revived, flattened, and republished.

A few weeks before that I became somewhat enamored by the idea of [Spring '83](https://www.robinsloan.com/lab/specifying-spring-83/) and ended up [writing a server implementation](https://github.com/brandur/neospring) which is now in prod and [hosts my board](https://neospring.brandur.org/). I don't think Twitter is being displaced anytime soon, but these indy web projects are great.

In 2023: write, move, visit France.

## Apr 5, 2020

It's 2020. Like for almost everyone else on Earth, COVID-19 is top-of-mind. I'm working from home, San Francisco is sheltering in place, and the future is a hugely uncertain time.

{{HTMLRender (ImgSrcAndAltAndClass "/assets/images/now/twin-peaks-stairs.jpg" "Stairs up to Twin Peaks" "overflowing")}}

As bad as our present day situation is, an indefinite work from home policy has given me more flexibility and more energy in my day-to-day than I've ever had in my adult life, and I'm going to do my best not to waste it.

Some things I’m working on:

* Write every day, and try to so fluidly. Instead of agonizing over every word, get content down, revise, and revise again. Some projects:

    * A (roughly) weekly newsletter called <em>Nanoglyph</em>. I’m challenging myself to send at least 30 editions in 2020, and do so without compromising content quality. You should <a href="https://nanoglyph-signup.brandur.org/">try subscribing</a>.
    * A development log with notes on daily software discoveries and projects. Most entries will be of minor interest, but frequent enough to build writing cadence. <a href="/fragments/google-cloud-run-deploy">For example</a>.

* Meditate every day.

## Dec 31, 2019

/ p I’m in Calgary for the winter break. It’s the last day of 2019 and we’re on the precipice of a new decade.
/ p
/   img.overflowing src="/assets/images/now/calgary-snow.jpg" srcset="/assets/images/now/calgary-snow@2x.jpg 2x, /assets/now/calgary-snow.jpg 1x"
/ p Some things I’m working on:
/ ul
/   li Writing a (roughly) weekly newsletter called <em>Nanoglyph</em>. I’m challenging myself to send at least 30 editions in 2020, and do so without compromising content quality. You should <a href="https://nanoglyph-signup.brandur.org/">try subscribing</a>.
/   li Write more, and more fluidly. Instead of agonizing over every word, get content down, revise, and revise again.
/   li The craft of software has landed on a plateau where big, complicated, fragile backend deployments that support slow, underwhelming, buggy frontend products is the norm. Find techniques and ideas to reverse this trend and boost their signal.
/   li Ran 1000 miles and did 42k pushups in 2019, do so again in 2020. Keep weight at ~150 lbs. Zero in on ~10% body fat.
/   li Reboot meditation practice. Did pretty well in 2018, but fell off the wagon completely in 2019. Aim for a couple 30 consecutive day runs.



## June 5, 2019

/ p I'm in Berlin.
/ p
/   img.overflowing src="/assets/images/now/molecule-man.jpg" srcset="/assets/images/now/molecule-man@2x.jpg 2x, /assets/now/molecule-man.jpg 1x"
/ p A few points of focus:
/ ul
/   li <A href="/sequences-project">A photography project called <strong><em>Sequences</em></strong></a> as an experiment to promote the independent web.
/   li A tiny static site framework that encourages stability through writing build recipes in a compiled language instead of in untyped templates, or having them implied through file organization.
/   li Writing on topics like <del>WebSockets</del> (<a href="/live-reload">done</a>), operable databases, and stability through data constraints.
/   li Nutrition and fitness: Leaner diet, run 1000 miles in 2019. Targeting <150 lbs. and ~10% body fat.



## April 20, 2018

/ p I'm in San Francisco, working on technology at Stripe.
/ p
/   img.overflowing src="/assets/images/now/sutro-giants.jpg" srcset="/assets/images/now/sutro-giants@2x.jpg 2x, /assets/now/sutro-giants.jpg 1x"
/ p A few points of focus:
/ ul
/   li Meditating every day.
/   li Learning <a href="/rust-web">Rust</a>, and using it to proof out resilient services that don't need constant human attention.
/   li Exercising my <a href="https://twitter.com/brandur/statuses/823588112488013824">attention muscle</a> by putting in more periods of deep thought and intense non-maintenance work. I wake up early, and try to focus on only one thing at a time.



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
