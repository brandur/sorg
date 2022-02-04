+++
hook = "A Hugo shortcode that renders retina and non-retina image assets automatically."
published_at = 2022-02-04T20:08:06Z
title = "A retina asset shortcode for Hugo"
+++

I've been focusing a lot on documentation and documentation quality lately, and one of the items that came across my plate was getting our image assets up to retina-level resolution.

In this blog I use a [pretty exotic pipeline](/fragments/static-site-asset-management) that produces good results (no need to put large files in the repo, final artifacts are highly optimized for size), but with a lot of other potential contributors at work, I needed something much less abrasive.

I settled on a [Hugo shortcode](https://gohugo.io/content-management/shortcodes/). Given how many of these Hugo provides out-of-the-box, I was a little surprised that what I wanted doesn't seem to already exist in some form, but luckily it does have a pretty feature-ful image asset pipeline built-in, so it wasn't hard to layer one in on top.

The shortcode's source:

``` html
{{HTMLSafePassThrough `
{{ $originalImage := resources.Get (.Get 0) }}

{{ $targetWidth1x := (.Get 1) }}
{{ $targetWidth2x := mul ($targetWidth1x) 2 }}

{{ $image1x := $originalImage.Resize (printf "%dx" $targetWidth1x) }}
{{ $image2x := $originalImage.Resize (printf "%dx" $targetWidth2x) }}

{{ $caption := (.Get 2) }}

<a href="{{ $image2x.RelPermalink }}" title="{{ $caption }}">
    <img src="{{ $image1x.RelPermalink }}" srcset="{{ $image2x.RelPermalink }} 2x, {{ $image1x.RelPermalink }} 1x"
        alt="{{ $caption }}">
</a>
`}}
```

And to use it:

``` html
{{HTMLSafePassThrough `
{{< imgsrcset "images/api-keys.png" 950 "API keys in account settings" >}}
`}}
```

I wrote a short guide for other people, saying how to choose paths and numbers:

1. Take a screenshot and place it in `assets/images/`.
2. Look at the image's width and round it to a nice evenly-divisible-by-two number. Divide it by two and it becomes your non-retina width. e.g. a screenshot is `1932` wide, round down to `1900`, divide by two to `950`.
3. Put asset path, non-retina width, and a caption into `<imgsrcset>`.

Not too hard, but there's a lot of shortcuts taken and copy/pasta that goes on at company, so we'll see if it turns out to be sustainable over the longer term. I get why people just use hosted CMSes instead, but I'd defend our choice of Hugo for docs as good, because although the technical bar is a little higher, it makes small content changes incredibly fast with a text/Git-based workflow, and it's allowed me to push through a total redesign full multiples faster than it would've taken on an overcomplicated CMS platform.
