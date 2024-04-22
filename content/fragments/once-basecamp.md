+++
hook = "ONCE's $299 self-hosted Campfire. Is a web app for chat fine nowadays?"
published_at = 2024-01-19T12:29:25-08:00
title = "Thoughts on ONCE + Campfire"
+++

37signals introduces its first product for [ONCE](https://once.com/), a [$299 self-hosted version of Campfire](https://twitter.com/jasonfried/status/1748097864625205586).

We used Campfire at Heroku in 2011. I never particularly liked it. It felt slow and had a lot of rough edges. We moved to HipChat briefly, which might've been slightly better at the time, but also nothing to write home about. Then we switched to Slack, and it felt like the sky opened up and a ray from heaven beamed down upon us. Truly a quantum leap over its predecessors. So fast. So responsive. A beautiful native app to use.

But a lot's changed since then. Slack became an Electron app, which is a web browser in a native-looking container, but whose disguise is thinly veiled. It got slower and less efficient -- messages sometimes take 15s+ to sync when I open the lid of my laptop; I have trouble sending a message from my phone if my cellular connection is anything less than perfectly ideal. Most recently they introduced a full redesign, optimized to look nice in screenshots, but degrading UX considerably [1].

Meanwhile, the web's caught up. React's sub-DOM rendering paradigm was invented and has become widespread, making web UIs noticeably more responsive where applied. A similar story for HTTP/2. Browsers and JavaScript engines became more optimized. Web notifications became possible.

In short, the benefits of products like Slack are more marginal, and the drawbacks of web apps less severe. Ten years ago when Slack came out, I would've said going back to a web app for chat was crazy. Not so much anymore.

I'll be curious to hear other people's experiences with ONCE's Campfire. I'm hopeful, although I think the cost of operations are being minimized. People are expensive, and if you have even one person who ends up spending a lot of time installing, upgrading, carrying a pager, and debugging, your savings from avoiding SaaS are wiped clean.

**Update 2024/01/28:** This piece originally said "Basecamp" instead of "Campfire" because honestly, I still get the two confused. Thanks Mathias for the correction!

[1] e.g. Workspaces now hidden behind a single tile, forcing the user to memorize their order to switch between them without an additional click. e.g. Unnecessarily breaking the main pane into "DMs" and "Activity", requiring frequent tab changes that weren't necessary before, and with poor shortcuts (partly Apple's fault for making the `Ctrl` button on Apple laptops so awful to use).