package assets

import (
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestFetch(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "asset")
	assert.NoError(t, err)
	defer os.Remove(tempfile.Name())

	assets := []Asset{
		Asset{"http://localhost", tempfile.Name()},
	}

	// Because the temp file already exists, no fetch will be made.
	err = Fetch(assets)
	assert.NoError(t, err)

	// We should also have a real fetch test, but this is not currently
	// implemented.
}
