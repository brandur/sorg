+++
image_alt = "The view off Sulphur Mountain"
image_url = "/photographs/nanoglyphs/018-ractors/sulphur-view@2x.jpg"
published_at = 2021-01-15T13:28:24Z
title = "Ruby 3's Ractors"
+++

On December 25th 2020, Christmas Day in much of the world, Ruby 3 was released. The timing might seem unusual, but it's part of an ongoing tradition of annual Ruby releases as a kind of high tech Xmas gift to the world. The team's hit their December 25th release date every year since Ruby 2, released in 2013, making their current shipping streak an impressive eight in a row.

In [015](/nanoglyphs/015-ruby-typing#ruby-typing) I wrote about the upcoming type annotation system in Ruby 3. In short, it's great that Ruby finally has one, but between the annotations being in a separate `.rbs` file and no prescribed static checking tool, Ruby 3's is roughly the software equivalent of homeopathy, and will push standardization on a common language toolchain years into the future. Still, types (and type annotations) are good, and it's progress nonetheless.

Typing aside, Ruby 3 came with some other interesting additions. Most notably is that of [Ractors](https://github.com/ruby/ruby/blob/master/doc/ractor.md) ("Ruby actors"), a new concurrency feature in the language, and what could fairly be called its first _true_ abstraction for parallelism [1]. They're still in their infancy and not yet broadly usable, but could be a major turning point in Ruby's trajectory.

---

Brief intermission: Welcome to _Nanoglyph_, a weekly newsletter about parallel computing and spruce trees. If you're reading this on the web, and you wouldn't mind seeing more posts like this one, you can [subscribe here](https://nanoglyph-signup.brandur.org).

---

## GVL ex GIL (#gvl)

_(If you're a well-versed Rubyist you might want to skip this section as you've probably been hearing about the GIL for the last decade or two.)_

Let's look at a simple Ruby program to calculate a Fibonacci Sequence:

``` ruby
def fib(n)
  new, old = 1, 0
  n.times { new, old = new + old, new }
  old
end

3000.times.each do |i|
  fib(i)
end
```

And then the same thing again, but in this case to divide the workload across two threads:

``` ruby
t1 = Thread.new do
  3000.times.each do |i|
    fib(i) if i % 2 == 0
  end
end
t2 = Thread.new do
  3000.times.each do |i|
    fib(i) if i % 2 == 1
  end
end
t1.join
t2.join
```

Now we run them. This is a job that should parallelize nicely, so given any modern computer that has more parallel cores than it knows what to do with, program #2 should be roughly twice as fast right?

Non-threaded (program #1):

``` sh
$ time ruby main.rb

real    1.00s
user    0.94s
sys     0.04s
```

Threaded (program #2):

``` ruby
$ time ruby main.rb

real    1.06s
user    1.00s
sys     0.04s
```

Threading isn't only _not_ faster, but it even clocks in slower than the single-threaded version in this example. We could even make this a 4, 8, or 16-thread variant, and none would do any better. What could possibly be happening here?

The answer: despite having the usual constructs like threads that would make you think otherwise, Ruby may be a concurrent language, but it's not a parallel one. A Global Interpreter Lock (GIL) ensures that Ruby in only place is running at any given time. The case above performs badly because despite threads, every operation is running sequentially.

But that's not to say that threading isn't useful in Ruby programs. Threads can never run in parallel while a program is executing Ruby code, but they _can_ preempt each other when waiting on I/O (e.g. reading a file, writing to a socket), and in practice, a lot of real world applications are I/O bound. Look at a typical web app, and it's spending the lion's share of its time waiting on database calls or sending/receiving other data over the wire, all of which is time where another thread could be doing useful work. My Fibonacci example above purposes demonstrates the most degenerate case in which a program is entirely Ruby-bound, but most programs will do better.

But I've been indulging in archaic terminology. Nowadays, the GIL is called the "GVL" (Global VM Lock) because it's no longer gated around the entire interpreter, but rather just around execution of bytecode in Ruby's VM. The GVL was an improvement, but it was a little like [StatOil](https://en.wikipedia.org/wiki/Equinor) rebranding itself to "Equinor" -- a little different, even if mostly the same thing, but a good opportunity to drop some old baggage by choosing a more friendly 21st century name. But still no language-level parallelism.

---

## Parallel at last (#parallel)

And that's where Ruby 3's Ractors come in. For the first time ever, they allow Ruby (MRI) code to be truly executed in parallel.

Think of a Ractor less like a thread, and more like a parallel environment. Each Ractor gets its own GVL, meaning that the language's bottleneck is now at the Ractor level instead of the entire executing Ruby environment. Each Ractor will have at least one thread, but just like any normal Ruby process, can spin up new ones with `Thread.new`. Each thread within a Ractor is bound to the traditional parallel limitations of a GVL, but the Ractor's more fine-grain GVL (RVL?) instead of one shared universally.

To facilitate isolation, Ractors are only allowed to inherit state which is known to be safe globally. That safety is determined by immutability, so an integer, a frozen string, or a frozen array with each element frozen are shareable. An un-frozen string or object with mutable fields are not. The Ractor API exposes a `.shareable?` method to help tell the difference:

``` ruby
Ractor.shareable?(1)            #=> true
Ractor.shareable?('foo')        #=> false (unless `freeze_string_literals: true` is on)
Ractor.shareable?('foo'.freeze) #=> true
```

### Message passing (#message-passing)

Similar to Erlang, Ractors are a faithful implementation of the [actor model](https://en.wikipedia.org/wiki/Actor_model). Ractors have an _incoming port_ and an _outgoing port_, each of which lends itself to a separate style of passing messages.

**"Push"** message passing with `receive`/`send` sends non-blocking messages to a Ractor's incoming port:

``` ruby
receiver = Ractor.new do
  while message = Ractor.receive
    puts message
  end
end

loop do
  receiver.send 'ping'
  sleep(1)
end
```

<!--
<img src="/assets/images/nanoglyphs/018-ractors/receiver@2x.png" alt="Ractor receiver" class="img_constrained">
-->

![Ractor receiver](/assets/images/nanoglyphs/018-ractors/receiver.svg)

The incoming queue has unlimited size and therefore `send` will never block. `receive` blocks until a message is available.

---

**"Pull"** type messages use `take`/`yield`. A ractor _yields_ a value to its outgoing port, and a receiving process takes it when ready. Unlike push-style messaging, both ends block in this model.

This example is similar to the one above, but with roles now reversed, with the Ractor yielding values back to main:

``` ruby
sender = Ractor.new do
  loop do
    Ractor.yield 'ping'
    sleep(1)
  end
end

while message = sender.take
  puts message
end
```

<!--
<img src="/assets/images/nanoglyphs/018-ractors/sender@2x.png" alt="Ractor sender" class="img_constrained">
-->

![Ractor sender](/assets/images/nanoglyphs/018-ractors/sender.svg)

(As usual, I'm skipping most of the fine detail. See [communication between Ractors](https://github.com/ruby/ruby/blob/master/doc/ractor.md#communication-between-ractors) for more information.)

### Channels (#channels)

Ractor's implementation of the actor model is more pure than the one in a language like Go's, with the intention that all communication happens through message passing on the Ractors themselves. Few other concurrency primitives are provided.

For example, there's no built-in channel, but you can make one for yourself by creating a Ractor with combined pull and push primitives:

``` ruby
channel = Ractor.new do
  loop do
    Ractor.yield Ractor.receive
  end
end

# share the channel between multiple worker Ractors
5.times do |i|
  Ractor.new(channel, name: "ractor-#{i}") do |channel|
    while message = channel.take
      puts "#{name}: message"
    end
  end
end

loop do
  channel.send 'ping'
  sleep(1)
end
```

<!--
<img src="/assets/images/nanoglyphs/018-ractors/channel@2x.png" alt="Ractor channel" class="img_constrained">
-->

![Ractor channel](/assets/images/nanoglyphs/018-ractors/channel.svg)

The channel yields to consumers through its outgoing port and blocks as it does so, but it receives messages from its incoming port, which recall is allowed unlimited depth. Producers can send as much as they want to the pseudo-channel without blocking.

### Applicability today (#applicability)

Keen on putting Ractors to work, I tried taking a small part of the static generator that generates the very newsletter you're reading and re-implementing it in Ruby. [Modulir](https://github.com/brandur/modulir) spins up a worker pool of fixed size, then throws all the work it can find at it. Each article, fragment, TOML file, photograph, and newsletter is sent in to be parsed and rendered, and the pool waits for it all to be done. This is exactly the sort of work that lends itself well to parallelism, and should have been a great way to see Ractors in action.

But I didn't get very far. The [worker pool](https://gist.github.com/brandur/af8ac446e6fcaf4120639ceb53391231) was implemented without trouble, but I discovered the hard way that I could run practically nothing inside of it. For example:

``` ruby
class RenderFragmentJob
  def initialize(source)
    @source = source
  end

  def name
    "fragment: #{File.basename(@source)}"
  end

  def work
    data = File.read(@source)
    _, frontmatter_data, markdown_data = data.split("+++")

    meta = TOML::Parser.new(frontmatter_data).parsed
    content = Kramdown::Document.new(markdown_data).to_html

    ...
  end
end
```

This is _supposed_ to read a blog source, parse its TOML frontmatter, then render its markdown. But, none of it works.

Both the TOML parser and Markdown renderer depend on a gem called [parslet](https://github.com/kschiess/parslet) to the heavy lifting in parsing. Parslet in turn maintains an internal cache as it goes about tokenizing the file. That cache is shared state -- banned in Ractors -- and the job raises the moment the cache accessed. Other libraries were similar. Redcarpet is a C-based Markdown renderer I tried as an alternative. It failed as well, but for a different reason -- its C calls weren't blessed for Ractor-safety.

I was able to get Ractors to succeed at operations available from the standard library like parsing JSON or YAML, but nothing I tried from a third party gem worked. In time, the ecosystem will update itself to be more Ractor-friendly, but it'll be some time before they're practical for most uses.

## Grass roots (#grass-roots)

[Kir imagines](https://kirshatrov.com/2021/01/06/ruby-concurrency-and-ecosystem/) that Ractors will ease themselves into the ecosystem by coming in from the "bottom" of programs rather than the "top":

* The "top" of a program wraps the lion's share of its code and functionality. Think like Puma or Unicorn where a thread or process wraps the entirety of the executing stack including HTTP endpoints, database operations, internal commands, and utilities. In a bright future, those threads and processes we use today could be Ractors instead, sharing more resources while benefiting from real parallelism.

* The "bottom" of a program is at it edges where relatively little of its stack needs to be wrapped. Developers could offload workloads conducive to parallel work, but which don't yet depend on too much existing code or gems to work in Ractors.

In this vision, Ractors would be introduced near the bottom and work their way up as a greater part of the stack becomes Ractor-friendly over time.

It's an open question as to how successful that'll be. Even much more popular languages like Python took many, _many_ years to overhaul their ecosystems on major changes like transition from Python 2 to 3, or from synchronous code to [asyncio](https://docs.python.org/3/library/asyncio.html). Ruby's waited so long to introduce much-needed features like types and parallelism that it may be too little too late -- the language has lost much of the momentum it once had, and it's reasonably likely that fewer gems are "proactively" maintained, with authors willing to sink considerable time and energy into full retrofits for Ractor compatibility. Time will tell.

<img src="/photographs/nanoglyphs/018-ractors/sulphur-gondola@2x.jpg" alt="The gondola at Sulphur Mountain" class="wide">

---

## Artefacts (#artefacts)

Today's photos are from a hike on Sulphur Mountain near Banff in Alberta. As we were walking up, we stumbled across what might be the world's tiniest mystery, finding an ancient plaque embedded in one of the area's old trees, completely illegible.

It was a few feet off the trail on the other side of a snow drift, so we couldn't get close. We took some photos and poured over them later, puzzling them out from the zoomed in perspective of a computer screen. We were just starting to accept that the plate might be completely unintelligible, but then, a breakthrough. The word "Spruce" resolved itself on the tag's upper right. Combine that with the visible letters "Eng" in a few places, along with a known list of spruce trees, and we got to ["Engelmann Spruce"](https://en.wikipedia.org/wiki/Picea_engelmannii). The plaque names the tree it's attached to, a high altitude spruce mainly found in the tight cluster of BC, Alberta, Montana, and Idaho, and with a smattering of growth elsewhere in the US, including northern California. The line below is its botanical name and that of the botanist who published it, "_Picea Engelmannii_ Engelm" ([more on that](/fragments/engelmann-spruce-plaque-sulfur-mountain)).

The plaque looks so far out of antiquity that it might've been placed by [Norman Sanson](https://banff.ca/492/Norman-Sanson), famous for hiking this trail tirelessly for 30 years to take weather measurements at the top of Sulphur in the early 1900s. We made up this story, but we hope its true.

I learned how to identify an Engelmann Spruce, and that low stakes detective projects are good for the soul, even the smallest amongst them.

Until next time.

<img src="/photographs/nanoglyphs/018-ractors/engelmann-plaque@2x.jpg" alt="Plaque on Sulphur Mountain showing an Engelmann Spruce" class="wide">

[1] Technically JRuby and Rubinius have previously allowed parallel Ruby execution, but I'm speaking specifically about the [MRI](https://en.wikipedia.org/wiki/Ruby_MRI) because it's what a vast majority of users are on.
