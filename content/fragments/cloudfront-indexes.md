---
title: CloudFront Indexes
published_at: 2016-07-18T16:34:33Z
hook: UNWRITTEN. This should not appear on the front page.
---

Some of us are vain enough to want to be able to use extensionless filenames
when hosting a static site through CloudFront. I previously wrote about [a
workaround][workaround] that can be used to achieve this effect by stripping a
file's extension and forcing its content type when uploading to S3. I now
realize that this is about half the solution; it does get us extensionless
URLs, but doesn't handle all the edge cases around indexes.

If I have a blog and want my articles to have URLs like `/articles/my-article`,
then it stands to reason that I might also want an index of all my articles at
`/articles`. If you're building your static content locally and then syncing it
up to S3 using `aws s3 sync`, you can't have both a file called `articles` (the
index) and a directory called `articles` (the container for articles) on your
filesystem. In other words, this isn't allowed:

```
filesystem/
  - articles
  + articles/
    - my-article
```

This might present us with a big problem, except that S3 is _not_ a filesystem,
it's a key/value store wherein we occasionally pretend that keys containing
slashes are a nested hierarchy of directories. So although not kosher on a
filesystem, this layout is perfectly okay on S3:

```
s3-bucket/
  - articles
  - articles/my-articles
```

Thus solving the problem above of having an index and container with the same
name.

Publishing the site isn't quite as trivial as running `aws s3 sync` anymore,
but can still be made easy by uploading in two passes:

1. Sync all non-index pages with `aws s3 sync`.
2. Upload all index pages with `aws s3 cp` after correcting their path name.

I generate index files locally with an HTML extension so that they're more
discoverable if I point web servers at the directory; they'll often know how to
handle files named `index.html` by default (like say Go's `http.FileServer`).
So our final local directory structure looks like this:

```
filesystem/
  - index.html
  + articles/
    - index.html
    - my-article
```

Below are a few slightly convoluted shell commands to do the upload. The last
step iterates a content directory called `public`, finds any files named
`index.html`, and uploads them to S3 as the name their parent directory them
(e.g. `articles/index.html` becomes `articles`):

``` sh
# Step 1: HTML
aws s3 kync ./public/ s3://$S3_BUCKET/ --acl public-read --content-type text/html --delete --exclude 'assets*'

# Step 2: Assets (CSS, images, JS)
aws s3 sync ./public/assets/ s3://$S3_BUCKET/assets/ --acl public-read --delete

# Step 3: HTML index files
find ./public -name index.html | egrep -v './public/index.html' | sed "s|^\./public/||" | xargs -I{} -n1 dirname {} | xargs -I{} -n1 aws s3 cp ./public/{}/index.html s3://$S3_BUCKET/{} --acl public-read --content-type text/html
```

There is also one finishing touch to be made: the system described here will
handle all index files _except_ the one at the very top-level, which we cannot
upload with a name like `/`. Luckily CloudFront has provided a workaround by
allowing us to configure a **Default Root Object**, which is the name of the
file to serve from the distribution's root. Configure it to `index.html`, make
sure that a corresponding object gets synced to the S3 bucket, and you're
done.

[workaround]: /fragments/aws-static-hosting-workaround
