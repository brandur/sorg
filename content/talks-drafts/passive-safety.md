---
title: Passive Safety
subtitle: Designing robust applications with transactions
published_at: 2018-10-07T21:55:22Z
location: San Francisco
event: FakeConf
---

class: middle

# Passive Safety

## Designing robust, passively safe applications with transactions

<!-- Title slide. Content hidden. Speaker notes used as intro. -->

???

This is a talk intro.

This talk was delivered as a 5-minute lightning talk and is by necessity light on detail. You can read about these ideas in greater depth in this article about [implementing Stripe-like idempotency keys in Postgres](/idempotency-keys).

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

Good **abstractions** are leverage -- they do something complicated while presenting an easy-to-use interface.

Some great abstractions: files, threads, libc memory management, TCP in your OS.

???

As engineers, we should always be looking to gain leverage to help us build reliable software more quickly. One pattern for leverage are abstractions -- building blocks that do something quite complicated, and present a comparatively easy-to-use interface.

We've all seen bad abstractions that are too complex or do too little, and end up costing more than they're worth. But if we think about the computing stack that we use every day, there are many successful abstractions that work *very* well: files, threads, libc memory management, or TCP.

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

Like with other places in the computing stack, databases are an abstraction, and provide abstractions. One of the most powerful of these is the **transaction**.

Transactions give us incredibly strong guarantees for manipulating data, guarantees that would be very difficult to implement for ourselves. A good example of this are the ACID properties which give us atomic commits and rollbacks, definitive data consistency in constraints like foreign keys, isolation from concurrent operations even if they're operating on the same data, and guaranteed persistence when a transaction commits.

---

# Application safety

Process enough volume for long enough and edge cases appear: application bugs, network connectivity problems, client disconnects, process crashes, out-of-memory, etc.

TODO: Diagram of benign failures along a transaction.

???

These guarantees are very useful for building reliable software. An application that processes enough volume will eventually see unhappy edge cases. Application bugs that raise errors midway through work, network connectivity problems, clients that disconnect mid-request, out-of-memory crashes, and many more.

Without transactions these sorts of failures would lead to broken state, but with them, partial state is rolled back. Failures still aren't benign from the perspective of our users, but at least they're safe.

---

# Map transactions to work

Map transactions onto units of work in an application like HTTP requests or asynchronous jobs.

TODO: Diagram of transaction mapped to work unit.

???

A good architectural strategy is to map transactions onto logical units of work in an application. For example, that might be an HTTP request in a web service where we'd begin a transaction when the request started processing, and commit it when we're ready to send results back to the client. A unit of work can also be something like an asynchronous executing background job.

The transaction guarantees safety for the duration of the work unit. If we get halfway through serving an HTTP request and realize that the client has disconnected, the transaction is aborted and the database is rolled back to a consistent state.

---

# Foreign mutations and leaking state

Watch for breaches in state encapsulation where a work unit manipulates state outside of the local database.

e.g. Charging a credit card through Stripe. Provisioning a server with AWS.

The database can roll back local changes, but leaves foreign state orphaned.

TODO: Diagram of transaction with foreign state mutation partway.

???

The simple technique of mapping transactions to work units works well in most cases, but breaks down when we start to manipulate non-local state that lives outside of our database. For example, we might want to charge a user's credit card through Stripe, or provision a new server with AWS.

If the request fails, the transaction of course can only roll back local changes, so foreign state changes continue to exist. This is made worse because if we had persisted some information about those changes, like the identity of the user whose credit card we charged, that information is lost with the transaction rollback.

That foreign state is now orphaned, which means that we might have a credit card charge or provisioned server that we now know nothing about.

---

# Mutate foreign state idempotently

Idempotency: Execute any number of times to produce the same result.

e.g. Create only one charge on Stripe. Provision only one server through AWS.

An **idempotency key** uniquely identifies a request. Naming can vary: `Idempotency-Key` in Stripe, `ClientToken` in AWS.

???

A key to solving this problem is idempotency. With idempotency, clients can make the same request any number of times and be guaranteed the same result. If we're creating a charge through Stripe, only one charge will ever be created for a given idempotent request. If we're provisioning a server, only one server will ever be provisioned.

Idempotency is made possible with a client-transmitted token called an **idempotency key** which is a random string with enough entropy to uniquely identify a request. Subsequent retries of the same idempotent request transmit the same idempotency key.

The naming and implementation of this concept varies. With Stripe, these tokens are sent in a header called `Idempotency-Key`. With Amazon, they're `ClientToken`s and sent via query parameter.

---

# Design atomic phases

Always use idempotency keys when executing foreign mutations.

Work between foreign mutations are **atomic phases** that are safe to group into transactions. Store idempotency keys and parameters of foreign mutations as part of the previous atomic phase.

If the work unit subsequently fails, we roll back to where our last atomic phase committed. Foreign state is no longer orphaned.

???

So let's get back to ensuring safety with transactions even with foreign state mutations. When designing our work units we need to identify places where we're mutating foreign state, and make sure that we're sending an idempotency key out with those requests.

These foreign state mutations are our transactional boundaries. We can define all work between them as **atomic phases** and any number of database operations that take place within one are safe to wrap in a transaction.

Before executing a foreign state change we commit data produced in the preceding atomic phase, along with information about the foreign request we're about to make including its idempotency key and parameters.

Now if the request fails, we still have a local record of it. This applies to subsequent atomic phases as well. If the first atomic phase in a work unit succeeds but the next fails, we're still left with the committed results of the first.

---

# Converging consistency

Even some committed atomic phases might leave the entire work unit only partially complete.

Our own services should be idempotent, with clients retrying requests to push work to completion.

Use exponential backoff schedules. First retry should be soon in case a failure was just an intermittent network problem. Last retry should be much later in case a failure was an application bug that takes time to find and fix.

???

Now although atomic phases give us assurance that we won't lose track of anything, a work unit that's left only partially completed still leaves us in an overall undesirable state.

This is where idempotency comes into play for our own application. We should provide our own version of idempotency so that our clients can continue retrying requests until they're fully executed to satisfaction. An initial try might fail after only a single atomic phase completed, but subsequent retries push the work unit forward until all phases are complete.

Clients should use exponential backoff to protect against a variety of failure modes. The first retry should come quickly because there's a good chance the failure was the result of an intermittent problem like a network hiccup. The last retry should be hours or even days later in the case the failure was caused by an application bug that will take time to find and remediate.

---

# Cultivate passive safety

Wrap work units in transactions.

For more complex operations, use transactions for spans that we know to be safe. Use idempotency keys for foreign mutations.

Work towards **passive safety**: largely guaranteed consistency with little operator effort.

???

So to recap, the database transaction is a powerful abstraction, and powerful abstractions are good for building reliable software.

Wrap units of application work into transactions for an easy way to protect the consistency of your data.

For more complex operations, use transactions along spans that we know to be safe. Use idempotency keys when talking to foreign services and make sure to commit state before doing so so that it's not lost.

These steps go a long way towards ensuring **passive safety** which means that consistency is largely guaranteed with little effort on the part of a system's maintainers. That frees up their time to work on more useful things.

<!-- vim: set tw=9999: -->
