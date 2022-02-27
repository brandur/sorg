+++
image_alt = "Pinnacles National Park"
image_url = "/photographs/nanoglyphs/032-hook-toil/pinnacles-park-1@2x.jpg"
published_at = 2022-02-26T19:57:06Z
title = "Hook Toil"
+++

A brief reminiscence on webhooks.

An old idea from Heroku: do everything over HTTP -- everything in the modern age (programming languages, packages, tools) speaks a little HTTP, and it's a powerful platform to build off of -- cut with the grain instead of trying to build on SOAP, RPC, or your own protocol. This is an idea that was more contentious back in 2011 when it was devised -- these days, HTTP has "won" and the majority of what people are doing online happens over it.

Webhooks are the a simple idea that dovetails nicely with that. Rest APIs have mostly taken over the web where it comes to inter-service interoperability, but they don't offer any inherent capability for a "push" mechanism where the server sends something back to the user instead of vice versa. As [originally described all the way back in 2007](https://progrium.github.io/blog/2007/05/03/web-hooks-to-revolutionize-the-web/), webhooks fill this gap -- a user defines an endpoint, and an API provider makes HTTP requests to it, pushing any information it'd like.

Webhooks are a pretty good system, but they're one that seems simpler than it actually is, and most developers implement them too quickly and too enthusiastically. In a perfect world, webhooks work perfectly, but in an imperfect one, there are many degenerate conditions:

* **Non-delivery:** The likelihood is that 99% of the time, you send a webhook, an end user receives it, and the transaction is over. But the other 1% that aren't so smooth are still a lot of sends, and those delivery failures need to be handled.

    Any webhook delivery system will need to be able to retry failures, but even more fundamentally, it'll have to be _stateful_. Putting things into an in-memory queue and hoping for the best isn't good enough if you want to guarantee delivery -- you'll need a durable database.

* **Slow endpoints:** A webhook sender needs to wait for a successful 200 from a receiver to confirm receipt. An ideal webhook receiver gets a message, puts the message in an asynchronous queue on its end, and send back that 200 as quickly as possible, but most webhook receivers are not ideal.

    It's common for receivers to do heavy-lifting in-band with receipt and that's often slow -- taking multiple seconds to respond. Sometimes it's _very_ slow -- their database isn't responsive or something of the like, causing them to take 30+ seconds to respond (and with an error when they finally do). You'll need to put a timeout system into the sender, and ideally make sure it's written to be as parallel as possible.
    
    More broadly, receivers will tend to push all their problems back to the sender. If they have a down database, they'll time out. If they have a data model conflict, they'll 400 / 422 with an error. The sender's taken on trying to guarantee delivery so it'll try its best, but it's not an optimal situation because operational responsibility is farmed out to another party.

* **Down endpoints:** Production problems on the receiver can cause endpoints to go hard down, failing all sends thereafter. This can be a bigger problem than you'd think because a down endpoint often manifests as many time outs in the sender, which can tie up sending capacity if many webhooks are being dispatched to the down endpoint. Once again, it's ideal if the sender is as parallel so that even if a large endpoint becomes unresponsive, the rest of the system stays healthy.

