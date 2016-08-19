package pool

import (
	log "github.com/Sirupsen/logrus"
	"sync"
)

type Task struct {
	Err  error
	Func func() error
}

func NewTask(f func() error) *Task {
	return &Task{Func: f}
}

func (t *Task) Run(wg *sync.WaitGroup) {
	t.Err = t.Func()
	wg.Done()
}

type Pool struct {
	concurrency int
	tasks       []*Task
	tasksChan   chan *Task
	wg          *sync.WaitGroup
}

func NewPool(tasks []*Task, concurrency int) *Pool {
	return &Pool{
		concurrency: concurrency,
		tasks:       tasks,
		tasksChan:   make(chan *Task),
		wg:          new(sync.WaitGroup),
	}
}

func (p *Pool) Run() error {
	log.Infof("Running %v task(s) at concurrency %v",
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
		if task.Err != nil {
			return task.Err
		}
	}

	return nil
}

func (p *Pool) work() {
	for task := range p.tasksChan {
		task.Run(p.wg)
	}
}
