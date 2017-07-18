---
title: Sending An Email Newsletter
published_at: 2017-07-18T02:49:14Z
hook: TODO
---

## Email CSS & HTML (#css)

I thought our HTML and CSS standardization initiatives had
succeeded. That might be true, but not in the crazy world
of email HTML.

Things that don't work:

* `<style>` tags, so all CSS needs to be inlined (I use a
  CSS inlining library for Go). This limitation seems to be
  largely Google Mail specific these days, but it's so big
  that it creates a de facto standard of its own [1].
* Negative margins.
* Descendant selectors.

Rendering wildly different from browser defaults.

[1] Somewhat ironically, it might even be fair to call
    Google Mail the new Internet Explorer of email
    rendering.
