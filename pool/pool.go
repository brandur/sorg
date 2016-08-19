package pool

import (
	log "github.com/Sirupsen/logrus"
	"sync"
)

// Task encapsulates a work item that should go in the pool.
type Task struct {
	// Err holds an error that occurred during a task. Its result is only
	// meaningful after Run has been called for the pool that holds it.
	Err error

	f func() error
}

// NewTask initializes a new Task based on a given work function.
func NewTask(f func() error) *Task {
	return &Task{f: f}
}

// Run runs a Task and does appropriate accounting via a given sync.WorkGroup.
func (t *Task) Run(wg *sync.WaitGroup) {
	t.Err = t.f()
	wg.Done()
}

// Pool is a worker group that runs a number of Task instances at a configured
// concurrency.
type Pool struct {
	Tasks []*Task

	concurrency int
	tasksChan   chan *Task
	wg          *sync.WaitGroup
}

// NewPool initializes a new Pool with the given tasks and at the given
// concurrency.
func NewPool(tasks []*Task, concurrency int) *Pool {
	return &Pool{
		Tasks:       tasks,
		concurrency: concurrency,
		tasksChan:   make(chan *Task),
		wg:          new(sync.WaitGroup),
	}
}

// HasErrors indicates whether there were any errors from tasks run. Its result
// is only meaningful after Run has been called.
func (p *Pool) HasErrors() bool {
	for _, task := range p.Tasks {
		if task.Err != nil {
			return true
		}
	}
	return false
}

// Run runs all work within the Pool. The first error that's detected after all
// work is done is returned.
func (p *Pool) Run() {
	log.Debugf("Running %v task(s) at concurrency %v.",
		len(p.Tasks), p.concurrency)

	for i := 0; i < p.concurrency; i++ {
		go p.work()
	}

	p.wg.Add(len(p.Tasks))
	for _, task := range p.Tasks {
		p.tasksChan <- task
	}
	p.wg.Wait()

	// all workers return
	close(p.tasksChan)
}

func (p *Pool) work() {
	for task := range p.tasksChan {
		task.Run(p.wg)
	}
}
