+++
hook = "Treating UUIDs at 16-byte arrays instead of strings."
published_at = 2022-04-14T00:30:45Z
title = "UUID code stanzas in Ruby"
+++

Most Rubyists will be familiar with generating UUIDs by way of `SecureRandom`:

``` ruby
SecureRandom.uuid
```

This produces the normal UUID hex string representation like `27a532f4-5bc8-4810-b602-88475a93167c` that everyone knows well, and widespread practice in Ruby is to treat UUIDs exclusively as strings, passing these formatted strings around everywhere in code.

But although it's so common to see UUIDs in their string representation, they're more simply (and succinctly) defined as 16 bytes worth of data. Implementations in languages like Go do exactly that:

``` go
type UUID [16]byte
```

Combined with other binary-friendly packages like say the `pgx` driver for Postgres, UUIDs may never have even have to be materialized as a string -- generated, persisted, loaded, and spending their entire lifecycle as 16-byte arrays.

For fun, last week I tried implementing a lower-level UUID data type for Ruby. I'm not going to gem-ify it, but here's a couple code stanzas from the project.

## Generating a random UUID (#gen_random)

Generating a new V4 UUID:

``` ruby
sig { returns(Uuid) }
def self.gen_random
  byte_str = SecureRandom.random_bytes(16)

  # V4 random UUIDs use 4 bits to indicate a version and another 2-3 bits to
  # indicate a variant. Most V4s (including these ones) are variant 1, which
  # is 2 bits.
  byte_str.setbyte(6, T.unsafe((byte_str.getbyte(6) & 0x0f) | 0x40)) # version 4
  byte_str.setbyte(8, T.unsafe((byte_str.getbyte(8) & 0x3f) | 0x80)) # variant 1 (10 binary)

  new(byte_str)
end
```

We still use `SecureRandom` for a cryptographically-secure RNG, but get bytes instead, then doing a little bit manipulation to add UUID version and variant.

Note that byte strings in Ruby are still strings rather than a separate type. Along with the normal built-ins to work with characters, Ruby strings also has `#bytes`, `#bytesize`, `#getbyte`, etc. to work with bytes instead.

## Formatting to string (#to_s)

I mostly stole the byte string to string `#to_s` implementation from the Ruby standard library:

``` ruby
sig { returns(String) }
def to_s
  return @str if @str

  # shamelessly copied from Ruby's stdlib
  ary = @byte_str.unpack("NnnnnN")
  @str = "%08x-%04x-%04x-%04x-%04x%08x" % ary
  @str
end
```

Code sure doesn't get much harder to parse than that! Unpack's `N` returns an unsigned 32-bit integer in big-endian byte order, and `n` is the same except 16-bit. So here we unpack 128 bits of data into an array of six integers (32 + 16 + 16 + 16 + 16 + 32), then format them into hex with `%x`.

## Parsing from string (#parse)

For parsing a string to UUID, I elected for a simple implementation, although I'm sure there's a much more optimal form out there:

``` ruby
PATTERN = /\A[0-9a-f]{8}\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\b[0-9a-f]{12}\z/

sig { params(val: String).returns(Uuid) }
def self.parse(val)
  unless val.match? PATTERN
    raise DecodeError, "value not a UUID: #{val}"
  end

  id = new(val.tr("-", "").downcase.scan(/../).map { |z| T.unsafe(z).hex }.pack("c*"))
  id.instance_variable_set(:@str, val)
  id
end
```

And of course, comparison and hash implementations so UUIDs can be compared and used as keys in dictionaries:

``` ruby
# Implements standard equality. Two UUIDs are considered equal if they have
# the same underlying value (even if they're different objects).
sig { params(other: T.untyped).returns(T::Boolean) }
def ==(other)
  other.is_a?(Uuid) && other.instance_variable_get(:@byte_str) == @byte_str
end

# Implements hash equality so that in conjuction with `#hash`, UUIDs can be
# used as keys in hashes.
sig { params(other: T.untyped).returns(T::Boolean) }
def eql?(other)
  self == other
end

# Implements getting a hash value so that in conjuction with `#eql?`, UUIDs
# can be used as keys in hashes.
sig { returns(Integer) }
def hash
  @byte_str.hash
end
```

## To Sequel literal (#sequel)

`#sql_literal` can be implemented so that the UUID type can be used in Sequel queries:

``` ruby
sig { params(dataset: Sequel::Postgres::Dataset).returns(String) }
def sql_literal(dataset)
  %('#{self}')
end
```

String interpolation on `self` invokes `#to_s`, producing a string-formatted UUID.

The type is then used like:

``` ruby
Account.where(id: Uuid.parse(...))
```

## Redux (#redux)

Full code [is here](https://gist.github.com/brandur/1bddb5215540889983dc7e3a66ef4e41). You may have to strip the Sorbet signatures yourself.

Although treating UUIDs as a custom byte string data type might be a little more performant [1], we have some more compelling reasons to do it:

* We use a public-facing ID representation called an ["EID"](https://docs.crunchybridge.com/api-concepts/eid/) which looks like `rvf73a77ozfsvcttryebfrnlem` and is also 16 bytes. With EIDs and UUIDs stored as strings, converting one format to the other involves expensive parsing. But with both treated as byte strings, they can be interchanged for free since the underlying value is the same.

* We're using [Sorbet](https://sorbet.org/) for type checking, and annotating EIDs and UUIDs as proper types instead of strings makes code safer. We've previously had multiple bugs where an EID string was actually a UUID string by mistake and leaked somewhere where it wasn't compatible.

[1] Performance and memory overhead _should_ be better, but you'd want to check this because Ruby core types like strings are generally better optimized than anything you can build in pure Ruby code.
