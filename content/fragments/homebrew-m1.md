+++
hook = "In which Homebrew now works very well on M1 Macs, and a second Intel installation is no longer necessary."
published_at = 2021-01-09T18:54:21Z
title = "Homebrew on M1 is a go"
+++

As you've already heard from everyone on the internet, including me, Apple's new M1 computers are amazing. It's freeing being able to unplug your computer in the morning, carry it around with you all day, and just not worry about about battery life. And while the lower power usage is the best part, they're also fast and quiet [1].

Despite all the gains, I was reluctant to recommend them because of the Intel to Apple Silicon switching pains. As with PowerPC to Intel a decade ago, Rosetta works practically flawlessly, but there was still one major catastrophe in progress: Homebrew.

Some programs worked, but many didn't, and the team's recommendation was to keep two versions of Homebrew installed simultaneously, one for ARM, and the other for x86 to be run under Rosetta. It wasn't _that_ bad, but computers waste so much of your time already that why eat the pain when you can just wait another few months and have the problem disappear.

After running into Homebrew-related trouble [installing Ruby 3 yesterday](/fragments/ruby-3-on-m1), I checked in on Homebrew's M1 transition, and was _very_ pleased to see that it's much further along even compared to a few weeks ago. Almost everything works, and many formulae are now shipping with bottles for `arm64_big_sur` ("bottle" is Homebrew's term for a pre-built binary), reducing trouble and saving installation time. They've integrated [Go's 1.16 beta](https://tip.golang.org/doc/go1.16), which adds ARM support, which unblocks many other programs that depended on it.

Today, I removed my Intel installation completely, successfully reinstalling everything that'd been in it to the ARM instead. What remains are fast ARM binaries, and a much less confusing setup.

So I'm pulling out the stopper on M1 Macs, and recommending them unconditionally. They're new enough that I'd expect more problems to still exist, but with Homebrew in a good place, there's none even worth mentioning anymore. Go get one.

[1] Also: if you're upgrading from the last few years of MacBook Pro, butterfly keys are gone, and so is the Touch Bar if you go with the Air.
