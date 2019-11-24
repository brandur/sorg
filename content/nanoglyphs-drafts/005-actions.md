+++
published_at = 2019-11-23T02:57:46Z
title = "Actions"
+++

![Sumo strong eggnog](/assets/images/nanoglyphs/005-actions/eggnog@2x.jpg)

Welcome to Nanoglyph, a newsletter about software. It ships weekly, though like most software projects, deadlines are not always met.

This week involved the fulfillment of an old Heroku tradition by making a batch of [sumo strong eggnog](https://github.com/seaofclouds/sumostrong/blob/master/views/eggnog.md). The recipe’s convenient in that you can buy the heavy cream and whole milk involved in cartons of exactly the right size so there’s none leftover. It makes enough to nearly fill three [Kilner clip top bottles](https://www.amazon.com/Kilner-Square-Clip-Bottle-34-Fl/dp/B005N984I8/). If the idea of strong eggnog is even remotely appealing, it’s very much recommended.

The usual format is three links from the week with some commentary, but to keep things dynamic, I’m playing with the format to instead talk about a mini-project in a little more length — like a push version of a blog post.

---

## Migrating to Actions (#github-actions)

Over the years, Travis has been one of the most important services in my quiver.

Ever since Travis untimely [acquisition by a holding company](https://news.ycombinator.com/item?id=18978251) and the departure of a large number of its engineering staff, a number of us have been keeping an eye out for what to use next. For the time being Travis is still running, but the clock’s likely inching closer to midnight.

[GitHub Actions](https://github.com/features/actions) were an extremely timely addition to the company’s product surface. Although described in most places in such grandiose vernacular so as to obfuscate what it actually does, to put it simply, it allows you describe jobs that will run on certain repository events like a push, opened pull request, or cron — perfect for CI. It’s major differentiating feature is that the steps of a job can be defined as similar shell commands, they can also defined a Docker container to run. This makes the whole system modular — units of work of arbitrary size can live in their own project, providing good encapsulation and easy updates.

