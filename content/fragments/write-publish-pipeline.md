+++
hook = "A writing pipeline that I find works pretty well."
published_at = 2021-09-09T14:12:09Z
title = "Write -> publish pipeline"
+++

I nail out first drafts of everything in [iA Writer](https://ia.net/writer). I normally prefer a modal editor like Vim for anything text-related, but for drafts I'm just aiming to get words out on paper, and the fact that editing is slower in a non-modal editor acts as a slight discouragement in doing it, which in this case, is good.

iA Writer's approach to markdown is _perfect_. It highlights, links, and formats, but unlike the vast majority of markdown-compatible editors (e.g. Slack, Dropbox Paper, Linear), it does so without removing the original syntax glyphs. Text is still markdown when you want to copy it out, and in practice bugs and annoyances are vastly reduced (the other day in Linear I couldn't end a line with a colon because its Clippy-inspired editor wanted to force me to pick an emoticon to go with it).

I usually write on Mac, but sometimes I'll use iA Writer on an iPad with Magic Keyboard in a gambit to further reduce distractions. Switching between apps on Apple's mobile operating systems is slow as molasses, which disincentives it, and helps with focus.

{{FigureSingleWithClass "A screenshot of iA Writer." "/photographs/fragments/write-publish-pipeline/ia-writer-2.png" "overflowing"}}

The raw draft goes into Vim for editing and refinement. You can't beat a modal editor for speed and precision in amending text. Words, sentences, and paragraphs are sliced, diced, and flung around the document with unrivaled economy. All edit passes will be in Vim until it's finished. I should do [three or four of them](https://www.newyorker.com/magazine/2013/04/29/draft-no-4), but rarely do except for more elaborate pieces.

For email pushes where you only get one shot, I'll usually ask someone like [Caio](https://twitter.com/kch) take a proofreading pass. I seem to be mildly dyslexic because even after multiple self reads, I'll still occasionally have typos and grammatical mistakes. Non-emails like this one all have links at the bottom of page where I rely on the kindness of strangers to send me pull requests will corrections for mistakes that they notice.

All publishing is [automatic on Git push](/aws-intrinsic-static). I have separate drafts folders for pieces that I want to keep around in the repository on the back burner without sending live. Images go in Dropbox so that I don't have to put them in the Git repository, and are slurped down based on directives [in a TOML file](/fragments/static-site-asset-management) so that they're resized and optimized for size automatically.

Sometimes I step back and think, despite all the problems we have in the 21st century, damn, do we have cool tools or what.
