+++
image_alt = "A wind turbine near Vancouver"
image_url = "/photographs/nanoglyphs/021-ides/wind-turbine@2x.jpg"
published_at = 2021-02-28T17:09:58Z
title = "Editors, IDEs; O Gods, What If Microsoft Was Right All Along?"
+++

I still use Vim. I still write Ruby. But I'm having a crisis of faith. I'm writing Java, in an IDE so heavy that it practically has manifest physical weight ... and enjoying it.

Welcome to _Nanoglyph_.

---

As mentioned in [020](/nanoglyphs/020-alfred) last week, I've been doing a lot of work in Oracle's choice programming language. After many long years, it's been tacitly [1] acknowledged internally that Ruby's molluskesque performance characteristics, poor reliability guarantees (interpreter, no static typing), and continuing non-existent (or [very nascent](/nanoglyphs/018-ractors)) concurrency story are factors significantly holding back software speed and quality.

Why Java? There's no single answer. My interpretation: because Google likes it, and we want to be like Google. But although it could be better, it also could be a lot worse. Java is a rock solid B- of programming languages. (I wanted to say that it's the "C+" of programming languages, but that's confusing.)

## Boilerplate reduction (#boilerplate)

Java SE 5 introduced support for generics:

``` java
Map<String, List<String>> myMap = new HashMap<String, List<String>>();
```

Generics are an idea that'd been around a while, and already a familiar concept (if not exactly the same) to people working in other popular languages at the time, like in C++ with its templates. Java's generics are considered a very average implementation of the idea because unlike say C#, they use type erasure, which made them easier to implement, but is a strict disadvantage when it comes to runtime performance and memory usage. But, that's a story for another day.

The syntax for generics was a real chore to type out, so minor quality of life improvements were made over the years. Java 8 improved type inference so generic variables could be declared with a shorthand `<>` diamond instead:

``` java
Map<String, List<String>> myMap = new HashMap<>();
```

Java 10 went further by introducing the C#-inspired `var` keyword, which was even more succinct, and more flexible:

``` java
var myMap = new HashMap<String, List<String>>();
```

## Let the machine do the work (#machine-work)

We're a far throw from being bleeding edge adopters of anything, and are still sitting somewhere between Java 8 and 10, meaning we use the "diamond" syntax for generics.

It's mostly fine to type by hand, but is still often burdensome. Say you have an invocation like:

``` java
sendMap(getMyMap()));
```

And you realize that you want to have a local variable:

``` java
Map<String, List<String>> myMap = getMyMap();
sendMap(myMap);
```

That's still quite a few extra characters to add, which is where your IDE comes in. Using IntelliJ, you bring your cursor over `getMyMap`, bring up "intention actions" with `⌥ + Enter` and tell it to extract a local variable. Thanks to great static typing, IntelliJ already knows that's a `Map<String, List<String>>`, and in the blink of an eye, the refactor is complete.

