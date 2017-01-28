#!/bin/bash

#
# check_retina.sh
#
# This script tries to ensure that all scalar images (i.e. a ".jpg" or a
# ".png") have a companion retina file, that is one that has a "@2x" before the
# extension and is 4x the resolution of its original. The script runs in CI to
# try and keep me from getting lazy and not finding retina assets.
#
# This is so that images look better on retina displays like newer Mac laptops,
# iPhones, and iPads. While the sources of images still point to the non-retina
# version, a JS script that's loaded (retina.js) replaces them with retina
# content in the DOM if appropriate for a user's device.
#

allowed_exceptions=(
    # These are images that we can safely say won't ever have a good retina
    # equivalent, so they're allowed to fail.
    "./content/images/interfaces/yahoo-1995.jpg"

    # These are images that came in before I started doing the whole retina
    # thing and have been allowed as exceptions. It's not really a big deal,
    # but try to clean it up over time.
    "./content/images/breaktime/commit.png"
    "./content/images/fragments/hm-sportswear/hm.jpg"
    "./content/images/fragments/hm-sportswear/nike.jpg"
    "./content/images/fragments/ipad-mini/vista.jpg"
    "./content/images/fragments/monkeybrains/vista.jpg"
    "./content/images/fragments/new-york/apple-store.jpg"
    "./content/images/fragments/new-york/dont-block-the-box.jpg"
    "./content/images/fragments/new-york/high-line.jpg"
    "./content/images/fragments/new-york/met.jpg"
    "./content/images/fragments/new-york/rockefeller.jpg"
    "./content/images/fragments/new-york/washington-square-park.jpg"
    "./content/images/fragments/paper-books/borderlands-swartz-01.jpg"
    "./content/images/fragments/paper-books/borderlands-swartz-02.jpg"
    "./content/images/fragments/safety-razors/razor.jpg"
    "./content/images/fragments/sprawl-blues/vista.jpg"
    "./content/images/fragments/wgt-2015/vista.jpg"
    "./content/images/fragments/wgt-2015-brain-dump/jo-quail.jpg"
    "./content/images/fragments/wgt-2015-brain-dump/leaf.jpg"
    "./content/images/page/economist.jpg"
    "./content/images/page/fp.png"
    "./content/images/page/rane.jpg"
    "./content/images/page/salon.png"
    "./content/images/page/transworld-surf.jpg"
    "./content/images/page/weekend.png"
    "./content/images/request-ids/splunk-search.png"
)

find_images() {
    find . -type f \( \
        \( \
            -not \( \
                -iname "*@2x.*" \
            \) \
        \) -and \( \
            -iname "*.jpg" \
            -o \
            -iname "*.png" \
        \) \
    \)
}

bad_images=()
for image in $(find_images); do
    extension="${image##*.}"
    base=$(basename $image .$extension)
    retina=${image/"$base.$extension"/"$base@2x.$extension"}

    if [[ ! -f "$retina" ]]; then
        bad_images+=("$image")
    fi
done

# This insane bash-fu takes the difference between two arrays. If either we
# have an exception that no longer exists or have an image without retina
# that's not in the exceptions list, it will be returned.
#
# Note that bad_images changes from an array to a standard string.
bad_images=$(echo "${allowed_exceptions[@]} ${bad_images[@]}" | tr ' ' '\n' | sort | uniq -u)

if [[ -n "${bad_images}" ]]; then
    echo "!!! no retina assets for the following images: "
    echo "${bad_images}"
    exit 1
fi
