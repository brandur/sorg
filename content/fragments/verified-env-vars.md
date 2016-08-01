---
title: Verified Env Vars
published_at: 2016-07-29T16:30:17Z
---

While passing information to command line programs with environment variables
is a model that's both powerful and elegant, it's not without its problems.
Possibly the worst of them is the total lack of feedback of any kind on user
error. Can you tell which of these is right?

``` sh
GPGHOME=~/.gnupg/ gpg --list-secret-keys

# or

GNUPGHOME=~/.gnupg/ gpg --list-secret-keys
```

Well, the executable is called "gpg" so the first one seems to make sense, but
the suite is named "GnuPG", so the second seems like a reasonable assumption as
well.

The answer is that `GNUPGHOME` is right, despite the executable's naming. But
far more offensive than that design is what happens when you pick the wrong
one: GPG doesn't even acknowledge that there might be a problem and silently
behaves as if you hadn't specified anything at all.

As I was attending Heavybit's Devguild earlier this week, Rob Szumski pointed
out that Etcd solves this problem by verifying not just its command line flags,
but its env vars as well. It assumes ownership of any env var with a prefix of
`ETCD_` and notifies the user when:

1. When an env var with a prefix of `ETCD_` has been specified, but which the
   program doesn't know about.
2. When an env var has been specified, but "shadowed" by a flag which has
   equivalent meaning but whose value has taken precedence.

Neither problem produces an error (the first is a warning and the second a
notice) so that the program stays backward compatible even if a new version of
Etcd removes an env var that a user was sending a value in for. [Here's the
code][code] that accomplishes the effect

This design wouldn't directly solve the `GPG` vs. `GNUPG` problem above, but I
would suggest that in its spirit a well-designed GPG look for `GPG`-prefixed
env vars as well, and show a warning when a user tries to use one.

I'm a huge supporter of these minor tweaks to the "old ways" of CLI design that
have a disproportionately positive impact on improving usability compared to
the effort required to implement them. More information is always better.

[code]: https://github.com/coreos/etcd/blob/e2088b8073aa4f80a9d88b134ec71b749839db5b/pkg/flags/flag.go#L110,L127
