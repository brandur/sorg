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
    # Images that don't have a retina version by design.
    "./content/images/standin_00.jpg"
    "./content/images/standin_01.jpg"
    "./content/images/standin_02.jpg"
    "./content/images/standin_03.jpg"
    "./content/images/standin_04.jpg"
    "./content/images/standin_portrait_00.jpg"
    "./content/images/talks/standin_00.png"

    # Raws.
    "./content/raws/talks/paradise-lost/mongo-ad-censored.jpg"
    "./content/raws/talks/paradise-lost/mongo-ad.jpg"
)

find_images() {
    find ./content -type f \( \
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
