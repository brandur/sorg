+++
hook = "TODO."
location = "San Francisco"
published_at = 2019-05-15T16:54:25Z
title = "Building a Robust Live Reloader with Websockets and Go"
+++

For the last couple weeks I've been making a few upgrades
to the backend that generates this site
([previously][intrinsic], with an aim on rebuilding it to
be faster, more stable, and more efficient. Over the years
I've used [Hugo][hugo] for a few projects, and one of its
features that I fell in love with is its live reloading. If
you haven't seen this before, as a file changes and a build
is triggered, an open web browser is notified of the
connection and triggers a reload on the page. Here's a
video of it in action:

TODO: Video

I was never convinced just reading about it -- it doesn't
seem like a big deal to just ⌘-`Tab` over to the browser
and ⌘-`R` for a refresh -- but the first time you really
_try_ it it's hard not to get addicted. Its a minor quality
of life improvement, but it makes the writing experience
much more fluid. And where it's good for writing, it's
_wonderful_ for design, where it's common to tweak
properties one at a time by the hundreds to get everything
perfectly lined up.

I decided to try and implement the feature and was
pleasantly surprised by how easy it turned out to be -- the
libraries available for Go for the task turned out to be
robust, and were exposing encapsulating complicated
implementations as simple APIs. Browser-level technologies
like WebSockets are now reliable and ubiquitous enough to
lend themselves to an easy implementation with minimal fuss
-- just basic JavaScript without a heavy build pipeline.

Here's a short tour of the final design.

## Watching changes with fsnotify (#fsnotify)

### Saving files in Vim (#vim)

```
2019/05/21 11:49:32 event: "../content/4913": CREATE
2019/05/21 11:49:32 event: "../content/hello.md~": CREATE
2019/05/21 11:49:32 event: "../content/hello.md": RENAME
2019/05/21 11:49:32 event: "../content/hello.md": CREATE
2019/05/21 11:49:32 event: "../content/hello.md": CHMOD
2019/05/21 11:49:32 event: "../content/hello.md~": REMOVE
2019/05/21 11:49:33 event: "../content/hello.md": CHMOD
```

## Emitting changes via WebSocket (#websocket)

### Keeping connections alive (#alive)

XX

This implementation, although not particularly complicated,
ends up working out incredibly well. It's common for me to
shut down my build server with some frequency, and with
this code, the next time I restart it all background tabs
that I might've had open immediately find the new server
and start listening for changes again almost immediately,
even if it was down for hours.

## Black boxes (#black-boxes)

For me, building live reload reminded me of the important
of good implementations that are well-abstracted. Fsnotify
connects into one of three different OS-level APIs
depending on the operating system (`inotify`, `kqueue`, or
`ReadDirectoryChangesW`), but hides all that legwork behind
what amounts to one function and two channels:

``` go
watcher, err := fsnotify.NewWatcher()
...

err = watcher.Add("/tmp/foo")
...

for {
    select {
        case event := <-watcher.Events:
            ...

        case err := <-watcher.Errors:
            ...
    }
}
```

Likewise with WebSockets, the most basic implementation of
live reload is only five lines of code, despite the work
involved in getting a WebSocket open and connected. Making
that more robust is only a few dozen more lines.

[hugo]: https://gohugo.io/
[intrinsic]: /aws-intrinsic-static
