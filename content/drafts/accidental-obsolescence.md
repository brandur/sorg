---
title: Software Enhancement and Accidental Obsolescence
published_at: 2017-11-22T08:50:10Z
location: Tokyo
hook: TODO
attributions: SNES cartridge image by
  <strong><a href="https://www.flickr.com/photos/bochalla/4753073993/">Bryan
  Ochalla</a></strong>. Licensed under Creative Commons CC BY-SA 2.0.
---

Like many of you, I recently updated my phone to iOS 11.
Although complaints about increased sluggishness are a
common staple of any iOS upgrade, anecdotally iOS 11 seems
to be worse, with reports of performance degradations even
more common than the use of over-the-top superlatives
during an Apple keynote (_Incredible_!, _Unpredecented_!,
_Revolutionary_!).

I'm a little behind the hardware curve with a 3+ year old
iPhone 6, and using it has become comically slow. It's not
uncommon to experience splash screens that last ten seconds
or longer. Even simple actions like scrolling down a page
to be stuttering and tedious. The new version brought some
minor quality of life improvements [1], but with a crushing
cost.

!fig src="/assets/autoupdate/ios11.png" caption="A signature feature of iOS 11: a customizable (and more useful) Control Center."

Niklaus Wirth described this phenomena in a mid-90s paper
entitled [_A Plea for Lean Software_][leansoftware], and
its since become an adage known as [Wirth's law][wirth]
stating that software is getting slower more rapidly than
hardware is getting faster. Unfortunately for the entire
world, Wirth's law has shown a tireless doggedness that
makes Moore's law look unreliable by comparison.

A cynic might posit that companies are doing this on
purpose as part of a program of [planned
obsolescence][planned], but the reality is that it's much
more like ***accidental obsolescence***. Companies are
incentivized to favor development speed, new features, and
visual gimmickry over performance. And while Apple is the
worst in class in selecting for new and longer genie
animations over the usability of its software, it's
happening everywhere -- developers are regularly building
products only on pristine versions of recent hardware
iterations, and older platforms are suffering for it.

TODO: Photo of ruins

## Longevity's in dire straits (#longevity)

It used to be that new software (and any performance
implications that it carried with it) could be avoided by
simply not updating -- early versions of Windows and
Internet Explorer were still common years after they were
obsolete because users had the option to keep running them
and save on the considerable costs to upgrade to something
more modern.

But as the security repercussions of this sort of
stagnation become more apparent and autoupdate becomes more
ubiquitous, this is no longer going to be an option. And
while the industry's trained us to have a built-in expectation
that our smartphones and computers should be replaced every
few years, if devices like our TVs, cameras, and IOT power
outlets are updated with new firmware that has as little
regard for old platforms as iOS 11, we'll all be stuck with
devices that are best semi-usable within just a few years.

For better or for worse this is the way that the world's
moving, and we should plan for it.




One of the greatest advances in software in the last ten
years.

Maybe the best ever example of this is software that
shipped on physical media. Developers had one opportunity
to get it right; after that master disc went out the door,
it was too late to make changes.

!fig src="/assets/autoupdate/cartridge.jpg" caption="A SNES cartridge. Software that once shipped, was set in stone."

These days by comparison, you might pop a brand new game
that you bought yesterday into your PS4, only to find that
it already wants to update itself before you can play.

iOS

Chrome

Mac OS, even Windows.

Good for security.

But even better, it allows builders to move forward without
thinking about the maintenance overhead of older systems.

## Autoupdate in packages (#packages)

### Static builds (#static-builds)

Some success: Postgres crypto.

## Accidental obsolescence (#accidental-obsolescencec)

### Performance as a feature (#performance)

DJB cycle counting for encryption.

Rust compile benchmarks.

Benchmarks built-in.

Firefox Quantum.

### Exercising restraint (#restraint)

Maybe that third loop on that spiral animation isn't
necessary.

JavaScript and Electron.

[1] TODO: quality of life iOS 11.

[leansoftware]: http://doi.ieeecomputersociety.org/10.1109/2.348001
[planned]: https://TODO
[wirth]: https://en.wikipedia.org/wiki/Wirth%27s_law
