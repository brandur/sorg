package pool

import (
	log "github.com/Sirupsen/logrus"
	"sync"
)

// Task encapsulates a work item that should go in the pool.
type Task struct {
	err error
	f   func() error
}

// NewTask initializes a new Task based on a given work function.
func NewTask(f func() error) *Task {
	return &Task{f: f}
}

// Run runs a Task and does appropriate accounting via a given sync.WorkGroup.
func (t *Task) Run(wg *sync.WaitGroup) {
	t.err = t.f()
	wg.Done()
}

// Pool is a worker group that runs a number of Task instances at a configured
// concurrency.
type Pool struct {
	concurrency int
	tasks       []*Task
	tasksChan   chan *Task
	wg          *sync.WaitGroup
}

// NewPool initializes a new Pool with the given tasks and at the given
// concurrency.
func NewPool(tasks []*Task, concurrency int) *Pool {
	return &Pool{
		concurrency: concurrency,
		tasks:       tasks,
		tasksChan:   make(chan *Task),
		wg:          new(sync.WaitGroup),
	}
}

// Run runs all work within the Pool. The first error that's detected after all
// work is done is returned.
func (p *Pool) Run() error {
	log.Debugf("Running %v task(s) at concurrency %v.",
		len(p.tasks), p.concurrency)

	for i := 0; i < p.concurrency; i++ {
		go p.work()
	}

	p.wg.Add(len(p.tasks))
	for _, task := range p.tasks {
		p.tasksChan <- task
	}
	p.wg.Wait()

	// all workers return
	close(p.tasksChan)

	for _, task := range p.tasks {
		if task.err != nil {
			return task.err
		}
	}

	return nil
}

func (p *Pool) work() {
	for task := range p.tasksChan {
		task.Run(p.wg)
	}
}
