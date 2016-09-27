---
title: GitHub Projects
published_at: 2016-09-23T15:35:45Z
hook: The new organization feature for GitHub repositories.
location: San Francisco
---

Column-based project management tools have been the Holy Grail for agile shops
for years now, 

Between Heroku and Stripe I've tried just about all of them: we started with
Pivotal Tracker way back in the day before eventually standardizing on Trello.
Just in the last year at Stripe we've moved from Trello to Phab to JIRA.

You'd think that the space would be completely saturated by now, but I was
inordinately happy when GitHub announced at their recent GitHub Universe
conference that they were making their way into the space. Despite a
considerable amount of tooling in the area already existing, it's amazing how
every package out there has at least one major deficiency that makes it
overwhelmingly unpleasant to use.

GitHub's Projects is the best implementation I've seen to date, but could still
be vastly improved with the addition of a few more features.

## Best Features (#best-features)

### Speed (#speed)

### Interface (#interface)

GitHub's overwhelmingly excels compared to any other company at producing
interfaces that strike a perfect balance between power and minimalism. Most of
their pages have _exactly_ the right number of buttons on them to do what you
need to do; no more and no less. You can get work done with a minimum number
of clicks.

TODO: GitHub screenshot

This is something that a company like Atlassian will never understand. JIRA's
interface is _packed_ to the brim full of buttons and doodads, most of which no
one will ever use. Getting anything done will necessitate about a minimum of
about four clicks (e.g. click on issue, click edit button, change the state of
a dropdown, and save). It's a classic example of focusing on having the longest
possible feature list without ever giving even a stray thought to UX.

TODO: JIRA screenshot

Another poor example here is Phab.

TODO: Phab screenshot

### Flexibility (#flexibility)

Projects doesn't prescribe any particular 

### Lightweight Issues (#lightweight-issues)

Notes.

Link through for issue management.

JIRA pulls up a sidebar that allows some changes to be made to an issue, but
some features are mysteriously missing from it (for example lifecycle
management), and overall it wastes more time than it saves.

TODO: JIRA screenshot

## Missing Features (#missing-features)

### Can't Link Issues Between Repositories (#crosslinking)

Even referencing an issue's exact name doesn't create a link to it. This is a
particularly mysterious design decision because linking between issues on
normal issues is something that's worked for years.

TODO: GitHub screenshot of missing link

### No Backlog (#backlog)

TODO: GitHub screenshot of manually adding issue cards

### Heavy Edit Modals (#modals)

TODO: GitHub screenshot of modal

TODO: Trello screenshot of modal
