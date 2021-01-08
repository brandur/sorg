+++
hook = "Converting Mutelight, my ancient blog, from a custom, problematically undermaintained Ruby stack over to Modulir, so it can live on as a static site."
published_at = 2021-01-08T18:37:58Z
title = "Development log: Rescuing history"
+++

I have an ancient blog over at [mutelight.org](https://mutelight.org) which was the predecessor to this site. It spans the first ~5 years of my professional career, and it's full of bad ideas. I talked a lot about .NET, and thought that Ruby was a gift from god almighty to save us from heavy enterprise frameworks. I'd forgotten almost everything I wrote on there, and was entertained by [this article](https://mutelight.org/building-a-command-line-environment-for-net-development-with-rake), which suggests adding a Ruby dependency to your .NET project so you can use Rake. Ahead of its time really.

For the last ten years, Mutelight's been slowly but surely sinking beneath the waves. It ran on one my old Ruby-based blogging engines called Hekla, and thanks to Heroku's platform it _still_ runs, but the writing was on the wall. The app is on the `cedar-14` stack, which was put into readonly last year. Before I archived the project, GitHub showed half a dozen major vulnerabilities in dependent gems. I barely remember how to even go about updating the thing anymore.

But I like the idea of keeping URLs around and unbroken as long as possible ([cool URIs don't change](https://www.w3.org/Provider/Style/URI)), so I decided to mount a rescue operation. I spent last night converting it over to a static site built by [Modulir](https://github.com/brandur/modulir), the same framework that builds this one. The conversion involved:

* Reformatting source articles from `.md` + `.rb` for metadata (!!!) to `.md` with TOML frontmatter.
* Converting from SASS to plain old CSS.
* Converting from [Slim](https://github.com/slim-template/slim) to [ACE](https://github.com/yosssi/ace).
* Finding a static-site scheme that wouldn't break the built-in "tiny URLs" that I'd so wisely added back in 2012 when URL shorteners were all the rage.
* Configuring S3/IAM/CloudFront and getting a build pipeline in place to deploy from GitHub Actions.
* Unlinking Flickr URLs (I don't expect those to be around too much longer) and importing those photos into the project. For good measure, running all images through [MozJPEG](/fragments/libjpeg-mozjpeg) to get sizes down.
* Giving the styling an ever-so-slight facelift. Eliminated Google Fonts, increased text size, and added a smidgeon of additional whitespace.

The job is done, and it only took one evening. Mutelight is now a static site that has a good chance of outliving Heroku, and maybe even Ruby. It's also noticeably faster now that there's no dynamic component. It's still on a NIH stack, but one that has minimal dependencies aside from Go's standard library. And even if Modulir dies, the static files that it generated will be portable to any system for about as long as we can expect modern computing to be alive.

Chalk this up to another of "probably not worth the time" lockdown project, but hey, it beats Netflix.
