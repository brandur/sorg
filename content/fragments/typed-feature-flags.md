+++
hook = "A pipeline that moves feature flags defined in a backend YAML file through to OpenAPI, then onto generated TypeScript that uses types to see which flags are available and check their state of enablement."
published_at = 2024-05-09T08:41:36+02:00
title = "Plumbing fully typed feature flags from API to UI via OpenAPI"
+++

Our architecture consists of a backend Go API layer that handles state and domain logic, and a frontend TypeScript app that provides the UI.

We're allergic to too much novelty, so like many shops, we roll out features via feature flags. Almost every new feature involves checking flag state somewhere in the backend. For example, given the flag to roll out of our recent analytics product:

``` yaml
flags:
  analytics_enable:
    description: |
      Roll out flag for Crunchy Bridge for Analytics.
    kind: feature
    owner: brandur.leach@crunchydata.com
```

We'd have a check on it in our cluster creation path to make sure that only those teams flagged in can create those plans:

``` go
func (c *ClusterCreate) Run(
    ctx context.Context,
    e db.Executor,
    params *ClusterCreateParams,
) (*ClusterCreateResult, error) {

    ...

    if params.Plan.IsAnalytics() &&
        !c.Flags.IsEnabledFlaggable("analytics_enable", params.Team) {
        return nil, apierror.NewBadRequestErrorf(ctx, ErrMessageClusterPlanNotFound)
    }
```

Ideally, flags should often have a UI component as well. The flow for provisioning an analytics cluster is different than for a non-analytics one because it involves asking for a set of S3 credentials for reading/writing data sets. If a team doesn't have access to the analytics plans, the new flow should be hidden completely.

This presents a bit of a dilemma because while our API layer knows about feature flags, the frontend does not. We could give it its own flag system, but then it'd have to get into managing falg state, and we'd have duplicative systems that'd have to be synchronized and reconciled. Luckily, there's another way.

## By way of OpenAPI (#openapi)

Our Go code is built on a lightweight, in-house API framework. It makes life easier compared to raw `net/http` handlers in a variety of ways, but one of its main benefits is that it knows how to introspect itself. API endpoints are iterated, Go structs are reflected, and docstrings are parsed, after which the entirety of the corpus is translated to OpenAPI and dumped as a YAML artifact.

Amongst a plethora of other features, OpenAPI supports enums, so not only can we define the shape of a flag, but also emit every possible value:

``` yaml
FlagName:
    description: |
        The name of a flag.
    enum:
        - analytics_enable
        - metric_views_allow_unlimited_raw
        - metric_views_use_raw_metric_points
        - multi_factor_always_require
        - placeholder
        - postgres_version_12
        - postgres_version_13
    type: string
```

The generator normally emits enums for sets of constants that are defined in Go code, and flags are a little more tricky because they're defined in a YAML file that's embedded with `go:embed` and parsed on start up. We resolve this through an extra interface for types that need to define their values at runtime:

``` go
// SchemaEnumer is an interface that can be implemented by a type
// that'd like to define all its possible values dynamically at
// OpenAPI generation time rather than having them read from code.
// This is suitable for cases where values aren't available until
// runtime, like if they're read from a file.
type SchemaEnumer interface {
    SchemaEnumValues() []string
}
```

The implementation of the flag API resource:

``` go
// A flag that an account, cluster, or team is gated into, usually
// representing new features or special capabilities. Flags are
// only exposed on internal subresources.
type Flag struct {
    // Name of the flag, suitable for machine or human use. Can be
    // considered stable (flags won't suddenly be renamed unless
    // under very exceptional circumstance).
    Name FlagName `json:"name" validate:"required"`
}

// The name of a flag.
type FlagName string

// Generates enum values for OpenAPI.
func (FlagName) SchemaEnumValues() []string {
    flagInfos := pflag.GenerateDefaultFlagBundle().AllFlagInfos()

    // Control rods are an internal feature to Platform, and will
    // never be useful externally, so remove them.
    flagInfos = slices.DeleteFunc(flagInfos, func(i pflag.FlagInfo) bool {
        return i.Kind == pflag.FlagKindControlRod
    })

    return sliceutil.Map(flagInfos, func(i pflag.FlagInfo) string { return i.Name })
}
```

Flags are then exposed from the API in various places like on a special `_internal` object for at team:

``` go
// Internal-only team fields.
type TeamInternal struct {
    // A set of flags the account is gated into, usually
    // representing new features or special capabilities.
    Flags []*Flag `json:"flags" validate:"required,dive"`
    
    ...
```

## Generated TypeScript (#generated-typescript)

In the early days, our backend and frontend communicated by having the frontend manually make API invocations via an HTTP client like `GET /teams`, then manually interpreting the resulting JSON.

