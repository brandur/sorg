+++
hook = "Comparing the results of libjpeg and MozJPEG by optimizing the archive of image's from this very website."
published_at = 2020-04-15T02:15:37Z
title = "Development log: libjpeg / MozJPEG optimization shootout"
+++

A few weeks ago I decided to look into optimizing the file sizes of the images found on this website to save on bandwidth, and storage size in the Git repository. It's also just an easy thing to do, with the only reason I wasn't doing it already that I'd never bothered to look into it before.

I was referred to a project called [MozJPEG](https://github.com/mozilla/mozjpeg), which is based on [libjpeg-turbo](https://github.com/libjpeg-turbo/libjpeg-turbo), which is based on [libjpeg](https://en.wikipedia.org/wiki/Libjpeg). The turbo version of libjpeg is an independent project that takes advantage of SIMD instructions on common CPU architectures for more speed. All three implementations are API compatible, so as a user it doesn't make much difference.

MozJPEG differs from libjpeg in that it claims to be better optimized for the web:

> It's compatible with libjpeg API and ABI, and can be used as a drop-in replacement for libjpeg. MozJPEG makes tradeoffs that are intended to benefit Web use cases and focuses solely on improving encoding, so it's best used as part of a Web encoding workflow.

## By the numbers (#numbers)

Both libjpeg and MozJPEG seemed to do a good job based on a few samples, so I was generally okay taking its claim at face value and preferring its use in my builds, but I got curious. It's largely fine to just use MozJPEG in all situations, but libjpeg does have the slight advantage of being somewhat more ubiquitous.

Luckily, I had a perfect real-world sample to do the comparison: the images from this very site, none of which I'd made any effort to optimize over the years beyond choosing a sane-looking "quality" setting as a I exported them from ImageMagick or Pixelmator. I wrote a [tiny Ruby script](https://github.com/brandur/sorg/blob/860640e59ccd82d6d3f5f6bd59534bf28f0face4/scripts/compression_test.rb) to iterate each one, optimize it using both libjpeg and MozJPEG, and compare the results. A few minutes later it spit this out:

```
average libjpeg compression ratio: 0.4897232210914123
average mozjpeg compression ratio: 0.4113459253012328
```

(Here "compression ratio" means the size of the newly optimized image divided by the original size. So an optimized image of 100 kB that was originally 200 kB would have a ratio of `100.0 / 200 = 0.5`.)

**Takeaway no. 1:** Optimizing JPEGs _really_ works. My _average_ result with MozJPEG was an optimized image 40% the size of the original. I compared the before and after results by eyeball and had a hard time seeing any difference, so there's really no downside as far as I'm concerned.

**Takeaway no. 2:** MozJPEG's claim appears to be true. Its average compression ratio was 41% the size of the original, compared to 49% with libjpeg [1].

[1] I should add a brief disclaimer that this is a very broad comparison based on a single, realistic set of images, but (1) I didn't spend any time fine-tuning either implementation, preferring just to use defaults, and (2) I didn't make any effort to compare the quality of the results beyond an eyeball check that what both programs produced was sane.
