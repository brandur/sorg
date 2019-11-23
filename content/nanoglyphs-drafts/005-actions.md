+++
published_at = 2019-11-22T19:30:13Z
title = "Actions"
+++

![Sumo strong eggnog](/assets/images/nanoglyphs/005-actions/eggnog@2x.jpg)

Hello.

---

## Migrating to Actions (#github-actions)

Over the years, Travis has been one of the most important services in my quiver.

Ever since Travis untimely [acquisition by a holding company](https://news.ycombinator.com/item?id=18978251) and the departure of a large number of its engineering staff, a number of us have been keeping an eye out for what to use next. For the time being Travis is still running, but the clock’s likely inching closer to midnight.

[GitHub Actions](https://github.com/features/actions) were an extremely timely addition to the company’s product surface. Although described in most places in such grandiose vernacular so as to obfuscate what it actually does, to put it simply, it allows you describe jobs that will run on certain repository events like a push, opened pull request, or cron — perfect for CI. It’s major differentiating feature is that the steps of a job can be defined as similar shell commands, they can also defined a Docker container to run. This makes the whole system modular — units of work of arbitrary size can live in their own project, providing good encapsulation and easy updates.

