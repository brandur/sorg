---
title: GPG at Stripe
published_at: 2015-10-18T21:02:24Z
---

One surprise for me when finally getting a close-up look at Stripe's internal
architecture was a pleasant one (I started work there about a month ago): GPG
is deeply integrated into the deployment stack. Every engineer gets a laptop
provisioned with MacGPG, a key generated just for them, and a passphrase that
goes into the system keychain and which becomes available when they log into
the system.

The uses are twofold:

* Certain security-critical repositories require that every commit to Git is
  signed. Releases that are not signed by a whitelisted key won't run.

* Deployment infrastructure. Rather than going into servers and initiating a
  deploy via SSH, a client "asks" a deployment service for a deploy by building
  a small manifest containing the current time and the commit being deployed,
  signing it, and then sending it up. The service verifies the signature's
  authenticity and initiates the deploy.

The system's greatest technical achievement is making the process seemless
enough that most engineers don't have to understand exactly how the
architecture works or have to interact with any GPG tooling; usually a good
thing given that GPG is infamously difficult to use.

Realistically speaking, it's likely that these tools are homegrown solutions
that are unlikely to see wider distribution, but turning over this one stone
makes me wonder how many others out there are hiding other such localized
GPG ecosystems.
