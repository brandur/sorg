+++
hook = "A walkthrough of the design of a live reload feature for the static site generator that builds this site, touching on fsnotify, WebSockets, and the curious case of file 4913."
location = "San Francisco"
published_at = 2019-05-15T16:54:25Z
title = "Building a Robust Live Reloader with WebSockets and Go"
+++

For the last couple weeks I've been making a few upgrades
to the backend that generates this site
([previously][intrinsic]), with an aim on rebuilding it to
be faster, more stable, and more efficient. The source is
custom, and it'd accumulated enough cruft over the years
through incremental augmentations to justify a facelift.

I'd recently used I've used [Hugo][hugo] for a few
projects, another static site generate well-known for being
one of the first written in Go, and fell in love with one
of its features: live reloading. If you haven't seen it
before, as a file changes in development mode and a build
is triggered, live reload signals any web browsers open to
the site to reload.  Here's a video of it in action:

<figure>
  <p>
    <video controls class="overflowing">
      <source src="/assets/images/live-reload/live-reload.h264.mp4" type="video/mp4">
    </video>
  </p>
  <figcaption>A short video of the live reload feature in
    action: changes saved in the editor show up immediately 
    in the browser.</figcaption>
</figure>

It's hard to be convinced just reading about it -- it
doesn't seem like a big deal to just ⌘-`Tab` over to the
browser and ⌘-`R` for a refresh -- but the first time you
try it, it's hard not to get addicted. Its a tiny quality
of life improvement, but one that makes the writing
experience much more fluid. And where it's good for
writing, it's _wonderful_ for design where it's common to
make minor tweaks to CSS properties one at a time by the
_hundreds_ to get everything looking exactly right.

I decided to write my own implementation of the feature and
was pleasantly surprised by how easy it turned out to be.
The libraries available for Go to use as primitives were
robust, and nicely encapsulated complicated implementations
into simple APIs. Browser-level technologies like
WebSockets are now reliable and ubiquitous enough to lend
themselves to an easy implementation with minimal fuss --
just a few lines of basic JavaScript. No transpiling, no
polyfills, no heavy build pipeline, no mess.

Here's a short tour of the design.

## Watching for changes with fsnotify (#fsnotify)

The first piece of the puzzle is knowing when source files
change so that we can signal a reload. Luckily Go has a
great library for this in [fsnotify][fsnotify], which hooks
into operating system monitoring primitives and notifies a
program over a channel when a change is detected. Basic
usage is as simple as adding directories to a watcher and
listening on a channel:

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

When something in the `content` directory changes, the
program emits a message like this one:

```
2019/05/21 11:49:32 event: "./content/hello.md": WRITE
```

### Saving files in Vim, and the curious case of 4913 (#vim)

Now things are _almost_ that easy, but a few practical
considerations complicate things a little.

While saving a file in Vim (for example, but other editors
may behave similarly), instead of an ideal single write
event being emitted, instead we see a long stream of events
like this:

```
2019/05/21 11:49:32 event: "./content/4913": CREATE
2019/05/21 11:49:32 event: "./content/hello.md~": CREATE
2019/05/21 11:49:32 event: "./content/hello.md": RENAME
2019/05/21 11:49:32 event: "./content/hello.md": CREATE
2019/05/21 11:49:32 event: "./content/hello.md": CHMOD
2019/05/21 11:49:32 event: "./content/hello.md~": REMOVE
2019/05/21 11:49:33 event: "./content/hello.md": CHMOD
```

And all of this for one save! What could possibly be going
on here? Well, various editors perform some non-intuitive
gymnastics to help protect against edge failures. What
we're seeing here is a Vim concept called a "backup file"
that's created to protect against the possibility that
writing a change fails midway and leaves a user with lost
data [1]. Here's Vim's full procedure in saving a file:

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
friendly to our build loop because they'll trigger rebuilds
for changes that won't affect the built result. I solved
this ignoring certain filenames in incoming events:

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

