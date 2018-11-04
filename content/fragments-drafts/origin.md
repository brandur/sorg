---
title: Simple and stateless CSRF protection with the `Origin` header
published_at: 2018-11-01T16:05:47Z
hook: Implementing simplified CSRF protection using the
  `Origin` header.
---

I've previously written about how much [I like static
sites](/aws-intrinsic-static). You can't beat them for
simplicity or scalability (during relatively rare peak load
times this site will serve ~20k uniques/day without
breaking a sweat, and all for a few dollars a month), but
their inherent downside is that any kind of dynamic
elements become hard or impossible.

I've been running what is probably the world's [most
infrequently-sent newsletter](/newsletter) for about a year
now [1], and to get perfect control over email layout
decided (most likely unwisely) to build it on [a custom
tech stack](/newsletters). Receiving new signups and adding
them to the list requires dynamic communication with my
mail provider, and my static site obviously wasn't able to
provide that.

So I ended up building a separate Go app and linking
prospective users from here to there to sign up. It worked,
but the user experience left a lot to be desired because
getting signed up was a multi-step process (link to the
newsletter page, get linked to the signup app, then submit
the signup form). Because the app was hosted on Heroku's
free tier, the second step also usually involved a lengthy
wait for it to come out of hibernation.

## CSRF protection (#csrf)

If you're wondering why I couldn't just render a form on
the static site and have it submit to the dynamic signup
app -- well, it hasn't historically been possible to do
this _safely_.

Every application should protect itself against cross-site
submissions to prevent [CSRF attacks][csrf]. The
traditional way of doing so is to generate a CSRF
protection token and put in a user's cookie and also embed
it as a hidden field in the form to be submitted. When
receiving the submitted form, the token in the cookie is
compared to the token in the form, and the contents
rejected unless they match (or the form field is empty).
Because the token needs to be included in the form to be
submitted, this means that you need a dynamic app to
_render_ the form in addition to receiving its payload.

## The `Origin` header (#origin)

Fortunately, the addition of the [`Origin`][origin] header
gave us a new option. `Origin` is a little like the classic
`Referer` header except that it contains strictly less
information to reduce the amount of user information being
exposed to a destination site. It still contains an origin
domain, but the path is stripped. That means that even
conscientious web browsers can send the header liberally
without worrying about leaking as much browsing
information.

Applications can take advantage of this to implement
simplified CSRF protection that checks the value of
`Origin` against a known whitelist instead of using a token
and cookie. `Origin` is a [forbidden header][forbidden]
which means that it can't be altered programmatically
through JavaScript or the like, and therefore an attacker
can't prevent it from being sent or modify its value. Apps
receiving it can rely on its validity.

That's what I ended up doing for my newsletter. I wrote a
[Go CSRF protection middleware][go] that relies only on
`Origin` and started hosting a form directly on my static
site's [newsletter page](/newsletter). The signup app
whitelists the static site's URL `https://brandur.org`
along with its own URL
`https://passages-signup.herokuapp.com`, and will happily
allow form submissions from either origin.

!fig src="/assets/fragments/origin/submit.svg" caption="Submitting a form from a static site to a dynamic signup app."

I also "solved" (with a hammer) the app unidle wait by
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
the recommended approach to mitigating CSRF [according to
OWASP][owasp] (the Open Web Application Security Project).
An `Origin`-based solution has previously been recommended
as primary defense, but was moved to a "defense in depth"
recommendation after it came to light that certain browsers
don't include the header in some situations. For example,
IE11 when issuing a CORS request across sites in a trusted
zone.

For the time being, those cases are rare enough that
`Origin` approach should be fine, especially if you also
configure your protection middleware to allow an empty
`Origin` value (again, an attacker doesn't have a way to
spoof that). However, as noted by OWASP, token-based
protection will provide the best possible compatibility.

[1] And with a new issue on the way, now is a great time to
[sign up](https://brandur.org/newsletter).

[2] At the cost of some false positives for users who land
on the page and decide not to sign up. Luckily, unidling a
small Go app is not an overly expensive operation.

[csrf]: https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
[forbidden]: https://developer.mozilla.org/en-US/docs/Glossary/Forbidden_header_name
[go]: https://github.com/brandur/csrf
[origin]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Origin
[owasp]: https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)_Prevention_Cheat_Sheet#Defense_In_Depth_Techniques
