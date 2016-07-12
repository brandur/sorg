---
hook: Building a static site on AWS with a global CDN, free HTTPS with automatic certificate
  renewal, and a CI-based deployment process powered by GitHub pull requests.
location: San Francisco
published_at: 2016-04-09T02:11:19Z
title: The Intrinsic Static Site
---

I've written previously [about my misgivings with static sites on
AWS](/fragments/aws-static-hosting), and while some of those points are still
valid, there's no question that the maturity of the AWS product and the
incredible surface area that its products cover easily make it one of the best
choices for hosting on the Internet.

Until now I would have recommended a CloudFlare/other host hybrid so that you
could get free HTTPS support for a custom domain with automatic certificate
renewal, but the advent of AWS Certificate Manager (ACM) was the last missing
piece in the all-Amazon stack. By composing a few services, we'll get a static
site with some pretty amazing features:

* Approximately unlimited scalability.
* As close to an operations-free experience as you're ever likely to find.
* Full HTTPS support with a certificate that's renewed automatically.
* A CDN with dozens of edge locations around the globe so that it'll load
  quickly from anywhere.
* A GitHub-based workflow which will make publishing new content as easy as
  merging a pull request.

Keep in mind that although this article walks through setting up a static site
step-by-step, you can [look at the full source code][singularity] of a sample
project at any time to dig into how things work.

## Building the Site (#building)

I'll leave getting a static site built as an exercise to the reader. There are
_hundreds_ of static site frameworks out there and plenty of good choices.

In my own experience, I've found that assembling my own static site generation
script takes about the same amount of time as getting on-boarded with any of
the major site generation frameworks. This isn't as much a critique of them as
it is a nod to the inevitability of any general-purpose project to expand in
features until it takes a lot of documentation and false starts to get up to
speed. Writing your own script may be a greater maintenance burden over the
long run, but this is offset by the much improved flexibility that it gets you.

The [singularity][singularity] example site uses a Go build script and a small
standard library-based web server with **fswatch** (a small cross-plaform
program that watches a filesystem for changes) to get a nice development
workflow that's fast and easy.

## AWS Service Setup (#aws)

We'll be using the [AWS CLI][aws-cli]. If you're on Mac or Linux, you should be
able to install it as simply as:

```
$ pip install --user awscli
$ aws configure
```

Running `aws configure` will ask for an AWS access key and secret key which you
can get by logging into the [AWS Console][aws-console]. You may also want to
configure a default region because some of the commands below will require one.

### S3 (#s3)

Amazon's storage system, S3, will be used to store the contents of our static
site. First create a bucket named according to the custom domain that you'll be
using for your site:

```
$ export S3_BUCKET=singularity.brandur.org
$ aws s3 mb s3://$S3_BUCKET
```

Next up, let's create a `Makefile` with a `deploy` target which will upload the
results of your build above:

``` makefile
# Makefile

deploy:
ifdef AWS_ACCESS_KEY_ID
	aws --version

	# Force text/html for HTML because we're not using an extension.
	aws s3 sync ./public/ s3://$(S3_BUCKET)/ \
        --acl public-read --delete --content-type text/html --exclude 'assets*'

	# Then move on to assets and allow S3 to detect content type.
	aws s3 sync ./public/assets/ s3://$(S3_BUCKET)/assets/ \
        --acl public-read --delete --follow-symlinks
else
	# No AWS access key. Skipping deploy.
endif
```

Normally assets uploaded to S3 are assigned a content type that's detected from
their file extension. Because I want to give my HTML documents "pretty URIs"
(i.e. extensionless like `/hello` instead of `/hello.html`) we play a trick
with content type that necessitates an upload in two steps:

* The first uploads your HTML assets and explicitly sets their content type to
  `text/html`.
* The second uploads all other assets. Here we allows the content type to be
  detected based on each file's extension.

Now try running the task:

``` sh
$ export $AWS_ACCESS_KEY_ID=access-key-from-aws-configure-above
$ export $AWS_SECRET_ACCESS_KEY=secret-key-from-aws-configure-above
$ make deploy
```

You can use the AWS credentials that you've configured AWS CLI with for now,
but we'll want to avoid the risk of exposing them as much as possible. We'll
address that problem momentarily.

