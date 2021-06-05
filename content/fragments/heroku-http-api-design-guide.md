+++
hook = "A few reflections on the Heroku HTTP API Design Guide, and a link to a fork that brings it back to its original format."
published_at = 2021-06-05T21:06:04Z
title = "The Heroku HTTP API Design Guide, revisited"
+++

One of our most influential "late Heroku" thought artifacts was "The Heroku HTTP API Design Guide". We'd recently revamped our API with a new V3 iteration, and the guide was a terse description of the conventions we'd developed. It was written by [Mark McGranaghan](https://twitter.com/mmcgrana), and then entered common maintenance.

It blew up when it was dropped, accumulating 13.5k stars on GitHub, and translated into a number of foreign languages. It was a big deal for us at the time because we hadn't managed to produce all that much on the "thought leadership" front since the company's halcyon days of ~2011.

The project was eventually converted to GitBook format, which I thought was a little unfortunate because it lost a lot of that original succinctness. I've created [a fork of the project](https://github.com/brandur/heroku-http-api-design) to bring it back to its more easily digestible original state.

Reading back through it today, it stands up pretty well. There are a few suggestions that were posited without enough convincing rationale (e.g. use of V4 UUIDs, which I was partly complicit in proposing), but it's not a huge problem. One of my strongest beliefs when it comes to API design is that it's most important to establish _some_ convention, because without any, things start drifting all over the place very quickly. A perfectly reasonable approach would be to copy a lot of the guide, change the sections you disagree with, and then adopt that for internal guidance.

The guide's most important non-obvious innovation is its brevity. It's written in simple, candid language that makes the entire thing easy to consume in just a few minutes. At Stripe (as a counterexample), we had at least one document that attempted to state design principles, but it was lengthy and unfocused, which made it rarely referenced in practice. Having something in the form of a written artifact was good, but it was pretty uninspiring to read.

This came to mind recently because I'm looking at writing something similar for use at Crunchy, and I'm currently looking copying out large parts of the Heroku guide for reuse. More on this as the project gets further along.
