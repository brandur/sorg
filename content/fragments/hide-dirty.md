+++
hook = "A simple trick to measurably improve prompt performance in a large Git repository."
published_at = 2016-11-17T04:42:17Z
title = "Fixing a slow git $PS1 prompt"
+++

Like many other developers, I've configured my prompt's `$PS1` to display my
current branch in Git:

    st-brandur1 my-project(master*) $

This is accomplishing by invoking Oh-My-Zsh's `git_prompt_info` function. Bash
has a similar mechanism called `__git_ps1`.

This worked great for me for years until I started working in repositories with
very large Git histories. You can see it for yourself by cloning down the Linux
kernel tree, Rust, or Servo, which all demonstrate the effect.
`git_prompt_info` takes so long to return that the latency becomes disruptive
when typing commands in quick succession.

`git_prompt_info` calls into `git status`. Along with getting the current
branch name, it returns another piece of information: whether the repository is
"dirty" in that it has uncommitted changes. In the prompt line above, that
manifests as the `*` shown after the branch name. This dirty check is very slow
and was single-handedly responsible for my unresponsive prompt.

Luckily, Oh-My-Zsh is good enough to include an option to skip the dirty check.
It's set as part of Git configuration:

    $ git config --add oh-my-zsh.hide-dirty 1

Bash includes something very similar:

    $ git config bash.showDirtyState false

My problems with a slow prompt evaporated instantly.
