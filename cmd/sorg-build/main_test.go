package main

import (
	"testing"

	"github.com/brandur/sorg"
	assert "github.com/stretchr/testify/require"
)

func TestGetLocals(t *testing.T) {
	locals := getLocals("Title", map[string]interface{}{
		"Foo": "Bar",
	})

	assert.Equal(t, "Bar", locals["Foo"])
	assert.Equal(t, sorg.Release, locals["Release"])
	assert.Equal(t, "Title", locals["Title"])
}

func TestIsHidden(t *testing.T) {
	assert.Equal(t, true, isHidden(".gitkeep"))
	assert.Equal(t, false, isHidden("article"))
}

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
