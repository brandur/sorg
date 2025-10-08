+++
image_alt = "Komodo"
# image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/046-yaml/komodo-hill@2x.jpg"
published_at = 2025-10-06T21:52:15+08:00
title = "Negative, I am a YAML wrangler"
hook = "TODO"
+++

Readers --

So they _do_ have Starlink here. A nice upgrade from the last time I was in this part of the globe.

I'm always late to the game using all these things, but it's my "wow, this technology is amazing" phase right now. Located on a remote island in the middle of the Pacific and getting 50 mbps up / 9.25 mbps down. That's as good as Comcast can do in the middle of downtown Seattle, 100 meters from the offices of some of the largest tech companies on Earth.

Today, Kubernetes, Komodo, and a rant about TikTok.

---

## Return to K8S (#return-to-k8s)

I've been doing a lot of Kubernetes work for the last few months. The last time I touched Kubernetes was about ten years ago, deploying a podcast hosting service I was writing, and which turned out to be about as successful as the other 1,000 nascent podcast services of the era. There's new stuff, but a lot has stayed the same.

A few high level impressions below. None of this will new to anyone who uses Kubernetes, which has got to be a lot of you at this point.

* Oh boy is it impressively complicated. Pods, deployments, services, replica sets, and config maps, PDBs (pod disruption budget). The amount of YAML you need to achieve lift off is surprising even when you know going in there's going to be a lot of YAML involved.

