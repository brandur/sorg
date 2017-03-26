---
title: GitLab, and Centrally Hosted Platforms
published_at: 2017-03-26T06:35:28Z
hook: Thoughts on depending on the wrong primitive (Docker)
  and entering a very competitive space.
---

I attended the release party for GitLab 9.0 last week here
in San Francisco. It wasn't oversubscribed, so I got a
chance to speak to their CEO a bit about GitLab's strategy.
Apparently though, it's a topic so readily available they
they've [actually published on it on their site][strategy].

One prospect that I was especially interested in was the
possibility of an end-to-end development to production
pipeline. They obviously host Git already, but have also
gone so far as to provide CI services for repositories, and
then went even further by providing [review
apps][review-apps] [1] that are deployed automatically
based on Git branches that are pushed to GitLab's servers.

What if they took it one step beyond by hosting production
apps? Developers could create an account, push code, and
arrange to have it go live at a `*.gitlab.com` URL from the
same interface that they're using to merge pull requests.
GitLab does the heavy lifting to make sure that deployments
are continuous, and safe.

Between AWS, GCE, and Azure, the world now has more
container services than you can shake a stick at, but
they're all so low level that every serious user ends up
either building their own infrastructure management and
deployment tooling from scratch, or maintains an open
source installation of something like Kubernetes at great
personal expense.

There's still room for a platform that's _not_ self-hosted
and still meets 90% of the needs of most users. Even after
years of only modest advacements, Heroku is still relevant
in this space even though the IaaS goliaths should have
surpassed it years ago. They still make getting from zero
to production easy, and guarantee a low level of
maintenance overhead once you're there. Docker-based
services that ask you spin up and look after your own nodes
for them to run on aren't even close.

GitLab's potential opportunity here is to own the whole
pipeline from code to deployment, a feat totally
unparalleled in the world of developer tooling. By baking
in metrics, error tracking, and best practices like canary
deployments, they could zero in on a level of
sophistication that most users would never have built
themselves. If done right, it could be akin to Apple's
advantage in providing a far superior personal computing
experience thanks to having perfect control over their
entire stack.

All that said, users right now are expected to bring their
own platform, and app deployment isn't a priority for the
company, which is too bad. Competing for business against
leviathans like GitHub and Travis isn't a place that I'd
want to be in personally, but as demonstrated by 9.0's
feature list, the people at GitLab seem to have a knack for
moving quickly on new features despite having a very mature
codebase. Hopefully, it'll keep them alive.

[1] Identical to the Heroku feature of the same name.

[review-apps]: https://docs.gitlab.com/ce/ci/review_apps/
[strategy]: https://about.gitlab.com/strategy/
