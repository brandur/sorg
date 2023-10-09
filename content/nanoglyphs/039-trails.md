+++
image_alt = "Kings Canyon National Park"
image_url = "/photographs/nanoglyphs/039-trails/kings-canyon@2x.jpg"
published_at = 2023-10-08T08:53:23+02:00
title = "Trails, Charleston, `t.Parallel()`"
+++

Readers, it's been a while. Again.

I'll give you my usual spiel after a long absence. This is _Nanoglyph_, a semi-delinquent newsletter about software whose original premise was to be weekly, but which in reality is more like a quarterly affair. If you don't remember signing up for this, you _probably_ did, but it was a long time ago. If you want to nope right the eff out, do me a favor by using this [one-click unsubscribe](%unsubscribe_url%) instead of smashing that "mark as spam" button.

I want to keep this one short because I have more to write about later this week, so without further ado.

---

## The legacy of John Muir (#john-muir)

I spent the last month and a half in and out of the Sierra Nevada on the [John Muir Trail](https://en.wikipedia.org/wiki/John_Muir_Trail), which runs 350 km (although we didn't go that far) from Yosemite Valley south to Mount Whitney.

The sights were spectacular. Pristine wilerness, alpine lakes, towering walls of endless granite, and waterfalls in every direction. California's been in and out of drought conditions for a decade now, but this year was one of outsized precipitation, and when you're up in the mountains, you can see the difference with your own eyes. Lakes that in other years would look a little melancholy, overflowed. Entire creeks that don't normally exist sprung into being. Everything was green and lush.

It's been a long time since I went backpacking last, so my weeks leading up to the trip were focused inordinately on gear and logistics. What's the latest on gas stoves and water filters? How does one go about cramming a week's worth of dehydrated food into a [bear can](https://www.rei.com/learn/expert-advice/bear-resistant-canisters.html)? [1] What's the maximum weight I can carry for weeks at a time before my knees buckle? I was impressed by the advances in outdoor gear since I handled it last. The broad strokes are the same, but with everything a little lighter, a little easier to use, a little more water resistant, and a little better at what it did. I had a tent that weighed two pounds and with a little practice, which I could assemble in two minutes flat (the [Nemo Osmo 2P](https://www.nemoequipment.com/collections/tents-and-shelters/products/hornet-osmo)). 

The length of the John Muir Trail is shared by the much longer [Pacific Crest Trail](https://en.wikipedia.org/wiki/Pacific_Crest_Trail) (PCT) that runs along the western spine of the US from Canada to Mexico, so at any point if you ever make the mistake of thinking that the backpacking trip you're on is hardcore, you're being reminded constantly that there are people 10x more hardcore as PCT hikers zip by and up from behind. I couldn't help but be amused by them because despite being 100s of disparate people with no relation to one another, they all look the same. They're all wearing shorts when it's clearly long pants weather, all have the same brand of funny half-calf sock that I don't really know how to describe, often with gaiters, and carry packs about half the size of other backpackers. It's not clear how they have enough gear on them to last a night, let alone weeks, and those few we stopped to ask about it also had trouble explaining it. Wiry, muscular, and with a look in their eyes that was half focused, indomitable determination, and half let-it-all-go madness.

We had some long days, but every moment was worth it. There is nothing else like being out in the middle of some of the most remote parts in the country, sleeping under the stars. We'd settle into our pitches at night with the only noise being running water (if camped near a creek) or deafening silence (if not). Unlike your typical family campground, no dogs barking, diesel generators running, or low roar of nearby highway traffic. Only you and the endless wilderness.

Our trip ended earlier than planned as we took a wrong turn hitchhiking down to Bishop on festival weekend, became temporarily homeless as there wasn't a single room to rent inside of a hundred miles, and ended up being dropped off in a field by a man who called himself only "Nightrider", but two of us returned a week later to complete the rest of the journey. The [full story here](/john-muir-trail).

<img src="/photographs/nanoglyphs/039-trails/seldon-pass@2x.jpg" alt="Seldon Pass" class="wide_portrait" loading="lazy">

---

## Charleston (#charleston)

A few days later I flew out to Charleston, where Crunchy is headquartered, for our team's annual on-site where we gather in person to get to know each other, brainstorm ideas, and get some planning done. I didn't know a thing about Charleston before visiting last year, but it's a great little city, with a charming downtown/French quarter and lots of nice boutique cafes and restaurants to visit. The weather is typically southeastern though. The humidity is so thick that I when I made the dubious decision to go for a run, I thought I might be drowning, on land. I arrived back at the hotel 30 minutes later looking like I'd just been rescued out of the harbor.

On a visit to Charleston's aquarium (easily reachable by a short water taxi ride) I was impressed by Liberty, a bald eagle. It's a high crime up there with treason to keep these American symbols in captivity in the US, but there are some exceptions, like when a bird's been permanently injured. Liberty had been when she was young and could no longer fly. I've been seeing bald eagles from afar for a long time, but it's rare to get so close to one. Liberty wasn't trapped behind bars or glass and seemed to be quite happy with humans in reasonably close proximity to her. It's when you get this close to a bald eagle that you realize how true their storied reputation is -- enormous in stature, beak and talons like razors, and oozing menace and power.

<img src="/photographs/nanoglyphs/039-trails/liberty@2x.jpg" alt="Liberty, the bald eagle" class="wide_portrait" loading="lazy">

It's great to see everyone in person. Charleston's airport is small and it's a bit a journey to get there (not even people flying from up the coast in New York can get a direct), but most of us come away thinking that the few days spent together in-person were invaluable. Not a lot of hard work got done per se (that wasn't the point), but we did more to build concensus on hard problems in a few days than would otherwise happen in a few months.

I'll spare you another rant about my unpopular opinions on remote work (wrote one, deleted it), but I predict that in the near future any companies who make the mistake of equating in-person collaboration/ideation to what happens over Zoom/Slack are going to be leaving a lot on the table.

---

## `t.Parallel()` everywhere (#t-parallel)

An improvement I’ve put into our codebase over the last few months is to add [`t.Parallel()`](https://pkg.go.dev/testing#T.Parallel) annotations to every test case.

Go’s testing tool already parallelizes nicely by default by testing N packages at once, where N defaults to your local machine’s CPU count, and universal use of `t.Parallel()` isn’t by any means best practice by conventional wisdom, but it does help in cases were a single package grows large and it's common to run its tests frequently in isolation.

Our API package, before `t.Parallel()`:

``` sh
$ go test ./server/api -count=1
ok      github.com/crunchydata/priv-all-platform/server/api     1.486s
$ go test ./server/api -count=10
ok      github.com/crunchydata/priv-all-platform/server/api     11.786s
```

With `t.Parallel()` interspersed liberally, a 30-40% speed up:

``` sh
$ go test ./server/api -count=1
ok      github.com/crunchydata/priv-all-platform/server/api     0.966s
$ go test ./server/api -count=10
ok      github.com/crunchydata/priv-all-platform/server/api     3.959s
```

A few things to watch out for:

* You’ll want to be [using test transactions](/fragments/go-test-tx-using-t-cleanup) to give each parallel run its own fully isolated environment.

* To avoid log output being interleaved randomly, make sure your logger has a [test bridge that redirects to `t.Logf()`](https://github.com/neilotoole/slogt) instead of just sending to stdout. Using `t.Logf()` will correctly collate output even across parallel runs, even it was interleaved when emitted.

* Postgres upsert deadlocks may become a problem where common fixture data is being upserted, _despite_ each run happening in its own transaction. We had to change our model to upsert common fixtures as a step that happens before any test is run.

* Runs of [goleak](https://github.com/uber-go/goleak) to look for leaked goroutines can no longer run on a per test case basis (because checks will find goroutines from other test cases). We moved to single package-level checks in `TestMain` instead.

* Once you've added `t.Parallel()` in you'll make want to make sure new test cases have it applied because it's really easy to forget. That's where the [`paralleltest` lint](https://github.com/kunwardeep/paralleltest) comes in, which is bundled into the commonly used golangci-lint.

That’s the short version. I published [a longer article about this as well](/t-parallel).

This is another one of those great things about running a smaller codebase. Instead of letting things slip to the point of intolerability through a slow-moving [tragedy of the commons](https://en.wikipedia.org/wiki/Tragedy_of_the_commons) effect that eventually gets so bad that no one can really do anything about it anymore, you can apply vast refactors to make sure things stay good. Our API package above had crept passed 1,000+ test cases and we knew that without another plan to keep expanding it we'd have to find creative ways of making it sustainable, like breaking it up into many smaller packages that'd parallelize. Bringing in `t.Parallel()` has definitively solved that problem.

---

## A compact UI any day (#compact-uis)

A minor controvery in the tech world was Slack rebuilding its app's UI, rolling out slowly over the last few weeks team by team. Our UIs have been degrading in familiar ways since we lost the plot somewhere around 2013, and Slack's changes have a familiar shape:

* More whitespace.

* More color gradients.

* More UI elements hidden away (namely, your normal list of Slack workspaces disappear under a single stacked button).

* Useful features hidden, useless ones given new, prominent positions. Less functional in every way.

<img src="/photographs/nanoglyphs/039-trails/new-slack-ui-normal-zoom@2x.png" alt="Screenshot of new Slack UI" class="wide" loading="lazy">

Knowing that I have a bias for heterodox opinion, I searched high and low for anyone out there _defending_ the new design, but I don't think there's a person on Earth. Even if you're microdosing regularly and don't mind all the psychadelic colors, there's no additive features in the new UI to actually _like_ -- it's strictly a downratchet.

Companies make major design missteps every so often, so that wouldn't be new, but I'd argue it's only the latest instance of a worrying, long-running trend industry-wide. As the field of app/UI design has gone mainstream, a healthy majority of the expanded base of general practitioners don't actually hold dear the values of great UX anymore, if indeed they ever did. Rather than the end goal being to deliver an interface that empowers its user and maximizes productivity, it's to build one that'll make a nice mockup for management and pretty screenshot on Product Hunt.

Designers, most of whom have unexceptional creative skill of their own, go look at what Apple and Stripe are doing and parrot it, tweaking it just enough to avoid looking like an obvious copy. Hand-crafting custom, intricate UI details in Photoshop is a form of art that takes real skill that most don't have, so they fall back on approximating beautiful design through the use of more whitespace, more gradients, and leaning on beautiful pixel-perfect system font rendering (which is almost totally ubiquitous across systems now) as a crutch. [2]

The broad look of apps and the web is discouragingly uniform these days, but while everyone else is zigging, the best designers will zag. Reams of whitespace in every direction satisfies some innate human aesthetic, but know what makes a tool more powerful? Density. Dense UIs put more actionable information on screen at once for their user to take advantage of, and are more efficient to use because there's less context switching in the form of scrolling/tabs/windows. There is a downside in that it's harder for the human eye to latch onto things, but with continued practice with a tool, that sharp edge gets rounded out.

I'll leave you with an example of an app that does this well: VSCode. Notice just how much information and number of controls/widgets there is per square millimiter of real estate. It's one of the (or maybe _the_) most productive software packages currently in existence.

<img src="/photographs/nanoglyphs/039-trails/vscode@2x.png" alt="Screenshot of VSCode" class="wide" loading="lazy">

Until next week.

[1] Bear cans are mandated, so you don't have a choice but to figure it out.

[2] I'll fully concede to often doing this myself, but I also don't call myself a designer.