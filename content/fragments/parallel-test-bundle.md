+++
hook = "A Go convention that we've found effective for making subtests parallel-safe, keeping them DRY, and keeping code readable."
# image = ""
published_at = 2024-10-27T16:06:21-07:00
title = "The parallel test bundle, a convention for Go testing"
+++

A year ago we went through of process of getting every test case in our project tagged with [`t.Parallel` and ratcheted with `paralleltest`](/t-parallel). I was initially skeptical about this being worth the effort because testing across Go packages was already happening in parallel, but it turned out to be a major boon for running large packages individually where we reduced test time by 30%+. We did one more step from there to tag every _subtest_ with `t.Parallel` too. The gains from that weren't as big, but it helps when running tests with many subtests one off, and isn't much effort to sustain now that it's in place.

We're running close to 5,000 tests at this point. Large scale code refactoring tools aren't widespread in Go, so I did most of the refactoring with some _very_ gnarly multi-line regexes, and even with those, the only reason that it was possible was that we're obsessive with keeping strong code convention. Most test cases were structured with an identical layout, which might've seemed like unnecessary pedantry when it was first going in, but later paid off in reams as I refactored thousands of tests in hours instead of weeks.

Let me showcase a test convention that we've found to be useful for making subtests parallel-safe, keeping them DRY (unlike many languages, Go doesn't have built-in facilities for setup/teardown blocks in tests), and keeping code readable. I try to be honest in the assessment of programming conventions and am not always certain about new ones, but we've been using the parallel test bundle for months and I'd rate it a 10/10 strong recommendation. Better yet, it's all just plain Go code and doesn't require the adoption of anything weird/novel.

## The test bundle struct (#bundle-struct)

The test bundle itself is simple struct containing the object under test and useful fixtures to have available across subtests:

``` go
type testBundle struct {
    account *dbsqlc.Account
    svc     *playgroundTutorialService
    team    *dbsqlc.Team
    tx      db.Tx
}
```

## The setup function (#setup-function)

It's paired with a `setup` helper function that returns a bundle:

``` go
setup := func(t *testing.T) (*testBundle, context.Context) {
    t.Helper()

    // These two vars are standard across almost every test case.
    var (
        ctx = ptesting.Context(t)
        tx  = ptesting.TestTx(ctx, t)
    )

    // Group of data fixtures.
    var (
        team    = dbfactory.Team(ctx, t, tx, &dbfactory.TeamOpts{})
        account = dbfactory.Account(ctx, t, tx, &dbfactory.AccountOpts{})
        _       = dbfactory.AccessGroupAccount_Admin(ctx, t, tx, team.ID, account.ID)
    )
    ctx = authntest.Account(account).Context(ctx)

    return &testBundle{
        account: account,
        svc:     pservicetest.InitAndStart(ctx, t, NewPlaygroundTutorialService(), tx.Begin, nil),
        team:    team,
        tx:      tx,
    }, ctx
}
```

Along with a test bundle, the function also returns a context [1], which is useful for seeding context with a context logger that makes sure all [logging output is collated with the test](/t-parallel#logging) being run instead of `stdout` where its output would be interleaved with that of other tests running parallel. Tests that don't need a context omit the second return value.

## Subtest invocations (#subtests)

Each subtest marks itself as parallel, and calls `setup` to procure a test bundle:

``` go
t.Run("AllProperties", func(t *testing.T) {
    t.Parallel()

    bundle, ctx := setup(t)
    
    ...
```

Each instance of a test bundle is fully insulated from every other instance, ensuring that no side effects from a test can leak into any other. Every test case uses a test transaction so that it's got its own private snapshot into the database for purposes of raising fixtures or querying.

We tend to put test bundles in every test case, even where the bundle contains only a single field. This is a courtesy to a future developer who might need to augment the test and where a preexisting test bundle makes that faster to do. It also keeps convention strong in case we need to do another broad refactor down the line.

## Complete example (#complete-example)

Here's a full code sample with all the steps together:

``` go
func TestPlaygroundTutorialServiceCreate(t *testing.T) {
   t.Parallel()

   type testBundle struct {
      account *dbsqlc.Account
      svc     *playgroundTutorialService
      team    *dbsqlc.Team
      tx      db.Txer
   }

   setup := func(t *testing.T) (*testBundle, context.Context) {
      t.Helper()

      var (
         ctx = ptesting.Context(t)
         tx  = ptesting.TestTx(ctx, t)
      )

      var (
         team    = dbfactory.Team(ctx, t, tx, &dbfactory.TeamOpts{})
         account = dbfactory.Account(ctx, t, tx, &dbfactory.AccountOpts{})
         _       = dbfactory.AccessGroupAccount_Admin(ctx, t, tx, team.ID, account.ID)
      )
      ctx = authntest.Account(account).Context(ctx)

      return &testBundle{
         account: account,
         svc:     pservicetest.InitAndStart(ctx, t, NewPlaygroundTutorialService(), tx.Begin, nil),
         team:    team,
         tx:      tx,
      }, ctx
   }

   t.Run("AllProperties", func(t *testing.T) {
      t.Parallel()

      bundle, ctx := setup(t)

      resp, err := pservicetest.InvokeHandler(bundle.svc.Create, ctx, &PlaygroundTutorialCreateRequest{
         BootstrapSQL: ptrutil.Ptr(`SELECT unnest(array[1,2,3]);`),
         Name:         "My playground tutorial",
         Content:      "# My tutorial\n\nThis is my SQL tutorial, created by **me**.",
         IsPinned:     true,
         IsPublic:     true,
         TeamID:       eid.EID(bundle.team.ID),
         Weight:       ptrutil.Ptr(int32(100)),
      })
      require.NoError(t, err)
      prequire.PartialEqual(t, &apiresourcekind.PlaygroundTutorial{
         BootstrapSQL: ptrutil.Ptr(`SELECT unnest(array[1,2,3]);`),
         Content:      "# My tutorial\n\nThis is my SQL tutorial, created by **me**.",
         IsPinned:     true,
         IsPublic:     true,
         Name:         "My playground tutorial",
         TeamID:       eid.EID(bundle.team.ID),
         Weight:       ptrutil.Ptr(int32(100)),
      }, resp)

      _, err = dbsqlc.New().PlaygroundTutorialGetByID(ctx, bundle.tx, uuid.UUID(resp.ID))
      require.NoError(t, err)

      prequire.EventForActor(ctx, t, bundle.tx, "playground_tutorial.created", bundle.account.ID)
   })
}
```

See also the [`PartialEqual` helper](/fragments/partial-equal) which I wasn't completely sure about when I first put it in, but am now fully bought into now because it's shown itself to be so effective at keeping many consecutive assertions very tidy.

[1] The context could plausibly be added to the test bundle structure as well, and that's what I started with, but embedding contexts on structs is generally frowned upon and felt weird, so it became a return value instead.