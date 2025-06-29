+++
hook = "Fun times with Dropbox's new `~/Library/CloudStorage` location and File Provider API integration."
published_at = 2025-05-26T00:28:04-06:00
title = "Be careful with Dropbox"
+++

I've been a Dropbox users going on fifteen years now. It's one of the most frustrating products in my arsenal because fifteen years ago it was _perfect_, but every new release just makes it a little bit worse than it was before. It's still fine to use, but you can see the writing on the wall as the long term trend is all in the wrong direction.

Despite that, I previously would've lavished it with praise in that I've never once had trouble with data loss or data integrity. Despite increasing feature bloat, it did what it was supposed to, syncing files to the right places, and doing so _reliably_, which is pretty much all I need out of it.

That ended Friday, when I was installing Dropbox on a new laptop. My Dropbox size runs ~500 GB, so when credentialing a new machine, I copy it from another computer on the network for speed, and to conserve precious bandwidth [1] :

* Rsync `~/Dropbox` from an existing computer to the new one.
* `brew install dropbox`. Open it, log in, close it.
* Replace the contents of `~/Dropbox` with the `rsync`ed copy.
* Open Dropbox, let it sync against the new data. It should find everything it needs already there.

Dropbox made a change in the last couple years wherein they moved the standard `~/Dropbox` on Mac to a new `~/Library/CloudStorage/Dropbox` location. I now know that folders in this directory are meant for use with Apple's [File Provider API](https://developer.apple.com/documentation/fileprovider/).

Apparently the change had been introduced for macOS Ventura (two major versions ago), but there must've been an incremental roll out because I set up a computer last year and didn't run into it then. Once you've been opted into the feature, you cannot opt out. Changing back to `~/Dropbox` is not an option.

Normally I do a wholesale swap of `~/Dropbox` with my locally copied version, but seeing this new magic folder in `~/Library`, I worried there'd be some irreversible effect if I did it the normal way. Instead, I closed Dropbox, `cd`ed into the folder to `rm` all the files acting as cloud "stubs", intending to replace them with materialized versions from my local copy.

What a mistake. I dumbly assumed that with Dropbox closed, any changes I made to the folder would be safe, just like they were in every previous version of Dropbox. Not so. At all.

I got suspicious after about ten seconds. Normally an `rm` even on gigantic directories is near instant, but this one was running long. I `SIGINT`ed it, but the damage was done.

I'm sure you guessed what happened already. `~/Library/CloudStorage` is a magic location, and folders in it use macOS extension voodoo to make arbitrary changes in a cloud storage API. Despite Dropbox not being open, it'd used a Mac API to intercept the `rm` and started to remove everything. My other computers had already synced the deletions. 100s of GBs gone in seconds.

Dropbox has a good "undelete" function, so I was able to log into their web UI and recover all the deleted files, but I was left with the problem of all my other computers having purged their local contents, with potentially 100s of GBs on each needing to be synced back down (and I thought I was _saving_ bandwidth when I started doing this). Worse yet, Dropbox puts any files it deletes into a `~/Dropbox/.dropbox.cache` directory, but can't reuse any of that data when files are recovered, so it just makes a copy. Dropbox doesn't purge its cache often, even if disk space gets critically low, so every computer potentially needed 2 * 500 GB =~ 1 TB of free space for the full recovery, which they didn't have.

Two days later, I got everything back to where it should be, but all I could think afterwards was what a stupid, unforced error this all was. A mandatory move to `~/Library/CloudStorage/Dropbox`/File Provider API has no marginal utility for the user [2], even if it makes product managers at the company feel good about themselves.

Being particular incensed at this moment, I started looking into alternatives immediately. There's dozens out there, but my approximate evaluation is that there isn't one that's a crystal clear, unambiguous win that'd I'd be excited about doing the work to switch over to, which is too bad.

What I'm really looking for is Dropbox circa 2011. The one without the gratuitous/dangerous product changes, without an Electron app, and without the nags to upgrade my account which I already pay $120/year for.

Anyway, I doubt most users will run into this one as it was a confluence of stupid things that led me down this path, but I'd just caution like the title says: be careful with Dropbox. Don't `rm` too much. Don't assume intuitive cause and effect. Don't assume operations are safe even if the app is closed.

[1] I'm on Comcast, where data is worth more than gold, with a strict 1.2 TB cap.

[2] Dropbox vaguely describes the advantage as "to more deeply integrate with macOS and fix issues related to opening online-only files in third-party applications".
