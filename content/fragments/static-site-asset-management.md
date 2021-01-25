+++
hook = "Automatic generation, optimization, and deployment of large static assets, all managed from a TOML file."
published_at = 2021-01-25T01:07:01Z
title = "Large asset management in a Git-based static site"
+++

I've been a vocal proponent of [Git-based static sites](/aws-intrinsic-static) for a long time. Not only are they fast, versioned, and geo-distributed, but with the right framework, have an amazing development workflow. I run everything from my text editor, and have some elaborate bells and whistles like [live reloading](/live-reload), a fast incremental build loop that takes on average ~3 ms, and automatic deployment on Git merges. Deployment is a `git push origin master` and it's great for accepting corrections by way of pull requests.

One piece it took a while to get right was the management of non-text static assets, especially where those assets are large files like photographs. For years I just committed them straight to Git, which worked well enough, except that it was ballooning my repository to ever greater proportions, which has culminated in today's impressive standing at 682 MB. Luckily, GitHub is generous with space, but my runway was bound to run out eventually, especially given that I've gotten in the expensive habit of including high resolution photographs with newsletters. I needed a better way, but one that didn't compromise on workflow.

I had a brief stint with Git [LFS](https://git-lfs.github.com/) (Large File Storage), which GitHub conveniently supports, but found that overall it added too much configuration and confusion to the project, so rolled it back.

I settled on a custom system in [Modulir](https://github.com/brandur/modulir), my homegrown CMS. All original assets stay in Dropbox, where I keep everything anyway, and as they're needed, I generate a link for them with a right-click in Finder, and copy it into [a TOML file](https://github.com/brandur/sorg/tree/master/content/photographs/_other_meta.toml):

``` toml
#
# size conventions:
#
#   * landscape: 1200 wide
#   * portrait: 900 wide
#

#
# /nanoglyphs/018-ractors
#

[[photographs]]
  crop_width = 1200
  original_image_url = "https://www.dropbox.com/s/tvz4t4ggikrld6x/2W4A0178-cropped.jpg?dl=1"
  slug = "nanoglyphs/018-ractors/engelmann-plaque"
  title = "Plaque on Sulphur Mountain showing an Engelmann Spruce"

[[photographs]]
  crop_width = 1200
  original_image_url = "https://www.dropbox.com/s/yvhbe0c6wsydc4f/2W4A0150.JPG?dl=1"
  slug = "nanoglyphs/018-ractors/sulphur-gondola"
  title = "The gondola at Sulphur Mountain"

[[photographs]]
  crop_width = 1200
  original_image_url = "https://www.dropbox.com/s/0cf7x6o1ccew6qi/2W4A0161.JPG?dl=1"
  slug = "nanoglyphs/018-ractors/sulphur-view"
  title = "The view off Sulphur Mountain"
```

Each entry creates a parallel Modulir job which downloads the file, passes it through ImageMagick to resize it retina and non-retina proportions, then [through MozJPEG](/fragments/libjpeg-mozjpeg) or pngquant (depending on whether JPG or PNG) to optimize file size. The TOML file's watched through [fsnotify](https://github.com/fsnotify/fsnotify), so the only thing I need to do to kick off the process is hit the save button. When running in CI, a GitHub Action uploads everything to S3.

To avoid having to resize everything over and over again, on success the system leaves an empty `.marker` file in place to signal to the build system that the work is already done:

``` sh
$ ls -lh content/photographs/nanoglyphs/018-ractors
.rw-r--r--  brandur    211 KB  Jan 16 23:03:57 2021  engelmann-plaque.jpg
.rwxr-xr-x  brandur      0 B   Jan 16 23:03:59 2021  engelmann-plaque.marker
.rw-r--r--  brandur  661.7 KB  Jan 16 23:03:59 2021  engelmann-plaque@2x.jpg
.rw-r--r--  brandur  255.5 KB  Jan 16 23:03:58 2021  sulphur-gondola.jpg
.rwxr-xr-x  brandur      0 B   Jan 16 23:04:01 2021  sulphur-gondola.marker
.rw-r--r--  brandur  834.5 KB  Jan 16 23:04:01 2021  sulphur-gondola@2x.jpg
.rw-r--r--  brandur  169.6 KB  Jan 16 23:03:57 2021  sulphur-view.jpg
.rwxr-xr-x  brandur      0 B   Jan 16 23:04:00 2021  sulphur-view.marker
.rw-r--r--  brandur  568.2 KB  Jan 16 23:04:00 2021  sulphur-view@2x.jpg
```

I commit outstanding `.marker`s to Git every few weeks after giving the resizes a chance to run in CI and on the couple different computers where I run Modulir. Assets won't be present on new machines with a fresh Modulir install, so for that case I have a make task to sync the archive down from S3:

``` make
.PHONY: photographs-download
photographs-download:
ifdef AWS_ACCESS_KEY_ID
	aws s3 sync s3://$(PHOTOGRAPHS_S3_BUCKET)/ content/photographs/
else
	# No AWS access key. Skipping photographs-download.
endif
```

It's been a major boon for automation because although I had canned ImageMagick invocations at the ready, creating new directories and copying and pasting them into the terminal still took a few seconds. Duplicating some TOML and a simple `:w` from Vim is faster and easier. Highly recommended.
