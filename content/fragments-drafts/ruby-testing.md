---
title: Ruby Testing
published_at: 2016-10-19T01:45:15Z
hook: UNWRITTEN. This should not appear on the front page.
---

Writing Ruby code with no tests is as safe as jumping out an airplane with a
parachute packed by your untrained, blind, and drunk uncle because you hopped
off a bench a couple times down in the hangar and the weight of it felt roughly
right.

By default, productionizing Ruby is never safe. Not for
you, but _definitely_ not for that poor sap who's next in
line for maintainership. Know how hard it is to make sense
of your own code a year after you've written it? It's 10x
harder for someone else. Tests are absolutely the only way
to codify that certain behaviors are expected and not
accidental.

Besides a few basic checks on syntax, Ruby gives you zero guarantees. Code in
all languages needs testing, but code in dynamic/interpreted languages [1]
needs it doubly so.

Write tests for your production apps, write tests for your scripts, write tests
for your migrations; even write tests for that joke app that you built that one
time.

This sounds like a pointless rant. Everyone likes tests. "Testing is good" is
the platitude of our generation. However, I'm constantly floored by how many
people still do this and who should know better, even when those same people
might have just finished espousing the importance of testing to someone else
moments before.

[1] Compilers are better, but not foolproof. If you're a JavaScript programmer
    that just discovered Go/a compiler for the first time, please continue to
    write tests.
