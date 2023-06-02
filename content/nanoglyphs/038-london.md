+++
image_alt = "Hyde Park in London"
image_url = "/photographs/nanoglyphs/038-london/hyde-park@2x.jpg"
published_at = 2023-06-02T10:08:07-07:00
title = "London, Gardening 500s"
+++

Having never been before, the first thing that occurred to me arriving in London is that London is the most _f*ing_ London place ever. Trite, yes, but what I mean is that the decades worth of London stereotypes in TV, movies, and photographs actually gave me a pretty good impression of what the city really would be like.

All the features I'd come to expect were really there: stately mansions (which in British English, refers to a grand looking apartment block rather than a large house), double decker buses, "mind the gap" announcements, fish and chips, red phone booths, black cabs, beige trench coats (Burberry or a Burberry-like), [British Whippets](https://en.wikipedia.org/wiki/Whippet), intense _Lock Stock and Two Smoking Barrels_ accents, brick-covered [mews](https://en.wikipedia.org/wiki/Mews), and of course, tea. I arrived a few weeks after the coronation of the King Charles III, and thousands of British flags were flying all over town.

With the possible exception of Tokyo, it's the busiest city I've ever step foot in. In New York when I go to Times Square, I think, "wow, so many people! it's crazy!", but then go to Brooklyn and find that it's not so bad. In London, I go Westminster, and think "wow, so many people! it's crazy!". Then I go to Mayfair and think "wow, so many people! it's crazy!". And then Kensington. And SOHO. And Camden Town. And Chelsea. And Marylebone. At least a dozen neighborhoods have hubs that are bustling with more activity than the single liveliest district found in most cities.

## Before sunrise (#before-sunrise)

London wakes early. Paris is also a busy city, but there's a trick: get up early and you may as well own it. I'd started to quite enjoy running at the crack of dawn and going by the Eiffel Tower and Palais de Chaillot, normally mobbed by tourists, but so empty around sunrise that you could use them as a private parade ground. I'd head down to the miniature Statue of Liberty in the Seine, then over to the Inception Bridge (does it have another name?), giving courtesy waves to the five other runners out that time of the morning, and even once witnessing a wedding ceremony held in public, but made 100% private by virtue being held on the south side of 7 AM. (Okay, witnessed by me and one other runner.)

London is different. Roll out a little late at 6:30 AM (when even in May, the city's northern latitude means that you have blinding daylight flooding down through your windows already), and you feel like a delinquent, the last person in town awake. Run over to Hyde Park, already mobbed with early morning enthusiasts, and spend the next 60 minutes jostling for position with other runners (see my screed on British walking conventions at the bottom), most of whom are just about done with their morning work out and will be dressed and at work before Big Ben strikes seven.

And anecdotally, what they say about oligarchs buying up London seems to be true? Some friends tell me about the local dating pool, where they'll meet guys who live on the water, drive Lamborghinis, and who claim their wealth stems from having sold some tech startups, but for whom googling their name offers no results beyond their social media profiles. In other times or other places we might have whispered "drug dealer", but now, more likely they're the children of oligarchs and other landed rich putting down roots in the business and cultural capital of Europe.

And yet another city where it's not clear how non-oligarchs, who average < £60k annually and aspire to one day purchase £750k+ one bedroom apartments, make ends meet. Some of the people I met who made salaries which would be considered reasonably respectable, literally lived in social housing (although that's apparently pretty nice here) they'd won their way into long ago, and now the only foothold keeping them in the city. A familiar pattern we're seeing repeated again and again across every major city in the western world (although London may be the most intense version of it) and which we all seem to accept as a completely intractable problem.

The good news is that if you're used to being driven crazy by the abject madness of Bay Area real estate, one of the few therapeutics available that might really work to feel better is to visit London.

<img src="/photographs/nanoglyphs/038-london/hyde-park-2@2x.jpg" alt="Hyde Park in London" class="wide" loading="lazy">

## Tending the garden of 500s (#500-garden)

I don't have much in the way of technical content this week, so let me pose an open question: what is the correct methodology, if any, for tending exceptions/500s [1]?

Many projects will already be using an exception tracker like Sentry or Honeybadger. In a perfect world, the number of exceptions in the tracker is of course, zero. But, as anyone who's ever run any real life project can tell you, we do not live in a perfect world.