Special-casing byproducts of known editors isn't the most
elegant possible solution, but it's pragmatic choice. The
build would still work fine without the special cases, but
it'd be a little less efficient. Also, the pace of newly
created editors isn't _so_ frantic so we won't be able to
keep up with them.

### Hardening the build loop (#build-loop)

It's a nice feature to trigger a page reload as soon as
possible after a build finishes, so the build loop will
start immediately on changes to non-ignored files. This
introduces a bit of a problem in that there may be
additional changes that arrive in close succession after
the first one *while* the build is still running.
Time-to-reload is an important feature, but we can't let it
supersede correctness -- every change needs to be captured
to ensure that the final result is correct according the
current state of the source.

We'll cover that case by having two goroutines coordinate.
A ***watch*** goroutine watches for file system changes and
sends a signal to a ***build*** goroutine upon receiving
one. If however, the build is still ongoing when a new
change comes in, it will accumulate new events until being
signaled that the build completed, at which point it will
trigger a new one with the sum of the accumulated changes.

!fig src="/assets/images/live-reload/build-loop.svg" caption="Goroutines coordinating builds even across changes that occur during an active build."

Builds are fast (especially when they're incremental), so
usually only one change we're interested will come in at a
time, but in case many do, we'll rebuild until they've all
been accounted for. Multiple accumulated changes can be
pushed into a single build, so we'll also rebuild as many
times as possible instead of once per change.

The watcher code with an accumulating inner loop looks
something like this (simplified slightly for brevity):

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

## Signaling with WebSockets (#websockets)

To get WebSocket support in the backend we'll use the
[Gorilla WebSocket][gorilla] package, another off-the-shelf
library that abstracts away a lot of gritty details.
Creating a WebSocket connection is as simple as a single
invocation on an `Upgrader` object from the library:

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
completes. Unlike the much more common channel primitive, a
conditional variable allows a single controller to signal
any number of waiting consumers that a change occurred.

``` go
var buildCompleteMu sync.Mutex
buildComplete := sync.NewCond(&buildCompleteMu)

// Signals all open WebSockets upon the completion of a
// successful build
buildComplete.Broadcast()
```

Those goroutines will in turn pass the signal along to
their clients as a JSON-serialized message:

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

The browser API for WebSockets is dead simple -- a
`WebSocket` object and a single callback. Upon receiving
`build_complete` message from the server, we'll close the
WebSocket connection and reload the page.

Here's the minimum viable implementation:

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
failure, so this code will continually try to reconnect
until either its tab is closed, or it's successful.

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

  socket.onmessage = function(event) {
    ...
  }
}

connect();
```

This implementation, although still quite simple, ends up
working very reliably. It's common for me to shut down my
build server as I'm changing Go code in the backend, and
with these few extra lines for resilience, the next time I
restart it all background tabs that I had open immediately
find the new server and start listening again almost
immediately. The server could've been down for hours (or
days!) and it still works just fine.

## Black boxes and solid foundations (#black-boxes)

Building live reloading reminded me of the importance of
good foundational layers that are well-abstracted. Fsnotify
connects into one of three different OS-level monitoring
APIs depending on the operating system (`inotify`,
`kqueue`, or `ReadDirectoryChangesW`), and if you look at
its implementation, does quite a lot of legwork to make
that possible. But for us as the end user, it's all hidden
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

None of the package's underlying complexity leaks into my
program, which leaves me with a lot less to worry about.

Likewise with WebSockets, the most basic client
implementation of live reload is about five lines of code,
despite the behind-the-scenes work involved in getting a
WebSocket open and connected. This is exactly what the road
to reliable software looks like: layering on top of **black
boxes** that expose a minimal API and whose walls are
opaque -- they can be expected to "just" work, so there's
no need to think too hard about what's inside them.

[1] Vim's behavior with respect to backup files can be
tweaked with various settings like `backup` and
`writebackup`.

[cond]: https://golang.org/pkg/sync/#Cond
[fsnotify]: https://github.com/fsnotify/fsnotify
[gorilla]: https://github.com/gorilla/websocket
[hugo]: https://gohugo.io/
[intrinsic]: /aws-intrinsic-static
