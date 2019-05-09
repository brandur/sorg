package myaml

// Hopefully the beginnings of getting some testing started.
/*
import (
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/require"
)

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
*/
