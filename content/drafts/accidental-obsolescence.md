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
more common than the use of superlatives during an Apple
keynote (_Incredible!_, _Unprecedented!_,
_Revolutionary!_).

I was a little behind the hardware curve with a 3+ year old
iPhone 6, but even so, using it had become comically slow.
Splashing screens lasting ten seconds or longer weren't
uncommon. Simple actions like scrolling down a page became
stuttering and tedious. The new OS brought some minor
quality of life improvements [1], but with a crushing cost.

!fig src="/assets/accidental-obsolescence/ios11.png" caption="A signature feature of iOS 11: a customizable (and more useful) Control Center."

Niklaus Wirth described this phenomena in a mid-90s paper
entitled [_A Plea for Lean Software_][leansoftware], and
its since become an adage known as [Wirth's law][wirth],
stating that software is getting slower more rapidly than
hardware is getting faster. Unfortunately for the entire
world, Wirth's law has shown a tireless doggedness that
makes Moore's law look ill-conceived and unreliable by
comparison.

A cynic might posit that companies are doing this on
purpose as part of a program of [planned
obsolescence][planned], but the reality is that the source
of these usability losses is much more innocent than that.
Companies are incentivized to favor development speed, new
features, and visual gimmickry over performance. Developers
are regularly building products only on pristine versions
of recent hardware iterations rather than what's likely to
be out in the field. And while Apple is worst-in-class in
selecting for new and longer genie/fade/slide animations
over the usability of its software, it's happening
everywhere.

Rather than being a maliciously engineered schedule of
deprecation, the reality is much more akin to one of
unintended consequences. It's not planned obsolescence,
it's ***accidental obsolescence***. 

!fig src="/assets/accidental-obsolescence/ruins.jpg" caption="Adequate performance seems destined to be lost to time."

## Longevity's in dire straits (#longevity)

It used to be that new software (and any performance
implications that it carried with it) could be avoided by
simply not updating -- early versions of Windows and
Internet Explorer were still common years after they were
obsolete because users had the option to keep running them
and save on the considerable costs of upgrading to modern
hardware.

Some software could not be updated at all. Maybe the best
example of this is software that shipped on physical media
like game cartridges or the firmware of the game systems
they ran on. Developers had one opportunity to get it
right; after that master disc went out the door, it was too
late to make changes. Absolutely terrible if a bug is ever
found; but also somewhat advantageous in that performance
could never be accidentally degraded.

!fig src="/assets/accidental-obsolescence/cartridge.jpg" caption="A SNES cartridge. Software that once shipped, was set in stone."

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

These days by comparison, you might pop a brand new game
that you bought yesterday into your PS4, only to find that
it already wants to update itself before you can play.

iOS

Chrome

Mac OS, even Windows.

Good for security.

But even better, it allows builders to move forward without
thinking about the maintenance overhead of older systems.

## Performance as a feature (#performance-as-a-feature)

### Techniques (#techniques)

Cycle counting. DJB's [Ed25519 public-key signature system][djb]:

> * **Fast single-signature verification.** The software
>   takes only 273364 cycles to verify a signature on
>   Intel's widely deployed Nehalem/Westmere lines of CPUs.
>   (This performance measurement is for short messages;
>   for very long messages, verification time is dominated
>   by hashing time.) Nehalem and Westmere include all Core
>   i7, i5, and i3 CPUs released between 2008 and 2010, and
>   most Xeon CPUs released in the same period.
> * **Even faster batch verification.** The software
>   performs a batch of 64 separate signature verifications
>   (verifying 64 signatures of 64 messages under 64 public
>   keys) in only 8.55 million cycles, i.e., under 134000
>   cycles per signature. The software fits easily into L1
>   cache, so contention between cores is negligible: a
>   quad-core 2.4GHz Westmere verifies 71000 signatures per
>   second, while keeping the maximum verification latency
>   below 4 milliseconds.
> * **Very fast signing.** The software takes only 87548
>   cycles to sign a message. A quad-core 2.4GHz Westmere
>   signs 109000 messages per second.

[Rust compiler performance data][rustperf]. Number of CPU
instructions to compile various projects like a "hello
world", but also real-world software like Hyper, Syn, and
Servo.

Benchmarks built-in. `go test` and `cargo bench`.

```
running 1 test
test bench_xor_1000_ints ... bench:       131 ns/iter (+/- 3)

test result: ok. 0 passed; 0 failed; 0 ignored; 1 measured
```

[Firefox Quantum][quantum].

> Firefox Quantum is consistently about 2X faster than Firefox was.

Buy underpowered computers.

### Exercising restraint (#restraint)

Maybe that third loop on that spiral animation isn't
necessary.

JavaScript and Electron.

[1] TODO: quality of life iOS 11.

[djb]: https://ed25519.cr.yp.to/
[leansoftware]: http://doi.ieeecomputersociety.org/10.1109/2.348001
[planned]: https://en.wikipedia.org/wiki/Planned_obsolescence
[quantum]: https://blog.mozilla.org/firefox/quantum-performance-test/
[rustperf]: https://perf.rust-lang.org/?start=2017-12-01&end=2017-12-10&absolute=true&stat=instructions%3Au
[wirth]: https://en.wikipedia.org/wiki/Wirth%27s_law
