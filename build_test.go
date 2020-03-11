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

func TestResizeImage_NoMozJPEG(t *testing.T) {
	if conf.MozJPEGBin == "" {
		return
	}

	oldMozJPEGBin := conf.MozJPEGBin
	defer func() {
		conf.MozJPEGBin = oldMozJPEGBin
	}()
	conf.MozJPEGBin = ""

	tmpfile, err := ioutil.TempFile("", "resized_image")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	err = resizeImage(nil, "./content/images/about/avatar.jpg", tmpfile.Name(), 100)
	assert.NoError(t, err)
}
