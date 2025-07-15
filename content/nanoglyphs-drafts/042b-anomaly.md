+++
image_alt = "TODO"
# image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/042-anomaly/downtown-reno@2x.jpg"
published_at = 2025-03-26T10:40:47-07:00
title = "Transaction Anomaly"
hook = "TODO"
+++

Readers ---

I've been operating production systems going on fifteen years now and talking about databases for almost as long. It wasn't until a few weeks ago that I saw my first ever honest-to-god transaction anomaly in production.

Transaction anomalies are a class of problem that database academics talk about all the time, and which production operators never do. They probably do happen quite a lot in the real world, but it doesn't matter much that they do. Just one more rare, intermittent error to go onto the Sentry/Honeybadger heap.

Before anything else, some usual housekeeping: I'm Brandur, and this is my newsletter _Nanoglyph_. It's been comically delinquent for over a year now, and you may have signed up months ago and this is the first blast you've gotten so far, If you hate it already and don't want to see anymore of them, you can [instantly unsubscribe in one easy click](%unsubscribe_url%).

This newsletter is [also available on the web](/nanoglyphs/043-anomaly). Last year I spent some time [adding dark mode support](/fragments/dark-mode-notes), something that sounds easy until you realize you have a few thousand lines of legacy CSS to retrofit.

---

## In pursuit of Sentry zero (#sentry-zero)

I go through our Sentry on a regular basis looking for bugs that might be affecting real users. I'd love to achieve "Sentry zero" one day, but realistically it doesn't happen because there's always a small collection of intermittent exceptions (context cancellations, statement timeouts) that'll never be banished completely.

Beyond those though, we usually do have zero. A few new ones a week generally appear, but we'll fix them. Here's one of those that appeared:

``` txt
error deleting access group account "01956c80-9297-7754-69d5-e1fefcce5e02" (access group ID "01955d36-68e9-2b93-9e8c-878a8367770b", account ID "5dacc804-ca3a-436e-b826-a8b8b8998789"):
    no rows in result set
  File "/tmp/build_f0358c7e/pcommandkind/access_group_account_destroy.go", line 108, in (*AccessGroupAccountDestroy).Run
    return nil, xerrors.Errorf("error deleting access group account %q (access group ID %q, account ID %q): %w",
  File "/tmp/build_f0358c7e/server/api/access_group_account_service.go", line 76, in (*accessGroupAccountService).Delete.func1
    destroyRes, err := pcommandkind.NewAccessGroupAccountDestroy(svc.BaseParams).Run(ctx, tx, &pcommandkind.AccessGroupAccountDestroyParams{
```

TODO
https://crunchy-data.sentry.io/issues/6286172207/

I studiously examined the code path, expecting to find of the usual subtle, but pretty-obvious-once-you-see-it code bugs that's generally the cause of this type of thing.

But I couldn't see it. Early in the API request, we'd load an access group account (a join record that ties an access group to an account). The object was passed into a subcommand that'd run a few more queries, then go to delete it. But by that point, it was gone!

TODO code

``` go
// access group loaded
var accessGroupAccount *dbsqlc.AccessGroupAccount
if err := dbload.LoadOne(ctx, tx, dbload.LoaderCustomID(
    &accessGroupAccount,
    dbsqlc.AccessGroupAndAccount(req.AccessGroupID, req.AccountID)
), nil); err != nil {
    return nil, err
}
```

``` go
// but gone by the time we try to delete it!
deletedAccessGroupAccount, err := dbsqlc.New().AccessGroupAccountDeleteByID(ctx, e, params.AccessGroupAccount.ID)
if err != nil {
    return nil, xerrors.Errorf("error deleting access group account %q (access group ID %q, account ID %q): %w",
        params.AccessGroupAccount.ID, params.AccessGroupAccount.AccessGroupID, params.AccessGroupAccount.AccountID, err)
}
```

