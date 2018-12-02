#!/bin/bash

#
# When copying Dropbox links they'll include a `?dl=0` query parameter at the
# end which links you to a preview page instead of an actual download. You can
# change this to `?dl=1` to fix the behavior, but it's really easy to forget to
# do that when copying links around so this check script can be installed as
# part of a pre-commit hook to help prevent errors. (See `scripts/pre-commit`.)
#

PHOTOGRAPHS_DATA="content/photographs.yaml"

if grep "?dl=0" "$PHOTOGRAPHS_DATA"; then
    echo '!!! found `?dl=0` in `$PHOTOGRAPHS_DATA`'
    exit 1
fi
