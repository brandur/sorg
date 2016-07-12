package assets

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	assert "github.com/stretchr/testify/require"
)

// Number of test iterations to run while fetching assets.
const numIterators = 50

func TestFetch(t *testing.T) {
	//
	// Existing file
	//

	tempfile, err := ioutil.TempFile("", "assets")
	assert.NoError(t, err)
	defer os.Remove(tempfile.Name())

	assets := []*Asset{
		{URL: "http://localhost", Target: tempfile.Name()},
	}

	// Because the temp file already exists, no fetch will be made.
	err = Fetch(assets)
	assert.NoError(t, err)

	//
	// New files
	//

	dir, err := ioutil.TempDir("", "assets")
	assert.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			fmt.Fprintf(w, "test-contents")
		}
	}))
	defer ts.Close()

	assets = nil
	for i := 0; i < numIterators; i++ {
		asset := &Asset{
			URL:    ts.URL + "/asset" + strconv.Itoa(i),
			Target: dir + "/asset" + strconv.Itoa(i),
		}
		assets = append(assets, asset)
	}

	err = Fetch(assets)
	assert.NoError(t, err)

	for i := 0; i < numIterators; i++ {
		contents, err := ioutil.ReadFile(dir + "/asset" + strconv.Itoa(i))
		assert.NoError(t, err)
		assert.Equal(t, "test-contents", string(contents))
	}

	//
	// New files with error
	//

	dir, err = ioutil.TempDir("", "assets")
	assert.NoError(t, err)

	assets = []*Asset{{URL: ts.URL + "/error", Target: dir + "/asset-error"}}
	for i := 0; i < numIterators; i++ {
		asset := &Asset{
			URL:    ts.URL + "/asset" + strconv.Itoa(i),
			Target: dir + "/asset" + strconv.Itoa(i),
		}
		assets = append(assets, asset)
	}

	err = Fetch(assets)
	expectedErr := fmt.Errorf("Unexpected status code 500 while fetching: %v/error",
		ts.URL)
	assert.Equal(t, expectedErr, err)

	//
	// New files with *all* errors
	//

	dir, err = ioutil.TempDir("", "assets")
	assert.NoError(t, err)

	for i := 0; i < numIterators; i++ {
		asset := &Asset{
			URL:    ts.URL + "/asset/error",
			Target: dir + "/asset-error" + strconv.Itoa(i),
		}
		assets = append(assets, asset)
	}

	err = Fetch(assets)
	assert.Equal(t, expectedErr, err)
}
