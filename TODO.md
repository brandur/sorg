# TODO

* [ ] Render article and fragment drafts in development.
* [ ] Finish building out "about" and various other one-off pages.
* [ ] Fix logging: debug is currently too verbose for Travis runs, but normal
  is not verbose enough.
* [ ] Test for the majority of functions in build's main.go.
* [ ] Download images found in fragment frontmatter (maybe just stored them
  locally).
* [ ] Smarter asset symlinking that doesn't remove and create every time.
* [ ] Remove orphaned files between builds.
* [ ] Speed up the build (maybe incremental?).

## Done

* [x] Section headers in build's main.go; it's getting hard to track.
* [x] Change conf DB checks to nil checks on `db`.
* [x] Procfile rebuild on *.go file changes.
