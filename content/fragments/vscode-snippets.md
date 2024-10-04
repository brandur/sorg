+++
hook = "Using built-in variables in VSCode snippets to make publishing to this site incrementally faster."
published_at = 2024-10-04T11:18:21-07:00
title = "TIL: Variables in custom VSCode snippets"
+++

This blog is entirely driven by Markdown, TOML, and Git. Publishing an [atom](/atoms) or [sequence](/sequences) involves popping open a TOML file, adding a new item to the top, committing to Git, and pushing to origin to trigger a CI action that deploys the site:

``` toml
[[atoms]]
  published_at = 2024-10-04T10:24:22-07:00
  description = """\
Hello, world!
"""
```

This generally works quite well, and in this developer's humble opinion, far preferable to something involving a web UI with a little text box, but when I'm being honest with myself, I have to admit that the friction to editing is a little too high, and prevents me from publishing posts that I would've done if I was on a platform _with_ a web UI and a little text box, like Twitter.

I'd been using [VSCode snippets](https://code.visualstudio.com/docs/editor/userdefinedsnippets) to speed up inserting a new TOML item, but the `published_at` date wasn't automated, so I'd have to jump to a terminal, get a timestamp with `date`, then jump back and paste it. Not a big deal, but a little slow and mildly annoying.

I went back and RTFMed. It turns out that custom snippets support a number of built-in variables like `$TM_FILENAME`, `$CURRENT_SECONDS_UNIX`, or even `$UUID` for a random V4 UUID.

With a few more variables I got it to insert RFC3339 dates exactly like the ones I'd been grabbing from my terminal:

``` json
{
	"New atom": {
		"prefix": "at",
		"body": [
			"",
			"[[atoms]]",
			"  published_at = $CURRENT_YEAR-$CURRENT_MONTH-${CURRENT_DATE}T$CURRENT_HOUR:$CURRENT_MINUTE:$CURRENT_SECOND$CURRENT_TIMEZONE_OFFSET",
			"  description = \"\"\"\\",
			"$1",
			"\"\"\"",
			""
		],
		"description": "New atom"
	}
}
```

There's quite a few other useful built-ins (e.g. currently selected text, contents of clipboard, start comment), and [transformations with regex](https://code.visualstudio.com/docs/editor/userdefinedsnippets#_transform-examples) are supported.

I also took the time to get the whitespace around the inserted block exactly right, so no extra time is needed to correct it after insertion. All in all I probably saved myself about ten seconds for each snippet use, but it's enough of a gain to make myself marginally more likely to do it.

Next up (hopefully): a mobile publishing workflow, something that's been sorely missing for years.