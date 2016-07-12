---
hook: Learn how to start encrypting dotfile secrets with GPG, and some techniques
  for getting those encrypted files integrated with your toolchain.
location: San Francisco
published_at: 2014-11-10T23:46:34Z
title: Dotfile Secrets and GPG
---

A recent [scare installing Yosemite](/fragments/yosemite-progress), was a nice reminder to check the state of affairs of my backups. I do a decent job for most of my important data, but one weak point were my dot files, which I'd traditionally spread around as little as possible because the standard usage of so many programs had led me to storing a lot of plain text passwords and keys spread out all over the place.

This put me on the path toward the complete eradication of any secrets on disk in my dotfiles with the eventual goal of being able to back them up to the cloud, and feel comfortable about doing so. As it turns out, this is a more difficult prospect than you might imagine: although we have a pretty obvious tool to help with this in [the form of GnuPG](https://wiki.archlinux.org/index.php/GnuPG), integration across the system is spotty and varies widely in terms of design and setup difficulty. For better or for worse most Unix programs are best adapted for reading secrets out of plain text files, and as part of getting a complete GPG setup you'll have to fight this natural order of things.

In terms of basic GPG setup, I won't even try to one up [Thoughtbot's excellent article on the subject](http://robots.thoughtbot.com/pgp-and-you) which provides a comprehensive walkthrough on getting off the ground with GPG, including generating your key. As a prerequisite to doing anything in this article, you should give it a thorough read. After that's done, if you're still ready to use GPG to secure your secrets, you can turn to this article to help assemble a gpg-agent configuration and start integrating the suite with the rest of your tooling.

## gpg-agent (#gpg-agent)

gpg-agent's job is to remember your passphrase for some amount of time so that you don't have to compromise between encrypted secrets and your sanity. This would otherwise be a problem as you were forced to enter a passphrase three times a minute.

gpg-agent is a little more sophisticated than its cousin ssh-agent in that it can write out metadata about its process to a file, but like ssh-agent it still needs a little extra infrastructure built into your *rc files. The [Archlinux wiki page](https://wiki.archlinux.org/index.php/GnuPG#gpg-agent) on the subject contains a nice little bootstrap script that will launch a gpg-agent if necessary, but will otherwise export the settings of the copy that's already running:

``` sh
#!/bin/sh

# start-gpg-agent

gnupginf="$HOME/.gpg-agent-info"
gnupglog="$HOME/.gpg-agent.log"

if pgrep -x -u "$USER" gpg-agent >/dev/null 2>&1; then
    eval `cat $gnupginf`
    eval `cut -d= -f1 $gnupginf | xargs echo export`
else
    eval `gpg-agent -s --daemon --write-env-file "$gnupginf" \
      --log-file "$gnupglog"`
fi
```

This should then be included in your appropriate *rc file so that `gpg` can always find a running agent (note the leading `.` which is used to source the script):

```
. start-gpg-agent
```

Now when opening a new shell, you should be able to see that your agent is alive and well:

```
$ gpg-agent
gpg-agent: gpg-agent running and available
```

When you run a GPG command, it will initially ask for your passphrase, but will then cache it for the next time around:

```
$ echo "encrypt me" | gpg --armor --encrypt -
-----BEGIN PGP MESSAGE-----
Version: GnuPG v1.4.13 (Darwin)

hQEMA1XJl0SO//WLAQf/QsLhIqOSgfKtA3EwiIw290aNhpa1gl6rLXXPw3N66zuH
...

#
# CACHED! A passphrase is no longer necessary ...
#

$ echo "encrypt me" | gpg --armor --encrypt -
-----BEGIN PGP MESSAGE-----
Version: GnuPG v1.4.13 (Darwin)

hQEMA1XJl0SO//WLAQf/eBsnpMMoTZIBYEboXmdZcs73EaKD/HDcglQM9k7wyvt3
...
```

With the right configuration, gpg-agent can also double as ssh-agent by duplicating its public interface, allowing it to perform the jobs of both agents. I personally don't use this option because gpg-agent somewhat invasively insists on managing your keys itself by assigning them its own passphrases, but it might be a feature that's worth checking out.

## Ecosystem (#ecosystem)

Now that GPG is up and running, the next challenge is to get it integrated into your toolchain so that you can purge any secrets that are stored in plain texg. This is where things start to get a little more challenging because there really isn't a standardized methodology for getting programs to cooperate with GPG, and as a result, many programs have been written in ways that don't allow it to be plugged in easily. That said, thanks to the strong conventions of the Unix environment, it's amazing how many programs _can_ be backpatched for GPG support using nothing but simple shell primitives.

The most trivial example is probably Curl, which is used to reading from an unencrypted .netrc file, but with a simple stdin pipe trick it [can be made to read from an encrypted .netrc.gpg](/fragments/gpg-curl) instead. Other programs like the Heroku toolbelt integrate with GPG out of the box (in this case by preferring a .netrc.gpg over a .netrc if one is available), but this is rare.

I've converted quite a number of my standard tools now, and will curate the following list of mini-guides for GPG integration as I write them:

* [GPG and Curl](/fragments/gpg-curl)
* [GPG and the Heroku CLI](/fragments/gpg-heroku)
* [GPG and HTTP Git](/fragments/gpg-git)
* [GPG and s3cmd](/fragments/gpg-s3cmd)
