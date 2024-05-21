+++
hn_link = "https://news.ycombinator.com/item?id=40352686#40353255"
hook = "Maximizing Postgres connection economy by using a single connection per program to receive and distribute all listen/notify notifications."
location = "Berlin"
published_at = 2024-05-06T07:54:07+02:00
tags = ["postgres"]
title = "The Notifier Pattern for Applications That Use Postgres"
+++

[Listen/notify in Postgres](https://www.postgresql.org/docs/current/sql-listen.html) is an incredible feature that makes itself useful in all kinds of situations. I've been using it a long time, started taking it for granted long ago, and was somewhat shocked recently looking into MySQL and SQLite to learn that even in 2024, no equivalent exists.

In a basic sense, listen/notify is such a simple concept that it needs little explanation. Clients subscribe on topics and other clients can send on topics, passing a message to each subscribed client. The idea takes only three seconds to demonstrate using nothing more than a psql shell:

``` sql
=# LISTEN test_topic;
LISTEN
Time: 2.828 ms

=# SELECT pg_notify('test_topic', 'test_message');
 pg_notify
-----------

(1 row)

Time: 17.892 ms
Asynchronous notification "test_topic" with payload "test_message" received from server process with PID 98481.
```

But despite listen/notify's relative simplicity, when it comes to applications built on top of Postgres, it's common to use it less than optimally, eating through scarce Postgres connections and with little regard to failure cases.

---

Here's where the **notifier pattern for Postgres** comes in. It's an extremely simple idea, but in my experience, one that's rarely seen in practice. Let's start with these axioms:

* `LISTEN`s are affixed to specific connections. After listening, the original connection must still be available somewhere to successfully receive messages.

* There may be many components within an application that'd like to listen on topics for completely orthogonal uses.

* Despite optimizations over the years, connections in Postgres are still somewhat of a precious, limited resource, and should be conserved. We'd like to minimize the number of them required for listen/notify use.

* A single connection can listen on any number of topics.

With those stated, we can explain the role of the notifier. Its job is to **hold a single Postgres connection per process, allow other components in the same program to use it to subscribe to any number of topics, wait for notifications, and distribute them to listening components as they're received**.

The "single Postgres connection per process" piece is key. Use of a notifier keeps the number of Postgres connections dedicated to use with listen/notify down to **one per program**, a major advantage compared to the naive version, which is _one connection per topic per program_. Especially for languages like Go that make a in-process concurrency easy and cheap, the notifier reduces listen/notify connection overhead to practically nil.

<img src="/assets/images/notifier/notifier.svg" alt="Notifier distributing notifications to program components">

## A few implementation details (#implementation)

From a conceptual standpoint, the notifier's not difficult to understand, and with only this high level description, most readers would be able to implement it themselves. I'm not going to go through an implementation in full detail, but let's look at a few important aspects of one. (For a complete reference, you can take a look [at River's notifier](https://github.com/riverqueue/river/tree/master/internal/notifier), which is quite well vetted.)

Here's a listen function to establish a new subscription:

``` go
// Listen returns a subscription that lets a caller receive values from a
// notification channel.
func (l *Notifier) Listen(channel string) *Subscription {
    l.mu.Lock()
    defer l.mu.Unlock()

    existingSubs := l.subscriptions[channel]

    sub := &Subscription{
        channel:        channel,
        listenChan:     make(chan string, 100),
        notifyListener: l,
    }
    l.subscriptions[channel] = append(existingSubs, sub)

    if len(existingSubs) > 0 {
        // If there's already another subscription for this channel, reuse its
        // established channel. It may already be closed (to indicate that the
        // connection is established), but that's okay.
        sub.establishedChan = existingSubs[0].establishedChan
        sub.establishedChanClose = func() {} // no op since not channel owner

        return sub
    }

    // The notifier will close this channel after it's successfully established
    // `LISTEN` for the given channel. Gives subscribers a way to confirm a
    // listen before moving on, which is especially useful in tests.
    sub.establishedChan = make(chan struct{})
    sub.establishedChanClose = sync.OnceFunc(func() { close(sub.establishedChan) })

    l.channelChanges = append(l.channelChanges,
        channelChange{channel, sub.establishedChanClose, channelChangeOperationListen})

    // Cancel out of blocking on WaitForNotification so changes can be processed
    // immediately.
    l.waitForNotificationCancel() 

    return sub
}
```

A few key details to notice:

* Subscriptions use a **buffered channel** like `make(chan string, 100)` and **non-blocking sends** (using `select` with `default`). A notifier may receive a high volume of notifications, and if it were to block on every component successfully receiving and processing each one, it could easily fall behind. Instead, a received notification is sent into the channel using a non-blocking send. The non-blocking send means that the send operation will never block: instead the notification is discarded if the channel is full. The buffer provides a tunable amount of slack to make sure this won't happen too easily. It's each component's job to make sure its processing its inbox in a timely manner. This is important because even in the event of one component falling behind, the system as a whole stays healthy.

* Multiple components may want to subscribe to the same topic. Since only one connection is in use, the notifier only needs to issue one `LISTEN` per topic. Internally, it organizes subscriptions by topic, and if it notices that a topic already exists, a new subscription is added without issuing `LISTEN`.

* Subscriptions provide an **established channel** that's closed when a `LISTEN` has been successfully issued and the notifier is up and listening. This isn't strictly necessary for most production uses, but it's invaluable for use in testing. If a test case issues `pg_notify` before the notifier has started listening, that notification is lost -- a problem that can lead to tortuous test intermittency [1]. Instead, a test case tells the notifier to listen, _waits for the listen to succeed_, then moves on to send `pg_notify`.

``` go
// EstablishedC is a channel that's closed after the notifier's successfully
// established a connection. This is especially useful in test cases, where it
// can be used to wait for confirmation that not only that the listener is
// started, but that it's successfully established started listening on a
// channel before continuing. For a new subscription on an already established
// channel, EstablishedC is already closed, so it's always safe to wait on it.
//
// There's no full guarantee that the notifier can ever successfully establish a
// listen, so callers will usually want to `select` on it combined with a
// context done, a stop channel, and/or a timeout.
//
// The channel is always closed as a notifier is stopping.
func (s *Subscription) EstablishedC() <-chan struct{} { return s.establishedChan }
```

### Interruptible receives (#interruptible-receives)

There's no standard SQL for waiting for a notification. Typically, it's accomplished using a special driver-level function like [Pgx's `WaitForNotification`](https://pkg.go.dev/github.com/jackc/pgx/v5#Conn.WaitForNotification).

These commonly block until receiving a notification, which can be problem since we're only using a single connection. What if the notifier is in a blocking receive loop, but another component wants to add a new subscription that requires `LISTEN` be issued?

You'll want to handle this case by making sure that the wait loop is interruptible. Here's one way to accomplish that in Go:

``` go
func (l *Notifier) runOnce(ctx context.Context) error {
    if err := l.processChannelChanges(ctx); err != nil {
        return err
    }

    // WaitForNotification is a blocking function, but since we want to wake
    // occasionally to process new `LISTEN`/`UNLISTEN` operations, we put a
    // context deadline on the listen, and as it expires don't treat it as an
    // error unless it
    notification, err := func() (*pgconn.Notification, error) {
        const listenTimeout = 30 * time.Second

        ctx, cancel := context.WithTimeout(ctx, listenTimeout)
        defer cancel()

        // Provides a way for the blocking wait to be cancelled in case a new
        // subscription change comes in.
        l.mu.Lock()
        l.waitForNotificationCancel = cancel
        l.mu.Unlock()

        notification, err := l.conn.WaitForNotification(ctx)
        if err != nil {
            return nil, xerrors.Errorf("error waiting for notification: %w", err)
        }

        return notification, nil
    }()
    if err != nil {
        // If the error was a cancellation or the deadline being exceeded but
        // there's no error in the parent context, return no error.
        if (errors.Is(err, context.Canceled) ||
            errors.Is(err, context.DeadlineExceeded)) && ctx.Err() == nil {
            return nil
        }

        return err
    }

    l.mu.RLock()
    defer l.mu.RUnlock()

    subs := l.subscriptions[notification.Channel]

    if len(subs) < 1 {
        return nil
    }

    for _, sub := range subs {
        sub.listenChan <- notification.Payload
    }

    return nil
}
```

The inner closure calls into `WaitForNotification`, but has a default context timeout of 30 seconds that automatically cycles the function periodically. It also stores the special context cancellation function `l.waitForNotificationCancel`.

When `Listen` is invoked and a new subscription needs to be added, `l.waitForNotificationCancel` is called. The wait is cancelled immediately, new subscriptions are processed, and the closure is reentered to wait anew.

### Let it crash (#let-it-crash)

Given there's now a single master connection that's handling all notifications for a program, it's fairly critical that its health be monitored, and the notifier reacts appropriately. If not, all uses of listen/notify would degrade simultaneously.

The obvious way to react would be to close the connection, use a connection pool to procure a new connection, reissue `LISTEN`s for each active subscription, then reenter the wait loop.

It can be a little tricky sometimes to guarantee that state is reset cleanly, so another possibility is to adhere to the "let it crash" school of thought. If the connection becomes irreconcilably unhealthy, stop the program, and have it come back to a healthy state by virtue of its normal start up.

``` go
// If the notifier gets unhealthy, restart the worker. This will generally
// never happen as the notifier has a built-in retry loop that try its best
// to keep established before giving up.
notifier.AddUnhealthyCallback(closeShutdown)
```

We've found this sort of edge to be so rare (I've only seen it happen once in a year+ of use) that letting the program crash when it does happen hasn't produced any undue disruption.

## PgBouncer (#pgbouncer)

Using [PgBouncer](https://www.pgbouncer.org/features.html), `LISTEN` is only supported using session pooling (as opposed to transaction pooling) because notifications are only sent to the original session that issued a `LISTEN` for them.

Use of a notifier requires an app to dedicate a single connection per program for listen/notify, but every other part of the application is free to use PgBouncer in transaction pooling or statement pooling mode, thereby maximizing the efficiency of connection use.

[1] Regarding test intermittency: Trust me on this. We found out the hard way so that you don't have to.
