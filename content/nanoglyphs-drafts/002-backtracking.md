+++
published_at = 2019-07-17T22:29:45Z
title = "Collapse By Backtracking"
+++

![Iceland](/assets/images/nanoglyphs/002-backtracking/iceland@2x.jpg)

Hello.

---

## Backtrack collapse (#backtrack-collapse)

[Cloudflare's postmortem of their July 2nd
outage][cloudflare] was the highlight of the week. Like
most production incidents, there were a variety of
contributors to the problem, but the most direct for this
one was the introduction of a new regular expression.
Specifically, this part of it: `.*.*=.*`.

Backtracking is a technique that some regular expression
engines use to match as generously as possible, but which
has the serious disadvantage of potentially being
computationally expensive to do so, and with that expense
growingly non-linearly with input size. Matching `x=x` to
`.*.*=.*` takes 23 steps on a backtracking engine, `x=xx`
takes 33, and `x=xxx` takes 45. Scroll down to the post's
appendix for a deep explanation of the effect including an
animated visualization.

The solution is complicated, but luckily, simple for most
of us. Some regular expression engines like [RE2][re2] or
[Rust's][rustregex] are designed for speed and security,
and guarantee execution in linear time with respect to
expression and inpput size by avoiding backtracking. Regex
constructs that require backtracking like backreferences or
look-around assertions are explicitly not supported.

## Serverless Postgres (#serverless-postgres)

Amazoned [GAed their serverless version of
Postgres][aurorapostgres]. We've known for some time that
Aurora has found an impressive way to plug a more scalable
storage layer into popular databases like MySQL and
Postgres. What I hadn't realized is that they're apparently
able to plug a specific store into a generically
provisioned database server at speed.

The announcement post describes how serverless Aurora makes
this fast enough to work by maintaining a warm pool of
standby servers provisioned and ready to go. When a request
comes in for a database that's not running, they pull a
server out of the standby fleet, plug in the database's
storage, and it's ready to go. I'll withhold final
judgements before trying it, but from a distance it's
impressive that they got it working.

I still can't tell whether a serverless database is a good
idea, but it's so novel that it's great to see invention.
Within the next few years we should have a better idea of
whether the serverless craze really is the next deployment
stack of the future or a buzzword that's been popular to
latch onto.

---

[aurorapostgres]: https://aws.amazon.com/blogs/aws/amazon-aurora-postgresql-serverless-now-generally-available/
[cloudflare]: https://blog.cloudflare.com/details-of-the-cloudflare-outage-on-july-2-2019/
[re2]: https://github.com/google/re2/
[rustregex]: https://docs.rs/regex/1.1.9/regex/
