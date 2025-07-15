+++
image_alt = "Courtyard in Berlin"
image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/042a-crush/courtyard@2x.jpg"
published_at = 2024-05-28T14:04:42+02:00
title = "Berlin. N+1s. RBS. Crush!"
hook = "Berlin, don't go to Manchester, a general pattern for eradicating N+1s, Tobi Lutke on performance, trialing RBS in Ruby, and the disappointing legacy of iPadOS."
+++

Readers —

I'm just back from a month in Berlin. Going for a music festival, but taking the opportunity to stay a few extra weeks in one of my favorite cities.

Before anything else, some usual housekeeping: I'm Brandur, and this is my newsletter _Nanoglyph_. It's been comically delinquent lately, and you may have signed up months ago and this is the first blast you've gotten so far, If you hate it already and don't want to see anymore of them, you can [instantly unsubscribe in one easy click](%unsubscribe_url%).

This newsletter is [also available on the web](/nanoglyphs/042a-crush). I recently spent some time [adding dark mode support](/fragments/dark-mode-notes), something that sounds easy until you realize you have a few thousand lines of legacy CSS to retrofit.

In this issue: Manchester, a general pattern for eradicating N+1s, Tobi Lutke on performance, trialing RBS in Ruby, and the disappointing legacy of iPadOS.

---

## The trial (#the-trial)

Spending a month in one German city might seem like a weird way to spend time, but being a little more stationary was an intentional choice.

Last year, I went to Europe around this time, and also stayed for around a month, but it was a completely different trip. I'd established an aggressive itinerary for myself across France, the UK, and Germany, not staying in any one place for more than a few days at a time. Across the span of the month, I must have packed and unpacked my bag fifteen times, before and after lugging it to and from various airports and train stations.

<img src="/photographs/nanoglyphs/042a-crush/manchester@2x.jpg" alt="Manchester" class="wide" loading="lazy">

This culminated in the greatest travel fiasco I've ever experienced, where I went through security at Manchester only to find that my flight was cancelled. This is _theoretically_ not that big of a deal in Europe where strong regulatory frameworks exist to protect consumers, and which mandate that cancelled flights be replaced with equivalent ones, but airlines have found a not-so-surprising loophole: they can't rebook you on a new flight if there _are no other flights_. I searched high and low, and not only were there no flights to my destination with an available seat out of Manchester for a good week's time, but even were I to take the train back down to London, there was nothing out of Heathrow or Gatwick either.

Knowing that I wasn't getting on a plane that night led to my next surprise discovery: in the UK, one does not simply _leave_ an airport. There is no exit. You either get on a plane, or personnel from your _specific_ airline must escort you out. What if it's late and there are no airline personnel around? Well, heaven have mercy on your immortal soul dear traveler, you have just entered the most Kafka-esque labyrinth of pain and frustration found anywhere in the western world. Three hours later, and after more negotiating, cajoling, pleading, yelling, and waiting than I care to remember, I escaped, but even to this day I think that if I'd been just a little less lucky that day, to this day I might still be in that hellish prison, sleeping on benches and sustaining myself on chicken nuggets from the food court like [Mehran Nasseri](https://en.wikipedia.org/wiki/Mehran_Karimi_Nasseri).

I spent the next two days blasting across Europe by train, from London to Paris to Amsterdam to Berlin, first by local rail, then EuroStar, then slower and more scenic German trains. It's a fun memory in retrospect and I have no regrets, but it did make me establish a new, important life rule: don't go to Manchester.

<a href="https://www.youtube.com/watch?v=UxsYbT3sQZU">
    <img src="/photographs/nanoglyphs/042a-crush/avi@2x.jpg" alt="Avi from Snatch" class="wide" loading="lazy">
</a>

This time, I mostly stayed in one spot, and it was great. The busy travel version of seeing Europe is nice too, but logistics, moving, and sightseeing take so much time out of the day that there's not much time for anything else.

---

## If I had more time, I would have written a shorter letter (#shorter-letter)

I got quite a bit of technical writing done on the old continent, publishing three articles and ~15 shorter fragments, but this newsletter's still be totally delinquent (what else is new?). By the rules of my internal intuitive framework, the longer I haven't sent one, the more elaborate the first issue after a pause must be. Ideally including something big, loud, and _bold_.

Unfortunately I don't have much in that category right now, and as I've been thinking about what to send, I keep coming back to the realization that if anything, these dispatches should be shorter, not longer. As much as I appreciate the format of long, detailed newsletters full of swathes of interesting information, I like short, directed ones even more, and other people are probably the same.

