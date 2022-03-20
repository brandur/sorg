+++
hook = "iOS Shortcuts are useful, but they make easy things hard, and hard things (probably?) impossible."
published_at = 2022-03-20T00:13:12Z
title = "iOS Shortcuts: Appending dictated memos to Notes"
+++

I made my first ever foray into iOS Shortcuts today by copying the recipe [in this tweet](https://twitter.com/craigmod/status/1404083121763405829) to build a tiny Siri-driven shortcut that appends the content of a spoken memo to a specific note in (Apple) Notes.

The tweet's about a year old and some of the UI and naming in Shortcuts have changed subtly. Here's what (I think) it's supposed to look like today:

{{FigureSingle "Siri-activated shortcut for appending the contents of a voice memo to a note." "/photographs/fragments/shortcuts-notes-memos/shortcut.png"}}

Pretty simple. Two parts:

* Look up a note with a specific name, put it in a variable.
* Take spoken text and append it to found note.

I thought I'd be able to activate it by saying, "Hey Siri Scratch [1] Here's my memo", but there's a meaningful pause that has to be in there. Actual usage is more like, "Hey Siri Scratch <wait for Siri to activate shortcut> Here's my memo". You can activate it without unlocking the phone, which is nice for when you're hands-free out on a walk, although that's somewhat dependent on how much you trust Apple's transcription.

## So, is low code, good? (#low-code)

The experience was a heavy reminder as to why I'm not big on GUI-driven coding solutions ("low-code" according to latest tech lingo). Even having a near-perfect recipe to copy, it took me 20+ minutes to put it together. Initially, iOS insisted on presenting only a basic list of possible actions to add, rather than one that included a search bar, so I was scrolling through a hundred unordered actions trying to find specific ones by visual scan. I'm still not exactly sure what I did to finally get it to flip to a search-based version, but it wasn't obvious. After changing configuration on specific actions, in some cases existing actions would refuse to connect with the original, making me remove them and add a fresh version.

It works now, but gods, imagine doing something actually hard. Once again, I'll harp on the fact that despite the improvements to battery, security, and resource management brought to us by sandboxed operating systems like iOS, it's a shame how much we've lost in extensibility. Still, some programmability is better than none, so it's nice that Apple's trying.

After initial success with my first shortcut, I tried to build another, only to find that another major limitation at the moment is app support. Some apps like Strava and Streaks support them, but most don't -- for example, I tried to build one to pause any currently playing sound and a start a meditation session, only to find that none of the meditation apps I have installed can do that. I was able to make one that lets me set my morning calisthenics task as complete via voice, which is something.

[1] "Scratch" being the name of the shortcut.
