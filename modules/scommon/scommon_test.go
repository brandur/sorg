package scommon

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestExtractSlug(t *testing.T) {
	// Article, fragment, etc.
	assert.Equal(t, "hello", ExtractSlug("hello.md"))

	// Newsletter
	assert.Equal(t, "001-hello", ExtractSlug("001-hello.md"))

	// With path included
	assert.Equal(t, "hello", ExtractSlug("path/to/hello.md"))
}

func TestIsDraft(t *testing.T) {
	// Article, fragment, etc.
	assert.False(t, IsDraft("content/hello.md"))
	assert.True(t, IsDraft("content-drafts/hello.md"))
	assert.True(t, IsDraft("drafts/hello.md"))

	// Newsletter
	assert.False(t, IsDraft("content/001-hello.md"))
	assert.True(t, IsDraft("content-drafts/001-hello.md"))

	// Other outliers
	assert.False(t, IsDraft("hello.md"))
}
