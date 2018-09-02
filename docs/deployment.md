# Deployment

S3 buckets:

* `brandur.org`: Production deployment. Connected via CloudFront to
  [brandur.org](https://brandur.org).
* `brandur.org-dev`: Development deployment that includes drafts.
* `brandur.org-photographs`: A cache for photos that have been downloaded from
  the original Flickr source and resized so that we can avoid doing this work
  over and over again. Stored outside of the Git repository for size
  considerations.

CloudFront:

* `E2D97SPIHRBCUA`: Production. Points to S3 bucket `brandur.org`.
* `E10YZYIAIS23JX`: Development. Points to S3 bucket `brandur.org-dev`.

IAM:

* `brandur-org-policy` (policy): Policy that allows read/write access to the S3
  buckets listed above.
* `brandur-org-user` (user): User with `brandur-org-policy` attached. Owner of
  the credentials encrypted in `.travis.yml`.
