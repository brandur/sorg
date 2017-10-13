---
title: Active Versus Passive Safety in Production Systems
published_at: 2017-10-13T15:08:44Z
location: San Francisco
hook: TODO
---

Safety.

## Atomicity (#atomicity)

This is where the importance of using a database with
atomic guarantees comes in.



If you're running MongoDB or a database anything like it,
your system is actively safe. Data integrity problems are
introduced as operations fail before they're supposed to,
and you might have scripts designed to go through and fix
them. More likely though, the active safety system at play
is a human.

## Safety by human (#human)
