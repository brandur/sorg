+++
hook = "A few thoughts on GitHub's practice of keeping their code synchronized with Rails `main`."
published_at = 2023-04-07T11:24:07-07:00
title = "Stay mainline"
+++

GitHub writes about how they [stay current on mainline Rails](https://github.blog/2023-04-06-building-github-with-ruby-and-rails/). Takeaways:

* A pull request is opened automatically every week targeting Rails `main`.

* Weekly incremental updates avoid huge monolithic projects to bump a big dependency across many versions.

* GitHub developers always get access to the latest Rails features. Security posture is improved.

* All Rails monkeypatches have been unwound. Instead of monkeypatching, GitHub developers should contribute changes upstream to the public Rails project.

* The practice becomes a huge regression hedge for Rails itself. The load of tens of thousands of tests on new `main` contributions is going to catch a lot of problems that the Rails test suite may have missed.

I have huge admiration for GitHub going this direction. Staying mainline is obvious as a concept, but overwhelmingly non-default in practice.

Without strongly defined process, the overwhelming default is to leave the underlying stack as it is. Over time, what was brand new when the company started becomes dated, and eventually becomes old. At some point it's realized that an upgrade is necessary, but by then it's become an undertaking of epic proportion involving months of work and full teams of people. Instead of dependency updates being discrete, incremental work like climbing a flight of stairs, it becomes like trying to jump your way to the next floor.

And the pressure to _diverge_ is intense. In any modern endeavor inevitably a large set of dependencies is created to build faster and more accurately by taking advantage of the work that other people have already done. Eventually, an engineer notices either a bug, a small missing feature, or some suboptimal behavior in one of those dependencies. Trying to patch it via upstream contribution to the open-source project is _hard_ -- writing the code might be easy, but rationalizing the change could be less so, and negotiating its integration into the project and waiting for a release is a slow process at the best of times.

A _much_ faster way to do it is to fork the dependency, make your change, and have bundler target the fork's Git. At my last job we had a couple early engineers who wouldn't hesitate for _even a second_ to follow this playbook, and one-by-one every major dependency was forked: Thin, Rack, Sinatra, etc. Analyzing someone else's complex codebase and making a change takes non-trivial technical skill, so not only did they not feel bad about it, but afterwards they'd bask in their own engineering ingenuity and a job well done. What they weren't thinking about was the team of engineers who'd come years later, and who now not only needed to upgrade an ancient dependency, but also figure out what'd been done to it and unwind that. A fork created in 10 minutes at the drop of a hat was now a tech debt project that'd take long weeks to safely unwind.

Staying mainline is obvious in concept, rare in practice, and a win of incalculable magnitude where implemented. What GitHub is doing with Rails is doing this exactly right -- bravo.