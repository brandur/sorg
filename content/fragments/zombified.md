---
title: Zombified
published_at: 2016-12-29T05:10:25Z
hook: On duck typing, zombie processes, and language APIs that are unsafe by
  default.
---

I recently moved my media backups over to a Ruby script that wraps [rclone]'s
sync command. The script partitions rclone invocations onto different threads
by media type so that small files like photos are synced frequently and not
blocked by uploads of larger and less important files like videos. It all gets
sent up to Amazon Cloud Drive, which seems to me to be the best deal out there
at $60/year for unlimited space.

It was all working beautifully until a day after putting it in place, I tried
to SSH back into my home machine and, to my chagrin, got back an error message
telling me that it was unable to start a shell.

Historically these types of failures are usually because OS X is a bad
operating system for a headless box, so my first assumption was that some
system-level task had kicked in and rebooted my computer. Either that, or an
errant web browser had eaten the entirety of its memory. The problem's timing
was a little too suspicious though, so I inspected my (identical) locally
running backup script to make sure that nothing was amiss.

Well, something was amiss. In the few hours that it'd been running, the program
had accumulated dozens of zombified rclone child processes. This was almost
certainly happening on my home machine too, and it couldn't give me a shell
because every available file descriptor had been devoured by Ruby to
communicate with its zombies.

![Zombies](/assets/fragments/zombified/zombies.png)

So what went wrong? I eyeballed my code and looked at the line where I was
shelling out to rclone. I'm using Ruby's `IO.popen` to get better control over
the I/O coming out of the process. By default, it'll give you an `IO` object
which you're responsible for closing, but it also comes with a safer block
syntax which handles cleanup automatically. I'd intended to use the latter.
Here's what the call looked like:

``` ruby
IO.popen(...).each { ... }
```

See it? I'd accidentally included an `each` invocation and was passing my block
to that instead. And just my luck, the `IO` instance returned by `popen`
includes `Enumerable` so Ruby's duck typing makes `each` a perfectly valid
call. After being iterated, the object was being silently discarded and its
resources were left dangling.

The corrected invocation looks almost identical:

``` ruby
IO.popen(...) { ... }
```

This one simple mistake means that my home server is now out of comission for
two weeks until I can get back there to manually reboot it. This is one of
about a thousand different reasons as to why I'm very keen on Rust these days.

[rclone]: http://rclone.org/
