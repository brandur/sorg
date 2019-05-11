package modulir

import (
	"testing"

	assert "github.com/stretchr/testify/require"
	"github.com/fsnotify/fsnotify"
)

func TestShouldRebuild(t *testing.T) {
	// Most things signal a rebuild
	assert.Equal(t, true, shouldRebuild("a/path", fsnotify.Create))
	assert.Equal(t, true, shouldRebuild("a/path", fsnotify.Remove))
	assert.Equal(t, true, shouldRebuild("a/path", fsnotify.Write))

	// With just a few special cases that don't
	assert.Equal(t, false, shouldRebuild("a/path", fsnotify.Chmod))
	assert.Equal(t, false, shouldRebuild("a/path~", fsnotify.Create))
}
