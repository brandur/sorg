---
title: An AWS Static Hosting Workaround
published_at: 2016-01-05T17:49:42Z
---

After writing a short article yesterday about [static hosting on AWS][first], a
couple people [1] tipped me off that a workaround for the lack of index documents
is instead generate "pretty URIs" by creating extensionless files and forcing
their content type to `text/html` when uploading them to S3.

The static generator I was using at the time (Hugo) didn't support creating
files without an extension, but I ended up writing my own pretty minimal
[replacement Go script][script] and it worked like a charm. My upload to S3 is
now two phases: one to send up HTML content and force its type, and a second to
upload assets (CSS, images, JS) where content type is detected normally. Using
the official [AWS CLI][awscli], that looks like this:

``` sh
# HTML
aws s3 sync ./public/ s3://$S3_BUCKET/ --acl public-read --content-type text/html --delete --exclude 'assets*'

# Assets (CSS, images, JS)
aws s3 sync ./public/assets/ s3://$S3_BUCKET/assets/ --acl public-read --delete
```

This works because S3 remembers the content type of an object that's uploaded
to it, and will serve the file with that same type on subsequent requests. It
may not be a perfectly ideal design, but it works.

[1] Thanks Mark McGranaghan and Ryan Smith.

[awscli]: https://aws.amazon.com/cli/
[first]: /fragments/aws-static-hosting
[script]: https://github.com/brandur/singularity/blob/master/main.go
