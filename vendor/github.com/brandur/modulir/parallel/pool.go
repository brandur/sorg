package parallel

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/brandur/modulir/log"
)

// Job is a wrapper for a piece of work that should be executed by the job
// pool.
type Job struct {
	Duration time.Duration
	F        func() (bool, error)
	Name     string
}

func NewJob(name string, f func() (bool, error)) *Job {
	return &Job{Name: name, F: f}
}

// Pool is a worker group that runs a number of jobs at a configured
// concurrency.
type Pool struct {
	Errors []error
	Jobs   chan *Job

	// JobsExecuted is a slice of jobs that were executed on the last run.
	JobsExecuted []*Job

	// NumJobs is the number of jobs that went through a work iteration of the
	// pool.
	//
	// This number is not accurate until Wait has finished fully. It's reset
	// when Run is called.
	NumJobs int64

	// NumJobsExecuted is the number of jobs that did some kind of heavier
	// lifting during the build loop. That's those that returned `true` on
	// execution.
	//
	// This number is not accurate until Wait has finished fully. It's reset
	// when Run is called.
	NumJobsExecuted int64

	concurrency    int
	errorsMu       sync.Mutex
	jobsInternal   chan *Job
	jobsExecutedMu sync.Mutex
	jobsFeederDone chan bool
	log            log.LoggerInterface
	roundStarted   bool
	runGate        chan struct{}
	wg             sync.WaitGroup
}

// NewPool initializes a new pool with the given jobs and at the given
// concurrency.
func NewPool(log log.LoggerInterface, concurrency int) *Pool {
	return &Pool{
		concurrency: concurrency,
		log:         log,
	}
}

func (p *Pool) Init() {
	p.log.Debugf("Initializing job pool at concurrency %v", p.concurrency)
	p.runGate = make(chan struct{})

	// Worker Goroutines
	for i := 0; i < p.concurrency; i++ {
		go func() {
			<-p.runGate

			for {
				p.workForRound()
			}
		}()
	}

	// Job feeder
	go func() {
		for {
			<-p.runGate

			for job := range p.Jobs {
				atomic.AddInt64(&p.NumJobs, 1)
				p.wg.Add(1)
				p.jobsInternal <- job
			}

			// Runs after Jobs has been closed.
			p.jobsFeederDone <- true
		}
	}()
}

// StartRound begins an execution round. Internal statistics and other tracking
// is all reset from the lsat one.
func (p *Pool) StartRound() {
	if p.roundStarted {
		panic("StartRound already called (call Wait before calling it again)")
	}

	p.Errors = nil
	p.Jobs = make(chan *Job, 500)
	p.JobsExecuted = nil
	p.NumJobs = 0
	p.NumJobsExecuted = 0
	p.jobsFeederDone = make(chan bool)
	p.jobsInternal = make(chan *Job, 500)
	p.roundStarted = true

	// Close the run gate to signal to the workers and job feeder that they can
	// start this round.
	close(p.runGate)
}

// Wait waits until all jobs are finished and stops the pool.
//
// Returns true if the round of jobs all executed successfully, and false
// otherwise. In the latter case, the caller should stop and observe the
// contents of Errors.
//
// If the pool isn't running, it falls through without doing anything so it's
// safe to call Wait multiple times.
func (p *Pool) Wait() bool {
	if !p.roundStarted {
		panic("Can't wait on a job pool that's not primed (call StartRound first)")
	}

	// Create a new run gate which Goroutines will wait on for the next round.
	p.runGate = make(chan struct{})

	p.roundStarted = false

	// First signal over the jobs chan that all work has been enqueued).
	close(p.Jobs)

	// Now wait for the job feeder to be finished so that we know all jobs have
	// been enqueued in jobsInternal.
	<-p.jobsFeederDone

	p.log.Debugf("pool: Waiting for %v job(s) to be done", p.NumJobs)

	// Now wait for all those jobs to be done.
	p.wg.Wait()

	// Drops workers out of their current round of work. They'll once again
	// wait on the run gate.
	close (p.jobsInternal)

	if p.Errors != nil {
		return false
	}
	return true
}

// The work loop for a single round within a single worker Goroutine.
func (p *Pool) workForRound() {
	for j := range p.jobsInternal {
		// Required so that we have a stable pointer that we can keep past the
		// lifetime of the loop. Don't change this.
		job := j

		start := time.Now()
		executed, err := job.F()
		job.Duration = time.Now().Sub(start)

		if err != nil {
			p.errorsMu.Lock()
			p.Errors = append(p.Errors, err)
			p.errorsMu.Unlock()
		}

		if executed {
			atomic.AddInt64(&p.NumJobsExecuted, 1)

			p.jobsExecutedMu.Lock()
			p.JobsExecuted = append(p.JobsExecuted, job)
			p.jobsExecutedMu.Unlock()
		}

		p.wg.Done()
	}
}
