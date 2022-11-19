+++
hook = "A trivial-but-useful pattern in Go for retrying intermittently failed HTTP requests on a pre-configured backoff schedule."
published_at = 2020-08-29T01:54:50Z
title = "A simple HTTP retry and backoff loop in Go"
+++

I was writing a Go program to run my [self-updating GitHub `README`](https://github.com/brandur) and added a little touch to make CI runs more robust by retrying intermittent HTTP failures a few times. I like how it came out:

``` go
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// A backoff schedule for when and how often to retry failed HTTP
// requests. The first element is the time to wait after the
// first failure, the second the time to wait after the second
// failure, etc. After reaching the last element, retries stop
// and the request is considered failed.
var backoffSchedule = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	10 * time.Second,
}

func getURLData(url string) (*http.Response, []byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, body, nil
}

func getURLDataWithRetries(url string) (*http.Response, []byte, error) {
	var body []byte
	var err error
	var resp *http.Response

	for _, backoff := range backoffSchedule {
		resp, body, err = getURLData(url)

		if err == nil {
			break
		}

		fmt.Fprintf(os.Stderr, "Request error: %+v\n", err)
		fmt.Fprintf(os.Stderr, "Retrying in %v\n", backoff)
		time.Sleep(backoff)
	}

	// All retries failed
	if err != nil {
		return nil, nil, err
	}

	return resp, body, nil
}

func main() {
	_, body, err := getURLDataWithRetries("https://example.com")
	if err != nil {
		panic(err)
	}

	fmt.Printf("response = %v\n", string(body))
}
```

Here's what a failed run looks like:

```
Request error: Get https://example.com: dial tcp: lookup example.com: no such host
Retrying in 1s
Request error: Get https://example.com: dial tcp: lookup example.com: no such host
Retrying in 3s
Request error: Get https://example.com: dial tcp: lookup example.com: no such host
Retrying in 10s
panic: Get https://example.com: dial tcp: lookup example.com: no such host

goroutine 1 [running]:
main.main()
        /Users/brandur/Documents/projects/go-http-retry/main.go:63 +0x11c
exit status 2
```

Normally, a backoff schedule is determined with an equation like `2 ** num_failures - 1`, but for simple programs like mine that have no intention of retrying until infinity, I like how this implementation makes the schedule explicit in a very human-readable way (see the slice `backoffSchedule` at the top).

A serious program would also add some randomness to the backoff time ("jitter"), but I've left that out for simplicity.