It was awful. There was no way to know which endpoints existed without reading Go code. Once you knew about an endpoint, you'd have to read more Go code to figure out which request parameters were required, and where they need to go (body vs. query vs. path), and once you'd finally succeed in issuing a successful request, you'd have to read _yet more_ Go code to know what to expect in the response. The whole scheme ate up copious amounts of time, required a lot of coordination, and was brittle to boot. It'd break at the drop of a pin, and did so many times.

We scrapped the whole scheme. After successfully getting an initial OpenAPI spec generated and porting the back catalog of existing APIs to the new, introspectable API framework that enabled it, we brought in [openapi-generator](https://github.com/OpenAPITools/openapi-generator) to generate TypeScript bindings that could easily slot into the frontend's TypeScript/Remix code base. It was a game changer on a magnitude that's hard to overstate. Development got 100x easier overnight.

### Generated enums (#generated-enums)

Like everything else, an OpenAPI enum gets translated into TypeScript. Flags look like this:

``` ts
/**
 * The name of a flag. 
 * @export
 * @enum {string}
 */

export const FlagName = {
    AnalyticsEnable: 'analytics_enable',
    MetricViewsAllowUnlimitedRaw: 'metric_views_allow_unlimited_raw',
    MetricViewsUseRawMetricPoints: 'metric_views_use_raw_metric_points',
    MultiFactorAlwaysRequire: 'multi_factor_always_require',
    Placeholder: 'placeholder',
    PostgresVersion12: 'postgres_version_12'
} as const;

export type FlagName = typeof FlagName[keyof typeof FlagName];
```

We implement a thin wrapper over the enum to make it easy to check a flag on an account or a team:

``` ts
/**
 * Returns true if the given account has the flag of the given
 * name. Flags are defined in the Platform API, and may be
 * associated with an account, cluster, or team, usually to gate
 * an experimental feature or special behavior.
 *
 * An account is accessible with the `useAuthenticatedUserCtx`
 * React hook:
 *
 *     const { account } = useAuthenticatedUserCtx()
 *     const isFlagOn = hasFlag(account, FlagName.MetricViewsAllowUnlimitedRaw)
 *
 * @param account_or_team - Account or team on which to check flag.
 * @param name - Name of the flag to check for on the account.
 */
export const hasFlag = makeHas(
    (account_or_team: Account | Team) =>
        account_or_team._internal?.flags?.map(flag => flag.name) ?? [],
)
```

An invocation looks like:

``` tsx
<ClusterTable
    title="All Clusters"
    actions={
        canCreateCluster(team) ? (
            hasFlag(team, FlagName.AnalyticsEnable) ? (
                <CreateActionWithAnalytics teamId={team.id} />
            ) : (
                <CreateAction teamId={team.id} />
            )
        ) : null
    }
    clusters={clusters}
/>
```

The code is nice and tidy, but better yet, it's type safe! If we were to try referencing a flag that doesn't exist, TypeScript notices immediately:

``` sh
$ npm run type-check

> bridge-express@0.0.0 type-check
> tsc --noEmit

src/app/routes/__authenticated/teams/$teamId/dashboard.tsx:171:32 - error TS2339: Property 'DoesNotExist' does not exist on type '{ readonly AnalyticsEnable: "analytics_enable"; readonly MetricViewsAllowUnlimitedRaw: "metric_views_allow_unlimited_raw"; readonly MetricViewsUseRawMetricPoints: "metric_views_use_raw_metric_points"; readonly MultiFactorAlwaysRequire: "multi_factor_always_require"; readonly Placeholder: "placeholder"; readonly PostgresVersion12: "postgres_version_12"; }'.

171         hasFlag(team, FlagName.DoesNotExist) ? (
                                   ~~~~~~~~~~~~


Found 1 error in src/app/routes/__authenticated/teams/$teamId/dashboard.tsx:171
```

## Modest outlay, major return (#major-return)

At first glance this pipeline might seem quite elaborate, but I'd put forward that it's nowhere near as bad as it looks. By far the most complicated piece of the whole thing is generating TypeScript from OpenAPI, and we didn't have to write any of that (openapi-generator is open source). Everything else -- basic flag system, Go API framework, OpenAPI reflection, TypeScript utilities -- takes some time, but not inordinate quantities of it, and they're all components that most projects will eventually want to have anyway.

We're a small company, and for many of our peers a feature-complete flags system probably seems like a luxury that's not worth investment right now, but they're one of those things that's not as hard as it sounds. A world class flags system [1] is within reach.

[1] With the exception of a GUI, which will take a little longer depending on where you're at with internal-only interfaces. This is our big missing piece right now.