+++
hook = "_Not_ a dark mode tutorial, but a few notes on some specific refinements of a good dark mode implementation like tri-state instead of bi-state toggle, avoiding page flicker, and responding to theme changes from other tabs or the OS."
published_at = 2024-05-23T17:27:09+02:00
title = "Notes on implementing dark mode"
+++

As you can see from the pretty new toggle at the top, I recently added dark mode to this site. I thought this was something that'd never happen because over a decade I'd built up an inescapable legacy of fundamentally unmaintainable CSS, but for over a year I've been slowly making headway refactoring everything to Tailwind, and with that finally done, the possibility of dark mode was unlocked.

The internet is awash with tutorials on how to implement dark mode, so I won't cover the basics, but while those tutorials will get you to a rudimentary dark mode implementation, I found that every one I read lacked the refinements necessary to get to a _great_ dark mode implementation, with many of the fine details easy to get wrong. Here I'll cover the big ones that are commonly missing.

Some select code snippets are included below, but a more complete version should be found in this site's repository. [HTML including Tailwind styling](https://github.com/brandur/sorg/blob/master/views/_dark_mode_toggle.tmpl.html), [JavaScript](https://github.com/brandur/sorg/blob/master/views/_dark_mode_js.tmpl.html), and [a little custom non-Tailwind CSS](https://github.com/brandur/sorg/blob/ed0bc11a82d1159d73e1f0068f8eb37561903f23/content/stylesheets-modular/tailwind_custom.css#L41-L61).

---

## Frontend basics for backend people (#frontend-basics)

First, a couple core concepts that'll be referenced below.

If you're not a frontend programmer, then you should be aware of the existence of the [`prefers-color-scheme` CSS media selector](https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-color-scheme) which lets a web page react to an OS-level light/dark mode setting:

``` css
@media (prefers-color-scheme: dark) {
    // dark mode styling here
}
```

We'll also be making use of [local storage](https://developer.mozilla.org/en-US/docs/Web/API/Web_Storage_API). Although both are superficially key/value stores, local storage differs from cookies in that it's intended for use from a client's browser itself compared to cookies which are server-side constructs.

The only way to implement a _permanently_ lived light/dark mode setting that's persistent forever and across computers is to use a cookie to reference a server-side account where its stored in a database. That's not possible for this site because it doesn't have a server-side implementation or database, but local storage the next best thing. It also has the added benefits of having no default expiration date (so unless a user manually clears data, a light/dark mode setting is sticky for a long time), and sends less personal information to the server, which is broadly a good thing.

---

## Tri-state (#tri-state)

By far the most common mistake in dark mode implementations is to make it a bi-state instead of tri-state setting. At first glance it might seem like the only two relevant states are light or dark, but there are actually three:

* User has explicitly enabled dark mode.
* User has explicitly enabled light mode.
* User has enabled neither dark mode nor light mode. Fall back to the preference expressed by their OS in `prefers-color-scheme`.

