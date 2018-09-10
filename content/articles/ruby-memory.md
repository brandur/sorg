---
title: "The Limits of Copy-on-write: How Ruby Allocates Memory"
published_at: 2017-08-28T14:07:33Z
hook: Why Ruby's scheme for memory allocation doesn't play
  nicely with copy-on-write, and how a compacting garbage
  collector will help.
location: San Francisco
hn_link: https://news.ycombinator.com/item?id=15118307
---

Anyone who's run Unicorn (or Puma, or Einhorn) may have
noticed a curious phenomena. Worker processes that have
been forked from a master start with low memory usage, but
before too long will bloat to a similar size as their
parent. In a big production installation, each worker can
be 100s of MBs or more, and before long memory is far and
away the most constrained resource on servers. CPUs sit
idle.

Modern operating systems have virtual memory management
systems that provide ***copy-on-write*** facilities
designed to prevent this exact situation. A process's
virtual memory is segmented into 4k pages. When it forks,
its child initially shares all those pages with its parent.
Only when the child starts to modify one of them does the
kernel intercept the call, copy the page, and reassign it
to the new process.

!fig src="/assets/ruby-memory/child-processes.svg" caption="Child processes transitioning from mostly shared memory to mostly copied as they mature."

So why aren't Unicorn workers sharing more memory? Most
software has a sizeable collection of static objects that
are initialized once, sit in memory unmodified throughout a
program's entire lifetime, and would be prime candidates
for staying shared across all workers. Apparently though,
practically nothing is reused, and to understand why, we'll
have to venture into how Ruby allocates memory.

## Slabs and slots (#slabs-and-slots)

Let's start with a very brief overview of object
allocation. Ruby requests memory from the operating system
in chunks that it refers to internally as ***heap pages***.
The naming is a little unfortunate because these aren't the
same thing as the 4k page that the OS will hand out (which
I will refer to hereafter as ***OS pages***), but a heap
page is mapped to a number of OS pages in virtual memory.
Ruby sizes its heap pages so that they'll maximize use of
OS pages by occupying an even multiple of them (usually
4x4k OS pages = 1x16k heap page).

!fig src="/assets/ruby-memory/heap-slots.svg" caption="A heap, its heap pages, and slots within each page."

You might also hear a heap page referred to as a "heap"
(plural "heaps"), "slab", or "arena". I'd prefer one of the
last two for less ambiguity, but I'm going to stick with
***heap page*** for a single chunk and ***heap*** for a
collection of heap pages because that's what they're called
everywhere in Ruby's source.

A heap page consists of a header and a number of
***slots***. Each slot can hold an `RVALUE`, which is an
in-memory Ruby object (more on this in a moment). A heap
points to a page, and from there heap pages point to each
other, forming a linked list that allows the entire
collection to be iterated.

### Harnessing the heap (#heap)

Ruby's heap is initialized by `Init_heap` ([in
`gc.c`][initheap]), called from `ruby_setup` ([in
`eval.c`][rubysetup]), which is the core entry point for a
Ruby process. Along with the heap, `ruby_setup` also
initializes the stack and VM.

``` c
void
Init_heap(void)
{
    heap_add_pages(objspace, heap_eden,
        gc_params.heap_init_slots / HEAP_PAGE_OBJ_LIMIT);

    ...
}
```

`Init_heap` decides on an initial number of pages based on
a target number of slots. This defaults to 10,000, but can
be tweaked through configuration or environmental variable.

``` c
#define GC_HEAP_INIT_SLOTS 10000
```

The number of slots in a page is calculated roughly how
you'd expect ([in `gc.c`][heappagealignlog]). We start with
a target size of 16k (also 2^14 or `1 << 14`), shave a few
bytes off for what `malloc` will need for bookkeeping [1],
subtract a few more bytes for a header, and then divide by
the known size of an `RVALUE` struct:

