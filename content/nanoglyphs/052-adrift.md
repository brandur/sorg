+++
image_alt = "Forest in the Vitosha mountains"
# image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/052-adrift/vitosha-forest@2x.jpg"
published_at = 2026-06-01T12:08:01+02:00
title = "Adrift, Minimum Viable Unit of Saleable Software, Balkans, Bears?!"
hook = "On life after acquisition, the minimum viable unit of saleable software and whether River can qualify, Balkan Ruby in Sofia, and hiking in the Vitosha mountains."
+++

There's something calming about being in Germany. People ride around on mechanical (not E-assisted) bikes. People pay for things in cash. An overwhelmingly common pastime is drinking beer beside the river like it's 1984. A not-insignificant part of the population rebuilds cobblestone streets and sells white asparagus and strawberries on street corners. Aside from the smartphones that are as ubiquitous here as they are anywhere, you could almost believe it's the same quaint, storybook Europe as before the turn of the century.

If you didn't already know about the high stakes battle around LLMs and the very future of human work, art, and civilization consuming the other side of the planet, you could spend months here never even knowing that it was happening. In 10-20 years, this country is either going to look exactly the same as it does today, or the consequences of choosing to recuse itself from the heated economic competition between superscalers in the US and China are going to be as clear as day. I don't know which.

---

It's been a weird week. After 20 years of an uninterrupted daily grind (Mentor, iStock, Heroku, Stripe, Crunchy, Snowflake, Stainless) since nearly the day I graduated university, after [a second acquisition in a year](/nanoglyphs/051-that-was-fast), I'm officially unemployed. It's weird, but not necessarily bad, and in fact, kind of a good thing. It's going to be a great travel year, and of course, I'm going to try and build some things.

In this issue: thinking through whether it's still viable to sell software in the age of LLMs, a trip to Sofia, Bulgaria, and a hike up into the Vitosha mountains south of the city.

---

## Minimum Viable Unit of Saleable Software (#minimum-viable-unit)

Yesterday I published [The Minimum Viable Unit of Saleable Software](/mvuss).

Here's the broad thesis: LLMs have completely changed the calculus in buy vs. build decisions, making it plausible to build even large software projects with unbelievable speed.

However, although they've dramatically reduced costs, they haven't brought them to _zero_. Even when most of the work involved is prompting LLMs, the hours it takes an engineer to oversee an initial buildout and then patch projects with new features and bug fixes add up. At 40 hours a week, a person paid $200k/year costs ~$96/hour, so it doesn't take that many hours of prompting and context switching before a software subscription with reasonable pricing wins out over a homegrown buildout from a cost standpoint.

The operative keyword here is _reasonable_ pricing. The days when a SaaS company can charge $500/head/mo for relatively basic software that's plausible to rebuild with an LLM might be coming to an end. 50 seats at $500/mo is $25k/mo, and that's plenty of incentive for an organization to tackle an internal rebuild via LLM even when the project's completely orthogonal to their domain.

---

I think the other shoe here hasn't quite dropped. We're still manically fixated on our LLMs and the short-term results we can produce with them, but haven't fully acknowledged the ongoing price tag that this new software has. The next phase will be the age-old realization that just because you can doesn't mean you _should_.

When plotting software cost versus complexity, I postulate the existence of a _zone of viability_, in which a small software business may continue to successfully sell a product that's sufficiently complex not to be trivially reproducible by LLM, and at a cost that's reasonable enough not to provide strong incentive for potential buyers to reimplement it themselves.

<img src="/assets/images/minimum-viable-unit/zone-of-viability.svg" alt="Zone of viability" class="img_constrained" loading="lazy">

