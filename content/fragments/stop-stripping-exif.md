+++
hook = "Most EXIF metadata is benign, and occasionally of nominal interest to somebody. Try leaving it in."
published_at = 2023-10-04T12:50:00+02:00
title = "Being a good web denizen: Don't strip EXIF metadata from photos"
+++

A few days ago after posting photos from my [trip on the John Muir Trail](/john-muir-trail), a colleague asked what camera I was shooting with. It's an easy question to ask with a direct line over Slack, but also one that's quick to answer using a program like `ImageMagick`'s `convert` or any number of other tools that can read EXIF tags:

``` sh
$ identify -format '%[EXIF:*]' john-muir-trail-1_large@2x.jpg
exif:ApertureValue=393216/65536
exif:ExposureTime=1/320
exif:FNumber=8/1
exif:FocalLength=72/1
exif:LensModel=RF24-105mm F4 L IS USM
exif:LensSerialNumber=9234003630
exif:LensSpecification=24/1, 105/1, 0/1, 0/1
exif:Make=Canon
exif:Model=Canon EOS R6
...
```

Photos on this site go through a comprehensive pipeline before being published, being resized with ImageMagick and optimized with MozJPEG to maximize web-friendliness. But there's nothing special to leave EXIF in -- both these programs will do the right thing by default and preserve properties like camera model, lens, f-stop, and shutter speed, all of which another photographer might find interesting. And beyond my specific CLI tools, leaving EXIF data in place tends to overwhelming default -- the desktop publishing programs I use like Pixelmator do it as well.

To my dismay, I find that so many sites online _including_ those run by self-described photographers strip out this information, and my guess is that it isn't for a particularly good reason -- probably because they haven't thought about it or are cargo culting copy/pasted commands than anything else. There are some sensitive EXIF tags that you'd conceivably want to remove like serial numbers or GPS coordinates, but programs like `exiftool` can [cauterize those on a per-tag basis](https://kevinchen.co/blog/scrub-sensitive-photo-exif-metadata-with-exiftool/).

So in the same vein of being a good web denizen like [stop truncating RSS](/fragments/stop-truncating-rss): leave your EXIF metadata in! Most people will never know it's there, but it leaves the web a little richer, and is occasionally useful for someone who cares.
