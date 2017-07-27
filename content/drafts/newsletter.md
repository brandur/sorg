---
title: Pidgin CSS, and Building an HTML Newsletter
published_at: 2017-07-18T02:49:14Z
hook: TODO
---

I wanted to try sending a newsletter.

Indepedent web.

Initially I assumed that the best way to go would be
through one of the many newsletter services like
TinyLetter, but in all the cases I looked at they either
wanted to paste a branded footer on the end of all
messages, have you use an ugly WYSIWYG editor, or both. I
also wanted to archive old copies on my own domain, and it
looked like that was going to force me to reinvent a custom
layer of my own on top of their service.

After digging around a little more, I realized that Mailgun
offered a list management service. Mailgun offers mailing
APIs that are geared toward developers, and I've been using
them successfully for years, first at Heroku, then at
Stripe, and immediately knew it was the right primitive for
me.

## Email CSS & HTML (#css)

Over the last few decades, we've had pretty good success in
standardizing how HTML and CSS are rendered across
browsers, and I'd naively believed that this was a problem
that was well behind us. Even if that's true generally, it
certainly isn't when it comes to HTML/CSS in email.

The newsletter industry seems to have developed a form of
pidgin CSS 

Email clients seem to be trending better in how they handle
HTML/CSS, but it's still the wild west out there, and with
some of the world's favorite clients like Google Mail being
the most regressive. [This CSS support matrix][email-css]
does a good job of showing just how divergent feature
support is between mail clients.

General newsletter advice seems to be to keep HTML email as
basic as possible, and to avoid "fancy" CSS keywords like
`float` (`flex` and `grid` are totally off the table). For
more complicated layouts, `<table>` seems to be the state
of the art.

I found that everything beyond the most trivially basic CSS
usually caused problems in at least one mail client (often
Google Mail). For example:

* `<style>` aren't supported by Google Mail (meaning a very
  healthy fraction of all potential readers), so all CSS
  needs to be inlined like `<p style="...">` [1].
* Negative margins don't work.
* Descendant selectors (`#wrapper p`) and child selectors
  (`#wrapper > p`) can't be used.
* Ergonomic niceties like `rem` are out.

After a few false starts I fell back to the simplest and
best layout that I could think of: a centered single
column. I wrote CSS with normal `<style>` tags in my
templates, and used [Doucer][douceur] to inline it.

## Even newsletters have architecture (#architecture)

In the spirit of minimizing the number of components that
I'm running, I reused the code written for this site
([sorg]) to also render my newsletter's HTML and archives.

During development, [fswatch] feeds change events to a
build executable so that I can quickly iterate on content
and design. When I'm satisfied, [another executable][exec]
compiles the final HTML, inlines CSS, and generates a list
message using Mailgun's API. By default it sends a single
initial email to my personal address so that I can vet the
content for a final time, and keep an eye out for any
consistencies that might be email-specific. An additional
`-live` argument sends it for real.

## Subscription management (#subscription-management)

Mailgun helps with list management by automatically
generating person-specific unsubscribe links, but there's
no equivalent for signing up new addresses. You can add new
addresses manually from their control panel, but they won't
host a form for you that allows users to subscribe.

My site is static, so I built a small form that runs as a
[Go executable][passages-signup] and provides this one
function. It's a little painful that I need to run a
separate service to handle this miniscule function, but
hopefully it's micro enough that I won't have to look at it
very often. I chose Go for the job because it's got a
remarkable track record for API stability and minimal
upgrade churn.

## Introducing Passages & Glass (#passages)

My new newsletter is called _Passages & Glass_. It's
intended to be an digest of many topics that interest me
including travel, ideas, products, and software, but a
little more personal than what I'd normally put in a blog
post. If this sounds like something that interests you,
please consider [signing up to receive it][signup]. I
promise not to bother you often!

[1] Somewhat ironically, it might even be fair to call
    Google Mail the new Internet Explorer of email
    rendering.

[douceur]: https://github.com/aymerick/douceur
[email-css]: https://www.campaignmonitor.com/css/
[exec]: https://github.com/brandur/sorg/blob/master/cmd/sorg-passages/main.go
[fswatch]: https://github.com/emcrisostomo/fswatch
[passages-signup]: https://github.com/brandur/passages-signup
[signup]: https://passages-signup.herokuapp.com
[sorg]: https://github.com/brandur/osrg
