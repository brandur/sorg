+++
image_alt = "A bank of windows in Lyon."
# image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/048-llms/lyon-windows@2x.jpg"
published_at = 2026-02-01T10:43:33-08:00
title = "The AI Edition"
hook = "Mixed thoughts on LLMs, including the highs and lows (will I have a job in two years?), Ambon, and psychadelic frogfish."
+++

Where did January go?

I can't answer that question. I got mine going in Austin doing a colocated kickoff where we tried to get the basic shape of our new [Stainless](/nanoglyphs/047-stainless) product in place. That was followed by a few days at Stainless West (San Francisco), and now I'm back in Seattle. Somewhere in that span, January disappeared.

Product-wise, we got to a point where what we built is demo-able, but due to an unconventional choice for a data layer (we're trying some things with DuckDB), we've been running into some performance trouble and have spent the last couple of weeks trying to optimize things. I've learnt a lot about Parquet data formats and [Iceberg](https://iceberg.apache.org/), an open table format that lives in files or blob stores.

It's not quite ready to talk about publicly yet, but it will be soon. By the end of February maybe?

Notably, the project was my first where LLMs were not only a novelty, but a core piece of tooling that made it run. A core member of the team even. They are terrifyingly good. I'm a traditionalist, and was constantly looking for evidence of mistakes or malpractice so I could say loudly "this is sh*t" and declare that them unfit for use. I found none. Instead, I threw increasingly large chunks of work their way, and almost every time, I got results back that were close to what I would've written by hand, except slightly better. They've unquestionably been an incredible boon to me, so far. For society and the world at large? I'm not so sure.

---

<img src="/photographs/nanoglyphs/048-llms/lyon-montee-de-la-grande-cote@2x.jpg" alt="Lyon Montée de la Grande Côte" class="wide_portrait" loading="lazy">

## Lyon (#lyon)

The photos from this issue are from Lyon, France. Shortly after the Snowflake acquisition last year a small team of Crunchy and Snowflake employees went on-site together (Lyon sounds like an exotic place to hold such a thing, but some of our colleagues live in the city, and hotels there are a third of the price as staying in a bad airport-like hotel in Menlo Park or Bellevue) to get as much done on an MVP as we could in a week.

Gorgeous city. By day two I had a favorite walking route up to the famous Basilique on top of the hill. I'd descend back down down the other side, taking a different way back through Croix-Rousse to the hotel every day, and there was always something new to discover. Lookout points up on the city's famous hills provide incredible viewing opportunities to see out over its distinctive roofing. Some easy to find, others less so. Come evening, especially in and around Vieux Lyon, vibrant streets lined up and down with outdoor patios for bars and restaurants. When a visitor imagines stereotypical France in their mind's eye, this is what they see.

Ye gods I wish we still built cities like this. Give me narrow streets and pedestrian plazas! If there was one in the US--_anywhere_--I'd consider moving there tomorrow.

---

<img src="/photographs/nanoglyphs/048-llms/lyon-jardin-des-curiosites@2x.jpg" alt="Lyon Jardin des Curiosités" class="wide" loading="lazy">

## Confessions of a former skeptic (#skeptic)

This is the first time I've written anything in detail about AI/LLMs. That's not because the subject's unimportant, but rather an inditement of my own lack of creative output. Also, I figure that if if you want to read about LLMs, you need only open any newspaper, magazine, or newsletter in existence and turn to the first page. A whole cottage industry of a hundred thousand newly minted AI experts. If you ask me, the world doesn't need even more LLM writer. The market's saturated.

I avoided LLM use for a long time. Longer than most outside the tech industry, let alone in it. I gave in last year around the time Google added automatic Gemini answers to their search results. This wasn't something that was going away, so best to embrace it.

I hoped LLMs would be worse than they are. In my ideal world, they'd produce convincing output, but which was recognizable as B-grade semi-slop to a learned expert. This would've relegated them squarely to a bucket of useful tools, but not ones that we could expect to replace the expertise of veteran humans anytime soon. This would be convenient for me as it'd justify my continued employment.

No such luck. Sometime in the last year, models graduated from "neat party trick" to "holy sh*t this is _insane_". My generation (elder millenial) was quite lucky in that we experienced the invention of the home PC, and later the invention of the smartphone. Two of the largest paradigm shifts in history. Even stacked against those monsters, I think LLMs will be the most impactful technological event horizon of my generation.

I'd been hearing testimony of impressive results for many months, but I wasn't truly convinced until last December. For a hackathon, one of my colleagues vibe coded a working prototype of a product we'd been slated to work on. He started around 1 PM and was done by the end of the work day, and while it was a prototype, he managed to hit some stretch goals so that his demo even had a Jobs-style, "oh, and one more thing" clause at the end. Before LLMs, this would've been a week of work. It was incredible -- a technological unlock that made previously impossible things possible.

Amazing, and scary. Like seeing the gruesome wreck photos after an accident involving drinking and driving, or the X-ray photos of a smoker's black lungs, it set off alarm bells screaming in my head. The choice now seemed quite clear: learn to use these tools, or be left in the dust.

---

I've begrudgingly learned to love them. One of the first things that really WOWed me is that in conjunction with a local `AGENTS.md` file, it's not only possible to have an LLM generate code, but generate _conventional_ code. Ours makes sure to [mark tests as t.Parallel()](/t-parallel), use a [parallel test bundle](/fragments/parallel-test-bundle) as appropriate, and alphabetizes fields on structs where there's no other sorting rationale in play (something I'm kind of a hardass about). We have an extremely comprehensive list of required linters (every one of them in [golangci-lint](https://golangci-lint.run/) except the really dumb ones). AI will dutifully iterate on code until every one of them passes muster.

Scattered thoughts from my first months:

* I'm big on having unit tests, but don't love writing them, especially seeding a brand new test file. The LLMs now do this for me, generating a first pass that's almost always more comprehensive than what I would've written (but also eerily similar), with me following up with a refinement pass.

* There's a classic refactoring dilemma wherein you need to change a bunch of instances of something which are more than a few, but less than hundreds, say in the range of 50 or so. One approach is just to rely on your quickness on a keyboard and with Vim to brute force the problem. Go through and fix each by hand, taking advantage of efficient hand motion where possible. Another is to write a regex or Vim macro. The latter will take more activation energy, but if done correctly will be done in an instant.  It's not clear in advance which option will be faster because even if you're good at regex, there's still more often than not some stupid backslash that you missed or other mistake you made that requires debugging and would've made the manual change faster.

    With LLMs, we now have the perfect third option: get the LLM to do it. In general this turns out to be fast and correct. By far the best of the three approaches.

* Refactoring in general is so much better. As good as Gopls and LSPs have gotten over the years, the refactoring tools they offer are still dull and not expressive enough. LLMs are the first refactoring tool that truly gets me everything I need and they enable _widespread_ refactoing. Existing code is no longer at risk of ossifying in place the way it was before. LLMs can refactor huge swaths of it quickly and quite safely.

* It's made terminal work so much more plausible. When using infrequent-but-necessary CLI invocations, the amount of work needed to look up the right docs and string together a working command is *absurd*. An example I ran into recently was scaling an AWS Fargate service down and back up, which has an awscli command, but one which Amazon makes stupidly hard to discover. Now, the LLM figures it out.

My LLM is doing a lot of the work I don't love doing, leaving the choice parts for the human). I like writing code and will often still write the core of a feature myself, and with Claude helping me fill the edges and boilerplate.

---

<img src="/photographs/nanoglyphs/048-llms/lyon-skateboarding@2x.jpg" alt="Lyon skateboarding" class="wide" loading="lazy">

## Vibe codin' and beyond (#vibe-codin-and-beyond)

The second order effects of this brave new world are quite visible already. Conspicuously, I received more than a few PR review requests in the last month where it became apparent that I was the first human to have looked at the code as the ostensible author was completely unaware of any specific content. I would hazard to guess that the majority of posts on LinkedIn are now generated by LLM. People like Mitchell Hashimoto (HashiCorp and Ghostty) are considering [closing external pull requests](https://x.com/mitchellh/status/2018458123632283679) on repos due to LLM abuse.

Follow LLM-mania to its logical conclusion and you get to developers maintaining a black box for which they have only a vague understanding of what's inside. Changes are made by the LLM. When there's a bug present, the LLM troubleshoots.

It won't be long before there's no other choice. If you stop walking, you lose muscle mass. If you stop talking, you become inarticulate. If you stop coding and debugging, it won't be long before those skills are gone.

Now, go further. What about the generation of up and coming developers who never even experienced the pre-LLM world? They never develop any specific coding skills, and are only vaguely aware of names like "Go", "Tailwind", and "React" as their LLM builds finished products using these mysterious building blocks.

Does it matter? I'd like to think it does, that artisanal code that's fully comprehensible to a human, and a human who fully understands their code will produce a better operations and product experience. When I described this to a friend recently he countered, "you know, people said the same thing about compilers when they first came out".

Two questions. The most obvious one: in a couple years, will any human still write code? The more existential one: in a couple years, will any human bother _reading_ code?

---

<img src="/photographs/nanoglyphs/048-llms/lyon-roofs@2x.jpg" alt="Lyon roofs" class="wide" loading="lazy">

### Worthless predictions (#worthless-predictions)

A lot of smart people I know are asking, "will I have a job in two years?" LLMs have been around a while, but there's been an inflection point in the last few months. They're still being adopted industry-wide, so the other shoe might be about to drop as they're rolled out universally.

Simplifying to the extreme, I see three broad directions that our industry can go:

1. **Major contraction:** A pronounced reduction in the number of jobs available. Everyone does more work more efficiently. Companies need fewer people, leading to downsizing.

2. **Major expansion:** The optimist's case. LLMs let everyone get more done than ever before. This leads to a broad positive feedback cycle where more goods and services are sold, more products produced, and more money changes hands, ballooning GDP and opening more positions and growth. Even if 5k copywriters are lost, they may be replaced by 10k developers -- still a net expansion of 5k jobs.

3. **Relative stasis:** Companies do more with less, but keep their workforce more or less constant. Expansion happens, revenue grows, but the number of participants stay steady. Companies find that growth is no longer contingent on proportional growth in headcount.

I'm personally partial to option 3. In a conversation about it, I was reminded of [Amdahl's Law](https://en.wikipedia.org/wiki/Amdahl%27s_law):

> the overall performance improvement gained by optimizing a single part of a system is limited by the fraction of time that the improved part is actually used.

It's meant to apply to CPUs, but there's an analog in software development. We speed up writing code by an order of magnitude which you'd hope would lead to an order of magnitude in productive output, but a big piece of the advance might disappear into other parts of the product lifecycle that aren't as easily optimized: design and engineering reviews, Slack conversations, inter- and intra-team coordination, human code review, deployment processes, user support, etc.

Major innovations tend to have less of an effect than people predict. The workplace of today is still more alike the one I graduated into in 2007 than not. The tools are different, and the development stacks have turned over five times since, but squint a little and not much has changed in 20 years. The [Lindy effect](https://en.wikipedia.org/wiki/Lindy_effect) suggests it might be reasonable to bet on its continuation.

On the other hand, when I'm walking downtown Seattle around lunch and look at the other knowledge workers who are out and about, sometimes I think, "but really, what are all of us doing all day?" I've been in corporate America long enough to know the answer: they're talking on Slack, they're coordinating meetings, they're on Zoom calls, they're writing product requirements, they're building slide decks, they're posting on LinkedIn. Mostly tasks that LLMs excel at.

Once again, I worry that my predictions are self-serving because _I_ would still like to have a job in two years. Anything could happen. There's a distinct possibility that we're in the calm before the storm and we're about to live through seismic upheaval.

---

<img src="/photographs/nanoglyphs/048-llms/lyon-hat@2x.jpg" alt="Lyon hat" class="wide" loading="lazy">

## Pride is for inventors (#pride-inventors)

I listened to a podcast the other day that genuinely disturbed me. In it, two veteran programmers one up each other in how little code they write nowadays, leapfrogging each other to make increasingly hyperbolic claims, eventually climaxing in how reading or writing any code _at all_ is an anti-pattern. The disturbing aspect was the tone. They spoke in a state of pure ecstasy of a kind normally reserved for the bedroom.

LLMs are here. I'd never suggest anyone not use them, but a recurring theme in the programming world nowadays: LLM-pilled developers who describe the accomplishments of their LLM as if they were somehow responsible for these advances. Bragging, boasting, unadulterated pride of the loudest and most obnoxious kind. One of these guys laid off 75% of his technical staff two weeks ago, ostensibly due to AI, but with that event so distant in the rearview mirror (two weeks is a long time in the information age), not a moment is spent dwelling on it.

Every time I hear someone say something like this, I want to shake them and say, okay man! So you're telling me that you're using that you use a technology that **you didn't create and which you don't own and which you don't know anything about**, and you use this tool to do all the things that used to be *your* job? Congratulations indeed my friend. Next maybe tell me why someone should pay the $200-300k salary you've become accustomed to when they can hire someone younger and with more energy for $80k who's just as good at your one remaining skill? (Typing the English language into a terminal.)

Many of these people will concede the point that LLMs will lead to a decreased demand for programmers and other white collar jobs, and even find the idea quite thrilling. My question is, how are you so sure that _you_ are going to be the one left standing?

We can't stop this train, but to euphorically cheerlead the LLM-ification of the world? I just can't get there. If you're an inventor or shareholder of Anthropic or OpenAI, you should be excited. If you're a major asset holder, you should be excited. If you're an entrepreneur intending to leverage these technologies to the hilt, then maybe you should be excited. If you're working a salary job and have access to the same LLMs that everyone else does? I don't know man.

---

<img src="/photographs/nanoglyphs/048-llms/lyon-waterfront@2x.jpg" alt="Lyon waterfront" class="wide" loading="lazy">

## LLM anti-compression (#llm-anticompression)

One of the few things LLMs are better at than writing code is text, but having thought about it a lot, a concession that I'm not prepared to make is to have an LLM do my writing.

A conventional practice for execs at Snowflake was to send out what was called a "snippet". Usually on a weekly cadence, these were emails containing personal notes on ongoing action and details on what their divisions were working on. The first thing you notice about "snippets" is the sheer volume of them -- in the default set of on-boarding mailing lists you start getting them from every part of the organization on day one. The second thing you notice about snippets is their length -- comprehensive detail, painstaking even. Essays once a week.

One might even say a *suspicious* amount of detail. Detail that includes a few too many tables, emoji, and emdashes. Yes, most of these were undoubtedly LLM-generated.

But LLM use isn't just reserved for execs. In fact, Gemini was on by default, so everyone who received one of these long scrawls got a short, three point summary on top of it. The summary was so concise and so convenient that most recipients (including yours truly) read nothing further.

You have to step back and appreciate the absurdity of this situation. An executive enters three lines to produce a small novella which he then bulk emails to the rest of the organization. Receivers get an automatic three line summary that ... looks a lot like what the sender wrote in the first place. The novella's read by no one except a few stragglers who aren't in on the joke yet. Is this progress? There's a punch line about information theory in here somewhere.

---

<img src="/photographs/nanoglyphs/048-llms/lyon-street@2x.jpg" alt="Lyon street" class="wide" loading="lazy">

## Certified human (#certified human)

A friend of mine, now in his 40s, had never written a blog post prior to 2026. As of a month ago, he's now a prolific blogger, posting long essays about LLM use on a daily basis. Blog posts are accompanied by only marginally shorter LinkedIn posts, published multiple times a day. After not having produced code in ~20 years, he has four new repositories in the last week, all with long, meandering READMEs containing a lot of lists, tables and emojis.

This might be the grumpy old man coming out in me, but my only reaction is: if it takes you less time to generate this stuff than it takes me to read it, why in the name of Saint Clawd would I ever invest my time looking at anything you've created ever again?

It strikes me as wrong to ask someone to spend their precious time reading text that I didn't feel like it was worth my time to write. I've taken a pretty hard line on this, and for some time now have been actively unsubscribing from anything where I know or suspect the author is more like an "author".

So I wrote this mini-charter for my website: [certified human](/human). Unless something changes significantly, I pledge that all prose here is human written. I'll use LLMs as the best spell and grammar checkers in the world, but only as part of light refinement passes.

This newsletter is chronically late and inconsistent. Don't think that it didn't cross my mind to input a, "In the style of Brandur Leach, write five pages on ..." and fix the problem forever. It might even be better than what you're reading right now.

But again, it doesn't feel right to me. If I have one dark prediction about the future, it's that humans will spend outsized effort to push back against endless waves of LLM-generated spam and slop. I'll try my best to not be a contributor to the situation. Every word written here involved me spending far too much time thinking and fussing over it.

---

<img src="/photographs/nanoglyphs/048-llms/lyon-stairs@2x.jpg" alt="Lyon Jardin des Curiosités" class="wide_portrait" loading="lazy">

---

## Ambon, psychadelic frogfish

Another couple photos from Indonesia. This is the island of Ambon. It's a small one (pop. ~500k), but large enough to have an airport. It's one of the thousands of islands closer to the eastern side of Indonesia. There's probably only a few hundred people in the world who could point to it on a map without help, and that includes people from Ambon.

The dive industry on Ambon isn't gigantic, but it's known for its exceptional muck diving. The island is split in two my a massive inlet. Diving tends to be located between the two halves and reveal a trove of critters including rhinopias, frogfish, seahorses, scorpionfish, and octopuses.

<img src="/photographs/nanoglyphs/048-llms/ambon-spice-divers@2x.jpg" alt="Ambon Spice Divers" class="wide" loading="lazy">

There's a couple really rare fish found in Ambon and nowhere else. One is the Ambon scorpionfish, which we did get to see a couple times. Even for local guides, they're hard to find. On one dive we spent a full hour canvassing the same ~2500 sqft of sand over and over again. At about minute 55 our guide let out an underwater cheer and beckoned excitedly for us to come over, and there it was. I swear I'd passed that exact spot five times already. [1]

<img src="/photographs/nanoglyphs/048-llms/ambon-scorpionfish@2x.jpg" alt="Ambon Scorpionfish" loading="lazy">

The camouflage on these creatures is just amazing. Here's a photo of me and a rhinopias (a kind of scorpionfish) [3]. Can you see it? (Answer key at the bottom.)

<img src="/photographs/nanoglyphs/048-llms/ambon-brandur-rhinopias@2x.jpg" alt="Brandur and rhinopias" loading="lazy">

The _really_ rare spotting would've been a psychadelic frogfish, endemic to only this one place in the entire world. Its face is so distinctive that even non-divers might recognize it from the covers of books or nature TV. Alas, they're somewhat seasonal and hard to find at the best of times. Even having done dozens of dives we never came across one. Maybe next time. [2]

<img src="/photographs/nanoglyphs/048-llms/psychadelic-frogfish@2x.jpg" alt="Psychadelic Frogfish" loading="lazy">

I don't want to sugarcoat things too much on Ambon. Westerners who hate the West have internalized the idea that developed countries in North America and Europe savage the environment while the rest of the world lives in noble harmony with nature. (As a recent example, see James Cameron and his revisionist, heavy-handed _Avatar_ series.) Visit a place like Indonesia and this myth is dispelled in four seconds flat, immediately recognizable not only as false, but a perfect inversion of the truth.

For Ambon and much of Indonesia, the ocean is the *default* place to discard garbage, and the amount of trash one person on the island puts in the ocean in one day is enough to give your average CO<sub>2</sub>-hating, straw-banning coastal activist a heart attack. When Indonesians are dumping all together, they produce a rolling version of the Great Pacific Garbage Patch on a daily basis. Every beach is littered with garbage. The surface is littered with garbage. The ocean floor is littered with garbage. Plastic is an omnipresent companion everywhere you go. The only place you don't find it are the resorts, who knowing that westerners hate seeing this stuff, have made significant investments to have the trash picked clean daily.

The photo below isn't cherry picked. This is what every beach on Ambon looks like at all times, with the exception of the ones right in front of tourist resorts.

<img src="/photographs/nanoglyphs/048-llms/ambon-trash@2x.jpg" alt="Ambon beach trash" class="wide" loading="lazy">

That said, Ambon's a perfect place to spend a week. It's right at the edge of the "tourist frontier" so you still have a few resorts that give you enough comfort that you feel like you're on vacation, but everything else is as authentic as a traveler could ask for.

It's amazing traveling to these places how existential concerns like LLMs disappear. Back home, the coming paradigm shift is clear and present. Every conversation I have that's more than a couple minutes long eventually turns to the subject of LLMs. In a place like Ambon, life goes on the same way that it has for decades. People fish. People cook. People dive. No one thinks about LLMs at all.

Until next week.

<img src="/photographs/nanoglyphs/048-llms/ambon-villas@2x.jpg" alt="Ambon villas" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/048-llms/ambon-garden@2x.jpg" alt="Ambon garden" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/048-llms/ambon-bamboo@2x.jpg" alt="Ambon bamboo" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/048-llms/ambon-sunset@2x.jpg" alt="Ambon sunset" class="wide" loading="lazy">

## Rhinopias highlight (#rhinopias-highlight)

Continued from above, here's that photo of me and the rhinopias again, then the rhinopias again highlighted for easier viewing, then a clearer photo of a rhinopias [4] so you can see what you're supposed to be looking for.

<img src="/photographs/nanoglyphs/048-llms/ambon-brandur-rhinopias-labeled@2x.jpg" alt="Brandur and rhinopias" loading="lazy">

<img src="/photographs/nanoglyphs/048-llms/ambon-brandur-rhinopias-highlighted@2x.jpg" alt="Ambon rhinopias highlighted" loading="lazy">

<img src="/photographs/nanoglyphs/048-llms/ambon-rhinopias@2x.jpg" alt="Rhinopias highlighted" loading="lazy">

[1] Ambon Scorpionfish photo credit [Etienne Gosse](https://www.flickr.com/photos/steve\_childs/233842439/).

[2] Psycadelic Frogfish photo credit [Spice Island Diver's](https://www.instagram.com/p/Cka8p1LLskJ/).

[3] Brandur and rhinopias photo credit Dad.

[4] Rhinopias photo credit [Jason Marks](https://en.wikipedia.org/wiki/Rhinopias_eschmeyeri#/media/File:Rhinopias_eschmeyeri_JRM.jpg).