``` c
/* default tiny heap size: 16KB */
#define HEAP_PAGE_ALIGN_LOG 14
enum {
    HEAP_PAGE_ALIGN = (1UL << HEAP_PAGE_ALIGN_LOG),
    REQUIRED_SIZE_BY_MALLOC = (sizeof(size_t) * 5),
    HEAP_PAGE_SIZE = (HEAP_PAGE_ALIGN - REQUIRED_SIZE_BY_MALLOC),
    HEAP_PAGE_OBJ_LIMIT = (unsigned int)(
        (HEAP_PAGE_SIZE - sizeof(struct heap_page_header))/sizeof(struct RVALUE)
    ),
}
```

On a 64-bit system, an `RVALUE` occupies 40 bytes. I'll
save you some calculations, and just tell you that with its
defaults Ruby initially allocates 24 pages at 408 slots
each [2]. That heap is grown if more memory is needed.

### RVALUE: An object in a memory slot (#rvalue)

A single slot in a heap page holds an `RVALUE`, which is a
representation of an in-memory Ruby object. Here's its
definition ([from `gc.c`][rvalue]):

``` c
typedef struct RVALUE {
    union {
        struct RBasic  basic;
        struct RObject object;
        struct RClass  klass;
        struct RFloat  flonum;
        struct RString string;
        struct RArray  array;
        struct RRegexp regexp;
        struct RHash   hash;
        struct RData   data;
        struct RTypedData   typeddata;
        struct RStruct rstruct;
        struct RBignum bignum;
        struct RFile   file;
        struct RNode   node;
        struct RMatch  match;
        struct RRational rational;
        struct RComplex complex;
    } as;

    ...
} RVALUE;
```

For me this is where the mystique around how Ruby can
generically assign any type to any variable finally starts
to fall away; we immediately see that an `RVALUE` is just a big
list of all the possible types that Ruby might hold in
memory. These types are compacted with a C `union` so that
all the possibilities can share the same memory. Only one
can be set at a time, but the union's total size is only as
big as the largest individual type in the list.

To help concrete our understanding of a slot, lets look at
one of the possible types it can hold. Here's the common
Ruby string (from [ruby.h][rstring]):

``` c
struct RString {
    struct RBasic basic;
    union {
        struct {
            long len;
            char *ptr;
            union {
                long capa;
                VALUE shared;
            } aux;
        } heap;
        char ary[RSTRING_EMBED_LEN_MAX + 1];
    } as;
};
```

Looking at `RString`'s structure yields a few points of
interests:

* It internalizes `RBasic`, which is struct that's common
  to all in-memory Ruby types that helps distinguish
  between them.

* A union with `char ary[RSTRING_EMBED_LEN_MAX + 1]` shows
  that while the contents of a string might be stored in
  the OS heap, a short string will be inlined right into an
  `RString` value. Its entire value can fit into a slot
  without allocating additional memory.

* A string can reference another string (`VALUE shared` in
  the above) and share its allocated memory.

### VALUE: Both pointer and scalar (#value)

`RVALUE` holds many of Ruby's standard types, but it
doesn't hold all of them. Anyone who's looked at a Ruby C
extension will be familiar with the similarly named
`VALUE`, which is the general purpose type that's used to
pass around all Ruby values. Its implementation is quite a
bit simpler than `RVALUE`'s; it's just a pointer ([from
`ruby.h`][value]):

``` c
typedef uintptr_t VALUE;
```

This is where Ruby's implementation gets clever (or gross,
depending on how you think about these things). While
`VALUE` is often a pointer to an `RVALUE`, by comparing one
to constants or using various bit-shifting techniques, it
may also hold some scalar types that will fit into the
pointer's size.

`true`, `false`, and `nil` are the easiest to reason about;
they're all predefined as values in [ruby.h][rubyconsts]:

``` c
enum ruby_special_consts {
    RUBY_Qfalse = 0x00,		/* ...0000 0000 */
    RUBY_Qtrue  = 0x14,		/* ...0001 0100 */
    RUBY_Qnil   = 0x08,		/* ...0000 1000 */

    ...
}
```

A fixnum (i.e. very roughly a number that fits in 64 bits)
is a little more complicated. One is stored by
left-shifting a `VALUE` by one bit, then setting a flag in
the rightmost position:

