{{HTMLRender (ImgSrcAndAltAndClass "/assets/images/me.jpg" "Photo of Brandur" "overflowing")}}

I'm an engineer at Stripe where I help design, build, and run our API and maintain our public-facing developer tooling.

I recently spent quite a few years at Heroku where I helped create our V3 API and refine and operate its central supporting services. I'm still a strong believer that developers at every product company shouldn't be spending too much time thinking about infrastructure, and have access to a deployment mechanism as easy as `git push heroku master`.

Having written software professionally for many years now, I'm convinced that the default result given our modern processes and tools are products with undesirable levels of fragility. These days I'm especially interested in ways to improve the robustness and longevity of software, and reduce toil in operating it. I have little doubt that powerfully type-safe languages which expose more problems at compile time are the future. Though nascent today, I think that soon the most pragmatic option will be Rust.

My favorite movie is Sofia Coppola's _Lost in Translation_. My favorite book is Michael Crichton's _Rising Sun_ (although I like [a lot of others](/reading)). I also like running, photography, history, meditation, urban design, and metal.

A few times a year I publish a newsletter called _Passages & Glass_. If you liked some of the other content here, you should [consider subscribing](/newsletter).

## Technology (#technology)

This site is a static set of HTML, JS, CSS, and image files built using a [custom Go executable](https://github.com/brandur/sorg), stored on S3, and served by a number of worldwide edge locations by CloudFront to help ensure great performance around the globe. It's deployed automatically by CI as code lands in its master branch on GitHub. The architecture is based on the idea of [the Intrinsic Static Site](/aws-intrinsic-static).

| It was previously running [Ruby/Sinatra stack](https://github.com/brandur/org), hosted on Heroku, and using CloudFlare as a CDN.

## Design (#design)

The responsive design aims to improve readability and emphasize content through typography, whitespace, and clean lines compared [to earlier incarnations of my work](https://mutelight.org). It wouldn't have been possible without the timeless beauty of [Helvetica](http://en.wikipedia.org/wiki/Helvetica_(film\)).
