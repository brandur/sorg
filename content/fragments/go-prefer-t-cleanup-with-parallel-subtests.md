+++
hook = "Go's `defer` and `t.Cleanup` have similar semantics, but there's a good reason to prefer the use of `t.Cleanup` specifically in the presence of parallel subtests."
published_at = 2023-09-05T17:32:10-07:00
title = "Why to prefer `t.Cleanup` to `defer` in tests with subtests using `t.Parallel`"
+++

Go's [`defer` statement](https://go.dev/ref/spec#Defer_statements) and its [`t.Cleanup` test function](https://pkg.go.dev/testing#T.Cleanup) overlap each other in functionality, and at first glance seem to be interchangeable.

Some time ago I came across the [tparallel linter](https://github.com/moricho/tparallel), which **enforces the use of `t.Cleanup` over `defer` in parent test functions where `t.Parallel()` is used in subtests**.

I didn't initially understand why, and the project's README was light on details, so I [opened an issue asking about it](https://github.com/moricho/tparallel/issues/23). Recently [@telemachus](https://github.com/telemachus) responded to with translated content from the [original Japanese blog post](https://engineering.mercari.com/en/blog/entry/20220408-how_to_use_t_parallel/) that went into deeper background, and which I'll summarize here.

Consider the following example where subtests using `t.Run` make `t.Parallel` calls to run in parallel:

``` go
func Test_Func1(t *testing.T) {
    defer trace("Test_Func1")()

    t.Run("Func1_Sub1", func(t *testing.T) {
        defer trace("Func1_Sub1")()

        t.Parallel()

        // ...
    })

    t.Run("Func1_Sub2", func(t *testing.T) {
        defer trace("Func1_Sub2")()

        t.Parallel()

        // ...
    })

    // ...
}
```

Run with `-test.v` to see this output:

```
=== RUN   Test_Func1
Test_Func1 entered
=== RUN   Test_Func1/Func1_Sub1
Func1_Sub1 entered                          <- Func1_Sub1が開始
=== PAUSE Test_Func1/Func1_Sub1             <- Func1_Sub1が一時停止
=== RUN   Test_Func1/Func1_Sub2
Func1_Sub2 entered                          <- Func1_Sub2が開始
=== PAUSE Test_Func1/Func1_Sub2             <- Func1_Sub2が一時停止
Test_Func1 returned                         <- Test_Func1の呼び出し戻り（＊）
=== CONT  Test_Func1/Func1_Sub1             <- Func1_Sub1が再開
Func1_Sub1 returned                         <- Func1_Sub1が完了
=== CONT  Test_Func1/Func1_Sub2             <- Func1_Sub2が再開
Func1_Sub2 returned                         <- Func1_Sub2が完了
--- PASS: Test_Func1 (0.00s)                <- Test_Func1の結果表示
    --- PASS: Test_Func1/Func1_Sub1 (0.00s)
    --- PASS: Test_Func1/Func1_Sub2 (0.00s)
```

Look closely, and you'll see that the deferred trace on `Test_Func1` returns _before_ either `Func1_Sub1` or `Func1_Sub2` finish running, which seems to violate `defer`'s guaranteed LIFO (last-in-first-out) ordering.

Stopping for a moment to think, it makes sense. In order to known which tests can run in parallel, Go would first have to perform an initial informational pass, because otherwise it'd have no way of knowing which tests are marked with `t.Parallel` and which are not. It does this by running each test/subtest in a goroutine, pausing that goroutine when `t.Parallel()` is encountered, and later continuing each as appropriate for a complete run.

In cases where `t.Parallel` is used in subtests, the top-level test function is allowed to finish while subtests are still paused, so its `defer` statements can run before subtests get a chance to finish. This behavior is surprisingly and could easily leading to bugs, which is why the tparallel lint was born.

The fix is `t.Cleanup`. Like `defer`, `t.Cleanup` also guarantees LIFO order, but unlike `defer`, Go's test framework is aware of it, so even with pauses caused by `t.Parallel`, it behaves as expected.
