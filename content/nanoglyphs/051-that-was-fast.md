+++
image_alt = "Below the Brooklyn Bridge"
image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/051-that-was-fast/brooklyn-bridge@2x.jpg"
published_at = 2026-05-18T20:39:16+02:00
title = "Well, That Was Fast"
hook = "Stainless has been acquired by Anthropic, making this my second acquisition inside of a year, and this one of the shortest stints I've ever had at a company. Also, the Great Saunter."
+++

A few months back I wrote about [joining Stainless](/nanoglyphs/047-stainless).

Today, [Anthropic acquires Stainless](https://www.anthropic.com/news/anthropic-acquires-stainless).

It's pretty surreal. Last year I experienced the whirlwind process of having our small company (Crunchy) acquired by Snowflake. Visiting Berlin at the time the deal went through, I was told that waiting until I got back to the US would be too long, so I was to report immediately to Snowflake Berlin to be issued a temporary work machine. This was followed by two weeks of "Snowcademy" (a process in which one watches a long series of hopelessly outdated YouTube videos), and we went from that to implementing Snowflake Postgres in earnest, targeting a release date before EOY.

Given it was my first acquisition ever, I _knew_ that of course I wouldn't have to worry about something like that again for a long time. More probably, ever.

Almost exactly one year later, I'm in Berlin again, and the company I work for was just acquired, _again_. A speedrun of such absurd timing that I'm still having trouble believing that it really happened.

I'm less frantic about it this time because I won't be going to work at Anthropic. What am I going to do? Well, I'm still mulling that over:

* At first I'd sold myself on a form of early retirement in which I'd be an old man of 41 (approximately 150 in programmer years) traveling the world and screaming at young guns in their 20s about how we used to _write_ code instead of prompting it, hopefully with some skiing, diving, and shaking fists at clouds on the side.

* Next, it occurred to me that I should be focusing on growing [River](https://riverqueue.com/) into a sustainable business. It's maybe 25% of the way there, and with 6 to 12 months of concerted time and effort maybe I could get it to the point where it covers the costs of my degenerate Club-Mate habit.

* _Next_, I realized that the relative isolation of solo company building will probably drive me crazy, so maybe I should try to get another job (after taking some time for a minimum of one trip to the Asia-Pacific region of course). A couple threads appeared that are intriguing enough that I have _just_ enough self-awareness to recognize I'd be a fool not to tug on them a little.

In short, I don't know yet. For now, I'm plugging away on River:

* Last week we shipped [resumable jobs](https://riverqueue.com/blog/resumable-jobs). They let you subdivide jobs into skippable sections to avoid repeat work in case of failure. Useful when parts of a job are time-consuming or expensive, like invoking an LLM or executing a large analytical query.

``` go
func (w *DataPipelineWorker) Work(ctx context.Context, job *river.Job[DataPipelineArgs]) error {
    river.ResumableStep(ctx, "download", nil, func(ctx context.Context) error {
        return downloadData(ctx, job.Args.SourceURL)
    })

    // If River was forced to stop between download and transform or midway into
    // transform, the next run of this job skips download and picks up here.
    river.ResumableStep(ctx, "transform", nil, func(ctx context.Context) error {
        return transformData(ctx)
    })

    river.ResumableStep(ctx, "load", nil, func(ctx context.Context) error {
        return loadData(ctx)
    })

    return nil
}
```

* Within the next few days, we should have a V2 iteration on River workflows out the door. Our benchmarks show a 20x throughput improvement.

* Wrote a blog post on the [history of job queues in Ruby](https://riverqueue.com/blog/ruby-queue-history) over the years, starting with BackgrounDRb _way_ back in the day (circa 2008), to Solid Queue in 2026, and going over the improvements in each generation.

---

I feel much the same about this acquisition as the last one: kind of good, kind of bad. I can't complain much because especially given I was only at the company for ~6 months, it was a favorable outcome. That said:

* I genuinely felt the new product I'd been working on for Stainless had serious promise right up until the moment it was cancelled. We had a great team iterating on the thing at cannonball speed, and it's sad that it won't see the light of day.

* I don't feel great about [promoting a company](/nanoglyphs/047-stainless) only to have it be subsumed a few months later. I'm sure I'm flattering myself in that my opinion had much of an effect on anyone else's decision making process, but in case it did: I'm sorry. Sincerely, I had no idea this was coming, and even if I could have guessed, couldn't have imagined in any universe that it'd be this soon.

<img src="/photographs/nanoglyphs/051-that-was-fast/pickleball@2x.jpg" alt="Pickleball" class="wide" loading="lazy">

## The Great Saunter (#great-saunter)

During a last chance visit to New York, one of my colleagues tipped me off to the existence of [The Great Saunter](https://shorewalkers.org/great-saunter/), an epic walk where participants walk the entire perimeter of Manhattan.

I missed the official event after learning about it the evening after it'd already happened, but walking the shoreline of Manhattan is something I've wanted to do since first stepping foot on the island. Discovering that there were occasionally _sane_ people who did it greatly increased my confidence that the feat was possible, and not only that, but these other walkers had even left me a map, and one updated within the last 24 hours. Now I _had_ to do it.

Coming out at 58 km (38 miles) and taking over 10 hours with no stops, it's the single longest walk I've ever done. I wasn't particularly worried going in: it's mostly flat, my shoes were perfectly broken in, and there's no shortage of places to stop for food or water. But by the end of the day, in addition to a pervasive soreness permeating my body from legs to back, I'd actually managed to develop blisters on the _flats_ of my feet, just by virtue of them having been placed down on the pavement so many times in a row. I'm used to blisters on long hikes, but this is something I didn't even know this was possible.

As darkness fell and my gait transformed into a weird sideways shuffle that favored my toes and the balls of my feet (the only parts not yet blistered), it got harder and harder to ignore each new glowing subterranean portal down to the metro that I passed. One of these merciful machines could deliver me to my hotel's front door, where I could collapse and not get back up again for a good 24 hours. In 15 minutes it could all be over.

But I couldn't shake the feeling that it was now or never. If I stopped early on this walk, the chances I ever tackle it again, especially having a better idea of its full extent the next time around, were roughly zero. So down to the last 10 km, I dragged one foot after the other for as long as it took. Around the Battery, passed the East River Esplanade, through the mess of ongoing construction along FDR Dr, under the Brooklyn Bridge, passed the Manhattan Bridge, back west from the Williamsburg Bridge, north through East Village, and finally up to my hotel on Union Square. I was putting places to names for the neighborhoods and infrastructure I'd read about in [The Power Broker](https://en.wikipedia.org/wiki/The_Power_Broker), and knowing nothing about New York, had previously only had the vaguest possible notion of.

Limped into and through my hotel, dragging myself the last few feet from elevator to room. Mission accomplished. Sitting down never felt so good.

Until next week.

Brandur

<img src="/photographs/nanoglyphs/051-that-was-fast/great-saunter-map@2x.png" alt="Great Saunter map" class="img_constrained" loading="lazy">

<img src="/photographs/nanoglyphs/051-that-was-fast/harlem-pool@2x.jpg" alt="Harlem pool" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/051-that-was-fast/w-29th-st@2x.jpg" alt="Harlem pool" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/051-that-was-fast/runners@2x.jpg" alt="Runners" class="wide" loading="lazy">

<!-- <img src="/photographs/nanoglyphs/051-that-was-fast/park@2x.jpg" alt="Runners" class="wide" loading="lazy"> -->