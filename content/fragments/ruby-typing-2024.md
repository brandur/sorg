+++
hook = "Diving into the RBS ecosystem as an alterative to Sorbet for Ruby typing."
published_at = 2024-05-03T08:25:57+02:00
title = "Ruby typing 2024: RBS, Steep, RBS Collections, subjective feelings"
+++

I was writing a new Ruby gem recently, and being a strong proponent of a type checking step, I wanted to do right by the ecosystem so that anyone using it would get the full benefit of type checking against my gem's API in their own projects, so I dug into the current state of the art to find out how that'd be done.

I used Sorbet for years at Stripe, and although a little unwieldy in places, it was overall quite practical, and certainly useful for substantial bug reductions. About four years ago, Matz declared unilaterally that everything about Sorbet's approach to typing was wrong, and established a similar but entirely divergent technology involving [RBS files](https://github.com/ruby/rbs/blob/master/docs/syntax.md), companions to the traditional `.rb` containing type annotations. I'd never tried them before, but given they seem to be the preferred direction of the ecosystem, they were my point of entry.

RBS files look like Ruby, but are subtly different (notice the extra colons and syntax to express return values), and with types. For example:

``` rb
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
```

They're a pretty comprehensive type system, supporting common paradigms like interfaces, but also union types, intersection types, record types, tuples, and type parameters.

## The CLI and prototyping (#cli-prototyping)

RBS [ships a CLI](https://github.com/ruby/rbs) that can "prototype" an existing `.rb` file:

``` sh
rbs prototype rb lib/driver.rb > sig/driver.rbs
```

A scaffold is generated for the original file with its structure like modules, classes, and method definitions. Convenient, although once you look inside, you notice that almost everything is `untyped`, so the `.rbs` that was just generated has little marginal benefit for your project's typing compared to if it didn't exist at all.

That's fixable, so you correct type signatures for everything, and then maybe even start using some of the more novel constructs that RBS supports like [signatures on blocks or type parameters](https://github.com/ruby/rbs/blob/master/docs/syntax.md).

## Ergonomics, two by two (#ergonomics)

Once you do, your `.rbs` file has now taken on a life of its own, and needs to be maintained by hand rather than generated again from its source `.rb`. Having started programming in C, it reminded me a lot of `.h` header files, and I tried to have an open mind about them, but after a few days of work found the RBS files quite grating (post-C languages ditched the headers for a reason):

* Every addition or change is done twice. Once in `.rb`, and then repeated in `.rbs` with slightly altered syntax and type annotations added. Keep in mind that all private/internal methods and constants should also be in `.rbs` to maximize the benefit of the type check, so even internal refactors tend to update both types of files.

* Needing to have two files open for every file in your project means 2x the tabs in your editor. Not the end of the world, but adds a lot of organizational bloat.

## Type checking with Steep (#steep)

With RBS files in place, you're now ready to reap the fruits of your labor by detecting code problems via static analysis through the type system. [Steep](https://github.com/soutaro/steep) is the alternative to the Sorbet CLI that's been informally blessed by the RBS initiative.

You'll initialize your project with a `Steepfile` that points to your source and signature directories, then run `steep check` to run an analysis that hopefully turns up no errors:

``` shell
$ bundle exec steep check
# Type checking files:

..................................................................................................

No type error detected. 🍵
```

In case of a typing problem, Steep readily detects it:

``` shell
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

Steep worked quite well for my tiny little project, but given that it's written in pure Ruby, I wondered what would happen if you tried to run it for something like Stripe or Shopify's monoliths, which must be tracking hundreds of thousands of Ruby files by this point.

Best I could tell, Steep's support for DSLs is minimal. I'd made the apparent mistake of writing my test suite with RSpec, so I couldn't get any useful type checking for half the project.

### Magic comments only (#magic-comments)

One of Matz's main concerns about Sorbet's syntax was that it was too invasive in Ruby code. Worth nothing however, that a type system in a dynamic language always ends up needing escape hatches. With no runtime syntax available, Steep has chosen to implement things like type assertions as magic comments:

``` ruby
args_with_insert_opts = args #: _JobArgsWithInsertOpts
```

Is that really that much better than just having in-code type syntax? Dear reader, I'll leave it to you to decide.

## RBS Collections (#rbs-collections)

After a Steep user has written extensive RBS definitions for all their code, they're left with a dilemma, which is how to get type signatures for third party code. In the Sorbet ecosystem, this is a task generally handled by [Shopify's Tapioca](https://github.com/Shopify/tapioca).

RBS suggests another concept called an "RBS Collection". `rbs collection install` reads your project's `Gemfile.lock`, tries to resolve RBS definitions, and places the result in `.gem_rbs_collection`:

``` shell
# Create rbs_collection.yaml
$ rbs collection init

# Resolve dependencies and install RBS files from this repository
$ rbs collection install
```

Most gems don't ship RBS files of course, so [a central repository](https://github.com/ruby/gem_rbs_collection/tree/main/gems) maintains RBS files for a few dozen common gems that the command falls back to using. They're all generated RBS with everything marked `untyped`, so don't expect them to catch too many bugs.

I was frankly surprised by the existence of RBS Collections because I'd intuitively assumed that the core-blessed typing path would be able to live closer to already existing mainline tooling. So signatures could be installed as part of `bundle install` or something close to.

## IDE tooling (#ide-tooling)

I installed the Steep VSCode plugin. It produces error diagnostics in case of typing problems, but as far as I can tell, nothing more sophisticated than that like jump-to-definition or symbol renaming. [1]

## Who's using this? (#whos-using-this)

As I iterated through each new piece of the RBS typing toolchain, I couldn't shake the feeling that although I'd got up and running successfully with my brand new project of ~200 LOCs, this would be hard to get installed successfully into an existing codebase. There's many moving parts, tools like Steep are surprisingly strict in that they're unhappy if a project's types are only partially defined, and the overall ergonomics of the toolchain aren't exactly optimal.

I tried looking around for evidence of happy RBS users, but almost everything written about the system are the usual "intro to RBS" posts, most of them from years ago ([Honeybadger's is probably the best](https://www.honeybadger.io/blog/ruby-rbs-type-annotation/) in this category).

A defense of RBS might be that it's all quite new and will take some time to properly crystallize, which is reasonable, but it has been four years. and it's also reasonable to expect that these tools would've become a little more sticky after that much time in the field. If they're going to stick that is.

I'm a little afraid that RBS might fall victim to the same problem as [Ractors](/nanoglyphs/018-ractors), which were an ambitious idea, but turned out to be so fundamentally incompatible with the existing ecosystem that even years after their release, I've never heard of a production stack that's been able to make use of them. I know it's wishful thinking at this point, but I can't help but wonder whether if Ruby core had combined efforts with Sorbet's existing bulk of tooling, there might be a compromise by now that traded a little idealism for a little more practicality.

A good thing is that the RBS toolchain is certainly lighter weight than Sorbet's. I expect to keep using it for smaller projects where I'd like some internal type safety (poor RSpec support is indeed a conundrum), and maximally compatible public-facing type APIs for projects that use them.

[1]  Every time I come back to Ruby from Go, I find myself yelling in my head, "how do people live like this???".