This one's not that short, but it's not that meaty either. I've got a few technical pieces to cover, mostly outlinks, and one iPad rant.

---

<img src="/photographs/nanoglyphs/042a-crush/berlin-1@2x.jpg" alt="Berlin 1" class="wide" loading="lazy">

## N+1s and data loading

Ever visit a website and wonder why it's so slow despite your fast connection?

Well, software is complicated, so it could be any number of different reasons, but one of the top candidates are N+1 queries.

Imagine you're building a basic social network app. In the beginning, things are simple. All you want to do is pull the ID of a logged in user out of their session, and fetch their info so you can render a little avatar icon for them. Easy. With one database query it's hard for much to go wrong.

Now, say this user has a network of friends. You want to render them in a pretty list on their profile, so first you fetch the IDs of all friends, then in a loop iterate through them, fetch information on each friend, and render that friend's avatar icon next to their name.

This is the essence of an N+1 query. The "1" is getting the IDs of all friends, and the "N" is pulling information on each user inside a loop, with the sum being an arbitrary number of database queries required.

``` ruby
user = User.find(user_id)       # 1 query

friend_avatars =
  user.friends.map do |friend|
    friend.avatars.default      # N queries (1 query in loop of N items)
  end
                                # total: N + 1
```

Magical lazy loading ORMs like ActiveRecord make N+1s _really_ easy to introduce and really hard to notice. Modern databases are extremely fast, so even if rendering a page takes hundreds of database queries, it's still pretty quick to get through them all, and most of the time neither user nor developer notices. But they're an entropic problem: as new features are added, new queries come in 1 by 1 by N, and a site's load times degrades in small increments.

My guess would be that 99% of operating backend software has no generalizable solution for this problem. The way it generally works in industry at most shops is that load times continue to slowly degrade, and once they become truly intolerable, some lucky programmer will be tasked to do an optimization pass, where they go in armed with logs and maybe a profiler to patch up the slowest parts of code through techniques like eager loading. The problem is never definitively fixed, but the worst offenders are rehabilitated to get load time back within reason, and the clock starts all over again.

Projects like Rails have been leaders in coming up with builtins to help avoid the problem through features like [strict loading associations](https://rubyonrails.org/2020/12/9/Rails-6-1-0-release#strict-loading-associations), although these are very specific to the framework in question.

We're building all our domain logic in Go right now and we're not using a specific framework, so we came up with a code pattern that I call "two-phase load and render" that's general enough that it could be applied in any language/framework. It's taken our N+1s to zero, and even more importantly, makes it difficult to introduce new ones because doing so requires breaking convention.

I published [an article fully explaining it](/two-phase-render) with diagrams and code samples. It's long and will seem overwrought, but I stand by that having a framework for good performance is extremely important. Very few shops do, and by the time they realize they have a problem, it's nigh impossible to fix.

<img src="/assets/images/two-phase-render/render_load_bundle.svg" alt="Rendering a load bundle.">

---

<img src="/photographs/nanoglyphs/042a-crush/berlin-2@2x.jpg" alt="Berlin 2" class="wide" loading="lazy">

### Root of all evil (#root-of-all-evil)

Relatedy, Shopify's Tobi tweeted a [great rant on performance in software](https://x.com/tobi/status/1787139157078188180) last month.

His claim, which I agree with, is that the phrase "premature optimization is the root of all evil" has backfired in a major way. Early optimizing your programs to the bit level isn't good, but neither is not think about performance _at all_ when building things:

> For software engineering, my sense is that the phrase “premature optimization is the root of all evil” has massively backfired. Its from a book on data structures and mainly tried to dissuade people from prematurely write things in assembler. But the point was to free you up to think harder about the data structures to use, not leave things comically inefficient. This context is always skipped when it’s uttered.  
>
> Not all fast software is world-class, but all world-class software is fast. Performance is _the_ killer feature.  

Non-technical end users do care, even if they don't exactly know all the details of what they care about:

> Every user of your products cares exactly as much about latency as engineers do when typing in their terminal. They just don’t have the words to describe what they don’t like about the experience and neither should they.

Like with N+1s, making an app fast retroactively tends to be really hard. There may be a few possible changes that produce outsized performance gains, but there's also likely to be thousands/millions uses of fundamentally slow patterns that started as out as "this is sorta bad, but good enough for now" and which proflierated to an ungodly extent over the intervening years. Fixing all of them retroactively is probably difficult, and may be borderline impossible.

This is an area where I'm absolutely certain that an 80/20 exists that'd work for almost everyone. Don't go miles out of the way to optimize code, but also don't introduce terrible patterns that are predictably going to cause major grief later on. If one wasn't so predictable, but is now noticeably causing trouble, _fix it_. Don't wait until it's replicated endlessly and will be orders of magnitude more effort to repair.

See also [_Nanoglyph 037: Fast as a Service_](/nanoglyphs/037-fast).

---

<img src="/photographs/nanoglyphs/042a-crush/berlin-3@2x.jpg" alt="Berlin 3" class="wide" loading="lazy">

## Test driving RBS (#rbs)

I've long been a proponent of type checking, and quite critical of Ruby for not having this feature. At Stripe we built and used Sorbet (that's the royal "we" -- I was a user, not a contributor), and it was created in the nick of time. Given Stripe's size and scope, probably one of the biggest technical saves in history.

