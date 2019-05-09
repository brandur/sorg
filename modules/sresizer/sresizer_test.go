package sresizer

import (
	"io/ioutil"
	"path"
	"testing"

	assert "github.com/stretchr/testify/require"
)

// Doesn't really matter what this is as long as its one that we can rely being
// inside the repository.
const testImage = "../content/images/acid/pillars.jpg"

func TestResize(t *testing.T) {
	dir, err := ioutil.TempDir("", "photos")
	assert.NoError(t, err)

	job := &ResizeJob{
		SourcePath:  testImage,
		TargetPath:  path.Join(dir, "pillars_resized.jpg"),
		TargetWidth: 100,
	}

	err = Resize([]*ResizeJob{job})
	assert.NoError(t, err)
}
