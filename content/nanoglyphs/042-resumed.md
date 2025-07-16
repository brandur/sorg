+++
image_alt = "Bikers in Berlin next to Volkspark am Weinberg"
# image_orientation = "portrait"
image_url = "/photographs/nanoglyphs/042-resumed/berlin-bikes@2x.jpg"
published_at = 2025-07-15T09:17:39-07:00
title = "Service Resumed"
hook = "TODO"
+++

Readers —

I'm just recently from a few weeks in Berlin. Going for a music festival, but taking the opportunity to stay a few extra weeks in one of my favorite cities.

Before anything else, some usual housekeeping: I'm Brandur, and this is my newsletter _Nanoglyph_. It's been comically delinquent bordering on absurdism, with this being the first issue in over a year. You may have signed up months ago and this is the first blast you've gotten so far, If you hate it already, you can [instantly unsubscribe in one easy click](%unsubscribe_url%).

I've written this edition three times now, finding various excuses not to hit the "send" button each time. I've decided to try something shorter, hopefully to get some flow back. I'm predisposed to writing wordy communiqués the length of short novels, but in rare moments of reflection I remember that we all get a lot of email, and most of us probably don't make it past page one anyway.

This newsletter is [also available on the web](/nanoglyphs/042a-crush). Last year I spent some time [adding dark mode support](/fragments/dark-mode-notes), something that sounds easy until you realize you have a few thousand lines of legacy CSS to retrofit.

In this issue: don't mock the database, graceful request cancellation, occasionally injected clocks, and of course, Berlin.

---

<img src="/photographs/nanoglyphs/042-resumed/museum-island@2x.jpg" alt="Museum Island in Berlin" class="wide" loading="lazy">

## Berlin, a city alive (#berlin)

I first visited Berlin in 2011 and fell in love. Daunting, and a little foreboding. Underground bunkers. Unexpected art. The best bars on Earth. Sprawling and enormous, but dense, and with the city's U- and S-bahn systems providing perfect connectivity. I went on an assortment of walking tours, some where we read in on the thematic meanings of the city's prolific graffiti, and others where we jumped fences and explored abandoned factories. At the time it was December and cold, but that left an even greater impression on me, with the city's epic, cyclopean architecture looming out of the frigid nights. More romantic than anything produced in Hollywood.

These days when I visit, I take daily evening walks down through my favorite neighborhoods usually around Friedrichshain, Kreuzberg, and Neukolln. The difference between what I see in Berlin versus at home couldn't be more stark. Even after 9 PM, almost every street corner is alive. Some districts and more popular than others, but there are _hundreds_ of blocks where the vibrancy of life is incredible, with people out socializing and building community on patios, in parks, or perched on riverbanks. Contrast this to a city like San Francisco, where the only thing open passed 9 PM are the drug bazaars. There are less than half a dozen small areas in the entire city where you could find even a single soul outside around this time. The difference isn't 10%. It's not even 2x, 5x, or 10x. It's 1,000x. Every place on Earth should be trying to figure out what the hell exactly is going on in Berlin and trying to copy it.

The Berlin Wall fell in '89, which in 2011 would've been only twenty years before. The wall coming down was a seismic event---9.5 on the cultural Richter scale---leaving a historically depopulated city with a rare surplus of vacant real estate and land that led to the city's eclectic combination of abundant nightclubs, experimental art, and cheap rent.

Since then I've been coming back to the city every year or two, and it's been fascinating (and a little unnerving) watching it change during that time. Berlin is by far the oddest capital of Europe, but it's still a capital of Europe, and the further in time we move away from 1989, the more it normalizes to be like every other European capital, with the prices and tourists to go with it.

It'll be fascinating to see what the city looks like in another ten years. Until then, I'll return as often as I can.

<img src="/photographs/nanoglyphs/042-resumed/biergarten@2x.jpg" alt="Biergarten" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/042-resumed/gorlitzer@2x.jpg" alt="Gorlitzer Park in Berlin" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/042-resumed/berlin-bridge@2x.jpg" alt="A bridge in Berlin in the evening" class="wide" loading="lazy">

## 1 --- Database fixtures are fast enough (#1-database-fixtures)

