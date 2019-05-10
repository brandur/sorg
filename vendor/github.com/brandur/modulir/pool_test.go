package modulir

import (
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestEmptyPool(t *testing.T) {
	p := NewPool(&Logger{Level: LevelDebug}, 10)
	defer p.Stop()

	p.StartRound()
	p.Wait()

	assert.Equal(t, 0, len(p.JobsAll))
	assert.Equal(t, 0, len(p.JobsErrored))
	assert.Equal(t, 0, len(p.JobsExecuted))
	assert.Equal(t, []error(nil), p.JobErrors())
}

func TestWithWork(t *testing.T) {
	p := NewPool(&Logger{Level: LevelDebug}, 10)
	defer p.Stop()

	p.StartRound()
	j0 := NewJob("job 0", func() (bool, error) { return true, nil })
	p.Jobs <- j0
	j1 := NewJob("job 1", func() (bool, error) { return true, nil })
	p.Jobs <- j1
	j2 := NewJob("job 2", func() (bool, error) { return false, nil })
	p.Jobs <- j2
	p.Wait()

	// Check state on the pool
	assert.Equal(t, 3, len(p.JobsAll))
	assert.Equal(t, 0, len(p.JobsErrored))
	assert.Equal(t, 2, len(p.JobsExecuted)) // Number of `return true` above
	assert.Equal(t, []error(nil), p.JobErrors())

	// Check state on individual jobs
	assert.Equal(t, true, j0.Executed)
	assert.Equal(t, nil, j0.Error)
	assert.Equal(t, true, j1.Executed)
	assert.Equal(t, nil, j1.Error)
	assert.Equal(t, false, j2.Executed)
	assert.Equal(t, nil, j2.Error)
}

func TestWithError(t *testing.T) {
	p := NewPool(&Logger{Level: LevelDebug}, 10)
	defer p.Stop()

	p.StartRound()
	j0 := NewJob("job 0", func() (bool, error) { return true, nil })
	p.Jobs <- j0
	j1 := NewJob("job 1", func() (bool, error) { return true, nil })
	p.Jobs <- j1
	j2 := NewJob("job 2", func() (bool, error) { return true, fmt.Errorf("error") })
	p.Jobs <- j2
	p.Wait()

	// Check state on the pool
	assert.Equal(t, 3, len(p.JobsAll))
	assert.Equal(t, 1, len(p.JobsErrored))
	assert.Equal(t, 3, len(p.JobsExecuted)) // Number of `return true` above
	assert.Equal(t, []error{fmt.Errorf("error")}, p.JobErrors())

	// Check state on individual jobs
	assert.Equal(t, true, j0.Executed)
	assert.Equal(t, nil, j0.Error)
	assert.Equal(t, true, j1.Executed)
	assert.Equal(t, nil, j1.Error)
	assert.Equal(t, true, j2.Executed)
	assert.Equal(t, fmt.Errorf("error"), j2.Error)
}
