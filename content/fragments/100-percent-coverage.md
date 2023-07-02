+++
hook = "100% test coverage is annoying as hell, but probably good."
published_at = 2023-07-02T12:39:10-07:00
title = "100% test coverage"
+++

We run one of our projects not only with 100% test coverage, but 100% _branch_ coverage. It's deeply unpleasant. Every single change you make is likely to accidentally break test coverage, so if a normal development process is **1.** add feature, and **2.** write tests/fix bugs, here you have a three-step process of **1.** add feature, **2.** write tests/fix bugs, and **3.** hunt down every omitted branch and write increasingly convoluted tests to hit them.

And I know I'm not the only skeptic. Company coverage requirements are a commonplace complaint from working developers, right up there with slow test suites and annoying management.

But 100% coverage definitely helps, especially in languages like Ruby where an untaken branch can contain a serious error that's accepted by the interpreter, but will take down production if hit.

While thinking about whether 100% coverage was an net good or ill, it occurred to me that one of its underappreciated aspects is that it's a ratchet. Another project we run is built by a small team of two and doesn't require 100% coverage. Its test suite is good, and we have an informal policy that tests are exhaustive-but-not-100%-exhaustive, but the only thing to guarantee that are good taste on the part of each developer and to a less extent, code reviews [1]. If we were to add a third person, hopefully they'd share our aesthetic preference around tests, or at least learn to emulate it, but there's no guarantee. If that third person is still chronically underwriting tests months into their tenure, undertested features will proliferate, and more regressions can be expected as changes are riskier to make.

I use the example of going from two to three people, and person number three may or may not have good testing intuition, but once you have an org of 10 or 100, you're guarantee to have someone who does not.

So as unpleasant as it is, I think I still come down on the side that 100% coverage is good. The tests themselves are useful for preventing regressions, but the more important piece might be the tight baseline for testing practice that it enforces across a larger team.

[1] Code reviews tend to catch some missing tests, but it's a huge amount of effort for a reviewer to go through and try to identity _every_ test case that might've been desirable, but omitted.