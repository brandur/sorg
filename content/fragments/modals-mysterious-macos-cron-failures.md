+++
hook = "Bequeathing cron scripts permission to run."
published_at = 2024-03-05T12:54:57-08:00
title = "Modals and mysterious macOS failures"
+++

I have a Mac Mini that I try to use as a headless server. The last few years have trained me such that if there's ever a mysterious problem or failure, its most likely explanation is that there's a new pop up back in the GUI that I'm trying not to use.

The other day I upgraded [Sonarr](https://sonarr.tv/) and irkingly, it stopped moving completed files. Sure enough, the next time I logged onto the computer via screen, I had a new modal asking whether to grant Sonarr access to `~/Downloads`.

Last year I upgraded the server to macOS Ventura, and the script I used that ran a nightly back up stopped working, spewing a long stream of permission errors:

```
2024/03/04 03:00:01 ERROR : : failed to open directory "": open /Volumes/b-archive: operation not permitted
2024/03/04 03:00:01 ERROR : Local file system at /Volumes/backup-b-archive: not deleting files as there were IO errors
2024/03/04 03:00:01 ERROR : Local file system at /Volumes/backup-b-archive: not deleting directories as there were IO errors
2024/03/04 03:00:01 ERROR : Attempt 3/3 failed with 3 errors and: not deleting files as there were IO errors
2024/03/04 03:00:01 INFO  :
Transferred:              0 B / 0 B, -, 0 B/s, ETA -
Errors:                 3 (retrying may help)
Elapsed time:         0.0s

2024/03/04 03:00:01 Failed to sync with 3 errors: last error was: not deleting files as there were IO errors
/Users/brandur/bin/backup-archives:69:in `empty?': Operation not permitted @ rb_dir_s_empty_p - /Volumes/d-archive (Errno::EPERM)
        from /Users/brandur/bin/backup-archives:69:in `dir_empty?'
        from /Users/brandur/bin/backup-archives:51:in `block in main'
        from /Users/brandur/bin/backup-archives:38:in `each'
        from /Users/brandur/bin/backup-archives:38:in `main'
        from /Users/brandur/bin/backup-archives:82:in `<main>'
```

Being lazy, I didn't bother debugging the problem for a long time, opting instead to run the backup every so often manually. But I finally looked into it, and although not a modal this time, it was the same idea.

Cron scripts are now denied disk access by default. The fix is to grant the **cron** executable itself full disk access so that its child processes will also have access to it. Go to **System Settings** → **Privacy & Security** and add `/usr/sbin/cron`. Use `Shift``⌘`+`G` in the file selection dialog to navigate to `/usr/sbin`.

It's always nice to see security shoring up, but macOS's plausibility as a server OS is moving steadily in the wrong direction.

{{FigureSingle "" "/photographs/fragments/modals-mysterious-macos-cron-failures/full-disk-access.png"}}

{{FigureSingle "" "/photographs/fragments/modals-mysterious-macos-cron-failures/cron.png"}}