+++
hook = "Using FFmpeg and High Efficiency Video Coding (HVEC/H.265) to produce tiny video files that play very well in QuickTime."
published_at = 2019-05-28T03:32:28Z
title = "Encoding H.265/HEVC for QuickTime with FFmpeg"
+++

This weekend I experimented a little with re-encoding video
as H.265/HEVC, now that the codec has good support on both
hardware and in Mac OS/iOS. The [FFmpeg wiki][ffmpegh265]
estimates that you should be able to produce video files of
similar video quality as H.264 that are about half the
size. That sort of space saving should be interesting to
everybody, but especially for those who carry videos around
on laptops, which still tend to come with SSDs of modest
size even in the advanced year of 2019.

I wanted to produce something that was QuickTime
compatible, which means complying with Apple's very
[specific rules][applerules] around encoding. It might seem
crazy to want to do this, but I still find that QuickTime
is in the unique position of being simultaneously the best
_and_ worst video player ever made. While its narrow
versatility makes it incredibly frustrating to work with,
there's never been another video player in history that
scrubs as well, or uses laptop CPU/battery as efficiently.

For the copy/pasters out there, here's the FFmpeg
invocation that worked well for me:

```
ffmpeg -c:v libx265 -preset fast -crf 28 -tag:v hvc1 -c:a eac3 -b:a 224k -i <source> <target>
```

Or for GPU-based encoding, which is much faster, but
produces a larger file size:

```
ffmpeg -vcodec hevc_videotoolbox -b:v 6000k -tag:v hvc1 -c:a eac3 -b:a 224k -i <source> <target>
```

## The gritty details (#details)

The invocations above are all you really need to know, but
I'll walk through some caveats and explanations for those
interested.

### CPU versus GPU (#cpu-gpu)

`libx265` encodes via the CPU while `hevc_videotoolbox`
uses the GPU. `libx265` produces a much smaller file size,
but `hevc_videotoolbox` runs significantly faster (5 to 10
times the speed on my box).

You might have noticed that for `hevc_videotoolbox` I
specified a target bitrate with `-b:v 6000k`. I tried to
get away without doing that, but found that regardless of
the selected video profile (`main` or `main10`), unless I
forced a higher bitrate, the output video quality would be
garbage. And when I say garbage, I mean _garbage_; as in
totally unwatchable, with visual artifacts and blurriness
everywhere. Specifying a target video bitrate works around
the problem, at the cost of producing an even larger final
video.

From what I can tell, the right compromise is to use
`libx265` for any videos that you want to keep around, but
for videos that you'd prefer encoded quickly and will
probably delete (say you want them for a single trip),
`hevc_videotoolbox` is perfectly fine. Output videos are
still small enough, and the significantly faster encoding
speeds mean that FFmpeg finishes far more quickly.

### Audio, Dolby Digital Plus, and AAC (#audio)

The samples above use `-c:a eac3`, which QuickTime seems to handle more easily than AAC.

QuickTime will play multi-channel AAC, but it's more finnicky, and a video with even a slightly "wrong" audio configuration won't be openable. I've had luck forcing 5.1 with `channel_layout`:

```
-filter_complex "channelmap=channel_layout=5.1" -c:a aac
```

If your input sources are only stereo anyway, or you never expect to watch the output video on anything but a stereo device (i.e., headphones, TV minus sound system), FFmpeg can trivially downmix to stereo AAC, which QuickTime will play wit no trouble:

```
-c:a aac -ac 2
```

### The `fast` preset

The `libx265` `preset` setting accepts the wide array of
adjectives `ultrafast`, `superfast`, `veryfaster`,
`faster`, `fast`, `medium`, `slow`, `slower`, `veryslow`,
and `placebo`. Specifying a faster setting means that
encoding speed will be preferred over file size. Both ends
of the spectrum have extreme diminishing returns.

I found that there was very little file size difference
between `fast`, `medium`, and `slow`, but some encoding
speed difference, so I just default to `-preset fast` for
everything.

### The `hvc1` tag

The argument `-tag:v hvc1` tags the video with `hvc1`,
which is purely for QuickTime's benefit. It allows this
Very Smart Player to recognize the fact that it will be
able to play the resulting file.

[1] I don't even want to admit how long I spent trying to
figure out why QuickTime wouldn't open my encoded files. I
combed through my video settings about a hundred times
before realizing that it wasn't the video Apple didn't
like, it was the 5.1 AAC.

[ffmpegh265]: https://trac.ffmpeg.org/wiki/Encode/H.265
[applerules]: https://developer.apple.com/documentation/http_live_streaming/hls_authoring_specification_for_apple_devices
