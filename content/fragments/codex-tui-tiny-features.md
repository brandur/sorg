+++
hook = "Hiding in plain sight: copy the last answer, toggle raw Markdown, and disable Escape as an interrupt."
published_at = 2026-09-17T18:17:26+08:00
title = "A couple tiny Codex TUI features"
+++

It occurred to me today that I now spend more time in the Codex TUI than I do in Vim. Unbelievable.

I've spent a lot of time over the last couple days prompting Codex TUI to help me resolve little quality-of-life problems that I have with Codex TUI. This type of Codex inception works wonderfully.

Nearly every time I do it, there's a headslap moment. What I want to do isn't only possible; it's often available by default without custom configuration. Why, oh why, didn't I look this up earlier?! I figured I'd dump a few recent favorites in case they help anyone else.

## Copy last answer to clipboard (#copy-last-answer)

Codex TUI has a built-in shortcut for copying the last answer to clipboard: `Ctrl+O`.

Combine this with an alias to put it into Vim for easy pinpoint copying and remixing (`-R` is for read-only, so you're not warned about unsaved changes on exit):

``` bash
alias vio='pbpaste | nvim -R -'
```

## Get raw Markdown (#raw-markdown)

Codex renders answers as a terminal-friendly form of Markdown, which is usually what you want because it looks nicer. However, you occasionally want the raw Markdown to copy something like a table or code block. There's another built-in for this: `Alt-R`.

It toggles raw mode. `Alt-R` again toggles it back.

(`Ctrl+O` above also copies raw Markdown, but sometimes it's faster to use `Alt-R` if you just want to select one segment.)

## Disable Escape as interrupt (#disable-escape-interrupt)

For more than a decade I've had my Caps Lock remapped to Escape so I can use it as an easy Tmux prefix. More recently, I've been using [Tmux less](/fragments/tmux-15-years), but I still hit Caps Lock constantly from muscle memory, and when Codex TUI is active, this has the undesirable effect of Escape cancelling the current operation.

Escape is never what I want to use to cancel an operation anyway. [Just like in Vim](/fragments/no-escape), `Ctrl+C` is the more keyboardist-friendly shortcut and works in Codex TUI too.

This addition to `~/.codex/config.toml` will disable Escape but leave `Ctrl+C` working:

``` toml
[tui.keymap.chat]
interrupt_turn = []
```
