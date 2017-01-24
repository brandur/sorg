---
title: Marginal Security
published_at: 2017-01-24T16:48:49Z
hook: UNWRITTEN. This should not appear on the front page.
---

Some days I think that security, especially as it's applied
at big companies, is a race to the bottom in terms of who
can be the first to implement an environment that's so
restrictive that its employees are no longer able to
plausibly function at work.

Two of the latest casualties on an endless quest for
absolute security are Google's app passwords and second
factor security that's not based on the [Fido U2F][u2f]
standard. Both measures are designed to gain some marginal
security at the fringes. App passwords, although _perfectly
safe_ as long as they're stored on a secure system with an
encrypted disk, have a non-zero (and I mean non-zero, as in
_approximately_ zero, but not _exactly_ zero) chance of
leaking, and such a leak runs the risk of being
undiscovered for an extended period of time. Likewise,
existing second factor authentication methods have always
worked exactly as well as advertised, but U2F removes some
edges that the truly paranoid lose sleep over, like how a
TOTP-based factor is configured and entered by a fallible
human.

But those marginal security gains often come at the high
cost of destroying productivity as valuable tooling becomes
unusable. Restricting app passwords, for example,
necessarily implies the immediate death of any open
standard -- IMAP and clients like Mutt and offlineimap are
out. U2F is neat, but because it's new enough to not be
supported by most browsers [1]. Firefox has a plugin
available, but Google's U2F implementation in particular
relies on APIs that are coupled to Chrome, so using it
outside of Google's One True Browser is impossible.

It's important to note that the trade offs are not linear
-- eliminating the possibility of a highly improbable edge
case could cost hundreds of hours of employee productivity
every month as their tools become that much duller and work
slows as a result. Compromises need to be made somewhere --
the only way to achieve perfect information security is to
cut your network cables and bury your server at the bottom
of a mineshaft, and companies aren't exactly falling over
each other to do that. Security decisions should be
evaluated based on their impact on usability, preferably
with the input of the people being impacted, and security
teams should be willing to throw out a proposal that's too
expensive.

It's possible to build a room that's so impenetrable that
it's airtight, but doing so cuts off your own oxygen
supply.

[1] As of January 2017, only Chrome and Opera support U2F.

[u2f]: https://en.wikipedia.org/wiki/Universal_2nd_Factor
