package pool

import (
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestEmptyPool(t *testing.T) {
	p := NewPool(nil, 10)
	err := p.Run()
	assert.NoError(t, err)
}

func TestWithWork(t *testing.T) {
	tasks := []*Task{
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
	}
	p := NewPool(tasks, 10)
	err := p.Run()
	assert.NoError(t, err)
}

func TestWithError(t *testing.T) {
	tasks := []*Task{
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
		NewTask(func() error { return fmt.Errorf("error!") }),
	}
	p := NewPool(tasks, 10)
	err := p.Run()
	assert.Equal(t, fmt.Errorf("error!"), err)
}
