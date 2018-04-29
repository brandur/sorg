---
title: Reflections on Rust, and the sand castle metaphor
published_at: 2018-04-27T14:46:34Z
hook: Some personal progress with Rust.
---

A month ago, I wrote about [how I was frustrated with my
progress in Rust][walls]. These days, I'm still no expert,
but I've made progress.

Programming languages get easier to learn the more of them
you know. A lot of this is just pattern matching: syntax
changes, but there's a lot of overlap in features, and the
similarities get more apparent. Some of my recent
acquisitions have been very fast to learn -- with Go, I was
writing functional code within hours, and had a solid grasp
of the ecosystem and conventions within a few weeks.

Rust is in its own league. I'm about five months into
writing it on a pretty frequent basis, and its taken this long to
internalize a lot of its ideas. This is especially true
when it comes to the more complicated ones like moves and
futures, but also true for simpler ones like borrows. What
I didn't know when I wrote about it in frustration a month
ago is that it doesn't take a little longer to be
effective in Rust compared to other languages, it takes 10
to 20 times longer.

But that learning curve isn't without its benefits. When
I'm working with languages with poor guarantees (e.g.,
Ruby, JavaScript), I see myself building a sand castle on
the beach with the tide coming in around me. While I'm
focused on finishing one part of the structure, the rest of
it is being worn away by the rising waters of time and
entropy. The whole only retains some semblance of form as
long as I diligently rush from one part of it to another,
reshaping each of them before they disintegrate into the
sea.

Using compiled languages feels like I'm no longer building
out of sand. In Go or C#, between a test suite and the
compiler's guarantees, I rest easier knowing that the code
that works today is probably going to work tomorrow.
Entropy is still taking its toll, but more slowly.

Rust is another step into the beyond. When finishing a
feature and its test suite I'll run my program to see it
in action, but just as a formality -- I already know it
works. I also know that it's going to _keep_ working
because meticulousness of the compiler is so good at
catching regressions. My castle on the beach is made from
steel. Entropy will swallow it eventually, but on a totally
different scale of time.

The older I get, the more value I see in these sorts of
guarantees. Software is partly a production problem, but
it's mainly a maintenance problem. Powerful primitives like
Rust are our path to building sustainable services that
don't need human attention around the clock, and which can
be expected to operate reliably over epochs measured in
decades instead of weeks.

[walls]: /fragments/rust-brick-walls
