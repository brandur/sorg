# Images

Install ImageMagick and Mozjpeg via Homebrew:

    brew install imagemagick mozjpeg

## Export input and output paths

    export GMI=~/Pictures/photo-archive/2017-031-portland/L1010468.jpg
    export GMO=.
    export GMO=content/images/passages/001-portland

## Copying paths from finder

Right-click on an image and hold `Option`. The "Copy
<file>" option becomes "Copy <file> as Pathname".

## Resize for article/Nanoglyph hooks

    magick convert $GMI -resize 75x75^ -gravity center -extent 75x75 -quality 85 $GMO/hook.jpg
    magick convert $GMI -resize 150x150^ -gravity center -extent 150x150 -quality 85 $GMO/hook@2x.jpg

Note the `^` on `-resize` here which treats these numbers
as minimums.

## Resize for article images

    magick convert $GMI -resize 650x -quality 85 $GMO/${$(basename $GMI)/.${$(basename $GMI)##*.}/.jpg}
    magick convert $GMI -resize 1300x -quality 85 $GMO/${$(basename $GMI)/.${$(basename $GMI)##*.}/@2x.jpg}

Or resized for 3:2:

    magick convert $GMI -gravity center -crop 3:2 -resize 650x -quality 85 $GMO/${$(basename $GMI)/.${$(basename $GMI)##*.}/.jpg}
    magick convert $GMI -gravity center -crop 3:2 -resize 1300x -quality 85 $GMO/${$(basename $GMI)/.${$(basename $GMI)##*.}/@2x.jpg}

## Resize for fragment vistas

    magick convert $GMI -gravity center -crop 3:2 -resize 1024x -quality 85 $GMO/vista.jpg
    magick convert $GMI -gravity center -crop 3:2 -resize 2048x -quality 85 $GMO/vista@2x.jpg

# Resize for fragment images

    magick convert $GMI -resize 550x -quality 85 $GMO/${$(basename $GMI)/.${$(basename $GMI)##*.}/.jpg}
    magick convert $GMI -resize 1100x -quality 85 $GMO/${$(basename $GMI)/.${$(basename $GMI)##*.}/@2x.jpg}

Overflowing:

    magick convert $GMI -resize 650x -quality 85 $GMO/${$(basename $GMI)/.${$(basename $GMI)##*.}/.jpg}
    magick convert $GMI -resize 1300x -quality 85 $GMO/${$(basename $GMI)/.${$(basename $GMI)##*.}/@2x.jpg}

## Resize for Twitter cards

    magick convert $GMI -resize 1300x650^ -gravity center -extent 1300x650 -quality 85 $GMO/twitter@2x.jpg

## Resize for Passages/Nanoglyph images

Landscape:

    magick convert $GMI -gravity center -crop 3:2 -resize 1100x -quality 85 $GMO/${$(basename $GMI)/.${$(basename $GMI)##*.}/@2x.jpg}

Landscape wide/highlight:

    magick convert $GMI -gravity center -crop 3:2 -resize 1400x -quality 85 $GMO/${$(basename $GMI)/.${$(basename $GMI)##*.}/@2x.jpg}

Portrait:

    magick convert $GMI -auto-orient -gravity center -crop 2:3 -resize 1100x -quality 85 $GMO/${$(basename $GMI)/.${$(basename $GMI)##*.}/@2x.jpg}

Portrait wide/highlight:

    magick convert $GMI -auto-orient -gravity center -crop 2:3 -resize 1200x -quality 85 $GMO/${$(basename $GMI)/.${$(basename $GMI)##*.}/@2x.jpg}

Note we don't bother with a non-retina version because we
can't run Retina.JS.

Note that some systems like Mac OS actually understanding
the `@2x` suffix and will treat the image correctly.

## Identify

Look at EXIF information with:

    identify -verbose <file>

## Convert from HEIC to JPG and crop 3:2

    magick convert $GMI -gravity center -crop 3:2 +repage -quality 85 $(dirname $GMI)/${$(basename $GMI)/.HEIC/.jpg}

## Optimization

Use the wrapper script for `mozjpeg` and `pngquant` (works for JPG and PNG):

    scripts/optimize_image.rb <path>

JPGs with `mozjpeg`:

    brew install mozjpeg
    cjpeg -outfile <out> -optimize -progressive <in>

PNGs with `pngquant`:

    brew install pngquant
    pngquant --output <out> -- <in>

## Favicons

    # or .jpg as extension
    export GMI=content/images/favicon/favicon-2048.png
    export GMO=content/images/favicon/

    magick convert $GMI -resize 32x32^ $GMO/favicon-32.${GMI##*.}
    magick convert $GMI -resize 128x128^ $GMO/favicon-128.${GMI##*.}
    magick convert $GMI -resize 152x152^ $GMO/favicon-152.${GMI##*.}
    magick convert $GMI -resize 167x167^ $GMO/favicon-167.${GMI##*.}
    magick convert $GMI -resize 180x180^ $GMO/favicon-180.${GMI##*.}
    magick convert $GMI -resize 192x192^ $GMO/favicon-192.${GMI##*.}
    magick convert $GMI -resize 256x256^ $GMO/favicon-256.${GMI##*.}

    scripts/optimize_image.rb $GMO/*

### Nanoglyph

    # or .jpg as extension
    export GMI=content/images/favicon/nanoglyph-2048.png
    export GMO=content/images/favicon/

    magick convert $GMI -resize 32x32^ $GMO/nanoglyph-32.${GMI##*.}
    magick convert $GMI -resize 128x128^ $GMO/nanoglyph-128.${GMI##*.}
    magick convert $GMI -resize 152x152^ $GMO/nanoglyph-152.${GMI##*.}
    magick convert $GMI -resize 167x167^ $GMO/nanoglyph-167.${GMI##*.}
    magick convert $GMI -resize 180x180^ $GMO/nanoglyph-180.${GMI##*.}
    magick convert $GMI -resize 192x192^ $GMO/nanoglyph-192.${GMI##*.}
    magick convert $GMI -resize 256x256^ $GMO/nanoglyph-256.${GMI##*.}

    scripts/optimize_image.rb $GMO/*

### Passages

    # or .jpg as extension
    export GMI=content/images/favicon/passages-2048.png
    export GMO=content/images/favicon/

    magick convert $GMI -resize 32x32^ $GMO/passages-32.${GMI##*.}
    magick convert $GMI -resize 128x128^ $GMO/passages-128.${GMI##*.}
    magick convert $GMI -resize 152x152^ $GMO/passages-152.${GMI##*.}
    magick convert $GMI -resize 167x167^ $GMO/passages-167.${GMI##*.}
    magick convert $GMI -resize 180x180^ $GMO/passages-180.${GMI##*.}
    magick convert $GMI -resize 192x192^ $GMO/passages-192.${GMI##*.}
    magick convert $GMI -resize 256x256^ $GMO/passages-256.${GMI##*.}

    scripts/optimize_image.rb $GMO/*