We had so much Ruby code that it was becoming unmanageably difficult to work on. CI was slow. Booting up even on test case to run was slow. Judging by variable names only, it was impossible to understand the specifics of any implementation (is `users` an array of `User` objects or user IDs?), so changing it was laborious, time consuming, and extremely error prone. Every month that went by, adding features and shipping product got noticeably harder.

Insufficient test coverage was also a perennial issue. As a reminder, Ruby's an interpreted language, meaning that you have no idea whether code works ahead of time. The _only_ way to know whether it does is to run it. The common prescription is to write tests, but very, very few codebases are going to have [100% test coverage](/fragments/100-percent-coverage), which is the only way to be sure that what you're shipping is safe. If even _one branch_ isn't tested it can take production down.

``` ruby
# Is `last_login` nil in most of your tests? It probably is ...
if user.last_login.nil? && !user.login_history.any_abuse?
  # do stuff
end
```

``` ruby
main.rb:9:in `<main>': undefined method `login_history' for #<User:0x00000001046d69d8> (NoMethodError)

if user.last_login && user.login_history.any_abuse?
                          ^^^^^^^^^^^^^^
```

This happened to us an untold number of times, and when it did, our users felt it, viscerally, as potentially hundreds of thousands of requests failed in the time it took to roll back.

The key unlock with Sorbet was the introduction of a fast type checking pass. Methods and variables were annotated with types, and as these annotations continued to expand and were backfilled to existing code, the call graphs inferred by Sorbet became wider and richer. We got to a point where the largest classes of common problems could be detected and solved without having to run even one test, and even without leaving your editor (Sorbet has an LSP). It helped with harder problems too, like finding bad code in branches that realistically no test was ever going to be written for.

``` ruby
# typed: true
extend T::Sig

sig {params(name: String).returns(Integer)}
def main(name)
  puts "Hello, #{name}!"
  name.length
end

main("Sorbet") # ok!
main()   # error: Not enough arguments provided
man("")  # error: Method `man` does not exist
```

To the disappointment of many (including myself), Matz made the executive decision that type annotations in code are bad, and that Ruby core would be taking a different direction, involving a companion `*.rbs` file for every `*.rb` file containing type annotations.

RBS-based projects would maintain two files for every one of their sources, updating the RBS anytime an instance variable or method was added or changed (even private ones).

``` ruby
module River
  interface _Driver
    def advisory_lock: (Integer) -> void
    def job_get_by_kind_and_unique_properties: (Driver::JobGetByKindAndUniquePropertiesParam) -> JobRow?
    def job_insert: (Driver::JobInsertParams) -> JobRow
    def job_insert_many: (Array[Driver::JobInsertParams]) -> Integer
    def transaction: [T] () { () -> T } -> T
  end

  module Driver
    class JobGetByKindAndUniquePropertiesParam
      attr_accessor created_at: [Time, Time]?
      attr_accessor encoded_args: String?
      attr_accessor kind: String
      attr_accessor queue: String?
      attr_accessor state: Array[jobStateAll]?

      def initialize: (kind: String, ?created_at: [Time, Time]?, ?encoded_args: String?, ?queue: String?, ?state: Array[jobStateAll]?) -> void
    end
  end
end
```

That was a few years ago. Recently I had cause to be writing a small Ruby library, and wanted to see what it was like to be using Ruby's prescribed typing system these days. The toolchain has a few components:

* [An RBS CLI](https://github.com/ruby/rbs) for generating `*.rbs` files from `*.rb` code.

``` sh
rbs prototype rb lib/driver.rb > sig/driver.rbs
```

* A tool that runs static analysis on code called [Steep](https://github.com/soutaro/steep).

``` sh
$ bundle exec steep check
# Type checking files:

.............................................................................................F...

