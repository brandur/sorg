package resizer

import (
	"bytes"
	"fmt"
	"os/exec"
	"sync"
)

// Number of simultaneous resizes that we should perform.
const numWorkers = 5

// ResizeJob represents work to resize a single image.
type ResizeJob struct {
	SourcePath  string
	TargetPath  string
	TargetWidth int

	Err error
}

// Resize resizes a set of images based on the parameters of the jobs it receives.
func Resize(jobs []*ResizeJob) error {
	var wg sync.WaitGroup
	wg.Add(len(jobs))

	jobsChan := make(chan *ResizeJob, len(jobs))

	// Signal workers to stop looping and shut down.
	defer close(jobsChan)

	for i := 0; i < numWorkers; i++ {
		go workJobs(jobsChan, &wg)
	}

	for _, job := range jobs {
		jobsChan <- job
	}

	wg.Wait()

	// This is not the greatest possible approach because we have to wait for
	// all jobs to be processed, but practically problems should be relatively
	// rare.
	for _, job := range jobs {
		if job.Err != nil {
			return job.Err
		}
	}

	return nil
}

func resize(job *ResizeJob) error {
	cmd := exec.Command(
		"gm",
		"convert",
		job.SourcePath,
		"-resize",
		fmt.Sprintf("%vx", job.TargetWidth),
		"-quality",
		"85",
		job.TargetPath,
	)

	var errOut bytes.Buffer
	cmd.Stderr = &errOut

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%v (stderr: %v)", err, errOut.String())
	}

	return nil
}

func workJobs(jobsChan chan *ResizeJob, wg *sync.WaitGroup) {
	// Note that this loop falls through when the channel is closed.
	for job := range jobsChan {
		err := resize(job)
		if err != nil {
			job.Err = err
		}
		wg.Done()
	}
}
