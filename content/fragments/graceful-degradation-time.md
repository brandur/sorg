+++
hook = "How to format blog content so it'll last."
published_at = 2021-02-08T16:21:45Z
title = "Markdown and Git: Graceful degradation through time"
+++

_(Problem 1.)_ As I was [reviving my RSS habit the other week](/fragments/rss-abandon), I couldn't help but notice how many of my old subscriptions were dead. Not "doesn't publish anymore" dead, but "gone from the internet" dead. People had created blogs (many with great content), run them for a while, let them lapse, and eventually they were swallowed by time and entropy.

_(Problem 2.)_ The Wayback Machine is very good, and one of the great assets of the internet. But it's not perfect. It's fairly slow, and is often missing assets like images or entire pages (especially where smaller blogs are concerned). Most importantly, its content isn't discoverable -- it's difficult to find things without the precise original URL.

_(Problem 3.)_ For fun, I tried "de-archiving" an old colleague's blog from The Wayback Machine. Getting the first few pages was easy, but getting the whole thing, and with quality/precision, was very hard. First, I needed to write a spider to iterate the paginated archive. Then, an extractor to boil away the horrible Wordpress-generated, tracker-ridden HTML to get down to the actual content. Then, another crawler to download whatever images were still available.

## Run optimistically, plan pessimistically (#plan-pessimistically)

A simple proposal I'd like to make to current and future blog authors: run your blog in _whatever_ technology you want, whether its a static generator like Hugo or Gatsby, a custom CMS, or the JAM stack. Different technologies are good for different things, people are good with different tools, and there surely are plenty of great options.

_But_ remember at the end of the day that a blog is text and multimedia. Store your content in the simplest formats in which it can possibly be stored -- I'd suggest Markdown and Git -- and use the great frontend to read from that and render it however you want. Make that source public via GitHub or another favorite long-lived, erosion-resistant host. Git's portable, so copy or move repositories as you go.

It's not going to matter for a month, or a year, or probably five years, but once it reaches that horizon of 10+ years where things start disappearing, it might really make a difference. Interested parties can still see read content even after the original host is long gone. No one plans for their site to disappear, but most eventually do.

I practice what I preach. This entire site is generated from Markdown, and also [lives on GitHub](https://github.com/brandur/sorg). It could be remixed even after it's gone.

{{Figure "Markdown and Git." (ImgSrcAndAltAndClass "/photographs/fragments/graceful-degradation-time/markdown-and-git.png" "Markdown and Git." "overflowing")}}
