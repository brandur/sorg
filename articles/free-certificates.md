---
hook: Getting a certificate that most browsers would accept used to be difficult and
  expensive, but that's changing as we enter a relative golden age of web encryption.
  Read this for options for getting certificates issued for free.
location: Vancouver
published_at: 2016-01-31T17:48:53Z
title: A Guide to Free CA-Signed Certificates
---

Securing a website used to be an expensive process. Although certificates have
been slowly getting cheaper, they've still on par with the cost of the domain
name that they're protecting, and getting one issued was often complex and
error prone. Furthermore, in a pre-[SNI][sni] world, HTTPS connections needed
to be terminated at a unique IP address, making it prohibitively expensive for
hosts to offer low cost encryption to their users.

In an attempt to unwind some of the mistakes that were made around security in
the earlier ages of the Internet, browsers are starting to prod service providers
in the right direction. For example, in the near future [Chrome will start
shaming websites that aren't encrypted][chrome-shame] and [Firefox will start
red flagging login forms that come in over HTTP][firefox-shame].

The good news is that we're now living in a golden age of secure connections.
The price of CA-signed certificates is trending toward zero, and if you're a
savvy user who knows where to look, you can easily get one for free already
today. Support for SNI is now widespread enough that hosts have a cheap
mechanism for offering secure termination for all their users. Encryption may
be especially critical for banks and Facebook, but it belongs on every site
online: shopping sites (even pre-checkout), blogs, marketing landing pages,
personal websites, and everything in between. Hopefully by reading this guide,
you'll realize that there aren't any excuses for running an insecure website
anymore, so come on, let's encrypt!

## Services

### CloudFlare