### AWS Certificate Manager (#acm)

ACM is Amazon's certificate manager service. Using it we'll provision a
certificate for your custom domain which will then be attached to CloudFront.
After the certificate is issued, Amazon will automatically take care of its
renewal to keep your site accessible with perfect autonomy. You can even have
it issue you a wildcard certificate, and all for free!

Issue this command to request a certificate (and you can use a wildcard if you
want):

    aws acm request-certificate --domain-name singularity.brandur.org

AWS will email the domain's administrator to request approval to issue the
certificate. Make sure to track down that email and accept the request.

### CloudFront (#cloudfront)

CloudFront is Amazon's CDN service. We'll be using it to distribute our content
to Amazon edge locations across the world so that it's fast anywhere, and to
terminate TLS connections to our custom domain name (with a little help from
ACM).

Using the CLI here is a bit of a pain, so go to the [CloudFront control
panel][cloudfront-console] and create a new distribution. If it asks you to
choose between **Web** and **RTMP**, choose **Web**. Most options can be left
default, but you should make a few changes:

* Under **Origin Domain Name** select your S3 bucket.
* Under **Viewer Protocol Policy** choose **Redirect HTTP to HTTPS** As the
  name suggests, this will allow HTTP connections initially, but then force
  users onto HTTPS.
* Under **Alternate Domain Names (CNAMEs)** add the custom domain you'd like to
  host.
* Under **SSL Certificate** choose **Custom SSL Certificate** and then in the
  dropdown find the certificate that was issued by ACM above.

After it's created, you'll get a domain name for your new CloudFront
distribution with a name like `da48dchlilyg8.cloudfront.net`. It may take a few
minutes for the distribution to become available. You'll need this to set up
your DNS.

### Route53 (or Any Other DNS) (#dns)

Use Route53 or any other DNS provide of your choice to CNAME your custom domain
to the domain name of your new CloudFront distribution (once again, those look
like `da48dchlilyg8.cloudfront.net`). 

You should now be able to visit your custom domain and see the fruit of your
efforts!

### IAM (#iam)

Now that the basic static site is working, it's time to lock down the
deployment flow so that you're not using your root IAM credentials to deploy.

#### Create User (#create-iam-user)

Issue these commands to create a new IAM user:

    aws iam create-user --user-name singularity-user
    aws iam create-access-key --user-name singularity-user

Note that the second command will produce an **access key** and a **secret
key**. Make note of these.

#### Create Policy (#create-iam-policy)

Save the following policy snippet to a local file called `policy.json`. **Make
sure to replace the S3 bucket name with the one you used above.**

``` json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "s3:*",
            "Resource": "arn:aws:s3:::singularity.brandur.org"
        },
        {
            "Effect": "Allow",
            "Action": "s3:*",
            "Resource": "arn:aws:s3:::singularity.brandur.org/*"
        }
    ]
}
```

Now create the policy and attach it to your IAM user:

    aws iam create-policy --policy-name singularity-policy --policy-document file://policy.json

    # replace --policy-arn with the ARN produced from the command above
    aws iam attach-user-policy --user-name singularity-user --policy-arn arn:aws:iam::551639669466:policy/singularity-policy

The policy and user combination that you've just created scopes access to just
the S3 bucket containing your static site. If the worst should happen and this
user's credentials are leaked, an attacker may be able to take down this one
static site, but won't be able to probe any further into your Amazon account.

## Automating Contribution (#automating)

### Travis (#travis)

By putting file sychronization to our S3 bucket into a Make task, we've made
deployments pretty easy, but we can do even better. By running that same task
in a Travis build for the project, we'll make sure that anytime new code gets
merged into master, our static site will update accordingly with complete
autonomy.

We start by giving installing AWS CLI into the build's container and running
our Make task as the build's main target. That's accomplished by putting this
into `.travis.yml`:

``` yml
# travis.yml

# magic word to use faster/newer container-based architecture
sudo: false

install:
  - pip install --user awscli

script:
  - make deploy
 ```

