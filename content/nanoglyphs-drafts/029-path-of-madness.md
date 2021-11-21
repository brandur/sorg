+++
image_alt = "A road on Bunaken in Indonesia"
image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/029-path-of-madness/bunaken-road@2x.jpg"
published_at = 2021-10-02T21:29:43Z
title = "The Path of Madness"
+++

Hacker News is the Twitter tech elite's favorite outfit to hate, taking regular scorching criticism as that "orange website", and perceived to be full of nothing better than a hypercritical mob informally operating as the world's most cynical peanut gallery.

But even its most vocal critics can't seem to help themselves -- despite the frequent condemnation they lob its way, someway, somehow, they keep finding themselves back. And there's a reason for that -- it's not all bad, and actually, most of it is even pretty good. Not only that, but once in a while, you strike internet gold and come across information that you could never have found anywhere else, reminding you that this elaborate series of tubes has a few good things going for it after all.

One of my [favorite comments of all time](https://news.ycombinator.com/item?id=18442941) is from an ex-Oracle engineer who describes what development there is like. Not just _at_ Oracle the company, but on the core Oracle database product itself (to be read in the tone of a Lovecraft short story):

> Oracle Database 12.2.
>
> It is close to 25 million lines of C code.
>
> What an unimaginable horror! You can't change a single line of code in the product without breaking 1000s of existing tests. Generations of programmers have worked on that code under difficult deadlines and filled the code with all kinds of crap.
>
> Very complex pieces of logic, memory management, context switching, etc. are all held together with thousands of flags. The whole code is ridden with mysterious macros that one cannot decipher without picking a notebook and expanding relevant pats of the macros by hand. It can take a day to two days to really understand what a macro does.
> 
> Sometimes one needs to understand the values and the effects of 20 different flag to predict how the code would behave in different situations. Sometimes 100s too! I am not exaggerating.
>
> The only reason why this product is still surviving and still works is due to literally millions of tests!
>
> Here is the life of an Oracle Database developer:
>
> - Start working on a new bug.
> - Spend two weeks trying to understand the 20 different flags that interact in mysterious ways to cause this bag.
> - Add one more flag to handle the new special scenario. Add a few more lines of code that checks this flag and works around the problematic situation and avoids the bug.
> - Submit the changes to a test farm consisting of about 100 to 200 servers that would compile the code, build a new Oracle DB, and run the millions of tests in a distributed fashion.
> - Go home. Come the next day and work on something else. The tests can take 20 hours to 30 hours to complete.
> - Go home. Come the next day and check your farm test results. On a good day, there would be about 100 failing tests. On a bad day, there would be about 1000 failing tests. Pick some of these tests randomly and try to understand what went wrong with your assumptions. Maybe there are some 10 more flags to consider to truly understand the nature of the bug.
> - Add a few more flags in an attempt to fix the issue. Submit the changes again for testing. Wait another 20 to 30 hours.
> - Rinse and repeat for another two weeks until you get the mysterious incantation of the combination of flags right.
>
> ...
>
> The above is a non-exaggerated description of the life of a programmer in Oracle fixing a bug. Now imagine what horror it is going to be to develop a new feature. It takes 6 months to a year (sometimes two years!) to develop a single small feature (say something like adding a new mode of authentication like support for AD authentication).
>
> The fact that this product even works is nothing short of a miracle!

A miracle indeed.

---

## Edit-compile-run (#edit-compile-run)

Engineers who've spent their careers at smaller companies may not fully appreciate a situation like this. Although Oracle may be a particular egregious example, it's a disturbingly common scenario at larger shops where head count's been scaled up. They sink slowly into the quagmire, and once in, find it impossible to get themselves back out.

**Edit-compile-run** is a software engineer's work loop in which they (1) edit code, (2) compile the code (or start the interpreter), and (3) run the program or test suite. It's the all-important standard workflow an engineer will run hundreds of times a day, and the speed at which it's possible is one of the single most important factors for productivity in a codebase. Whether new code is being written or existing code modified, being able to run it and get feedback quickly is of paramount importance.

The ideal edit-compile-run loop for small programs is < 1s, or measured in a couple seconds. For larger programs, 10s of seconds is more likely and still acceptable, but go any higher than that and you're entering the danger zone. As soon as the loop is long enough that engineers will go find something else to do while they're waiting on it then you've just taken a massive productivity hit in large and frequent context switches.

---

## The fool's path (#the-fools-path)

I want you to run a thought experiment with me. It's the story of an expanding engineering organization, and not an uncommon one around the hypergrowth shops in the valley. It's the story of how taking the easy path can be disastrous, and how a series of [locally optimal decisions lead to an overall _very suboptimal_ end result](https://en.wikipedia.org/wiki/Tyranny_of_small_decisions).

You're building a company, and that company's product is a web service. Like many web services, it has a test suite, and this test suite is comprehensive by design -- a widely accepted best practice that's been grilled into you over the years.

The test suite's also important because you've been sold on an interpreted language as the fastest way to get work done, but a downside of that is lack of any built-in compiler or static analysis, so short of sending changes right into production, the test suite's the only way you know anything works.

Your product and users grow, and so does your application. What was hundreds of lines of code becomes tens of thousands. The test suite, which used to run in seconds, is now kind of slow, taking many full minutes to grind through end to end. You've ramped up hiring and are expanding the application aggressively, so this is expected to get worse before it gets better. Projecting a few months down the road, minutes will become tens of minutes, and tens of minutes will become hours.

A member of your team comes up with a clever solution. Although you're not writing in a language that's conducive to parallelism, by moving test runs into the cloud as part of CI, the entirety of the test run could be broken up across multiple worker nodes and its results coordinated by a leader, significantly improving its speed. Even better, this scales well horizontally with more tests and more people -- the very textbook definition of parallelizable. As more engineers join (who are pushing more branches) and as more lines of code are written, the size of the worker pool can be continually increased. Horizontal scaling yields results that would have been impossible before the advent of cloud computing. 1,000 nodes cooperating for a mere 10 minutes of real time produces a full week's worth of computation.

It works beautifully. Tests are fast again, and engineers are productive and happy. But a new problem appears on the horizon, first as a tiny pinprick point, but quickly coming into horrible, razor sharp relief. Test workers are being scaled as predicted, but the number needed isn't growing linearly with engineers and lines of code as expected, but rather quadratically. It's not well understood why, but it appears that the suite isn't only getting slower through additional test cases, but also because _each individual test_ is getting slower too. It doesn't take long for the problem to become apparent -- with all testing being offloaded to the cloud, and little feedback when an unintentionally slow test is added to the suite, engineers no longer have any intuitive visibility into how long tests case, and no incentive in writing efficient ones.

Running the entirety of the test suite locally used to be slow. Now, it's a pipe dream. Adding up the timing of a single test run across all parallel workers shows that a full run would now take 3+ days if performed sequentially. Massive parallelization is now the only way that running it is possible at all.

But despite aggressive cloud parallelization, the suite's now become slow enough that not even the cloud run is fast, now taking around 10 minutes of wall clock time. The slow development loop is back, and once again eating into engineering productivity. A new system is devised to ping users on Slack when their builds finish or error. This doesn't solve the speed problem, but at least cuts out the unnecessary downtime between when a test suite fails and when an engineer realizes it failed.

Another problem emerges. Even with a sizable Silicon Valley budget and a policy of spend-what-you-need when it comes to AWS, this is getting expensive. _Really_ expensive. Hundreds of large, highly priced worker nodes are needed to keep the system afloat, and total costs have more zeroes on the end of them than you like to think about.

A new clever scheme is devised. Engineers tend to focus work on core Pacific hours, so an autoscaling policy is created to put most of the workers to sleep for most of the day. Unfortunately, this win is short-lived as the organization continues to expand and new offices are added in more distant timezones. There are now engineers working around the clock.

To address slow feedback and expense, significant investment is poured into an advanced selective test running scheme. It analyzes which code is run by which tests, and uses the inversion of that data to aggressively prune the test graph for any given change. Patches that affect hot code in the core still run most of the suite, but changes closer to the periphery skip large parts of it, making builds faster.

It works. Costs are reduced, and most engineers (who are working closer to the edges of the dependency graph) get faster test runs. But the win doesn't come for free: the selective test runner is very complex software and exhibits bugs on a regular basis, causing build problems and confusion amongst engineers. CI as a whole is now so big and so complicated that it's down on a semi-regular basis, hamstringing the entirety of the engineering force for hours at a time. A full team is spun up to maintain and troubleshoot it.

Meanwhile, test overhead has grown enough that even individual test files are quite slow to run locally. A local run of the entirety of the test suite has been such an impossible fantasy for so many years now that no one even thinks about it anymore. Cloud development boxes are procured for each engineer to streamline the development setup process, and to make test runs faster by offloading work to more powerful cloud machines. That works somewhat, but costs more money, isn't much faster, and development sans high speed internet connection is now impossible.

Your programming language's lack of a rigid type system makes coding mistakes common -- a feature that doesn't pair well with slow iteration loops. The company sets out to build an in-house type annotation system and static analysis tool for a language whose creators want neither. The project proves successful -- with the static analysis tool usually able to run faster for the whole codebase than even a single test case -- but its development and maintenance are now a permanent carrying cost, with no hope of making it back to core, and it will never work as well as the built-in equivalents of any sister language.

This trajectory continues unabated, made possible by a combination of brute forcing infrastructure costs and human sweat. The organization is now well passed the point of no return, and no real alternative is possible. It's healthy in every non-technical sense so it may be able to tread water forever, but without any real hope of rescue. [1]

---

Now, I'll leave it at that this story isn't purely hypothetical. It could even happen to you.

The lesson: Clever patches might seem ingenious, but clever patches have secondary effects. Worse yet, the secondary effects _have secondary effects_. This pattern propagates endlessly, growing in complexity and expense with every new layer.

There is a solution: Instead of taking the easy path today, Fix. Root. Problems. If tests are slow, the right fix isn't to send them to the cloud, it's to _fix the tests_ and implement ratchets to make sure they stay fixed. The former is enticing in that it's easier today, but it will be much, much harder tomorrow.

---

<img src="/photographs/nanoglyphs/029-path-of-madness/bunaken-pier@2x.jpg" alt="A view from a pier on Bunaken in Indonesia" class="wide" loading="lazy">

## Predestination (#predestination)

I've gotten into more than one argument over the years with apologists who posit that at scale, there's no alternative to a slow, multi-day test suite -- that Oracle's sordid situation is inevitable for any company interested in growth. I don't buy it, and neither should you.

To make my case, I'm going to fall back to my old, reliable friend, Postgres. Like Oracle, it's a relational database, and has by almost all objective and subjective standards, now surpassed Oracle in speed and features for everything that matters. Like Oracle, it's also written in C. Like Oracle, it has a build farm for changes to be checked in the cloud.

_Unlike_ Oracle, compiling it and running its full barrage of test suites is shockingly fast. Here it is clocking in at ~31 seconds (compile) and 14 seconds (test) on my machine, which is a consumer-grade laptop unplugged from its power source:

```
$ time make -j8 -s
All of PostgreSQL successfully made. Ready to install.

real    31.36s
user    135.81s
sys     25.86s

$ time make install -j8 -s
PostgreSQL installation complete.

real    1.95s
user    1.44s
sys     1.09s
```

```
$ time make installcheck-parallel

...

=======================
 All 209 tests passed.
=======================


real    14.03s
user    1.17s
sys     1.71s
```

Postgres does have a more exhaustive test suite (separated out because that edit-compile-run is important), but even that's pretty quick for such a thing:

```
$ time make installcheck-world -j8


======================
 All 98 tests passed.
======================


real    90.66s
user    6.52s
sys     4.80s
```

So for argument's sake, we'll take the higher number of 90 seconds. 90 seconds in Postgres versus 30 hours in Oracle, both of which are functionally similar products. We're not talking about a difference of 20%, 2x, or even 20x here -- we're talking about a difference of **1200x**. Even if you think that Oracle's providing some kind of rich feature set that Postgres isn't, or is tested more robustly, that can't come even close to explaining that sort of vast differential. Undoubtedly, there are many intertwined factors that led to the two very different outcomes, but methodology, taste, and attention to detail are absolutely amongst them.

### Getting testing right (#testing-right)

Go is another favorite. Its lack of generics are a common complaint, but we should remember the upsides of that, one of which is that it keeps the compiler really, _really_ fast.

Compilation speed is often not even a factor in the edit-compile-run loop because it's practically negligible. And that's only one of the tricks Go uses to stay quick. Some others:

* Test results are cached aggressively. When iterating you can still call your top-level `go test ./...` command on every loop and it only takes a second as just the affected tests get run.

* Tests for different packages run in parallel automatically without the user having to configure anything.

* If there are slow tests within a specific package, Go provides an easy `t.Parallel()` test helper to flag tests that should be run in parallel with each other. No need for a heavy non-standard test framework to come in on top.

The result is a development loop where there's no hesitation in compiling or running the test suite. Want to try something quick? Do it, then just run everything. Make a syntax mistake? That happens. Lucky for you, that's only a second in lost time. Test quickly. Test constantly.

And that's not even saying anything about its great LSP engine that plugs easily into IDEs like VSCode -- most of the time you get all your compilation problems fixed without even leaving the editor. (F8, F8, F8, F8, ...)

Go has its blemishes, but these tight and fast development loops are a game changer in productivity. If only all languages felt so fast and productive software development would be on a totally different level.

### Anecdata (#anecdata)

Some anecdotes from Crunchy. Our software is much smaller than my last job, but it's written by people with good intuition around potential bottlenecks and pitfalls that may be encountered during testing, and I'd contend that even were it to grow 100x, we'd still be in a pretty good place.

Our Go codebase runs in a little under four seconds of real time:

```
$ time go test -count=1 ./...
?       github.com/crunchydata/priv-all-crunchy-cloud-manager/backend   [no test files]
ok      github.com/crunchydata/priv-all-crunchy-cloud-manager/backend/apiendpoint       0.139s
?       github.com/crunchydata/priv-all-crunchy-cloud-manager/backend/apiendpoint/apiendpointtest       [no test files]
ok      github.com/crunchydata/priv-all-crunchy-cloud-manager/backend/apierror  0.094s

...

ok      github.com/crunchydata/priv-all-crunchy-cloud-manager/backend/tools/src/uuid-to-eid     0.142s
ok      github.com/crunchydata/priv-all-crunchy-cloud-manager/backend/util      0.248s
ok      github.com/crunchydata/priv-all-crunchy-cloud-manager/backend/validate  0.253s

real    3.66s
user    14.87s
sys     3.26s
```

That's with caching disabled, so a vanilla `go test ./...` is even faster. This is also commodity hardware running on battery power.

The suite stands at 1120 test cases total, most of which hit the database. That's not a lot compared to a larger company's, but it's still sizable, and with a compiled language, it's not as necessary to test every possible branch which could contain potentially non-sensical code. The test suite is complete, but not exhaustive.

Here's our Ruby codebase (Owl) clocking in at 15 seconds:

```
$ time bundle exec rspec

Randomized with seed 59463
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
..............................................................................
.........

Finished in 12.17 seconds (files took 2.07 seconds to load)
1413 examples, 0 failures

Randomized with seed 59463

Coverage report generated for RSpec to /Users/brandur/Documents/crunchy/owl/coverage. 15088 / 15977 LOC (94.44%) covered.

real    15.03s
user    7.46s
sys     1.61s
```

A little slower, but for a Ruby codebase, still very good. At my last job we had (a great number of) individual test cases that were full multiples slower than this whole suite, and 15 seconds isn't too far off from what a developer could expect to wait for a local test runner in development to spin up and execute a single test.

And this is with very little effort towards optimization. The fact that `user` (CPU time running user code) << `real` (wall clock) suggests that it's probably spending a lot of time blocking on I/O, which means that there's a good chance that even Ruby's [GIL-blocking threads](/nanoglyphs/018-ractors) would speed things up.

---

<img src="/photographs/nanoglyphs/029-path-of-madness/padan-fain@2x.jpg" alt="Wheel of Time: Padan Fain" class="wide" loading="lazy">

## The wheel turns (#the-wheel-turns)

Exciting news for fantasy fans last week -- the TV adaptation of the Wheel of Time (WoT) started steaming on Amazon Prime. For anyone not initiated, WoT's one of the major Tolkien-inspired fantasy epics from the last 50 years similar to the much better known Game of Thrones (GoT). WoT's a little less focused on the intrigue of rival nations, and much more heavy on the world building and depth of astounding magnitude. Although a disciplined and prolific writer, Robert Jordan in later years also fell victim to what I'll call the "GRRM trap", and each book in the series got progressively longer, spaced further apart, and with less happening in each one. As it was becoming undeniably obvious that the series would never be tied up in the original ten-book target, Jordan passed away, leaving the future of the work in limbo. Eventually, Brandon Sanderson, another fantasy author prolific in his own right, was nominated by the estate to finish WoT, and managed to get it wrapped up in a not-so-brief set of thirteen books, each of which is comparable in length to the entirety of Lord of the Rings.

Fantasy-wise, the series is distinctive because you won't find a single mention of magic, elves, or wizards throughout, although parallels exist. The word "Dragon" does appear many times, but as the title of foretold hero rather than Smaug-esque fire-breathing beast.

What sets WoT apart from its sibling epics is the sheer depth of _everything_. Established Tolkien readers will know that books like the Silmarillion dived deeply into the history of Middle Earth, painting a rich tapestry of background and concepts that didn't even garner a mention in the books most people know. WoT takes that to the next level -- not only do you get history, but a deep familiarity with the customs of a dozen different nations, intricate deals on how The One Power (magic) works and is used, and hundreds of well-defined characters. Concepts only briefly hinted at early on are often fully fleshed out into a deep mythos by the end. Unlike WoT, GoT isn't so reliant on the unexplained mysteries of the non-human world (dragons, white walkers, Valyrian steel, etc.) to drive narrative -- those alternative worlds are still be important, but they're fully deconstructed in their own right.

Capturing this beast in a TV series is a nigh impossible task. So far though (after the initial three-episode drop) ... I'm cautiously optimistic. They _nailed_ the casting for Lan and Moraine, and Rand, Mat, and Perrin all look promising too. It's still too early to call, but certainly going better than [_Cowboy Bebop_](/fragments/netflix-cowboy-bebop).

Something I found incredible is that Brandon Sanderson, the author who finished the series, is regularly active in the Wot subreddit, going as far as to create [fan posts deconstructing  each episode](https://www.reddit.com/r/WoT/comments/qy2r52/some_thoughts_from_brandon_episode_two/), and being refreshingly [candid about his opinion](https://www.reddit.com/r/Fantasy/comments/qwy6xu/wheel_of_time_megathread_episodes_1_3_discussion/hlhp4ea/?context=3) on some of the show's creative changes.

Until next week.

<img src="/photographs/nanoglyphs/029-path-of-madness/whitecloak@2x.jpg" alt="Wheel of Time: Whitecloak" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/029-path-of-madness/shadar-logoth@2x.jpg" alt="Wheel of Time: Shadar Logoth" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/029-path-of-madness/lan-and-nynaeve@2x.jpg" alt="Wheel of Time: Lan and Nynaeve" class="wide" loading="lazy">

---

[1] Actually, the rescue is probably a rewrite in Java.
