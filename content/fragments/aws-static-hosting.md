---
title: AWS Static Hosting
published_at: 2016-01-04T09:18:18Z
---

**Addendum &mdash;** I've published a [second post][second] containing a
workaround for the problems described in this document.

While working on a tiny project today to demonstrate publishing a static
website through to S3, I was surprised to find that despite almost nonstop
improvements since its release, there are still some key features missing from
the AWS static hosting offering.

A common technique for static website builders like Hugo to generate "pretty"
URIs (i.e. `/about` instead of `/about.html`) is to generate a folder
(`/about`), put a default file in it (`index.html`), and rely on a frontend web
server to detect and serve it implicitly without incoming users being any the
wiser. Amazon was obviously aware of this, and their static website hosting for
S3 allows you to specify an "index document" which will be served when a
directory is accessed.

So everything's fine right? Well, no. The problems begin with [S3 not
supporting HTTPS access on its static website endpoints][s3-endpoints] [1]; a
decision that makes things awkward for would-be customers, but presumably one
that was made with difficulty for reasons related to limitations in internal
architecture.

Luckily, Amazon gives you a way to serve files in S3 over HTTPS through the use
of CloudFront, their global CDN service that can be easily linked to a bucket.
So that gets us easy, fast, and secure static hosting right? Well, almost, but
no. The problem now is that Cloudfront doesn't support index documents like S3
does. It allows you to specify a "default root object" that will forward
requests to `example.com` along to `example.com/index.html`, but [this applies
only at the top level][cloudfront-root]; subdirectories below the root aren't
allowed any special treatment (so `/about` cannot go to `/about/index.html`).

So with S3 you can have index documents but no HTTPS, and with CloudFront you
can have HTTPS but no index documents; not the end of the world (determined
customers can give up the pretty URIs), but a sad state of affairs nonetheless
for a product that's ostensibly in a mature state.

[1] Note that S3 does allow access over HTTPS through domains that look like
`*.s3.amazonaws.com`, but these domains don't get the features of an S3 static
website (index documents for example).

[cloudfront-root]: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/DefaultRootObject.html
[second]: /fragments/aws-static-hosting-workaround
[s3-endpoints]: https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteEndpoints.html
