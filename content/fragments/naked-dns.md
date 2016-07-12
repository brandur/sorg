---
title: Naked DNS
published_at: 2016-01-31T19:39:33Z
---

A recent link from HN re-raised the discussion about whether it's a good idea
to use "naked DNS" (i.e. a root domain like "brandur.org" without a "www"
subdomain). The case against it is that because DNS technically only supports
`A` records at the root level, and `A` records can only be linked to an IP
address, it's not possible to route traffic as flexibly as you might with a
`CNAME`. If the site is experiencing a large scale attack, a service operator
can't dynamically re-route traffic because they're married to a particular IP
address.

This argument was stronger five years ago. Since then, DNS providers have been
inevitably moving in the direction of providing non-standard options for root
domains that compensate for the downsides of the `A` record. For example:

* [CloudFlare's "CNAME Flattening"][cloudflare]
* [DNSimple's ALIAS record][dnsimple]
* [Route 53's "alias" record][route53]
* [DNS Made Easy's ANAME record][dnsmadeeasy]

Using any of these will neutralize this argument against using a naked domain
[1]. The disapora of options is unfortunate, but because DNS didn't provide an
adequate answer and because the demand existed, service providers complied by
building custom implementations.

[1] Although `A` records tend to be the champion argument against naked DNS,
there are other somewhat compelling points. Maybe the most effective being that
you have the overhead of sending cookies (which are set on the main domain) to
any subdomains that are serving static assets. This can be alternatively solved
by just serving static assets from a separate domain, which is likely to a
fairly nominal additional cost if cookie size on static assets is important to
you.

[cloudflare]: https://blog.cloudflare.com/introducing-cname-flattening-rfc-compliant-cnames-at-a-domains-root/
[dnsimple]: https://support.dnsimple.com/articles/alias-record/
[dnsmadeeasy]: http://www.dnsmadeeasy.com/services/anamerecords/
[route53]: http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/resource-record-sets-choosing-alias-non-alias.html
