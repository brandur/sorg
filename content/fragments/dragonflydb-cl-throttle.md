+++
hook = "The rate limiting command provided by redis-cell is now available in DragonflyDB."
published_at = 2023-03-12T15:51:14-07:00
title = "`CL.THROTTLE` implemented in DragonflyDB"
+++

Way back when Redis Modules were first introduced I wrote one called [redis-cell](https://github.com/brandur/redis-cell). I wrote it because in my experience you couldn't invent a more perfect fit for Redis than rate limiting, and we'd used it for such at practically every job I'd ever had.

The command provides easy limiting in a single Redis call:

```
CL.THROTTLE user123 15 30 60 1
               ▲     ▲  ▲  ▲ ▲
               |     |  |  | └───── apply 1 token (default if omitted)
               |     |  └──┴─────── 30 tokens / 60 seconds
               |     └───────────── 15 max_burst
               └─────────────────── key "user123"
```

I'd still contend that it's useful, but haven't been maintaining it well. It picked up a few users, but one could make the argument that using a Redis Module for this isn't enough of advantage to be worthwhile. You might gain a little speed, but have added overhead in configuration hassle and a non-standard Redis set up.

But here's something that's cool: `CL.THROTTLE` was recently [implemented in DragonflyDB](https://github.com/dragonflydb/dragonfly/pull/714) by [Daria Sukhonina](https://github.com/ZetaNumbers). [DragonflyDB](https://dragonflydb.io/) is a Redis-compatible DB written for speed, and seems to be more open about what kind of default utilities they include. Having it available out of the box is a better fit for `CL.THROTTLE` because you get access to it without configuring anything, and you won't have to worry about its availability across different cloud hosts. It also gives you better protection against delinquent maintainers like me.