+++
hook = "UNWRITTEN. This should not appear on the front page."
published_at = 2026-07-28T15:48:15-05:00
title = "I used to add links by hand"
+++

I was reflecting on how the rise of LLMs have changed my workflows when it comes to this blog. I continue to maintain a strong [certified human](/human) charter that guarantees all content written here is written by a human hand, but I use LLMs frequently for automation:

> `give me an empty fragment template with the slug 'llm-workflows'`

Or:

> `add an atom entry that links the new article`

Or:

> `update my /now page with the poster image https://dropbox.com/image.jpg and content:`
>
> `lorem ipsum ...`

I don't have agent skills for any of these things because with a failure rate of ~0%, Codex does fine without them. I'm never compelled to ask it to do something for me the same way twice.

For years, Ruby on Rails has shipped with a [`bin/rails generate` command](https://guides.rubyonrails.org/command_line.html). I always took issue with it because it's one that you'd use only rarely (most of the time you're editing existing code), but which to use effectively you had to remember long combinations of obscure syntax:

``` sh
bin/rails generate resource post title:string body:text
```

`bin/rails generate` was an ingenius idea, and unlocked superpowers if you knew it inside and out, but I'd always have to look up how to use it, which defeated the point.

LLMs have replaced `bin/rails generate` for me. No more subcommands or flags to remember -- natural human language becomes the user interface.

I imagine that when you read this years from now, this is all going to be so obvious that you'll wonder why I'm even talking about it. I want to remind you and myself that there was a point where we didn't have any of this stuff. If I wanted to a link, I'd edit a file manually. A new article, I'd `cp` an existing one to a new location and edit it. A new photo, paste in a URL where the CMS could fetch it and write in alt text manually.

Now, the machine does everything. Let these anecdotes from the past be a time capsule into the future.
