+++
hook = "Engineering blog posts are at least partly a marketing tool. Exercise some skepticism (but not too much)."
published_at = 2021-02-12T16:03:25Z
title = "Engineering blog posts: Apply healthy skepticism"
+++

Engineering blogging's been a foundation cornerstone of modern software shops for more than a decade now, with majors like GitHub, Stripe, Dropbox, and Shopify diligently maintaining well-curated and well-produced blogs, partly to scrupulously provide a learning resource, partly to spread bread content, and partly as a tool for marketing developer products and to developers for recruiting.

I love engineering blogs. I've read many hundreds, and probably more like thousands. They're not only useful for learning, but also for catching a rare glimpse of insight into how things are done outside the confines of your own box. The state of the art would unquestionably be further behind today if the industry had stuck with the old world Oracle/Microsoft/Apple "closed doors' approach.

But one word of advice around blogging though: practice healthy skepticism. Remember that engineering posts are often a marketing tool, and even where they're not, authors generally don't want to get into the gritty shortcomings of the technologies and techniques they're talking about. When it comes to software, it's trade offs all the way down. Maybe not _always_ always, but often enough to effectively be.

It's a high-pass filter -- the good stuff gets through, the bad is left out:

{{Figure "Only the good makes the blog cut." (ImgSrcAndAltAndClass "/assets/images/fragments/engineering-skepticism/high-pass-filter.svg" "Only the good makes the blog cut." "overflowing")}}

This applies to talk about microservices, new web frameworks, monorepos, test-driven development, agile methodology, programming languages -- you name it. The cause stems from enthusiasm rather than conscious intention to mislead. I aim to be as honest as possible in my posts, but am still guilty of it.

Like with many things, skepticism is best practiced in moderation. Believe what you read, but also think about what wasn't written.

Aspiring bloggers: a type of post I've always wanted to see more of is the "downsides of X". Like after a post like "why we love monorepos" a companion post like "why we _don't_ love monorepos". The dualistic light _and_ dark aspects. More of these would help lead to more informed decisions by younger projects and companies.
