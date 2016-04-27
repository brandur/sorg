package main

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestSplitFrontmatter(t *testing.T) {
	frontmatter, content, err := splitFrontmatter(`---
foo: bar
---

other`)
	assert.NoError(t, err)
	assert.Equal(t, "foo: bar", frontmatter)
	assert.Equal(t, "other", content)

	frontmatter, content, err = splitFrontmatter(`other`)
	assert.NoError(t, err)
	assert.Equal(t, "", frontmatter)
	assert.Equal(t, "other", content)

	frontmatter, content, err = splitFrontmatter(`---
foo: bar
---
`)
	assert.NoError(t, err)
	assert.Equal(t, "foo: bar", frontmatter)
	assert.Equal(t, "", content)

	frontmatter, content, err = splitFrontmatter(`foo: bar
---
`)
	assert.Equal(t, errBadFrontmatter, err)
}
