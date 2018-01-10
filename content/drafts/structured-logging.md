---
title: Structured Logging Revisited
published_at: 2017-12-29T21:36:51Z
location: Calgary
hook: TODO
---

Structured logging looks something like this:

    at=info method=GET path=/ host=mutelight.org fwd="124.133.52.161"
      dyno=web.2 connect=4ms service=8ms status=200 bytes=1653

The basic idea is keep logs useful to humans, but also to
preserve the semantic properties of data you're emitting so
that it can be searched through in a system like
[Splunk][splunk].

logfmt is often useful, but not always reliably parsable,
so configuring programs to emit JSON might be more
advisable for when your apps hit production.

_Just_ printf debugging is wrong. _Just_ debugging by way
of a debugger is wrong. The right way to do do things is to
have a variety of tools at your disposal, each of which can
be deployed given the right situation.

Found [slog][slog]. Configure a destination like `stdout`
or a file, configure a format. A terminal format is useful
in development while JSON can be configured for production.
Can chooose an async drain to offload serialization and IO
to a different thread.

``` rust
let decorator = slog_term::PlainSyncDecorator::new(std::io::stdout());
let drain = slog_term::CompactFormat::new(decorator).build().fuse();
let async_drain = slog_async::Async::new(drain).build().fuse();
slog::Logger::root(async_drain, o!("env" => "test"))
```

Extensibility:

``` rust
impl<'a> DirectoryPodcastUpdater<'a> {
    pub fn run(&mut self, log: &Logger) -> Result<()> {
        log_timed(&log.new(o!("step" => file!())), |ref log| {
            self.conn
                .transaction::<_, Error, _>(|| self.run_inner(&log))
                .chain_err(|| "Error in database transaction")
        })
    }

    ...
}
```

Example:

```
env: test
 step: src/mediators.rs
  Dec 29 13:03:18.511 INFO Start
  step: fetch_feed
   Dec 29 13:03:18.511 INFO Start
   Dec 29 13:03:18.511 INFO Finish, elapsed: 628.866µs
  step: parse_feed
   Dec 29 13:03:18.614 INFO Start
   Dec 29 13:03:18.678 INFO Finish, elapsed: 63.896ms
  step: convert_podcast
   Dec 29 13:03:18.678 INFO Start
   Dec 29 13:03:18.678 INFO Finish, elapsed: 3.993µs
  step: insert_podcast
   Dec 29 13:03:18.679 INFO Start
   Dec 29 13:03:18.679 INFO Finish, elapsed: 952.833µs
  step: insert_podcast_feed_contents
   Dec 29 13:03:18.679 INFO Start
   Dec 29 13:03:18.691 INFO Finish, elapsed: 11.241ms
  step: convert_episodes
   Dec 29 13:03:18.691 INFO Start
   Dec 29 13:03:18.693 INFO Finish, elapsed: 2.062ms
  step: insert_episodes
   Dec 29 13:03:18.693 INFO Start
   Dec 29 13:03:18.708 INFO Finish, elapsed: 14.865ms
  step: save_dir_podcast
   Dec 29 13:03:18.708 INFO Start
   Dec 29 13:03:18.708 INFO Finish, elapsed: 665.530µs
  Dec 29 13:03:18.709 INFO Finish, elapsed: 198.321ms
```

Implementation:

``` rust
#[inline]
fn log_timed<T, F>(log: &Logger, f: F) -> T
where
    F: FnOnce(&Logger) -> T,
{
    let start = precise_time_ns();
    info!(log, "Start");
    let res = f(&log);
    let elapsed = precise_time_ns() - start;
    let (div, unit) = unit(elapsed);
    info!(log, "Finish"; "elapsed" => format!("{:.*}{}", 3, ((elapsed as f64) / div), unit));
    res
}

#[inline]
fn unit(ns: u64) -> (f64, &'static str) {
    if ns >= 1_000_000_000 {
        (1_000_000_000_f64, "s")
    } else if ns >= 1_000_000 {
        (1_000_000_f64, "ms")
    } else if ns >= 1_000 {
        (1_000_f64, "µs")
    } else {
        (1_f64, "ns")
    }
}
```

[slog]: https://todo
[splunk]: https://todo
