---
hook: A few words on the Unix philosophy of building small programs that do one thing
  well, and compose for comprehensive functionality.
location: San Francisco
published_at: 2014-12-12T17:20:30Z
title: Small, Sharp Tools
---

Upon joining Heroku a few years back, one of the maxims that I heard cited frequently was that of building _small, sharp tools_, the idea of building minimalist, composable programs that worked in concert to a degree of effectiveness that was more than the sum of their parts. The classic example of such a system was the Unix shell, which contains a set of small utilities that provide a highly effective operating environment. Despite its origin in Unix, a number of people at Heroku including a founder and a few lead engineers had found that it lent itself quite naturally to building modern web platforms.

The best documented original source for this idea is the book [_The Art of Unix Programming_](http://www.catb.org/esr/writings/taoup/) written by Eric S. Raymond. In the book, the author boils down the overarching philosophies of Unix into a number of digestible rules, three of which are particularly applicable:

1. **Rule of Modularity:** Write simple parts connected by clean interfaces.
2. **Rule of Composition:** Design programs to be connected to other programs.
3. **Rule of Parsimony:** Write a big program only when it is clear by demonstration that nothing else will do.

(I'm highly recommend reading the rest of the rules and the whole book in depth as well.)

More background for the idea is described under a section on Unix philosophy titled _Tradeoffs between Interface and Implementation Complexity_:

> One strain of Unix thinking emphasizes small sharp tools, starting designs from zero, and interfaces that are simple and consistent. This point of view has been most famously championed by Doug McIlroy. Another strain emphasizes doing simple implementations that work, and that ship quickly, even if the methods are brute-force and some edge cases have to be punted. Ken Thompson’s code and his maxims about programming have often seemed to lean in this direction.

One example of the Unix-based small, sharp tools that are being referred to here are the basic shell primitives like `cd`, `ls`, `cat`, `grep`, and `tail`, which can be composed in the context of with pipelines, redirections, the shell, and the file system itself to work in tandem to build more complex workflows than any could provide on their own. The "simple and consistent" interfaces are the Unix conventions such as common idioms between programs for specifying input to consume, and [exit codes](/exit-status) that are re-used between programs.

Notably, it's never claimed that this is a fundamental of value of Unix, but more one of the competing philosophies that was baked into the system in its early history.

## The Web (#the-web)

The idea can be applied to the web in a similar way. Many modern web applications are choosing to build out their architecture as a set of services that communicate over the network layer, which has some parallels with the Unix model of programs that communicate via OS primitives.

This may be especially application to today with the recent popularization of [microservices](/microservices), a philosophy that advises building web architecture as a set of services that are kept small so that they can be easily reasoned about, operated, and evolved (also known as SOA).

## The Catch (#the-catch)

Unfortunately, rather than being a solution that's perfectly applicable to all problems, the trade-offs small, sharp tools are fairly well understood. In a section discussing the "compromise between the minimalism of ed and the all-singing-all-dancing comprehensiveness of Emacs", Raymond talks about how building tools that too small can result in an increased burden on their users:

> Then the religion of “small, sharp tools”, the pressure to keep interface complexity and codebase size down, may lead right to a manularity trap — the user has to maintain all the shared context himself, because the tools won’t do it for him.

He continues with another passage on the subject under the section titled _The Right Size of Software_:

> Small, sharp tools in the Unix style have trouble sharing data, unless they live inside a framework that makes communication among them easy. Emacs is such a framework, and unified management of shared context is what the optional complexity of Emacs is buying.

Once again, it's incredible how well this carries over to building modern service-oriented systems. Consider for instance two small programs that handle account management and billing. Although each is small and sharp in its own right with concise areas of responsibility and boundaries, an operator may find that even if each is well-encapsulated, complexity may arise from the shared context between the programs rather than either program in its own right. Where a monolithic system might be able guarantee data integrity and consistency by virtue of having a single ACID data store, two separate programs suddenly have to consider overhead like drifts between data sets, and the acknowledged complexity of distributed transactions.

I'm reminded of this pitfall almost every week. Just a few days ago, it turned out that an internal identifier generated by our billing component and which had previously seen only restricted use was now expected to be more broadly available. Backfilling the missing data would require a time-consuming manual operation.

## Reconciliation (#reconciliation)

As a general rule to help select tool boundaries, Raymond goes on to suggest the **Rule of Minimality**:

> Choose the shared context you want to manage, and build your programs as small as those boundaries will allow. This is “as simple as possible, but no simpler”, but it focuses attention on the choice of shared context. It applies not just to frameworks, but to applications and program systems.

We recently got this wrong. We wanted to build a system that improved the experience of managing a shared app by inventing the concept of an organization that would own an app and grant certain permissions to developers who wanted to collaborate on it. It was implemented a separate service that would use the existing Heroku API to transparently enrich the end user's experience. Although initially regarded as a very positive concept, we soon observed that the surface area of the private API's required for the new service swelled to an unmanageable size and the shallow integration made it difficult to implement new features without extensive changes across both components and their communication surface. The decision was made to re-integrate it.

The mistake we made was that although each service was kept fairly simple with the separation, the sum of their complexity was much greater than that of the integrated system. Despite being less sound from an idealogical standpoint, the monolith could be improved more quickly, operated more easily, and could better guarantee data consistency.

"Small, sharp tools" is a principle that any developer should consider while architecting their systems, but it's important to remember that it dictates that a tool can be too large _or_ too small. The boundaries of any service should be considered in depth, and if built to be small, it should be for more reason than virtue of size alone.
