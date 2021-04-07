+++
image_alt = "Sand near Ocean Beach"
image_url = "/photographs/nanoglyphs/023-enhancement/sand@2x.jpg"
published_at = 2021-04-04T21:32:25Z
title = "Gradual Enhancement: New Language Features, Steadily"
+++

As the calendar flipped to the first day of April, so came about the internet's most schismatic non-holiday. The one time a year when people suddenly start thinking critically about what they read online -- on the lookout for the latest in a long tradition of [tech company pranks](https://www.microsoftcoffee.org/) -- before once again abandoning those newfound faculties one day later.

Major companies have pulled back on April 1st jokes in recent years as they started taking sharp criticism, but like many people, I still read the internet assuming everything is a joke until proven innocent, and yet this year, it wasn't enough. There was one that got me: everyone's favorite walks-like-a-duck-quacks-like-a-duck idiosyncratic language, Ruby.

Ladies and gentlemen, I give you downward variable assignment:

``` c#
puts("Hello" + "World")  #=> HelloWorld
     ^^^^^^^x  ^^^^^^^y

p x  #=> "Hello"
p y  #=> "World"
```

As outlined in the snippet, downward assignment augments the typical "leftward" assignment found in most programming languages (`x = "Hello";`) by allowing syntax to send a value to set on a variable _downwards_. In this sample, we set `"Hello"` to `x` and `"World"` to `y`.

