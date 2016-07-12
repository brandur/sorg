---
hook: Implementing rate limiting using Generic Cell Rate Algorithm (GCRA), a sliding
  window algorithm without a drip process.
location: San Francisco
published_at: 2015-09-18T16:42:18Z
title: Rate Limiting, Cells, and GCRA
---

Rate limiting is a mechanism that many developers (especially those running
services) may find themselves looking into at some point in their careers. It's
useful for a variety of purposes, some of which might be:

1. **Sharing access to limited resources.** e.g. Requests made to an API where
   the limited resources are your server capacity, database load, etc.
2. **Security.** e.g. Limiting the number of second factor attempts that a user
   is allowed to perform, or the number of times they're allowed to get their
   password wrong.
3. **Revenue.** Certain services might want to limit actions based on the tier of
   their customer's service, and thus create a revenue model based on rate
   limiting.

Here we'll explore a few different rate limiting algorithms and culminate at
a sophisticated one called GCRA.

## Time Bucketed (#time-bucketed)

A very simple rate limiting implementation is to store a remaining limit in a
bucket that will expire after a certain amount of time. We can do this by
starting a bucket when the first action comes in, decrementing its value as
more actions appear, and make sure it expires after the configured rate
limiting period. The pseudo-code looks like:

``` ruby
# 5000 allowed actions per hour
RATE_BURST  = 5000
RATE_PERIOD = 1.hour

def rate_limit?(bucket)
  if !bucket.exists?
    bucket.set_value(RATE_BURST)
    bucket.set_ttl(RATE_PERIOD)
  end

  if bucket.value > 0
    bucket.decrement
    true
  else
    false
  end
end
```

The Redis `SETEX` command makes this trivial to implement; just set a key
containing the remaining limit with the appropriate expiry and let Redis take
care of clean-up.

### Downsides (#time-bucketed-downsides)

This method can be somewhat unforgiving for users because it allows a buggy or
rogue script to burn an account's entire rate limit immediately, and force them
to wait for the bucket's expiry to get access back.

By the same principle, the algorithm can be dangerous to the server as well.
Consider an antisocial script that can make enough concurrent requests that it
can exhaust its rate limit in short order and which is regularly overlimit.
Once an hour as the limit resets, the script bombards the server with a new
series of requests until its rate is exhausted once again. In this scenario the
server always needs enough extra capacity to handle these short intense bursts
and which will likely go to waste during the rest of the hour. This wouldn't be
the case if we could find an algorithm that would force these requests to be
more evenly spaced out.

GitHub's API is one such service that implements this naive algorithm (I will
randomly pick on them, but many others do this as well), and I can use a
benchmarking tool to help demonstrate the problem [1]. After my 5000 requests
are through I'll be locked out for the next hour:

``` sh
$ curl --silent --head -i -H "Authorization: token $GITHUB_TOKEN" \
    https://api.github.com/users/brandur | grep RateLimit-Reset
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1442423816

$ RESET=1442423816 ruby -e 'puts "%.0f minute(s) before reset" % \
    ((Time.at(ENV["RESET"].to_i) - Time.now) / 60)'
51 minute(s) before reset
```

## Leaky Bucket (#leaky-bucket)

Luckily, an algorithm exists that can take care of the problem of this sort of
jagged rate limiting called the [leaky bucket][leaky-bucket]. It's intuitive to
understand by comparing it to its real-world namesake: imagine a bucket
partially filled with water and which has some fixed capacity (τ). The bucket
has a leak so that some amount of water is escaping at a constant rate (T).
Whenever an action that should be rate limited occurs, some amount of water
flows into the bucket, with the amount being proportional to its relative
costliness. If the amount of water entering the bucket is greater than the
amount leaving through the leak, the bucket starts to fill. Actions are
disallowed if the bucket is full.

``` monodraw
┌──────────────────┐               *                      
│                  │░             ***      User           
│   LEAKY BUCKET   │░         │   ***    actions          
│                  │░         │            add            
└──────────────────┘░         │          "water"          
 ░░░░░░░░░░░░░░░░░░░░         │                           
                              │    *                      
                        ═╗    ▼   ***        ╔═      ▲    
                         ║        ***        ║       │    
                         ║                   ║       │    
                         ║                   ║  τ = Bucket
                         ║                   ║   capacity 
                         ║*******************║       │    
                         ║*******************║       │    
                         ║*******************║       │    
                         ╚════════╗*╔════════╝       ▼    
                                  ║*║                     
                                  ║*║                     
                                 ═╝*╚═                    
                                   *                      
                              │    *                      
                              │    *    Constant          
                              │    *    drip out          
                              ▼    *                      
                                   *                      
```

The leaky bucket produces a very smooth rate limiting effect. A user can still
exhaust their entire quota by filling their entire bucket nearly
instantaneously, but after realizing the error, they should still have access
to more quota quickly as the leak starts to drain the bucket.

The leaky bucket is normally implemented using a background process that
simulates a leak. It looks for any active buckets that need to be drained, and
drains each one in turn. In Redis, this might look like a hash that groups all
buckets under a type of rate limit and which is dripped by iterating each key
and decrementing it.

### Downsides (#leaky-bucket-downsides)

The naive leaky bucket's greatness weakness is its "drip" process. If it goes
offline or gets to a capacity limit where it can't drip all the buckets that
need to be dripped, then new incoming requests might be limited incorrectly.
There are a number of strategies to help avoid this danger, but if we could
build an algorithm without a drip, it would be fundamentally more stable.

## GCRA (#gcra)