``` c
enum ruby_special_consts {
    RUBY_FIXNUM_FLAG    = 0x01,	/* ...xxxx xxx1 */

    ...
}

#define RB_INT2FIX(i) (((VALUE)(i))<<1 | RUBY_FIXNUM_FLAG)
```

Similar techniques are used to store "flonums" (i.e.
floating point numbers) and symbols. When the time comes to
identify what type is occupying a `VALUE`, Ruby compares
pointer values to a list of flags that it knows about for
these stack-bound types; if none match, it goes to heap
([from `ruby.h`][rbclassof]):

``` c
static inline VALUE
rb_class_of(VALUE obj)
{
    if (RB_IMMEDIATE_P(obj)) {
        if (RB_FIXNUM_P(obj)) return rb_cInteger;
        if (RB_FLONUM_P(obj)) return rb_cFloat;
        if (obj == RUBY_Qtrue)  return rb_cTrueClass;
        if (RB_STATIC_SYM_P(obj)) return rb_cSymbol;
    }
    else if (!RB_TEST(obj)) {
        if (obj == RUBY_Qnil)   return rb_cNilClass;
        if (obj == RUBY_Qfalse) return rb_cFalseClass;
    }
    return RBASIC(obj)->klass;
}
```
Keeping certain types of values on the stack has the
advantage that they don't need to occupy a slot in the
heap. It's also useful for speed. "Flonum" was a relatively
recent addition to the language, and its author [estimated
that it sped up simple floating point calculations by
~2x][flonum].

#### Avoiding collision (#collision)

The `VALUE` scheme is clever, but how can we be sure that
the value of a scalar will never collide with a pointer?
This is where the cleverness gets kicked up a notch.
Remember how we talked about how an `RVALUE` is 40 bytes in
size? That sized combined with the use of an aligned
`malloc` means that every address for an `RVALUE` that Ruby
needs to put into a `VALUE` will be divisible by 40.

In binary, a number that's divisible by 40 will always have
three 0s in its rightmost bits (`...xxxx x000`). All the
flags that Ruby uses to identify stack-bound types like
fixnums, flonums, or symbols involve one of those three
bits, therefore guaranteeing perfect exclusivity between
them and an `RVALUE` pointer.

Folding extra information into pointers isn't specific to
Ruby. More broadly, a value that uses this technique is
called [a "tagged pointer"][taggedpointer].

## Allocating an object (#allocating)

Now that we've seen some basics of the heap, we're getting
closer to understanding why our mature Unicorn processes
can't share anything with their parent (some readers may
have guessed already). Let's get the rest of the way by
walking through how Ruby initializes an object; in this
case a string.

The entry point is `str_new0` (from [`string.c`][strnew0]):

``` c
static VALUE
str_new0(VALUE klass, const char *ptr, long len, int termlen)
{
    VALUE str;

    ...

    str = str_alloc(klass);
    if (!STR_EMBEDDABLE_P(len, termlen)) {
        RSTRING(str)->as.heap.aux.capa = len;
        RSTRING(str)->as.heap.ptr = ALLOC_N(char, (size_t)len + termlen);
        STR_SET_NOEMBED(str);
    }

    if (ptr) {
        memcpy(RSTRING_PTR(str), ptr, len);
    }

    ...

    return str;
}
```

Just like we speculated when examining `RString` earlier,
we can see that Ruby embeds the new value into the slot if
it's short enough. Otherwise it uses `ALLOC_N` to allocate
new space for the string in the operating system's heap,
and sets a pointer internal to the slot (`as.heap.ptr`) to
reference it.

### Initializing a slot (#slot-initialization)

After a few layers of indirection, `str_alloc` calls into
`newobj_of` back in [`gc.c`][newobjof]:

