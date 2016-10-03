package downloader

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

// Number of test iterations to run while fetching files.
const numIterators = 50

func TestFetch(t *testing.T) {
	//
	// Existing file
	//

	tempfile, err := ioutil.TempFile("", "files")
	assert.NoError(t, err)
	defer os.Remove(tempfile.Name())

	files := []*File{
		{URL: "http://localhost", Target: tempfile.Name()},
	}

	// Because the temp file already exists, no fetch will be made.
	err = Fetch(files)
	assert.NoError(t, err)

	//
	// New files
	//

	dir, err := ioutil.TempDir("", "files")
	assert.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			fmt.Fprintf(w, "test-contents")
		}
	}))
	defer ts.Close()

	files = nil
	for i := 0; i < numIterators; i++ {
		file := &File{
			URL:    ts.URL + "/file" + strconv.Itoa(i),
			Target: dir + "/file" + strconv.Itoa(i),
		}
		files = append(files, file)
	}

	err = Fetch(files)
	assert.NoError(t, err)

	for i := 0; i < numIterators; i++ {
		contents, err := ioutil.ReadFile(dir + "/file" + strconv.Itoa(i))
		assert.NoError(t, err)
		assert.Equal(t, "test-contents", string(contents))
	}

	//
	// New files with error
	//

	dir, err = ioutil.TempDir("", "files")
	assert.NoError(t, err)

	files = []*File{{URL: ts.URL + "/error", Target: dir + "/file-error"}}
	for i := 0; i < numIterators; i++ {
		file := &File{
			URL:    ts.URL + "/file" + strconv.Itoa(i),
			Target: dir + "/file" + strconv.Itoa(i),
		}
		files = append(files, file)
	}

	err = Fetch(files)
	expectedErr := fmt.Errorf("Unexpected status code 500 while fetching: %v/error",
		ts.URL)
	assert.Equal(t, expectedErr, err)

	//
	// New files with *all* errors
	//

	dir, err = ioutil.TempDir("", "files")
	assert.NoError(t, err)

	for i := 0; i < numIterators; i++ {
		file := &File{
			URL:    ts.URL + "/file/error",
			Target: dir + "/file-error" + strconv.Itoa(i),
		}
		files = append(files, file)
	}

	err = Fetch(files)
	assert.Equal(t, expectedErr, err)
}