This leads us to the leaky bucket variant called ["Generic Cell Rate
Algorithm"][gcra] (GCRA). The name "cell" comes from a communications
technology called [Asynchronous Transfer Mode][atm] (ATM) which encoded data
into small packets of fixed size called "cells" (as opposed to the variable
size frames of IP). GCRA was the algorithm recommended by the [ATM
Forum][atm-forum] for use in an ATM network's scheduler so that it could either
delay or drop cells that came in over their rate limit. Although today [ATM is
dead][atm-dead], we still catch occasional glimpses of its past innovation with
examples like GCRA.

GCRA works by tracking remaining limit through a time called the "theoretical
arrival time" (TAT), which is seeded on the first request by adding a duration
representing its cost to the current time. The cost is calculated as a
multiplier of our "emission interval" (T), which is dervied from the rate at
which we want the bucket to refill. When any subsequent request comes in, we
take the existing TAT, subtract a fixed buffer representing the limit's total
burst capacity from it (τ + T), and compare the result to the current time.
This result represents the next time to allow a request. If it's in the past,
we allow the incoming request, and if it's in the future, we don't. After a
successful request, a new TAT is calculated by adding T.

The pseudo-code for the algorithm can look a little daunting, so instead of
showing it, I'll recommend taking a look at our reference implementation
[Throttled](#throttled) (more on this below) which is free of heavy
abstractions and straightforward to read. Instead, let's take a look at visual
representation of the timeline and various variables for successful request.
Here we see an allowed request where t<sub>0</sub> is within the bounds of TAT
- (τ + T) (i.e. the time of the next allowed request):

``` monodraw
┌───────────────────┐                                                    
│                   │░                                                   
│  ALLOWED REQUEST  │░                                                   
│                   │░                                                   
└───────────────────┘░                                                   
 ░░░░░░░░░░░░░░░░░░░░░                                                   
                                                                         
                ┌────────┐                                               
                │allow at│               ┌───────┐     ┌───────┐         
                │ (past) │               │  t0   │     │  tat  │         
                └───┬────┘               └───┬───┘     └───┬───┘         
                    │                        │             │             
                    ▼                        ▼             ▼             
────────────────────+──────────────────────────────────────+───────▶     
                    │//////////////////////////////////////│        time 
                    │//////////////////////////////////////│             
                    └──────────────────────────────────────┘             
                     ◀────────────────τ + T───────────────▶              
                                                                         
                                                                         
┌─────────────────────────────────────┐                                  
│T     = Emission interval            │░                                 
│τ     = Capacity of bucket           │░                                 
│T + τ = Delay variation tolerance    │░                                 
│tat   = Theoretical arrival time     │░                                 
│t0    = Actual time of request       │░                                 
└─────────────────────────────────────┘░                                 
 ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                 
```

For a failed request, the time of the next allowed request is in the future,
prompting us to deny the request:

``` monodraw
┌───────────────────┐                                                    
│                   │░                                                   
│  DENIED REQUEST   │░                                                   
│                   │░                                                   
└───────────────────┘░                                                   
 ░░░░░░░░░░░░░░░░░░░░░                                                   
                                                                         
                ┌────────┐                                               
  ┌───────┐     │allow at│                             ┌───────┐         
  │  t0   │     │(future)│                             │  tat  │         
  └───┬───┘     └───┬────┘                             └───┬───┘         
      │             │                                      │             
      ▼             ▼                                      ▼             
────────────────────+──────────────────────────────────────+───────▶     
                    │//////////////////////////////////////│        time 
                    │//////////////////////////////////////│             
                    └──────────────────────────────────────┘             
                     ◀────────────────τ + T───────────────▶              
```

Because GCRA is so dependent on time, it's critical to have a strategy for
making sure that the current time is consistent if rate limits are being
tracked from multiple deployments. Clock drift between machines could throw off
the algorithm and lead to false positives (i.e. users locked out of their
accounts). One easy strategy here is to use the store's time for
synchronization (for example, by accessing the `TIME` command in Redis).

### Throttled (#throttled)

My colleague [Andrew Metcalf][andrew-metcalf] recently upgraded the open-source
Golang library [Throttled][throttled] [2] from a naive rate limiting
implementation to the one using GCRA. The package is well-documented and
well-tested, and should serve as a pretty intuitive reference implementation
for the curious. It's already taking production traffic at Stripe and should be
soon at Heroku as well. In particular, check out [rate.go][throttled-rate]
where the bulk of the implementation of GCRA is located.

And of course, if you need a rate limiting module and have some Go in your
stack, we'd love for you to give Throttled a whirl and let us know how it went.

[andrew-metcalf]: https://github.com/metcalf
[atm]: https://en.wikipedia.org/wiki/Asynchronous_Transfer_Mode
[atm-dead]: http://technologyinside.com/2007/01/31/part-1-the-demise-of-atm…/
[atm-forum]: https://en.wikipedia.org/wiki/ATM_Forum
[boom]: https://github.com/rakyll/boom
[gcra]: https://en.wikipedia.org/wiki/Generic_cell_rate_algorithm
[leaky-bucket]: https://en.wikipedia.org/wiki/Leaky_bucket
[throttled]: https://github.com/throttled/throttled
[throttled-rate]: https://github.com/throttled/throttled/blob/ef1aa857b069ed60f6f859f8b16350e5b7c8ec96/rate.go#L155-L239

[1] Note that it appears that GitHub also has a secondary limiting mechanism
    that protects its API from fast instantaneous bursts. I'm still using them
    as an example because although this does help protect them, it doesn't help
    very much in protecting me from myself (i.e. blowing through my entire
    limit with a buggy script).

[2] I am also a maintainer of Throttled.
