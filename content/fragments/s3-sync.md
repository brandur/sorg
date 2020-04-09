+++
hook = "Investigating excessive data transfer via `aws s3 sync`, and why it pays to read the manual more carefully."
published_at = 2020-04-08T23:31:33Z
title = "Development log: Minimizing S3 transfer during sync"
+++

Over the years my AWS bill has been slowly creeping up. Nothing serious -- now in the $10s/month where it started in the $1s -- but I've had a sneaking suspicion for a while that it's not because my monthly visitor numbers have improved that much.

Popping open AWS' Cost Explorer, I wasn't surprised to find that S3 was the main culprit, but _was_ surprised to find that the bulk of my costs weren't monthly retention or data transfer out, but data transferred _in_.

This seemed somewhat incredible, so I started examining the details of the builds of [this website](https://github.com/brandur/sorg) as it runs its CI from GitHub Actions and auto-deploys to S3. The basic process involves (1) compiling the Go program from source, (2) having it build the static site, then (3) using the AWS CLI, specifically `aws s3 sync`, to send those to S3:

```
aws s3 sync ... s3://...
```

## When `sync` degrades to `cp` (#sync-cp)

It seemed to be working correctly, except that on closer inspection of the last step, for a command called `sync`, there sure wasn't a lot of _syncing_ going on. Every file was being copied to S3 every time.

In some places that made sense. `aws s3 sync` figures out what needs to be resent based on the modification times (mtimes) of a local and remote target sharing the same name. HTML sources are built from scratch on every CI run, so it's not surprising that they're sent every time. However, a large bulk of the site's total size are static assets and those are not rebuilt, and rarely even modified. Once pushed to S3, they _should_ be fresh for a long time to come.

So what was going on? Well as it turns out, Git. Git preserves a large amount of data on every file that comes into its graph -- contents, author, commit date and context, but quite notably, doesn't preserve any file's modification time. Each of my CI runs was doing a fresh `git clone` from source and every static asset was getting the current time as its last modification date, thereby making every file in my S3 bucket stale. I was pushing 100+ MB of static assets to S3 on every build, and a cron job was triggering builds hourly. So on the order of gigabytes per day. Again, it's S3, so luckily that won't break the bank, but it was still a lot of needless waste.

## Stabilizing mtimes (#stabilizing-mtimes)

The most obvious mitigation would be to stop relying on modification times and use an alternative comparison to function figure out what to sync. If using `rsync` for example, the `--checksum` flag would tell it to use checksums for the comparison instead of mtime and size. Unfortunately though, `aws s3 sync` doesn't support that or any other alternatives ...


**... vvvvvvvvvvvwwppppp ztÅ±rÅ‘ tÃ¼kÃ¶rfÃºrÃ³g \*RECORD SCRATCH\***

## Hubris (#hubris)

Okay, full disclosure time. I wrote a custom Go program to solve the problem by checking the commit history of every image asset via `git log`, then using the timestamps of those commits to set stable mtimes for sync. It was a little slow, and a lot of a hammer, but it worked perfectly, saving me 100+ MB of S3 ingress on every build.

But as I was writing this article about it, I decided to go back and read the manual one more time. I'd done it before to verify that nothing like `--checksum` existed in the AWS CLI, but not carefully enough. On that second pass I discovered the `--size-only` option. `aws s3 sync` works by comparing mtimes _and_ sizes, and although `--size-only` doesn't fundamentally change that strategy, it does disable the mtime half of the equation -- files are considered equivalent as long as their name and size are equal.

`--size-only` has some risk for for text-based content where minor changes (e.g. one character changed) might lead to a file of the same size. But for media assets where even a minor tweak tends to rewrite many bytes, it's fine, and very fast.

I threw out my custom script solution and replaced it with `--size-only`, which _also_ worked perfectly. I decided to publish this article anyway as a reminder to not be too quick to jump into writing custom software -- sometimes the right answer is to read more carefully.
