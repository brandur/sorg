+++
image_alt = "The gardens at Vizcaya in Miami"
image_url = "/photographs/nanoglyphs/019-api-libraries/vizcaya@2x.jpg"
published_at = 2021-02-15T15:58:27Z
title = "Anatomy of a Great API Library"
+++

For roughly the last six months, I've been doing as much [WaniKani](https://www.wanikani.com/) as my brain can hold. It's a beautifully built app for learning Japanese vocabulary, along with the meaning and readings of the kanji from which the words are made up. WaniKani makes a hard thing possible, but it doesn't make a hard thing easy -- learning 2,000 new symbols and 6,000 words of Japanese still feels like walking backwards up Mount Everest with fifty pounds of bricks lashed to your back -- but until Elon gets Neuralink shipped, WaniKani is the best we've got.

A delightful side benefit is that WaniKani has [an excellent API](https://docs.api.wanikani.com/20170710/). It's _so_ good and so complete that its preferred iPhone app (Tsurukame) -- which is well-polished and fully functional -- is actually _not_ written by the parent company itself, but rather just a random, unusually talented guy in Australia.

My memory is terrible, so looking to give myself an unfair advantage, I started looking into writing little plugin programs to give me some remedial, after hours education. Things snowballed, and I ended up publishing a complete [package of WaniKani API Go language bindings](https://github.com/brandur/wanikaniapi) to support my own projects.

``` go
client := wanikaniapi.NewClient(&wanikaniapi.ClientConfig{
    APIToken: os.Getenv("WANI_KANI_API_TOKEN"),
})

// Get a combined list of all "subjects": radicals, kanji, and vocabulary
subjects, err := client.SubjectList(&wanikaniapi.SubjectListParams{})
if err != nil {
    panic(err)
}
```

When I announced it on the WaniKani community Discourse forums, crickets. My post garnered a grand total of three likes (all of which I deeply appreciated). But by comparison, an average meme involving an dancing anime girl or yipping Shiba Inu gets about 100 likes, ... per hour. I didn't take it personally though. The Venn diagram of overlap between WaniKani users and Go programmers probably looks roughly like the widely-spaced round spectacles found on the likes of Harry Potter or John Lennon.

<img src="/photographs/nanoglyphs/019-api-libraries/wani-kani@2x.png" alt="WaniKani's home page" class="wide">

<!--
<img src="/photographs/nanoglyphs/019-api-libraries/glasses-cat@2x.jpg" alt="A cat with round glasses (like John Lennon? Okay yes, I'm pushing it)" class="wide">
-->

---

## Resocializing code (#resocializing-code)

Stripe is well known for its rest API and executable cURL examples in the API reference, but its real API edge are the pre-packaged API libraries built on top of it -- currently available in seven languages from Ruby to C#, and often includable in a project in just a few lines of code. Hand-building HTTP requests is slow and painful [1]. Sending them via API library is fast and easy.

One of my first projects during my first few months at Stripe was taking those API libraries and rehabilitating them. It'd been a long time since they'd been shown any love, and had since fallen into a state of deep disrepair. Restoring them to health involved, for example:

* Adding missing API resources/endpoints and correcting existing ones.

* Guiding them back to normal language-specific conventions. (I'm not sure exactly what happened there originally, but I suspect most of the original authors were Java programmers wearing thin disguises, and who enthusiastically applied Java AbstractAbstractFactory design patterns to every new language they tried.)

* Adding utilities for frequently-requested features like pagination and resource expansion.

It took time, but we got there, and Stripe's API libraries are to this day all in a good state of upkeep (although not to say that there aren't improvements that could still be made).

As I was writing my WaniKani API library, I tried to exact the best elements of stripe-go, and apply them to this much smaller, but very much alike package. I took notes as I did it.

---

### Autocomplete to success (#autocomplete-to-success)

For the last four weeks, I've been a professional Java programmer for the first time in my life. Java's an unashamedly shambling mass of a language. It got a couple things right -- good typing, good parallelism, and it's fast -- but syntax-wise it's grotesque. Overly verbose boilerplate in every direction, builders on every class to make up for the lack anything like a named or optional parameter, and few quality of life of improvements even after decades of development. The toolchain isn't standardized beyond the need for the JVM, and normal projects are a twisted mess of heavy dependencies like build frameworks and code generators just to get compiling.

But Java does have one killer feature, and it has little to do with the language itself: [IntelliJ IDEA](https://www.jetbrains.com/idea/). It generates class outlines, test suites, constructor definitions -- you name it, with just a couple taps to the keyboard. Jump-to-definition works every time instead of sometimes-you-get-it-sometimes-you-don't Vim/VSCode-based CTag-esque solutions. Renaming symbols takes one second instead of long minutes. In short, Java's verbosity is monstrous, but luckily IntelliJ does most of the typing for you.

This same principle can be applied to API libraries. Leverage language type systems to put parameters and response fields as named properties, then let a user with a smart editor/IDE auto-complete their way to success. Documentation becomes something infrequently referenced to understand high-level concepts instead of a necessary, ever-present companion that users `Cmd-Tab` to every three seconds to look up names.

### Don't make me paginate (#dont-make-me-paginate)

Every modern API implementations pagination for list endpoints -- only allowing the user to, for example, retrieve 100 objects at a time, and iterating every 100-object page until they get the entire collection, with each page sending back a cursor to tell the user how to ask for the next one.

Writing client code to paginate isn't _that_ hard, but getting it right is hard enough, and users have better things to be doing with their time. Include built-in helpers in your API library to do pagination automatically.

For my money, the best approach is to use each pagination API call as the unit of granularity (as opposed to each individual object):

``` go
var allSubjects []*wanikaniapi.Subject

//
// Invokes its closure once for every API call made
// (where each API call is one page)
//
err := client.PageFully(func(id *wanikaniapi.WKID) (*wanikaniapi.PageObject, error) {
    page, err := client.SubjectList(&wanikaniapi.SubjectListParams{
        ListParams: wanikaniapi.ListParams{
            PageAfterID: id,
        },
    })
    if err != nil {
        return nil, err
    }

    allSubjects = append(allSubjects, page.Data...)
    return &page.PageObject, nil
})

if err != nil {
    panic(err)
}
```

That way, API calls aren't hidden away from the user, and it gives them perfect control over starting the next API call or stopping iteration.

### Provide porcelain, but expose the plumbing (#expose-plumbing)

Years ago, the Git book put forward the evergreen analogy of providing both [porcelain and plumbing](https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain):

> This book covers primarily how to use Git with 30 or so subcommands such as `checkout`, `branch`, `remote`, and so on. But because Git was initially a toolkit for a version control system rather than a full user-friendly VCS, it has a number of subcommands that do low-level work and were designed to be chained together UNIX-style or called from scripts. These commands are generally referred to as Git’s “plumbing” commands, while the more user-friendly commands are called “porcelain” commands.

Expose high-level constructs, but also make sure that low-level primitives are provided for users that need them.

The same should go for API libraries -- make sure that users have all the tools they need. The best example of this for API libraries is making sure that the underlying HTTP transport can be configured to the user's precise specifications. Here's `wanikaniapi` being configured with a custom `http.Client`:

``` go
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:       10,
        IdleConnTimeout:    30 * time.Second,
        DisableCompression: true,
    },
}

client := wanikaniapi.NewClient(&wanikaniapi.ClientConfig{
    APIToken:   os.Getenv("WANI_KANI_API_TOKEN"),
    HTTPClient: httpClient,
})
```

That way, the user can tweak key configuration like read and write timeouts, an HTTP proxy, or even a fully custom transport layer in case they've got a highly customized internal setup.

### Write a great README (#readme)

We're so used to it now that we don't even think about it anymore, but one of the greatest innovations that GitHub ever made was automatically plastering the beautiful, Markdown-rendered contents of a project's README right on its landing page. This one, simple advancement dawned a new age of README-driven development where users can often get bootstrapped on a new project in mere minutes instead of through a lengthy configuration process.

READMEs are increasingly useful beyond GitHub too. Rust automatically renders a project's `README.md` [on a crate's page](https://crates.io/crates/actix), and Go's new [pkg.go.dev](https://pkg.go.dev/github.com/brandur/wanikaniapi) documentation system pulls in a project's README to put in the introductory section above the reference. Documentation in two places for the price of one.

<!--
<img src="/photographs/nanoglyphs/019-api-libraries/pkg-go-dev-short@2x.png" alt="A README rendered on pkg.go.dev" class="wide">
-->

### Bake in retries (#retries)

Any integration, regardless of the size, is eventually going to have some requests that fail due to transient network or server problems. Put in a basic retry mechanism for people to quickly and easily take advantage of. Implementing a basic exponential backoff algorithm is as simple as `2 ^ num_retries` (1, 2, 4, 16, ...), and any RNG can be used to add a little bit of jitter. Allow retries to be disabled in case a sophisticated user wants to bring their own implementation.

### Pluggable instrumentation (#pluggable-instrumentation)

People who run production systems care a lot about observability, and good API libraries will provide hooks to enable that. A simple example of this is to allow a logger to be injected so that users can inject their own logging implementation and control thereby control exactly how much information is logged.

The right way to do this depends on the language because conventions vary widely. In Go, I've found an interface like this one to be effective:

``` go
type LeveledLoggerInterface interface {
    Debugf(format string, v ...interface{})
    Errorf(format string, v ...interface{})
    Infof(format string, v ...interface{})
    Warnf(format string, v ...interface{})
}
```

Popular loggers like [Logrus](https://github.com/sirupsen/logrus/) and [Zap's SuggaredLogger](https://godoc.org/go.uber.org/zap#SugaredLogger) support this interface out-of-the-box, so you automatically have broad support. Include a simple, default implementation so that logging still works for integrations that aren't using one of those.

It might also be a good idea to add extensible hooks to various key places that allow users to customize behavior (e.g. request start, request end, retry, etc.). The larger the user, the more they will care about this. Notably, stripe-ruby's hooks system was [contributed by a developer from Shopify](https://github.com/stripe/stripe-ruby/pull/870) (as big of a user as there is) because they wanted a way to emit custom StatsD metrics as requests were being made.

``` ruby
Stripe::Instrumentation.subscribe(:request_end) do |request_event|
  tags = {
    method:   request_event.method,
    resource: request_event.path.split('/')[2],
    code:     request_event.http_status,
    retries:  request_event.num_retries
  }
  StatsD.distribution('stripe_request', request_event.duration, tags: tags)
end
```

---

<img src="/photographs/nanoglyphs/019-api-libraries/brickell@2x.jpg" alt="Around Brickell in Miami" class="wide">

## Miami Vice, live (#miami-vice-live)

I'm writing this, laptop out, sitting along the sunny shoreline in Miami's Brickell district. As I look out onto the water, my perfectly tranquil day is interrupted when a passing Florida state police boat suddenly cranks its throttle to 11 and races passed an inbound boat out into the bay. A few hundred feet on it comes to a sudden stop, and one of the officers on board uses a net to fish something small out of the water -- a bundle that appeared to have been dropped by the inbound boat.

With the retrieval finished, the police boat does a hard 180-degree turn, opens its throttle back to full, and races back towards the inbound boat. They activate their sirens and hail it, telling them to stop their engines. I don't know the first thing about anything naval, but it's obvious even to my landlubber eyes that the police boat's helmsman is incredibly skilled. As currents tug at both boats, he's able to maneuver his in tight circles around the other vessel, staying mere feet from its hull the whole time.

The police call in another boat for backup, lash themselves to the target, and board them. They were too far away to hear anything, but a few minutes later everyone on board is arrested, and their boat being towed up the Miami River. I have no idea what I just witnessed. Is this what a smuggling bust looks like?

I messaged a friend about it, and he responded, "Miami Vice _LIVE_, man. You're living it."

Until next week.

[1] Especially for an API like Stripe's which bundles its own mutant strain of Rack's mechanism for encoding arrays and maps, neither of which `application/x-www-form-urlencoded` was ever intended to support.
