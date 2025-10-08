+++
image_alt = "Komodo hill"
# image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/046-yaml/komodo-hill@2x.jpg"
published_at = 2025-10-09T06:45:12+08:00
title = "Negative, I am a YAML wrangler"
hook = "A short issue on Kubernetes and progress through Indonesia."
+++

Readers --

So they _do_ have Starlink here --- a nice upgrade from the last time I was in this part of the globe.

I'm always late to the game when it comes to these things, but it's my "wow, this technology is amazing" phase right now. Located on a remote island in the middle of the Pacific and getting 50 mbps up / 9.25 mbps down. That's as good as Comcast can do in the middle of downtown Seattle, 100 meters from the offices of some of the largest tech companies on Earth.

I don't know if it's actually a good thing. I can send this newsletter, which is good, but the locals (and a few westerners too) huddle underneath the Starlink antenna like the life-giving warmth of a hearth, their faces glued to TikTok/Instagram Reels/YouTube Shorts ~55 minutes out of every hour. They enjoy this arrested, vegetative state like any drug user enjoys their high, but it's tragic to watch. On warm, otherwise peaceful evenings, the sound of TikTok on speakerphone and ear-shattering iPhone _¡DINGs!_ ring out over the water.

Today: Kubernetes, Komodo.

---

## Return to K8S (#return-to-k8s)

I've been doing a lot of Kubernetes work for the last few months. The last time I touched Kubernetes was about ten years ago, deploying a podcast hosting service I was writing, and which turned out to be about as successful as the other 1,000 nascent podcast services of the era (as in, it doesn't exist anymore). There's new stuff, but a lot has stayed the same.

A few high level impressions below. None of this will new to anyone who uses Kubernetes, which, by now, must be many of you.

* Oh boy is it impressively complicated. Pods, deployments, services, replica sets, and config maps, PDBs (pod disruption budget). The amount of YAML you need to achieve lift-off is surprising even when you know going in there's going to be a lot of YAML involved.

* [Overlays](https://notes.kodekloud.com/docs/Certified-Kubernetes-Application-Developer-CKAD/2025-Updates-Kustomize-Basics/Overlays) are a thing now. You don't just write YAML, you write a YAML _template_ which then generates the rest of the YAML. We have one overlay per region, generating to 50-100 per change.

* It's a little surprising there isn't a `kubectl lint`? We had some remarkably stupid errors make it all the way to rollout like breached name length limits. (I know there are third-party linters, but a built-in solution would be such an obviously good idea.)

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

We've wrapped up three days of diving now. Two days lost to travel (17 hours in the air, 5-hour layover, 13 hours time zone difference), two days in Bali, a one hour hop over to Labuan Bajo, a half day's wait and 90-minute boat ride over to the resort.

Our neighbors here are a tiny Indonesian village of a few hundred and a small herd of water buffalo that live up in the hills. We're told there's a reasonable incidence of macaques in the jungle (only a few obscured sightings so far). Komodo dragons are known to occasionally range out to this island (they saw one next to the resort last week), but they're rare this far out.

Dives on day one:

* Dive 1: Octopus, mantis shrimp, sea turtles.
* Dive 2: A site known for its manta ray cleaning stations, and 45+ minutes of the dive were dedicated to watching that. The largest of the three mantas had a wingspan of over 20 feet.
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