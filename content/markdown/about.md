{{HTMLRender (ImgSrcAndAltAndClass "/assets/images/me.jpg" "Photo of Brandur" "overflowing")}}

I'm an engineer at [Crunchy Data](https://www.crunchydata.com/), where I work on the company's platform API, and spend a lot of time talking about and working with my favorite database, Postgres.

I recently spent five and a half years at Stripe, where I helped design, build, and run our API, and maintain our public-facing developer tooling. Stripe's API design ethos is notable for aiming to make complex flows _as easy as possible_, while still providing enough flexibility to facilitate even the most complex flows.

Before that, I spent four years at Heroku, where I helped create our V3 API and refine and operate its central supporting services. I'm still a strong believer that developers at every product company shouldn't be spending too much time thinking about infrastructure, and have access to a deployment mechanism as easy as `git push heroku master`.

Having written software professionally for many years now, I'm convinced that the default result given our modern processes and tools are products with undesirable levels of fragility. These days I'm especially interested in ways to improve the robustness and longevity of software, and reduce toil in operating it. I have little doubt that powerfully type-safe languages which expose more problems at compile time are the future.

My favorite movie is Sofia Coppola's _Lost in Translation_. My favorite book is Michael Crichton's _Rising Sun_ (although I like [a lot of others](/reading)). I also like running, photography, history, meditation, urban design, and metal.

I publish to this blog with some frequency, and most often to my weekly-ish newsletter [Nanoglyph](/newsletter#nanoglyph). If you like the content here, you should [check out a sample edition](/nanoglyphs/018-ractors) and consider subscribing.

## Technology (#technology)

This site is a static set of HTML, JS, CSS, and image files built using a [custom Go executable](https://github.com/brandur/sorg), stored on S3, and served by a number of worldwide edge locations by CloudFront to help ensure great performance around the globe. It's deployed automatically by CI as code lands in its master branch on GitHub. The architecture is based on the idea of [the Intrinsic Static Site](/aws-intrinsic-static).

| It was previously running [Ruby/Sinatra stack](https://github.com/brandur/org), hosted on Heroku, and using CloudFlare as a CDN.

## Design (#design)

The responsive design aims to improve readability and emphasize content through typography, whitespace, and clean lines compared [to earlier incarnations of my work](https://mutelight.org). It wouldn't have been possible without the timeless beauty of [Helvetica](http://en.wikipedia.org/wiki/Helvetica_(film\)).
