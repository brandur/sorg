+++
image_alt = "Globe and crowd at Rails World 2025."
# image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/043-rails-world-2025/globe@2x.jpg"
published_at = 2025-09-20T13:10:55-07:00
title = "Rails World 2025, the Event Horizon, the Ruby Embassy"
hook = "A visit to Amsterdam for Rails World 2025. CI loops measured in days and why to make sure that CI can run locally. A visit to the Ruby Embassy at Beurs van Berlage."
+++

Readers --

I recently attended the 2025 edition of Rails World in Amsterdam, my third year of Rails World in a row, or a perfect grade of three out of three. I don't attend many other conferences, but Rails World is _the_ best conference, so I'll keep going back as long as I can keep getting tickets (it was very close this year).

Like the inaugural Rails World in 2023, it was held in the gorgeous Beurs van Berlage, a stately edifice in central Amsterdam built in 1903 as the city's stock exchange. Attending an event there feels like you're part of an exclusive diplomatic summit in the old world.

The first talk of the first day was of course DHH's annual state of the union keynote, which for my money is the single best talk in the entire tech industry every year. I used to watch Apple events, but as they devolved, I started reading their five sentence summary as instead. At Rails World on the other hand, I'm dialed in for every minute. He's been a 10 out of 10 public speaker for some time, but every year he seems to get a little better at it. The pace, jokes, and timing all incrementally tighter.

