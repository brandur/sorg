---
title: Fast OSX Screen Lock
published_at: 2016-01-22T22:28:01Z
---

The fastest way to lock your screen a modern Mac is to use the key combination
`Ctrl` + `Shift` + `Power` [1].

Using an external PC keyboard is a more tricky because it probably won't have
either a power or eject button. To get one working, install
[Karabiner][karabiner] and remap Pause/Break to Power:

![Karabiner settings](/assets/fragments/fast-osx-lock/karabiner.png)

`Ctrl` + `Shift` + `Break` will now trigger the lock.

[1] `Power` used to be `Eject`, but then Macs started shipping without optical
drives.

[karabiner]: https://pqrs.org/osx/karabiner/
