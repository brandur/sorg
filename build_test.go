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
