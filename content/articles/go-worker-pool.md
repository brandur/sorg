---
title: The Case For A Go Worker Pool
published_at: 2016-08-19T21:22:23Z
hook: Error handling and fragility; or why a worker pool belongs in Go's
  standard library.
location: San Francisco
---

When it comes to the question of what the right constructs for concurrency that
a language should expose to developers, I'm a true believer that Go's channels
and goroutines are as good as it gets. They strike a nice balance between power
and flexibility, while simultaneously avoiding the propensity for deadlocks
that you'd see in a pthread model, the maintenance hell introduced by
callbacks, or the incredible conceptual overhead of promises.

However, there's one blind spot in Go's concurrency APIs that I find myself
implementing in new Go programs time and time again: the worker pool (or
otherwise known as a [thread pool][thread-pool]).

Worker pools are a model in which a fixed number of _m_ workers (implemented in
Go with goroutines) work there way through _n_ tasks in a work queue
(implemented in Go with a channel). Work stays in a queue until a worker
finishes up its current task and pulls a new one off.

Traditionally, thread pools have been useful to amortizing the costs of
spinning up new threads. Goroutines are lightweight enough that that's not a
problem in Go, but a worker pool can still be useful in controlling the number
of concurrently running tasks, especially when those tasks are accessing
resources that can easily be saturated like I/O or remote APIs.

!fig src="/assets/go-worker-pool/worker-pool.svg" caption="A visualization of a worker pool: few workers working many work items."

Implementing a worker pool in Go is by no means a tremendously difficult feat.
In fact, [Go By Example][gobyexample] describes one implementation that's only
a few dozen lines of code:

``` go
package main

import (
	"fmt"
	"time"
)

func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		fmt.Println("worker", id, "processing job", j)
		time.Sleep(time.Second)
		results <- j * 2
	}
}

func main() {
	jobs := make(chan int, 100)
	results := make(chan int, 100)

	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	for j := 1; j <= 9; j++ {
		jobs <- j
	}
	close(jobs)

	for a := 1; a <= 9; a++ {
		<-results
	}
}
```

In this example, 3 workers are started and 9 work items are in put onto a job
channel. Workers have a work loop with a `time.Sleep` so that each ends up
working 3 jobs. `close` is used on the channel after all the work's been put
onto it, which signals to all 3 workers that they can exit their work loop by
dropping them out of their loop on `range`.

This implementation is meant to show the classical reason that a worker pool
doesn't need to be in Go's standard library: the language's concurrency
primitives are already so powerful that implementing one is trivial to the
point where it doesn't even need to put into a common utility package.

So if primitives alone already present such an elegant solution, why would
anyone ever argue for introducing another unneeded layer of abstraction and
complexity?

## Error handling (#error-handling)

Well, there's a simplification in the above example that you may have spotted
already. While it's perfectly fine if the workload for our asynchronous tasks
is going to be to multiply an integer by two, it doesn't stand up quite as well
when work items may or may not have to return an error. And in a real world
system, you're _always_ going to have to return an error.

But we can fix it! To get some error handling in the program, we introduce a
new channel called `errors`. Workers will inject an error into it if they
encounter one, and otherwise put a value in `results` as usual.

``` go
errors := make(chan error, 100)

...

// check errors before using results
select {
case err := <-errors:
    fmt.Println("finished with error:", err.Error())
default:
}
```

We need to make one other small change too. Because some threads may now output
over the `errors` channel rather than `results`, we can no longer use `results`
to know when all work is complete. Instead we introduce a `sync.WaitGroup` that
workers signal when they finish work regardless of the result.

Here's a complete version of the same program with those changes:

``` go
package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup, jobs <-chan int, results chan<- int, errors chan<- error) {
	for j := range jobs {
		fmt.Println("worker", id, "processing job", j)
		time.Sleep(time.Second)

		if j%2 == 0 {
			results <- j * 2
		} else {
			errors <- fmt.Errorf("error on job %v", j)
		}
		wg.Done()
	}
}

func main() {
	jobs := make(chan int, 100)
	results := make(chan int, 100)
	errors := make(chan error, 100)

	var wg sync.WaitGroup
	for w := 1; w <= 3; w++ {
		go worker(w, &wg, jobs, results, errors)
	}

	for j := 1; j <= 9; j++ {
		jobs <- j
		wg.Add(1)
	}
	close(jobs)

	wg.Wait()

	select {
	case err := <-errors:
		fmt.Println("finished with error:", err.Error())
	default:
	}
}
```

