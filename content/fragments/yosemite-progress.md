---
title: Yosemite & Progress Bars
published_at: 2014-11-10T16:37:30Z
---

I recently upgraded my system from Mavericks to Yosemite and ran into a pretty
scary situation resulting from poor installer UX. For anyone with a non-trivial
number of files in their `/usr/local` directory, Yosemite's installer will
stall on its "2 minutes remaining" for much longer than two minutes; I estimate
that mine lasted upwards of two hours, but I saw reports of as much as six
hours.

Yosemite's "2 minutes remaining" trap is so common that it's now a
well-documented effect across the net, but during my initial installation pass,
I didn't know about it. After waiting a little under two hours, I assumed that
the installer had crashed, and did a hard reboot to try another pass. That
turned out to be a big mistake, and ended with me spending the rest of my day
finding and imaging a USB stick so that I could recover my system.

Yosemite's installer is no better on a clean install either; while installing
it another one of my systems, I noticed that a couple "2 seconds remaining"
screens that lasted longer than ten minutes. In fact, it was rare to see any
screen that wasn't off by less than an order of magnitude.

It's telling that we still can't get progress bars right even on one of the
world's most modern operating systems. Back when I built Windows apps, we'd
often run into situations where performing the calculations to have an accurate
status bar would be almost as expensive as just doing the work, so we'd favor
speed over correctness. Some of the code was downright dirty; for example,
making a rough estimate that routing points would take 40 times longer to load
than the routes themselves. As you may imagine, this resulted in some
interesting experiences for users with outlying data sets.

Anyway, the lesson from these is clear: progress bars are often wrong.
Especially where Yosemite is considered, just wait no matter what the progress
screen says or how long it's been.
