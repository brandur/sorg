---
title: Google Engineering Principles
published_at: 2016-02-03T19:37:54Z
---

Here's a good talk on engineering principles at Google entitled [_Engineering
for the Long Game_][youtube].

My takeaways (paraphrased by me):

* Build your infrastructure so that engineers are running aggregate services
  instead of single machines.
* Any kind of standardization that can be put in early will probably pay
  dividends as it amortizes over time (naming, configuration files,
  statistic/informational endpoints, etc.).
* Don't build a new system where an existing one will do (be skeptical on
  the "special requirements" that are being used as an excuse to bully the new
  system through).
* "Don't let the weeds grow higher than the garden." Instead of letting debt
  grow unchecked, perform constant maintenance on the sharper points that are
  slowing productivity.

[youtube]: https://www.youtube.com/watch?v=p0jGmgIrf_M
