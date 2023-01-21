+++
hook = "~4 lines of Go to produce succinct, stable, URL-friendly slugs from timestamps."
published_at = 2022-12-29T10:46:06-07:00
title = "Short, friendly base32 slugs from timestamps"
+++

I recently added [Atoms](/atoms) to this site, which are short, tweet-length posts published via Git and [TOML file](https://github.com/brandur/sorg/blob/master/content/atoms/_meta.toml).

One of my goals was that they should be similarly easy to publish as a tweet, which is partly accomplished using a VSCode snippet and shell alias that copies the current timestamp to clipboard.

Beyond that, I wanted each atom to have a short, stable slug that could be included in a permalink for lasting posterity, but didn't want to have to choose these slugs myself because it'd add friction to publishing.

On Bridge we've been making heavy use of public [16-byte identifiers similar to a UUID](https://docs.crunchybridge.com/api-concepts/eid/), but encoded in base32 to be more succinct and more easily copy/pastable. Common base32 encodings also do clever things like remove characters that could be visually confused by a human like "1" and "l". From [RFC 4648](https://www.ietf.org/rfc/rfc4648.txt):

> The characters "0" and "O" are easily confused, as are "1", "l", and "I".  In the base32 alphabet below, where 0 (zero) and 1 (one) are not present, a decoder may interpret 0 as O, and 1 as I or L depending on case. (However, by default it should not; see previous section.)

I took a base32 approach for Atom slugs, converting a timestamp to a unix time integer, then encoding its bytes to base32 by way of `math.Big` [1]:

``` go
func atomSlug(publishedAt time.Time) string {
	i := big.NewInt(publishedAt.Unix())
	return lexicographicBase32Encoding.EncodeToString(i.Bytes())
}
```

The generated slugs are short and URL friendly like `gioee22`, `giofrfk`, or `giooep2`.

## Lexicographic base32 (#lexicographic-base32)

I also added in one more user convenience trick. RFC 4648 dictates that the numbers appear at the end of the encoding character set like `ABCDEFGHIJKLMNOPQRSTUVWXYZ234567`. You're probably used to hex encoding, so think of RFC 4648 base32 like the opposite of that. Hex's character set would look like `0123456789ABCDEF` with `0-9` appearing at the beginning.

Numbers at the end would generally go unnoticed, but the one problem with it is that in standard lexicographic (alphabetic) sorting numbers are sorted before other characters, and therefore sortable values won't necessarily sort the same after being base32 encoded. It's not the end of the world, but if I'm generating a series of atom slugs from timestamps that'll land as S3 objects for my site like `gioee22`, `giofrfk`, `giooep2`, ..., they're not always sorted chronologically when I list them. (They're _mostly_ chronologically sorted, but with some swapped when two share a prefix and a number comes into play.)

The problem is minor, but so is the remedy. I modified the base32 character set so that numbers encoded first like `234567abcdefghijklmnopqrstuvwxyz` [2]:

``` go
// Very similar to RFC 4648 base32 except that numbers come first instead of
// last so that sortable values encoded to base32 will sort in the same
// lexicographic (alphabetical) order as the original values. Also, use lower
// case characters instead of upper.
var lexicographicBase32 = "234567abcdefghijklmnopqrstuvwxyz"

var lexicographicBase32Encoding = base32.NewEncoding(lexicographicBase32).
		WithPadding(base32.NoPadding)
```

So now when I list atoms from my filesystem or in S3, they come back in the same order that I wrote them.


## Derivatives

- This post inspired a [Rust crate](https://crates.io/crates/lexicoid)

[1] `math/big`'s integer `Bytes()` produces bytes in big-endian order (most significant byte first).
[2] Character set is also downcased. Characters are easier to distinguish and it looks better.
