+++
hook = "A rare, periodic reevaluation of one of the oldest and dearest tools in my kit."
published_at = 2026-09-17T16:12:24+08:00
title = "15 years later, a day without Tmux"
+++

A few years ago I switched from doing all my coding in command-line Vim to VSCode. As much as I like [the terminal as a computer interface](/interfaces), I had it in the back of my head at the time that it was the beginning of the end. Every day, multimedia was getting more important, and terminals didn't seem to be evolving to meet the challenge.

So color me surprised to find that in 2026, I'm using the terminal more than ever. All that time spent writing code in VSCode is now in a Codex TUI instead. Most of the questions I used to ask Google are sent to Codex. Ghostty is one of two apps on my computer (along with a web browser) that are always open.

Yesterday, I did something else I wouldn't have predicted -- I tried running my sessions without Tmux, a package I've been using daily for 15+ years.

The most proximate reason was that I'd hit yet another small snag with Tmux sitting as a layer between my terminal and shell. A remote Tmux session enabled xterm's `modifyOtherKeys` in my local Tmux pane over SSH. If SSH exited normally, it'd be reset and I'd never notice, but if SSH timed out, the local pane would be left in extended-key mode, making shortcuts like `Ctrl+R` or `Ctrl+D` manifest as escape sequence jumbles like `;5;114~`. Fixable, but it was the fourth or fifth time I'd run into one of these papercuts. It was this one that planted the question in my head: wait, is this worth it?

So I ran a little experiment: go to my terminal direct. Skip Tmux for the day. Try Ghostty's built-in tabs and panes. Within minutes I discovered that during the last 6+ months, the Codex TUI has been giving me clickable hyperlinks in its answers through the **OSC 8 hyperlinks** specification. I'd never noticed because Tmux wasn't passing them through. I clicked one and my jaw dropped. So simple, but so powerful. (This problem is also fixable, but it's another incompatibility in the bundle.)

I want to make it clear: I am not disparaging Tmux at all. The lifetime value I've gotten out of this project is incalculable, and I'll continue to be a regular user for remote sessions. It's important context that back when I started using Tmux (and dinosaurs roamed the Earth), terminals didn't have tabs, they most definitely didn't have panes, and weren't very configurable. A terminal multiplexer wasn't a nice-to-have -- it was a necessity. But modern terminals have all these things, so Tmux's advantage has shrunk over time.

So what's replaced Tmux? Nothing yet. I'm using Ghostty's tabs/panes and a few shortcuts to make them more ergonomic:

``` toml
keybind = cmd+left=previous_tab
keybind = cmd+right=next_tab
```

``` toml
keybind = ctrl+h=goto_split:left
keybind = ctrl+j=goto_split:bottom
keybind = ctrl+k=goto_split:top
keybind = ctrl+l=goto_split:right
```

It's not a perfect replacement. Persistent sessions are gone (for now, anyway), but I'm very much enjoying my clickable links, faster shortcuts (no prefix), native scrollback, and a command palette for those actions that I use infrequently enough not to have bound a hotkey for.

It's amazing how active this space suddenly is again. It wasn't too long ago that we got Ghostty, but more recently projects like Herdr and Superlogical have popped up too. It's been a long time since I did a good tool sharpening in this area. I'm experimenting.
