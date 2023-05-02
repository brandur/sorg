+++
image_alt = "Leafy seadrgon at the Georgia Aquarium"
image_url = "/photographs/nanoglyphs/036-queues/leafy-seadragon@2x.jpg"
published_at = 2023-05-01T08:35:17+02:00
title = "Atlanta, Job Queues, Batch-wise Operations"
+++

I'm back from Atlanta, and now in Paris.

This was the first time I'd ever been to Atlanta, or anywhere in  Georgia for that matter. Trying to pick a hotel, I pulled up Google Maps as I usually do to cross-reference the locations of hotels with those of parks and other areas of interest, only to find to my dismay that Atlanta appeared to have precious little in the way of either.

A lot of cities are shaped around a major geographic feature like a waterfront (San Francisco, Vancouver), river (London, Paris, Chicago, Calgary, countless others), or large park (Manhattan, Berlin). Downtown Atlanta is shaped around a highway. And the only green to be found in the vicinity are a large stadium (a park on technicality), pseudo-park to commemorate the '96 Olympics, and a little space next door to hold World of Coca-Cola, a theme park built around everybody's favorite sugary confection Coca-Cola, which yes, is apparently really a thing.

But it wasn't as bad as the maps made it look. Knowing nothing about Georgia, I was surprised as I stepped off the plane to find the area surrounded by tall, stately trees, made up of magnolias, dogwoods, Southern pines, and oaks. One of Atlanta's nicknames is "city in a forest", and although it'd be a stretch to apply that alias to downtown, it shows through in the surrounding area.

