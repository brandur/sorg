---
hook: We may have sacrificed loading time, responsiveness,
  and accessibility, but at least we've got pretty
  animations and lots of whitespace.
location: Calgary
published_at: 2017-01-07T00:21:51Z
title: The Future of Interfaces
---

Web apps are still being hailed as the future. Many
companies are still developing their main products online,
and even when they're not, Electron-based products are
considered a reasonable alternative.

An ex-colleague who is a major JavaScript pundit recently
called me out on Twitter for gross hyperbole after I
claimed that Slack (an Electron-based app) regularly takes
30+ seconds to open with three teams. I sent over a 45
second video of opening slack and loading each of my teams
in succession. He replied that this was obviously just
networking code being untimely and has nothing to do with
HTML, Node, or Electron. That may or may not be so, but to
an end user like me, it doesn't matter.

<video autoplay loop class="overflowing">
  <source src="/assets/interfaces/slack-45s.mp4" type="video/mp4">
</video>

In 2007, after releasing the iPhone, Steve Jobs told
developers that they could all write apps for the iPhone
_today_, as long as they did it in HTML5. Within a year,
he'd reversed his position after realizing how many
compromises were involved in the web experience.

In 2012, Mark Zuckerberg ignited JavaScript developers
everywhere after announcing that Facebook's biggest mobile
mistake was focusing on HTML5. Meanwhile, consumers
everywhere celebrated as they were given a native app that
was far faster and more responsive.

Everyone one of us knows that when it comes to a
smartphone, we'd use a native app over an in-browser HTML5
any day of the week. Yet when it comes to the desktop, we
still consider web-based apps acceptable. At least when it
comes to users, not necessarily for ourselves.

Bertrand

Bad UI animation examples:

* OS X spaces.
* OS X Lion fullscreen.
* 1Password unlock.
* iOS to home screen.
* iOS app switching.

I find the animation trends on OS X especially terrifying.
In its next version, what if Apple insistent the Cmd + Tab
app switches be animated? Just imagine how much
productivity would evaporate overnight, especially amongst
the power users that use this feature. Aggregated on a
world scale, you could probably measure it 100s of millions
of dollars.

## ThemWare

<video autoplay loop class="overflowing">
  <source src="/assets/interfaces/terminal.mp4" type="video/mp4">
</video>

Developers use terminals.

Here's why I like terminal apps:

* Startup time: zero.
* Screen transition time: zero (no animations or loading
  time).

* Interface elements: limited, but uniform.
* Learing curve: steep, but rewarding.

* Composability: I'm not a zealot singing the praises of
  the Unix Religion, but _most_ terminal apps produce
  output that I can process in some way to get into another
  program. It could be _a lot_ better, but it's leaps and
  bounds over what I can do on the desktop. Much of the
  time, even copying text out of a modern web app can be a
  tricky proposition because of how HTML elements are
  nested.

VIDEO: Copying "you're receiving notifications" text out of
GitHub.

## The Core Tenets of Design

If you ask a web designer about the elements of practical
design in interfaces today (I say _practical_ to
disambiguate from much more vague design tenets like
[Dieter Rams' ten principles of good design][dieter-rams]),
they'd talk to you about text legibility and whitespace.
I'd argue 

Let's dig into it by looking at the aspirational future
interface concept from a great movie: _Minority Report_.

<img src="/assets/interfaces/minority-report.jpg" data-rjs="2" class="overflowing">

I think we can all agree that this an interface that's
incredible and which all of us want, but if we drill into
it, what are its most amazing aspects?

Years ago, I might have said that it was the wafer thin
touch screens, but we have this now! In fact, ours are
arguably better because they have far greater color depth
than anything they seem to have in Philip K. Dick's
dystopian future.

Today, by far the most amazing aspect is that its an
interface that's keeping up to its user. Instead of waiting
on the computer to think about some text completion, show
him an animation because he's switching an app, or start up
a program, it's keeping up with everything he tells it do
_exactly_. Besides a terminal, we don't have any modern UI
that is even close to being able to do this.

A successful interface isn't one that looks good in a still
screenshot, it's one that maximizes our productivity and
lets us _keep moving_. Legibility and whitespace are great,
but they're of vanishing unimportance compared to speed and
responsiveness.

## The Future

Neither a terminal nor today's web apps are what the future
should look like, but the terminal is closer.

Things we can learn from modern web app design:

* Rich media elements: images, videos, tabulated results,
  etc. The terminal has needed an answer to these since
  1985, but still doesn't have one.

* Fonts. Monospace is the best family of fonts for
  programming, but are terrible for reading.

* Whitespace and line-height: This does help to make UI
  elements more distinctive and text more legibile.

Terminals also need a lot of other things before they're
ever going to a plausible interface replacement for most
people. For example, UI elements that aren't built around
ASCII bar characters.

We need a reboot.

We need toolkits that produce interfaces that are fast,
consistent, bug free, and composable _by default_ so that
good interfaces aren't just something produced by the best
developer/designers in the world, but could be reasonably
expected from even junior people in the industry.

[dieter-rams]: https://www.vitsoe.com/us/about/good-design
