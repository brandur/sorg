package markdown

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	assert.Equal(t, "<p><strong>strong</strong></p>\n", Render("**strong**"))
}
