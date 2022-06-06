+++
image_alt = "Creek at the Forest of Nisene Marks"
image_url = "/photographs/nanoglyphs/035-generics/forest-of-nisene-marks@2x.jpg"
published_at = 2022-06-05T21:05:47Z
title = "1.18 and Generics"
+++

Go 1.18's been released for about two and a half months. We upgraded within a day or two, but decided to forego its exotic new syntax in the hopes that [golangci-lint](https://github.com/golangci/golangci-lint) would get some quick patches to be more compatible with it. It got a few, but a month later many of its lints still weren't compatible. Meanwhile, we felt the beckoning call of a whole new world of Go generics, ready to use right behind the curtain. We timed out and took for the plunge.

So far, so good. [Planet Scale](https://planetscale.com/blog/generics-can-make-your-go-code-slower) wrote a deep dive on how the use of generics makes Go slow, and although that's certainly true when it comes down to optimizing low-level code, when it comes to domain uses like ours (our project is largely a CRUD API), they've been purely beneficial, and I mean like, _very_ beneficial. Even if generics weren't a thing, my cramped hands are thanking me already with the substitution of the comically unwieldy `interface CURLY BRACE CURLY BRACE` with `any`.

Since the very beginning, Go's been a language of very strong opinions, and those strong opinions made even more stark because they're often directly contrary to what everyone else is doing. Generics were always the elephant in the room, with the only other major languages without them having deep roots in the 70s.

And generics are just the last in a long history of contrarianism. For years the most hardcore of Gophers claimed that package management was something that only other languages needed -- real men just `cp` things into a local `vendor/` directory. What you say? You care about security? Ergonomics? [_What problem are you really trying to solve_?](https://groups.google.com/g/golang-nuts/search?q=%22What%20problem%20are%20you%20trying%20to%20solve%22) Well, security and ergonomics for starters. And others:

* For the longest time, the toolchain mandated that all projects be present in a preconfigured `$GOPATH` hierarchy as opposed to where a programmer wanted to put them. There has never been any concept in programming that could so reliably confuse even the most battle-hardened veteran developers -- with practically perfect reliability I had to explain it to every single person that I onboarded to stripe-go over a period of five years (luckily, this was also solved by Go Modules).

* `go:embed` was added to the language after thousands of projects showed incredible demand for the feature, even if it meant having to use unmaintained packages.

* [Syntax highlighting is for children](https://groups.google.com/g/golang-nuts/c/hJHCAaiL0so/m/E2mQ1RDiio8J) (not an April Fools' joke).

But to their credit, in every one of these cases Go core has eventually come around and reversed course. Now with generics, Go Modules, and a host of other minor embellishments added over the years (not to mention the many good things it started with), it's on the trajectory to being one of the best languages out there. I'm even holding out hope that maybe, just maybe, one day we might even get assert helpers and stack traces.

Notably, generics don't break Go's `1.*` run. They'd originally been slated for a 2.0 release, but although 1.18 brings in the largest syntax changes since the language's first release, it manages to stay backwards compatible with all previous `1.*` releases. [I've beaten this drum before](/go-lambda), but it is a massive, _massive_ language feature not to have new versions of your language constantly breaking all your existing projects, with precious few other languages get this right.

## APIs

In order to facilitate API endpoints being able to self-describe for reflection into OpenAPI (see [docs docs docs](/nanoglyphs/031-api-docs)), and to wrap up common facilities to make developer faster and less error prone, we have a lightweight framework. A typical endpoint definition looked like this, with documentation, request/response samples, and an invocation to get a service handler that should be called into the API endpoint is executed:

``` go
func (e *ClusterCreateEndpoint) Materialize() apiendpoint.APIEndpointer {
    return &ClusterCreateEndpoint{
        &apiendpoint.APIEndpoint{
            Public:            true,
            Method:            http.MethodPost,
            Route:             "/clusters",
            Request:           &ClusterCreateRequest{},
            Response:          &apiresourcekind.Cluster{},
            ServiceHandler:    func(svc any) any { return svc.(ClusterService).Create },
            SuccessStatusCode: http.StatusCreated,
            Title:             "Create cluster",
        }
    }
}
```

It worked fine, but one of its downsides was that because we wanted service handler functions to be strongly typed with their right request and response structs, we had to use reflection to run handlers, with a core execution path dynamically instantiating a request struct, using the `reflect` package to call the handler, and then unwrapping the results as `interface{}` and interpreting them as either a successful response or error:

``` go
func (e *APIEndpoint) Execute(w http.ResponseWriter, r *http.Request) {
	req := reflect.New(e.requestType).Interface()
	
	...

	res := e.serviceHandlerVal.Call([]reflect.Value{
		reflect.ValueOf(r.Context()),
		reflect.ValueOf(req),
	})

	if len(res) != 2 {
		panic("expected exactly two arguments back from service handler")
	}

	var resp Response
	if !res[0].IsNil() {
			...
}
```

With generics, we were able to compact our API endpoint definitions down to something that looks like this, with request/response moving from structs interpreted as `interface{}` to being first-class type parameters instead:

``` go
func (e *ClusterCreateEndpoint) Materialize() apiendpoint.APIEndpointer {
    return &apiendpoint.APIEndpoint[ClusterCreateRequest, apiresourcekind.Cluster]{
        Public: true,
        Method: http.MethodPost,
        Route:  "/clusters",
        ServiceHandler: func(svc any) func(ctx context.Context, req *ClusterCreateRequest) (*apiresourcekind.Cluster, error) {
            return svc.(ClusterService).Create
        },
        SuccessStatusCode: http.StatusCreated,
        Title:             "Create cluster",
    }
}
```

This allows API instantiation and execution to become normal Go code, with `reflect` no longer being harmed anywhere in our core API paths:

``` go
func (e *APIEndpoint[TReq, TResp]) Execute(w http.ResponseWriter, r *http.Request) {
	req := new(TReq)
	
	...
	
	ret, err := e.serviceHandler(r.Context(), req)
	if err != nil {
		WriteError(r.Context(), w, err)
		return
	}

}
```

This makes things faster, but also makes the code easier to read and safer to change. Endpoint definitions are safer too, now producing a compile error if a service handler takes or responds with the wrong type:

```
`server/api/cluster_transport.go:55:11: cannot use svc.(ClusterService).Update (value of type func(context.Context, *ClusterUpdateRequest) (*apiresourcekind.Cluster, error)) as type func(ctx context.Context, req *ClusterCreateRequest) (*apiresourcekind.Cluster, error) in return statement`
```

Previously, this could only be detected at runtime, which meant that every endpoint definition needed a trivial test case to ensure that a type problem was caught in CI instead of catastrophically late once the problem in production.

## Other uses

Although 1.18 doesn't bring any new generic-based helpers directly into core, new [`x/exp/maps`](https://pkg.go.dev/golang.org/x/exp/maps) and [`x/slices`](https://pkg.go.dev/golang.org/x/exp@v0.0.0-20220518171630-0b5c67f07fdf/slices) packages have been made available with some useful helpers that many of us have been writing as manual boilerplate for every possible type ad nauseam for going on a decade now.

Another big annoyance resolved by generics are the per-type pointer helpers used when distinguishing between a `nil` versus an empty value is important, originally popularized in the AWS Go SDK, but later brought into `stripe-go` and many other projects. Previously, you'd have a separate function for `Bool`, `Int`, `Int32`, `Time`, and every other common type under the sun. Now, reduced to a single three-liner:

``` go
// previously (one of these needed for every time)
func String(v string) *string {
	return &v
}

func Ptr[T](v T) *T {
	return &v
}
```

TODO: Collect other uses.

## Limitations

By far the most noticeable limitation is that generic functions can't be defined on struct functions. Structs can have types and their functions can use those types, but functions can't define their own. So this is allowed:

``` go
type Node[T comparable] struct {
		Value T
}

func (n *Node) Equals(other T) bool {
		return n.Value == other
}
```

But this **is not**:

``` go
type Node[T comparable] struct {
		Value T
}

func (n *Node) Equals[U comparable](other U) bool {
		return n.Value == other
}
```

TODO: Other uses

## Other 1.18 niceties

The world's simplest possible crowd pleaser (well, aside from (`gen_random_uuid` in Postgres](TODO)) is `strings.Cut`, which very simply, returns two parts of a string broken on whitespace, which as it turns out is what XX% (TODO) of calls to `strings.Split` were trying to do:

``` go
tokenType, token := strings.Cut("Bearer tok_123")
```

And I don't have a clue how this one slipped in under the radar, but the [`x/sync` package now has the beginnings of a worker pool](TODO), a feature that Go has desperately needed for a long, long time:

``` go
TODO
```

## The Dropout

Since I just can't get enough of tech, last week I watched [The Dropout](TODO), about Elizabeth Holmes and Theranos. _Bad Blood_ by John Carreyrou is one of those precious few nonfiction books that reads like a thriller and keeps you on the edge of your seat the whole way through -- a legitimate 10/10, and being one of the most dramatic tech scandals of all time, I was looking forward to the TV adaptation as well.

I was a little worried when the first couple episodes started a tad slowly, but it quickly got it hooks into me. The pacing is a little uneven and it probably could've been shorter than the eight episodes it turned out to be, but by the end I appreciated the length -- especially compared to if it'd been compacted into a 120 minute movie, it gave the show enough time to explore every major character in depth. The genre is even somewhat malleable as it at times dips into the surreal, and bounces all the way to some laugh-out-loud comedy moments like those found throughout episode four (TODO) where Theranos closes a deal with Walgreens executives desperate to appear young and hip by making an imprudent deal with a darling unicorn of Silicon Valley.

The acting is all top-notch. Amanda Seyfried not only perfects Holmes' fake deep voice, but affects the perfect amount of cringe for the odder moments like her Steve Jobs worship scenes, or Theranos dance parties. Naveen Andrews seems to have been born to play the part of Balwani, who oscillates from a sympathetic character nearer to the beginning to one capable of explosive bursts of white hot anger towards good people doing the right thing, but who've come up against him. Sam Waterston's nuanced perform as George Schultz was also great -- not a malicious figure despite supporting Theranos well passed the point he should have, but an old man too proud to admit to his mistake and unable to walk it back.

There's been some post-hoc rationalization by people like [Ellen Chou (?)](TODO) that actually, against all evidence, Theranos was driven by evil men and that Elizabeth Holmes is really a misunderstood good guy who is just a victim of a misogynistic press. A [jury of her peers disagrees](TODO), and anyone laboring under this illusion should refer back to the well-cited long form Carreyrou source material -- not only was Holmes responsible for one of the largest scale frauds of all time, but toyed with peoples' lives through fake medical technology, and her vindictive character led to the suicide of one, and ruinous legal debt on the parts of multiple others.

Without giving away too much, the final scene shows a distracted Holmes apparently unable to grapple with reality as she distractedly plays with her dog and talks about her new boyfriend while her ex-legal director tries to explain the damage she's done. Again, A+++ acting down to a tee, and exactly consistent with the impression you get from Carreyrou's book -- not an inherently evil force, but one who incrementally slid ever further into the deep end until there was no going back.

Until next week.