**Website:** [https://www.cloudflare.com/][cloudflare]

Although CloudFlare is largely known for being a CDN, they've been more quietly
offering a great certificate-issuing and TLS terminating service for some time
now. It's easy to use, and is especially ideal for anyone who's hosting content
on another service that already offers secure termination (like Heroku or
GitHub pages), but who would like to have a custom domain name. You also get
the added benefit of CloudFlare's CDN services, which can be had for free.

The good:

* Unbelievably easy. Especially if you're already hosting your DNS with them,
  getting a secured endpoint involves as little as creating a new record,
  specifying a target origin, and clicking the little cloud icon to turn it on.
* Automatic rotation. You don't even have to know or understand what's
  happening, but can restly safely reassured that your users will have
  continued access to your services without suddenly getting hit with a red
  expired certificate page because one of your ops people forgot to get a new
  one issued.

The bad:

* Certificates are local to CloudFlare, and you can't export them and bring
  them with you. Using CloudFlare's CDN services is appropriate to just about
  everyone though, so that's fine for a lot of different cases, but 
* CloudFlare still defaults to "flexible mode" that allows a target origin
  server to be serving content over HTTP even if CloudFlare itself is
  terminating over HTTPS. This option is provided for user convenience (and it
  should be noted that it's still better than no TLS), but allows unwary users
  to unwittingly build themselves an unsafe setup.

I should also note for the pundits that CloudFlare's magic works by by SNI, and
as such may not work for clients that are using absolutely ancient technology
for browsing. As of today, ["Can I Use ...?" estimates support at 97+%
globally][caniuse], so an SNI-based solution is probably appropriate for you as
long as you're running an operation that's smaller than Google.

### Let's Encrypt

**Website:** [https://letsencrypt.org/][lets-encrypt]

Let's Encrypt is free CA run by the ISRG (Internet Security Research Group)
with the charter of providing free certificates in an open and transparent way
to help secure the Internet. They're been making waves lately, and the turning
point that we're seeing around the cost of CA-signed certificates on the
Internet could reasonably be attributed to their work.

The good:

* Let's Encrypt is by far the most flexible of any of these solutions in that
  they'll issue a certificate with private key and all. That means you can take
  these certificates with you and use them with any other service of your
  choice.
* They've built out a great set of tools that allow you to easily get a
  certificate safely installed for common web servers like Nginx or Apache.
* They've been working hard on building a standardized protocol to verify
  domain ownership and issue certificates called [ACME][acme] which will help
  further commoditize Internet security, and curb user error during the issuing
  process.
* Let's Encrypt is a project built in collaboration with the Linux Foundation,
  and they don't have any hidden agendas (or relative to any of these other
  services as least). You can feel good about yourself for using the service.

The bad:

* No wildcard certificates for the foreseeable future. That said, a great API
  that allows for easy automation goes a long way towards compensating for
  this.

### AWS Certificate Manager

**Website:** [https://aws.amazon.com/certificate-manager/][acm]

A brand new entrant is AWS Certificate Manager (ACM), which finally gives us
the missing link for building secure services on Amazon. ACM is AWS-only, but
is easy to use through either their API or web console, and plugs right into a
CloudFront distribution or ELB (Elastic Load Balancer).

The good:

* ACM will issue wildcard certificates for free, a feature which isn't
  currently available from any other provider. This isn't a profound difference
  given that certificates are free anyway, but it saves you from having to
  re-issue a certificate for every new domain you deploy. It's perfect if you
  have a microservices-like setup hosted on AWS.
* Like with CloudFlare, you automatic certificate rotation. This kind of peace
  of mind is worth paying for, but you'll get it for free.
* ACM finally gives us a free (or at least low-cost) way of protecting
  statically built websites served out of S3 buckets. Just create a bucket, a
  certificate in ACM, and a CloudFront distribution, link them all together,
  and you're done.

The bad:

* Like CloudFlare, certificates are local to Amazon and can't be exported. But
  if you're on Amazon, there'a a pretty good chance that you're _all_ on
  Amazon, and this won't be a huge problem.

### StartSSL

**Website:** [https://www.startssl.com/][startssl]

Event though StartSSL is probably not what most people want to use to get
certificates created these days, I'm still going to give them an honorable
mention because they were _the original_ free issuer, and really helped to get
the ball rolling towards a better future.

The good:

* This operation has been around for years and they know what they're doing.
  Being a fully fledged certificate authority, you'll get an easy upgrade path
  if you need something a little heavier like an EV cert.

The bad:

* Getting certificates issued is a totally manual process and someone will need
  to walk through it periodically to make sure your services stay online. This
  used to be an unavoidable reality, but we've got much better options these
  days.

The ugly:

* StartCom uses a client certificate system to log into their control panel and
  the certificate issuing flow is long and fairly obtuse. Only more advanced
  users will be able to understand what's going on here and get a certificate
  issued safely.

## Summary

There's a lot of information above, so here's a simple heuristic that should do
the trick for most people:

* If you're hosted on Amazon, you should use [ACM][acm].
* If you're hosted on another service that gives you some kind of secure
  terminate (like Heroku or GitHub Pages), you should use
  [CloudFlare][cloudflare].
* Otherwise, you should use [Let's Encrypt][lets-encrypt].

For example, this site runs on Heroku. I have my domain terminated by
CloudFront at "https://brandur.org", and CloudFront securely transports content
from my HTTPS Heroku address at "https://brandur-org-next.herokuapp.com".

That's it! Now please go out and secure your web properties.

[acm]: https://aws.amazon.com/certificate-manager/
[acme]: https://github.com/ietf-wg-acme/acme/blob/master/draft-ietf-acme-acme.md
[caniuse]: http://caniuse.com/#feat=sni
[chrome-shame]: https://motherboard.vice.com/read/google-will-soon-shame-all-websites-that-are-unencrypted-chrome-https
[cloudflare]: https://www.cloudflare.com/
[firefox-shame]: https://hacks.mozilla.org/2016/01/login-forms-over-https-please/
[lets-encrypt]: (https://letsencrypt.org/)
[sni]: https://en.wikipedia.org/wiki/Server_Name_Indication
[startssl]: https://www.startssl.com/
