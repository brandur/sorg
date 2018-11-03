---
title: Simple and stateless CSRF protection with the `Origin` header
published_at: 2018-11-01T16:05:47Z
hook: Implementing simplified CSRF protection using the
  `Origin` header.
---

I've previously written about how much [I like static
sites](/aws-intrinsic-static). You can't beat them for
simplicity or scalability (during heavy load from an HN day
this site will serve ~20k uniques/day without breaking a
sweat, and all for a few dollars a month), but their
inherent downside is that any kind of dynamic elements
become hard or impossible.

I've been running [an infrequently-sent
newsletter](/newsletter) for about a year now (and with a
new issue on the way, now's a great time to sign up), and
to get perfect control over the layout decided to build it
on [a custom stack](/newsletters). By extension, I needed
to have a dynamic backend somewhere that could receive a
new email address and add it to the list, something my
static site obviously couldn't provide.

I worked around the problem by building a separate Go app
hosted on Heroku and linking prospective users from here to
there. It worked, but the user experience left a lot to be
desired because getting signed up was a two step process
(1. get linked to the signup app, and 2. submit the signup
form) with the second step probably involving a lengthy
wait for a free tier Heroku app to dethaw.

## CSRF protection (#csrf)

If you're wondering why I couldn't just render a form on
the static site and have it submit to the dynamic signup
app -- well, it hasn't historically been possible to do
this _safely_. Every app should protect itself against
cross-site submissions to prevent abuse. The traditional
way of doing this is to generate a CSRF protection token
and put in a user's cookie and also embed it as a hidden
field in the form to be submitted. When receiving the
submitted form, the token in the cookie is compared to the
token in the form, and the contents rejected unless they
match. The unfortunate side effect is that you need a
dynamic app to _render_ the form in addition to receiving
its payload.

## The `Origin` header (#origin)

Fortunately, the addition of the [`Origin`][origin] header
gave us a new option. `Origin` is a little like the classic
`Referer` header except that it contains strictly less
information to reduce the amount of user information being
exposed to the target site. It still contains an origin
domain, but path information is stripped. That means that
even conscientious web browsers can send the header
liberally without worrying about leaking user data.

Applications can take advantage of this to implement
simplified CSRF protection that checks the value of
`Origin` against a known whitelist instead of using a token
and cookie. An attacker can't prevent a web browser from
sending a correct `Origin` value [1], so a target app can
rely on its validity.

That's what I ended up doing for my newsletter. I wrote a
[Go CSRF protection middleware][go] that relies only on
`Origin` and started hosting a form directly on my static
site's [newsletter page](/newsletter). The signup app
whitelists the static site's URL of `https://brandur.org`
along with its own URL
`https://passages-signup.herokuapp.com`, and will happily
allow form submissions from either origin.

!fig src="/assets/fragments/origin/submit.svg" caption="Submitting a form from a static site to a dynamic signup app."

I also "solved" (with a hammer) the app dethaw wait by
having the static newsletter page try to load a tiny image
from the dynamic app that it submits its form to, giving
the app a chance to spin up while the user is entering an
email address [2].

So anyway, in practically every case you should just use
whatever CSRF protection the framework you're using offers,
but in case you need a simplified setup for some reason,
this is a useful trick.

## For the pedantic (#pedantic)

Lastly I'll note that token-based CSRF protection is still
the [overall recommended approach][csrf] to mitigating
CSRF. Although `Origin` is more or less universally
available in browsers now (XX% according to
[caniuse.com][caniuse]), Edge is known to not send `Origin`
under certain conditions, and there's a chance that more of
these exceptions might appear in the future.

For the time being, those cases are rare enough that
`Origin` approach should be fine, especially if you also
configure your protection middleware to allow an empty
`Origin` value (again, an attacker doesn't have a way to
spoof that). However, token-based protection will provide
the best possible compatibility.

[caniuse]: https://TODO
[csrf]: https://TODO
[go]: https://github.com/brandur/csrf
[origin]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Origin

[1] An attacker can of course change an `Origin` value with
perfect control over a victim's web browser, but at that
point all bets are off, and traditional CSRF protection
mechanisms won't protect them either.

[2] At the cost of some false positives for users who land
on the page and decide not to sign up. Luckily, dethawing a
small Go app is not an overly expensive operation.