The prank was dressed up as a [formal feature proposal](https://bugs.ruby-lang.org/issues/17768) on the Ruby bug tracker, and triggered such a visceral reaction in me that it had me ensnared for the full minute it took to read through.

The reason downward assignment works so well is that even though it's completely ridiculous, when it comes to Ruby in particular, it's a plausible addition. Ruby's long practiced a mantra of one-upmanship over Python's "batteries included" to make it something more akin to "_everything_ included." New versions come with a host of improvements, which is great, but also ship with many additions with more dubious usefulness, seemingly present as change for change's sake.

For example, downward variable assignment is a direct play on Ruby 3's very real "rightward" assignment, which enables syntax like this:

``` c#
"Hello" => x
```

Undoubtedly, rightward syntax is useful to _someone_ out there, but is very Ruby-esque in that although it introduces some marginal improvement, it ever-so-slightly increases cognitive overhead for every Ruby developer, while also making every Ruby interpreter that much more complicated. Many language designers would have looked at rightward assignment, acknowledged it as "pretty neat", then firmly rejected it as nowhere near compelling enough of a cost/benefit equation.

And that's how they got me. Downward assignment is too outlandish to be a real feature, but not by much. It's right in that April Fools sweet spot of _just_ compelling enough to be believable in those early moments, then absurd as your rational brain finally catches up. My hat's off to you, Yusuke.

---

## Slow and steady (#slow-and-steady)

This is an excellent article on [new features in Java from version 8 to 16](https://advancedweb.hu/a-categorized-list-of-all-java-and-jvm-features-since-jdk-8-to-16/). It's written in bullet point form, broken down into categories like "feature", "new API", or "performance improvement", and each change is tagged with the specific JDK version where it was introduced.

A few weeks ago I wrote about how writing Java these days is actually, against all odds ... [pretty good](/nanoglyphs/021-ides), a position I never would have imagined I'd ever take when I first learned the language back in college.

Java is the anti-Ruby. It's literally designed [by committee](https://jcp.org/en/participation/committee), and improvements are not only introduced at glacial pace, but with extreme prejudice. New language syntax comes in only when it's absolutely, indisputable clear to everyone involved that it'll be widely useful. Lambdas (`x -> x + 1`) entered the language in March 2014 (Java 8), which is roughly 18 years after it was apparent they were dearly needed (~Java 1.0, released January 1996).

But that slow and steady approach has its advantages. Perusing the new features list, there isn't a single one amongst ~a hundred of questionable value. They're all useful, and keeping the language relatively minimal means less opportunity to make a mistake and less complexity to bear forward in future maintenance. And while they took a long time to get there, a lot of the most important quality of life improvements for Java programmers finally exist (e.g. lambdas, switches, the stream API, tuples/records), making today's Java quite pleasant to use.

---

## Boilerplate flexibility (#boilerplate)

But although Java's far more usable today than it was before, there are definitely still some rough areas that leave something to be desired. Take for example, its infamous getters/setters:

``` java
public class Person {
    private String firstName;

    public String getFirstName() {
        return this.firstname;
    }

    public void setFirstName(String firstName) {
        this.firstName = firstName;
    }
}
```

This was a technique that was really brought en vogue by Java, and there's quite a good idea at its core. Exposing variables directly through an API is easy, but doing so forever decreases the flexibility you have to change said API while keeping it backwards compatible. Instead, introduce indirection by wrapping variable access with getter/setter methods that can be reimplemented as need be, while having the nice property of maintaining compatibility for existing users.

The problem with getters/setters is boilerplate. Most of these methods end up as simple one liners, and some classes need dozens of them. They're not only slow to type, but introduce a lot of visual noise that does little to improve anyone's life.

I'm not the only one with this opinion, and over the years a myriad of code generation tools have appeared to make the syntax more succinct like [AutoValue](https://github.com/google/auto/tree/master/value), [Immutables](https://immutables.github.io/immutable.html), and [Lombok](https://projectlombok.org/). Once your project is up and running, working with any one of these is fine, but they tend to produce obtuse compilation errors if you make a mistake, are yet another box to tick when starting a new project, and have fractured the ecosystem so that Java code looks a little different depending on where you look. Again, mostly fine, but a pretty clear sign of missing functionality in core.

## The history of properties (#properties)

Sharing a design ethos, VM-based runtime model, and a lot of syntax, C# and Java are sister languages. Examining any particular code snippet, it was often hard to tell which language you were looking at, especially in earlier versions before they started to diverge.

But largely guided by a single company, C# tends to move much more quickly than the language that inspired it. It also had the benefit of hindsight, and took advantage of that to avoid some of the pitfalls found in Java right from the beginning. Java's verbose getters and setters in particular were in its sights, and right from version 1.0 C# shipped with _properties_, a less verbose and more standardized way of doing the same thing.

C#'s language designers didn't stop with just the basics. Almost every version since has introduced small enhancements to properties that make them a little more powerful, a little tighter, or a little more ergonomic. C# has found the perfect compromise in development pace on the wide spectrum between Ruby and Java -- slow, gradual enhancement that makes good things even better over time.

### C# 1.0: The beginning (#c-sharp-1)

C# 1.0 brings properties to life:

```c#
public class Person
{
    private string _firstName;
    private string _lastName;

    public string FirstName {
        get { return _firstName; }
        set { _firstName = value; }
    }

    public string LastName {
        get { return _lastName; }
        set { _lastName = value; }
    }
}
```

Pretty straightforward `get`/`set` keywords that wrap instance variables. Programmers use this syntax to access a property:

``` c#
person.FirstName = "Jin";
```

Again, very similar in spirit to getters/setters, but with terse syntax, and standardized across every project that uses C#. No third party libraries necessary.

### C# 3.0: Auto-properties and object initializers (#c-sharp-3)

The dawn of _auto-properties_. Getter/setter implementations and instance variables may be left out completely for properties that just need to get and set a value with no complications:

``` c#
public class Person
{
    public string FirstName { get; set; }
    public string LastName { get; set; }
}
```

Because most getters/setters are simple pass throughs, auto-properties turn out to be inordinately useful. Flexibility around API compatibility is maintained because if needed, an auto-property's `get`/`set` can always be unrolled into full implementations in a future version.

C# 3.0 also brought in object initializer syntax:

``` c#
Person person = new Person{
    FirstName = "Jin",
    LastName = "Sakai"
};
```

This is a big improvement because it obviates the necessity to have constructor overloads for every possible set of parameters. Make required properties constructor arguments, and leave optional ones as properties:

``` c#
Person person = Person("Jin", "Sakai") {
    CountryOfOrigin = "Japan",
};
```

### C# 4.0: Optional parameters (#optional-parameters)

Not strictly property-related, but C# 4.0 brings in optional method parameters:

``` c#
public class Person
{
    public Person(string firstName, string lastName,
        string countryOfOrigin = "Japan") {

        if (!isCountry(countryOfOrigin)) {
            throw new ArgumentException(...);
        }

        _firstName = firstName;
        _lastName = lastName;
        _countryOfOrigin = countryOfOrigin;
    }

    ...
}
```

This is another good way of making an easy distinction between required and optional properties that need to be set on a class, and has the advantage over object initializer syntax in that it's an easy way to set defaults.

Java still has no equivalent, and its absence has been the main factor that's led to the explosion of builders for practically every non-trivial class.

### C# 6.0: Defaults and expressions (#c-sharp-6)

C# 6.0 shifted the ergonomics for optional fields back to properties by letting them be easily annotated with default values:

``` c#
public class Person
{
    public string CountryOfOrigin { get; set; } = "Japan";
}
```

It also introduced expression body definitions, a more succinct way of implementing one liners:

```c#
public class Person
{
    public string FullName => $"{FirstName} {LastName}";
}
```

### C# 7.0: Expanded expression bodies (#c-sharp-7)

C# expanded expression-bodied properties by letting both the getter and the setter to take an expression:

``` c#
public class Person
{
    private string _firstName;
    private string _lastName;

   public string FirstName
   {
      get => _firstName;
      set => _firstName = value;
   }

   public string LastName
   {
      get => _lastName;
      set => _lastName = value;
   }
}
```

A small improvement, but one that makes one-liners more readable.

### C# 9.0: Easy immutables (#c-sharp-9)

C# 9.0 introduced the `init` keyword as an alternative to `set`:

``` c#
public class Person
{
    public string FirstName { get; init; }
    public string LastName { get; init; }
}
```

Properties with `init` can be set when an object is first initialized, but not after, making it a convenient way to write easy immutable classes.

C# 9.0 released late last year so that's the end of the story so far, but we'll see what new refinements future versions hold.

---

## Gradual enhancement (#gradual-enhancement)

C# excels in gradual enhancement -- moving slowly enough to make sure it's bringing in the right things, but not _so_ slowly as to be punitive to users or exacerbate bad community patterns emerging from features lacking in core. None of the improvements we saw above to properties was strictly needed, but every one makes the language a little more ergonomic, and a little more pleasant to use. Bringing them in gradually reduces the likelihood of mistakes being made in an overly broad upfront design, and allows the team to be reactive to how users are making use of the language in the real world.

There's a [very readable change log of major language additions](https://docs.microsoft.com/en-us/dotnet/csharp/whats-new/csharp-version-history). It's a great historical document, makes it really easy to tell exactly which version brought in which features, and is a testament to how C# is always on the move.

---

{{NanoglyphSignup .InEmail}}

<img src="/photographs/nanoglyphs/023-enhancement/city-pop@2x.jpg" alt="Japanese city pop" class="wide">

## "I miss the future" (#japanese-city-pop)

Now for something completely different -- I'm not an avid reader when it comes to music editorials, but I've been obsessing over this article from Pitchfork: [_The Endless Life Cycle of Japanese City Pop_](https://pitchfork.com/features/article/the-endless-life-cycle-of-japanese-city-pop/). In an era of clickbait-driven neo-journalism, this is a refreshing dose of good prose on a novel subject that's nothing short of fascinating.

City pop is roughly put, a loosely defined genre of music inspired by flourishing Japan of the 70s and 80s, which was in turn inspired by the golden age of the West. The name "city pop" was retconed much later, but invokes a time of optimism and prosperity, with the wave of technology on the rise, and the newfound luxuries of cosmopolitan life:

> The music is often exuberant and glitzy, drawing inspiration from American styles like funk, yacht rock, boogie, and lounge music. Emulating the easy vibes of California, the music’s sense of escapism is often embodied by the sun-soaked cover art of Hiroshi Nagai, one of city pop’s iconic designers: Sparkling blue water, slick cars, and pastel buildings evoke fantasies of a weekend vacation at sea.

But nowadays, in the shadow of Japan's lost decades, and with the sheen of a utopic future worn thin, city pop's symbolism is more complex, maybe best described as a "retro-futurist melancholy". Nostalgia is a core theme:

> Nostalgia, as the theorist and media artist Svetlana Boym once wrote, involves “a superimposition of two images—of home and abroad, of past and present, of dream and everyday life.” The emotional response to city pop centers on these twin imaginations: of Japan and the United States, the ’80s and now, the prior promises of capitalism and its current reality. Online, listeners dwell on artificial memories of boom-era Tokyo but also idyllic childhoods watching cartoons, reaffirming Boym’s claim that nostalgia “appears to be a longing for a place but is actually a yearning for a different time.”

And on the promises of a future never quite fulfilled:

> Boom-era Japan, with its neon metropolises and abundant consumer freedoms, embodies a lost promise of capitalist utopia that was crushed in the ’90s by the country’s recession. By savoring its music, listeners can both indulge in and mourn the beautiful, naive optimism that seemingly defined the time—as well as its bracing visions of what would lie ahead. As one commenter on a YouTube city pop mix wrote, echoed by many others, “I miss the future.”

Here's a [YouTube city pop video mix](https://www.youtube.com/watch?v=qXC4AyjRikg&t=3246s) showcased by the article. The whole thing is good, but I've taken the liberty of deep linking directly to minute 54, which is a song titled _Echoplex (w/ Tendencies)_ by FIBRE, and one that I've listened to about fifty times in the last two weeks.

I quit my job yesterday. More on that soon.

Until next week(-ish).
