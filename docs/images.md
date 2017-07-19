# Images

Homebrew:

    brew install graphicsmagick

Documentation:

http://www.graphicsmagick.org/convert.html

## Export input and output paths

    export GM_INPUT=~/Pictures/photo-archive/2017-031-portland/L1010468.jpg
    export GM_OUTPUT=.
    export GM_OUTPUT=content/images/passages/001-portland

## Copying paths from finder

Right-click on an image and hold `Option`. The "Copy
<file>" option becomes "Copy <file> as Pathname".

## Resize for article hooks

    gm convert $GM_INPUT -resize 75x75^ -gravity center -extent 75x75 -quality 85 $GM_OUTPUT/hook.jpg
    gm convert $GM_INPUT -resize 150x150^ -gravity center -extent 150x150 -quality 85 $GM_OUTPUT/hook@2x.jpg

Note the `^` on `-resize` here which treats these numbers
as minimums.

## Resize for article images

    gm convert $GM_INPUT -resize 650x -quality 85 $GM_OUTPUT/$(basename $GM_INPUT)
    gm convert $GM_INPUT -resize 1300x -quality 85 $GM_OUTPUT/${$(basename $GM_INPUT)/.jpg/@2x.jpg}

## Resize for fragment vistas

    gm convert $GM_INPUT -resize 1024x -quality 85 $GM_OUTPUT/vista.jpg
    gm convert $GM_INPUT -resize 2048x -quality 85 $GM_OUTPUT/vista@2x.jpg

## Resize for Passages images

    gm convert $GM_INPUT -resize 1100x -quality 85 $GM_OUTPUT/${$(basename $GM_INPUT)/.jpg/@2x.jpg}

Note we don't bother with a non-retina version because we
can't run Retina.JS.

Note that some systems like Mac OS actually understanding
the `@2x` suffix and will treat the image correctly.
