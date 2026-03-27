package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/joeshaw/envdecode"
	"github.com/stretchr/testify/require"
)

func init() {
	if err := envdecode.Decode(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding conf from env: %v", err)
		os.Exit(1)
	}
}

func TestExtCanonical(t *testing.T) {
	require.Equal(t, ".jpg", extCanonical("https://example.com/image.jpg"))
	require.Equal(t, ".jpg", extCanonical("https://example.com/image.JPG"))
	require.Equal(t, ".jpg", extCanonical("https://example.com/image.jpg?query"))
}

func TestExtImageTarget(t *testing.T) {
	require.Equal(t, ".jpg", extImageTarget(".jpg"))
	require.Equal(t, ".webp", extImageTarget(".heic"))
}

func TestLexicographicBase32(t *testing.T) {
	// Should only incorporate lower case characters.
	require.Equal(t, lexicographicBase32, strings.ToLower(lexicographicBase32))

	// All characters in the encoding set should be lexicographically ordered.
	{
		// This can be replaced with `strings.Clone` come Go 1.20.
		b := make([]byte, len(lexicographicBase32))
		copy(b, lexicographicBase32)

		slices.Sort(b)
		require.Equal(t, lexicographicBase32, string(b))
	}
}

func TestPagePathKey(t *testing.T) {
	require.Equal(t, "about", pagePathKey("./pages/about.ace"))
	require.Equal(t, "about", pagePathKey("./pages-drafts/about.ace"))

	require.Equal(t, "deep/about", pagePathKey("./pages/deep/about.ace"))
	require.Equal(t, "deep/about", pagePathKey("./pages-drafts/deep/about.ace"))

	require.Equal(t, "really/deep/about", pagePathKey("./pages/really/deep/about.ace"))
	require.Equal(t, "really/deep/about", pagePathKey("./pages-drafts/really/deep/about.ace"))
}

func TestSimplifyMarkdownForSummary(t *testing.T) {
	require.Equal(t, "check that links are removed", simplifyMarkdownForSummary("check that [links](/link) are removed"))
	require.Equal(t, "double new lines are gone", simplifyMarkdownForSummary("double new\n\nlines are gone"))
	require.Equal(t, "single new lines are gone", simplifyMarkdownForSummary("single new\nlines are gone"))
	require.Equal(t, "space is trimmed", simplifyMarkdownForSummary(" space is trimmed "))
}

func TestTruncateString(t *testing.T) {
	require.Equal(t, "Short string unchanged.", truncateString("Short string unchanged.", 100))

	exactly100Length := strings.Repeat("s", 100)
	require.Equal(t, exactly100Length, truncateString(exactly100Length, 100))

	require.Equal(t,
		"This is a longer string that's going to need truncation and which will be truncated by ending it w …",
		truncateString("This is a longer string that's going to need truncation and which will be truncated by ending it with a space and an ellipsis.", 100),
	)
}
