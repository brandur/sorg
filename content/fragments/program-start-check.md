+++
hook = "A helpful CI check that takes 10 minutes to implement, and which will save you some day."
published_at = 2021-09-30T17:36:35Z
title = "Program start check"
+++

In my first couple weeks at Stripe I broke the API's ability to start -- I can't remember if there was production fallout or not because someone else lucked out with the bad deploy and notified me later. It shouldn't have caused any damage, but this was before we even had a canary, let alone a more sophisticated progressive roll out, so it very well might've.

In my defense, it wasn't wholly my fault. The code I wrote was well-reviewed, seemed logically sound, and worked fine from the test suite. The details are hazy now, but it turned out to be a subtle problem around load order where a dependency was available when the file was loaded in by the test suite, but not available when loaded in from the executable that would run in production.

Something we added to the CI matrix afterwards is a program start check. The premise is about as simple as it gets: start the program, make sure all is well, and stop the program. Here's the one for my current employer's API in GitHub Actions:

``` yaml
jobs:
  program_starts:
    runs-on: ubuntu-latest
    timeout-minutes: 3

    steps:
      # ... setup steps ...

      - name: Check programs start
        run: |
          ( sleep 10 && killall -SIGTERM crunchy-platform-api ) &
          build/crunchy-platform-api
```

It backgrounds a `killall`, then starts the API in the foreground. Ten seconds later, the `killall` sends it a `SIGTERM`, causing it to fall through and finish the job.

We shut down gracefully on receipt of `SIGTERM`, finishing any outstanding requests and exiting cleanly with a status code of zero (and even if you're not using Go, practically every language will have a way of doing something similar):

``` go
httpServer := &http.Server{
    Addr:    addr,
    Handler: handler,
}

idleConnsClosed := make(chan struct{})
go func() {
    sigterm := make(chan os.Signal, 1)
    signal.Notify(sigterm, syscall.SIGTERM)
    <-sigterm

    logger.Info("Performing graceful shutdown")
    if err := httpServer.Shutdown(context.Background()); err != nil {
        // Error from closing listeners, or context timeout
        logger.Printf("HTTP server Shutdown: %v", err)
    }

    close(idleConnsClosed)
}()

logger.Infof("API listening on '%s'", httpServer.Addr)
if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
    logger.Errorf("Listening on '%s' failed: %v", httpServer.Addr, err)
}

<-idleConnsClosed
```

Why program start checks are a good idea:

* Test bootstraps and executable entry points bootstrap differently. Even if they're executing exactly the same core code, the small differences around the periphery might be enough to cause a problem.

* Your `main.go` or equivalent almost always has some code that's not tested. You generally test as much of your program's interior as you can, but getting right up to the edge of the outer later isn't easy.

Even if you run deploys through a staging environment, the start check is still better -- you get instant feedback right in your CI output instead of having to spalunk through logs, and you don't break staging for anyone else.

Start checks are also just an easy thing to do. Ours has saved us from staging/prod failures multiple times and I implemented it in about 10 minutes -- ROI is through the roof.
