# TODO

* [ ] Finish building out "about" and various other one-off pages.
* [ ] Move to a leading slash system in paths.
* [ ] Don't use so many path constants ... it's not really helping.
* [ ] Fix logging: debug is currently too verbose for Travis runs, but normal
  is not verbose enough.
* [ ] Test for the majority of functions in build's main.go.
* [ ] Download images found in fragment frontmatter (maybe just stored them
  locally).
* [ ] Smarter asset symlinking that doesn't remove and create every time.
* [ ] Remove orphaned files between builds.
* [ ] Speed up the build (maybe incremental?).

## Done

* [x] Render article and fragment drafts in development.
* [x] Section headers in build's main.go; it's getting hard to track.
* [x] Change conf DB checks to nil checks on `db`.
* [x] Procfile rebuild on *.go file changes.
