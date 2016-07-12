---
title: Avast
published_at: 2016-03-31T18:10:55Z
---

A user opened an issue on an open-source package that I maintain indicating
that it was being flagged by Avast (an antivirus package) as malware. I checked
that there nothing nefarious was going on, told the user that it was probably a
false positive, and opened a ticket with the company to see if they could do
anything about it.

They responded that it wasn't their problem and that if I wanted to see the
issue sorted out, I should open a request with them to have my package
inspected and whitelisted. I declined, and asked whether they thought it was
reasonable for the burden of responsibility to fall to the developer of the
software that happened to fall into the wrong end of an Avast bug to jump
through hoops to correct it. They asserted that it wasn't just reasonable, but
standard practice in the antivirus industry, then followed up with a trailing
platitude on how user safety was their top priority.

## Ghosts of the Past

As the lower bound of exploits becomes increasingly handled by ever more secure
operating systems, and the upper bound becomes increasingly more sophisticated
and harder to generally handle, I can't help but think that this apathetic
attitude might be a relic from another age. There was a time when running
antivirus was so integral to the safe use of a computer that they could have
demanded protection money from practically any software manufacturer, and those
manufacturers would scramble to appease them. Today, with the future of
antivirus being far from assured, it might be time to start rethinking these
policies.

I've been lucky enough to have never run into this problem before, but I
imagine that it must something that developers producing software for the
desktop have been dealing with for ages. Has anyone else seen this problem with
a project that they were shipping?
