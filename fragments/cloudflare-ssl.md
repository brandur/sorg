---
title: CloudFlare Universal SSL
published_at: 2014-11-12T04:56:26Z
---

I recently activated [CloudFlare's free Universal SSL][cloudflare] for a few of
my apps (including this one), and I'm impressed. The product works exactly as
advertised in that it gets you instant SSL with minimal configuration and no
cert wrangling.

Although I was finally in a position where I could do a fair job managing my
own SSL configuration using free certs signed by StartCom, I'd still
occasionally forget to renew certs and that would result in embarassing
downtime. No more!

Universal SSL combined with a Heroku app turns out to be a perfect match.
CloudFlare can terminate HTTPS requests made to a custom domain and proxy them
through to an app's `*.herokuapp.com` address (which also responds on HTTPS),
keeping both jumps fully secure. CloudFlare calls this mode "Full SSL". "Full
SSL (Strict)" doesn't appear to be suitable because it requires that the target
endpoint both support SNI and respond to requests for the custom domain.

![CloudFlare SSL options](/assets/fragments/cloudflare-ssl/ssl-options.png)

This may be one of the biggest steps taken to bring SSL to the masses in the
history of the Internet. The major problem with SSL is that it just never got
easier; every step from getting a certificate CA-signed to configuring it in
Nginx is confusing and error prone to the extreme, and a natural barrier to
widespread adoption. CloudFlare solves the problem by making the process
completely opaque, which under the circumstances is the right thing to do.

[cloudflare]: http://blog.cloudflare.com/introducing-universal-ssl/
