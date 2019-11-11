package modulir

import (
	"sort"
	"sync"
	"time"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Public
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Job is a wrapper for a piece of work that should be executed by the job
// pool.
type Job struct {
	// Duration is the time it took the job to run. It's set regardless of
	// whether the job's finished state was executed, not executed, or errored.
	Duration time.Duration

	// Err is an error that the job produced, if any.
	Err error

	// Executed is whether the job "did work", signaled by it returning true.
	Executed bool

	// F is the function which makes up the job's workload.
	F func() (bool, error)

	// Name is a name for the job which is helpful for informational and
	// debugging purposes.
	Name string
}

// Error returns the error message of the error wrapped in the job if this was
// an errored job. Job implements the error interface so that it can return
// itself in situations where error handling is being done but job errors may
// be mixed in with other sorts of errors.
//
// It panics if the job wasn't errored, so be careful to only use this when
// iterating across something like Pool.JobsErrored.
func (j *Job) Error() string {
	if j.Err == nil {
		panic("Error called on a non-errored Job")
	}

	return j.Err.Error()
}

// NewJob initializes and returns a new Job.
func NewJob(name string, f func() (bool, error)) *Job {
	return &Job{Name: name, F: f}
}

// Pool is a worker group that runs a number of jobs at a configured
// concurrency.
type Pool struct {
	Jobs chan *Job

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

	colorizer      *colorizer
	concurrency    int
	initialized    bool
	jobsInternal   chan *Job
	jobsErroredMu  sync.Mutex
	jobsExecutedMu sync.Mutex
	jobsFeederDone chan struct{}
	log            LoggerInterface
	roundStarted   bool
	runGate        chan struct{}
	stop           chan struct{}
	wg             sync.WaitGroup

	// Current state of workers. Used for debugging.
	workerJobs   []*Job
	workerStates []workerState
}

// NewPool initializes a new pool with the given jobs and at the given
// concurrency. It calls Init so that the pool is fully spun up and ready to
// start a round.
func NewPool(log LoggerInterface, concurrency int) *Pool {
	// By default a pool gets a no-op colorizer. NewContext may set one
	// separately for pools created within the package.
	pool := &Pool{
		colorizer:   &colorizer{LogColor: false},
		concurrency: concurrency,
		log:         log,
	}
	pool.Init()
	return pool
}

// Init initializes the pool by preparing state and spinning up Goroutines. It
// should only be called once per pool.
func (p *Pool) Init() {
	if p.initialized {
		panic("Init called for a pool that's already been initialized")
	}

	p.log.Debugf("Initializing job pool at concurrency %v", p.concurrency)

	p.initialized = true
	p.runGate = make(chan struct{})
	p.stop = make(chan struct{})

	p.workerJobs = make([]*Job, p.concurrency)
	p.workerStates = make([]workerState, p.concurrency)

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
		workerNum := i
		go func() {
			wg.Done()

			p.setWorkerState(workerNum, workerStateWaitingOnRunOrStop, nil)

			for {
				select {
				case <-p.runGate:
				case <-p.stop:
					break
				}

				p.workForRound(workerNum)
			}

			p.setWorkerState(workerNum, workerStateFinished, nil)
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
		errs[i] = job.Err
	}
	return errs
}

// LogErrors logs a limited set of errors that occurred during a build.
func (p *Pool) LogErrors() {
	p.LogErrorsSlice(p.JobErrors())
}

// LogErrorsSlice logs a limited set of errors from the given slice.
func (p *Pool) LogErrorsSlice(errors []error) {
	if errors == nil {
		return
	}

	for i, err := range errors {
		// When dealing with an errored job (in practice, this is going to be
		// the common case), we can provide a little more detail on what went
		// wrong.
		job, ok := err.(*Job)

		if ok {
			p.log.Errorf(
				p.colorizer.Bold(p.colorizer.Red("Job error:")).String()+
					" %v (job: '%s', time: %v)",
				job.Err, job.Name, job.Duration.Truncate(100*time.Microsecond))
		} else {
			p.log.Errorf(
				p.colorizer.Bold(p.colorizer.Red("Build error:")).String()+
					" %v",
				err)
		}

		if i >= maxMessages-1 {
			p.log.Errorf("... too many errors (limit reached)")
			break
		}
	}
}

// LogSlowest logs a limited set of executed jobs from the last build starting
// with the slowest jobs on top.
func (p *Pool) LogSlowest() {
	p.LogSlowestSlice(p.JobsExecuted)
}

// LogSlowestSlice logs a limited set of executed jobs from the given slice.
func (p *Pool) LogSlowestSlice(jobs []*Job) {
	sortJobsBySlowest(jobs)

	for i, job := range jobs {
		// Having this in the loop ensures we don't print it if zero jobs
		// executed
		if i == 0 {
			p.log.Infof("Jobs executed (slowest first):")
		}

		p.log.Infof(
			p.colorizer.Bold(p.colorizer.Cyan("    %s")).String()+
				" (time: %v)",
			job.Name, job.Duration.Truncate(100*time.Microsecond))

		if i >= maxMessages-1 {
			p.log.Infof("... many jobs executed (limit reached)")
			break
		}
	}
}

// Stop disables and cleans up the pool by spinning down all Goroutines.
func (p *Pool) Stop() {
	if !p.initialized {
		panic("Stop called for a pool that's not initialized")
	}

	if p.roundStarted {
		panic("Stop should only be called after round has ended (hint: try calling Wait)")
	}

	p.initialized = false
	p.stop <- struct{}{}
}

// StartRound begins an execution round. Internal statistics and other tracking
// is all reset from the lsat one.
func (p *Pool) StartRound() {
	if !p.initialized {
		panic("StartRound called for a pool that's not initialized (hint: call Init first)")
	}

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

	// Prints some debug information to help us in case we run into stalling
	// problems in the main job loop.
	done := make(chan struct{}, 1)
	go func() {
		select {
		case <-time.After(waitSoftTimeout):
			p.printWaitTimeoutInfo()
		case <-done:
		}
	}()

	p.log.Debugf("pool: Waiting for %v job(s) to be done", len(p.JobsAll))

	// Now wait for all those jobs to be done.
	p.wg.Wait()

	// Kill the timeout Goroutine.
	done <- struct{}{}

	// Drops workers out of their current round of work. They'll once again
	// wait on the run gate.
	close(p.jobsInternal)

	if p.JobsErrored != nil {
		return false
	}
	return true
}

func (p *Pool) printWaitTimeoutInfo() {
	p.log.Errorf(
		"Wait soft timeout (jobs executed: %v, errored: %v, left: %v)",
		len(p.JobsExecuted),
		len(p.JobsErrored),
		len(p.jobsInternal),
	)
	for i := 0; i < p.concurrency; i++ {
		workerJob := "<none>"
		if p.workerJobs[i] != nil {
			workerJob = p.workerJobs[i].Name
		}

		workerState := p.workerStates[i]

		p.log.Errorf("    Worker %v state: %v, job: %v",
			i, workerState, workerJob)
	}
}

func (p *Pool) setWorkerState(workerNum int, state workerState, job *Job) {
	p.workerJobs[workerNum] = job
	p.workerStates[workerNum] = state
}

// The work loop for a single round within a single worker Goroutine.
func (p *Pool) workForRound(workerNum int) {
	for j := range p.jobsInternal {
		// Required so that we have a stable pointer that we can keep past the
		// lifetime of the loop. Don't change this.
		job := j

		p.setWorkerState(workerNum, workerStateJobExecuting, job)

		// Start a Goroutine to track the time taken to do this work.
		// Unfortunately, we can't actually kill a timed out Goroutine because
		// Go (and we rely on the user to make sure these get fixed instead),
		// but we can at least raise on the interface which job is problematic
		// to help identify what needs to be fixed.
		done := make(chan struct{}, 1)
		go func() {
			select {
			case <-time.After(jobSoftTimeout):
				p.log.Errorf("Job soft timeout (job: '%s')", job.Name)
			case <-done:
			}
		}()

		start := time.Now()
		executed, err := job.F()
		job.Duration = time.Now().Sub(start)

		// Kill the timeout Goroutine.
		done <- struct{}{}

		if err != nil {
			job.Err = err

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

		// Unset active job
		p.setWorkerState(workerNum, workerStateJobFinished, nil)
	}
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private
//
//
//
//////////////////////////////////////////////////////////////////////////////

const (
	// When to report that a job is probably timed out. We call it a "soft"
	// timeout because we can't actually kill jobs.
	jobSoftTimeout = 15 * time.Second

	// Maximum number of errors or jobs to print on screen after a build loop.
	maxMessages = 10

	// When to report that a wait round is probably timed out. We call it a
	// "soft" timeout because no jobs are killed -- it's just for reporting and
	// debugging purposes.
	waitSoftTimeout = 60 * time.Second
)

// Keeps track of the state of a worker. Used for debugging purposes only.
type workerState string

// The possible states that a worker can be in. Used for debugging purposes
// only.
const (
	workerStateFinished           workerState = "worker_finished"
	workerStateJobExecuting       workerState = "job_executing"
	workerStateJobFinished        workerState = "job_finished"
	workerStateWaitingOnRunOrStop workerState = "waiting_on_run_or_stop"
)

// Sorts a slice of jobs with the slowest on top.
func sortJobsBySlowest(jobs []*Job) {
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[j].Duration < jobs[i].Duration
	})
}
