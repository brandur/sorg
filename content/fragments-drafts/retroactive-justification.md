+++
hook = "Be wary of anyone hawking technology -- there's snake oil around every corner."
published_at = 2016-10-16T21:33:13Z
title = "Retroactive justification & skepticism"
+++

Last week [Slack wrote an engineering blog post][slack-php] whose overriding
message is that despite its dangerous type system (especially around coercion),
poor function calling semantics, and totally inconsistent standard library, PHP
really ain't all that bad. The implied message is that now that there's a total
fork of the language and runtime in the form of Hack and HHVM that addresses
some of these problems and makes the performance not totally atrocious, even
you might, you know, want to think about trying it out.

This is a prototypical example of retroactive justification. It comes up all
too often in our industry:

* Facebook selling PHP through Hack and the HHVM.
* [GitHub selling MySQL](/fragments/gh-ost).
* [Uber selling MySQL][uber-mysql].
* iOS developers who defended Objective-C as a great language right up until
  the very second that Swift rolled around.

I'd go as far as to say that _most_ of us are tied into some technologies that
we don't like and which are difficult enough to move off of that they're not
likely to go anywhere. 

The good news is that with enough engineering blood, sweat, and tears it's
possible to make _any_ system work, and not just work, but work well. Just ask
the banking industry, who have managed to keep an incredibly reliable product
that makes up the backbone of our economy despite running on a dead language
(COBOL) for decades now. But that doesn't mean that anyone else should _ever_
use COBOL.

However, some companies take things a step further by 

We could write a blog post about how MongoDB is a pretty good technology
because we've been able to build a system on it that's of high quality and very
reliable (which is true). We would of course want to omit the parts that to do
so we've had to start maintaining our own ORM/ODM library, need to put
incredible workarounds in place because there's no such thing as an atomic
transaction, and have written tens of thousands of lines of add-on code to keep
it stable and running. Doing so would be highly disingenuous, but I guarantee
that we could pen a convincing essay if there was the will to do so.

I would beg you to stay skeptical when reading these types of posts. Read
between the lines; ask yourself not just what they're saying, but also what
they're _not_ saying. When picking technologies, look for the ones that are in
wide use and which have accumulated only minute amounts of valid criticism
while having successfully withstood the test of time.

[slack-php]: https://news.ycombinator.com/item?id=12703751
[uber-mysql]:
