+++
hook = "Prefer an HTTP 200 OK with an empty body over a 204."
published_at = 2021-09-07T13:44:39Z
title = "HTTP API design: 204 No content is mildly harmful"
+++

An HTTP response for 204 No Content indicates success, but that nothing will be sent back in body. I was interested to find in my latest job that we were using 204 on `DELETE` endpoints and on update a few places, which seemed like a novel idea.

However, after thinking about it longer, I wouldn't endorse using 204s for two reasons:

* 204s have no advantage over 200s.
* 204s reduce future change flexibility to zero.

Say you have an endpoint that doesn't need to return anything for now, but then later on your realize that it'd be useful if it did. A 204 has boxed you in: it's against the spec to return a body with one, and languages like Go will error if you try. You could change the status code, but that should be considered to be a minor breaking change because so many integrations are written like `if resp.StatusCode != http.StatusNoContent { <error> }`.

Instead, just return a 200 with an empty JSON object for endpoints that don't need a response. The difference is immaterial to clients, and adding new data later is fully backwards compatible.
