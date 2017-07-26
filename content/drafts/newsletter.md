---
title: Sending An Email Newsletter
published_at: 2017-07-18T02:49:14Z
hook: TODO
---

I wanted to try sending a newsletter.

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

Email clients seem to be trending better in how they handle
HTML/CSS, but it's still the wild west out there, and with
some of the world's favorite clients like Google Mail being
the most regressive. [This CSS support matrix][email-css]
does a good job of showing just how divergent feature
support is between mail clients.

The general advice out there seems to be to keep HTML email
as basic as possible, and to avoid fancy CSS keywords like
`float` (`flex` and `grid` are totally off the table). For
more complicated layouts, `<table>` is the recommended way
to go!

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
templates, but used [Doucer][douceur] to inline it.

## Backend (#backend)

I reused the code I've already written for this site
([sorg]) to build a slight variant on my normal pages that
was suitable for email.

During development, [fswatch] feeds notifications to a
build executable so that I can quickly iterate on content
an design (these generated pages also act as my list
archive). When I'm satisfied, [another executable][exec]
renders HTML, inlines CSS, and generates an email with
Mailgun's API. By default it sends messages to my personal
address so that I can vet the content directly from a real
mail client and keep an eye out for any consistencies
compared to normal web rendering. An additional `-live`
argument sends to the list.

## List management (#list-management)

Mailgun helps with list management by automatically
generating person-specific unsubscribe links, but there's
no equivalent for signing up new addresses. You can add new
addresses manually from their control panel, but they won't
host a form for you.

My site is static, so I built a small form that runs as a
single Go executable and provides this one function. It
runs on Heroku.

[1] Somewhat ironically, it might even be fair to call
    Google Mail the new Internet Explorer of email
    rendering.

[douceur]: https://github.com/aymerick/douceur
[email-css]: https://www.campaignmonitor.com/css/
[exec]: https://github.com/brandur/sorg/blob/master/cmd/sorg-passages/main.go
[fswatch]: https://github.com/emcrisostomo/fswatch
[sorg]: https://github.com/brandur/osrg
