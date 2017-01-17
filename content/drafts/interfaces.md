---
hook: We may have sacrificed loading time, responsiveness,
  and ability to be productive in general, but at least we've
  got pretty animations and lots of whitespace.
location: Calgary
published_at: 2017-01-07T00:21:51Z
title: The Future of User Interfaces
---

Modern apps and interfaces frustrate me.

Somewhere around the 00s we made the decision to jump ship
off desktop apps and start writing the lion's share of new
software for the web. This was largely for pragmatic
reasons: the infrastructure to talk to a remote server
became possible for the first time, good cross platform UI
frameworks had always been elusive beasts [1], and desktop
development frameworks were intimidating compared to more
approachable languages like Perl and PHP.

The other reason was aesthetic: HTML and CSS gave web
developers perfect cosmetic control over what their
interfaces looked like, allowing them to custom brand them
and build experiences that were pixel-perfect according to
what they wanted their users to see. This seemed like a big
improvement over more limiting desktop development, but led
to the world we have today where every interface is a
different size and shape, and any conventions that used to
exist (and which were a boon to usability) have disappeared
to the wind.

Today, web apps are still being hailed as the future. With
the possible exception of mobile, most software companies
are building their products for the web, and even when
they're not, web technology is considered a reasonable
alternative for the desktop. Vocal group proclaim that
Electron-based apps convey huge benefits compared to
traditional options.

I'm not on a mission to destroy the web, and will admit
more than willingly that the lower barrier to publishing
HTML has done a huge amount of good for the world. But as
the web is continually augmented with ever more unwieldy
retrofits, there's a widening disparity between what we can
build with it compared to the best-written native apps.
Software we build today takes too long to load, depends too
heavily on slow networks, overemphasizes visual gimmickry,
and lacks the refinement that allows mastery by more
experienced users to make huge strides in productivity.

## The Worst Kept Secret (#worst-kept-secret)

An ex-colleague who is a major JavaScript pundit recently
called me out on Twitter for gross hyperbole after I
claimed that Slack (an Electron-based app) regularly takes
30+ seconds to open with three teams. I sent over a video
of opening slack and loading each of my teams in
succession. It was 45 seconds long. He replied that this
was obviously just networking code being untimely and has
nothing to do with HTML, Node, or Electron. That may or may
not be so, but to an end user like me, it doesn't matter.

<figure>
  <p>
    <video autoplay loop class="overflowing">
      <source src="/assets/interfaces/slack-45s.mp4" type="video/mp4">
    </video>
  </p>
  <figcaption>This is a video of me waiting for Slack
    configured with three teams to fully load. It's 45
    seconds long.</figcaption>
</figure>

In 2007, after releasing the iPhone, Steve Jobs told
developers that they could all write apps for the iPhone
_today_, as long as they did it in HTML5. To his credit
though, he reversed his position inside a year after
realizing how compromised the web experience was compared
to native options.

In 2012, Mark Zuckerberg ignited JavaScript developers
everywhere after announcing that Facebook's biggest mobile
mistake was focusing on HTML5. Meanwhile, consumers
everywhere celebrated as they were given a native app that
was far faster and more responsive.

Everyone one of us knows that when it comes to a
smartphone, we'd use a native app over an in-browser HTML5
any day of the week. Yet when it comes to the desktop,
we're still using Gmail, Reddit, Trello, and JIRA.
Computers and networks tend to be fast enough that this
software is "good enough". Tellingly though, we tend to
avoid this software whenever better options are available,
like with our terminals and text editors.

Bertrand

## It's Not Just Technology (#not-just-technology)

Bad UI animation examples:

* OS X spaces.
* OS X Lion fullscreen.
* 1Password unlock.
* iOS to home screen (from lock screen or another app).
* Switching between iOS apps with double home button click.
* Switching tabs in iOS' mobile Safari.

<figure>
  <p>
    <video autoplay loop class="overflowing">
      <source src="/assets/interfaces/1password.mp4" type="video/mp4">
    </video>
  </p>
  <figcaption>1Password's unlock animation. The stuttering
    isn't a problem with the video on this page; it's
    actually how the animation looks.</figcaption>
</figure>

