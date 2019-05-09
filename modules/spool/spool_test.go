package spool

import (
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestEmptyPool(t *testing.T) {
	p := NewPool(nil, 10)
	p.Run()

	assert.Equal(t, []error(nil), poolErrors(p))
	assert.Equal(t, false, p.HasErrors())
}

func TestWithWork(t *testing.T) {
	tasks := []*Task{
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
	}
	p := NewPool(tasks, 10)
	p.Run()

	assert.Equal(t, []error(nil), poolErrors(p))
	assert.Equal(t, false, p.HasErrors())
}

func TestWithError(t *testing.T) {
	tasks := []*Task{
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
		NewTask(func() error { return fmt.Errorf("error") }),
	}
	p := NewPool(tasks, 10)
	p.Run()

	assert.Equal(t, []error{fmt.Errorf("error")}, poolErrors(p))
	assert.Equal(t, true, p.HasErrors())
}

// Gets a list of errors from a pool.
func poolErrors(p *Pool) []error {
	var errs []error
	for _, task := range p.Tasks {
		if task.Err != nil {
			errs = append(errs, task.Err)
		}
	}
	return errs
}
