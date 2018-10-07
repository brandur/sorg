---
title: Passive Safety
subtitle: Designing robust applications with transactions
published_at: 2018-10-07T21:55:22Z
location: San Francisco
event: FakeConf
---

# Designing Robust, Passively Safe Applications With Transactions

???

This is a talk intro.

---

Follow along: https://brandur.org/passive-safety

Find me on Twitter: [@brandur](https://twitter.com/brandur)

???

I publish most of my work on this site or [Twitter](https://brandur.org/twitter).

---

# Abstractions

Engineers need leverage to build reliable software more quickly.

Good **abstractions** are levarage -- they do something complicated while presenting an easy-to-use interface.

Some great abstractions: files, threads, libc memory management, TCP in your OS.

???

As engineers, we should always be looking to gain leverage to help us build reliable software more quickly. One pattern for leverage are abstractions -- modules that do something quite complicated, but present a comparatively easy-to-use interface.

We've all seen bad abstractions that are too complex and do too little, which end up costing more than they're worth. But if we think about the computing stack that we use every day, there are many abstractions with a very long history of working *very* well: files, threads, libc memory management, or TCP.

---

# Transactions

Databases provide abstractions. **Transactions** are amongst the most powerful of them.

ACID: atomic commit and rollback, definitive data consistency, isolation from concurrent operations, and guranteed persistence.

``` sql
BEGIN TRANSACTION;

INSERT ...;

UPDATE ...;

COMMIT;
```

???

Like with other places in the computing stack, databases are an abstraction and provide abstractions. One of the most powerful of these is the **transaction**.

Transactions are practically magical in the strong guarantees they make and give us a way to do hard things that would be very difficult to implement for ourselves. A good example of this are the ACID properties which give us atomic commits and rollbacks, definitive data consistency in constraints like foreign keys, isolation from concurrent operations even if they're operating on the same data, and guaranteed persistence when our transaction commits.

---

# Application safety

Process enough volume for long enough and edge cases appear: application bugs, network connectivity problems, client disconnects, process crashes, out-of-memory, etc.

TODO: Diagram of benign failures along a transaction.

???

These guarantees are very useful for building reliable services. If you process enough traffic for long enough you're going to start to see edge cases appear like application bugs that raise an error midway through some work, network connectivity problems, clients that disconnect mid-request, out-of-memory crashes, and many more.

Normally these sorts of failures would lead to broken state, but not so with transactions, which will roll back partial state and make sure that state is always consistent at the beginning or end of any transaction. Failures still aren't benign from the perspective of our users, but at least they're safe.

---

# Map transactions to work

It's useful to define a **work unit**: some discrete amount of work that maps nicely to a transaction.

Might be an HTTP request being served or an async job being worked.

TODO: Diagram of transaction mapped to work unit.

---

# x

TODO: The transaction guarantees safety along this dimension.

---

# Foreign mutations

Watch for breaches in state encapsulation where state is manipulated that's external to your datastore.

TODO: Diagram of transaction with foreign state mutation partway.

---

# Leaking state

With foreign state mutations, transactions can no longer guarantee safety.

Local changes are discarded on a rollback, but foreign changes persist. State and resources are leaked.

e.g. A financial transaction executed against Stripe, or server provisioned against AWS.

---

# Design atomic phases

We're not going to throw transactions out, but we do need to be more careful with them.

Identify foreign mutations. Wrap operations between them in transactions. These are **atomic phases**.

---

# Mutate idempotently

At the end of atomic phases, checkpoint what's happened so far, and save properties of the outgoing foreign mutation.

Use idempotency primitives of those foreign APIs like **idempotency keys**. These ensure that we don't do double work

---

# Coverge on consistency

Have clients retry requests.

Use a backoff schedule. First retry should be soon to protect against intermittent network problems. Last retry should be much later to hedge against hard down services or application bugs that take time to fix.

When retrying a work unit on the server, skip what's already been done. Reuse idempotency keys.

---

# Passive safety
