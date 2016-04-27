package sorg

import (
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestCreateTargetDirs(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "sorg-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	err = os.Chdir(dir)
	assert.NoError(t, err)

	err = CreateTargetDirs()
	assert.NoError(t, err)

	_, err = os.Stat(TargetDir)
	assert.NoError(t, err)
}
