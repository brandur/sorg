package modulir

import (
	"sync"
	"time"
)

// Job is a wrapper for a piece of work that should be executed by the job
// pool.
type Job struct {
	// Duration is the time it took the job to run. It's set regardless of
	// whether the job's finished state was executed, not executed, or errored.
	Duration time.Duration

	// Error is an error that the job produced, if any.
	Error error

	// Executed is whether the job "did work", signaled by it returning true.
	Executed bool

	// F is the function which makes up the job's workload.
	F func() (bool, error)

	// Name is a name for the job which is helpful for informational and
	// debugging purposes.
	Name     string
}

// NewJob initializes and returns a new Job.
func NewJob(name string, f func() (bool, error)) *Job {
	return &Job{Name: name, F: f}
}

// Pool is a worker group that runs a number of jobs at a configured
// concurrency.
type Pool struct {
	Jobs   chan *Job

	// JobsAll is a slice of all the jobs that were fed into the pool on the
	// last run.
	JobsAll []*Job

	// JobsErrored is a slice of jobs that errored on the last run.
	//
	// See also JobErrors which is a shortcut for extracting all the errors
	// from the jobs.
	JobsErrored []*Job

	// JobsExecuted is a slice of jobs that were executed on the last run.
	JobsExecuted []*Job

	concurrency    int
	jobsInternal   chan *Job
	jobsErroredMu  sync.Mutex
	jobsExecutedMu sync.Mutex
	jobsFeederDone chan struct{}
	log            LoggerInterface
	roundStarted   bool
	runGate        chan struct{}
	stop           chan struct{}
	wg             sync.WaitGroup
}

// NewPool initializes a new pool with the given jobs and at the given
// concurrency.
func NewPool(log LoggerInterface, concurrency int) *Pool {
	return &Pool{
		concurrency: concurrency,
		log:         log,
	}
}

func (p *Pool) Init() {
	p.log.Debugf("Initializing job pool at concurrency %v", p.concurrency)
	p.runGate = make(chan struct{})
	p.stop = make(chan struct{})

	// Allows us to block this function until all Goroutines have successfully
	// spun up.
	//
	// There's a potential race condition when StartRound is called very
	// quickly after Init and can close runGate before the Goroutines below
	// have a chance to start selecting on it.
	var wg sync.WaitGroup

	// Worker Goroutines
	wg.Add(p.concurrency)
	for i := 0; i < p.concurrency; i++ {
		wg.Done()
		go func() {
			for {
				select {
				case <-p.runGate:
				case <-p.stop:
					break
				}

				p.workForRound()
			}
		}()
	}

	// Job feeder
	wg.Add(1)
	go func() {
		wg.Done()
		for {
			select {
			case <-p.runGate:
			case <-p.stop:
				break
			}

			for job := range p.Jobs {
				p.wg.Add(1)
				p.jobsInternal <- job
				p.JobsAll = append(p.JobsAll, job)
			}

			// Runs after Jobs has been closed.
			p.jobsFeederDone <- struct{}{}
		}
	}()

	wg.Wait()
}

// JobErrors is a shortcut from extracting all the errors out of JobsErrored,
// the set of jobs that errored on the last round.
func (p *Pool) JobErrors() []error {
	if len(p.JobsErrored) < 1 {
		return nil
	}

	errs := make([]error, len(p.JobsErrored))
	for i, job := range p.JobsErrored {
		errs[i] = job.Error
	}
	return errs
}

func (p *Pool) Stop() {
	if p.roundStarted {
		panic("Stop should only be called after round has ended (hint: try calling Wait)")
	}

	p.stop <- struct{}{}
}

// StartRound begins an execution round. Internal statistics and other tracking
// is all reset from the lsat one.
func (p *Pool) StartRound() {
	if p.roundStarted {
		panic("StartRound already called (call Wait before calling it again)")
	}

	p.Jobs = make(chan *Job, 500)
	p.JobsAll = nil
	p.JobsErrored = nil
	p.JobsExecuted = nil
	p.jobsFeederDone = make(chan struct{})
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

	p.log.Debugf("pool: Waiting for %v job(s) to be done", len(p.JobsAll))

	// Now wait for all those jobs to be done.
	p.wg.Wait()

	// Drops workers out of their current round of work. They'll once again
	// wait on the run gate.
	close (p.jobsInternal)

	if p.JobsErrored != nil {
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
			job.Error = err

			p.jobsErroredMu.Lock()
			p.JobsErrored = append(p.JobsErrored, job)
			p.jobsErroredMu.Unlock()
		}

		if executed {
			job.Executed = true

			p.jobsExecutedMu.Lock()
			p.JobsExecuted = append(p.JobsExecuted, job)
			p.jobsExecutedMu.Unlock()
		}

		p.wg.Done()
	}
}
