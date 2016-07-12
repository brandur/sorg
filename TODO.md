# TODO

* [ ] Rewrite about page.
* [ ] Test for the majority of functions in build's main.go.
* [ ] Download images found in fragment frontmatter (maybe just stored them
  locally).
* [ ] Remove orphaned files between builds.
* [ ] Faster build by passing changes paths to program.

## Done

* [x] Speed up the build (maybe incremental?).
* [x] Fix logging: debug is currently too verbose for Travis runs, but normal
  is not verbose enough.
* [x] Smarter asset symlinking that doesn't remove and create every time.
* [x] Move to a leading slash system in paths.
* [x] Don't use so many path constants ... it's not really helping.
* [x] Atom feeds.
* [x] Make sure that all viewport widths have been accounted for.
* [x] Finish building out "about" and various other one-off pages.
* [x] Figure out a schema for talks/about pages title.
* [x] Render article and fragment drafts in development.
* [x] Section headers in build's main.go; it's getting hard to track.
* [x] Change conf DB checks to nil checks on `db`.
* [x] Procfile rebuild on *.go file changes.
