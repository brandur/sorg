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

The first piece of the puzzle is knowing when source files
change so that we can signal a reload. Luckily Go has a
great utility for this in [fsnotify][fsnotify] which hooks
into operating system primitives and notifies a program
over a channel when a change is detected. Basic usage looks
like this:

``` go
watcher, err := fsnotify.NewWatcher()
...

err = watcher.Add("./content")
...

for {
    select {
        case event := <-watcher.Events:
            log.Println("event:", event)
    }
}
```

When something in the `content` directory changes, a
message like this one is emitted:

```
2019/05/21 11:49:32 event: "./content/hello.md": WRITE
```

### Saving files in Vim, and the curious case of 4913 (#vim)

Now things are _almost_ that easy, but a few practical
considerations complicate things.

While saving a file in Vim, instead of an ideal single
write event being emitted, instead we see a long stream of
events like this:

```
2019/05/21 11:49:32 event: "./content/4913": CREATE
2019/05/21 11:49:32 event: "./content/hello.md~": CREATE
2019/05/21 11:49:32 event: "./content/hello.md": RENAME
2019/05/21 11:49:32 event: "./content/hello.md": CREATE
2019/05/21 11:49:32 event: "./content/hello.md": CHMOD
2019/05/21 11:49:32 event: "./content/hello.md~": REMOVE
2019/05/21 11:49:33 event: "./content/hello.md": CHMOD
```

And all of this for one save! What's going on here? Various
editors perform somewhat non-intuitive procedures to help
protect against various problems. Vim for example has a
concept called "backup" file that's created in case writing
a file fails midway and leaves a user with lost data [1].
Here's the full procedure involved in saving a file:

1. Test to see if the editor is allowed to create files in
   the target directory by creating a file named
   (somewhat arbitrarily) `4913`.

2. Move the original file (`hello.md`) to the backup file,
   suffixed by a tilde (`hello.md~`).

3. Write the new contents at the original filename
   (`hello.md`).

4. Copy the old permissions to the new file with chmod.

5. On successful execution of all of the above, remove the
   backup file `hello.md~`.

It's good to know that Vim has our back in preventing
corruption, but all these changes aren't particularly
friendly to our build loop because they'll likely to
trigger rebuilds for false positives that won't affect the
built result. I solved this ignoring certain filenames in
incoming events:

``` go
// Decides whether a rebuild should be triggered given some input
// event properties from fsnotify.
func shouldRebuild(path string, op fsnotify.Op) bool {
    base := filepath.Base(path)

    // Mac OS' worst mistake.
    if base == ".DS_Store" {
        return false
    }

    // Vim creates this temporary file to see whether it can write
    // into a target directory. It screws up our watching algorithm,
    // so ignore it.
    if base == "4913" {
        return false
    }

    // A special case, but ignore creates on files that look like
    // Vim backups.
    if strings.HasSuffix(base, "~") {
        return false
    }

    ...
}
```

### Hardening the build loop (#build-loop)

It's a nice feature to trigger a page reload as soon as
possible after a build finishes, so we'll trigger a build
immediately a non-ignored file changes. This presents a bit
of a problem though in that there may be additional changes
that arrive in close succession after the first one *while*
the build is still running. Time-to-reload is an important
feature, but we can't let it supersede correctness -- every
change needs to be captured to ensure that the final result
is correct according the current state of the source.

We'll cover this case by having two goroutines coordinate.
A ***watch*** goroutine watches for file system changes and
sends a signal to a ***build*** goroutine upon receiving
one. If however, the build is still ongoing when a new
change comes in, it will accumulate new events until being
signaled that the build completed, at which point it will
trigger a new one with the accumulated changes.

!fig src="/assets/images/live-reload/build-loop.svg" caption="Goroutines coordinating builds even across changes that occur during an active build."