(An "access group account" is a join relation between an account and an access group, which is like a custom team with configurable permissions, membership, and permissions, but that's not important for this example. It could just as easily have been any object.)

On a transaction-less Mongo stack like Stripe's, this would've been no surprise. Without snapshot isolation or foreign keys, all data belongs to Schrodinger. It may be there, it may not. Who knows! Just make sure to write if conditions around literally every statement.

But here, the entire API request was protected by a transaction, which means this isn't supposed to happen. It better not at any rate -- one of the beautiful things about code written on ACID is you get to make more assumptions safely. More safe assumptions leads to simpler code, and we've got a hell of a lot of this "simple" code. If I'd made a fundamental mistake somewhere, it could mean that this one bug is actually a hundred separate bugs repeated over and over again.

<img src="/photographs/nanoglyphs/042-anomaly/bridge@2x.jpg" alt="Bridge near downtown Reno" class="wide" loading="lazy">

---

## Repeatable read (#repeatable-read)

After not finding any obvious code mistakes, I begrudgingly went back to the docs to check my assumptions. Anyone who's read Postgres' [transaction isolation](https://www.postgresql.org/docs/current/transaction-iso.html) will be eminently familiar with this table, which has been an unchanging staple of the page for two decades:

## TODO (#todo)

+ Isolation level + Dirty Read + Nonrepeatable read + Phantom read + Serialization anomaly

Pay particular attention to the definition of a nonrepeatable read:

> A transaction re-reads data it has previously read and finds that data has been modified by another transaction (that committed since the initial read).

Postgres' default isolation level is **Read Committed**, and as I was re-reading this in the docs, I found the answer to my problem right there in the last sentence, clear as day:

> Read Committed is the default isolation level in PostgreSQL. When a transaction uses this isolation level, a `SELECT` query (without a `FOR UPDATE`/`SHARE` clause) sees only data committed before the query began; it never sees either uncommitted data or changes committed by concurrent transactions during the query's execution. In effect, a `SELECT` query sees a snapshot of the database as of the instant the query begins to run. However, `SELECT` does see the effects of previous updates executed within its own transaction, even though they are not yet committed. **Also note that two successive `SELECT` commands can see different data, even though they are within a single transaction, if other transactions commit changes after the first `SELECT` starts and before the second `SELECT` starts.**

Transaction isolation provides many guarantees arounds data integrity, but it's not absolute, and especially not at every isolation level.

## *Too* parallel (#too-parallel)

There's a widget in our UI that's a list of checkboxes that let an administrator add or remove a team member from each access group [1] on a team. A user may select or deselect any number of the checkboxes, and submit them all at once.

For a special set of system access groups, being added to one implies removal from another. For example, if a team member's added to the "admin" access group, they're removed from the "member" access group by the system automatically, even if they hadn't explicitly deselected that box in the UI.

There wasn't a batch API for multiple access group additions/removals, so on submit, our Dashboard was dispatching multiple separate API requests to converge on the desired state. It's written in Typescript, so it'd model those requests as promises whose results would be all be waited on at once, letting the fast JS runtime issue them all in near simultaneity.

Here's two sample API requests of an access group removal and access group addition causing _implicit_ removal on the same record. They happened so close together that our canonical log upserter batched them into the same transaction, accidentally assigning them exactly the same `created_at` down to the microsecond because `CURRENT_TIMESTAMP` stays constant for the transaction's duration:

``` sql
=> SELECT created_at, http_method, http_path, status
   FROM canonical_api_line
   WHERE request_id = '223e092a-6bb4-ce74-59c1-29b88dded96f'
      OR request_id = '1ef5ff29-9c01-69b8-a44d-6a2ffce63b93';
          created_at           | http_method | status |                                   http_path                                   
-------------------------------+-------------+--------+-------------------------------------------------------------------------------
 2025-03-06 17:28:58.122762+00 | PUT         |    200 | /access-groups/agkv2ntbdelkpb6o4yjbehuc6i/accounts/lwwmqbgkhjbw5obgvc4lrgmhre
 2025-03-06 17:28:58.122762+00 | DELETE      |    500 | /access-groups/agkv2nti5evzhhumq6figz3xbm/accounts/lwwmqbgkhjbw5obgvc4lrgmhre
```

So one request would remove the database record, and the other would also remove the same record. The failing request saw the record initially on start, but only a few milliseconds later experienced a nonrepeatable read as the other request beat it to the punch. 

## Are nonrepeatable reads ... everywhere? (#nonrepeatable-reads)

Database textbooks make implications that'd have you believe that transaction anomalies are a constant companion to database users, but it's more like the opposite. I've been doing this a long time, and this is the first instance of a definitive, provable transaction anomaly in production that I recall seeing.

But to be fair, a lot of bugs of this shape are prevented by accident:

* Most requests are orthogonal to each other, manipulating unrelated data. (i.e. Different authenticated users that have their own independent data sets, requests within data sets are few enough that they manipulate independent data most of the time).

* Most software is pretty slow. If our client wasn't extensively built on promises all the way through, those two conflicting API requests wouldn't have been so parallel, and wouldn't have triggered the bug. One of them would still have 404ed unexpectedly, but I'm not sure anyone would've even noticed.

That said, thousands of these errors are surely happening daily across the planet, with most of them going unnoticed.

<img src="/photographs/nanoglyphs/042-anomaly/quad@2x.jpg" alt="Quad at University of Nevada" class="wide" loading="lazy">

---

## Potential fixes (#potential-fixes)

With the cause nailed down, there's a few plausible ways to go about fixing this. We'll go from hardest to easiest.

### Cranking isolation level (#isolation-level)

Another possibility would be to tighten the transaction's isolation level from **read committed** to **repeatable read**, which was the name would suggest, makes nonrepeatable reads impossible:

``` sql
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
```

Vetting this possibility was beyond the scope of work I was willing to do in that moment. More transaction guarantees means more bookkeeping for Postgres, and tightening the isolation level might lead to novel performance degradations that we'd have to look into. Unlike read committed, going to repeatable read would also [necessitate writing code that's smart enough to retry transactions](https://www.postgresql.org/docs/current/transaction-iso.html) in case of failure. That's doable, but would involve some major work in the depths of our program's plumbing.

### Select for update (#select-for-update)

The first is to make sure to lock the row as it's being read the first time using `SELECT ... FOR UPDATE` [locking clause](https://www.postgresql.org/docs/current/sql-select.html#SQL-FOR-UPDATE-SHARE):

``` sql
-- name: AccessGroupAccountGetByID :one
SELECT *
FROM access_group_account
WHERE id = @id
FOR UPDATE;
```

`SELECT ... FOR UPDATE` takes a lock on the row that other transactions would have to wait on before operating on it. In our case both transactions finish on the order of milliseconds, so the blocking duration would be practically unnoticeable, and it'd all resolve just fine.

### Recover on error (#recover-on-error)

Those both sound like pretty good solutions. Too good maybe. As a slightly strapped for time developer trying to get this bug knocked out, what did we do instead? Well, something a little, erm, simpler for the time being:

``` go
deletedAccessGroupAccount, err := dbsqlc.New().AccessGroupAccountDeleteByID(ctx, e, params.AccessGroupAccount.ID)
// Tolerate a missing access group account by the time we try to delete it,
// despite having already loaded the record for this operation.
if err != nil && !errors.Is(err, db.ErrNoRows) {
    return nil, xerrors.Errorf("error deleting access group account: %w", err)
}
```

The `SELECT ... FOR UPDATE` approach strike me as the most obviously correct and sustainable approach, and one that we should be able to bake nicely into our data loading layer, so at some point soon I'll go back through and see if I can work it in. To be continued.

<img src="/photographs/nanoglyphs/042-anomaly/truckee@2x.jpg" alt="Path along the Truckee" class="wide" loading="lazy">

## Reno (#reno)

Today's photos are from Reno. Always on the look out for nice places to live, earlier this year I visited Reno for the first time. Flying into its airport, the first thing you notice is how absolutely gorgeous the city is. Growing up in Calgary, I got used to being able to look west and see an absolutely picturesque view of the snow-tipped rockies out in the distance, a constant reminder of Alberta's famed natural abundance of parks, mountains, and forests.

But where in Calgary those mountains are far, in Reno they're near. Descending from the air, you see what seems to be what's on the smaller side for a city, walled in all sides by perfect white-tipped peaks that are practically right on top of it. Besides the mountains, your immediate impression is that for a town of its size Reno seems to have an outsided number of high rises, interspersed with some major oddities like a giant white dome at the center of town. The city's tagline is "the biggest little city in the world", which is quite apt.

I picked up a VRBO for the week, and things started well when I got a call from its manager within an hour of when I landed to say that the unit three floors up had started to leak, and water was coming down through the walls, putting everything in danger of rot. Workers would need access to bring in some industrial strength fans and air conditioners to dry the place out (I lived with the fans for about a week, but they didn't turn out to be a major problem, it was kind of a fun experience by the end of it).

Among other things, I discovered that as it has in many other places, third wave coffee has spread from the Bay Area out to Reno, and although that part is great, no so amazing are the $15 for a coffee + pastry prices that didn't escape the contagion. The city isn't a major site for snow (for that, you go east into Tahoe), but I got lucky and had a few hours worth of a gorgeous white display that enveloped the town.

<img src="/photographs/nanoglyphs/042-anomaly/snow@2x.jpg" alt="Snow over the Truckee" class="wide" loading="lazy">

Reno's a nice city, but navigating around it you can't help but see the bones of what could've been a truly _great_ city. It's split down the middle by the Truckee river, and like any good municipal planners would do, there's been an obvious effort to plumb a series of walking and bike paths up and down it. It's a good idea, but foiled by the seeming difficulty in acquiring land for the task. You'll be following a path through a nice park when suddenly it stops and is replaced by a road with no good way across. Or the park abruptly ends, expelling you out into a subdivision. Parkland generally only occupies one side of the river at any given time, and bridges across the Truckee are few and far between. The result is less of a green belt, and more of a series of discontinuous green spots.

Still, low taxes, a decent airport, a well run state, far enough north to avoid Las Vegas' climate, and pretty good ski access. There's a lot to like about this place.

## What happened here? (#what-happened-here)

For a delinquent newsletter with long periods between sends, it's impossible not to notice that this is by far its worst period of delinquency with its long_est_ period between sends. I don't have much in the way of an excuse, except that this is the third time over the space of a year I've written issue 042, and in the previous two times I came right up to the finish line, and just didn't hit "send". Why? Don't know.

It's not lost on me that every one of these is way too much of a production. Not only is a longer edition more work for me to write, but it's more work for people to read. I've noticed in recent years that most often even for newsletter which I receive and _like_, I don't get through the entire thing, and I expect that's true of other people too. For the next couple sends I'm going to experience with a shorter format thet's faster to write, and more likely to be read.

Until next week.

Brandur

[1] An "access group" in our parlance is a customizable security group offering a configurable set of permissions on a configurable set of clusters.

<img src="/photographs/nanoglyphs/042-anomaly/nevada-n@2x.jpg" alt="The Nevada N" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/042-anomaly/into-downtown-reno@2x.jpg" alt="View into downtown Reno" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/042-anomaly/rancho-san-rafael@2x.jpg" alt="View from Rancho San Rafael" class="wide" loading="lazy">