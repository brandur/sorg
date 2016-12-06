---
title: No Escape
published_at: 2016-11-19T04:36:05Z
hook: Curbing the use of the Escape key in Vim.
---

A few weeks ago Apple announced their new MacBook Pro. To make way for their
latest gimmick (the TouchBar) they removed the row of the keyboard that
contained the Esc and function keys.

I'll do what I can to avoid spending money on this sort of thing lest I further
encourage more backwards progress on Apple's part, but just in case they've
started a trend, I've decided to hedge my bets.

The use of Esc in Vim has always been considered a bit of an anti-pattern due
to the travel distance required off of the home row. Despite that, I continued
to use it anyway because the recommended alternative of Ctrl+[ is awkward,
especially on a Mac with its diminished Ctrl key. There's also something to be
said for the satisfying feeling of reaching up and banging the Esc key.

I decided to make the transition, but to use Ctrl+C instead of Ctrl+[, which is
also a shortcut that's enabled by default. The first problem I ran into was
that despite knowing what key I _wanted_ to hit, muscle memory would always
take me back to Esc. A quick Google search revealed a good solution to that
problem:

    inoremap <esc> NO ESCAPE FOR YOU

Now intead of taking you out of Insert mode, Esc will now inject an unhelpful
string of junk which must then be painstakingly erased. I've already had to do
it about twenty times writing this article. Now it's all up to me: sink or
swim.

One annoying thing about Ctrl+C is that when you accidentally use it from
Normal mode, it shows you this beginner-friendly message:

    Type  :quit<Enter>  to exit Vim

Having learnt this the hard way the first time I opened Vim about ten years
ago, I don't really need to see it anymore. Here's a binding that will stop it
from coming up:

    nnoremap <C-c> <silent> <C-c>
