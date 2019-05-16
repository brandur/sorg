package modulir

import (
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Public
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Listens for file system changes from fsnotify and pushes relevant ones back
// out over the rebuild channel.
//
// It doesn't start listening to fsnotify again until the main loop has
// signaled rebuildDone, so there is a possibility that in the case of very
// fast consecutive changes the build might not be perfectly up to date.
func watchChanges(c *Context, watchEvents chan fsnotify.Event, watchErrors chan error,
	rebuild chan map[string]struct{}, rebuildDone chan struct{}) {

	for {
		select {
		case event, ok := <-watchEvents:
			if !ok {
				c.Log.Infof("Watcher detected closed channel; stopping")
				return
			}

			c.Log.Debugf("Received event from watcher: %+v", event)
			lastChangedSources := map[string]struct{}{event.Name: {}}

			if !shouldRebuild(event.Name, event.Op) {
				continue
			}

			// The central purpose of this loop is to make sure we do as few
			// build loops given incoming changes as possible.
			//
			// On the first receipt of a rebuild-eligible event we start
			// rebuilding immediately, and during the rebuild we accumulate any
			// other rebuild-eligible changes that stream in. When the initial
			// build finishes, we loop and start a new one if there were
			// changes since. If not, we return to the outer loop and continue
			// watching for fsnotify events.
			//
			// If changes did come in, the inner for loop continues to work --
			// triggering builds and accumulating changes while they're running
			// -- until we're able to successfully execute a build loop without
			// seeing a new change.
			//
			// The overwhelmingly common case will be few files being changed,
			// and therefore the inner for almost never needs to loop.
			for {
				if len(lastChangedSources) < 1 {
					break
				}

				// Start rebuild
				rebuild <- lastChangedSources

				// Zero out the last set of changes and start accumulating.
				lastChangedSources = nil

				// Wait until rebuild is finished. In the meantime, accumulate
				// new events that come in on the watcher's channel and prepare
				// for the next loop..
			INNER_LOOP:
				for {
					select {
					case <-rebuildDone:
						// Break and start next outer loop
						break INNER_LOOP

					case event := <-watchEvents:
						if shouldRebuild(event.Name, event.Op) {
							if lastChangedSources == nil {
								lastChangedSources = make(map[string]struct{})
							}

							lastChangedSources[event.Name] = struct{}{}
						}
					}
				}
			}

		case err, ok := <-watchErrors:
			if !ok {
				c.Log.Infof("Watcher detected closed channel; stopping")
				return
			}
			c.Log.Errorf("Error from watcher:", err)
		}
	}
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Decides whether a rebuild should be triggered given some input event
// properties from fsnotify.
func shouldRebuild(path string, op fsnotify.Op) bool {
	base := filepath.Base(path)

	// Mac OS' worst mistake.
	if base == ".DS_Store" {
		return false
	}

	// Vim creates this temporary file to see whether it can write into a
	// target directory. It screws up our watching algorithm, so ignore it.
	if base == "4913" {
		return false
	}

	// A special case, but ignore creates on files that look like Vim backups.
	if strings.HasSuffix(base, "~") {
		return false
	}

	if op&fsnotify.Create != 0 {
		return true
	}

	if op&fsnotify.Remove != 0 {
		return true
	}

	if op&fsnotify.Write != 0 {
		return true
	}

	// Ignore everything else. Rationale:
	//
	//   * chmod: We don't really care about these as they won't affect build
	//     output. (Unless potentially we no longer can read the file, but
	//     we'll go down that path if it ever becomes a problem.)
	//
	//   * rename: Will produce a following create event as well, so just
	//     listen for that instead.
	//
	return false
}
