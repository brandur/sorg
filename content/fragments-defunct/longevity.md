---
title: Optimizing For Longevity
published_at: 2015-12-27T18:12:33Z
---

I've been thinking a lot lately about how to architect the tech stack of a
simple web site like this one in a way that it's most likely to be sustainable
over a long period &mdash; say thirty years. The existential threats over that
kind of time come in all shapes and sizes: code rot that leads to an app that's
very inefficient to maintain, the obsolescence of packages the app depends on
as they fall out of maintenance, the disappearance of core services that run
the app like Heroku or Cloudflare, or even a dramatically reduced time
investment on the part of its creator.

With an eye towards optimizing for longevity, Ruby is one of my top concerns.
Although I love the language, the more of it I write, the more I realize that
every line of Ruby written that's not accompanied by exhaustive test coverage
is an accident waiting to happen. The language itself provides no guarantees
besides the basic checks in the interpreter, and even there you'll need a code
runner like a test suite to get them consistently before production. This isn't
a serious problem for apps that are in a state of constant supervision and
maintenance (i.e. the core components of well-funded companies), but for a
small codebase like this one with only a single maintainer, when trying to fix
a problem down the road I'm just as likely to create a new bug as I an to close
the one I'm looking at.

Some of its other problems are a complex runtime (Ruby may have been simple
enough, but it's highly recommended to bring in augmentations like Rbenv these
days), a complex package management system that relies on multiple
donation-fueled third party services (Rubygems and Bundler API), and an
ecosystem that strongly encouranges the use and proliferation of dependencies
regardless of their long-term viability.

The solution to this type of code rot in a dynamic language is probably to move
to a compiled language; all technology is imperfect, but a compiler provides a
certain baseline guarantee that makes all compiled code a little bit safer. I'd
tend to favor Go these days if only because its compiler tends to be easy to
work with and stay out of the way while still doing its job nicely by providing
orders of magnitude more default safety than the interpreter of a dynamic
language.

Go's controversial dependency management system is also well-suited for
longevity; although vendoring creates a myriad of problems, it is remarkably
resistant to external changes and rot. Even if GitHub was gone in ten years,
you could still build and deploy your app out of a self-hosted Git repository
on any computer in the world.

Removing over-depedence on third party services is also key to operating over a
long time horizon. If Heroku disappeared tomorrow, I could run this app on any
VPS behind a Ruby-standard Nginx and Puma deployment, but would prefer not to
because any box under my control will eventually need care and maintenance of
its own in the form of OS upgrades or package updates as new vulnerabilities
appear. For a simple app like this one, the answer is probably static site
generation given that it's likely that basic HTML, CSS, and JavaScript will be
runnable forever. A Docker configuration may be better-suited for a more
complex app given that Docker images are now runnable by many major providers
which helps to ensure portability, but even there you run the risk of Docker
repositories or base images going offline or becoming stale.

Incredibly, TLS on a custom domain may still be one of the biggest challenges
to long-term uninterrupted operation. [CloudFlare's](/fragments/cloudflare-ssl)
SNI-based product provides a near-perfect solution for this today, but if that
went away we'd be back to dealing with more traditional certificate authorities
and the need for periodic manual intervention as certificates expire and need
to be rotated. That said, the recent fantastic work by Let's Encrypt in making
free certificates widely and easily available is an encouraging sign that the
industry may become more commoditized.

One mental exercise that we often engaged in at Heroku was how to reduce the
number of moving parts in a system. No matter how cost-free you think it will
be when putting it in, each piece of a wider system, whether it be Postgres,
Redis, or a new AWS service, will eventually fail and require that someone take
a look at it. The only sure way to minimize this type of pain is to
uncompromisingly keep the number of these critical dependencies to a minimum,
and although this conservative behavior may bar entrance to the latest and
greatest technologies of today, it tends to pay off for the architect who goes
long.