* **Permanently gone endpoints:** Some endpoints will disappear permanently, most often [Ngrok](https://ngrok.com/) listeners used for development, and it's rare for users to bother removing them. You'll need a way of disabling dead endpoints after a certain number of failures to reclaim capacity that would otherwise be permanently wasted.

* **Zombie endpoints:** Endpoints that vacillate between alive and dead. Dead enough to cause many failed and slow requests, but never quite dead enough to be permanently disabled.

* **Lack of economy:** Webhooks are fundamentally inefficient in that they're generally one message per endpoint per send attempt. It may be possible to share connections if you're sending infrastructure is smart enough (i.e. new webhooks have an affinity for boxes from which previous webhooks for that endpoint have been recently sent), but webhooks are hugely wasteful compared to sending many messages over a single stream.

* **Security:** Because webhooks are pushed from provider to user instead of vice versa like in a normal REST API, they need a separate authentication technique. This is generally done these days using webhook signatures, and although there's nothing inherently insecure with signatures, it's difficult to verify that your users are actually verifying them. If they're not, they may be leaving a gaping security hole open that allows a malicious party to push forgeries to your users.

---

## Webhooks at Stripe (#stripe)

Webhooks at Stripe were easily one of the most operationally fraught systems in the entire company.

This is at least partly a historical anomaly. The webhook senders were written in Ruby and only single threaded. That combined with the fact that each Ruby process was around 1 GB of memory (see [Stripe-flavored Ruby in 027](/nanoglyphs/027-15-minutes#stripe-ruby) for more info) meant that the webhook sending fleet had a very fixed sending capacity, and was vastly overprovisioned at great expense so that it had the capacity to deal with spikes in demand. Even Ruby's pseudo-threads would've helped the situation because the senders were spending almost all their time I/O-bound, but there was always enough fear in the air around the safety of using threads in a mature Ruby stack that had spent so many years without them, that we never turned them on. It was a product of its time, and if it'd been written even five years later, it's reasonably likely that a lot of these mistakes would've been avoided -- less memory overhead, and with a better parallel story.

More than a few times, a single large customer would have an endpoint go down, having the effect of a huge number of messages getting queued for retry, and with each attempt taking 30 seconds before finally timing out. The end effect is that not only were messages to that single user disrupted, but to all users as the shared webhook infrastructure suddenly had much more work to do.

Over the years this was shored up in various ways -- the error-prone home-grown queueing infrastructure was replaced with Kafka, a sharding system was introduced so that not everyone was affected by a single shard in bad shape, the retry schedule was revamped for better economy, and dead endpoints were identified more quickly. But over the intervening years, a lot of blood and tears was lost to the name of webhooks. Changing anything in the original system was harder than it sounds -- Stripe is so serious about backwards compatibility that even reducing the timeout from 30 seconds to something more reasonable was consider a breaking change and had to be eased out.

---

## Boxes to tick (#boxes)

So say you are going to build a webhooks system. Here's some things you'll eventually want:

* **Retry schedule:** In case of failure, a retry schedule with exponential backoff and random jitter. The first retry should be seconds after the original failure. The last retry, days.

* **Disable schedule:** Down / gone / zombie endpoints should be given a chance to be healthy again, but eventually be disabled. Keep state to track the overall performance of each endpoint and turn them off when it's too low. Send the owner notifications to tell them what's happened.

* **Control rods:** Operational knobs that can be used in case of emergency. For example, being able to disable webhooks for a single misbehaving user who's bleeding into the rest of the system.

* **Resend system:** For extra points, give users a way to resend webhooks that failed all their retries and were to set as permanently failed. This gives them an easy way to remediate after they might've had a bad production bug.

* **Development system:** Generally speaking, Ngrok is the gold standard developing an endpoint receiver, which isn't a great situation (think about how much more difficult this is compared to testing against the normal API). It's worth considering providing additional tools to make this easy as possible.

---

## Join the resistance (#resistance)

All that said, I'm going to do my best to never develop another webhooks system. All in all, they're hugely wasteful, their DX isn't good, and I've seen them cause way too many problems. I'm not sure what we'll do instead, but I think it'll look more like SSE or WebSockets that stream events back to an active listener. A receiver still has a constant connection open, but only one of them over which thousands of events can stream. More on this as we go.

---

<img src="/photographs/nanoglyphs/032-hook-toil/pinnacles-park-2@2x.jpg" alt="Pinnacles National Park" class="wide" loading="lazy">

## Condors and peaks (#condors)

Yesterday, I had the chance to visit [Pinnacles National Park](https://en.wikipedia.org/wiki/Pinnacles_National_Park), quite notable for being one of the few places where it's possible to see [California condors](https://en.wikipedia.org/wiki/California_condor), one of the rarest birds in the world.

The story of this species is amazing, tragic, and more recently, worthy of some optimism. Very long-lived (60 years), with a late age of sexual maturity (6), laying very few young (one egg every other year), and vulnerable to a host of human-created threats, condors were particularly susceptible to extinction in the modern age. At one point they were down to only 27 birds left on the planet. Every one of them was captured, and a comprehensive breeding program was started with the hopes of repopulating the species. After many years of work it had some success, and condors were reintroduced in California, Arizona, and Mexico. Their number's now up to 504, every one of them carefully tagged and tracked.

Unlike a lot of critically endangered species, it's not that hard to see them -- head down to Pinnacles, up through the High Peaks trail, and there's a reasonable chance of seeing a number of them flying overhead. And it's an impressive sight -- their full-grown wingspan of nine and a half feet is wider than any other North American bird. Just don't expect too many Instagram opportunities -- the photo below is shot with a 500 mm zoom lens and there's still not much to see.

Until next week.

<img src="/photographs/nanoglyphs/032-hook-toil/california-condor@2x.jpg" alt="California condor in flight" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/032-hook-toil/pinnacles-park-3@2x.jpg" alt="Pinnacles National Park" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/032-hook-toil/pinnacles-park-4@2x.jpg" alt="Pinnacles National Park" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/032-hook-toil/balconies-cave@2x.jpg" alt="Balconies Cave at Pinnacles National Park" class="wide" loading="lazy">

