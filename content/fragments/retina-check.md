---
title: Retina check
published_at: 2016-08-22T06:02:48Z
hook: A CI trick for some accountability around retina assets. Set exclusion in
  Bash.
---

I try to put in retina versions of any JPGs or PNGs that I upload to this site
for higher definition screens. These high resolution assets are given a
filename ending with `@2x`, and [Retina.js][retina-js] swaps them in for the
standard image assets as a page is loading. Occasionally though, I forget to
upload those retina assets, or just get lazy.

I put a fun little script in CI last week [to check that all scalar images have
retina compatible versions][check-retina]. If an image comes in on a pull
request to publish a new article, the build fails and I get a chance to go fix
it.

One quirk that I had to build in was a whitelisting system for the existing set
of images that I have which don't have retina companions. I wanted to take my
set of "bad" images (`A`), exclude any found in the whitelist (`B`), and print
the remainder as those that need fixing. It turns out that in Bash array
operations are a hugely non-trivial operation, and I ended up with quite an
epic hack to work around the limitation:

``` sh
bad_images=$(echo "${allowed_exceptions[@]} ${bad_images[@]}" | tr ' ' '\n' | sort | uniq -u)
```

Which produces the `A - B` set operation that I wanted [1]. It works by
printing both sets, changing the item delimiter to a new line, sorting the
result, and using `uniq` with the unusual `-u` operation to produce the result.
`-u` is the secret sauce: it changes `uniq`'s standard behavior to only
printing lines that are _not_ repeated in the input; in this case uncovering
any filenames that weren't in both sets `A` and `B`.

The next time someone tells you that Bash represents the culmination of the
elegance that is the Unix philosphy, hit them with this set exclusion problem.
If they can write a script without Google's help that doesn't fall back on a
`O(n^2)` operation involving nested loops, then you're dealing with a truly
fearful programmer indeed.

[1] Actually, `A ⊕ B`, but that still works out nicely here.

[check-retina]: https://github.com/brandur/sorg/blob/master/scripts/check_retina.sh
[retina-js]: https://imulus.github.io/retinajs/
