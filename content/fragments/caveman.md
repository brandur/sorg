+++
hook = "In 1980, Michael Crichton characters in _Congo_ spoke like cavemen to save satellite bandwidth. It was absurd. Ridiculous! Forty-five years later, we're doing the same thing with LLMs to save tokens."
# image = ""
published_at = 2026-04-12T11:41:03-05:00
title = "Caveman"
+++

An excerpt from Michael Crichton's [Congo (1980)](https://en.wikipedia.org/wiki/Congo_(novel\)):

> “I don't understand," Elliot said. Ross explained that the "M" meant that there was more message, and he had to press the transmit button again. He pushed the button several times before he got the message, which in its entirety read:
>
> > REVUWD ORGNL TAPE HUSTN NU FINDNG RE AURL SIGNL INFO-COMPUTR ANLYSS COMPLTE THNK ITS LNGWGE.
>
> Elliot found he could read the compressed shortline language by speaking it aloud: "Reviewed original tape Houston, new finding regarding aural signal information, computer analysis complete think it's language." He frowned. "Language?”

Crichton was a gear guy. The story's protagonists took high tech satellite uplinks into the field, allowing transmission back to HQ, but due to the extreme expense of satellite bandwidth, having to read messages in shorthand like, "REVUWD ORGNL TAPE HUSTN NU FINDNG".

I always found it ridiculous. Although these words have had their vowels removed, they're still uniquely intelligible in the English language. It'd be trivial to write a short algorithm that'd use a dictionary to expand the message back to uncompressed English on the receiving end. Or better yet, stop with the vowel thing and use a standard compression algorithm [1]. You'd get better results.

---

Yesterday, I came across [Caveman](https://github.com/JuliusBrussee/caveman). Its job is to save tokens in Claude by having the LLM speak like a caveman, removing filler words and other niceties that make up a more fluently legible human language.

Before:

> "Sure! I'd be happy to help you with that. The issue you're experiencing is most likely caused by your authentication middleware not properly validating the token expiry. Let me take a look and suggest a fix."

After:

> "Bug in auth middleware. Token expiry check use `<` not `<=`. Fix:"

Crichton would've loved it. 45 years later we've come full circle, are back to speaking like cavemen again, and as an at-least-somewhat legitimate technical workaround. I don't know what I thought I knew anymore.

[1] Although Zip and Gzip wouldn't be invented for another decade, earlier precursors [LZ77/LZ78](https://en.wikipedia.org/wiki/LZ77_and_LZ78) were around at the time.