That gets us pretty close, but the build will need valid AWS credentials in
order to properly deploy. We don't want to compromise our credentials by
putting them into our public repository's `.travis.yml` as plaintext, but
luckily Travis provides a facility for [encrypted environment
variables][encrypted-variables]. Get the Travis CLI and use it to secure the
IAM credentials for deployment that you generated above:

```
$ gem install travis
$ travis encrypt AWS_ACCESS_KEY_ID=access-key-from-iam-step-above
$ travis encrypt AWS_SECRET_ACCESS_KEY=secret-key-from-iam-step-above
```

After encrypting your AWS keys, add those values to your `.travis.yml`. under
the `env` section (make sure to use the special `secure:` prefix) so that our
build can pick them up:

``` yaml
# travis.yml

env:
  global:
    - S3_BUCKET=singularity.brandur.org

    # $AWS_ACCESS_KEY_ID (use the encrypted result from the command above)
    - secure: HR577...

    # $AWS_SECRET_ACCESS_KEY (use the encrypted result from the command above)
    - secure: svmpm...
```

Note that the plaintext values of these secure keys are only available to
builds that are happening on the master branch of your repository. If someone
forks your repository and builds their own branch, these values will not be
available and upload to S3 will occur.

[See a complete Travis configuration here][travis-yml].

### GitHub (#github)

Now that CI configuration is in place, you can push to a GitHub repository and
activate Travis for it.

Builds that occur on the master branch will automatically deploy their results
to S3 and they'll be available immediately. Pull requests still get a build and
have a test suite run, but because configured secrets are not available on
non-master branches, the deploy phase gets skipped, but you need only merge
them to master to have it run.

### Periodic Rebuilds With Lambda (#lambda)

One final (and optional) step in the process is to set up an AWS lambda script
that will be triggered by a periodic cron and which will tell Travis to rebuild
your repository. If you tell Travis to notify you on build failures in
`.travis.yml`:

``` yaml
notifications:
  email:
    on_success: never
```

Then you'll get an e-mail if that build ever fails. In case your content
repository isn't seeing very regular contributions, this will act as a canary
to tell you when if your build starts failing for any reason. Say that your IAM
credentials are accidentally invalidated for example.

First, you'll need to acquire your Travis API token. Get it using their CLI:

    gem install travis
    travis login --org
    travis token

Go to the [Lambda console][lambda-console] and select **Create a Lambda
function** (this is another one that's a little awkward from the CLI). When
prompted to select a blueprint, click the **Skip** button at the bottom of the
page. Give the new function a name and copy in the [script found
here][rebuild-script]. Change the configuration section at the top to include
your GitHub repository's name and the Travis API token acquired above. Under
**Role** choose **Basic execution handler**. Click through to the next page and
create the function. Use the **Test** button to make sure it works.

Now create a scheduled event so that the script will run periodically. Click
the **Triggers** tab and then **Add trigger**. Click the dotted grey box and
choose **CloudWatch Events - Schedule**. For **Schedule expression** put in
something like **rate(1 day)**. Note that Travis will rate limit you, and you
really don't need to be rebuilding very often, so a daily schedule is a
reasonable choice.

Now you're all set. AWS will handle triggering rebuilds, and if one fails,
Travis will notify you by e-mail.

## Summary (#summary)

In short, we now have a set of static assets in S3 that are distributed around
the globe by CloudFront, TLS termination with an evergreen certificate, nearly
unlimited scalability, and a deployment process based on pull requests that's
so easy that within five years you'll probably have forgotten how it works. And
despite all of this, unless you're running a _hugely_ successful site, costs
will probably run in the low single digits of dollars a month (or less).

[aws-cli]: https://aws.amazon.com/cli/
[aws-console]: https://aws.amazon.com/console/
[cloudfront-console]: https://console.aws.amazon.com/cloudfront/home
[encrypted-variables]: https://docs.travis-ci.com/user/environment-variables/#Encrypted-Variables
[lambda-console]: https://console.aws.amazon.com/lambda/home
[rebuild-script]: https://github.com/brandur/singularity/blob/master/scripts/lambda/index.js
[singularity]: https://github.com/brandur/singularity
[travis-yml]: https://github.com/brandur/singularity/blob/master/.travis.yml
