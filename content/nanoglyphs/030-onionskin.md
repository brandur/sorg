+++
image_alt = "Union Square in the rain"
image_url = "/photographs/nanoglyphs/030-onionskin/union-square-rain@2x.jpg"
published_at = 2021-12-31T23:41:48Z
title = "No Docker; Onionskin Stacks"
+++

Readers --

I've been thinking all morning about a single short topic that I could use to close out the year with one last issue of _Nanoglyph_. My mind wandered between a few ideas before settling on the evergreen vitriolic battlefield of microservice versus monolith versus monorepo. But as I was jotting down notes, I could already feel it ballooning to a multi-thousand word essay ("monolithic proportions"?) that'd never ship on time.

So instead, I'm going to leave you with a short screed on what may be one our controversial internal tech choices at Crunchy (or within its cloud division at least): no Docker, no containers.

---

[OCI containers](https://opencontainers.org/) have in a period of less than a decade become the gold standard for deployment. Originally you'd hear about them only in association with Docker, but they've since grown to a wide standard that's in use by the vast majority of major tech companies. Every major cloud provider provides a service that deploys containers. Cutting edge operational paradigms like Kubernetes and serverless use containers as basic foundational building blocks. Even companies like Heroku that were using their own version of containers before Docker existed, have since retrofitted Docker-style containers into their product. In short, containers are the future, and they're already everywhere.

Run a quick Google search and you'll find them lauded with hundreds of operational, organizational, and technical benefits. Some major ones:

* They produce an isolated environment for an app to run in, and one which is much lighter weight than a traditional VM, with better performance and less overhead.

* They're portable -- the container engine provides a layer of abstraction between a container and the underlying OS, allowing it to run regardless of whether it's hosted on Linux, Mac, or Windows.

* They ease pain around setup and development. Containers encapsulate recipes for how they're initialized, allowing complex applications to be bootstrapped quickly and with minimal context.

* They're scalable -- the whole scheme requires encapsulation, so many of the same container can be spun up simultaneously, all from a single image.

And yet, if you were to examine the repos for our backend, our API, or our frontend, you wouldn't even find a `Dockerfile`. So what are we even doing over here?

## The deepness (#deepness)

In [029](/nanoglyphs/029-path-of-madness) I talked about the development experience at Stripe. Let's briefly visit that again.

Developers would start their environment via one simple command: `pay up`. This would kick off a plethora of activity that among other things would:

* Connect to a cloud development box (provisioning one if necessary).

* Start an rsync process to send local files up the cloud and watch for new changes.

* But remote processes for local commands to talk to.

* ... and about 400 other miscellaneous things.

Back in the old days, I used to know pretty much exactly how it worked. Big Ruby processes start up very slowly, so it's not uncommon to start an environment once and fork pristine processes as required, a model established by [Zeus](https://github.com/burke/zeus). We'd layered on a NIH project called "Hera", but it worked roughly the same. After getting Ruby and its dependencies up and running it was just a matter of spinning up the constellation of adjacent daemons -- Mongo, ElasticSearch, Redis, etc., and bingo.

Our dev productivity team had been pushing remote development for some time, but it had significant downsides -- it was slow, and naturally made it impossible to work offline, something that some of us still did back then. A small group of engineers collaborated to maintain an informal Hackpad titled "local development setup" with homegrown instructions on how to get the stack running minus the cloud bootstrap. Along with conveying additional speed and keeping us effective in low connectivity environments, there was another side benefit -- every person who'd run through that document had a better understanding of how the stack worked than the other 95% of the engineering contingent.

But as time went by, the stack got deeper. The stack got wider. The stack grew adornments, and it grew thorns. As the flywheel accelerated, trying to keep pace with changes made in the cloud became increasingly untenable, and one by one, those of us who'd been running local were forced to migrate to the blessed path. Eventually, I was running `pay up` just like everyone else -- and just like everyone else, not really understanding the specifics of what was happening within.

And for a company of that size, this might've been the right answer. Engineers run a command, a whole bunch of magic occurs behind the scene, and from there they have a mostly functional development environment. This is very similar to the model put forward by containers -- run something like `docker compose up`, and in one command you've got your whole contingent of services up and running just like that. It's fast, and anyone can do it.

## The murkiness (#murkiness)

But you know what they say about things that sound too good to be true. The model is largely functional, but comes with bad along with the good.

A problem is that thanks to the near perfect opacity, the majority of users don't understand how anything works, and lose the ability to diagnose problems and any hope of divining their way to a solution. In the case of `pay up`, the underlying infrastructure was _so_ complex that the _only_ remediation for 95%+ of the org when encountering the problem was to report it to someone else and get them to fix it. Not only does this mean that problems now eat at least two peoples' time (and usually more), but it's also a negative feedback loop: problem appears, problem is reported, debugging skills atrophy, problem appears, problem is reported, ...

An opaque stack also means that significant complication can be hidden below the surface thanks to the sophisticated facade. This often includes complication that by all rights shouldn't exist -- akin to cleaning your room by shoving everything under the bed instead of being forced to address each item head on.

## Onionskin stacks (#onionskin)

Back to Crunchy: an alternative to the ease-of-use of a single Docker command to do setup is to keep your stack so thin that you can see through it. But despite that, strong and lightweight -- an onionskin stack.

Here are our README instructions for bootstrapping and running the API's test suite:

``` sh
psql < sql/raise_databases.sql
migrate -source file://./migrations -database $TEST_DATABASE_URL up
go test ./...
```

That's it -- three commands.

Granted, it depends on a few external prerequisites (Postgres, Go, [direnv](https://direnv.net/), and [migrate](https://github.com/golang-migrate/migrate)), but all common software that most engineers at the company have already, which is easy to install in case they don't, and none of which needs to be upgraded very often.

Go helps a lot in keeping things this simple -- if this is the first time `go test` is being run, the command will automatically detect that dependencies need to be installed and go fetch them. Also, practically every Go dependency is written in Go, so installing those dependencies works with almost 100% reliability.

Go is good, but that said, our Ruby app (the database state machine) isn't too far off:

``` sh
asdf install
gem install bundler
bundle install
ALLOW_DB_LOCAL_SETUP=true bundle exec rake db:localsetup
bundle exec rspec
```

It needs Postgres and [asdf](https://github.com/asdf-vm/asdf) to fetch Ruby, but not much else.

Not visible in these command sets are the improvements to ease-of-set-up that have trickled in to many modern stacks over the years. Circa 2013 you would have _wanted_ Docker to compose your Ruby environment because there was so many steps to get to a successful installation. Nowadays, between improvements in version managers, package managers, and more streamlined dependency sets (e.g. jettisoning pain-in-the-rear dependencies like [Nokogiri](https://nokogiri.org/) that never quite compile right), it's much more plausible to be running a thin Ruby stack with no orchestration involved.

## The joy of container-free life (#container-free)

So why are we avoiding containers? Am I filibustering to gloss over what can only be explained by an elaborate rationalization for [Neo-Luddism](https://en.wikipedia.org/wiki/Neo-Luddism)? Well, that's not how we'd put it at least. In a nutshell:

* **Speed:** Running a process or database outside of a container is faster than running it inside a container, and has no boot overhead. Our development loops are ~instant (see [029](/nanoglyphs/029-path-of-madness)).

* **Insight:** Not having an additional abstraction layer forces us to keep our stacks thin and simple. And because every engineer is interacting with every element in the stack directly, they can fix their own problems, and build useful debugging muscle while doing so.

* **Commodity:** Each component in our stacks -- Go, Postgres, Ruby, Bundler, etc. -- is commodity software. It's actively developed, and we gain automatic benefit from advancements. If there's a problem, you can google it. Compare this to NIH infrastructure that's developed only by you and where by necessity all troubleshooting comes from you or another internal engineer.

A key element in making sure this works is keeping our stacks _aggressively_ thin. A notable omission from both of the above is Redis, a component so common these days that it's probably found in the majority of production stacks around the world. (And no Kafka either!) I like Redis a lot, and we may yet bring it or other elements in eventually, but are living on a just-Postgres model for as long as possible.

None of this means that we're excluding the possibility of using containers either. For now we still deploy on Heroku via `git push` (maybe our second most controversial tech decision), but if we migrated somewhere else, it's likely we'd write some thin `Dockerfile` shims because as stated above, OCI is more or less the de facto standard of cloud deployment, and isn't going anywhere.

So there you have it. These days, every engineer and their dog will preach the virtues of minimizing dependencies and keeping things simple, but few actually do it. Maybe we don't either, but we're giving it our best shot.

---

## Web dispatches (#dispatches)

A few links from around the web:

* [**The Road to Valhalla**:](https://openjdk.java.net/projects/valhalla/design-notes/state-of-valhalla/01-background) Valhalla is a major push on the JVM to better align its runtime with hardware by reducing indirection and boxing. This is largely a solution for a very Java problem, but the scope of a project like this so late in the language's lifecycle is impressive, and will eventually be a major boon to countless important software stacks around the world (right after they get log4j sorted out).

* [**Can "Distraction-Free" Devices Change the Way We Write?**:](https://www.newyorker.com/magazine/2021/12/20/can-distraction-free-devices-change-the-way-we-write) (paywall-free [archive link](https://archive.md/0XWsG)) Computers make everything so easy that writing becomes hard. This is a deep dive into various apps and hardware that help with focus. 2021 was yet another mediocre writing year for me, and I'll be trying more than a few tricks next year to help me do better.

* [**Dear Self; We Need To Talk About Social Media**:](https://acesounderglass.com/2021/12/04/dear-self-we-need-to-talk-about-social-media/) I'm not going to lie -- I'm addicted to Reddit, and it makes working on any large project requiring lengthy focus difficult. I like Elizabeth's concept of a "Quiet" state and the various techniques around how to achieve it.

Happy New Year -- see you in 2022.
