---
title: Understand Deeply
published_at: 2017-01-18T16:03:01Z
hook: Aim to be an engineer and not just a user by learning
  in depth.
---

One of the most notable parts about my arrival in the
Valley was the depth to which people were willing to go to
understand the technology they worked with [1]. Back home,
people were happy to learn the basic interfaces of the
programming languages they used, or the web server that
backed their production work, but wouldn't go beyond.

For a database, that might look like learning some SQL and
some basic administrative functions, but the underlying
operation of the product was an entirely inscrutable black
box. If anything went seriously wrong, you called Microsoft
enterprise support.

The best people are not like that at all. Along with deep
insight into advanced features of SQL and administration,
they know how to peer beyond the curtain. They can explain
its in-memory data structures, how it lays out information
on disk, or describe the precise mechanics of its
replication scheme.

People at home had been users, but I'd found new colleagues
who were true engineers; able to pop the hood and get their
hands dirty when the situation called for it.

For day-to-day work, users and engineers function about the
same. The value of technology products is that you _can_
treat them like a black box and maximize your productivity
with them by only thinking about their outward interface.

It's the times where you run into non-standard trouble that
the difference shows. My favorite story is where I ran into
[overflowing job queues at Heroku](/postgres-queues). The
manual couldn't help us, but a colleague of mine understood
the internals of Postgres so well that he was able to
reason about what was happening just by thinking about it.
We later verified his hypothesis empirically.

My approach to learning new technology has always been to
install it, and bounce around between documentation and
Stack Overflow answers until I built up a feel for it by
rote trial and error. These days, I try to build my muscles
slowly, but correctly, by attacking the learning from first
principles. My latest endeavor is with Rust; it's possible
to write it by copying and pasting examples, but you'll be
fighting the borrow checker in perpetuity unless you really
understand what it's trying to do.

Aim to be an engineer and not just a user. Start with the
manual, and aim not just to understand, but to understand
deeply.

[1] In retrospect, this turned out to be an especially
    prevalent trait among early Heroku engineers, but I use
    it to describe the Valley because in my experience,
    you'll find more people with it here than elsewhere.
