+++
hook = "On how LLMs now drive most of this blog's workflows, replacing what used to be a lot of small, manual edits."
published_at = 2026-07-28T16:43:01-05:00
title = "I used to add links by hand"
+++

I was reflecting on how the rise of LLMs has changed my workflows when it comes to this blog. I continue to maintain a strong [certified human](/human) charter that guarantees all content here is written by a human hand, but I use LLMs frequently for automation:

> `give me an empty fragment template with the slug 'llm-workflows'`

Or:

> `add an atom entry that links the new article`

Or:

> `update my /now page with the poster image https://dropbox.com/image.jpg and content:`
>
> `lorem ipsum ...`

I don't have agent skills for any of these things because with a failure rate of ~0%, Codex does fine without them. I'm never compelled to ask it to do something for me the same way twice.

My favorite part about it is that Codex pretends that all my site's little idiosyncracies are serious, legitimate practices. When I tell it to make an [atom](/atoms) (a word I made up for a tweet-like snippet), it figures out exactly what TOML file those belong in and inserts it. When it promotes a draft by moving from `./fragment-drafts` to `./fragments`, it treats this entirely local convention as if it were as natural as breathing. When I ask it to write me a "hook" (a 1-2 sentence summary of an article), I don't have to ask twice. It finds the right spot in TOML frontmatter and throws one in there.

---

For years, Ruby on Rails has shipped with a [`bin/rails generate` command](https://guides.rubyonrails.org/command_line.html). I always took issue with it because it's one that you'd use only rarely (most of the time you're editing existing code), but to use it effectively you had to remember long combinations of obscure syntax:

``` sh
bin/rails generate resource post title:string body:text
```

`bin/rails generate` was an ingenious idea and unlocked superpowers if you knew it inside and out, but I'd always have to look up how to use it, which defeated the point.

LLMs have replaced `bin/rails generate` for me. No more subcommands or flags to remember -- natural human language becomes the user interface.

I imagine that when you read this years from now, this is all going to be so obvious that you'll wonder why I'm even talking about it. I want to remind you and myself that there was a point where we didn't have any of this stuff. If I wanted to add a link, I'd edit a file manually. For a new article, I'd `cp` an existing one to a new location and edit it. For a new photo, I'd paste in a URL where the CMS could fetch it and write in alt text manually.

Now, the machine does everything. Let these anecdotes from the past be a time capsule for the future.
