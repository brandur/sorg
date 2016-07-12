---
title: Do One Thing and Do It Well?
published_at: 2016-01-04T08:48:16Z
---

I came across [an old Slashdot interview with Rob Pike][pike-interview] today,
and was amused to see that he raised some unusually pragmatic ideas on the old
Unix philosophy of ["do one thing and do it well"][do-one-thing]. The related
question asked by a reader:

> Given the nature of current operating systems and applications, do you think
> the idea of "one tool doing one job well" has been abandoned? If so, do you
> think a return to this model would help bring some innovation back to
> software development? 

And Pike's response:

> Those days are dead and gone and the eulogy was delivered by Perl.

This is often one of those places where developers suddenly get philosophic to
the point where otherwise good programs are written so that they become
unnecessarily difficult to use in the name of correctness. The best example of
this is still the modern command line; very functional for people like yours
truly who enjoy the throwback to the 70s and all the esoteric arcana that comes
with it, but which presents an artificially high bar to entry for anyone who
hasn't already memorized the usage of about 50 common Unix programs, 500
associated flags, and how to compose them to do what you want.

A modernized version of the philosophy might run something like "identify core
competencies and do them well". Such a tenet should serve any developer well as
long as they keep the converse problem dictated by [Zawinski's Law][zawinksi]
in mind as a counterbalance (i.e. "every program attempts to expand until it
can read mail").

[do-one-thing]: https://en.wikipedia.org/wiki/Unix_philosophy#Do_One_Thing_and_Do_It_Well
[pike-interview]: http://interviews.slashdot.org/story/04/10/18/1153211/rob-pike-responds
[zawinksi]: http://www.catb.org/jargon/html/Z/Zawinskis-Law.html
