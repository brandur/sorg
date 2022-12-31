+++
hook = "Converting Appletech™ to friendly formats easily usable on the web."
published_at = 2022-12-31T10:20:38-07:00
title = "Notes on iOS live photos with ffmpeg and <video>"
+++

After [a visit to the zoo](/atoms/girm632), I played around with converting iPhone "live" photos to short videos that could be uploaded to the web with reasonable ease. For anyone not familiar with the "live" feature, it's one in which iOS captures a short video around a photo to give it a more dynamic feel (think the moving newspaper photos from _Harry Potter_ movies).

Live photos are convenient compared to video because their video component is captured automatically on a normal shutter press. There's no mucking around with stop/start in video mode which makes the process fast and streamlined. It takes a little more storage, but you can always pick and choose what you want to keep retroactively, like keeping only the photo component and discarding the video. The process below also works fine with normal video so live photos aren't a requirement.

## Export from phone (#export)

Live photos normally export from phone to computer as a still like any other, but by selecting "Options ➝ Include All Photos Data" when sharing one, iOS will export the entire "bundle" which includes both a `.heic` still and `.mov`. The `.mov` retains a little more footage around the photo than what iOS normally shows when you touch and hold a live.

## Convert with ffmpeg (#ffmpeg)

If I'm going to be doing this on a semi-regular basis, a key requirement of the workflow is that it should require as little manual finnagling as possible. Ideally, no GUIs are harmed in the process, which takes us back to the old reliable programmable workhorse: ffmpeg.

Here's a sample of a conversion command that I ended up with:

> `ffmpeg -ss 00:00:00.50 -to 00:00:02.60 -i IMG_0347.MOV -filter:v "crop=in_w:in_w*(2/3):0:in_h-(in_w*(2/3)),scale=1200:-1" cougar-1.webm`

Explanation of each part:

* `-ss 00:00:00.50 -to 00:00:02.60`: Start at second 00.50 (that's half a second in) and end at 02.60. Cuts the less interesting start and end of the video. I guestimate bounds initially and refine them until they look good.

* `crop=in_w:in_w*(2/3):0:in_h-(in_w*(2/3))`: Apply a crop in the format of `(output_width, output_height, x_start, y_start)`. In this case we do some math based on `in_w` (input width) and `in_h` (input height) to crop to an aspect ratio of 3:2 with central gravity. This isn't strictly necessary, but I keep media in 3:2 for consistency across devices.

* `scale=1200:-1`: A second half of the video filter that scales output width to 1200 and `-1` which picks a height according to the video's aspect ratio. The video was cropped to 3:2 in the previous step so this will scale to `1200x800`.

* `-an`: Strip the audio track. We're going to make sure that videos are muted when rendered as HTML for reasons discussed below, but since there's nothing worth keeping in the audio anyway, just get rid of it.

* `cougar-1.webm`: ffmpeg will pick an output container of WebM based on the extension of the output filename, and will default to the VP9 codec for WebM which is what we'd want to use anyway.

### VP9 versus HEVC and iOS support (#vp9-hevc-ios)

WebM/VP9 are widely supported and theoretically so on iOS versions of Safari, but I couldn't easily get them working there [1], so I opted to export a second video file encoded with MP4/HEVC instead which is well-supported by Apple.

> `ffmpeg -ss 00:00:00.50 -to 00:00:02.60 -i IMG_0347.MOV -filter:v "crop=in_w:in_w*(2/3):0:in_h-(in_w*(2/3)),scale=1200:-1" -c:v libx265 -tag:v hvc1 -an cougar-1.mp4`

It's mostly the same as the above, with a couple tweaks:

* Add `-c:v libx265` to encode as x265/HEVC.
* Add [the magic tag](/fragments/ffmpeg-h265#hvc1-tag) `-tag:v hvc1` to hint to Apple that it'll be able to play the video.
* Change the extension to `.mp4` to use a more conventional Apple-friendly container.

### WebP (#webp)

WebP is another modern format that's WebM's counterpart for images instead of video, and has the benefit of being compatible with an `<img>` tag. Out of curiosity I tried exporting to it as well since it can support animation. ffmpeg supports it easily by just changing the target extension from `.webm` to `.webp`, but I didn't end up using it because the files produced were ~5x the size, presumably lacking some of the video specific size optimizations that `.webm` includes.

If you do encode to WebP and want something like an animated GIF, also consider including the option `-loop 0` which will prompt the WebP to loop infinitely after it finishes playing.

## The &lt;video&gt; tag (#video)

With encoding finished, it's easy to add the product to an HTML5 `<video>` tag:

``` html
<video autoplay loop muted playsinline>
    <source src="/videos/atoms/girm632/cougar-1.mp4" type="video/mp4">
    <source src="/videos/atoms/girm632/cougar-1.webm" type="video/webm">
</video>
```

<video autoplay loop muted playsinline>
    <source src="/videos/atoms/girm632/cougar-1.mp4" type="video/mp4">
    <source src="/videos/atoms/girm632/cougar-1.webm" type="video/webm">
</video>

We add `autoplay` and `loop` so that the short videos behave like a classic animated GIF. Users can right-click on them to show controls and pause if they're too annoying.

Note that despite stripping the audio track above we still include the special keyword `muted`. This is because browsers like Chrome won't allow a video to autoplay _unless_ it's muted, which seems like a fair compromise to help us all retain our sanity while browsing the web.

Lastly, `playsinline` is a magic Apple keyword that seems to be required to have the video start on an iPhone specifically (it seemed to work without on iPad, but it's hard to say anything for sure on Apple platforms given the absence of a hard refresh capability).

[1] There may be a trick to getting WebM/VP9 working on iOS, but I timed out trying to find it.