This is implemented by way of a three state radio button that's heavily styled to look like the toggle you see above (Tailwind classes have been removed for clarity, but see [the toggle's template](https://github.com/brandur/sorg/blob/master/views/_dark_mode_toggle.tmpl.html) for gritty details):

``` html
<input value="light" name="theme_toggle_state" type="radio" />
<input value="auto" name="theme_toggle_state" type="radio" />
<input value="dark" name="theme_toggle_state" type="radio" />
```

On input change, store the selected value to local storage and add the CSS class `dark` to the page's HTML element so that Tailwind knows to style itself with the appropriate theme:

``` js
// Runs on initial page load. Add change listeners to light/dark
// toggles that set a local storage key and trigger a theme change.
window.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('.theme_toggle input').forEach((toggle) => {
        toggle.addEventListener('change', (e) => {
            if (e.target.checked) {
                if (e.target.value == THEME_VALUE_AUTO) {
                    localStorage.removeItem(LOCAL_STORAGE_KEY_THEME)
                } else {
                    localStorage.setItem(LOCAL_STORAGE_KEY_THEME, e.target.value)
                }
            }

            setThemeFromLocalStorageOrMediaPreference()
        }, false)
    })
})
```

Use of `auto` (no explicit light/dark preference) styles the page based on `prefers-color-scheme`:

``` js
// Sets light or dark mode based on a preference from local
// storage, or if none is set there, sets based on preference
// from the `prefers-color-scheme` CSS media selector.
function setThemeFromLocalStorageOrMediaPreference() {
    const theme = localStorage.getItem(LOCAL_STORAGE_KEY_THEME) || THEME_VALUE_AUTO

    switch (theme) {
        case THEME_VALUE_AUTO:
            if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
                setDocumentClasses(THEME_CLASS_DARK)
            } else if (window.matchMedia('(prefers-color-scheme: light)').matches) {
                setDocumentClasses(THEME_CLASS_LIGHT)
            }
            break

        case THEME_VALUE_DARK:
            setDocumentClasses(THEME_CLASS_DARK, THEME_CLASS_DARK_OVERRIDE)
            break

        case THEME_VALUE_LIGHT:
            setDocumentClasses(THEME_CLASS_LIGHT, THEME_CLASS_LIGHT_OVERRIDE)
            break
    }

    document.querySelectorAll(`.theme_toggle input[value='${theme}']`).forEach(function(toggle) {
        toggle.checked = true;
    })
}
```

## Avoiding theme flicker on load (#flicker)

The next most common mistake is page flicker on load. The flicker is caused by the page initially styling itself with its default theme (usually light mode), then noticing that a different theme should be set and reskinning itself, but not before there's an observable flash.

A lot of sites have so much crap happening when they're loading that a flicker is lost amongst a sea of jarring effects (e.g. UI elements jumping around as sizes are determined suboptimally late), but a tasteful dark mode implementation takes care to avoid it.

The key to doing so is to make sure that the theme is checked initially by JavaScript that's run inline with the page's body. Putting it in an external file `<script src="...">` or in a listener like `DOMContentLoaded` is _too late_, and will cause a flicker.

Common convention is to use a `<script>` tag right before body close:

``` html
<body>
    ...

    <script>
        ...

        // script must run inline with the page being loaded
        setThemeFromLocalStorageOrMediaPreference()
    </script>
</body>
```

## Theme changes in other tabs (#theme-other-tabs)

If a user has multiple tabs open to your site and changes the theme in one of them, it should take effect immediately in all others.

Luckily, our use of local storage makes this trivially easy. JavaScript provides the `storage` event for when a local storage key changes. Hook into that, and this problem is solved with five lines of code:

``` js
// Listen for local storage changes on our theme key. This lets
// one tab to be notified if the theme is changed in another,
// and update itself accordingly.
window.addEventListener("storage", (e) => {
    if (e.key == LOCAL_STORAGE_KEY_THEME) {
        setThemeFromLocalStorageOrMediaPreference()
    }
})
```

A changed theme takes effect instantly. A user clicks back to another tab and the new theme is there. No page reload required.

Take care that along with the page's theme, `setThemeFromLocalStorageOrMediaPreference()` also sets any light/dark toggles to the right place.

## Theme changes from the OS (#theme-os)

The page should respond to OS-level changes in theme, which is easy via the [`matchMedia()` API](https://developer.mozilla.org/en-US/docs/Web/API/Window/matchMedia):

``` js
// Watch for OS-level changes in preference for light/dark mode.
// This will trigger for example if a user explicitly changes
// their OS-level light/dark configuration, or on sunrise/sunset
// if they have it set to automatic.
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', e => {
    setThemeFromLocalStorageOrMediaPreference()
})
```

If you're on macOS and have appearance set to "Auto" for light/dark mode to change at sunrise and sunset, this code makes sure that a site restyles itself automatically at that time. It also responds if a user manually sets their OS-level appearance. This is another small detail that most people won't even notice (manually reloading the page will also set the right theme), but a good one nonetheless, and takes mere seconds to get right.

Side note: As a frequent critic of JavaScript, I have to acknowledge just how good its browser APIs are at helping developers _get things done_. Local storage and media matchers are powerful, complicated features, and yet we can plug into them knowing practically nothing about the elaborate effort that went into their internal implementations, and with only a handful of lines of code. Excellent work.

## Syntax highlighting with Shiki (#shiki)

A site's syntax highlighting for code blocks likely involves elaborate styling, and while some themes might look okay in either light or dark, it's even better if the code theme changes along with the rest of the site.

I'd gotten tired of Prism's various quirks, and recently made the move [over to Shiki](/fragments/shiki). One of its many benefits is easy support for [dual light/dark themes](https://shiki.matsu.io/guide/dual-themes) with minimal configuration:

``` js
// Shiki will add its own `<pre><code>`, so go to parent `<pre>`
// and replace the entirety of its HTML.
codeBlock.parentElement.outerHTML = 
    await codeToHtml(code, {
        lang: language,
        themes: {
            dark: 'nord',
            light: 'rose-pine'
        }
    })
```

``` css
html.dark .shiki,
html.dark .shiki span {
    background-color: var(--shiki-dark-bg) !important;
    color: var(--shiki-dark) !important;
    font-style: var(--shiki-dark-font-style) !important;
    font-weight: var(--shiki-dark-font-weight) !important;
    text-decoration: var(--shiki-dark-text-decoration) !important;
}
```