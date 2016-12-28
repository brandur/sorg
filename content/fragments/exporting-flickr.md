---
title: Exporting Flickr
published_at: 2016-12-28T20:22:04Z
hook: Moving photos from Flickr back to plain folders on disk.
---

I spent a few hours over the holiday break doing some consolidation of my
backups and putting in a few improvements. One item that's been on the back
burner for a while was to do a full dump of everything that I've ever stored in
Flickr.

About five years back I got more concerned about data portability, so started
storing all my photo archives as plain on disk folders instead of in the
database of any particular photography software. However, there was a time
before that when I was so enthralled with Apple products that I trusted my
canonical copies to an Aperture vault. In retrospect, we know how good an idea
that was [1], and for me the experience imparted a larger lesson about trusting
Apple products with my data.

Luckily, I'd been uploading everything to Flickr the whole time, and have
photography on there going back to 2006. The site's always served me well, but
given Yahoo's recent troubles, it seems wise to hedge my bets and spread some
other copies around.

I was initially afraid that I'd have to come up with some kind of scripted
solution to get my photos out, and was pleasantly surprised to find out that
Flickr's made image export trivial. Every album has a download arrow that will
collect its originals and send them to you bundled up in a zip file. It's fast
and easy; the files are even named according to your album's title so that it's
very little effort to get them organized after they're downloaded.

![Album download](/assets/fragments/exporting-flickr/album-download.png)

I've now got all my photos stored safely into a set of plain folders, and have
them syncing bidirectionally between Amazon Cloud Drive and every computer I
own with a small daemon script that wraps the magnificently functional [rclone]
utility.

But I'm not done with Flickr either. It's served me very well for a long time
now, and has proven remarkably resilient to the sorts of product defacements
that tend to occur when you're acquired by a much bigger and more bureaucratic
company. I'll probably still be uploading photos until its dying day.

[1] Apple discontinued Aperture in 2015.

[rclone]: http://rclone.org/
