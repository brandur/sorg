---
title: A Hundred Little Things to Like About Rust
location: San Francisco
published_at: 2018-03-07T16:02:00Z
hook: TODO
---

* **Rustup**: 

* **Rustfmt**: 

* **Clippy**: 

* **The module system with file system conventions:** You
  start any new module (or tree or modules) in a single
  file (`endpoints.rs`), and break it up as it starts to
  grow and when it makes sense (`endpoints/mod.rs`,
  `endpoints/middleware.rs`, etc.). This makes new projects
  fluid to build and keeps mature projects well-organized.
  A file's contents inherits a module name from its
  filename and that keeps indentation in files minimal (as
  opposed to wrapping it all in `mod endpoints {`.

* **Snake-case module names and camel-case types:** It's
  trivial to differentiate a module (`view_models`) and a
  type (`ViewModel`) at a glance and it also guarantees no
  import collisions.

* **A compiler that warns you when you break naming
  conventions:** Even more important than the conventions
  themselves is a toolchain that respects them. If you name
  a module camel-case or a type snake-case, the compiler
  will tell you about it.
