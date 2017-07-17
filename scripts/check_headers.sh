#!/bin/bash

#
# check_headers.sh
#
# This project uses a custom Markdown addition that allows headers to be
# annotated with an ID so that we can name their permalink. This looks
# something like this:
#
#     ## The Best Section (#the-best-section)
#
# Unfortunately, these are easy to forget, and when I do, the result is a
# generically-named header. This script checks to make sure that all headers
# have an identifier. It runs in CI.

# That first character class should probably be something like [^\(\n], but for
# the life of me I can't get it to not match newlines, which leads to a huge
# multi-line match (even in single line mode). This sucks because I'll probably
# have to add new characters to the class, but it's mostly okay for now.
rx="^###?#?#? [A-Za-z0-9.:'/\-_ ]+(?! \(#.*\))$"

dirs="content/articles/* content/drafts/* content/fragments/* content/fragments-drafts/* content/passages/* content/passages-drafts/*"

# Use ag for Macs (where grep lacks PCRE) and fall back to gnugrep for Linux
# systems. Macs without ag installed won't be able to run this script
# successfully.
if command -v ag >/dev/null; then
    bad_headers=$(ag -C 0 "$rx" $dirs)
# grep will exit with 1 for no match and 2 if we passed a bad option
elif command -v grep > /dev/null && (echo 'abc' | grep --perl-regexp abc >/dev/null 2>&1); then
    bad_headers=$(grep --perl-regexp "$rx" $dirs)
else
    echo "no suitable matchers"
    exit 1
fi

if [[ -n "${bad_headers}" ]]; then
    echo "!!! the following headers do not have IDs: "
    echo "${bad_headers}"
    exit 1
fi
