+++
hook = "A little background on cookies and local storage, where they overlap, and where they don't."
published_at = 2020-12-05T20:57:58Z
title = "Cookies vs. local storage"
+++

The phrase "cookies vs. local storage" was something I found myself googling the other day. If I'd thought about it longer and harder, the not-so-unsubtle difference would have come to me, but it was a sloppy google-so-I-don't-have-to-think moment.

It did end up putting me down a bit of a rabbit hole though. Cookies are a deeply entrenched internet keystone, but tend to be so well-abstracted that even web developers don't think about what's happening very often. Local storage is a newer idea that does a different job, but with some overlap.

## Cookies (#cookies)

The basic mechanism of a cookie is very simple. A server sends back a header telling a client that it should now set a cookie:

```
Set-Cookie: id=123
```

The client remembers that value, and when making subsequent requests to the same server, sends it through:

```
Cookie: id=123
```

A lot of us are used to thinking about cookies as tarnished technology these days for spying and selling ads, but we should remember that the fundamental reason they exist is to facilitate complex interactions on the web. Unlike with a desktop application which stays open and keeps all its state around moment to moment, the web is fundamentally _stateless_. From a server's perspective, there's no intrinsic way in which two HTTP requests from the same user are related.

Absent cookies, clients would have to send something like a session ID in the payload of every request they made to remind servers of who they are so they could, for example, stay logged in. That's what a cookie does as well, but with a standardized scheme that does the work automatically.

Cookies may also be available on the client side via JavaScript's [`document.cookie`](https://developer.mozilla.org/en-US/docs/Web/API/Document/cookie) API.

Over the years, cookies have picked up new bells and whistles, largely for reasons of security and privacy. `Secure` ensures cookies are sent over encrypted HTTPS only. `HttpOnly` disables client-side JavaScript access. `Path` can scope cookies down to specific HTTP paths. `SameSite` defines behavior around cross-origin requests to mitigate CSRF attacks. Most apps will want to use most of these to keep things as tight as possible. Adtech will intentionally relax use them to improve the chances of users checking in.

Beyond security, the blunt instrument of a cookie can have other downsides as well. Back when I worked at iStock, we were always careful to serve static assets from a separate domain as a resource optimization measure to prevent clients from sending sizable `Cookie` payloads for thousands of image requests where no cookie was needed. This is often less of a problem nowadays by virtue of that the fact that a lot of big apps will be using a separate domain for static assets for CDN purposes anyway, but still worth thinking about.

Lastly, cookies have a size limit of roughly 4096 bytes. This is a hallmark of the fact that they're intended as a vehicle for an identifiers of server-side state, and not intended for use transmission anything heavy.

## Local storage (#local-storage)

Like cookies, local storage also saves state, and can do so across multiple requests. Unlike cookies, it can't be used to transmit anything to servers -- it's a mechanism that lives purely on the client side.

It has a number of advantages. Most obviously, a [much improved API](https://developer.mozilla.org/en-US/docs/Web/API/Window/localStorage) (no need to do your own encoding like with a cookie):

``` js
localStorage.setItem('myCat', 'Tom');
const cat = localStorage.getItem('myCat');
localStorage.removeItem('myCat');
localStorage.clear();
```

Cookies require a dynamic server process to receive a cookie, interpret the results, and do some persistence. Since local storage lives purely in the browser, none of that is required, making it perfect for applications that don't have or want a dynamic component. Static sites for example.

It can also store _more_, up to 5 MB, although it's not recommended to do so because the API for setting and getting data is synchronous and will block the main thread.

Lastly, there's an argument to be made in favor of local storage for user privacy. Why send something back to the cloud when you don't have to? With local storage, users are in possession of their data and can purge it at will.

## What to use (#what-to-use)

There's a reason that cookies exist, and a lot of what they do can't be replicated by local storage, but there's certainly a local storage sweet spot for applications that want to store some minimal configuration needed on the client side, but don't otherwise need cookies for anything.

Personally, I've found it perfect for things like remembering a user's preference between light mode and dark mode -- no server-side persistence required, and local storage is long-lived enough that it still feels changing the setting is permanent, lasting until the user explicitly clears their caches.
