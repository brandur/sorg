+++
published_at = 2019-11-23T02:57:46Z
title = "Actions"
+++

![Sumo strong eggnog](/assets/images/nanoglyphs/005-actions/eggnog@2x.jpg)

Welcome to Nanoglyph, a newsletter about software. It ships weekly, though like most software projects, deadlines are not always met.

This week involved the fulfillment of an old Heroku tradition by making a batch of [sumo strong eggnog](https://github.com/seaofclouds/sumostrong/blob/master/views/eggnog.md). The recipe’s convenient in that you can buy the heavy cream and whole milk involved in cartons of exactly the right size so there’s none leftover. It makes enough to nearly fill three [Kilner clip top bottles](https://www.amazon.com/Kilner-Square-Clip-Bottle-34-Fl/dp/B005N984I8/). If the idea of strong eggnog is even remotely appealing, it’s very much recommended.

The usual format is three links from the week with some commentary, but to keep things dynamic, I’m playing with the format to instead talk about a mini-project in a little more length — like a push version of a blog post.

---

## Migrating to Actions (#github-actions)

A few weeks ago I spent some time migrating the program that generates my blog and these emails over to use [GitHub Actions](https://github.com/features/actions). Its CI goes a little beyond running a test suite by [running the actual deployment](/aws-intrinsic-static) that puts it onto AWS.

It’s automated to the degree that making changes are as easy as merging pull requests, which has been a huge advantage over the years as I’ve had [almost 50 people](https://github.com/brandur/sorg/graphs/contributors) send me PRs to fix typos and grammar. Wherever I am -- in a meeting, on a run, or on a beach in Indonesia -- I hit the "Merge" button from my phone and it's done.

### Commodifying CI (#commodifying)

Travis is one of the most important service innovations of the decade. With a little help from GitHub, they made setting up CI _so_ easy that it rarely made sense not to do it. Even pushing a repo with 50 lines of code that you never intend to look at again, you may as well put a few lines into `.travis.yml` and get CI running just as a little bit of assistance in case future you who wants to make a change and has forgotten most of everything about the program. Even if the CI doesn’t do anything useful, it still serves as a codified reference for how to build it and run the test suite.

Unfortunately, Travis [was acquired](https://news.ycombinator.com/item?id=18978251), and based off the buyer and the subsequent attrition of much of its engineering staff, it’s hard to imagine that the terms were favorable. Things are still running more or less the same as they always were, but given the sheer expense that must be involved in doing free builds for a sizable part of the world’s open source software, some of us have been looking for alternatives.

GitHub’s Actions were a timely arrival. Although described in most of their docs in such grandiose terms so as to obfuscate what they actually do, to put it simply, they describe jobs that will run on certain repository events like a push, opened pull request, or cron — perfect for CI, though they unlock other uses as well. Their major differentiating feature is that the steps of a job can be defined as similar shell commands, they can also defined a Docker container to run. In my own recipe I have shell steps like:

``` yml
- name: Install
  run: make install
```

Intermingled with containers like:

``` yml
- name: Install Go
  uses: actions/setup-go@v1
  with:
    go-version: 1.13.x
```

The path in `uses` refers to a GitHub repository, so the code above refers to the [actions](https://github.com/actions) organization which contains a number of core containers. Versioning is possible in a number of ways:

``` yml
steps:    
  - uses: actions/setup-node@74bc508
  - uses: actions/setup-node@v1
  - uses: actions/setup-node@v1.2
  - uses: actions/setup-node@master
```

And steps can also reference Docker Hub with the magic `docker://` prefix:

``` yml
- name: My first step
  uses: docker://alpine:3.8
```

### Container as unit of modularity (#container-modularity)

This leads into what I believe to be Actions’ most important innovation: the container as a unit of workflow modularity.

Containers have always been modular in that containers reference other containers during builds, but Actions takes it a step further in making containers the convention for encapsulating reusable code — whether it’s checking out code, setting up a Go environment, or deploying to a specific service like AWS, Actions allow complex functionality to be reused easily in a generic workflow.

This is interesting because despite the popularity of containers, many services have been pushing in a much different direction — JavaScript. If you use AWS Lambda or Twilio Functions, workflows should be written in JavaScript and the unit of reuse is the Node package.

As someone who believes that JavaScript is one of the biggest mistakes in the history of computation, this is hugely exciting. If I want to write a package for use with GitHub actions I can do so using the widely understood convention of containers, and I can do so in a language that isn’t terrible.

### What Actions gets right (#right)

Containers are a nice touch. In Travis, you could get some code reuse by manually pulling down scripts and running them, but it was an overly difficult and haphazard process. A single, prescribed system that provides easy built-in modularity is a huge step forward.

Also, acknowledging that builds are really just a series of steps, and that it’s not hugely advantageous to differentiate the category of step like setup vs. build, is a simplification that works:

``` yml
steps:
  - name: Step 1
  - name: Step 2
  - name: Step 3
```

Travis differentiated phases with `install`, `script`, `before_script`, `after_success`. It was a leaky abstraction in two ways:

* Even with the plethora of phases, you’d eventually have to start chaining commands within one of them (usually `script`). Travis allowed separate steps with a YAML array, but made no qualms if any of the failed, so users have to either `set -e` or chain commands with `&&` to get the behavior they wanted.

* The ordering of phases wasn’t that intuitive, so you’d have to look it up.

Steps can be configured individually using `with` to specify options for containers or `env` to specify step-specific variables. I like this because it lets you see which particular steps need specific variables instead of mixing everything into a global env. Explicit usually beats implicit.

``` yml
- name: "Create database: sorg-test"
  run: createdb sorg-test
  env:
    PGHOST: localhost
    PGPORT: ${{ job.services.postgres.ports[5432] }}
    PGUSER: postgres
    PGPASSWORD: postgres
    PGDATABASE: postgres
```

### What Actions continues to get wrong (#wrong)

---