[River's](https://github.com/riverqueue/river) an open source project that makes almost all job-related features (periodic jobs, scheduled jobs, unique jobs, web UI, …) available for free, but reserves some advanced features (workflows, sequential jobs, concurrently limited jobs, …) and billing capability (billing by invoice) for a Pro version that we charge for. An LLM could reproduce the latter features, but we’ve put enough thought into their API design and performance properties that it’d take some work to get back to something of similar fidelity.

In terms of price, we used a sublinearly scaling pricing model based on team size rather than head count, starting at $125/mo for up to 20 developers, and scaling up to a multiple of that for an unlimited site license. So for a small-to-medium development team, $125/mo is the all-in cost across everyone.

I'm hoping that it's got a sufficient price-versus-complexity tradeoff to fall inside that band of viability, and for the next few months at least I'm going to be banking my livelihood on it by working on it full time instead of trying to immediately jump to another job.

---

<img src="/photographs/nanoglyphs/052-adrift/sofia-portrait@2x.jpg" alt="Sofia (portrait)" class="wide_portrait" loading="lazy">

## Balkan Ruby (#balkan-ruby)

A few weeks ago, on a whim I attended [Balkan Ruby](https://balkanruby.com/). I was in Europe already, the Anthropic acquisition was in its last throes of closing, and I didn't have much to do.

I haven't actually coded Ruby professionally in about a year now, and on darker days of the week I seriously question its plausibility as a modern language, but it's still got the best conferences and nicest people. Bulgaria also rates as one of the places I'd like to see, but it's just enough off the beaten path that without a good reason to go, I'd probably never end up visiting. A Ruby conference creates the perfect excuse. I did the same for Latvia last year where Baltic Ruby (that's "Baltic" not "Balkan") was being held, and that trip worked out spectacularly.

<img src="/photographs/nanoglyphs/052-adrift/sofia@2x.jpg" alt="Sofia" class="wide" loading="lazy">

Balkan Ruby takes place in Bulgaria's capital, Sofia. I didn't know anything about it coming in, and was pleasantly surprised. It's got all the hallmarks a lot of us North Americans are looking for in European cities---beautiful architecture, grand churches, a vibrant, walkable downtown, and even a nice train system from the airport---but with a fraction of the cost and tourist crowds that you'd find in more popular destinations like France, Italy, or Spain. With a population of 6.7M and no other countries with a shared language (I learned from other attendees that Bulgarian's closest sister tongue is _Macedonian_, also a small language spoken by ~2M people), a lot of people speak English quite well, making it easy to get around as a tourist.

The venue was so on the nose. It was held at Sofia's "Earth and Man" national museum, which houses one of the largest mineralogical displays in the world. So you watch talks on Ruby even as you're surrounded by immense gems and geodes inches from your seat. It's a 25 minute walk from the city center where most of the hotels are, absolutely perfect for a little morning exercise before sitting down for most of the day, all along nice parks and picturesque promenades.

<img src="/photographs/nanoglyphs/052-adrift/earth-and-man@2x.jpg" alt="Balkan Ruby at the Museum of Earth and Man" class="wide" loading="lazy">

A few conference takeaways: Ruby's getting close to a full [server-side rendered Rails-compatible reactive framework](https://github.com/marcoroth/reactionview), fibers [continue to be popular](https://github.com/rage-rb/rage), and I learned that even when [Stephen's not present](https://fractaledmind.com/), someone else will be there to pick up the slack in telling the Postgres-to-SQLite migration story. (Is it time for us Postgres fanatics to start worrying yet? I keep on telling myself that I will once SQLite adds a sixth data type.)

---

## On odd coincidences (#odd-coincidences)

The week rolled into a couple odd, once-in-multi-decade coincidences that just happened to occur while we were there. The first was that on the last day of the conference, [a person was killed by a bear](https://tvpworld.com/93309644/bulgarian-man-killed-by-brown-bear-in-rare-attack-near-sofia), Bulgaria's first fatal bear attack in almost two decades (2010). (To be clear, the person killed was _not_ a conference-goer, but it happened the day of.)

The organizers had kindly arranged a post-conference hike in the gorgeous surroundings of Sofia. At the after-party (held at a craft beer bar called "KANAAL"), the subject came up while I was talking to one of them, with the conversation going roughly like this:

Organizer: "Huh, today a guy was mauled to death by a bear in the mountains."

Me: "Ahaha, you're joking, right?"

"No. It was the first death by bear in Bulgaria in 16 years."

"Well, at least that happened somewhere far away, right?"

"No. In the Vitosha mountains. They're a few kilometers south of Sofia."

"Ahaha, you're _definitely_ joking now, right?"

"No. It's where you're going hiking tomorrow."

I smiled, and joked weakly that he should know that most conferences try to aim for a running average of zero attendees killed by bears. Maybe it's different in the Balkans, but even one fatality is generally considered one too many in most countries. Even RubyKaigi in Japan, a country that regularly posts an [unusual number of bear attacks](https://www.japantimes.co.jp/news/2025/12/06/japan/society/japan-casualties-record/) (13 deaths and 217 injuries last year), still has no conference fatalities on record.

---

<img src="/photographs/nanoglyphs/052-adrift/golden-bridges@2x.jpg" alt="Zlatnite Mostove (Golden Bridges)" class="wide" loading="lazy">

It was an excellent hike, and thankfully, no bears in sight. We took a bus up to a unique feature on the mountain called "Zlatnite Mostove", or "The Golden Bridges". It's a river of oval stones that cascades down the slope, with an actual stream running below it. It's called "golden" due to the color of the lichens covering the rocks, and "bridges" because the stones can be used to cross the water. Very beautiful at any rate.

We hiked over to the Kopitoto TV Tower, a tall feature up on the mountain easily visible from Sofia. The area is the terminus of an old lift built in the 60s during the Soviet era, which fell into disrepair decades ago.

An ominous cloud had clung to the mountain all day, but we got just enough of an opening for a reasonable view of the city below. Diffuse light and gorgeous forest colors provided braindead conditions for photography -- aim a camera in any direction and you'd get something reasonable. We got a little rain throughout the day, and at one point one of our guides broke out a set of ponchos, but it was never enough to be a serious problem.

On the way out we stopped at [Septemvri Hut](https://septemvrihut.bg/) for a warm meal and tea with locally produced honey. An advantage of the area (compared to North America) is that there's enough development to make quite a number of these little huts available, enough that they're not completely smothered with crowds all the time, creating a relaxing, rustic ambience.

---

The aftermath of the hike led to coincidence number two: as we were taking the metro back to our hotels, our host, examining his phone, exclaimed, "Bulgaria has won Eurovision!"

Again, this seemed a little too coincidental to be real, but it was true. More than 20 years after their first entry (2005, Kyiv), that day Bulgaria had won Eurovision with the song "Bangaranga" by Dara, winning both the jury and popular vote.

I'll leave it at that. I watched the winning song and was reminded that Eurovision ... isn't for me. That is unless it's [_Eurovision Song Contest: The Story of Fire Saga_](https://www.imdb.com/title/tt8580274/) with Will Ferrell and Rachel McAdams, which is a masterpiece for the ages.

Until next week.

Brandur

<img src="/photographs/nanoglyphs/052-adrift/gondola@2x.jpg" alt="Gondola" class="wide_portrait" loading="lazy">