``` c
static inline VALUE
newobj_of(VALUE klass, VALUE flags, VALUE v1, VALUE v2, VALUE v3, int wb_protected)
{
    rb_objspace_t *objspace = &rb_objspace;
    VALUE obj;

    ...

    if (!(during_gc ||
          ruby_gc_stressful ||
          gc_event_hook_available_p(objspace)) &&
        (obj = heap_get_freeobj_head(objspace, heap_eden)) != Qfalse) {
        return newobj_init(klass, flags, v1, v2, v3, wb_protected, objspace, obj);
    }

    ...
}
```

Ruby asks the heap for a free slot with
`heap_get_freeobj_head` ([in `gc.c`][heapgetfreeobj]):

``` c
static inline VALUE
heap_get_freeobj_head(rb_objspace_t *objspace, rb_heap_t *heap)
{
    RVALUE *p = heap->freelist;
    if (LIKELY(p != NULL)) {
        heap->freelist = p->as.free.next;
    }
    return (VALUE)p;
}
```

Ruby has a global lock (the GIL) that ensures that Ruby
code can only be running in one place across any number of
threads, so it's safe to simply pull the next available
`RVALUE` off the heap's `freelist` and repoint it to the
next free slot in line. No finer grain locks are required.

After procuring a free slot, `newobj_init` runs some
generic initialization on it before it's returned to
`str_new0` for string-specific setup (like copying in the
actual string).

### Eden, the tomb, and the freelist (#eden)

You may have noticed above that Ruby asked for a free slot
from `heap_eden`. ***Eden***, named for the biblical garden
[3], is the heap where Ruby knows that it can find live
objects. It's one of two heaps tracked by the language.