* [Overlays](https://notes.kodekloud.com/docs/Certified-Kubernetes-Application-Developer-CKAD/2025-Updates-Kustomize-Basics/Overlays) are a thing now. You don't just write YAML, you write a YAML _template_ which then generates the rest of the YAML. We have one overlay per region, generating to 50-100 per change.

* It's a little surprising there isn't a `kubectl lint`? We had some remarkably stupid errors make it all the way to rollout like breached name length limits. (I realize there's third party lints, but a built-in solution would be so obvious.)

* It's admirably functional. In my experience so far, pods schedule how they're supposed to, deploys just work, and features work as advertised. This is not a given for product this complex.

* Is Kubernetes ever really going to be a general purpose platform for anyone? I'd say the same thing about it that I did years ago: Kubernetes is great once you're an organization big enough to have your own Kubernetes _team_. Smaller than that, and you'll be spending as much time configuring deployments as writing software to deploy.

But my main, overarching finding is that [Argo CD](https://github.com/argoproj/argo-cd) is something that exists now, and it's amazing.

<img src="https://argo-cd.readthedocs.io/en/stable/assets/argocd-ui.gif" alt="Argo CI" class="wide" loading="lazy">

Argo CD provides a gorgeous front end for Kubernetes, with an interactive UI showing a clear topology for what a namespace looks like and all the deployments therein, status of each pod, and best of all, a way to zoom in and get log traces for each grouped by container. Ours is set up to spring open at the touch of a button in any deployment and namespace with `snowkube open argocd <DEPLOY> <NS>`.

All of this information is available by Kubernetes CLI of course, but gated behind its awkward UX, in a way that's not intuitive or discoverable, and makes use by non-experts all but impossible.

I'd go as far to as to say that Argo CI is as important to a modern Kubernetes deployment as Kubernetes. What an incredible package.

---

We've got a pretty legitimate infrastructure-as-code set up going on right now. To take out a change, bump a commit number in a Go file, generate overlays, and send it up in a pull request. Upon merge, the change in the repo is picked up automatically and rolled out.

An Argo post-sync hook runs a migrator to handle any DB migrations. (A post-sync hook is used instead of an init container to guarantee migrations run before _all_ deployments rather than just the one with the init container configured.)

It's a little less clear what happens when, say someone raises an index non-concurrently or adds a `NOT NULL` column to a large table. Those emergency procedures are still being tooled out.

---

## Indonesia so far (#indonesia)

We just finished up day one of diving. Two days lost to travel (17 hours in the air, 5 hours layover, 13 hours time zone difference), two days in Bali, a one hour hop over to Labuan Bajo, a half day's wait and 90 minute boat ride over to the resort.

Our neighbors here are a tiny Indonesian village of a few hundred and a small herd of water buffalo that live up in the hills. We're told there's a reasonable incidence of macaques in the jungle (so far only one unclear sighting so far). Komodo dragons are known to occasionally range out to this island (they saw one next to the resort last week), but they're rare this far out.

Great day out on the water:

* Dive 1: Octopus, mantis shrimp, sea turtles.
* Dive 2: A site known for its manta ray cleaning stations, and 45+ minutes of the dive were dedicated to watching that. The largest of the three mantas had a wingspan over 20 feet.
* Dive 3: Muck dive, with sightings of 3 sea horses, two pipefish, one ornate ghost pipefish (a very different looking creature), zebra crab, and a half dozen nudis of different kinds.

It was such a good day that we could probably call the trip here if we needed to, but if we stick to our plan, there's still 60+ dives to go!

<img src="/photographs/nanoglyphs/046-yaml/vira-bali@2x.jpg" alt="Vira Bali" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/046-yaml/labuan-bajo-1@2x.jpg" alt="Labuan Bajo 1" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/046-yaml/labuan-bajo-2@2x.jpg" alt="Labuan Bajo 2" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/046-yaml/labuan-bajo-3@2x.jpg" alt="Labuan Bajo 3" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/046-yaml/labuan-bajo-4@2x.jpg" alt="Labuan Bajo 4" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/046-yaml/labuan-bajo-5@2x.jpg" alt="Labuan Bajo 5" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/046-yaml/scuba-junkies-1@2x.jpg" alt="SCUBA Junkies" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/046-yaml/scuba-junkies-2@2x.jpg" alt="SCUBA Junkies" class="wide" loading="lazy">

---

## The other opiate (#tik-tok)

During a dinner in Bali we were treated to a performance of [traditional legong dancing](https://en.wikipedia.org/wiki/Legong). The dancers themselves, made up in their impressive costumes and demonstrating the art’s and precise, refined movements and emotive expressions, seemed to be otherworldly creatures, floating above us like ethereal spirits.

But as each dancer sat backstage waiting for their turn, they outed themselves as truly like the rest of us. Rather than engage in some ritual of eastern enlightenment like meditation, mantra, or silent contemplation, they did what the rest of the world does. They waited, faces buried into phones … staring into the abyss of TikTok, as intent on the content as they were utterly helpless to turn away from it.

(I use the name “TikTok” as a convenient stand in for either TikTok itself, or another algorithmic black hole app. Instagram Reels or YouTube Shorts is exactly the same thing.)

I look out at the table next to us to see Chinese tourists … on TikTok. Walk out onto the street and there’s a dozen waiting motorcycle taxis. While they bide their time waiting for a fare … TikTok. Store owners next door … TikTok. We head to our remote dive island and watch the local fishermen taking a break on … TikTok. Our boat’s deck hands, smoking cigarettes and … TikTok. In every case, slaved to their screen.

The algorithmic app is Huxley’s soma in its most distilled form, ensuring dull pacification and an unfulfilling life free of creativity, ambition, or sophistication beyond the most cursory thinking. It's insidious because when you're cooking heroin over a tetanus-contaminated spoon with a butane torch, or sending fentanyl into the side of your arm with a hypodermic needle on the corner of a filthy street corner in San Francisco, or even just popping the day's fifth benzo, there's an awareness that you're not making the best decisions in life. Even the doped up junkie on his 20th "bowl of cereal" this morning kinda knows that this isn't a great situation and feels the occasional tinge of existential guilt. But smartphones? There's no such adaptation or obvious faux pax. These are devices of higher learning right? They like, teach us stuff.

In most parts of the world now, it’s perfectly within social bounds to have your smartphone blaring on speakerphone as you watch TikTok on a bus, at the airport, or in a restaurant. Aside from being annoying as hell, it’s not good for the TikTok addict either. Like a bad drug our alcohol dependence, their capacity to moderate diminishes with every new Reel that comes on screen, prefrontal cortex weakening one neuron at a time.

---

The genie's not going back in the bottle. The plan's already been carried out and the technology is spread far and wide. It’s too late to save the world, so it's up to each person to save themselves:

* Admit that a problem even exists. Smartphones are not a value equation of pure benefit.
* Recognize that the human brain isn't well adapted to defend itself from the endorphin streams of algorithmic crack.
* Resist it. Ideally, don't use the apps. If you do, minimize exposure time. Train focus.

If you can turn on your phone and use it to accomplish a task without it becoming an inescapable, attention sucking vortex for the next hour, it’s a competitive advantage. Most of the world can’t.

If you can pick up a book and read it all the way to the end, it’s a competitive advantage. Most of the world can’t.

If you can sit with your own thoughts for two minutes without being pulled into your smartphone, it's a competitive advantage. Most of the world can't.

Depending on your disposition, a morbidly comforting thought if you care about adversarial geopolitics and the west’s position therein: aside from a few exceptional cases, the third world will _never_ be catching up. During the darkest parts of China’s opium crisis in the early 1900s, 10s of millions were addicted to the drug, paralyzing the country for decades, letting the west pull even further ahead.

Well, the masses have a new form of opium, and ironically it's China this time that got it started. TikTok's not physically wasting in the same way, but nonetheless guarantees permanent arrested development. In the case of this new drug though, the victims range in the 100s of millions, possibly the billions, with countries or sociographic stratas (i.e. the poor) where residents own _only_ smartphones the hardest hit. A lesser calamity on a much wider scale. They’ll be lucky to be alone with their thoughts for five consecutive minutes in their entire lifetimes, let alone build an Apple, Google, or Nvidia.