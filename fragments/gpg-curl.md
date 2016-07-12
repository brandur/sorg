---
title: GPG + Curl
published_at: 2014-11-10T20:14:35Z
---

A very convenient feature of Curl is that if invoked with `curl -n`, it will
try to read credentials of a target service our of a local `.netrc` file, and
if found, use them for authentication. The trouble is that these credentials
default to being stored in plain text, which is something that we'd like to
avoid by using GPG.

The first step here is to encrypt your `.netrc`:

``` sh
$ gpg -r <your email> -e ~/.netrc
$ ls ~/.netrc.gpg
$ rm ~/.netrc
```

Now we can can pipe the decrypted output of our `.netrc` file from gpg, and
have Curl read it in (this should go in your appropriate *rc file):

``` sh
$ alias curl="gpg --batch -q -d ~/.netrc.gpg | curl --netrc-file /dev/stdin"
```

Because we've folded this into an alias, curl can be invoked normally:

```
$ curl -n https://api.heroku.com/apps
```
