+++
hook = "Surviving the Herokupocalypse by jettisoning dynamicism."
published_at = 2023-01-01T13:00:35-07:00
title = "The static rip cord"
+++

November 28th 2022 was the Herokupocalypse, where the platform engaged in its most aggressive cost-cutting project to date by deleting free apps and free databases.

Always leaving things until the last minute, the day before I cycled through the apps in my account to see whether I needed to save anything. Luckily, there wasn't much. I had ~50 apps, but precious few that were doing anything, strongly suggesting the tendency to use in abundance where there is abundance, regardless of need.

The couple apps I rescued were some slides from ancient talks like [Composable](https://brandur.github.io/composable/) from Frozen Rails in 2012 [1].

I initially tried to boot them up locally from the source in GitHub, but as you might expect from Ruby apps that are 10 years old, this proved challenging. I had a lot of good-ideas-that-turned-out-to-be-bad-ideas back then such as bundling in elaborate asset pipelines like Sprockets that'd pull in a whole Node runtime to perform dark feats with opaque magic, and even after upgrading every gem in the project it still wouldn't run. I timeboxed the modernization effort to ten minutes and gave up.

Luckily Heroku's whole [erosion resistance](/nanoglyphs/022-entropy#erosion-resistance) thing actually works pretty well, and I could still get to a runnable version of the projects despite them being roughly as old as my entire professional career. So instead of trying to get the dynamic element running again and finding a new container host (e.g. Google Cloud Run), I took the easy path: the static rip cord.

* In Chrome, File -> Save Page As.

* Massage the generated source to make paths to assets like images and fonts prettier, and agnostic as to whether you're opening the page up locally or from a remote host.

* Commit the product to Git, and add a GitHub Actions workflow to [auto-deploy to GitHub Pages on changes](https://github.com/brandur/composable/blob/master/.github/workflows/github-pages.yml).

And voila! Saved from Herokupocalypse, obsolescence-prone dynamic scaffolding relegated to a note in Git history, and now in a format that's not only easy to open locally with only a web browser as a dependency, but interchangeable between web hosts (in case gods forbid, free GitHub Pages ever goes the way of free Heroku). In retrospect, I'm actually thankful for the Heroku free tier going away because it acted as a forcing function for some much-needed housekeeping.

Obviously this won't work for any app that needs real dynamic elements, but a lot of things don't. For content that doesn't, consider publishing it in a format that not only works now, but will work 10 or 20 years from now, namely Markdown or HTML in Git. See [degrade gracefully through time](/fragments/graceful-degradation-time).

[1] Rescued for posterity, but I wouldn't recommend trying to navigate this mess, especially on mobile.