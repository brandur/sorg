---
title: Global Recovery in Go With Defer
published_at: 2014-09-21T17:53:09Z
---

Having spent the last few weeks trying to become literate in Golang, I can
confidently say that all of my favorite tricks in the language [involved the
`defer` statement](https://golang.org/doc/effective_go.html#defer). Here's one
that I gleaned from the source code of hk:

``` go
package main

import (
	"fmt"
	"log"
	"os"
)

var (
	logger *log.Logger
)

func init() {
	f, _ := os.Create("log")
	logger = log.New(f, "", log.LstdFlags)
}

func main() {
	defer recoverPanic()
	panic(fmt.Errorf("This unhandled error will be handled"))
}

func recoverPanic() {
	if rec := recover(); rec != nil {
		err := rec.(error)
		logger.Printf("Unhandled error: %v\n", err.Error())
		fmt.Fprintf(os.Stderr, "Program quit unexpectedly; please check your logs\n")
		os.Exit(1)
	}
}
```

The output of this program is:

``` bash
$ ./recover
Program quit unexpectedly; please check your logs

$ echo $?
1

$ cat log
2014/09/21 11:13:11 Unhandled error: This unhandled error will be handled
```

In this simplified example above, we use a combination of `defer` and
`recover()` to handle any exception that might have been thrown with `panic()`,
log some information about it, and exit the program cleanly.

hk takes this a step futher by sending the error up to Rollbar so that all of
its unexpected crashes can be examined in aggregate.
