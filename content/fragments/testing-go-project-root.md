---
title: Testing Go Packages From Project Root
published_at: 2016-07-15T16:37:30Z
---

By default, running `go test` executes a package's tests with the current
working directory set to that package's path. This is fine for single-package
projects, but for a more complex one, it's often useful to execute tests from a
single root directory so to simplify access to common test resources or other
files on-disk.

I've been using a simple trick involving Go's `runtime` package to accomplish
this feat. Given a project package/directory hierarchy like this one:

```
github.com/brandur/project
  + cmd
    + executable
      - main.go
  + testing
    - testing.go
```

In `testing.go`:

``` go
package testing

import (
	"os"
	"path"
	"runtime"
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	path.Join(path.Dir(filename), "..")
	err := os.Chdir("../../")
	if err != nil {
		panic(err)
	}
}
```

The filename returned by `runtime.Caller(0)` will be the one currently
executing; `testing.go` in this case. Join the result to `..` gets the
project's root, and a call to `os.Chdir` takes the running program there.

Then from any other package in the project, import the testing package with a
blank identifier, which will run its `init` function, but including any of its
symbols:

``` go
package main_test

import (
	_ "github.com/brandur/project/testing"
)
```
