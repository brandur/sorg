+++
hook = "On getting features out the door with a small team and sharp toolset."
published_at = 2023-04-09T14:47:23-07:00
title = "Lean = fast"
+++

A feature notably missing from Bridge for too long was multi-factor authentication, but as of last week, it's finally available. The reason it took so long was largely a historical accident -- a decision was made early on to use Keycloak for identity management, and [for a variety of reasons](/atoms/gkoxmy2) this tended to make account-related features slow to ship. Multi-factor was one of them, and we recommended use of Azure/Google SSO as a security stopgap until we had it.

Keycloak was fully decommissioned last month, and we started on MFA work almost right away. With a couple people working on it part-time we got it done in a little over four weeks:

* **Week 1:** I start writing the APIs for multi-factor activation and recovery codes. We will support both TOTP and WebAuthn out of the gate.

* **Week 2:** MFA APIs ship, a colleague on the Frontend teams starts integrating them. I start working on the API interaction for a multi-factor challenge/response.

* **Week 3:** Multi-factor challenge/response ships, the UI for MFA activation is mostly finished, Frontend integration starts for MFA challenge/response. I write internal docs.

* **Week 4-5:** Frontend for MFA ships. Internal users dogfood it or QA. Bugs are fixed. Bolts are tightened to refine what was built. External docs are written.

~4.5 weeks total for a very complete implementation, on a mature product, with hundreds of new test cases written to avoid regression, and both us doing work on other projects in parallel. A very favorable timeline in my view.

I know some readers are thinking "5 weeks? big deal lol", but especially for a cross-team project on a product with many current users that needs to be mindful to ensure stability, it _is_ a big deal. Even back at Heroku where we were relatively agile, getting the first pass of TOTP and recovery codes in was a multi-month process. Yubikeys came much later (and it'd be years before U2F or WebAuthn existed) as a separate project that would've been another few months. Our security team insisted on SMS-based 2FA [1], which involved its own team, a new microservice, and many months more. I was never involved in MFA at Stripe, but if it was anything like any other project it involved 20+ people, 500,000 hours of CI time, and the better part of a year.

## The productivity cliff (#productivity-cliff)

We're certainly benefiting from technological progress -- WebAuthn is now quite mature, has a straightforward browser API, and is simpler to integrate than its predecessors. But that's only a part of the story. I give heavy due to:

* **Small teams _ship_.** Every IC in our org has a huge amount of personal discretion and is empowered to make decisions without sending up the chain for approval. We write code individually,  communicate over Slack and GitHub in moderation, and pair when we need a higher-bandwidth channel to do something hard.

* **Tech stack _does_ matter.** Languages like Go compile and run quickly, making edit-compile-test loops lightning fast. Types mean I find most problems without leaving the editor. Our project is larger now so its total test run is now measured in the 10-20 second range instead of just seconds, but no cloud CI loop is ever required (we have one of course, but you don't depend on it during dev iterations).

* **There's a right size of services.** We're using a "macroservices" model with a frontend, API layer, and backend state machine, and unless something changes dramatically, those are all the services we ever plan to run. The frontend talks to the API by way of strongly-typed TypeScript bindings, making integration faster and more accurate, and we've optimized our development flows such that every frontend pull request spins up a review app that's fully functional against a staging API, and is a single click to log into. Everything can be boostrapped to run locally in less than a minute by following quickstart instructions in READMEs.

Our industry still has a lot to learn when it comes to keeping development productivity high when operating at scale. The all-too-common pitfall is that a stack gets to a certain maturity with a certain number of users and falls off a productivity cliff as making a substantial change gets really hard due to slow loops, brittle tooling, and overwhelming complexity. I can't be sure that we're not going to fall off in the same way eventually, but we're doing everything we can in the meantime to avoid it -- fast edit-compile-run loops, comprehensive regression tests (but again, fast ones!), and sharp tooling end to end.

[1] For reasons I still have trouble understanding. By that time, SMS was already well understood to be an inferior 2FA method due to the risk of SIM-jacking and unreliability of delivery, and it's expensive to run.
