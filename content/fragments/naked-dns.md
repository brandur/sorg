---
title: Naked DNS
published_at: 2016-01-31T19:39:33Z
hook: UNWRITTEN. This should not appear on the front page.
---

A recent link from HN re-raised the discussion about
whether it's a good idea to use "naked DNS", which means
hosting a site at a root domain like "brandur.org" without
a "www" subdomain in front of it. The case against it is
that because DNS technically only supports `A` records at
the root level, and `A` records can only be linked to an IP
address (and not a hostname like `CNAME` can), it's not
possible to route traffic with as much flexibility as you
might by using a subdomain. If you're under heavy load or a
large scale attack, you can't dynamically re-route traffic
because they're married to one IP.

This argument was stronger five years ago. Since then, DNS
providers have been inevitably moving in the direction of
providing non-standard options for root domains that
compensate for the downsides of the `A` record. For
example:

* [CloudFlare's "CNAME Flattening"][cloudflare]
* [DNSimple's ALIAS record][dnsimple]
* [Route 53's "alias" record][route53]
* [DNS Made Easy's ANAME record][dnsmadeeasy]

Using any of these will neutralize this argument against
using a naked domain [1]. The disapora of options is
unfortunate, but because DNS didn't provide an adequate
answer and because the demand existed, service providers
adapted by building their own.

[1] Although `A` records tend to be the champion argument
against naked DNS, there are other somewhat compelling
points. Maybe the most effective being that you have the
overhead of sending cookies (which are set on the main
domain) to any subdomains that are serving static assets.
This can be alternatively solved by just serving static
assets from a separate domain, which is likely to a fairly
nominal additional cost if cookie size on static assets is
important to you.

[cloudflare]: https://blog.cloudflare.com/introducing-cname-flattening-rfc-compliant-cnames-at-a-domains-root/
[dnsimple]: https://support.dnsimple.com/articles/alias-record/
[dnsmadeeasy]: http://www.dnsmadeeasy.com/services/anamerecords/
[route53]: http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/resource-record-sets-choosing-alias-non-alias.html
