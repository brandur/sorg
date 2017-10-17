---
title: "Changing a Universal Constant: How 280-character Tweets Break an API"
published_at: 2017-10-17T17:10:01Z
hook: Even changing the length of a field is a
  backwards-incompatible change if it's deeply entrenched.
---

Recently Twitter announced that they're [rolling out
support for 280-character tweets][ann]; up from the
140-character limit that's been in place since the company
was founded. The controversy around whether that was a good
decision continues to rage on.

More interesting than the debate is how Twitter treated
adding the new longer tweets to their API. Returning new
longer tweet content in a field that's been expected since
the beginning of time to always hold 140 characters would
be tumultuous to say the least -- for thousands of existing
applications it would be as if the universal constant of
gravity, something they've always taken for granted to be a
_constant_, suddenly shifted from 9.81 m/s<sup>2</sup> to
15 m/s<sup>2</sup>.

To avoid breaking those applications Twitter has treated
moving to 280 characters as a [backwards-incompatible API
change][stripe]. Tweets come back from the API containing a
set of fields including `text`, which is each tweet's
140-character content:

```
GET /1.1/statuses/user_timeline.json?screen_name=brandur

[
    {"id":123, "text":"tweet123 content", ...},
    {"id":124, "text":"tweet124 content", ...},
    {"id":125, "text":"tweet125 content", ...},
]
```

Even today for tweets that contain up to 280 characters,
`text` comes back with no more than 140 characters; content
that runs over is truncated with an ellipsis (…).

The rest of each tweet is still accessible, but clients
must ask for it explicitly by specifying
`tweet_mode=extended`. When they do, `text` is replaced by
`full_text`, which can hold up to 280 characters.

```
GET /1.1/statuses/user_timeline.json?screen_name=brandur&tweet_mode=extended

[
    {"id":123, "full_text":"tweet123 *extended* content", ...},
    {"id":124, "full_text":"tweet124 *extended* content", ...},
    {"id":125, "full_text":"tweet125 *extended* content", ...},
]
```

The API change ensures that only clients who have vetted
themselves get the upgrade, and by extension guarantees a
smooth transition. There is a downside though: especially
for new users, Twitter's API is forever more complex to
understand. If you're writing a program from scratch and
have never made the 140-character assumption, having to
specify a special parameter to get a normal-length tweet
back isn't going to make a lot of sense.

This is where API versioning comes in. Twitter's
technically using path-based versioning, and could
increment the `/1.1` that prefixes every URL to a `/1.2`.
They haven't though, and I suspect that they won't for some
time. The trouble with such an explicit versioning scheme
is that it introduces two divergent schemes that will both
need to be maintained in near perpetuity (you're not going
to have everyone moving off of `1.1` anytime soon).

This is a near-perfect case study in compatibility because
a change in content length seems totally benign on the
surface, but the effort involved shows just how far careful
platforms are willing to go not to break their users.

[ann]: https://blog.twitter.com/official/en_us/topics/product/2017/Giving-you-more-characters-to-express-yourself.html
[stripe]: https://stripe.com/blog/api-versioning
