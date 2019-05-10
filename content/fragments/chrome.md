+++
hook = "On joining the party eight years late."
published_at = 2017-01-27T19:39:57Z
title = "Chrome"
+++

So put this in the bucket of "things that I do five years
behind the curve", but I recently switched my default
browser from Firefox to Chrome. The change wasn't exactly
by choice, but more the result of [new security
policy][marginal-security] combined with [lack of U2F
support in Firefox][firefox-u2f].

Luckily, it wasn't all bad. As surely everyone in the world
but me already knew, Chrome feels like a much more
responsive browser. The UI, tab switching, animations, and
page rendering are all noticeably speedier, and minor
features like tab pinning and proportional page zooming are
a lot of fun to play with.

One feature in particular that I've really enjoyed is
Chrome's "People" mechanic, which allows for browsing
sessions that are completely isolated from each other, for
say work and home. This, combined with Google account sign
on that syncs extensions and other settings between
computers, makes Chrome highly ergonomic to use.

Firefox's best feature, and the reason that I stuck with
the browser for so long, is [Vimperator][vimperator] which
brings Vim motions and shortcuts into the browser
environment. Vimperator is still the closest to a mouseless
experience on the web that you can get, and I wanted to
stick with it right up until the bitter end. It's an
amazing add-on, but more recently it hasn't been all
sunshine and rainbows, as Mozilla has announced it's moving
Firefox to [the WebExtensions API][web-extensions] which
will severely constrain what Vimperator will be able to do,
and has also been [breaking their add-ons API][tabopen-bug]
on a pretty regular basis and making developers wait a week
to get fixes approved for distribution in their add-ons
catalog.

Since the switch I've been using an equivalent in Chrome
called Vimium, and it tends to do as good job or better
than Vimperator for most things. Of particular note is its
"Vomnibar" that lets you pull up a special textbox to
search bookmarks, history, or open tabs, along with being
able to enter a new URL. Its major downside compared to
Vimperator is that it can't activate any shortcuts on
"special" Chrome pages like settings or a tab where the
browser is just displaying a raw media file like a JPG.
This seems pretty minor, but turns out to be annoying in
practice.

I'll miss Firefox, but am cautiously hopeful that ambitious
projects like Servo will eventually get them their lead
back by producing features that are harder to build within
the frameworks of existing browsers. For now, Chrome is
pretty good.

[firefox-u2f]: https://bugzilla.mozilla.org/show_bug.cgi?id=1065729
[marginal-security]: /fragments/marginal-security
[tabopen-bug]: https://github.com/vimperator/vimperator-labs/issues/671
[vimperator]: http://www.vimperator.org/
[web-extensions]: http://fasezero.com/
