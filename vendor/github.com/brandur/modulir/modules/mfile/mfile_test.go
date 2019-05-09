package mfile

// Hopefully the beginnings of getting some testing started.
/*
import (
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/require"
)

var targetDir = "./public"

func TestCreateTargetDirs(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "sorg-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	err = os.Chdir(dir)
	assert.NoError(t, err)

	err = CreateOutputDirs(targetDir)
	assert.NoError(t, err)

	_, err = os.Stat(targetDir)
	assert.NoError(t, err)
}
*/
