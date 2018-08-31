package resizer

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ResizeJob represents work to resize a single image.
type ResizeJob struct {
	SourcePath  string
	TargetPath  string
	TargetWidth int
}

// Resize resizes a set of images based on the parameters of the jobs it receives.
func Resize(jobs []*ResizeJob) error {
	for _, job := range jobs {
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
	}

	return nil
}
