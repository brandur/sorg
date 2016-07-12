---
title: AWS Certificate Manager
published_at: 2016-01-21T20:47:11Z
---

Yesterday Amazon shipped their best addition to the AWS toolchain in a long
time in form of the [AWS Certificate Manager][blog]. The service issues TLS/SSL
certificates under Amazon's authority and allows you to easily attach them to a
CloudFront distribution or Elastic Load Balancer. That may not sound
spectacular, but the real win is the price: free with the use of other Amazon
services.

ACM won't change how people use AWS, but it does play very nicely for anyone
using Amazon for other services. Where previously I'd have to use CloudFlare's
SNI proxy to get HTTPS for a static site built on CloudFront and S3, I can now
issue a certificate and have CloudFront serve it directly, taking a big hop out
of the critical path.

Another nice surprise is that ACM will allow you to issue wildcard certificates
(I issued a `*.brandur.org` that can be examined [here][drop] as a sample).
This is tremendously convenient for microservices-style architectures because a
single certificate can be created and then attached to a number of different
nodes. ACM also manages certificate rotation automatically, so everything will
stay online even through the expiration of the original certificate.

One of the downsides brought up on HN is that unlike Let's Encrypt, Amazon
doesn't give users access to certificate private key, meaning that they're not
portable outside of the AWS product infrastructure. It may be somewhat
controversial, but this is the perfect kind of product trade-off for them to
make: it offers a very convenient path for users of existing products while
also preventing any kind of abuse that the service might incur by being overly
generous with its free offerings.

Domain verification works by having Amazon e-mail a domain's public contact and
having them follow back a verifiction link. This system may be a little less
open than some of the [Let's Encrypt ACME challenges][acme] that have been
submitted to the IETF, but it's more convenient and faster to use.

The ACM certificate chain:

![Certificate chain](/assets/fragments/acm/chain.png)

[acme]: https://github.com/ietf-wg-acme/acme
[blog]: https://aws.amazon.com/blogs/aws/new-aws-certificate-manager-deploy-ssltls-based-apps-on-aws/
[drop]: https://drop.brandur.org/certificate.txt