The other is the ***tomb***. If the garbage collector
notices after a run that a heap page has no more live
objects, it moves that page from eden to the tomb. If at
some point Ruby needs to allocate a new heap page, it'll
prefer to resurrect one from the tomb before asking the OS
for more memory. Conversely, if heap pages in the tomb stay
dead for long enough, Ruby may release them back to the OS
(in practice, this probably doesn't happen very often,
which we'll get into in just a moment).

We talked a little about how Ruby allocates new pages
above. After being assigned new memory by the OS, Ruby will
traverse a new page and do some initialization ([from
`gc.c`][heappageallocate]):

``` c
static struct heap_page *
heap_page_allocate(rb_objspace_t *objspace)
{
    RVALUE *start, *end, *p;

    ...

    for (p = start; p != end; p++) {
        heap_page_add_freeobj(objspace, page, (VALUE)p);
    }
    page->free_slots = limit;

    return page;
}
```

Ruby calculates a memory offset for the page's `start` and
`end` slots, then proceeds to walk from one end of it to
the other and invoke `heap_page_add_freeobj` ([from
`gc.c`][heappageaddfreeobj]) on each slot along the way:

``` c
static inline void
heap_page_add_freeobj(rb_objspace_t *objspace, struct heap_page *page, VALUE obj)
{
    RVALUE *p = (RVALUE *)obj;
    p->as.free.flags = 0;
    p->as.free.next = page->freelist;
    page->freelist = p;

    ...
}
```

The heap itself tracks a single `freelist` pointer to a
slot that it knows is free, but from there new free slots
are found by following a `free.next` on the `RVALUE`
itself. All known free slots are chained together by a long
linked list that `heap_page_add_freeobj` has constructed.

!fig src="/assets/ruby-memory/freelist.svg" caption="A heap's freelist pointer to a free RVALUE, and the continuing linked list."

`heap_page_add_freeobj` is called initializing a page. It's
also called by the garbage collector when it frees an
object. In this way, slots get added back to `freelist` so
that they can be reused.

## Closing the case on bloated workers (#workers)

Ruby has an elaborate scheme for memory management, but
reading between the lines, you may also have noticed that
something that's not going to mesh well with an operating
system's copy-on-write. Ruby allocates expansive heap pages
in memory, stores objects to them, and GCs slots when able.
Free slots are tracked carefully, and the runtime has an
efficient way of finding them. However, despite all this
sophistication, _a live slot will never change position
within or between heap pages_.

In a real program where objects are being allocated and
deallocated all the time, pages quickly become a mix of
objects that are alive and dead. This gets us back to
Unicorn: the parent process sets itself up, and by the time
it's ready to fork, its memory looks like that of a typical
Ruby process with live objects fragmented across available
heap pages.

Workers kick off with the entirety of their memory shared
with their parent. Unfortunately, the first time child
initializes or GCs even a single slot, the operating system
intercepts the call and copies the underlying OS page.
Before long this has happened on every page allocated to
the program, and child workers are running with a copy of
memory that's completely divergent from their parent's.

Copy-on-write is a powerful feature, but one that's not of
much practical use to a forking Ruby process.

## Copy-on-write on the mind (#copy-on-write)

The Ruby team is well-acquainted with copy-on-write, and
has been writing optimizations for it for some time. As an
example, Ruby 2.0 introduced heap "bitmaps". Ruby uses a
mark-and-sweep garbage collector which traverses all of
object space and "marks" live objects it finds before going
through and "sweeping" all the dead ones. Marks used to be
a flag directly on each slot in a heap page, which had the
effect of GCs running on any fork and performing their mark
pass to cause every OS page to be copied from the parent
process.

The change in Ruby 2.0 moved those mark flags to a
heap-level "bitmap" which is a big sequence of single bits
mapping back slots on the heap. The GC performing a pass on
a fork would only copy the OS pages needed for bitmaps,
allowing more memory to be shared for longer.

### The future with compaction (#compaction)

Upcoming changes are even more exciting. For some time
Aaron Patterson has been [talking publicly about
implementing compaction in Ruby's GC][aaroncompact], and
has suggested that [it's spent time in production at GitHub
with some success][aaronprod]. Practically, this would look
like a method named `GC.compact` that's called before
workers fork:

``` ruby
# Called before a parent forks any workers
before_fork do
  GC.compact
end
```

The parent would get a chance to finish churning objects as
part of its initialization, then take the objects that are
still living and move them into slots on a minimal set of
pages that are likely to be stable for a long time. Forked
workers can share memory with their parent for longer.

!fig src="/assets/ruby-memory/compaction.svg" caption="A fragmented heap before and after GC compaction."

For anyone running big Ruby installations (GitHub, Heroku,
or like we are at Stripe), this is _really_ exciting work.
Even deploying to high-memory instances, memory is still
usually the limiting resource on the number of runnable
workers. GC compaction has the potential to shave off a big
chunk of the memory that every worker needs. With the
savings we can run more workers per box and fewer total
boxes -- with very few caveats, the fleet is immediately
cheaper to run.

[1] `malloc`'s bookkeeping is compensated for so that we
can keep a heap page fitting nicely into a multiple of OS
pages without overflowing onto another OS page. Because
pages are the smallest unit that an OS will allocate to a
process, this would make for an inefficient use of memory.

[2] Astute readers may notice that we start with "only"
9,792 (24 * 408) total slots, despite requesting 10,000.

[3] The naming "eden" is also somewhat conventional in the
world of garbage collectors. You'll find reference to "eden
space" in Java's VM too.

[aaroncompact]: https://twitter.com/tenderlove/status/801576703361355776
[aaronprod]: https://twitter.com/tenderlove/status/844248259631566848
[flonum]: https://bugs.ruby-lang.org/issues/6763
[heapgetfreeobj]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/gc.c#L1783
[heappageaddfreeobj]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/gc.c#L1432
[heappagealignlog]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/gc.c#L660
[heappageallocate]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/gc.c#L1517
[initheap]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/gc.c#L2376
[newobjof]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/gc.c#L1952
[rbclassof]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/include/ruby/ruby.h#L1970
[rstring]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/include/ruby/ruby.h#L954
[rubyconsts]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/include/ruby/ruby.h#L405
[rubysetup]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/eval.c#L46
[rvalue]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/gc.c#L410
[strnew0]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/string.c#L702
[taggedpointer]: https://en.wikipedia.org/wiki/Tagged_pointer
[value]: https://github.com/ruby/ruby/blob/917beef327117cfeee4e1f455d650f08c2268d7e/include/ruby/ruby.h#L79
