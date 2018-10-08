---
title: Passive Safety
subtitle: Designing robust applications with transactions
published_at: 2018-10-07T21:55:22Z
location: San Francisco
event: FakeConf
---

class: middle

# Designing Robust, Passively Safe Applications With Transactions

<!-- Title slide. Content hidden. Speaker notes used as intro. -->

???

This is a talk intro.

---

class: middle

Follow along:<br>
https://brandur.org/passive-safety

Find me on Twitter:<br>
[@brandur](https://twitter.com/brandur)

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

Map transactions onto units of work in an application like HTTP requests or asynchronous jobs.

TODO: Diagram of transaction mapped to work unit.

???

Generally a good way to do this is to map transactions onto units of work in your application that are meant to execute atomically. For example, that might be an HTTP request in a web service where we'd begin a transaction when the request started processing and commit the transaction when we're ready to send the result back to the client. A unit of work can also be something like an synchronous executing job.

The transaction guarantees safety for the duration of the work unit. If we get halfway through serving an HTTP request and realize that the client has disconnected, we abort the transaction and roll back its results. Our data's consistency is safe.

---

# Foreign mutations and leaking state

Watch for breaches in state encapsulation where a work unit manipulates state outside of the local database.

e.g. Charging a credit card through Stripe. Provisioning a server with AWS.

The database can roll back local changes, but leaves foreign state orphaned.

TODO: Diagram of transaction with foreign state mutation partway.

???

The simple technique of mapping transactions to units of work works well in most cases, but it breaks down in cases where we start to manipulate non-local state that lives outside of our database. For example, during the lifetime of our HTTP request we might want to charge a user's credit card through Stripe, or provision a new server with AWS.

If the request fails, the transaction of course can only roll back our local changes, and any foreign state we mutated continues to exist. This is made even worse because if we did persist some information about the foreign state we changed, that information is rolled back along with the rest of the transaction. That foreign state is now orphaned, and that might mean we execute a costly operation like charging a user or provisioning a server and now know nothing about.

---

# Mutate idempotently

At the end of atomic phases, checkpoint what's happened so far, and save properties of the outgoing foreign mutation.

Use idempotency primitives of those foreign APIs like **idempotency keys**. These ensure that we don't do double work

???

The first thing we need to do to solve this problem is to introduce idempotency, which we'll want in both our own service and any foreign services that we're calling out to. With idempotency clients can make the same call any number of times and be guaranteed the same result. If we're creating a charge through Stripe, only one charge will ever be created. If we're provisioning a server, only one server will ever be provisioned.

Idempotency is normally achieved through the use of an **idempotency key**, which is a random string with enough entropy to uniquely identify a request, and which we'll hand off to foreign services as we're making changes. The naming of this idea can vary. With Stripe it's called an `Idempotency-Key` and sent in via a header. With Amazon, it's called a `ClientToken` and sent via a query parameter.

---

# Design atomic phases

We're not going to throw transactions out, but we do need to be more careful with them.

Identify foreign mutations. Wrap operations between them in transactions. These are **atomic phases**.

---

# Coverge on consistency

Have clients retry requests.

Use a backoff schedule. First retry should be soon to protect against intermittent network problems. Last retry should be much later to hedge against hard down services or application bugs that take time to fix.

When retrying a work unit on the server, skip what's already been done. Reuse idempotency keys.

---

# Passive safety

<!-- vim: set tw=9999: -->
