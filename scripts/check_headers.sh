#!/bin/bash

# That first character class should probably be something like [^\(\n], but for
# the life of me I can't get it to not match newlines, which leads to a huge
# multi-line match (even in single line mode). This sucks because I'll probably
# have to add new characters to the class, but it's mostly okay for now.
bad_headers=$(ag -C 0 '^###?#?#? [A-Za-z0-9'\-]+(?! \(#.*\))$' content/articles/* content/drafts/*)

if [[ -n "${bad_headers}" ]]; then
    echo "!!! the following headers do not have IDs: "
    echo "${bad_headers}"
    exit 1
fi
