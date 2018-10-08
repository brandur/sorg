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

# Mutate foreign state idempotently

Idempotency: Execute any number of times to produce the same result.

e.g. Create only one charge on Stripe. Provision only one server through AWS.

An **idempotency key** uniquely identifies a request. Naming can vary: `Idempotency-Key` in Stripe, `ClientToken` in AWS.

???

The first thing we need to do to solve this problem is to introduce idempotency, which we'll want in both our own service and any foreign services that we're calling out to. With idempotency clients can make the same call any number of times and be guaranteed the same result. If we're creating a charge through Stripe, only one charge will ever be created. If we're provisioning a server, only one server will ever be provisioned.

Idempotency is normally achieved through the use of an **idempotency key**, which is a random string with enough entropy to uniquely identify a request, and which we'll hand off to foreign services as we're making changes. The naming of this idea can vary. With Stripe it's called an `Idempotency-Key` and sent in via a header. With Amazon, it's called a `ClientToken` and sent via a query parameter.

---

# Design atomic phases

Always use idempotency keys when executing foreign mutations.

Work between foreign mutations are **atomic phases** that are safe to group into transactions. Store idempotency keys and parameters of foreign mutations as part of the previous atomic phase.

If the work unit subsequently fails, we roll back to where our last atomic phase committed. Foreign state is no longer orphaned.

???

So let's get back to ensuring safety with transactions even with foreign state mutations. When designing our units of work we need to identify any places where we're mutating foreign state, and making sure that we're sending an idempotency key out with those requests.

These foreign state mutations are our transactional boundaries. We can define all work between them as **atomic phases** and any number of database operations that take place there are safe to wrap in a transaction. Before executing any foreign state change we commit any data produced in the atomic phase so far along with information about the request we're about to make including the idempotency key and any parameters.

Now if that request were to fail we will not have lost track of it in our local database. We will have information on the charge that we tried to create through Stripe or the server we tried to provision through AWS. This applies for for all subsequent parts of the request as well. Our next atomic phase could fail and roll back, but we're still left with the committed results of the first one.

---

# Converging consistency

Even some committed atomic phases might leave the entire work unit only partially complete.

Our own services should be idempotent, with clients retrying requests to push work to completion.

Use exponential backoff schedules. First retry should be soon in case a failure was just an intermittent network problem. Last retry should be much later in case a failure was an application bug that takes time to find and fix.

???

Now although introducing atomic phases has given us a little extra assurance that we won't lose track of anything, a unit of work that's left only partially completed will still leave us in an undesirable state.

This is where idempotency comes into play for our own service. We should provide our own version of an idempotency key so that our own clients can continue retrying a request until it's been fully executed to satisfaction. The first try might only complete a single atomic phase before a failure, but subsequent retries will continue to push the work unit forward until all phases have completed successfully.

We'd normally recommend an exponential backoff schedule to protect against many types of failure. After a failed request the first retry should come quite quickly because there's a decent likelihood that the failure was the result of an intermittent network problem. The last retry should be hours or days later in case the failure was due to an application bug that will take time for engineers to find and remediate.

The same idea applies to other types of work units like background jobs as well. For a background job we'd probably want an automatic retry schedule built in.

---

# Cultivate passive safety

Wrap work units in transactions.

For more complex operations, use transactions for spans that we know to be safe. Use idempotency keys for foreign mutations.

Work towards **passive safety**: largely guaranteed consistency with little operator effort.

???

To recap, the database transaction is a powerful abstraction, and we like powerful abstractions.

Wrap units of application work into transactions for an easy way to protect the consistency of your data.

For more complex operations, use transactions along spans that we know to be safe. Use idempotency keys when talking to foreign services and make sure to commit that state before talking to one so that we don't lose track of it.

These steps go a long way towards ensuring a form of **passive safety** which means that consistency is largely guaranteed with very little effort on the part of a system's maintainers, and that frees up their time so that they can do more useful things.

<!-- vim: set tw=9999: -->
