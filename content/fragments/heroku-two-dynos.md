+++
hook = "Contemplating whether the Heroku platform would fit on two standard 512 MB dynos if it could be ported from Ruby to Go."
published_at = 2024-05-11T11:04:23+02:00
title = "Heroku on two standard dynos?"
+++

One of the aspects that I appreciate most about Go is its incredible economy around CPU and memory.

Bridge's API runs on two standard 512 MB dynos. Even with these very modest containers, only a fraction of available resources are used, with memory hovering around 40 to 50 MB, level as a laser beam.

{{FigureSingle "" "/photographs/fragments/heroku-two-dynos/bridge-memory.png"}}

It actually runs quite well on only one dyno too. Having two of them is an availability hedge to protect against Heroku itself. We'd been running one, but doubled them after a Heroku incident where single dyno apps became randomly unavailable after being deployed.

Admittedly, it's certainly not the world's highest volume API, but it does serve every customer request regardless of whether they came in by API, CLI, or GUI, and receives tens of thousands of metric data points from Postgres servers every hour. Let's say it's hundreds of requests per minute.

Thinking back to the time I was running the Heroku API ~2011 to 2015, that's about the right order of magnitude for it as well. The API acted as Heroku's orchestration layer, serving every CLI and Dashboard request, and serving as the go between for the platform's constellation of microservices. The deployment specifics are foggy, but it ran on about a dozen large EC2 instances, eating memory for breakfast through the process forking model typically used to achieve real parallelism in the presence of Ruby's GIL, first on Unicorn and later Puma.

There was already enough code at the time that a rewrite from Ruby to Go was infeasible (we tried), but it makes me think now. If we could've waved that magic wand and convert to Go and had reasonable techniques for good data loading economy similar to what we have today, I think the entire Heroku API service could've run comfortably on two standard dynos with breathing room to spare?

A reminder that despite the lofty language of tech founders and VCs, at the end of the day most of us are just serving HTTP requests. Unfortunately there's no way for me to prove this counterfactual, but it's a fun thought experiment nonetheless.