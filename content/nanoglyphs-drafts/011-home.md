+++
image_alt = "Wind chimes in my backyard"
image_url = "/assets/images/nanoglyphs/011-home/chimes@2x.jpg"
published_at = 2020-03-25T02:24:28Z
title = "At Home; Asynchronous I/O, in Circles"
+++

Well, the world’s changed since the last time we spoke. How many times in life can you legitimately say that?

This is _Nanoglyph_, an experimental newsletter on software, ... and occasionally back yards and mountains. If you want to get straight to hard technology, scroll down a little ways for some close-to-the-metal content on kernel I/O. If you're pretty sure you never signed up for this wordy thing and want to be shown to the closest escape hatch, unsubscribe in [one painless click](%unsubscribe_url%).

---

The windows are dark in San Francisco’s museums, bars, and bookstores -- everywhere minus a few exceptions like grocery stores and restaurants with a take out business. Paradoxically, despite a pervasive malaise, the city feels nicer than they’ve ever been (we’re still allowed out for exercise) -- traffic is less voluminous and more calm, the streets are quieter, the air is fresher [1].

Before the shelter in place order went out, Stripe’s offices formally closed to all personnel. Before the office closure order went out, we’d been encouraged to work from home. That makes this my third week working out of my apartment, where before present circumstances I’d never spent even a fraction of this much contiguous time.

It’s had its ups and downs. My ergonomic situation is dire -- my working posture isn’t good even in the best of times, and right now I’m changing sitting positions (low desk, cross-legged on floor, couch, bed, cycle back to start) every 30 minutes so my later life doesn’t see me taking up residence in a chiropractor’s office for live-in back treatment. But that little bit of extra pain is balanced by reduced stress -- my commute is probably better than 90% of America’s, but even so, it’s incredible how much time it eats and anxiety it adds to the day.

---

A few years ago I moved to the Twin Peaks neighborhood of San Francisco. Just up the hill from the Castro, and right across from the giant three-pronged TV/radio tower on top of Mount Sutro (Sutro Tower), still the most conspicuous landmark on the city’s skyline, even after the notorious addition of Salesforce Tower.

![Sutro Tower](/assets/images/nanoglyphs/011-home/sutro-tower@2x.jpg)

If you’re going to be spending time at home, Twin Peaks isn’t a bad place to do it. Even with surprisingly high apartment density and everyone holed up in their units, it’s impressively quiet. I’ve been doing daily runs through our local trail system and then up to the top of the peaks themselves. The building’s back yard (pictured at the top) makes a good place to write. The occasional meditative lap around the filled in pool on the upper terrace (pictured below; the compound’s most unusual feature) helps to focus.

![Pool on the roof](/assets/images/nanoglyphs/011-home/pool@2x.jpg)

---

Despite rarely even seeing another human being in the flesh, I manage to have 15 conversations about the virus a day. News is breaking all the time, and yet, changes very slowly. Twitter is at its most unhelpful ever.

A thought that hits me hourly is that despite extraordinary events, this might just be one of the best self-development opportunities of all time. You know all those things that every adult claims to want to do but which is impossible because there’s no time? Learning a language. Nailing down a healthy diet. Picking up an instrument. Writing a book. Learning to draw. This is it -- an extended snow day for most of humanity and the perfect time and excuse to stay home and do something constructive.

In theory anyway. So far I’ve just been playing video games. I’m working on that.

---

## I/O classic (#io-classic)

## Generation one (#generation-one)

## Great circles (#great-circles)

### Progressive enhancement (#progressive-enhancement)

``` c
struct io_uring ring;
io_uring_queue_init(ENTRIES, &ring, 0);

struct io_uring_sqe sqe;    
struct io_uring_cqe cqe;    

/* get an sqe and fill in a READV operation */
sqe = io_uring_get_sqe(&ring);
io_uring_prep_readv(sqe, fd, &iovec, 1, offset);    

/* tell the kernel we have an sqe ready for consumption */
io_uring_submit(&ring);    

/* wait for the sqe to complete */ io_uring_wait_cqe(&ring, &cqe);   

/* read and process cqe event */
app_handle_cqe(cqe); io_uring_cqe_seen(&ring, cqe);
```

---

![Tea set -- made in Calgary and Japan](/assets/images/nanoglyphs/011-home/tea@2x.jpg)

I’m overhauling my day-to-day by starting small. _Really_ small.

* **Sitting cross-legged:** I was taught to do this from the moment I entered daycare, but it never really took and I’ve always regretted it. Eating meals on tatami in Japan, I was the only person in the room with a “cheat” pillow -- roughly equivalent in the land of faux pas to asking for a fork. Ease back into some flexibility doing some cross-legged sitting every day (pictured above: my tea set seen from my new, lower perspective).
* **Daily scheduling/routine:** With fewer commitments to be in certain places at certain times, my schedule’s been on a collision course with its destiny as an unstructured, amorphous blob. It’s not working. Go back to having a routine, even if not strictly necessary.
* **Healthy meals:** I like carbs way too much. To do: Eat food, not too much, mostly plants.
* **Technical reading:** Unlike fiction, technical reading requires time and concentration. And unlike fiction, it also makes you learn something. Do some every day.

If things go well, I’ll work my way up to learning Latin later.

Take care.

[1] Air particulates in SF are down ~40% year over year.
