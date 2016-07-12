---
title: json_ schema
published_at: 2014-05-24T16:17:45Z
---

After a few primitive attempts at developing some basic [JSON
Schema][json-schema] validation systems for [Committee][committee-github], I
finally mustered a bit of initiative and developed a fully spec-compliant
schema and hyper-schema validator which I've officially released as
[`json_schema`][json-schema-github]).

During a validation, `json_schema` will build JSON Pointers to any data that
failed, and to the exact schema rule that failed it. This is designed to help
with the validation debugging process:

```
schema.json#/properties: failed schema #/properties/properties: Expected data to be of type "object"; value was: [].

schema.json#: failed schema #: Data did not match all subschemas of "allOf" condition.
```

I tried to pay special attention to the more difficult features that weren't
well covered by my previous implementations like self-referencing schemas,
tuple validation in `items`, schema `dependencies`, along with a multitude of
other things.

Besides validation, `json_schema` should also provide a nice way to introspect
and traverse schemas programmatically. Committee's already been reworked to use
it, and I hope to soon start moving over other [Interagent][interagent]
projects.

A final thought: there is no better way to understand a spec than to write an
implementation for it. I thought that I had an okay understanding the V4 JSON
Schema spec before starting the project, but then as you might expect, I ran
into every mental grey area, gap in understanding, obscure feature, and special
case. I now realize just how tenuous my understanding actually was.

Check it out on [GitHub][json-schema-github].

[committee-github]: https://github.com/interagent/committee
[interagent]: https://github.com/interagent
[json-schema]: http://json-schema.org/
[json-schema-github]: https://github.com/brandur/json_schema