That's one simple example, but a pattern that's repeated again and again in modern day Java development. The language itself is punishingly verbose, full of warts, and lacking the syntax niceties of its modern counterparts (or its contemporaries that were able to more successfully modernize, e.g. C#). But in practice, it doesn't matter that much because the machine is doing all the work.

And the extra annotation pays off fast. Coming down from Ruby (even [Sorbet-annotated Ruby](/nanoglyphs/015-ruby-typing)), it's incredibly noticeable how much less time you're spending in a debugger. I think I've used one about once in a month's worth of Java development, versus dropping into Pry twenty times a day in Ruby. IDE-driven autocomplete coding means you introduce fewer problems in the first place, and a well-informed compiler means you catch most of the ones that do make it in well before testing.

I was surprised to find how much I was enjoying coding again. I couldn't even remotely be called a fan of Java, but the feeling of slogging through mud all day had slipped away.

---

{{NanoglyphSignup .InEmail}}

## Full circle (#full-circle)

Undoubtedly, many of you are already using powerful IDEs, so none of this will be new to you. The weird thing is that it's not new to _me_ either.

I started my career as a C# developer way back in the prehistory of the 2000s. At the time, Visual Studio wasn't just a nice augmentation for C#, it was a necessity, often the only way to access the compiler and build toolchain. It was good too -- IntelliSense and its refactoring facilities were great.

And yet, when I joined Heroku around 2011, I was happy to see it go. Exchanging C#'s heavy syntax for Ruby's expressive DSLs was so freeing, and we were all into the idea of small, sharp tools, running small programs from our terminals instead of through a monolithic IDE that even on the bulkiest of machines, was slow to open a file or bring up and a context menu, and kept your fans spinning at full speed all day. It felt like the wild west, like were pushing towards a leaner, meaner, new frontier of computing.

But do that enough years, and you get tired. Tired of squashing the tenth bug of the day that a compiler and better type system would have detected trivially. Tired of your editors painstakingly-configured LSP integration not working right or suddenly failing mysteriously, leading to wasted hours of troubleshooting. Tired of looking up the API reference for `Net::HTTP` for the 2,000th time. Tired of your duct tape CTags-based jump-to-definition working 60% of the time.

IDEs _are_ good. Microsoft was right all along.

<img src="/photographs/nanoglyphs/021-ides/chairlift@2x.jpg" alt="A chairlift off season" class="wide">

## Jetbrains: Good, but heavy (#jetbrains)

While I'm a fully-licensed IntelliJ user at work, I needed something that would work for personal development too. Although best known for IntelliJ, Jetbrains now produces a full suite of IDEs for most of the languages you'd care to use, like [CLion](https://www.jetbrains.com/clion/) (C/C++) or [GoLand](https://www.jetbrains.com/go/).

Unfortunately, licensing isn't cheap. They do have a progressive model where individual licenses are much cheaper than corporate, but it's still a subscription model starting at $249/year for the full product pack, dropping to $199/year for year two, and going to $149/year after that.

And Jetbrains IDEs do have a few other downsides:

* They are _heavy_, and it shows in their everyday "feel". e.g. Not being able to use shortcuts in the interface while a build is running.

* On that same note, one of the beautiful things about M1 Macs is that for the first time in laptop history, you're free from the wall outlet. The battery lasts as long as you need and then some. I haven't tried IntelliJ on an M1 Mac as of yet, but if the battery life I get on an Intel is any indication, it won't be good.

* While they added support for M1 Macs, many commenters report a subpar experience.

* A vast number of shortcuts are based on `Ctrl` and the Function keys, all of which of which are awkward to use on Mac. This is more Apple's fault than Jetbrains', but still, reality is reality.

None of these problems will stop me from demoing GoLand at some point, but I wanted other options.

## VSCode (#vscode)

VSCode is the decided victor as the programming world's favorite new editor, and in less than a decade has gone from mere idea to on par in sophistication with the best in class.

I booted it up again recently to see how it faired against my experience with IntelliJ. There are so many things to love about VSCode (extensible, configurable, fast, free!), but one of the places it really shines is its language flexibility. I tried both a Go project ([sorg](https://github.com/brandur/sorg)) and a C project (Postgres), installed the official plugins for Go and C/C++, and got to work. While there's certainly more prompts to install various tooling compared to IntelliJ, within a few minutes of a fresh installation I was up and running.

I was most curious about the quality of jump-to-definition and refactoring, and was largely impressed. Support for that in Vim for languages like Go/C is decent these days, but I've often had problems with accuracy and flakiness. It was nice to have something that worked ~100% of the time, and with zero configuration. This is an area where IntelliJ really shines, and VSCode was comparable, with its major downside being that it has far fewer refactoring options, with renaming being the only major one support out of the box. Some extensions have additional refactoring commands like Go's "extract to function" and "extract to variable", but none of the ones I tried worked (score one for IntelliJ). Jump to definition, on the other hand, was perfect.

I tried VSCode a year ago, but eventually churned back to Vim because while the former was good, there were just enough little quirks in the interface to be annoying. Like you'd be trying to be keyboard only, and focus would get stuck in some random part of the UI, forcing you to use the mouse to rescue yourself. I'm already seeing little problems of this sort crop up, but am going to try and stick with it anyway.

---

## The push model (#push-model)

With Substack's spectacular rise in popularity, newsletters are a hot topic recently. Can they actually be an adequate replacement for traditional news media, which doesn't seem to have a working business model anymore, or for social media, which more and more just seems to be a medium that the human mind isn't well-suited for?

I suspect that the importance of Substack is a little overblown, but I was looking through my inbox today, and I've gone from subscribed to about two newsletters ~3 years ago (and when I say "newsletters", I'm talking about newsletters from people, not from marketers/companies, which like all of you, I'm constantly playing whack-a-mole with), to somewhere around 40 today.

A big difference in terms of how I read newsletters versus my RSS feed or articles from a link aggregator is that I almost invariably read them all top to bottom, usually because the average quality of content is that much better, something I'd link to two major factors:

* There's friction to sending a newsletter. It's not free, and it's a write-once deal where after it's sent, it's sent. This leads writers to be more rigorous around content and editing.

* My own selection bias. I'm picky about what's allowed to land in my inbox in the first place, and if I'm no longer interested in reading someone's content top to bottom, I'll unsubscribe.

I don't know whether newsletters can save us. Their strengths are also their weaknesses in that while they're more resistant to defamation and censorship, they're also less discoverable. They almost certainly suffer from the same power-law dynamics as we see on sites like YouTube, where a handful of users have a huge number of subscribers, while the long tail has almost none. That's fine, unless the ability to monetize is an important aspect of the platform, and if this is supposed to be a replacement for print media, it kind of is.

### The microtransaction, finally? (#microtransaction)

Personally, I'm hoping that where Substack goes next is to be the first company to put in a working "microtransaction" patron model, which is something we've talked about since the inception of the internet, but for which a popular, practical implementation has never emerged. I don't necessarily want to pay $5/month for every newsletter I'm subscribed to, but if I could say, put $25/month into a "newsletter pot", and have Substack distribute it every month based on the set of newsletters I'm currently subscribed to, I would be really into _that_.

Important new model, or trendy craze -- what do you think?

Until next week.

[1] Tacitly ("in a way that is understood or implied without being directly stated") is the perfect word for this case. No official stance will ever be taken.