<figure>
  <p>
    <video autoplay loop class="overflowing">
      <source src="/assets/interfaces/spaces.mp4" type="video/mp4">
    </video>
  </p>
  <figcaption>OS X Spaces, introduced in Leopard. A
    nominally useful feature, but the mandatory animations
    make them slow and unwieldy.</figcaption>
</figure>

I liked every one of these the first time I saw them. The
next five thousand times were less impressive.

I find the animation trends on OS X terrifying. What if in
the next version of OS X apple makes Cmd + Tab task
switching animated? Multiply the new animation's length by
the average number of task switches by the number of users
of the feature by the average value of their time, and
you'll probably find that 10s or 100s of millions of
dollars a year in productivity has evaporated overnight.

## ThemWare (#themware)

<figure>
  <p>
    <video autoplay loop class="overflowing">
      <source src="/assets/interfaces/terminal.mp4" type="video/mp4">
    </video>
  </p>
  <figcaption>Contrary to any "modern" interfaces, a
    terminal is fast and responsive. There are no animations
    or other superfluous visual baggage.</figcaption>
</figure>

Developers use terminals.

Here's why I like terminal apps:

* Startup time: zero.

* Screen transition time: zero (no animations or loading
  time).

* Interface elements: limited, but uniform.

* Learing curve: steep, but rewarding. Another way of
  putting this is that they're optimized for the
  experienced user rather than the first timer. Given that
  successfully onboarded users may spend tens of thousands
  of hours in the UI over the course of their lifetimes,
  this is just good sense.

* Composability: I'm not a zealot singing the praises of
  the Unix Religion, but _most_ terminal apps produce
  output that I can process in some way to get into another
  program. It could be _a lot_ better, but it's leaps and
  bounds over what I can do on the desktop. Much of the
  time, even copying text out of a modern web app can be a
  tricky proposition because of how HTML elements are
  nested.

<figure>
  <p>
    <video autoplay loop class="overflowing">
      <source src="/assets/interfaces/uncomposable.mp4" type="video/mp4">
    </video>
  </p>
  <figcaption>Modern UIs have next to zero composability.
    Even copying text can be a tricky proposition.
    </figcaption>
</figure>

## The Core Principles of Interface Design (#core-principles)

If you ask a web designer about the elements of practical
design in interfaces today (I say _practical_ to
disambiguate from much more vague design tenets like
[Dieter Rams' ten principles of good design][dieter-rams]),
they'd talk to you about text legibility and whitespace.
I'd argue 

Let's dig into it by looking at the aspirational future
interface concept from a great movie: _Minority Report_.

<figure>
  <p>
    <img src="/assets/interfaces/minority-report.jpg" data-rjs="2" class="overflowing">
  </p>
  <figcaption>A futuristic and unrealistic concept interface:
    the computer waits on the human instead of the human
    waiting on the computer.</figcaption>
</figure>

I think we can all agree that this an interface that's
incredible and which all of us want, but if we drill into
it, what are its most amazing aspects?

Years ago, I might have said that it was the wafer thin
touch screens, but we have this now! In fact, ours are
arguably better because ours can display more than two
colors; superior to anything they seem to have in Philip K.
Dick's dystopian future.

Today, by far the most amazing aspect is that it's an
interface that's keeping up to its user. Instead of waiting
on the computer to think about some text completion, show
him an animation because he's switching an app, or start up
a program, it's keeping up with everything he tells it do
_exactly_. The computer waits on the human rather than the
other way around. Besides a terminal, we don't have any
modern UI that is even close to being able to do this.

A successful interface isn't one that looks good in a still
screenshot, it's one that maximizes our productivity and
lets us _keep moving_. Legibility and whitespace are great,
but they're of vanishing unimportance compared to speed and
responsiveness.

## The Road Ahead (#the-road-ahead)

Neither a terminal nor today's web apps are what the future
should look like, but the terminal is closer.

The problem with terminals is that _they suck_. Although
still better than the alternative in so many ways, they've
failed to keep up with any advancements from the last
thirty odd years.

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

APIs

[dieter-rams]: https://www.vitsoe.com/us/about/good-design

[1] Fans of Qt (and maybe others) will vehemently disagree
    that there's never been a good cross platform UI
    library. I'd argue that SDKs like Qt were never quite
    accessible enough and never produced good enough
    results to be suitable for universal adoption.