Builds are fast (especially when they're incremental), so
usually only one change we're interested will come in at a
time, but in case many do, we'll rebuild until they've all
been accounted for. Multiple changes can be accumulate and
be pushed into a single build, so we'll also rebuild only
as many times as necessary instead of doing one per change.

The code to do that looks something like this (simplified
slightly for brevity):

``` go
for {
    select {
    case event := <-watchEvents:
        lastChangedSources := map[string]struct{}{event.Name: {}}

        if !shouldRebuild(event.Name, event.Op) {
            continue
        }

        for {
            if len(lastChangedSources) < 1 {
                break
            }

            // Start rebuild
            rebuild <- lastChangedSources

            // Zero out the last set of changes and start
            // accumulating.
            lastChangedSources = nil

            // Wait until rebuild is finished. In the meantime,
            // accumulate new events that come in on the watcher's
            // channel and prepare for the next loop.
        INNER_LOOP:
            for {
                select {
                case <-rebuildDone:
                    // Break and start next outer loop
                    break INNER_LOOP

                case event := <-watchEvents:
                    if !shouldRebuild(event.Name, event.Op) {
                        continue
                    }

                    if lastChangedSources == nil {
                        lastChangedSources = make(map[string]struct{})
                    }

                    lastChangedSources[event.Name] = struct{}{}
                }
            }
        }
    }
}
```

## Emitting changes via WebSocket (#websocket)

To get WebSocket support in the backend we'll use the
[Gorilla WebSocket][gorilla] package, another off-the-shelf
library that abstracts away all the hard work. Creating a
WebSocket connection is as simple as a single invocation on
an `Upgrader` object from the library:

``` go
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func handler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    ... Use conn to send and receive messages.
}
```

There's a little plumbing involved in the HTTP backend that
we'll skip over, but the important part is that the build
goroutine will use a [conditional variable][cond] to signal
the goroutines serving open WebSockets when a build
completes:

``` go
var buildCompleteMu sync.Mutex
buildComplete := sync.NewCond(&buildCompleteMu)

// Signals all open WebSockets upon the completion of a
// successful build
buildComplete.Broadcast()
```

Those goroutines will in turn pass the message along to
their clients as a JSON-serialized payload:

``` go
// A type representing the extremely basic messages that
// we'll be serializing and sending back over a websocket.
type websocketEvent struct {
    Type string `json:"type"`
}

for {
    select {
    case <-buildCompleteChan:
        err := conn.WriteJSON(websocketEvent{Type: "build_complete"})
        if err != nil {
            c.Log.Errorf("<Websocket %v> Error writing: %v",
                conn.RemoteAddr(), writeErr)
        }

    ...
}
```

!fig src="/assets/images/live-reload/signaling-rebuilds.svg" caption="The build goroutine broadcasting a completed rebuild to WebSocket goroutines that will message their clients."

### Client-side JavaScript (#client)

The browser API for WebSockets is dead simple -- involving
a `WebSocket` object and a single callback. Upon receiving
`build_complete` message from the server, we'll close the
WebSocket connection and reload the page.

Here's the simplest possible implementation:

``` js
var socket = new WebSocket("ws://localhost:5002/websocket");

socket.onmessage = function(event) {
  var data = JSON.parse(event.data);
  switch(data.type) {
    case "build_complete":
      // 1000 = "Normal closure" and the second parameter is a
      // human-readable reason.
      socket.close(1000, "Reloading page after receiving build_complete");

      console.log("Reloading page after receiving build_complete");
      location.reload(true);

      break;

    default:
      console.log(`Don't know how to handle type '${data.type}'`);
  }
}
```

### Keeping connections alive (#alive)

We want to keep the amount of JavaScript we write to a
minimum, but it'd be nice to make sure that client
connections are as robust as possible. In the event that a
WebSocket terminates unexpectedly, or the build server
restarts, they should try and reconnect so that the live
reload feature stays alive.

Here we use a WebSocket's `onclose` callback to set a
timeout that tries to reconnect after five seconds.
`onclose` is called even in the event of a connection
failure, so this code will continually try to connect until
it's successful.

``` js
function connect() {
  var socket = new WebSocket("ws://localhost:5002/websocket");

  socket.onclose = function(event) {
    console.log("Websocket connection closed or unable to connect; " +
      "starting reconnect timeout");

    // Allow the last socket to be cleaned up.
    socket = null;

    // Set an interval to continue trying to reconnect
    // periodically until we succeed.
    setTimeout(function() {
      connect();
    }, 5000)
  }

  ...
}

connect();
```

This implementation, although not particularly complicated,
ends up working very reliably. It's common for me to shut
down my build server with some frequency, and with this
code, the next time I restart it all background tabs that I
might've had open immediately find the new server and start
listening for changes again almost immediately, even if it
was down for hours.

## Black boxes (#black-boxes)

For me, building live reload reminded me of the important
of good implementations that are well-abstracted. Fsnotify
connects into one of three different OS-level APIs
depending on the operating system (`inotify`, `kqueue`, or
`ReadDirectoryChangesW`), and if you look at its
implementation, does quite a lot of legwork to make that
possible. But for us as the end user, it's all hidden
behind a couple function calls and two channels:

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

Likewise with WebSockets, the most basic client
implementation of live reload is about five lines of code,
despite the work involved in getting a WebSocket open and
connected. Making that more robust is only a few dozen more
lines.

[1] Vim's behavior with respect to backup files can be
tweaked with various settings like `backup` and
`writebackup`.

[cond]: https://golang.org/pkg/sync/#Cond
[fsnotify]: https://github.com/fsnotify/fsnotify
[gorilla]: https://github.com/gorilla/websocket
[hugo]: https://gohugo.io/
[intrinsic]: /aws-intrinsic-static
