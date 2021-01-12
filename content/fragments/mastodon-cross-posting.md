+++
hook = "Using Go, GitHub Actions, and a well-designed API to cross-post tweets to Mastodon."
published_at = 2021-01-12T05:54:58Z
title = "Cross-posting to Mastodon"
+++

I spent a few hours on Sunday writing a small program that [cross posts my tweets to Mastodon](https://github.com/brandur/mastodon-cross-post). The idea was seeded by talk of censorship, but I followed through because even though I'm not super optimistic about Mastodon's chances in appealing to the general public, I'll enthusiastically explore any possibility that might help compromise Twitter's harmful dominance of public discourse.

Mastodon's API is a breath of fresh air compared to Twitter's, which reminds you at every turn that it wants you the hell out, has the worst documentation since old school [MSDN](https://en.wikipedia.org/wiki/Microsoft_Developer_Network), and mandates use of ancient and difficult-to-implement protocols like OAuth1. With Mastodon, I clicked on the "developer" link and was up and running with a new app in about three seconds.

![Mastodon OAuth keys and API token](/assets/images/fragments/mastodon-cross-posting/keys.png)

Note that although I get the standard OAuth2 client key and secret, I'm also conveniently given an access token which I can put to use immediately -- no OAuth dances involved. This is a _critical_ development feature in getting up and running with minimal resistance, and every API platform should do it (looking at you Strava). It's especially great for indy hackers like me who want to build something simple which won't be productized.

Tweets can largely be re-posted to Mastodon without changes, but the program makes a few amendments for usability/clarity:

* Mastodon's avoided the cancer of tiny URLs (e.g. `t.co`), so these are expanded back to their original links.
* Retweets get truncated to 140 characters when returned through the Twitter API, so add a link back to the original RT to make them not completely inscrutable.
* Strip extraneous links that Twitter adds automatically. For example, any tweet containing photos automatically gets a link back to the tweet itself which isn't needed when the media is present (presumably for the benefit of older clients).

The program is stateless, so it scans the Mastodon API for where it left off before posting anything new. Comparing a Mastodon "toot" to an origin tweet isn't trivial because Mastodon marks up content with HTML after it's posted. I have a subroutine that does its best to deconvert it back to its original form, but allow some margin of error through a [Levenschtein distance](https://en.wikipedia.org/wiki/Levenshtein_distance) fuzzy match in case I missed something.

Copying over multimedia like photos was blessedly simple thanks to good API design on both the backend and in its [Go bindings](https://godoc.org/github.com/mattn/go-mastodon). Fetch data from Twitter's URLs, post it back up as a series of new media objects, then include those IDs when creating a Mastodon status. So easy that I had it working in minutes.

As usual, I always write this type of project in Go because I expect it to still be working five or ten years from now as it makes heavy use of the standard library (pulling in few dependencies) and the language has introduced no breaking changes.

![A sample status cross-posted to Mastodon](/assets/images/fragments/mastodon-cross-posting/mastodon-status.jpg)

The result? Working so far. Except I don't actually know anyone on Mastodon. Find me at [mastodon.social/@brandur](https://mastodon.social/@brandur) if you are.
