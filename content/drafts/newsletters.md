---
title: "Pseudo-HTML and Pidgin CSS: Building an Email Newsletter"
published_at: 2017-07-28T17:35:07Z
hook: Finding a backend and building a toolchain for
  sending an email newsletter. Realizations that
  standardization in email HTML and CSS is a distant dream.
---

After a recent trip to Portland, I decided to start writing
a newsletter. I don't post on Facebook very often and am
not very good at staying in touch with friends and family,
so a curated message that I could send every so often
describing what I'm up to seemed like a good way to help
with that problem. I've been a subscriber to a few high
quality newsletters from other people for years, and
reading them has always been something that I've enjoyed.

I also like the idea of supporting the independent web. I'm
one of the holdouts who refuses to move to a centralized
publishing service, or publish content exclusively a social
network. Bloggers used to have a powerful distribution
channel in the form of RSS, and although the technology
still exists today, but it's been fading for years, with
more and more people moving exclusively towards their
favorite social platform for content discovery. Email is a
flawed technology in many ways, but it's one of the few
communication channels that every connected person in the
can reliably be expected to have, and it fully supports
content more than 140 characters long and rich media over
open standards.

This post contains nothing revelatory, but describes a few
of the building blocks I used to build a modern newsletter,
and some of the surprises along the way.

## The right shape of service (#service)

Initially I assumed that the best way to go would be
through one of the many newsletter services like MailChimp
or TinyLetter. Maybe it is, but in all the cases I looked
at they either wanted to paste a branded footer on the end
of all messages, have you use a WYSIWYG editor, or both;
aspects that I'm pretty allergic to. I also wanted to
archive old copies on the web somewhere, and it was looking
like I'd have to reinvent a custom layer on top of whatever
service I ended up using anyway.

I sure wasn't about to start sending email myself, so I
still wanted to use a service, but kept looking for one
that exposed the right primitives for me. I wanted perfect
control over the visuals, but to fully offload subscription
management to someone else.

After a little more Googling I discovered that Mailgun
offered an API for mailing lists. I've been using them for
sending mail for years at Heroku, then at Stripe, and have
always found their service to be well-designed and
reliable. I poked around their control panel for a while
and experimented by sending a few messages to myself with
their Go SDK, and it was off to the races.

## Pidgin CSS (#css)

Over the last few decades, we've had pretty good success in
standardizing how HTML and CSS are rendered across
browsers, and during the time some huge victories were won.
Progressively more sophisticated tests like Acid2 and Acid3
helped drag browsers up to spec and establish widespread
consistency, and even hopeless stragglers like IE would
eventually fall in line. I'd naively believed that this was
a war that had been won, only to realize that this whole
time there's been a battle raging on the frontier of email.

Email clients are a million miles away from rendering
anything that's even remotely compliant with anything, and
they're all uncompliant in their own exotic ways. Some
clients are better than others, and somewhat ironically the
companies that we tend to think of as the most advanced in
the world are some of the most regressive, like Google. If
you threw Acid2 at Google Mail for example, you'd be lucky
to see a lone yellow pixel on screen.

The newsletter industry has dealt with this sad state of
affairs by developing a form of pidgin CSS made up of the
lowest common denominator of what the world's diverse set
of clients will handle. [This CSS support
matrix][email-css] does a good job of showing just how
divergent (and underwhelming) feature support is between
clients. Best practice in newsletters is to keep HTML email
as basic as possible. Fancy CSS keywords like `float` are
best avoided, anything developed this decade like `flex`
and `grid` are totally out, and `<table>` is still the
state of art when it comes to building more complex
layouts.

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

After a few false starts I fell back to a layout that
offers a maximum in the tradeoff between simplicity and
style: a single centered column. I wrote CSS in my
templates with normal `<style>` tags in my templates, and
used [Doucer][douceur] to inline it for email.

## The architecture of a newsletter (#architecture)

In the spirit of minimizing the number of components that
I'm maintaining, I reused the code written for this site
([sorg]) to also render my newsletter's HTML and archives.

During development, [fswatch] feeds change events to a
build executable so that I can quickly iterate on content
and design. When I'm satisfied, [another executable][exec]
compiles the final HTML, inlines CSS, and generates a list
message using Mailgun's API. By default it sends a single
initial email to my personal address so that I can vet the
content for a final time, and keep an eye out for any
rendering problems that might be email-specific. An
additional `-live` argument sends it for real.

## Subscription management (#subscription-management)

Mailgun helps with list management by automatically
generating person-specific unsubscribe links, but there's
no equivalent for signing up new addresses. You can add new
addresses manually from their control panel, but they won't
host a form for you that allows users to subscribe.

My site is static, so I built a small form that runs as a
[Go executable][passages-signup] and provides this one
function. It's a little painful that I need to run a
separate service to handle this uninteresting function, but
it should be micro enough enough that I won't have to look
at it very often. I chose Go for the job because it's got a
remarkable track record for API stability and minimal
upgrade churn. I'm hopeful that in ten years it'll still be
running as expected with minimal maintenance on my part.

## Passages & Glass (#passages)

My newsletter is called _Passages & Glass_. It's a digest
of many topics that interest me including travel, ideas,
products, and software, but with a few more personal
touches compared to what I'd put in a blog post. If this
sounds like something that interests you, consider [signing
up to receive it][signup]. I won't bother you often.

[douceur]: https://github.com/aymerick/douceur
[email-css]: https://www.campaignmonitor.com/css/
[exec]: https://github.com/brandur/sorg/blob/master/cmd/sorg-passages/main.go
[fswatch]: https://github.com/emcrisostomo/fswatch
[passages-signup]: https://github.com/brandur/passages-signup
[signup]: https://passages-signup.herokuapp.com
[sorg]: https://github.com/brandur/osrg