lib/client.rb:194:40: [error] Cannot pass a value of type `::Integer` as an argument of type `(32 | 64)`
│   ::Integer <: (32 | 64)
│     ::Integer <: 32
│
│ Diagnostic ID: Ruby::ArgumentTypeMismatch
│
└           FNV.fnv1_hash(lock_str, size: 512)
                                          ~~~

Detected 1 problem from 1 file
```

* An "RBS Collections" bundler for bringing in type annotations from third party dependencies. Very few gems are typed, so it also provides a public registry that provides generated annotations for common ones.

``` sh
# Create rbs_collection.yaml
$ rbs collection init

# Resolve dependencies and install RBS files from this repository
$ rbs collection install
```

* A VSCode plugin that produces some basic error diagnostics, although not much else right now (e.g. no jump-to-definition, no symbol renaming)

It all worked, but I couldn't shake the feeling that it's still a long way from achieving parity with Sorbet, let alone to a point of being as friendly and useful as a good compiler like `clang`, `go`, or `rustc`.

I wrote a longer piece with more details in [Ruby typing 2024](/fragments/ruby-typing-2024).

---

<img src="/photographs/nanoglyphs/042a-crush/crush1@2x.jpg" alt="Apple's Crush! ad" class="wide" loading="lazy">

## Let's not destroy the creative tools quite yet maybe? (#creative-tools)

A few weeks ago Apple announced their new iPad Pro, similar in form and function to the last iPad Pro, but with a new M4 processor (which has once again leapfrogged the competition), "tandem" OLED display, and 5.1 mm body, as thin as the original iPod Nano.

Far more interesting than the iPad Pro's release though was the meta commentary surrounding it. They ran and [an campaign called "Crush!"](https://www.youtube.com/watch?v=ntjkwIXWtrc) in which a hydraulic compactor crushes a plethora of creative and entertainment objects from typewriter to trumpet to dust in order to reveal how they're reborn as the new, razor thin iPad Pro. To say that it was poorly received was an understatement, with Apple raked over the coals on every social platform they cared to release it, with many customers saying they'd experienced a physical reaction of pure disgust to have seen these beautiful objects destroyed for the pleasure of Apple's Marketing team.

I'm endlessly skeptical that angry users on the internet are in fact as angry as they claim to be (rather than ragebaiting for engagement), but I won't lie, I did experience some schadenfreude from the rare event of Apple having to come face to face with the fact that what they're doing just isn't that good anymore. From the historically famous 1984 "Big Brother" commercial to the decades long run of heavy hitting Mac, iPod and, iPhone ads, Apple has been responsible for some of the best work in the history of discipline. But just like their products, what they've done in recent years has looked like a child's finger painting by comparison. Check out their cringe piece on [Apple 2030, starring Mother Nature](https://www.youtube.com/watch?v=QNv9PRDIhes) for more contemporary Apple state of the art.

What's most rhetorically offensive about this commercial isn't the objects being crushed, but what those objects symbolize, because they're telling of what the iPad Pro is pretending to be.

Some of what Apple chose to destroy checks out:

* TV.
* Arcade machine.
* Light.
* Clock with alarm.
* Vinyl records.
* Globe.

This category is consumption and small utilities, all of which the iPad Pro is pretty good at. It has a gorgeous screen for watching movies on the airplane, a touch interface is the perfect way to zoom and pan the world with Google Maps, and the clock app and flash light function work well. I'll even give them the typewriter. The iPad makes a great little writing machine because its narrow scope of utility and bad multitasking encourages focus.

Where I take issue is all the other stuff:

* Drum kit.
* Guitar.
* Metronome.
* Pro camera lenses.
* Clapperboard (that thing with the diagonal black stripes that they snap together on movie sets while yelling "scene X take Y" to start a scene).
* PC (running film editing software).
* Paint.
* Easel.
* Drawing mannequin.
* Drafting table.

The inclusion of these objects spells out clearly that Apple believes that the iPad Pro is a device for creating music, works of art, photos, films, and paintings.

It may technically be possible to do so, but no one would. Not because the iPad isn't some of the sleekest, slimmest hardware every conceived (it is), but because the software that runs on it is so determinedly hostile to people who are trying to, you know, _do_ things.

But don't take my word for it. Let's check in with some Apple fanatics so loyal to Apple that they'd take a bullet for Tim Cook, or even Jony Ives (despite the traiterous turncoat daring to part ways with the mothership!).

<img src="/photographs/nanoglyphs/042a-crush/crush2@2x.jpg" alt="Apple's Crush! ad" class="wide" loading="lazy">

### BEYOND the tablet (#beyond-the-tablet)

The sphere of Apple fandom is a strange place, and one where Steve Jobs' reality distortion field is still not only on, but operating at maximum power. In this dark mirror of our own reality, a common claim from Apple influencers is that they like iPads _so_ much that their iPad is their primary driver. Meaning, it's their main computer, and if they go on a vacation or work trip, it's the iPad that gets put in the carry bag rather than the laptop. Here's one dramatically titled example of this genre: ["Beyond the Tablet: Seven Years of iPad as My Main Computer"](https://www.macstories.net/stories/beyond-the-tablet/).

In the world of Apple enthusiasts, this is a flex. Translated to English it means, "I love Apple so much that I'm going to say that I do something so stupidly impractical that everybody knows that it's not really possible, and if it were possible it'd mean adding hours of time unnecessarily wasted time and frustration to my day. Please like and subscribe my YouTube channel."

The same person above who claimed their iPad is their main computer recently published a new, somewhat contradictory article: ["Not an iPad Pro Review: Why iPadOS Still Doesn’t Get the Basics Right"](https://www.macstories.net/stories/not-an-ipad-pro-review/), a long laundry list of its daily frustrations. Some choice section titles:

* "Files: A Slow, Unreliable File Manager"
* "Multitasking: A Fractured Mess"
* "Inefficiency by a Thousand Cuts"
* "The Need for Change"

Every complaint within is correct, but it misses the forest for the trees. It's not that iPadOS has a few blemishes or is missing a couple of useful features, it's that the entire stack is built from the ground up to reign in general purpose computing. Programs don't work well because they're built on dull, high-level abstractions that don't work well.

<img src="/photographs/nanoglyphs/042a-crush/crush3@2x.jpg" alt="Apple's Crush! ad" class="wide" loading="lazy">

### Cracked open by millimeter (#millimeter)

The fundamental difference between more traditional PC operating systems and iOS/iPadOS is that one started open, and the other closed. The flexibility of the former was near infinite: not only could you launch any program, but you could edit that the memory of running programs, or customize them by patching their raw binaries. By contrast, when the iPhone launched in 2007, you were allowed to open Safari.

The original OSes assumed they were the property of the user, and defaulted to open. iOS on the other hand was perfectly closed — a hermetically sealed black box. It took more than a decade for it to catch up with the most foundational features that Windows 3.0 shipped with in 1990:

* Third party apps.
* Copy/paste (two years later -- kind of amazing in retrospect).
* Multitasking
* Background processing.
* Windowing (as in, arranging multiple windows in a workspace).
* File system and file browser.

If it was only that these features took a long time to arrive it'd be one thing, but it was worse than that. Each was implemented, poorly. A copy/paste that between awkward tooltips, horrible text selection, and slow app change animations, makes a previously two second operation a 60 second one. Multitasking and file management so indescribably awful that no one on Earth would use them by choice (with the exception of the Jobs-philes mentioned above).

Here's what the article above has to say about multitasking, a concept that was originally mastered in the early 60s, but sent back four thousand years to the stone age for iOS:

> You don’t need to look far to see where iPadOS is failing in this regard. If you use Apple’s own Final Cut Pro for iPad – one of the company’s very showcases of the new iPad Pro – and begin exporting a video, then switch apps for even a second, the export is canceled. If you simply switch workspaces in Stage Manager or accidentally click on an incoming notification, an entire project’s export will fail.

Just imagine if multitasking had been that dysfunctional on 90s-era Windows. Personal computers as a general concept might have fizzled out and failed. Luckily for the iPad, it could afford to be a total failure, because when it was, people could still fall back to a real operating system.

A lot of us are still carrying around the internal intuition that the iPhone is "new" technology, and just needs a little more time to get there. But it's not new, and hasn't been for some time. The first iPhone shipped in 2007, which makes iOS 17 years old in 2024. For comparison, the first Macintosh came out in 1984, and Windows 1.0 in 1985. Let that sink in: **iOS is now older than Windows or macOS was when most of us started using them**. For me, it's more than twice as Windows was when I got ahold of it in the early 90s.

Just like every other iPad, the M4 iPad Pro is an an incredibly powerful processor in the sleekest possible shell, with a Fisher Price operating system.

And unless something fundamental changes at Apple, it's always going to be like that. The iPad will never be a general purpose computing platform.

I don't say this as an iPad hater. I've owned an iPad since they were first released, and my iPad Pro 2018 is fast, modern, and has great battery life six years into owning it. It's my preferred platform for reading comics and writing first drafts. I don't hate the iPad, I hate how off target it is for what it should've been.

---

That's issue 042 done, and once again, more words than I'd intended to write. Next time I'm going to try and take a page out of Apple's playbook and _crush!_ this thing down to size.

Until next week.