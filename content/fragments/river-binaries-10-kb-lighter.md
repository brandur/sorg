+++
hook = "River v0.43 removes test-only transitive dependencies from non-test builds, shrinking Go binaries that include River by about 10 kB."
published_at = 2026-08-06T02:09:08-05:00
title = "River binaries will get 10 kB lighter"
+++

With [v0.43](https://github.com/riverqueue/river/releases/tag/v0.43.0), Go binaries linking River will get 10 kB lighter. This is good, except they never should've had the extra weight in the first place.

River has a test-oriented package called [`testsignal`](https://pkg.go.dev/github.com/riverqueue/river/rivershared/testsignal) that's used in main code paths to help check for conditions that'd otherwise require tricky timing workarounds involving sleeps. It's invoked in non-test code, but no-ops outside of tests:

``` go
// notifies when the service deletes a batch
DeletedBatch testsignal.TestSignal[struct{}]

// send the test signal
s.TestSignals.DeletedBatch.Signal(struct{}{})
```

`testsignal` has few dependencies, but we used a helper from another River package (`riversharedtest`) for test helpers which was pulling in a slew of test dependencies like Goleak, Testify, and even YAML through Testify.

This meant that River's non-test code was also picking up all these test dependencies transitively through `testsignal` and including them in any River program, adding ~10 kB in binary overhead (from `go tool nm -size`):

``` txt
github.com/stretchr/testify  18 symbols / 1168 bytes
github.com/davecgh/go-spew   11 symbols / 1312 bytes
go.uber.org/goleak            8 symbols /  896 bytes
gopkg.in/yaml                22 symbols / 4320 bytes
riversharedtest              13 symbols / 2248 bytes
```

Obviously a silly mistake to make, but one that's hard to spot and which Go doesn't protect you from, leaving latent binary bloat until someone notices it months later (in this case a contributor -- thanks [@e-yavuz-1](https://github.com/e-yavuz-1)!).

Thankfully, there's a linter to help prevent the problem. [Depguard](https://github.com/OpenPeeDeeP/depguard) (part of [golangci-lint](https://golangci-lint.run/)) lets you add rules that target a configured set of files and allow or deny a configured set of packages. I added a couple new ones for River:

``` yaml
# The next two blocks have the same intent: don't allow testsignal,
# which is used in non-test code, to have non-stdlib dependencies.
# Previously, we ran into a problem where it was accidentally importing
# riversharedtest, which was importing Goleak, Testify, YAML (through
# Testify), etc. which added 10 kB overhead to all binaries built with
# River. testsignal does use testutil, so the second block makes sure
# that testutil has no stdlib dependencies so that testsignal doesn't
# pick one up transitively.
testsignal-no-test-deps:
  files:
    - "**/testsignal/*.go"
    - "!$test"
  allow:
    - $gostd
    - "github.com/riverqueue/river/rivershared/util/testutil$"
testutil-no-test-deps:
  files:
    - "**/util/testutil/*.go"
    - "!$test"
  allow:
    - $gostd
```
