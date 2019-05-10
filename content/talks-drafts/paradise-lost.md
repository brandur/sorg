+++
event = "FakeConf"
location = "San Francisco"
published_at = 2018-09-07T15:36:33Z
subtitle = "The trials of building software outside of an ACID and relational environment"
title = "Paradise Lost"
+++

class: middle

# Paradise Lost

<!-- Title slide. Content hidden. Speaker notes used as intro. -->

???

Intro text.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce non aliquet
urna. Fusce nisl ex, lacinia sit amet suscipit eget, consectetur sit amet
nunc. Ut suscipit erat quis molestie feugiat. Ut pellentesque, nisi ultrices
sodales tristique, dui velit gravida ante, sit amet sodales erat arcu a
purus. Integer purus ligula, ornare vitae mollis consequat, finibus quis
dolor. Praesent sit amet molestie enim. Nunc ultricies sem quis convallis
efficitur. Proin a ornare sem.

---

class: middle

Follow along:<br>
https://brandur.org/paradise-lost

Find me on Twitter:<br>
[@brandur](https://twitter.com/brandur)

???

I publish most of my work on this site or [Twitter](https://brandur.org/twitter).

I'll save you the boring history of my career, but the short version of it is that for a long time I ran a large production critical application on a large Postgres installation. I still run a large production critical application, but I now do it on NoSQL instead.

---

background-image: url(./raws/talks/paradise-lost/mongo-ad-censored.jpg)

???

I'm not here to slime anyone, but the company who makes the product I use is somewhat notorious for selling their weaknesses as strengths. This billboard is real unfortunately, with the only Photoshopped part being on the right where I blurred the identity of the defendant.

And although this may seem pretty egregious, it's actually a marked improvement from their old days. Their first claim to infame was when one of their engineers published a set of benchmarks that showed how much faster their data store was compared to well-known databases at the time, and it did indeed show that their product was really, really fast.

But they ran into a bit of trouble when people realized that at the time they were only persisting changes in memory -- not to disk -- and they didn't have a WAL or anything of the kind, whereas the databases they were comparing themselves to all did. It turns out that when you're just setting values in memory over and over again you can produce some pretty awesome results.

---

???

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce non
aliquet urna. Fusce nisl ex, lacinia sit amet suscipit eget, consectetur
sit amet nunc. Ut suscipit erat quis molestie feugiat.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce non
aliquet urna. Fusce nisl ex, lacinia sit amet suscipit eget, consectetur
sit amet nunc. Ut suscipit erat quis molestie feugiat.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce non
aliquet urna. Fusce nisl ex, lacinia sit amet suscipit eget, consectetur
sit amet nunc. Ut suscipit erat quis molestie feugiat. Ut pellentesque,
nisi ultrices sodales tristique, dui velit gravida ante, sit amet sodales
erat arcu a purus. Integer purus ligula, ornare vitae mollis consequat,
finibus quis dolor. Praesent sit amet molestie enim. Nunc ultricies sem
quis convallis efficitur. Proin a ornare sem.

---

???

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce non
aliquet urna. Fusce nisl ex, lacinia sit amet suscipit eget, consectetur
sit amet nunc. Ut suscipit erat quis molestie feugiat. Ut pellentesque,
nisi ultrices sodales tristique, dui velit gravida ante, sit amet sodales
erat arcu a purus. Integer purus ligula, ornare vitae mollis consequat,
finibus quis dolor. Praesent sit amet molestie enim. Nunc ultricies sem
quis convallis efficitur. Proin a ornare sem.

<!-- vim: set tw=9999: -->
