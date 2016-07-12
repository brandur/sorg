---
title: GPG + s3cmd
published_at: 2014-11-10T20:27:47Z
---

[s3cmd](http://s3tools.org/s3cmd) is a simple tool for use with Amazon's S3 and
CloudFront networks, which I tend to use quite a lot. Like many programs, it
deaults to storing your very sensitive AWS crdentials in plain text in a file
called `~/.s3cfg`, which is something that we can correct using GPG.

s3cmd makes this a little more challenging than average because its convention
is to generate `.s3cfg` by dumping its entire set of configuration. Luckily for
us though, as of s3cmd 1.5, configuration values are allowed to be the names of
environment variables, so we can pull in our sensitive values while leaving
most of the file unencrypted for ease-of-use:

```
[default]
access_key = $AWS_ACCESS_KEY_ID
...
secret_key = $AWS_SECRET_KEY
...
```

(Note that version 1.5 is still currently under development, and may have to be
installed as a pre-release through something like `brew install --devel
s3cmd`).

I then created a simple shell file containing my secrets which I stored to
`~/.aws-credentials`:

``` sh
AWS_ACCESS_KEY_ID=my-access-key
AWS_SECRET_KEY=my-secret-key
```

And encrypted with:

``` sh
$ gpg -r <your email> -e ~/.aws-credentials
$ ls ~/.aws-credentials.gpg
$ rm ~/.aws-credentials
```

Then elected for a simple wrapper script for `s3cmd`, which reads the encrypted
credentials file and exports environment appropriately (saved as
`~/bin/s3cmd-gpg`):

```
#!/bin/sh

# s3cmd-gpg

eval `gpg -q -d $DOTFILES/aws/credentials.gpg`
export AWS_ACCESS_KEY_ID
export AWS_SECRET_KEY

s3cmd "$@"
```

And finally, added to simple alias to my *rc file:

```
$ alias s3cmd="s3cmd-gpg"
```

From there, s3cmd can be invoked normally:

```
$ s3cmd ls
```
