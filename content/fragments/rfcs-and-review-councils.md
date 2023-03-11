+++
hook = "Thoughts on whether it's possible to get most of the value from RFC-driven engineering, but save on the committees and heavy process."
published_at = 2023-03-11T13:11:45-08:00
title = "RFCs and review councils"
+++

Squarespace writes about [The power of "Yes, if"](https://engineering.squarespace.com/blog/2019/the-power-of-yes-if) on the topic of RFCs and review councils (from 2019, but recirculated recently). Major takeaways:

* RFCs can never have a "rejected" status, only a "not yet".

* Councils like an Architectural Review meet to work through RFCs and towards consensus:

> So we introduced Architecture Review, a twice-weekly meeting where a small group of our most senior engineers review an RFC in depth. Each major RFC gets a one hour session and, since the same people review them, the reviewers build up a picture of everything that’s happening in the organization, and can notice overlaps or incompatible initiatives.

* An explicit aim is to find rough concensus. Unanimous opinion isn't necessary.

## Thought as process (#thought-as-process)

The superpower of an RFC process is as a forcing function to have the writer themselves really think through the scope of the change they're proposing. There's a bad tendency in a lot of people to have an idea, not think about it in too much depth, start implementation, and only then discover a colorful assortment of complications, necessitating patch after patch to follow up the original change, and unless there's a more senior person to insist on quality control, more often than not hack after hack as the deadline draws nearer.

I'm more skeptical of RFCs as a consensus-making process. We had one at Heroku modeled on the original GitHub RFCs involving a Markdown document in a pull request which would be merged into mainline if the RFC was accepted. I fully acknowledge  and somewhat regret that I was more cantankerous at the time and unnecessarily prickly in some of them, but we had some _epic_ debates on APIs and system architecture in those pioneering days, and all via async, long-form, Markdown-styled pull request comment such that I was as often as not talking to people in Copenhagen as San Francisco. (I would come to appreciate this even more later when my new world became one of pain, trying to hold discussions of depth in tiny comment boxes on the side of Google Drive or Dropbox Paper documents.)

Something we did well at the time is that titles really didn't matter much, and by widespread convention no one pulled rank (although you might pull "ownership rank", like if someone is proposing a change that'd add a lot of operational burden to something you run). But as good as it was in many ways, it was noticeably difficult to come to consensus on topics where there was more disagreement. Most often, these would be resolved over drinks or not at all, in the latter case lingering in an open zombie state for far too long.

I found myself often thinking that it'd be nice to add maybe one infrequently-used layer to the largely flat system, so there'd be at least one authority to act as tiebreaker in cases that got stuck, but later learned that's fraught in its own way. Stripe had a process close to what Squarespace describes with RFCs by another name, and committees for API, Infrastructure, etc. But the reception you could expect was highly correlated with rank. High status people would get a few cosmetic notes for appearance sake, then a rubber stamp. Lower status people would need to _wage war_ in a process that was often drawn out into multiple months. Changes would usually go through in the end because it was a high functioning environment, and projects tend to at least get started in high functioning environments, but after hundreds of hours lost to a lengthy review process, proposals might be 10% better than they'd been originally, and with long delays added in. It really left you wondering if it'd all been worth it. Things got even worse when the term "staff engineer" was invented (~2018 [1]), and there were now crystal clear delineations of authority. Pulling rank went from implicit and rare to explicit and common. And this was even worse than it sounds because an appellation of "staff" is more a recruiting incentive than anything else, so there was little correlation between seniority and quality of opinion.

## RFCs without the committee (#without-the-committee)

Assuming you're hiring well, I believe that most people will mostly be doing the right thing, but there are the occasional projects that need heavy course correction, or in the rarer case, sometimes to die (because they really are bad ideas even if well-intentioned). When you're small, backchanneling works, but once you have an organization of non-trivial size, the process probably has to be more formal. My haphazard preferences would be something like:

* Default "yes", assume a core value proposition is the learning act of authoring the RFC itself, and to be fairer to the person doing the work to write the proposal.

* Low-touch. Foster a culture of feedback-where-necessary rather than drive-by opinions.

* Minimal overhead so that projects can start without massive activation energy. The twice-weekly meetings described above feels like way too much process to me.

* I still like the idea of a tiebreaker who can unmyre proposals that've come to an impasse and consensus is proving elusive. But you have to be _really_ careful here. It needs to be someone who's expertise is widely respected, but also someone who's not excited about wielding it. The classic "type of person we need in politics versus type who enters it" problem.

## Lossy descriptions (#lossy-descriptions)

Engineering posts can be thought of as prisms that refract a real thing from inside of an organization to its outside, but in a way that reveals only its most positive qualities, and tends to obscure the inconvenient details. When I think about how Stripe's dysfunctional committee process would be written about in a blog post, I can't help but be quite skeptical about Squarespace's write up. A comment [posted to HN](https://news.ycombinator.com/item?id=34947274):

> It’s a bitter pill to swallow seeing this article. I am a former Squarespacer who left in the past year. All I will say is consider a few things carefully about this:
> 
> 1. This introduces a lot of process, bureaucracy, and chances of gatekeeping politics. The whole “yes if” mantra just centralizes the political power to the hands of fewer senior leaders by cutting off legitimate blocking criticism from the rank and file.
>
> 2. Judge processes by outcomes. What has Squarespace done / shipped / accomplished in the last four or five years? I’m not alone in feeling like it’s a head-scratcher why the company would pay the most senior engineers to erect slow moving bureaucratic decision-by-committee disempowerment structures like this, and then turn around and act like that’s a good thing or constitutes what you’d want to see as productivity from staff or principal engineers.
>
> While I’m a proponent of writing things down, not surprising peer teams, pulling blockers or conflicts forward quickly … I don’t think templates and councils help more than they hurt, and often domain experts in different teams need empowerment to just totally bypass that time-wasting crap and do what needs to be done. If they don’t uphold the responsibility to do due diligence on review or finding conflicts, fire them or take them off the project, but don’t make everyone wear design-by-committee training wheels and act like that’s non-dysfunctional good engineering.
> 
> YMMV of course, but I reckon it’s really worth it to ask what is the real engineering output, speed, productivity following from this kind of large process overhead.

It's anonymous, so take it with a grain of salt and all of that, but it sounds plausible to me.

[1] Or rather, came into common use. I save all recruiting emails. Never saw the term before Dec 2018, was relatively rare in 2019, and then plastered all over everything from 2020 on.