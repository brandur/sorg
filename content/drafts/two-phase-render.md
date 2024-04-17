+++
hook = "Using a two-phase data load and render pattern to prevent N+1 queries in a generalized way. Especially useful in Go, but applicable in any language."
image = "/assets/images/two-phase-render/vista.jpg"
location = "Taveuni, Fiji"
published_at = 2024-02-23T09:44:46-08:00
title = "Eradicating N+1s: The Two-phase Data Load and Render Pattern"
# hn_link = "https://news.ycombinator.com/item?id=38349716"
+++

*Author’s note:* This is a longer piece that starts off with exposition into the nature of the N+1 query problem. If you're already well familiar with it, you may want to skip my description of N+1 to a story involving a creative use of [Ruby fibers at Stripe](#fibers-and-intents) to try and plug this hole, or the [two-phase load and render](#two-phase) that I've put in my current company's Go codebase, a pattern we've been using for two years now that's rid of us N+1s, and for which I'd have trouble citing any deficiency (aside from Go's normal trouble with verbosity).

---

## N+1 in a nutshell (#n-plus-one)

Let's say we have a model `Product` that can render a public-facing API resource for itself by implementing `#render`. I'll be talking about API resources a lot because that's what I'm used, but keep in mind that this could also be an object that's used to render an HTML view and all the same concepts apply.

``` ruby
class Product < ApplicationRecord
  belongs_to :owner

  def render
    {
      id:          self.id,
      name:        self.name,
      owner_id:    self.owner_id,
      owner_email: self.owner.email,
    }
  end
end
```

Some of the properties in `#render` like `id` or `name` come directly from the model itself, and nothing beyond the initial model needs to be loaded from the database. But some, like `owner_email` must be accessed through an associated record (`product.owner`), which the data framework (ActiveRecord in this case) will happily lazy load.

Now, say ten products are rendered in a loop:

``` ruby
Product.limit(10).map do |product|
  product.render
end
```

In this naive loop, the number of database queries issued to render all products is one (`Product.limit(10)`) plus ten as `owner` is lazily loaded on each product. That's where we get "N+1" -- one initial fetch, and N as its objects are iterated and do their own loading.

This practically invisible problem is probably number two to only forgotten indexes as the most common reason for poor performance of web apps around. It's an easy mistake to make, and there's a broad lack of guard rails to protect against it.

### N*M+1 and more (#n-m-plus-one)

11 queries doesn't sound like much, but in the real world it never stops there. Let's look at a more complicated example where `Product` now has multiple associated resources along with a `Widget` subresource that has its own associations.

``` ruby
class Product < ApplicationRecord
  belongs_to :owner
  belongs_to :team
  has_many :widgets

  def render
    {
      id:          self.id,
      name:        self.name,
      owner_id:    self.owner_id,
      owner_email: self.owner.email,
      team_id:     self.team_id,
      team_name:   self.team.name,
      widget:      self.widgets.map { |w| w.render },
    }
  end
end

class Widget < ApplicationRecord
  belongs_to :factory

  def render
    {
      id:           self.id,
      factory_id:   self.factory_id,
      factory_name: self.factory.name,
      name:         self.name,
    }
  end
end
```

We're now at more like N*M+1. This is the more realistic example, and in real life it just keeps snowballing from there. Models have dozens of associations, and their subresources have subresources which have subresources. Rendering a single API resource might take hundreds, or even thousands, of database queries.

# TODO: Diagram

Luckily for all of us, databases are pretty fast, and even when abused in this fashion can still get the job done in a timely manner. ORMs like ActiveRecord also have features like [eager loading](https://guides.rubyonrails.org/active_record_querying.html#eager-loading-associations), that can be used to prefetch what otherwise would've been loaded lazily.

``` ruby
Product.includes(owner: [], team: [], widget: [:factory]).limit(10)
```

But even these sophisticated strategies have their own problems. In a large application with lots of layers, it's not obvious from any particular query if the right prefetching is happening, and it's easy to forget eager loads or put them in the wrong place.

---

## A digression: Fibers and intents (#fibers-and-intents)

Sometimes you have to get creative to solve N+1s.

A story from Stripe: due to an architecture built around Mongo, records were almost always point loaded by nothing more complex than a point index lookup (i.e. no fancy joins, eager loading, or anything else). N+1s were the rule, not the exception, but with fast hardware and modest performance expectations, it’s amazing how far you can get with this brute force model. A normal API request could easily run thousands of lookups.

It’s a good example of how pernicious N+1s can be. Databases are fast, and especially in the beginning, you can have the sloppiest internal practices imaginable and they’ll still be viable. A request might be making 50 database calls, 45 of which would be unnecessary in a better-designed system, but with each taking only 1-2 ms, everything’s still done in well under a second.

But over the years 50 calls becomes 1,000, and users start to notice that things are slow. And once things are this far gone, there’s no obvious fix. The latency isn’t due to only one factor, it’s a confluence of years worth of haphazardly written code, and now there's millions of lines worth of it.

With no easy solutions in sight, one of my colleagues came up with what to this day is still the most novel hack I've ever seen work in production.

API endpoints mapped to an API resource that they render. API resources were backed by a database model. Sometimes properties on the API resource mapped directly 1:1 to properties on the model, but especially over time, these representations tended to diverge, and custom overrides were required to map internal schema to public representation.

``` ruby
class Charge < APIResource
  prop :amount_total # maps to model directly
  prop :refund_total, render: :render_refund_total
  prop :user_email, render :render_user_email
  
  def render_refund_total
    @model.refunds.sum { |r| r.amount_total }
  end
  
  def render_user_email
    @model.user.email
  end
end
```

It was these custom overrides where N+1s were most pervasive. Models used an ORM similar to ActiveRecord or Sequel that lazily loaded related records, and rendering would more often than not require loading relations. Custom overrides often rendered subresources of their own, each of which might have its own N+1s, amplifying expense to unbounded magnitudes.

### Dynamic aggregates (#dynamic-aggregates)

This is where the innovation came in. Ruby has a construct called [fibers](https://docs.ruby-lang.org/en/master/Fiber.html) which are coroutines with a smaller memory footprint than a thread (using only small 4 kB stacks), and which can be paused and started again. The devised scheme:

* Every custom `#render_*` override would be wrapped in a fiber during invocation.
* If the fiber called into the database layer, it'd be paused with a record of its "intent" to query tracked, and the next fiber started.
* After every fiber was either paused or completed, paused fibers were examined and the database calls they’d make aggregated into batch operations if possible.
* Batch operations were invoked. Their results were disassembled and the appropriate data returned to each parked fiber.
* Paused fibers were continued. If new database calls were made, the sequence would repeat.

So from the example above, if 10 charges were rendered that mapped to 10 separate accounts, the accounts were bulked loaded with `account_id IN (?, ?, ?, ...)` instead of a single `account_id = ?`, but each fiber would get back a single account as if it'd performed a point load.

TODO: Diagram.

The system had broad limitations (e.g. only point loads could be aggregated; no complex queries were supported), but despite some gnarly code, it worked, and helped knock considerable latency off API calls. Importantly, options were limited and this was one of the few ways to have a large effect across millions of lines of code. The time where a prettier/more optimal abstraction could've been applied was long past.

---

## Rails strict loading (#rails-strict-loading)

N+1s are a constant threat in frameworks like ActiveRecord where lazy loading is common. Lazy loading is preventable with eager loading like `#includes` / `#eager_load` / `#preload`, but is difficult to guarantee because even if all relations were eager loaded initially, it’s easy to accidentally regress as a new lazy load is introduced.

To help ratchet down on the problem, [Rails 6.1 introduced **strict loading**](https://rubyonrails.org/2020/12/9/Rails-6-1-0-release#strict-loading-associations), wherein lazy loading becomes an error. The idea is that tests will exercise code which will fail if it performs a lazy load, allowing all instances of it to be banished before deployment.

``` ruby
config.active_record.strict_loading_by_default = true
```

``` ruby
class Article < ApplicationRecord
  self.strict_loading_by_default = true

  has_many :comments
end
```

Strict loading is an important feature, but not a panacea. Test coverage needs to be very substantial to make sure problems are caught before hitting production.

---

## Loading data in Go, exceptional verbosity (#go-verbosity)

This brings us to Go, where loading data is hard even without considering N+1s.

Go can aptly be described as a newer, safer C, but with even less leeway. You couldn’t write a good ORM for the language if you wanted to (they do exist, but rely on a lot of untyped `any` shenanigans, which obviate the advantages of Go in the first place since problems are only caught at runtime), and in the absence of one, the Go philosophy is to avoid abstraction -- if you need something like an API resource, piece it together query-by-query, with requisite `if err != nil { ... }` blocks after every statement.

For larger applications with dozens or hundreds of associations, the default result is a breathtaking amount of boilerplate to accomplish what would be a modest amount of code in a language with more succinct syntax and a dynamic ORM.

The increased verbosity does nothing to make N+1s less likely, which are still easy to introduce in a loop, especially with layers of indirection. It also makes them harder to fix because there might be a lot of refactoring involved. One of the first bugs I ever fixed at my job was an N+1:

``` sh
commit de58e3552eaef78c9b3d7779ddf9c646d5009985
Author: Brandur <brandur@brandur.org>
Date:   Thu Jun 3 13:06:56 2021 -0700

    Fix N+1 query getting replicas on cluster list

    We currently have an N+1 situation when listing clusters wherein we query
    replicas for every cluster picked up in the original list. This leads to
    poor performance where a user has many clusters.

    Here we fix the problem by introducing a new query that's able to select
    replicas based on a set of input IDs, and after fetching them, we assign
    them to cluster objects appropriately.
```

It was about as classic of a mistake as is possible. A query in a loop:

``` go
for _, cluster := range clusters {
    replicas, err := svc.getReplicasByClusterID(ctx, svc.executor(), cluster.ID)

    if err != nil {
        plog.Logger(ctx).Errorf("could not retrieve replicas for cluster id=[%s]: %s",
            cluster.ID, err.Error())
        continue
    }

    cluster.Replicas = replicas
}
```

This one's is easy to spot, but once queries are folded into functions and other abstractions, they get less visible and harder to address.

The fix was to query many clusters at once before the loop, and piece them together inside of it, requiring an impressive amount of code. (This was before generics arrived in 1.18, so even basic operations like rearranging models into a map wasn't possible with less than four lines of code.)

``` go
// Code in this block retrieves any replicas for these clusters and assigns
// them appropriately. All replicas are selected in one query to avoid an N+1
// problem. It would be nice to generalize this pattern because it's not pretty.
{
    clusterIDs := make([]pgtype.UUID, len(clusters))
    for i, cluster := range clusters {
        clusterIDs[i] = db.MakeUUID(cluster.ID).UUID
    }

    replicas, err := svc.getReplicasByClusterIDs(ctx, svc.executor(), clusterIDs)
    if err != nil {
        return nil, err
    }

    clusterMap := make(map[string]*dbops.Cluster)
    for _, cluster := range clusters {
        clusterMap[cluster.ID] = cluster
    }

    for _, replica := range replicas {
        cluster := clusterMap[replica.ClusterID]
        cluster.Replicas = append(cluster.Replicas, replica)
    }
}
```

Beyond the eyesore, this case-by-case approach doesn't scale well codewise either. Even this example for a single API resource with one sublist is already messy. What would happen for one with dozens of subresources, each of which might have dozen of subresources of their own? Then add a half dozen different developers into the equation, none of whom will have perfect insight into or understanding of code that anyone else wrote.

Despite Go's ad nauseum verbosity, it's no less susceptible to N+1s than a metaprogramming heavy language like Ruby.

---

## Two-phase load and render (#two-phase)

This is where our generalized data loading pattern comes in. It doesn't make N+1s impossible, but it forces developers to break convention to introduce them, making adding a new one harder than not adding one.

As the name suggests, it's broken down into two distinct render phases:

1. **Load phase:** Generates a **load bundle** from the database containing everything needed to render an **arbitrary number** of resources. Load phases always load data for N resources, even if only a single one is being rendered.

2. **Render phase:** Using a load bundle, renders a single resource. No database access is allowed.

The key insight is that the load phase knows how to load data to a bundle that's sufficient to render N resources. For a list endpoint, render may then be called using that bundle for N resources in the list. For a point retrieval endpoint, it'll render only one resource. Either way, the process is the same.

Let's look at a basic example. A product API resource, each of which has one admin and belongs to a team:

``` go
package apiresourcekind

type Product struct {
    apiresource.APIResourceBase

    ID         uuid.UUID `json:"id"`
    Name       string    `json:"name"`
    OwnerID    uuid.UUID `json:"owner_id"`
    OwnerEmail string    `json:"owner_email"`
    TeamID     uuid.UUID `json:"team_id"`
    TeamName   string    `json:"team_email"`
}
```

``` go
//
// Phase 1: Load data into a bundle
//

type ProductLoadBundle struct {
    accounts map[uuid.UUID]*dbsqlc.Account // account ID -> account
    teams    map[uuid.UUID]*dbsqlc.Team    // team ID -> team
}

func (_ *Product) LoadBundle(
    ctx context.Context, e db.Executor, baseParams *pbaseparam.BaseParams, products []*dbsqlc.Product
) (*ProductLoadBundle, error) {
    var (
        bundle  = &ProductLoadBundle{}
        queries = dbsqlc.New(e)
    )

    // Load owners for all products, map them in bundle by ID.
    {
        accounts, err := queries.AccountGetByIDMany(ctx,
            sliceutil.Map(products, func(p *dbsqlc.Product) uuid.UUID { return p.OwnerID }))
        if err != nil {
            return nil, xerrors.Errorf("error getting accounts: %w", err)
        }
        bundle.accounts = sliceutil.KeyBy(accounts, func(a *dbsqlc.Account) uuid.UUID { return a.ID })
    }

    // Load teams for all products, map them in bundle by ID.
    {
        teams, err := queries.TeamGetByIDMany(ctx,
            sliceutil.Map(products, func(p *dbsqlc.Product) uuid.UUID { return p.TeamID }))
        if err != nil {
            return nil, xerrors.Errorf("error getting teams: %w", err)
        }
        bundle.teams = sliceutil.KeyBy(teams, func(t *dbsqlc.Team) uuid.UUID { return t.ID })
    }

    return bundle, nil
}
```

``` go
//
// Phase 2: Use a bundle to render a single resource
//

func (_ *Product) Render(
    ctx context.Context, baseParams *pbaseparam.BaseParams, bundle *ProductLoadBundle, product *dbsqlc.Product
) (*Product, error) {
    return &Product{
        ID:         product.ID,
        Name:       product.Name,
        OwnerID:    product.OwnerID,
        OwnerEmail: bundle.accounts[product.OwnerID].Email,
        TeamID:     product.TeamID,
        TeamName:   bundle.teams[product.TeamID].Name,
    }, nil
}
```

A `Product` is rendered from a `ProductLoadBundle` bundle and `dbsqlc.Product` database model. Some properties like `ID` and `Name` are inherent to the product itself and are reflected directly into the API resource, but others like `OwnerEmail` and `TeamName` are only accessible by loading other database records and accessing their properties.

So, the full render process is:

1. `LoadBundle` is invoked once (regardless of the number of products being rendered).
    * Owner and team records are loaded in bulk for every product (e.g. `queries.AccountGetByIDMany` is generated by [sqlc](/sqlc), and maps to roughly `SELECT * FROM account WHERE id = any(@id::uuid[])`).
    * Owners and teams are placed into maps on `ProductLoadBundle` key to their IDs.
2. `Render` is invoked for each product individually, but reusing the same load bundle from (1). 
    * Properties like `ID` and `Name` map directly from model to API resource.
    * Indirect properties like `OwnerEmail` and `TeamName` are pulled off the records added to the load bundle in (1).

TODO: Diagram.

### Renderable (#renderable)

Implementing a full two-phase render involves a fair bit of code (again, it's Go), but once it's done, that type of API resource can easily be rendered from anywhere else:

``` go
resource, err := apiresource.Render[*apiresourcekind.ProductLoadBundle, *apiresourcekind.Product](
    ctx, tx, svc.BaseParams, product
)
if err != nil {
    return nil, err
}
```

And rendering many API resources at once (like on a list endpoint) looks like:

``` go
resources, err := apiresource.RenderMany[* apiresourcekind.ProductLoadBundle, * apiresourcekind.Product](
    ctx, tx, svc.BaseParams, products
)
if err != nil {
    return nil, err
}
```

Returned API resources implement `Renderable`, which holds types for bundle, model, and API resource:

``` go
package apiresource

// Renderable is an API resource that can be rendered by Render or RenderMany.
type Renderable[TLoadBundle any, TModel any, TResource any] interface {
    // LoadBundle loads a load bundle for the given models, usually from a
    // database, which can then be used along with a model to render a full API
    // resource.
    //
    // It may seem odd that this takes a slice of models instead of a model, but
    // this is for a good reason: it lets us batch load all data dependencies
    // all at once instead of loading them one-by-one, causing an N+1 problem.
    LoadBundle(ctx context.Context, e db.Executor, baseParams *pbaseparam.BaseParams, models []TModel) (TLoadBundle, error)

    // Render renders an API resource using a load bundle and model as input.
    Render(ctx context.Context, baseParams *pbaseparam.BaseParams, bundle TLoadBundle, model TModel) (TResource, error)
}
```

From there, implementations for `Render` and `RenderMany` are trivial, each loading a bundle once, and then rendering either a single or slice of API resources:

``` go
package apiresource

// Render renders an API resource.
//
// The type parameters may appear to be in a weird order as you might expect
// TModel before TRenderable, but it's like this for a good reason. Type
// parameters that can be inferred can be omitted, and although use of this
// function won't be able to infer TLoadBundle or TRenderable, it will infer
// TModel because it's in a parameter.
func Render[TLoadBundle any, TRenderable Renderable[TLoadBundle, TModel, TRenderable], TModel any](
    ctx context.Context, e db.Executor, baseParams *pbaseparam.BaseParams, model TModel,
) (TRenderable, error) {
    var renderable TRenderable

    bundle, err := renderable.LoadBundle(ctx, e, baseParams, []TModel{model})
    if err != nil {
        return renderable, xerrors.Errorf("error loading bundle: %w", err)
    }

    resource, err := renderable.Render(ctx, baseParams, bundle, model)
    if err != nil {
        return renderable, xerrors.Errorf("error rendering resource: %w", err)
    }

    return resource, nil
}

// RenderMany is similar to Render, but renders many API resources at once.
func RenderMany[TLoadBundle any, TRenderable Renderable[TLoadBundle, TModel, TRenderable], TModel any](
    ctx context.Context, e db.Executor, baseParams *pbaseparam.BaseParams, models []TModel,
) ([]TRenderable, error) {
    var renderable TRenderable

    bundle, err := renderable.LoadBundle(ctx, e, baseParams, models)
    if err != nil {
        return nil, xerrors.Errorf("error loading bundle: %w", err)
    }

    resources := make([]TRenderable, len(models))

    for i := range resources {
        resources[i], err = renderable.Render(ctx, baseParams, bundle, models[i])
        if err != nil {
            return nil, xerrors.Errorf("error rendering resource: %w", err)
        }
    }

    return resources, nil
}
```

### Nested resources (#nested-resources)

Watchful readers might be asking right now: what about subresources? If we need to call `apiresource.Render` inside the `Render` implementation of another resource, N+1s boomerang right back.

This is where the pattern shines. N+1s are avoided by composing load bundles onto _other load bundles_ so the `Load` implementation of a resource invokes `Load` for its subresources as well, always ensuring that there is never more than one `Load` per resource type.

This is best demonstrated by example. Let's augment `Product` above so that it renders a list of `Widget` subresources. Widgets need to do some data loading of their own, to get the location of the factory they're produced at. `Widget`'s `Renderable` implementation (widget is a leaf resource so there's nothing exotic here):

``` go
package apiresourcekind

type Widget struct {
	apiresource.APIResourceBase

	ID              uuid.UUID `json:"id"`
	FactoryID       uuid.UUID `json:"factory_id"`
	FactoryLocation string    `json:"factory_location"`
	Name            string    `json:"name"`
}

//
// Renderable implementation
//

type WidgetLoadBundle struct {
	factories map[uuid.UUID]*dbsqlc.Factory // factory ID -> factory
}

func (_ *Widget) LoadBundle(ctx context.Context, e db.Executor, baseParams *pbaseparam.BaseParams, widgets []*dbsqlc.Widget) (*WidgetLoadBundle, error) {
	var (
		bundle  = &WidgetLoadBundle{}
		queries = dbsqlc.New(e)
	)

	// Load factories for all widgets, map them in bundle by ID.
	{
		factories, err := queries.FactoryGetByIDMany(ctx,
			sliceutil.Map(widgets, func(w *dbsqlc.Widget) uuid.UUID { return w.FactoryID }))
		if err != nil {
			return nil, xerrors.Errorf("error getting factories: %w", err)
		}
		bundle.factories = sliceutil.KeyBy(factories, func(f *dbsqlc.Factory) uuid.UUID { return f.ID })
	}

	return bundle, nil
}

func (_ *Widget) Render(ctx context.Context, baseParams *pbaseparam.BaseParams, bundle *WidgetLoadBundle, widget *dbsqlc.Widget) (*Widget, error) {
	return &Widget{
		ID:              widget.ID,
		FactoryID:       widget.FactoryID,
		FactoryLocation: bundle.factories[widget.FactoryID].Location,
		Name:            widget.Name,
	}, nil
}
```

Now, back to product's (the parent resource) `Renderable` implementation, now modified to include widgets. `WidgetLoadBundle` is embedded on `ProductLoadBundle` and populated on `Load`. Product's `Render` invokes `Render` for each of its embedded widgets, passing through the common load bundle:

``` go
package apiresourcekind

type Product struct {
	apiresource.APIResourceBase

	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	OwnerID    uuid.UUID `json:"owner_id"`
	OwnerEmail string    `json:"owner_email"`
	TeamID     uuid.UUID `json:"team_id"`
	TeamName   string    `json:"team_email"`
	Widgets    []*Widget `json:"widget"`     // NEW!!
}

//
// Renderable implementation
//

type ProductLoadBundle struct {
	accounts     map[uuid.UUID]*dbsqlc.Account  // account ID -> account
	teams        map[uuid.UUID]*dbsqlc.Team     // team ID -> team
	widgetBundle *WidgetLoadBundle              // <-- the product load bundle has a widget load bundle!
	widgets      map[uuid.UUID][]*dbsqlc.Widget // product ID -> widgets
}

func (_ *Product) LoadBundle(ctx context.Context, e db.Executor, baseParams *pbaseparam.BaseParams, products []*dbsqlc.Product) (*ProductLoadBundle, error) {
	var (
		bundle  = &ProductLoadBundle{}
		queries = dbsqlc.New(e)
	)

    ...

	// Load widgets for all products, group them in bundle by product ID, and load widget bundle.
	{
		widgets, err := queries.WidgetGetByProductIDMany(ctx,
			sliceutil.Map(products, func(p *dbsqlc.Product) uuid.UUID { return p.ID }))
		if err != nil {
			return nil, xerrors.Errorf("error getting widgets: %w", err)
		}
		bundle.widgets = sliceutil.GroupBy(widgets, func(w *dbsqlc.Widget) uuid.UUID { return w.ProductID })

		bundle.widgetBundle, err = (&Widget{}).LoadBundle(ctx, e, baseParams, widgets)
		if err != nil {
			return nil, err
		}
	}

	return bundle, nil
}

func (_ *Product) Render(ctx context.Context, baseParams *pbaseparam.BaseParams, bundle *ProductLoadBundle, product *dbsqlc.Product) (*Product, error) {
	// Render widget subresources.
	var widgetResources []*Widget
	if widgets, ok := bundle.widgets[product.ID]; ok {
		widgetResources := make([]*Widget, len(widgets))
		for i, widget := range widgets {
			var err error
			widgetResources[i], err = (&Widget{}).Render(ctx, baseParams, bundle.widgetBundle, widget)
			if err != nil {
				return nil, err
			}
		}
	}

	return &Product{
		ID:         product.ID,
		Name:       product.Name,
		OwnerID:    product.OwnerID,
		OwnerEmail: bundle.accounts[product.OwnerID].Email,
		TeamID:     product.TeamID,
		TeamName:   bundle.teams[product.TeamID].Name,
		Widgets:    widgetResources,
	}, nil
}

```

The beauty of this approach is that even if your resources which have subresources _which have subresources_, it's still okay. All load bundles map 1:1:1, and regardless of number of resources or hierarchy, we still perform a constant number of database operations. Predictable performance and database load is always maintained.

### Beyond Go (#beyond-go)

Go is special because of its overwhelming verbosity and total lack of dynamic features. Even if we hadn't designed a framework to avoid N+1s, we would've had to build one to help with basic data loading, so with the two-phase load and render approach we kill two birds with one stone.

With that said, Rails' strict loading feature is a bit of an abberation. Many similar ORMs offer similar dynamic APIs that perform lazy loading, but without safety rails, making N+1s practically the default. Common practice is to live with them, and if a particular hot spot becomes a performance problem, to go in and whack-a-mole N+1s one at a time.

The two-phase approach could be extended to other languages to help make N+1s less common and more easily addressable. The syntax above looks intimidating, but once again that's mostly a Go verbosity problem. In most languages, you could do something similar with half the lines of code.

The specific code above is meant more for inspiration than anything else, and I'm not providing any particular package prescriptions. But it involves only a few plain Go structs, one interface, and two functions, so it's easy to reproduce.