<img src="/photographs/nanoglyphs/038-london/sentry-errors@2x.png" alt="" class="img_constrained" loading="lazy">

At Heroku, we'd use our tracker to inspect new errors, but we gave up on ever clearing it early on. There were some exceptions in there that I worked with longer than every one of my human colleagues. One in particular -- a unique constraint violation on app releases -- was there the day I joined, and there the day I left, having occurred tens of thousands of times in the interim [2].

In five years of working at Stripe, not one person I ever met even pretended to care about exception hygiene. The trackers weren't just littered with hundreds of thousands of exceptions, they were littered with hundreds of thousands of _kinds_ of exceptions -- a veritable encyclopedia of every novel things that could possibly go wrong in the application, and there were many. They'd occupy similar shelf real estate as the _Encyclopedia Britannia_ if printed out.

But this time around, seeing a possible opportunity with a smaller application constrained in size and scope, I tried to see if we could do it better. For every new bug that appeared, I'd jump to investigate it immediately, making sure to track it down and squash it definitively before moving on.

So far it's working, at least I think? We have zero exceptions that are the result of undiagnosed bugs. This is partly possible for us because of a strong ownership model. If I see a new exception, I know that if I don't fix it, no one else will. Most shops will inherently see [the bystander effect](https://en.wikipedia.org/wiki/Bystander_effect) in that if errors are going to a shared bucket, it's no specific person's job to go in and fix them, so they don't get fixed. Instead, issues are filed for particularly severe problems, which scales, but is slow and not thorough.

But zero exceptions from undiagnosed bugs doesn't mean zero exceptions. We regularly have a half dozen or so errors that appear as some of our core services like the database have little service blips and fail some requests. I could resolve all of them to make them disappear temporarily, but they'll be back inside of a week or two.

I wrote a prototype wherein a circuit breaker is added to the error reporting module to make some types of knowingly intermittent errors only be reported in case they pass a threshold over a given period (e.g. 10 exceptions within 30 seconds). So if we had a handful of Postgres errors they wouldn't be sent, but if we were to get dozens (suggesting a real problem with the database), the circuit breaker would open and they'd get sent up.

It was a creative solution, but the more I thought about it, the less liked it. Was I really going to add a big, complicated gear to the system just so I wouldn't have to see a couple exceptions in Sentry? An interesting idea, but the possible added trouble didn't feel commensurate with the value. I never sent it out.

So my question to you is: what is the right technique for managing intermittent exceptions? Is the right move to just to live with the fact that exception trackers are inherently unruly, and will never be brought to zero?

Speaking of exception trackers, kudos to Honeybadger for making the only swag t-shirt I've ever received that's actually going to enter my wear rotation (this from [RailsConf last month](/nanoglyphs/036-queues)).

<img src="/photographs/nanoglyphs/038-london/honeybadger-shirt@2x.jpg" alt="" class="img_constrained" loading="lazy">

<img src="/photographs/nanoglyphs/038-london/honeybadger-shirt-2@2x.jpg" alt="" class="img_constrained" loading="lazy">

It took me a moment to see it, but look carefully at what the honeybadger is battling against. Not insects, but _bugs_. A dramatic, but accurate metaphor for life as a service operator. Also, metal.

## Chaos under the crown (#chaos-under-the-crown)

Over the years I've read many long comment threads by British writers on the propensity of the country to "queue", so while uncivilized barbarians in other countries would elbow each other out of the way intent to be the rat that wins the race, His Majesty's subjects would line up in considerate, first-in-first-out order for everything from the bus to the ski lift, showing the rest of the world what true, high-minded social cohesion really looks like.

Coming to London, I was very much looking forward to observing this marvel of first-world civilization, ready to be embarrassed by my own medieval manners, but granted the rare opportunity to bear witness to this spectacle of great society, being totally okay with it.

But within five minutes of arriving I'd forgotten completely about queueing, and two weeks later as I leave the UK, it still hasn't reentered my mind. Disembarking the EuroStar, I was somewhat surprised to encounter disorder at the train station so total that it made what I'd just left at the busy Gare du Nord station in Paris look like a poster child for orderliness.

In most countries on Earth, by convention people tend to stay to one side or the other, either right or left, and even if it's not universally adhered to, keeping two broad streams of traffic works well enough for good general order. In the UK, there is only one convention: Absolute. Total. Chaos. Everybody moves in a random direction, at a random angle, along a random side of the path, and yields or gives way only when they come bowling head on into someone bigger than themselves, and often, not even then. If you've ever played a video game in the bullet hell genre like Cuphead or Contra, you'll immediately see its close relation to UK public spaces as you dodge out of the way of disoriented, directionless Britons coming at you from every conceivable angle.

But this is the cradle of western civilization! Surely I must be missing something? Mystified, I google the problem in search of answers, only to find little beyond a few pieces written [by outlets like the BBC](https://archive.is/GKutj) with post-modern anarchist rationalizations for the country's lack of order like "the best kind of rule is NO RULES" with bylines that may as well be from Tyler Durden himself.

I still refuse to believe it, and over the coming days try out various walking conventions intended to achieve greater harmony with my fellow pedestrians. But if anything, it just keeps getting worse. Walking down Constitutional Hill from Buckingham Palace to Wellington Arch, I notice that the side of the path each party is navigating along breaks down almost exactly along 50 / 50 lines. So if you stay left, you cleanly pass a party staying left going the other direction, dodge a party staying right going the other direction, cleanly pass another party staying left, dodge another party, pass another party, dodge another party, ... and everywhere you go, these small, constant assaults on sanity and common sense go on and on, all day every day.

On many paths the government has stepped in to helpfully add even more confusion to the situation by drawing a line down the center of them, and declaring via painted glyph that one side is for walking and the other for biking, except with neither side being anywhere near wide enough to support bidirectional traffic, making the system a non-starter from day one. Then, time erodes the legibility of the markings so they're barely visible, and now not even one single person in the entire country could tell you what the hell anyone is supposed to be doing. From the depths of the abyss, Lucifer emits a hearty bellow of approval. Chaos reigns supreme.

6 AM, Hyde Park. In the distance, along a perfectly straight trail, I see another runner coming in the opposite direction, maybe a half mile away. Trying to respect what I believe to be local convention, I stay left, and to my relief, I see that they do too. Finally, an encounter in this country where two humans can pass by each each other without hassle or drama, a much needed breath of fresh air after days of pinball along London's busy sidewalks. I stay my course forward, straight as a laser beam.

But as the gap between us slowly closes, I see to my absolute horror that over the corresponding distance, millimeter by millimeter, the runner opposite to me has been slowly tacking to my side of the path. At first it's just that we can no longer pass each other cleanly, but before long they're on a direct collision course, bearing straight down on me like the black sails of the Queen Anne's Revenge.

At first I thought this might be some European aggro thing, a low key demonstration of dominance, something I encountered a few times from swarthy looking guys on the streets of Paris. But in many of these cases my would be interceptor is a mature lady maybe half my size. If the worst should happen and we stayed our courses into head-on collision, there'd be exactly one clear loser in the situation, and without putting too fine of a point on it, it wouldn't be me. At the last possible second we narrowly veer around each other in an awkward dance -- the only two people in hundreds of meters in any direction -- yet each of us missing the other by mere inches. A close encounter as totally absurd as it is unnecessary, and better fit for Monty Python sketch than daily London life.

Over the few weeks of being here I've distilled my own version of the cardinal rule for walking in Britain:

> On any UK footpath, one should walk on the side which will maximize conflict with other walkers.

A long way of saying: the next time someone from the UK gives you a hard time about queuing in your country, counter by politely asking them what side of the path a person should walk on.

Only a few weeks post-coronation, the Union Jack flies proudly above London on a hundred thousand flags, but one has to wonder whether a better fit might be the Jolly Roger.

Until next week.

<img src="/photographs/nanoglyphs/038-london/london-streets@2x.jpg" alt="London streets, flags flying high" class="wide" loading="lazy">

<img src="/photographs/nanoglyphs/038-london/london-streets-2@2x.jpg" alt="London streets, flags flying high" class="wide" loading="lazy">

[1] 500 = internal server error, the bane of web application developers.

[2] Probably a unique violation resulting in two racing releases for the same app occuring nearly simultaneously. Knowing what I do now, I'm fairly certain I could've squashed this one in a day or two, but it was over our heads at the time.