You can [watch it here](https://www.youtube.com/watch?v=gcwzWzC7gUA&list=PLHFP2OPUpCebhAv1ZWb_978cTl1o-yue-&index=1). My rough highlights:

* New product features include Markdown support in Rails, a new rich text editor named Lexxy, resumable Active Jobs, Action Push, SQLite replication with Beamer, geo routing in Kamal, and free [Campire with ONCE](https://once.com/campfire) (previously, $299).

* DHH talked about [how they run test suites locally](https://world.hey.com/dhh/we-re-moving-continuous-integration-back-to-developer-machines-3ac6c611). With modern chipsets and _without_ the inherent sloppiness of "just throw it over the fence into CI", they run their entire suite in 70 seconds on the fastest AMD chips.

* [Omarchy](https://omarchy.org/), an opinionated Linux setup on Arch and Hyprland, got a lot of air time. They pulled off a great demo on stage that combined a brand new Framework laptop with a USB stick containing Omarchy where they got a fresh system installed, booted, and up and running to the point where they created a Rails app, al in less than five minutes of real world clock time. In the following days, they'd get this install time down to [less than two minutes](https://x.com/dhh/status/1966919727173038546) with the community's help.

* A fact that surprised me even if maybe it shouldn't have is that their test suite runs **2x faster on Linux** given similar hardware, mainly due (if I understood correctly) optimizations that have gone into Linux over the years, and Apple's tireless work at deoptimizing the Mac.

For the third year in a row Heroku got a special shout out. DHH revealed in a surprise twist that the conference's Campfire instance was running entirely off a $300 mini-PC from his closet in Copenhagen. He talked about how computing resources of the same scale would cost $600 _per month_ on Digital Ocean, and an awe dropping $1,200 a month on Heroku.

<img src="/photographs/nanoglyphs/043-rails-world-2025/ruby-crowd@2x.jpg" alt="Post-keynote Ruby crowd" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/043-rails-world-2025/scaffold@2x.jpg" alt="The wine list" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/043-rails-world-2025/sponsor-floor@2x.jpg" alt="Sponsor floor" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/043-rails-world-2025/pair-programming@2x.jpg" alt="Pair programming" class="wide" loading="lazy">

---

## When build time is measured in days (#days)

One of the [best ever HN comments](https://news.ycombinator.com/item?id=18442941) talks about product development at Oracle. CI (continuous integration) is 20 to 30 hours for a full test run, with no interim feedback, making the edit-compile-debug loop about 20 to 30 hours. Developing a feature or bug fix takes dozens of iterations, leading to lag times of six months to a year. This was written in 2018, so it's probably all even worse now.

As another data point, a company I interacted with recently runs a CI loop that's days long. This was too long so they decided to move a subset of testing to a "precommit" check which takes "only" two to fours to run. No one expects any change to take less than a week. Multiple weeks is normal.

Stripe wasn't quite as bad. When I left, CI was in the neighborhood of 15 to 20 minutes, with a single test run 30 seconds to a minute (most of that going to start up effort). That's an order of magnitude better than Oracle, but one minute test runs and 15 minute CI loops are still absolute killers when you're trying to make work happen. And that's discounting spend completely. The 15 minute CI time was only possible through the use of 100s/1000s of test nodes running in parallel. I don't know how much a single run of the test suite cost, but I would guess it's measured in tens of dollars rather than cents.

The more data I collect from the industry, the more it's cemented for me that these anecdotes aren't aberrations, but rather the "gold" standard, even amongst even the most sophisticated big tech companies.

Let me restate that for emphasis: even with the advent of CPUs with the compute output equivalent to entire countries from a few decades ago, high bandwidth chipsets, and NVMe drives, combined with the labor of the smartest, best compensated engineers money can buy, iteration times are measured in hours at best, and just as often days. A typical engineer making a half million in total compensation will come to work, push some code in the morning, go to eat free lunch down at the cafeteria, fix some tests flagged from their morning commit, push once more in the afternoon, then go home. The company's "staff" engineers will then attend conferences to brag about the lean, refined efficiency of their development process, taking great care to omit all the bad parts [1].

### The overlooked power of commodity hardware (#commodity-hardware)

This leads me back to what I think is David's best idea from the last couple years (even amongst a constellation of good ones): local CI.

At Basecamp they run the [test suite locally](https://world.hey.com/dhh/we-re-moving-continuous-integration-back-to-developer-machines-3ac6c611), harnessing the incredible potential of a modern dev machine instead of adhering to the traditional dogma that everything must be outsourced the cloud. It used to be the case that it took the power of a build farm to run a comprehensive test suite, but as long as the bounds of a project are kept with reason, that's not only no longer the case, but local runs can be considerably faster. Theirs takes 70 seconds on a fast AMD chip.

<img src="/photographs/nanoglyphs/043-rails-world-2025/hey-test-suite@2x.jpg" alt="Hey test suite" loading="lazy">

The other reason that CI is important is reproducibility, or to prevent the "works on my machine" phenomena. This is still important, but more easily achievable locally compared to the old days through the use of containers or other isolation/clean build techniques.

Basecamp uses a [signoff script](https://github.com/basecamp/gh-signoff) that a human runs explicitly after confirming that all checks are green on a local branch. It mostly boils down to a single GitHub API call:

``` sh
if gh api \
    --method POST \
    "repos/:owner/:repo/statuses/${sha}" \
    -f state=success \
    -f context="${context_name}" \
    -f "description=${user} signed off" >/dev/null; then

    # Build success message
    if [[ -z "$context" ]]; then
    success_messages+=("${STATUS_SUCCESS} Signed off on ${sha}")
    else
    success_messages+=("${STATUS_SUCCESS} Signed off on ${sha} for ${context}")
    fi
```

This is an enforcement technique, but not a _strong_ enforcement technique, and of course it's possible for someone to fib during signoff, but that's the case for almost anything. Even with a normal GitHub mandatory check it's common to leave a "bypass" checkbox available that can be abused. It turns out though that this is ... okay? If someone starts to show a history of breakages from erroneous sign offs, it can be dealt with specifically, just like any other workplace problem.

And the benefits are crystal clear! 70 seconds test times for a large real world app. That's better than Google, better than Apple, better than Dropbox, better than Netflix, and better than Stripe, likely by 10-100x. A test suite that fast keeps developers happy and productivity high. And all in Ruby! One of the world's slowest programming languages.

### The event horizon (#event-horizon)

Broadly, there are three stages in the long ark of a company's CI trajectory:

1. Early on, the test suite is run only locally.
2. CI is set up. Test suite is runnable both locally or in CI.
3. Test suite gets too big, or too custom, or has too many dependencies. Test suite is runnable _only_ in CI.

I've never actually been at a company where they transition from 2 → 3. I was pretty early at Stripe, but it still happened before I got there. I can't say _exactly_ how senior engineers rationalize the change, but I can imagine how the conversation goes. Something along the lines of, "We're just too smart/too sophisticated/have too many special requirements to be running things on laptops now. We are web scale. We _have_ to use the cloud." [2]

After the transition to stage 3 there's a brief moment where things are still theoretically recoverable, like if a small team of dedicated engineers worked day and night for a few weeks they could walk things back from the brink, maybe. But generally speaking, you're caught in the gravity well. After crossing the event horizon, there's no going back. The overwhelming default will be to descend further into the black hole. The build continues to get more custom and requires more configuration and leverages more cloud constructs. There's ~0 performance feedback now, so engineers don't even notice half the time when they write slow tests, further degrading the morass.

In this developer's opinion, it's not only important to keep an eye out for that 2 → 3 transition and avoid it, but absolutely _crucial_ to do so. A test suite that can still run on one machine can be shaped and sped up. Once you're cloud only, all bets are off.

---

<img src="/photographs/nanoglyphs/043-rails-world-2025/second-floor@2x.jpg" alt="Second floor" class="wide" loading="lazy">

## A visit to the Ruby embassy (#ruby-embassy)

You climb to the top of the Beurs van Berlage clutching your Ruby passport. You need to get it validated because it's your ticket. Formal travel authorization from The Community for future Ruby events. Rails World is today in Amsterdam, but Friendly.rb and EuRuKo are coming up mere days/weeks from now in Bucharest and Viana do Castelo, Portugal.

You enter a distinguished, high-ceiling room. Its understated decor is tasteful, with wood paneling surrounding the room, a series of impressive stained glass window overhead, and furniture that's ornate, but not luxurious. It's quiet. The mood is professional, and a little somber.

There's a security checkpoint at the door. A guard is searching a person in front of you and as you accidentally cross a line you didn't on the floor he holds up his baton and says, "Sir, step back please. I'll be with you in a minute." He finishes with the person in front of you and calls you up. He waves the wand around your body and asks solemnly, "Sir, have you been in the vicinity recently of any packages that were emitting ticking noises?" You answer in the negative and are waved through.

Next up, a waiting area to fill out documentation. You take a form, furrow your brow, and get to work, expecting it to take only a few moments. It's surprisingly comprehensive, taking a full 10 minutes to get through, and with many questions requiring careful thought to answer definitively. During that time more Rubyists are ushered in and start work on their own forms. By the time you finish it's standing room only.

<img src="/photographs/nanoglyphs/043-rails-world-2025/embassy-1@2x.jpg" alt="Ruby embassy 1" class="wide" loading="lazy">

You advance, and after finishing with a person in front you, a lady at a desk in the center of the room calls out, "NEXT! Come on now, I don't have all day."

You walk up and surrender your paperwork. "Okay, let's see what we have here. Number of stars in the universe? You put 2,342,952,459,236,419 [3]. Okay, very precise. Counted them all did you. Best smell in the world? Sharpie. Okay, I might've said baking bread, but I can accept that answer. You said on here the last time you cuddled with a kitten was ten years ago? That seems like an awfully long time. What about this, when was the last time you cuddled a _cat_? i.e. "Kittie" rather than "kitten". A few weeks ago. Okay, we can work with that."


<img src="/photographs/nanoglyphs/043-rails-world-2025/embassy-2@2x.jpg" alt="Ruby embassy 1" class="wide_portrait" loading="lazy">

The interview takes a few minutes and after it's done your documents are returned to you. You photo is taken and you're sent over to one last table to have your passport is validated and finalized.

A++ bit. I'd visit the embassy again any day.

<img src="/photographs/nanoglyphs/043-rails-world-2025/ruby-passport@2x.jpg" alt="Ruby passport" class="wide" loading="lazy">

---

## Back in blue (#blue)

We were sponsoring again this year, but had set it up before [the acquisition](/nanoglyphs/042-resumed#acquired) in June, so that our branding was still the old Crunchy Data like it's been every other year.

It was bittersweet. Great to be there again under our old banner, but Rails World will have the special distinction of being the last ever appearance of the Crunchy Data booth in the wild. I'm not sure what'll happen next year, but if we do come back, it'll be accompanied by snowflake iconography in baby blue, surrounded by a familiar halo of signature arrow glyphs.

We went out for dinner at an Italian place along one Amsterdam's labyrinthine network of canals. We ask for the wine list. The waiter returns, and drops a crate of empty wine bottles onto the table.

Love it.

Until next week.

<img src="/photographs/nanoglyphs/043-rails-world-2025/wine-list@2x.jpg" alt="The wine list" class="wide" loading="lazy">

[1] The less generous amongst us would call this sort of aggressive fact filtering "lying", but an inordinate number of thought leaders in the engineering space pull it off with a bald face, presumably having had their conscience cleansed by checking the balance of their brokerage account.

[2] There's probably a few instances where this is actually true, like if you're building Google Chrome, but like "big data", for every one org that has it, there are a thousand that think they have it.

[3] I looked this one up later and despite jamming as many numbers into the box as I could fit, I was still off by about ten orders of magnitude. Actual estimated number is more like 200 sextillion stars (200,000,000,000,000,000,000,000).
