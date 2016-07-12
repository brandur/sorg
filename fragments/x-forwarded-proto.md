---
title: X-Forwarded-Proto
published_at: 2014-05-15T09:37:33Z
---

We use [rack-ssl](https://github.com/josh/rack-ssl) to redirect users making
requests over raw HTTP and force users onto a secure connection. We also
terminate TLS at a reverse proxy, and have it make privileged internal requests
over plain HTTP.

This architecture tends to work very well, but can make issuing internal
requests with a tool like Curl somewhat difficult because individual servers
will deny HTTP connections in favor of HTTPS, but those same servers don't know
how to serve HTTPS!

This is where `X-Forwarded-Proto` comes in. This unstandardized but de facto
standard allows a proxy to hint to backends that a particular protocol was used
to serve the request, avoiding a spurious redirect from something like
rack-ssl. The same technique can be used with Curl to "fake" a secure request
internally:

```
curl -H "X-Forwarded-Proto: https" ...
```
