# TODO

* [ ] Change conf DB checks to nil checks on `db`.
* [ ] Render article and fragment drafts in development.
* [ ] Finish building out "about" and various other one-off pages.
* [ ] Fix logging: debug is currently too verbose for Travis runs, but normal
  is not verbose enough.
* [ ] Test for the majority of functions in build's main.go.
* [ ] Download images found in fragment frontmatter.
* [ ] Smarter asset symlinking that doesn't remove and create every time.
* [ ] Remove orphaned files between builds.
* [ ] Speed up the build (maybe incremental?).

## Done

* [x] Procfile rebuild on *.go file changes.
