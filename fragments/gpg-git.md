---
title: GPG + HTTP Git
published_at: 2014-12-11T21:48:06Z
---

Along with a few other protocols like SSH, Git can communicate with a remote
over HTTP. Git's HTTP transport uses basic authentication to identify a user,
which is nice because it allows credentials to be consolidated with other tools
like Curl.

Git actually uses libcurl to support its HTTP implementation which means that
like Curl, it supports credentials stored in `.netrc` out of the box, although
this seems almost somewhat accidental. Modern practice suggests the [use of
credential helpers][credential-helpers] to help procure credentials. Luckily,
someone has taken the time to write a [good credential helper for
`.netrc`][netrc] that will automatically prefer `.netrc.gpg` if one is
available. Simply install it somewhere in your `$PATH` and make it executable:

``` sh
$ curl https://raw.githubusercontent.com/git/git/master/contrib/credential/netrc/git-credential-netrc > ~/bin/git-credential-netrc

$ chmod +x ~/bin/git-credential-netrc
```

Now you need to tell Git to use this credential helper. Add this section to
`~/.gitconfig`:

```
[credential]
    helper = netrc
```

That's it! Git will now read host out of your `~/.netrc.gpg` when attempting to
authenticate with a remote server over HTTP.

[credential-helpers]: https://www.kernel.org/pub/software/scm/git/docs/v1.7.9/technical/api-credentials.html
[netrc]: https://github.com/git/git/blob/master/contrib/credential/netrc/git-credential-netrc