As you can see, it's fine code, but not quite as elegant as the original.

### Potential fragility (#fragility)

In our example above, we've accidentally introduced a fairly insidious problem
in that if our error channel is smaller than the number of work items that will
produce an error, then workers will block as they try to put an error into it.
This will cause a deadlock.

We can simulate this easily by changing the size of our error channel to 1:

``` go
errors := make(chan error, 1)
```

And now when the program is run, the runtime detects a deadlock:

```
$ go run worker_pool_err.go
worker 3 processing job 1
worker 1 processing job 2
worker 2 processing job 3
worker 2 processing job 5
worker 1 processing job 4
worker 1 processing job 6
worker 1 processing job 7
fatal error: all goroutines are asleep - deadlock!
```

It's quite possible to address that problem as well, but it helps to show that
developing a useful and bug-free worker pool in Go isn't quite as simple as
it's often made out to be.

## Implementing a robust worker pool (#implementing)

Putting together a good worker pool abstraction is pretty simple, and can even
be done reliably with a minimal amount of code. Here's the worker pool
implementation that builds this website for example (or [on GitHub][github]):

``` go
import (
	"sync"
)

// Pool is a worker group that runs a number of tasks at a
// configured concurrency.
type Pool struct {
	Tasks []*Task

	concurrency int
	tasksChan   chan *Task
	wg          sync.WaitGroup
}

// NewPool initializes a new pool with the given tasks and
// at the given concurrency.
func NewPool(tasks []*Task, concurrency int) *Pool {
	return &Pool{
		Tasks:       tasks,
		concurrency: concurrency,
		tasksChan:   make(chan *Task),
	}
}

// Run runs all work within the pool and blocks until it's
// finished.
func (p *Pool) Run() {
	for i := 0; i < p.concurrency; i++ {
		go p.work()
	}

	p.wg.Add(len(p.Tasks))
	for _, task := range p.Tasks {
		p.tasksChan <- task
	}

	// all workers return
	close(p.tasksChan)

	p.wg.Wait()
}

// The work loop for any single goroutine.
func (p *Pool) work() {
	for task := range p.tasksChan {
		task.Run(&p.wg)
	}
}
```

And also simple implementation for the `Task` that goes with it. Note that we
store errors on the task itself to avoid the problem of a saturated Go channel
above:

``` go
// Task encapsulates a work item that should go in a work
// pool.
type Task struct {
	// Err holds an error that occurred during a task. Its
	// result is only meaningful after Run has been called
	// for the pool that holds it.
	Err error

	f func() error
}

// NewTask initializes a new task based on a given work
// function.
func NewTask(f func() error) *Task {
	return &Task{f: f}
}

// Run runs a Task and does appropriate accounting via a
// given sync.WorkGroup.
func (t *Task) Run(wg *sync.WaitGroup) {
	t.Err = t.f()
	wg.Done()
}
```

And here's how to run and performing error handling on it:

``` go
tasks := []*Task{
    NewTask(func() error { return nil }),
    NewTask(func() error { return nil }),
    NewTask(func() error { return nil }),
}

p := pool.NewPool(tasks, conf.Concurrency)
p.Run()

var numErrors int
for _, task := range p.Tasks {
    if task.Err != nil {
        log.Error(task.Err)
        numErrors++
    }
    if numErrors >= 10 {
        log.Error("Too many errors.")
        break
    }
}
```

## Summary (#summary)

Even though putting together a robust worker pool isn't overly problematic,
right now it's something that every project needs to handle on its own. The
size of the pattern is also almost a little _too_ simple for an external
package, as evidenced by the dozens (if not hundreds) of Go worker pool
implementations that you can find on GitHub. Coming to community consensus at
this point on a single preferred third party package would be nigh impossible.

It seems to me that this is one easy place that the Go maintainers team could
help guide developers and prevent a wild amount fracturing by introducing a One
True Path. I'd love to see a worker pool in core.

[github]: https://github.com/brandur/sorg/tree/master/pool
[gobyexample]: https://gobyexample.com/worker-pools
[thread-pool]: https://en.wikipedia.org/wiki/Thread_pool
