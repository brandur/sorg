package main

import (
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestResizeImage(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "resized_image")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	err = resizeImage(nil, "./content/images/about/avatar.jpg", tmpfile.Name(), 100)
	assert.NoError(t, err)
}
