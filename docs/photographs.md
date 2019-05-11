# Photographs

Photographs that should appear on the home and `/photos`
pages are added manually to Dropbox and into a YAML index,
but then resized and added to an S3 cache by the build
process.

The process works like this:

1. Make sure that photo is uploaded to Dropbox.
2. Append an entry to `content/photographs/_meta.toml`
   which includes a title, description, and Dropbox share
   URL.
    * It's easy to obtain a share URL by right clicking in
      Finder and selecting "Copy Dropbox Link".
    * Note when copying a Dropbox link, change the `?dl=0`
      at the end to a `?dl=1` ("dl" is "download").
3. Add and commit the entry in Git and push to GitHub.
4. The build process pulls the image down from Dropbox,
   resizes it into a number of sizes, then pushes them to
   an S3 bucket `brandur.org-photographs`.

## Marker files

To prevent the same images from being downloaded with every
build, "marker" files (`content/photographs/*.marker`) are
used to represent that the work has been done. Marker files
are stored in Git while JPGs are not (for size reasons).

The build process skips resizing for any photos that have
markers, but since it won't modify the Git repository
itself, markers still need to be committed manually every
so often.

To commit a marker:

1. **Make sure the build process has run successfully at
   least once in CI** (if you commit a marker before it
   gets a chance to run, no resized images will be uploaded
   to S3).

2. Make sure Graphics Magick is installed:

    ```
    brew install graphicsmagick
    ```

3. Run the build/resize process locally:

    ```
    make install && make build
    ```

4. Add and commit the resulting marker file (but not the
   general `.jpg`s which are `.gitignore`ed) to Git and
   push to GitHub.
