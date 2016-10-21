---
title: Black Boxes All the Way Down
hook: TODO
location: San Francisco
published_at: 2016-10-20T22:59:04Z
---

* Don't use low-level databases like document-oriented stores or those lacking ACID -- treat them as a black box.
* Don't use slow language runtimes that don't really scale (like Ruby) -- treat your fast compiled language as a black box.
* Don't pull your infrastructure in house -- treat it as a black box.
* Don't pull your exception handling in house -- treat it as a black box.
* Don't do code management -- use GitHub as a black box.
* Don't run Jenkins -- use 
* Don't run Kafka

I don't need to worry about my infrastructure when I'm on AWS (usually). I can
guarantee that on the other side of that boundary things are failing
_constantly_, but that noise never reaches me.

Even when there is a serious problem I usually just need to wait, and it gets
fixed.

Conversely, when services leak, I need to worry about them constantly, even if
I'm not directly involved in fixing them.

* Breakages are generally more frequent because not as much rigor is applied to any individual service.
* Breakages tend to draw everyone's attention, even those who are not involved.
* More in house services mean more engineers. More productivity lost into management and communication overhead. The death of the small focused team.

The quality of a boundary is still important.

* Don't use JIRA because the product is terrible.
* Don't use a platform with a bad history of downtime.
