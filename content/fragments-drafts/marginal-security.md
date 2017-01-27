---
title: Marginal Security
published_at: 2017-01-27T17:35:36Z
hook: Your servers are only perfectly secure when they're
  buried at the bottom of a mineshaft without network
  connections. Security is about trade offs, and must be
  designed rationally.
---

Some days I think that security, especially as it's applied
at big companies, is a race to see who can be the first to
implement an environment that's so uncompromisingly
obstructive that its employees can no longer plausibly
function at work.

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
existing second factor authentication methods work exactly
as well as advertised, but U2F removes some edges that the
truly paranoid lose sleep over, like how a TOTP-based
factor is configured and entered by a fallible human.

But those marginal security gains often come at the high
cost of suppressing productivity as valuable tooling becomes
unusable. Restricting app passwords, for example,
necessarily implies the immediate death of any open
standard -- IMAP and clients like Mutt and offlineimap are
out. U2F is neat, but because it's new enough to not be
supported by most browsers [1]. Firefox has a plugin
available, but Google's U2F implementation in particular
relies on APIs that are coupled to Chrome, so using it
outside of Google's One True Browser is impossible.

The trade offs are not linear. Eliminating the possibility
of a highly improbable edge case could cost hundreds of
hours a month in employee productivity as their tools
become duller and work slows as a result. Compromises need
to be made somewhere -- the only way to achieve perfect
information security is to cut your network cables and bury
your server at the bottom of a mineshaft, and companies in
big enterprise aren't falling over each other to do that.
Security decisions should be evaluated based on their
impact on usability, preferably with the input of the
people being impacted, and security teams should be willing
to throw out a proposal that's too expensive.

It's possible to build a room that's perfectly
impenetrable, but doing so will cut off your own oxygen
supply. We rely only on our own rationality and ability to
recognize and balance trade offs not to do so.

[1] As of January 2017, only Chrome and Opera support U2F.

[u2f]: https://en.wikipedia.org/wiki/Universal_2nd_Factor
