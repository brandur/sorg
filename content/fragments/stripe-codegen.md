+++
hook = "A rough sketch on how Stripe's API libraries go from DSL to code."
published_at = 2021-04-25T22:32:12Z
title = "The state of Stripe API library codegen"
+++

One thing that Stripe's well known for are its API libraries. Rather than talking HTTP directly to the API, users will usually opt to use a specific library provided for the language of their choice -- Ruby, Go, C#, etc. The libraries adhere to language conventions reasonably well, and are released in lockstep with new features.

One question I've gotten a few times now is how the pipeline of keeping these up-to-date works, and as I've transitioned out of Stripe, I figured I'd put down some broad strokes.

For many years, the process was very manual, and the libraries were kept updated by virtue of one person who watched API changes like a hawk, then opened pull requests on seven languages when new ones appeared. This was never sustainable, but through a combination of superhuman discipline and very long work weeks, it worked for a long time.

Eventually, most of the libraries moved to a code generated model, and soon _all_ of them will be codegen'ed in a way that works reasonably well. Codegen was difficult to bring in because we wanted to maintain compatibility with the existing versions. It would've been easy to try an open source solution and just run with that, but we had such a large base of existing users that it would've meant either a very painful breakage for them to upgrade, or us permanently maintaining two divergent lines at considerable cost.

Here's a rough sketch of how it works now:

* The canonical source of truth for an API's shape (routes, input parameters, output fields, ...) is a DSL built in Ruby code around the API server's implementation.

* A script walks the DSL and dumps its structure to OpenAPI. There are a few variants of it -- one very complete version for internal use, one version for the Dashboard, and a public version. The public version is [openly available on GitHub](https://github.com/stripe/openapi). A check in CI makes sure that OpenAPI specs on mainline are accurate.

* A JavaScript program reads OpenAPI and uses it to codegen API library code in all support languages. Code is defined using [JSX](https://reactjs.org/docs/introducing-jsx.html)-style templates (this is all completely fake code and at least 50% wrong, but should give you the right idea):

    ``` js
    const apiInvocation = (
        <Function name="...">
            obj, err := <InvokeHTTP method="{verb}" path="{path}" />
            return obj, err
        </Function>
    );
    ```

    Generic primitives can then be composed at a higher level:

    ``` js
    const chargeResource = (
        <APIResource name="Charge">
            <APIInvocation name="refund" />
        </APIResource>
    );
    ```

    Templates are written to be as generic as possible -- ideally every API resource follows the same form and we can just use one template to read out of OpenAPI and spit them all out, but enough special-cased API library code accumulated over the years that customized templates to maintain compatibility are common.

* API libraries are written in two layers. The outer layer contains the models and API resources which codegen creates, and the inner is the infrastructure and common utilities that codegen'ed code calls into, and which will continue to be maintained by hand.

The most important takeaway is that the system works, but it's built on a pipeline that's completely custom, and really not reusable anywhere else. The real question is, what should someone do today to accomplish something similar (and hopefully at a lower cost). I don't know that answer to that, but will be exploring that question in coming months.
