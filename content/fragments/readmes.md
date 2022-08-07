+++
hook = "An ode to well-written READMEs. Detailed, but not overly so, and scannable with examples."
published_at = 2022-08-07T21:02:07Z
title = "READMEs are a great idea, still"
+++

One of the great tech ideas of the 2010s that's always stuck with me is [Readme-driven development](https://tom.preston-werner.com/2010/08/23/readme-driven-development.html).

This was written contemporaneously with GitHub's innovation to start rendering READMEs in Markdown as stylish HTML on the web. This seems completely obvious today, but was a totally exotic feature at the time, and helped a lot in making READMEs more readable and more accessible.

In what's hopefully not news to anyone, a good README is still a major asset for any project. A well-written one shares a lot in common with a good novel: it's got an initial hook, mounting exposition as more and more detail is revealed, and concludes with some obligatory detail like attributions and licensing.

[Pressly's fork of Goose](https://github.com/pressly/goose/tree/0a729707373536519a3aef3d3a4d0c4d4ced5766) is an example of an excellent one (and one that I happened to be reading this morning). It starts with background, moves onto installation and basic usage, gets into more detail and advanced usage, then moves onto some best practices and philosophy ("hybrid versioning") before concluding with credits and license. With every step there's succinct CLI or code samples. Sentences are short. Paragraphs are often only one line, with the longest being four, keeping the file easy to visually scan in a hurry.

As a potential user, this is **awesome**. It's like a one-stop shop for the 80% of detail I'm going to need to integrate it, and 80% of the reference I'm going to need afterwards. By the end, and without having written a line of code, I not only have a clear idea on how it works, but how I'd bake it into my own projects. I don't have any major outstanding questions.

As good of a project as Postgres is, [it can be used](https://github.com/postgres/postgres/tree/7e29a79a46d30dc236d097825ab849158929d977) as an example of a near perfect foil. No usage information, examples, or detail, with an unlinked reference to an `INSTALL` file that's not even in the same directory. This type of thing was pretty common pre-2010s when clicking into a README had to be a conscious act, before GitHub effectively made them a project's homepage.

## Not too ancient and not too long (#too-ancient-too-long)

An unfortunate tendency is for READMEs to ossify. Someone throws something together in the beginning (often without thinking about it too much) and new sections are added to it incrementally, but at no point has anyone gone back to reevaluate the totality of what's in there and whether it makes sense. I'm guilty of this on a number of my projects.

It's also possible for a README to be doing too much. A README that's too long or goes into overextended detail is hard to parse for an introductory user, and hard to scan for an existing one. There isn't a great heuristic for when "good and thorough" transitions to "just too much", but sufficed to say that sections that are too long or infrequently used should be broken out to their own documentation file.

Lastly, I'll just say that spending the time to write a good README is partly a selfish act. For a lot of my own projects that I haven't looked at in a while, I've forgotten almost everything about their structure, and don't even remember basic details like how to run their test suite. Their READMEs are like a time capsule to myself, reminding me how to go about maintaining the thing.
