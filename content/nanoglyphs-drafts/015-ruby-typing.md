+++
image_alt = "Cole Valley around noon on Bladerunner day"
image_url = "/assets/images/nanoglyphs/015-ruby-typing/cole-valley@2x.jpg"
published_at = 2020-10-07T22:23:01Z
title = "Bladerunner Day, Ruby Type Signatures, Typing at Stripe"
+++

Over the years, I've learnt to take stories about San Francisco with a grain of salt. It's an old city with a deep history and an outsized cultural presence, and something about that, along with a citizenry with a flair for the dramatic, produces stories exaggerated to an extent that they're flirting dangerously with fiction. A coyote spotting in a city park will occasionally spark so many flabbergasted posts on social media and Reddit threads expressing pure incredulity that you'd think we were facing the imminent return of the [Beast of Gévaudan](https://en.wikipedia.org/wiki/Beast_of_G%C3%A9vaudan).

But with that said, there is the occasional story that's true, and borders on the unbelievable without the helping hand of hyperbole. You've probably already heard about San Francisco's [Bladerunner 2049 day](https://www.youtube.com/watch?v=h9ZUtFQZbq4) a few weeks ago as it was startling enough to make international news. I come from a city where it's normal to have a few days a year of -35C temperatures, and which will once in a while see storms that produce car-and-roof-wrecking golf ball sized hail, but in terms of extraordinary atmospheric phenomena, I've never seen anything like it, and don't expect to again.

I was lucky enough to be out for a long walk out to Ocean Beach that morning. We started around 7:30, and it was a little surprising that it was still dark. We were still coming down from long summer days, with great light even in the morning's earliest hours. I chalked it up to the lateness of the season. It's fall now Brandur -- _get used to it_ -- winter's coming and all of that.

But as the hours passed, that explanation rapidly lost credibility. Walking through Golden Gate park at 9 o'clock -- still dark. The beach around 10 o'clock -- still dark. Returning along JFK towards Haight -- _still dark_. I distinctly remember walking under some medium-heavy forest canopy around 11 o'clock and wishing I had a flashlight -- I couldn't see the ground or where I was putting my feet. Fear of the unknown crept into my pagan heart. We'd offended a great sun god -- Apollo, or Ra, or Sól, or Helios -- and they were inflicting punishment by taking the day off.

The Bladerunner-esque orange hue was jaw droppingly bizarre, but what I'll remember about the day was the darkness. Cameras these days do such a good job of compensating for low light that most photos you see online don't do it justice, including mine above. (Of course it had to be the one day I decided not to bring a "real" camera ...)

---

(And my usual repartee: I'm Brandur. This is _Nanoglyph_, my newsletter about software, dystopian orange skies, and type signatures. My discipline on sending it has been so rough lately that you might have signed up long ago and forgotten about it, but if you're sure you didn't, or just never want to see it again, [unsubscribe right here](%unsubscribe_url%).)

---

## Better typing in Ruby (#ruby-typing)

A few months ago, the Ruby core team announced their [plans for typing in Ruby 3](https://developer.squareup.com/blog/the-state-of-ruby-3-typing/), a long awaited feature which, if done right, will have profound benefits to the language's safety and productivity, especially where large codebases are concerned.

The history of typing in programming languages _generally_ is interesting. The idea's been around forever, with some of the earliest languages like C having relatively good type systems, but over the next few decades there'd be wild divergence in thinking as language creators experimented with different approaches. Some went for stronger typing, giving us C++, D, Java, and on the furthest reach of the spectrum where types became a religion, Haskell. But simultaneously, there was an equal push for weaker typing, with the appearance of Perl, Python, JavaScript, PHP, Ruby, and the like.

Since then, there's been movement towards typing convergence that's a compromise between the two extremes, but errs on the side of more typing. Python, JS, and PHP have all picked up support for type signatures, either in the language itself, or in popular variants like TypeScript. Conversely, strongly typed languages stayed strongly and statically typed, but in a nod to user ergonomics, have walked back somewhat by allowing type inference within the bounds of a single function -- see the addition of the `auto` keyword in C++ and Java, `var` in C#, or newer languages which have baked function-local inference in from the beginning like `:=` in Go, or `let` Rust.

``` go
//
// Type declared explicitly in code
// (traditionally required by most old world languages)
//
var myInt int
myInt = 3

//
// Type inferred by compiler for brevity
// (supported at least within functions by most modern languages)
//
myInt := 3
```

Ruby's been a laggard. Technically speaking, Ruby is already typed -- any particular variable has a type, whether it's an integer, string, or class instance, and it's an error to call methods on it which it doesn't support (Rubyists will have seen the infamous `NoMethodError` a hundred times by the end of their first day on the job). But it's a _dynamically_ typed language because the interpreter has no idea what anything is until it starts executing code. When we're talking about typing in Ruby 3, we're not so much adding typing as _type signatures_, but in practice, those can be as important as the language's core typing model.

Type signatures have a few benefits, with an important one being that by defining types in the source code itself, we can gain a lot of insight about the correctness of code without running it.

Compiled languages get this for free as the compiler does the job -- getting the typing right is a prerequisite to a working program. In dynamic languages, it's done by a static analysis tool that runs independently of the interpreter. Python's ecosystem recommends the use of [mypy](http://mypy-lang.org) for analyzing its type signatures, but they're very careful to point out that it's an optional extension -- Python users never have to go beyond the interpreter if that's their preference.

Matz has been talking about types in Ruby since at least as far back as 2016, but until very recently, precious little progress had been made. I think it's quite possible that serious movement would _never_ have happened, but their hand was forced as developments beyond their control threatened to make decisions for them.

---

## Typing at Stripe (#typing-at-stripe)

_(I should note before I start this section that I work at Stripe where Sorbet was developed, but am not on the Sorbet team, have never been party to any discussions between Sorbet and Ruby Core, and have no relation to the project except as a regular user.)_

Rewind to me joining Stripe in 2015. I was used to working in fairly large Ruby codebases, but was floored by the sheer number of LOCs Stripe had produced. A few vocal early engineers had been staunchly opposed to microservices, so most of the code doing anything with domain logic had landed in one giant Ruby codebase. Ruby does a poor job of encouraging modularity, so the code had ended up as one giant amorphous blob, with everything calling into everything else, and all boundaries purely theoretical in nature.

Working in it was the stuff of nightmares. The company's greenfield days were long over, so most of the time engineers were modifying existing code rather than writing it anew. Examining any section of it, all you had to go by in figuring out the type of anything was naming. `num_*` was probably an integer. `is_*` -- hopefully a boolean. `config` was probably some kind of configuration object, or maybe a hash. If no type was readily apparent, the best thing you could do is throw a Pry statement into the path, try to find a test case that exercised it, and inspect variables at runtime. This was an excruciatingly slow process (Ruby's interpreted, so more code means more startup overhead), and not a reliable one. The original author might've intended for a variable to be one type, but it may have subsequently picked up new uses and had its interface broadened to include many possible types. Even if you'd observed an incoming variable as a `Charge` object at some point, it could be a `Payment` when called from somewhere else.

The uncertainty didn't just make code hard to change -- it caused real production problems on a regular basis. Developers would make an incorrect assumption while changing something, have enough success in the test suite that it reported no failures, only to find 500s thrown in production on an untested path (we write a lot of tests so _in theory_ there shouldn't be too many of those, but actually, there's a lot). It was a mess, and it made even minor changes difficult and risky. The risk was so great that small, incremental changes were all that were even possible -- as more lines were changed on a deploy, the likelihood of a mistake trended exponentially towards 100%.

---

## Sorbet (#sorbet)

Sorbet started life as a runtime type checking library. Type signatures which were syntactically valid Ruby were attached to methods and variables, and at runtime would bind themselves into the invocation chain and throw a type error if the wrong type was passed in:

``` ruby
sig {params(x: Integer, y: Integer).returns(Integer)}
def add(x, y)
  x + y
end
```

And a corresponding failure:

```
TypeError: Parameter 'x': Expected type Integer,
    got type String with value "a"

Caller: test_module.rb:24
Definition: test_module.rb:18
    ...
```

A method like `add` above is a good, simplified demonstration of a method in Ruby that could go bad. Although it may have originally been intended for use with integers, at some point someone notices that it works with strings as well, and starts passing those in too. Ruby's duck typing is fine with this (it's even considered a feature) because the `+` operator is present on both strings and integers. The problem arises when later, someone tries to change something in `add` assuming its specific to integers like its name would suggest, only to either realize that they've inherited a big refactoring job they never signed up for (if there was enough test coverage to catch the string inputs), or accidentally blow up production (if there wasn't).

Sorbet's runtime checking system also came with an annotation to indicate that check failures should notify someone instead of throwing an exception. This helped us gradually roll out typing to modules that never had it before. Engineers would deploy a signature to production, check that it was error free, then lock it down. Test coverage was good, but even knowing line of code was called from the test suite wasn't enough because it's not possible to know whether it's been called with every possible type that could be sent to it. The only way to roll out type checking was to do so very, very carefully.

### Static analysis (#static-analysis)

Sorbet's next evolutionary step was the introduction of a mypy-like static analysis engine. It parses code without running it, and uses its type information to calculate whether the call graph of variables and method invocations is sound.

This was a major improvement in several ways:

* Static analysis was faster to run than a single suite (i.e. one file) of tests, and often faster than even a single test case (our tests have a significant startup overhead). This has become less true over time as the amount of code continued to grow, but it's still fast given the quantity of code it has to analyze.

* It produced better errors than a runtime failure. A specific variable or invocation gets highlighted, and the user is told exactly what's wrong with it.

```
test_module.rb:24: Expected Integer but found String("a") for argument x http://go/e/7002
    24 |      add("a", "b")
                  ^^^
    test_module.rb:17: Method TestModule#add has specified x as Integer
    17 |    sig {params(x: Integer, y: Integer).returns(Integer)}
                        ^
  Got String("a") originating from:
    test_module.rb:24:
    24 |      add("a", "b")
                  ^^^
```

* Code is analyzed whether it was tested or not, so we have an extra layer of protection even where there are holes in the test coverage.

Ignoring any benefit to the wider Ruby community, Sorbet has been a resounding success internally. In the beginning, it was just type enthusiasts who were adding type signatures to files. Now, having recognized the value they provide, it's everyone. Many (if not most) new files even specify the `typed: strict` pragma, which _requires_ that every method and instance variable in it has a type specified. Types have been back ported to the majority of the pre-Sorbet Ruby files, and the benefit of that is compounding -- the more code that's typed, the more problems Sorbet can detect.

---

## Introducing `.rbs` (#rbs)

That brings me back to typing in Ruby 3. Rather than adopting Sorbet's type signature design, the core team forged ahead with their own. Fair enough, Sorbet's is verbose, and at least a little ugly. But Ruby 3's bespoke type design comes with a galaxy-sized deviation from other such systems -- they've refused to allow type information in the source `*.rb` files themselves. Instead, developers specify type signatures in separate `*.rbs` files that mirror the declarations of a companion `*.rb` file.

For an `*.rb` file like this:

```
# sig/merchant.rb

class Merchant
  attr_reader :token
  attr_reader :name
  attr_reader :employees
  
  def initialize(token, name)
    ...
  end
  
  def each_employee(&block)
    ...
  end
end
```

You'd have this corresponding `*.rbs`:

```
# sig/merchant.rbs

class Merchant
  attr_reader token: String
  attr_reader name: String
  attr_reader employees: Array[Employee]

  def initialize: (token: String, name: String) -> void

  def each_employee: () { (Employee) -> void } -> void
                   | () -> Enumerator[Employee, void]
end
```

You can probably tell by now that I think this is a mistake of fairly colossal proportion. Although theoretically the static analysis of typing will be the same, it's an unspeakably large compromise in ergonomics. Even the simple act of renaming a method now involves changes across multiple files.

The announcement blog post also makes the mistake of comparing `*.rbs` files to TypeScript's `*.d.ts`. Although there's superficial similarity purely at the cosmetic level, it misses the raison d'être of `*.d.ts`. In TypeScript, these files are used to add types to untyped JavaScript files that you _don't_ control -- say one that's part of a package imported from NPM, or if you need files to be interoperable between JavaScript and TypeScript -- say if you're publishing a package to NPM. We do this [in stripe-node](https://github.com/stripe/stripe-node/tree/master/types) for example, so that both JavaScript and TypeScript users can use the package, but TypeScript users still get the benefit of type information. The critically important difference is that in TypeScript you still have the _option_ of putting type information inline with TypeScript code, and that's vastly preferred over a `*.d.ts` file when possible.

### Man *and* machine (#man-and-machine)

And while static analysis is great, we shouldn't forget that type signatures are _for people too_. Being able to see what the expected types of any particular variable or method while reading code is a huge boon for comprehension. Sure, a great IDE can help with this too, but why not both? It's free.

There isn't much chance that Ruby Core backpedals at this point, but as someone who has grown to like the language's syntax, it's disappointing to see it fall yet another step behind its sister language, Python. Along with better performance, _much_ better documentation, a [concurrency model](https://docs.python.org/3/library/asyncio.html), and an ever growing popularity disparity in Python's favor, Python can now definitively boast the better type system, despite Ruby having had [more than five years](https://www.python.org/dev/peps/pep-0484/) longer to think about its design and implementation. A decade ago Ruby and Python were neck and neck. Today, there's no comparison.

Still, more typing is usually better, and after decades of paralysis, it's good that Ruby's moving forward once again.

---

## Mavericks to Mojave (#mavericks-to-mojave)

![Leopard default background](/assets/images/nanoglyphs/015-ruby-typing/leopard@2x.jpg)

Outside of software, I really enjoyed this [MacOS wallpaper gallery](https://512pixels.net/projects/default-mac-wallpapers-in-5k/) which has every default system background since the original OS X 10.0 Cheetah, and all in vivid 5k resolution that makes even the most ancient among them look great on today's most modern monitors.

![Mavericks default background](/assets/images/nanoglyphs/015-ruby-typing/mavericks@2x.jpg)

The Mac's wallpapers are so artfully designed that they're the only product sequence I can think of [1] where every instance has wowed me beyond belief and yet the next is still _somehow better_; a feeling that's been consistent now for a dozen iterations. From abstract art, to galaxies, to California landscapes (and seascapes), every step along the way has shown transcendent taste in creative direction.

Thanks for reading and happy Canadian Thanksgiving! Until next week.

[1] With the possible exception of the iPhone lineup.
