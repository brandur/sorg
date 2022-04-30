+++
hook = "Safely refactoring Ruby traditionally involves many incremental steps and feature flags. Sorbet and 100% branch coverage make it possible to refactor a lot of code safely."
published_at = 2022-04-25T18:11:13Z
title = "Sorbet + 100% cov makes Ruby refactoring possible"
+++

Last week, we deployed a refactoring patch to our Ruby codebase that was big for us [1] -- 178 files changed:

{{FigureSingleWithClass "Changeset." "/photographs/fragments/large-scale-ruby-refactoring/pr.png" "overflowing"}}

If you'd asked me five years ago how to safely deploy a change of this size in Ruby, I'd have a very simple answer: you can't.

Ruby's parser is no help in detecting regressions -- it'll notice invalid syntax, and that's about it. A good test suite helps, but it's never complete enough -- the happy path and a few error cases are tested, but in production all kinds of alternate data tends to be flowing through. A similar story for staging environments -- activity there isn't of high enough fidelity to catch problems. Even a test suite with 100% line coverage (a rare thing anyway) is no sure fire bet because even covered lines may still have alternate branches which aren't.

So traditionally, big changes in Ruby are pulled off very slowly and very incrementally -- small sections are refactored individually in a hundred steps. Extra careful companies will use flags so that both new code and old code coexists, allowing fast roll back in case of error, and new code to be vetted in a minimally destructive way by taking only a small percent of total traffic. It's hugely time consuming. Very often, Ruby code is just _not_ refactored -- it's too much effort with too much risk for too marginal of a benefit.

## Strict typing (#strict-typing)

[Sorbet](https://sorbet.org/) does a lot to address this. All non-spec files in our codebase are annotated with `typed: strict`, which requires type signatures on all methods and types assigned to all constants and instance variables. This makes static analysis very effective, and during a big refactor, the vast majority of problems can be caught that way.

## Total coverage (#total-coverage)

We're also doing something a little more controversial by requiring not only have 100% line coverage, but 100% _branch_ coverage too. Some old files from the first big development push are exempted from full branch coverage, but all new development has it, and we're back patching existing code, with current coverage sitting at 95+%.

I'll tell you first hand that 100% branch coverage is pretty annoying -- you often end up contorting yourself to write tests for vanishingly unlikely branches. But you remember why you're doing it while refactoring, where it's a big comfort. It also has a compounding effect with Sorbet because along with static analysis, Sorbet also has a runtime component that checks that objects passed into methods are the type they're supposed to be. So if all code paths are covered, every method is robustly called with every object it'll be called with.

By our best measurements, only one problem fell out of the refactor (some network code that relied on heavy stubbing in tests), and it had no production effect thanks to our [transactionally-driven state machines](https://www.citusdata.com/blog/2016/08/12/state-machines-to-run-databases/) [2]. Sending it out took about a day and a half, but most of that was letting it idle in staging so we could keep an eye out for problems.

So as usual, consider not writing Ruby, but if you do, use type signatures, and maybe 100% branch coverage too, as annoying as it seems.

[1] It's all relative. 178 files is small for a big codebase, but is almost every file touched in ours.

[2] Also, we caught it in staging, but we'd generally prefer not to break that either.
