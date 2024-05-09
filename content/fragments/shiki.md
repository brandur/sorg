+++
hook = "At long last, adding syntax highlighting for code blocks to this website, and what I like about Shiki, a syntax highlighter that uses on the same engine as VSCode."
published_at = 2024-05-09T11:29:25+02:00
title = "Shiki"
+++

Last year I redesigned most of this blog. It'd been a long time since it's last facelift, but the more pressing problem was the rat's nest of CSS which had become fundamentally unmaintainable -- anything I changed had a high change of cascading unintentionally and breaking something else. Migrating to Tailwind was a must, and soon.

The rebuild was a huge project that had to be broken up so I could work on the pieces incrementally. One that fell to the bottom of the stack was syntax highlighting for the site's various code blocks. I kept telling myself that I could live without it. Code without highlighting has a certain Spartan look that's kind of badass.

But it was obvious cope. Code blocks without highlighting look awful and are unnecessarily hard to read. It was starting to bother me.

## Google: "syntax highlighter 2024" (#syntax-highlighter-2024)

And so my grand search for a new syntax highlighting library began.

I'm not that picky. Traditionally I've used Prism, but have never been in love with it, and since it's quite long in the tooth at this point, wanted something new. My only hard requirement was that it had to work entirely client-side. A build step as necessitated by the likes of Pygments was a non-starter, as was having to put NPM/Node anywhere in my stack.

I came across [Shiki](https://shiki.style/), and it immediately ticked a lot of boxes:

* Styled code looks good. This is important! So many syntax highlighters look only marginally better than no styling at all.
* Documentation exists, is easy to find, and thorough.
* It has [long list of bundled languages](https://shiki.style/languages) that doesn't force me to manually activate languages which are slightly less popular, like Ruby.
* Actively maintained. Commits as recently as a few hours ago.
* Light/dark mode support. This site notably doesn't support dark mode, but I wanted the option to be available.

## Like your styling? Keep your styling. (#keep-your-styling)

After trying it out, what became my favorite Shiki feature **by far** was one that I didn't even know I was looking for. Beyond the color/italics of the code itself, Shiki doesn't restyle anything. The borders, margins, paddings, line heights, and font sizes that you set _stay_ set. Prism by comparison assumes that it knows better than you, and changes all these things in the most invasive way imaginable, which you have to undo line by line if you care enough to do so.

It took only a few hours to set up. The only problem I ran into is that the sample code in docs for how to enable highlighting was a little _too_ trivial, and I had to write a fair bit of JavaScript myself to do the bare minimum of scanning for code blocks, parse a language from class attributes like `class="language-go"`, and unescape HTML entities. [Here's the code I ended up with](https://github.com/brandur/sorg/blob/master/views/_shiki.tmpl.html) in case anyone else wants to do the same thing.

## Sample gallery (#sample-gallery)

A few pretty code blocks from a recent piece on [typed feature flags](/fragments/typed-feature-flags), in a sprinkling of different languages:

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

``` sh
$ npm run type-check

> bridge-express@0.0.0 type-check
> tsc --noEmit

src/app/routes/__authenticated/teams/$teamId/dashboard.tsx:171:32 - error TS2339: Property 'DoesNotExist' does not exist on type '{ readonly AnalyticsEnable: "analytics_enable"; readonly MetricViewsAllowUnlimitedRaw: "metric_views_allow_unlimited_raw"; readonly MetricViewsUseRawMetricPoints: "metric_views_use_raw_metric_points"; readonly MultiFactorAlwaysRequire: "multi_factor_always_require"; readonly Placeholder: "placeholder"; readonly PostgresVersion12: "postgres_version_12"; }'.

171         hasFlag(team, FlagName.DoesNotExist) ? (
                                   ~~~~~~~~~~~~


Found 1 error in src/app/routes/__authenticated/teams/$teamId/dashboard.tsx:171
```