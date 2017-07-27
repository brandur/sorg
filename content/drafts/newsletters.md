---
title: "Archaic HTML and Pidgin CSS: Building HTML Newsletters"
published_at: 2017-07-18T02:49:14Z
hook: TODO
---

After a recent trip to Portland, I decided that I wanted to
try and send a newsletter. I don't post on Facebook very
often, and I'm horrible with staying in touch with friends
and family to let them know what I'm doing, and a tasteful
note sent ever so often with stories about what I'm up to
and what I'm thinking would be a pretty good way of helping
with the problem. I've been a subscriber to a few high
quality newsletters for years, and have always enjoyed
reading them.

It's also about the independent web. I'm one of those
holdouts who refuses to move to a centralized publishing
service, or publish content exclusively a social network.
Bloggers like me used to have a powerful distribution
channel in the form of RSS, and the technology still exists
today, but it's been fading for years, with more and more
people moving exclusively towards their favorite social
platform for content discovery. Email is a flawed
technology in many ways, but it's one of the few
communication channels that every connected person in the
world will reliably have, and it fully supports writing
that's more than 140 characters long and complex media sent
via open standards.

This post won't contain anything revelatory, but describes
a few of the building blocks I used to build a modern
newsletter, and some of the surprises along the way.

## The right size of service (#service)

Initially I assumed that the best way to go would be
through one of the many newsletter services like MailChimp
or TinyLetter, but in all the cases I looked at they either
wanted to paste a branded footer on the end of all
messages, have you use a horrible WYSIWYG editor, or both.
I also wanted to archive old copies on the web somewhere,
and it was looking like I was going to have to reinvent a
custom layer on top of whatever service I ended up using
anyway.

I still wanted to use a service, but one that exposed the
right primitives for me. I wanted perfect control over the
visuals, but the last thing I wanted to do was get into
subscription management myself. It was then that I realized
that Mailgun offered an API for mailing lists. I've been
using them for years at Heroku, then at Stripe, and the
service has always been well-designed and reliable. I
experimented by sending a few messages to myself with their
Go SDK, and then it was off to the races.

## Pidgin CSS (#css)

Over the last few decades, we've had pretty good success in
standardizing how HTML and CSS are rendered across
browsers, and during the time some huge victories were won,
including dragging even hopeless stragglers like IE up to
spec. I'd naively believed that the war had been won, but
this whole time there's been a neverending battle raging on
the frontier of email HTML/CSS.

The newsletter industry has developed a form of pidgin CSS
which is made up of the lowest common denominator of what
the world's diverse set of email clients will support.
[This CSS support matrix][email-css] does a good job of
showing just how divergent (and underwhelming) feature
support is between clients. The result is that best
practice in newsletters is to keep HTML email as basic as
possible. Fancy CSS keywords like `float` (`flex` and
`grid` are totally off the table) are best avoided, and
`<table>` is still the state of art when it comes to more
complex layouts.

Email clients are trending towards better support, but it's
going to be a long time before they hit parity with modern
web browsers, and there's a good chance they'll never get
there. Amazingly, clients from companies that we tend to
think of as the most advanced in the world are the most
regressive, like Google Mail.

My experience was that everything beyond the most trivially
basic CSS usually caused problems in at least one mail
client (often Google Mail). For example:

* `<style>` aren't supported by Google Mail (meaning a very
  healthy fraction of all potential readers), so all CSS
  needs to be inlined like `<p style="...">` [1].
* Negative margins don't work.
* Descendant selectors (`#wrapper p`) and child selectors
  (`#wrapper > p`) can't be used.
* Ergonomic niceties like `rem` are out.

After a few false starts I fell back to a layout that
offers a maxima in the tradeoff between simplicity and
style: a single centered column. I wrote CSS in my
templates with normal `<style>` tags in my templates, and
used [Doucer][douceur] to inline it for email.

## The architecture of a newsletter (#architecture)

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

## Passages & Glass (#passages)

My new newsletter is called _Passages & Glass_. It's
intended to be an digest of many topics that interest me
including travel, ideas, products, and software, but a
little more personal than what I'd normally put in a blog
post. If this sounds like something that interests you,
please consider [signing up to receive it][signup]. I
won't bother you often.

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
