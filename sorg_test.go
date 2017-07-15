package sorg

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

func TestSplitFrontmatter(t *testing.T) {
	frontmatter, content, err := SplitFrontmatter(`---
foo: bar
---

other`)
	assert.NoError(t, err)
	assert.Equal(t, "foo: bar", frontmatter)
	assert.Equal(t, "other", content)

	frontmatter, content, err = SplitFrontmatter(`other`)
	assert.NoError(t, err)
	assert.Equal(t, "", frontmatter)
	assert.Equal(t, "other", content)

	frontmatter, content, err = SplitFrontmatter(`---
foo: bar
---
`)
	assert.NoError(t, err)
	assert.Equal(t, "foo: bar", frontmatter)
	assert.Equal(t, "", content)

	frontmatter, content, err = SplitFrontmatter(`foo: bar
---
`)
	assert.Equal(t, errBadFrontmatter, err)
}
