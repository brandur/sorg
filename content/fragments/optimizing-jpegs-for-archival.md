+++
hook = "Writing a wrapper script around MozJPEG to achieve ~80% compression on large JPEGs with little downside."
# image = ""
published_at = 2025-03-29T12:35:10-07:00
title = "Optimizing JPEGs with MozJPEG for local archival"
+++

Call me old fashioned, but I like to keep my photo collection as local files on disk rather than symbolic pointers in the cloud, or sent off to deep storage on large archival drives, neither of which I'm likely to ever look at again. It's nice having quick access to them that still works over a bad internet link or on an airplane.

It's a great system, but it's been getting more difficult as time goes by. My photo collection grows year by year, but Apple's hard drive sizes stay frozen circa 2012. I'm running the same 1 TB drive that I was five years ago, which is only incrementally larger than five years before that (and even the mizerly 1 TB is still a $200 upcharge over the default _512 GB_ that's somehow a thing that Apple sells in 2025).

Realistically, I know that I'll never look at the majority of these photos again, so I already prune the collections aggressively to keep only the highlights, but was looking for storage opportunities beyond that. Years ago I wrote about [optimizing JPEGs for this site using MozJPEG](/fragments/libjpeg-mozjpeg), and knowing that a lot of cameras produce suboptimally compressed JPEGs, realized there was a similar opportunity for archival.

I ended up writing a wrapper around around MozJPEG that saves about 80% of space compared to what comes out of my camera. Here's a sample run:

```

    $ optimize 001-ana-nuevo/*
    created: 001-ana-nuevo/2W4A6210.jpg (9.02MB -> 2.11MB / saved 77%)
    created: 001-ana-nuevo/2W4A6212.jpg (8.21MB -> 1.79MB / saved 78%)
    created: 001-ana-nuevo/2W4A6216.jpg (11.0MB -> 2.68MB / saved 76%)
    created: 001-ana-nuevo/2W4A6218.jpg (6.36MB -> 1.29MB / saved 80%)
    created: 001-ana-nuevo/2W4A6219.jpg (12.11MB -> 3.01MB / saved 75%)
    created: 001-ana-nuevo/2W4A6224.jpg (7.3MB -> 1.69MB / saved 77%)
    created: 001-ana-nuevo/2W4A6228.jpg (7.75MB -> 1.72MB / saved 78%)
    created: 001-ana-nuevo/2W4A6230.jpg (8.62MB -> 1.99MB / saved 77%)
    created: 001-ana-nuevo/2W4A6236.jpg (8.14MB -> 1.87MB / saved 77%)
    created: 001-ana-nuevo/2W4A6237.jpg (6.65MB -> 1.48MB / saved 78%)
    created: 001-ana-nuevo/2W4A6238.jpg (7.59MB -> 1.69MB / saved 78%)
    created: 001-ana-nuevo/2W4A6240.jpg (9.38MB -> 2.21MB / saved 76%)
    created: 001-ana-nuevo/2W4A6242.jpg (9.26MB -> 2.22MB / saved 76%)
    created: 001-ana-nuevo/2W4A6243.jpg (10.17MB -> 2.44MB / saved 76%)
    created: 001-ana-nuevo/2W4A6247.jpg (10.49MB -> 2.56MB / saved 76%)
    created: 001-ana-nuevo/2W4A6251.jpg (7.92MB -> 1.84MB / saved 77%)
    created: 001-ana-nuevo/2W4A6252.jpg (8.97MB -> 2.12MB / saved 76%)
    created: 001-ana-nuevo/2W4A6253.jpg (7.74MB -> 1.75MB / saved 77%)
    created: 001-ana-nuevo/2W4A6254.jpg (9.43MB -> 2.3MB / saved 76%)
    created: 001-ana-nuevo/2W4A6255.jpg (10.78MB -> 2.65MB / saved 75%)
    created: 001-ana-nuevo/2W4A6258-pups.jpg (9.13MB -> 2.22MB / saved 76%)
    created: 001-ana-nuevo/2W4A6259.jpg (10.46MB -> 2.55MB / saved 76%)
    created: 001-ana-nuevo/2W4A6260.jpg (8.54MB -> 2.04MB / saved 76%)
    created: 001-ana-nuevo/2W4A6262.jpg (10.3MB -> 2.59MB / saved 75%)
    created: 001-ana-nuevo/2W4A6266.jpg (8.81MB -> 2.19MB / saved 75%)
    created: 001-ana-nuevo/2W4A6267.jpg (9.64MB -> 2.31MB / saved 76%)
    created: 001-ana-nuevo/2W4A6268.jpg (9.83MB -> 2.33MB / saved 76%)
    created: 001-ana-nuevo/2W4A6269.jpg (8.93MB -> 2.14MB / saved 76%)
    created: 001-ana-nuevo/2W4A6271.jpg (7.38MB -> 1.74MB / saved 76%)
    created: 001-ana-nuevo/2W4A6272.jpg (7.19MB -> 1.68MB / saved 77%)
    created: 001-ana-nuevo/2W4A6283-water-fight.jpg (7.65MB -> 1.73MB / saved 77%)
    created: 001-ana-nuevo/2W4A6284.jpg (8.02MB -> 1.77MB / saved 78%)
    created: 001-ana-nuevo/2W4A6286.jpg (5.82MB -> 1.11MB / saved 81%)
    created: 001-ana-nuevo/2W4A6287.jpg (6.03MB -> 1.14MB / saved 81%)
```

I'm sure there's some subtle downside to the extra compression, but I've tried zooming all the way in on a couple samples before and after, and I can see differences right at the pixel level, but the optimized version isn't clearly worse to my eye.

My script's use-at-your-own-risk me-ware that I'm not publishing in any official sense, but [here it is for reference](https://gist.github.com/brandur/8a7a7c7870fce52bcf1ac0c34d66af30).

Some gotchas I ran into and which might save someone else time/trouble:

* The MozJPEG binary to compress JPEGs is called `cjpeg`. This is an old Linux style project, and naming the binary after the project would make things too easy and too obvious for users. Under the strict edicts of 1970s Unix philosophy, that's completely unacceptable.

* You might have multiple packages on your system providing `cjpeg`. Make sure you're using MozJPEG's because it offers much better compression than libjpeg or libjpeg-turbo. You can see here that my default `cjpeg` is *not* MozJPEG's:

``` sh
$ which cjpeg
/opt/homebrew/bin/cjpeg

$ ls -l /opt/homebrew/bin/cjpeg
lrwxr-xr-x@ 1 brandur  admin    36B Feb 10 11:45 /opt/homebrew/bin/cjpeg -> ../Cellar/jpeg-turbo/3.1.0/bin/cjpeg
```

* The original libjpeg `cjpeg` didn't support _reading_ JPEGs, only writing them, and would encourage you to read JPEGs with another binary called `djpeg` and pipe that into `cjpeg` (again, the wonders of Unix philosophy). You can do that with MozJPEG too, but DO NOT DO THAT! Piping will strip EXIF data, which [you shouldn't do](/fragments/stop-stripping-exif). Unlike libjpeg's version, MozJPEG's `cjpeg` does read JPEGs, so piping is not necessary.

* If you're writing to a new a file and then replacing the original after (which you probably should for safety), make sure to copy the original create/modify timestamps to the new file. The easiest way to do this is with `touch -r <original> <new`>`

``` txt
TOUCH(1)						      General Commands Manual							  TOUCH(1)

NAME
     touch – change file access and modification times

SYNOPSIS
     touch [-A [-][[hh]mm]SS] [-achm] [-r file] [-t [[CC]YY]MMDDhhmm[.SS]] [-d YYYY-MM-DDThh:mm:SS[.frac][tz]] file ...

DESCRIPTION
     The touch utility sets the modification and access times of
     files. If any file does not exist, it is created with default
     permissions.

     By default, touch changes both modification and access times.
     The -a and -m flags may be used to select the access time or
     the modification time individually.  Selecting both is
     equivalent to the default.  By default, the timestamps are set
     to the current time. The -d and -t flags explicitly specify a
     different time, and the -r flag specifies to set the times
     those of the specified file.  The -A flag adjusts the values
     by a specified amount.

     The following options are available:

     ...

     -r      Use the access and modifications times from the
             specified file instead of the current time of day.
```

Another approach would be to do away with JPEG completely and go to HEIC or WebP, but I'm still finding support for those a little spotty, and navigating them in a file browser feels slow because the compression takes longer to render. I'll check in on that again in a year or two.