The third largest aquarium in the world (and formerly, largest) is in Atlanta, and its signature feature is a 6.3 million gallon tank holding a plethora of aquatic life including whale sharks, which I wrote about [in sequences 046](/sequences/046). Its extensive collection contains everything from hammerheads, dolphins, and belugas, to the mesmerizingly gorgeous [leafy seadragons](https://en.wikipedia.org/wiki/Leafy_seadragon) pictured above.

By the end of the trip I had a daily morning run that took me passed Centennial Park, around the aquarium and World of Coca-Coca, and down through the campus of Georgia Tech (which is very pretty). Thoroughly enjoyable.

---

## You can't take the name out of the conference, but you can take the conference out of the name (#railsconf)

The ostensible reason for being in Atlanta was RailsConf, of which I've grown more skeptical over the years.

Its organizing committee took a hard ideological turn a while back, and although the effects of such are long  the long effects of it are starting to show. Infamously, last year DHH, the creator of Rails, was [uninvited from the conference](https://world.hey.com/dhh/no-railsconf-faa7935e) for making _shocking_ and _offensive_ public statements, like that people at work should do work instead of fighting with each other about politics. Absolutely outrageous -- he deserved worse. Also in 2022, RubyConf (also organized by Ruby Central) announced they were giving up a $500k deposit so they could move the conference from Houston to San Diego, because as we all know, the state of Texas and all 30 million souls that call it home, are evil. Money well spent. Naturally, vaccine passports were still in full effect at RailsConf in mid-2023 -- not for anything to do with Covid control, but to make sure that no undesirables guilty of wrongthink could be in attendance.

Programming-wise, the subject matter has shifted away from the technical towards an emphasis on soft skills. Common themes were mentorship, job progression, and unionization, which might all be fine, except for those confused few who might've come to RailsConf to learn about Rails, there wasn't much of an alternate tech track to opt into instead. (This is generally true but not completely -- my commendations to the couple great technical talks that snuck in past the review board.)

The name "RailsConf" is an institution, but time may be demonstrating that it's only possible for so long to run a successful conference that is ostensibly about one subject, but which is largely in pursuit of unrelated objectives. The exhibitor floor this year was pared down again even compared to last, which had already been modest. The crowds were sparse, the yawning expanse of the job boards unnervingly empty, and one couldn't help but get the feeling that the echoing halls of the Atlanta venue were oversized compared to the capacity necessary for the number of people there.

Ruby/Rails may not be the most cutting edge technology anymore, but it's still its own very interesting little niche of a broader ecosystem, and it's a shame not to have a more vibrant community that cares about it. Last November, the [Rails Foundation was created](https://rubyonrails.org/2022/11/14/the-rails-foundation) with a seed of one million dollars from its eight founding members: Cookpad, Doximity, Fleetio, GitHub, Intercom, Procore, Shopify, and 37Signals. Later this year it'll hold [the first ever Rails World](https://rubyonrails.org/2023/4/6/rails-world-is-coming) in Amsterdam, limited to 650 tickets this year, but presumably aiming to provide a more focused (and dare I say, more inclusive) alternative to the Ruby Central encumbents.

I'm cautiously optimistic about the new event. It could just be the phoenix-from-the-ashes resurrection that the Rails community needs.

---

## Job queues and the hallway track (#job-queues)

As with most conferences, the best track was the hallway track, and I could some good discussions with a few people, notably around job queues (for which some selection bias might've been present). A few recent links on the subject:

* [Building a simple dynamic job scheduler with Sidekiq](https://www.tines.com/blog/building-a-simple-dynamic-job-scheduler-with-sidekiq), in which it was found that a stateless scheduler wasn't up to the task due to the possibility of it crashing, missing the hour in which it was supposed to work, and losing scheduled jobs. The solution was to move to a Postgres-based scheduler.

* [Why we ditched RabbitMQ and replaced it with a postgres queue](https://www.prequel.co/blog/sql-maxis-why-we-ditched-rabbitmq-and-replaced-it-with-a-postgres-queue), in which RabbitMQ's job prefetching feature, assuming jobs to all be short-lived, ended up holding jobs hostage for hours as they were stuck behind a slow job at the worker's head.

I got a chance to catch up with someone who worked the Heroku API team after my own time under the mast, and asked about its job queue.

During my tenure the project had moved from Delayed Job to Queue Classic to [Que](https://github.com/que-rb/que) as we tracked the new hotness in the ecosystem. Having a job queue in Postgres was interesting for minimizing dependencies and the possibilities around transaction consistency, but the years before I left had been plagued by [major queue bloat problems](/postgres-queues) caused by long-running queries initiated by other teams, filling the queue's index with dead tuples and making it inordinately expensive for workers to find live jobs. We'd found a tentative workaround that worked reasonably well, but the whole scheme was calling out to be rearchitected.

## Why the queue matters (#why)

It might not be intuitively obvious why a job queue and its steady health were of critical importance, but they were. A primary function of the queue was to push slower operations like network calls out of the synchronous API path, thereby keeping the API responsiveness while (in theory anyway) not taking any hit to UX.

One example of an out-of-band call were to register domains for new Heroku apps, linked back to the Heroku router. Every app creation would get a background job that'd perform:

```
POST https://maestro.heroku.com/domains/serene-example-4269
```

A struggling queue meant delayed domain registrations, which meant that new apps might not be able to serve requests for many minutes after being created. Similar delays would occur for activating new app releases and deprovisioning add-ons.

<img src="/photographs/nanoglyphs/036-queues/leafy-seadragon-2@2x.jpg" alt="Another shot of the leafy seadragon at the Georgia Aquarium" class="wide" loading="lazy">

## Batch-wise operations (#batch-wise)

Heroku's queueing problems weren't so technically challenging that they couldn't be fixed, but these days I find myself wondering whether we couldn't have made life easier for ourselves by shifting the paradigm.

This style of high volume queue where jobs are queued frequently and expected to be worked quickly works okay when all systems are green, but is anti-fragile in that it degenerates badly when they're not. Index bloating aside, even when we resolved the immediate problem by killing a long-running transaction, a large backlog had been enqueued, and since jobs are worked one by one, it took time to work it back down to zero.

The service I operate at Crunchy doesn't have a job queue. Instead, we have single, dedicated worker goroutines that fulfill specific async tasks. This might otherwise be scary because it'd be easy for a single goroutine to fall behind in the presence of a large amount of work, but each one's been specifically written to operate efficiently by batched work into as few operations as possible, and then running what can't be batched in parallel. Concretely:

* Selects are all done together instead using `id = any(@id::uuid[])` to avoid chatty round trips to the database that could easily number in the thousands.

* Updates and inserts are batched using the normal `INSERT INTO tbl VALUES ('a', 1), ('b', 2), ('c, 3)` or upsert. We use ULIDs as IDs in most places for [faster and more index-friendly insertion](/nanoglyphs/026-ids).

* When invoking remote APIs we'd use batch APIs where provided. Batch APIs are so rare as to be practically non-existent though, so we use [`errgroup` as a worker pool](https://pkg.go.dev/golang.org/x/sync/errgroup) to crunch through these invocations in parallel.

Selecting a batch is always constrained by a `LIMIT` clause so that in case of an enormous large backlog, we can always make sure that we're making forward progress instead of timing out trying to select or update a degenerately large data set.

``` sql
-- name: EmailNotificationGetUnsent :many
SELECT *
FROM email_notification
WHERE email_to IS NOT NULL
    AND sent_at IS NULL
ORDER BY id
LIMIT @max;
```

Workers go to sleep between loops if they don't have anything to do, but we have a framework for waking them via `pg_notify` to tell them there's work avaialble. So if this had been in use for the Heroku case, the API request that created a new app would wake the DNS worker to make sure that new records were registered in a timely matter.

### Example: Sending mail (#sending-mail)

Here's some sample code for our worker that sends mail. An initial select pulls up to 100 email notifications that are ready to go out:

``` go
const emailNotificationSenderMaxPerRun = 100

notifications, err := queries.EmailNotificationGetUnsent(ctx,
    emailNotificationSenderMaxPerRun)
if err != nil {
    return nil, xerrors.Errorf("error getting unsent notifications: %w", err)
}
```

Then select other data that'll be needed to send everything in the batch:

``` go
notificationAttachments, err := queries.EmailNotificationAttachmentGetByEmailNotificationID(ctx,
    sliceutil.Map(notifications, func(n dbsqlc.EmailNotification) uuid.UUID { return n.ID }))
if err != nil {
    return nil, xerrors.Errorf("error getting email notification attachments: %w", err)
}

// email notification ID -> slice of attachments
notificationAttachmentsMap := make(map[uuid.UUID][]dbsqlc.EmailNotificationAttachment)
```

Email notifications may have attachments, so use an `errgroup` to pull all the ones we'll need down from S3:

``` go
errGroup, ctx := errgroup.WithContext(ctx)
errGroup.SetLimit(awsclient.MaxParallelRequests)

for _, attachment := range notificationAttachments {
    ...

    errGroup.Go(func() error {
        out, err := w.awsClient.S3_GetObject(ctx, &s3.GetObjectInput{
            Bucket: ptrutil.Ptr(awsclient.S3Bucket),
            Key:    ptrutil.Ptr(attachmentS3Key),
        })
        if err != nil {
            return xerrors.Errorf("error getting attachment with key %q from S3: %w",
                attachmentS3Key, err)
        }
```

And similary, send each notification up to Mailgun in a parallel `errgroup`. Results are _not_ updated as a batch in this case because Mailgun's API offers no way of guaranteeing idempotency, and we want to be as careful as possible to avoid double-sending mail:

``` go
for _, notification := range notifications {
    ...

    errGroup.Go(func() error {
        message := w.mailgunClient.NewMessage(notification.EmailFrom, notification.Subject,
            notification.BodyHTML, notification.BodyText, notification.EmailTo)

        if attachments, ok := notificationAttachmentsMap[notification.ID]; ok {
            for _, attachment := range attachments {
                message.AddBufferAttachment(attachment.Filename, attachmentsDataMap[attachment.S3Key])
            }
        }

        _, id, err := w.mailgunClient.Send(ctx, message)
        if err != nil {
            return xerrors.Errorf("error sending message: %w", err)
        }

        _, err = queries.EmailNotificationUpdate(ctx, dbsqlc.EmailNotificationUpdateParams{
            ID:                       notification.ID,
            MailgunMessageIDDoUpdate: true,
            MailgunMessageID:         sql.NullString{String: id, Valid: true},
            SentAtDoUpdate:           true,
            SentAt:                   sql.NullTime{Time: time.Now(), Valid: true},
        })
        if err != nil {
            return xerrors.Errorf("error updating notification: %w", err)
        }
```

API clients should be configured with reasonably aggressive timeouts so that a single slow invocation can't delay an entire batch from finishing.

I'd never claim that anything in software is a panacea, but am fairly confident that between batch-wise operations and Go's speed, most of our workers wouldn't have much trouble burning down backlogs even in the millions if it ever came to that.

## Queues, like everything, ossify (#queue-ossification)

Back to meeting another ex-Heroku employee during the hallway track: naturally, I was excited to hear about what had come next for the Heroku API job queue. An upgrade to the newer and more bloat-resistant version of Que? Sidekiq combined with a [transactional job drain](/job-drain)? A total shift away from high-throughput single-function work and over to more efficient batch-wise operations?

But I'd forgotten my own theses around architectural longevity and [the Lindy effect](https://en.wikipedia.org/wiki/Lindy_effect):

> by which the future life expectancy of some non-perishable things, like a technology or an idea, is proportional to their current age. Thus, the Lindy effect proposes the longer a period something has survived to exist or be used in the present, the longer its remaining life expectancy.

Due to a combination of chronic trouble with headcount and overwhelming weight of sheer inertia, that job queue never changed, becoming yet another case study for the [the disproportionate influence of early tech decisions](/fragments/early-tech-decisions).

---

I'm in Europe for the whole month of May, making my way from Paris to London to Berlin.

I'm off to a good start, having totally forgotten over the last three years that outlets over here are not the same as outlets in North America, left my Euro-to-US outlet dongle at home, and not being able to charge any of the myriad of gizmos that I brought along with me. This morning I begin a pilgrimage towards Notre-Dame in search of a fabled site called "FNAC" which may be able to sell me one at a price that doesn't make me wish I'd never left my own continent, and on French Labour Day no less.

Until next week.