On the Crunchy blog I published [Don't mock the database: Data fixtures are parallel safe, and plenty fast](https://www.crunchydata.com/blog/dont-mock-the-database-data-fixtures-are-parallel-safe-and-plenty-fast), arguing that the database should _never_ be mocked.

Using our project as a sample empirical baseline, I added a module to measure test statistics to try and quantify the runtime cost of hitting the database for real (the most commonly seen argument against doing so is that it's too slow). Our test suite runs 4,900 tests in ~23s uncached on my commodity notebook:

``` sh
$ gotestsum ./... -- -count=1
✓  apiendpoint (235ms)
✓  apierror (370ms)
✓  apiexample (483ms)
...
✓  util/urlutil (1.058s)
✓  util/uuidutil (1.084s)
✓  validate (1.077s)

DONE 4876 tests, 4 skipped in 23.156s
```

During the course of one run, we generate a little north of 18,000 fixtures:

``` sql
=# select * from test_stat;
                  id                  |          created_at           | num_fixtures
--------------------------------------+-------------------------------+--------------
 9E06C8B9-EA6E-490F-A0D3-1A18310376CF | 2025-05-28 07:42:49.500298-07 |        18132
```

18k fixtures / 23 seconds = 780 fixtures/s, although the real number is higher because not all tests are hitting the database. This seems to me to be plenty fast enough.

## 2 --- Graceful request cancellation (#2-graceful-cancellation)

I verified that our API stack [gracefully supports request cancellation](https://brandur.org/fragments/testing-request-cancellation).

If a request is cancelled midway through, either because the client did so explicitly (e.g. due to a timeout) or a connection was unceremoniously interrupted, it's nice to be able to stop executing immediately instead of continuing to serve an HTTP request whose response will never be received.

That sounds simple enough, but I'd bet money it's a trick that the vast majority of applications in production today can't handle. For most of them it wouldn't even be safe an interrupted request would leave mutated data in a partial state. The use of a transaction is critical to make sure that it's all or nothing.

I found that with Go's built in support for contexts (and our total use of transactions), it was pretty straightforward to get graceful cancellation handling working properly:

``` go
// API service handler error handling. Repeated from above.
ret, err := e.serviceHandler(ctx, req)
if err != nil {
    if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
        // Distinct error message when the request itself was
        // canceled above the API stack versus we had a
        // cancellation/timeout occur within the API endpoint.
        if r.Context().Err() != nil {
            // This is a non-standard status code (499), but
            // fairly widespread because Nginx defined it.
            err = apierror.NewClientClosedRequestError(ctx, errMessageRequestCanceled).WithSpecifics(ctx, err)
        } else {
            err = apierror.NewRequestTimeoutError(ctx, errMessageRequestTimeout).WithSpecifics(ctx, err)
        }
    }

    WriteError(ctx, w, err)
    return
}
```

We can test the feature by interrupting a request on an endpoint that sleeps artificially:

``` sh
$ curl -i http://localhost:5222/sleep
^C
```

The client will never see a response status code, but for logging purposes we return a `499 Client Closed Request` which isn't in any spec, but defined informally by Nginx years ago:

``` txt
canonical_api_line GET /sleep -> 499 (4.162702459s)
    api_error_cause="context canceled"
    api_error_internal_code=client_closed_request
    api_error_message="Context of incoming request canceled; API endpoint stopped executing."
```

## 3 --- Occasionally injected clocks (#3-postgres-clocks)

A trick almost too small to mention: [Occasionally injected clocks in Postgres](https://brandur.org/fragments/postgres-clocks).

An app deployed to many nodes will show minor variance in generating times as clocks are subtly skewed across the fleet. Usually it doesn't matter, but a low touch way to synchronize clock times is to general timestamps using a single, shared source of truth. A single, shared source of truth like a database.

But using _only_ a database can be problematic for the test suite because times in databases are hard to stub.

We've had reasonable luck in getting the best of both worlds by using `coalesce` along with a nullable parameter. When the clock is stubbed, the input parameter gets a value and the database operation uses that, but falls back to `now()` when it's not:

``` sql
-- name: QueuePause :execrows
UPDATE queue
SET paused_at = CASE
                WHEN paused_at IS NULL THEN coalesce(
                    sqlc.narg('now')::timestamptz,
                    now()
                )
                ELSE paused_at
                END
WHERE name = @name;
```

And we use it with a `NowUTCOrNil()` function that returns a value in case the clock is stubbed, or `nil` otherwise:

``` go
func (c *Client[TTx]) QueuePauseTx(ctx context.Context, tx TTx, name string, opts *QueuePauseOpts) error {
    executorTx := c.driver.UnwrapExecutor(tx)

    if err := executorTx.QueuePause(ctx, &QueuePauseParams{
        Name:   name,
        Now:    c.baseService.Time.NowUTCOrNil(), // <-- accessed here
        Schema: c.config.Schema,
    }); err != nil {
        return err
    }
    
    ...
```

Take this one as a loose take-it-or-leave-it recommendation. Remember that in SQL/Postgres, `CURRENT_TIMESTAMP`/`now()` returns the current time _at the beginning of the current transaction_ rather than the current time. This is sometimes a feature because every new record gets the same timestamp, but given transactions can be long-lived, this behavior is desirable or undesirable in roughly equal proportion.

<img src="/photographs/nanoglyphs/042-resumed/berlin-graffiti@2x.jpg" alt="Berlin graffiti" class="wide" loading="lazy">

## Acquired (#acquired)

In early June, we were acquired: [Snowflake to Buy Crunchy Data for $250 Million](https://archive.is/60MjY).

I got quite a few congratulatory texts after the event. I appreciated them all, but there was a sizable distance between their tone and my observations about the event to date. A friend of mine put it best, pairing two great virtues in being pithy and correct, messaging:

> Congrats?

We're entering a new future with a lot of uncertainty. The bull case is that we're able to leverage Snowflake's enormous portfolio to bring Postgres to a wider base than we'd otherwise be able to reach, including large enterprise users who'd otherwise be naturally predisposed to use whichever database has the biggest and best outfitted sales team (Oracle? Mongo?). There's a larger battle going on between Postgres and MySQL for the hearts and minds of future app developers, and assuming we like nice things like partial indexes and `RETURNING` syntax, we better hope that Postgres wins, even without Vitess.

I made a very modest sum from the sale. That's good I guess. But now, our small, agile team has been replaced by a much larger vessel with the rough maneuverability of an aircraft carrier. One where it takes a "secure" endpoint, password, and three multi-factor confirmations to use the head. It'll probably take some time to see what our new flow looks like, and to sort through what's been gained and what's been lost.

Until next week.

<img src="/photographs/nanoglyphs/042-resumed/berlin-portal@2x.jpg" alt="Berlin portal" class="wide_portrait" loading="lazy">