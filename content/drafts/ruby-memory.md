---
title: "The Limits of Copy-on-write: How Ruby Allocates Memory"
#title: How Ruby Allocates Memory, and the Limits of Copy-on-write
#title: "The Curious Case of the Bloated Unicorn: How Ruby Manages Memory"
published_at: 2017-08-24T13:39:04Z
hook: TODO
---

Anyone who's run Unicorn (or Puma, or Einhorn) will have
noticed a curious phenomena. Worker processes that have
been forked from a master start with low memory usage, but
before too long will bloat to a similar size as their
parent. In a big production installation, workers can be
100s of MBs or more, and before long memory is far and away
the constrained resource on your boxes. CPUs sit mostly
idle.

Modern operating systems have virtual memory management
systems that provide ***copy-on-write*** facilities
designed to prevent this exact situation.

This is in spite of the ***copy-on-write*** (COW) features
provided by the virtual memory systems in modern operating
systems. As many readers will be aware, these systems will
reduce startup and runtime overhead by having forked
children share the same memory space as their parent. Only
when a child modifies shared memory does the OS intercept
the call and copy the page for exclusive use by the child.

!fig src="/assets/ruby-memory/child-processes.svg" caption="Child processes transitioning from mostly shared memory to mostly copied as they mature."

So what's going on here? Most programs have a sizeable
collection of static objects that are initialized once, and
sit in memory largely unmodified throughout its entire
lifetime. Child processes should have no problem sharing
that collection with their parent, but apparently they're
not, or at least not doing it well. To get to the heart of
the problem, we'll have to understand how Ruby allocates
memory.

## Slabs and slots (#slabs-and-slots)

### A heap (#heap)

### RVALUE: An object (#rvalue)

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

### VALUE: Objects and scalars (#value)

[From `ruby.h`][rbclassof]:

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

## Allocating an object (#allocating)

## Closing the case on bloated workers (#bloated-workers)

## Towards compaction (#compaction)

[rbclassof]: https://
