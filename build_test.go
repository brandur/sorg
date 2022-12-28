package main

import (
	"fmt"
	"os"
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

func TestPagePathKey(t *testing.T) {
	require.Equal(t, "about", pagePathKey("./pages/about.ace"))
	require.Equal(t, "about", pagePathKey("./pages-drafts/about.ace"))

	require.Equal(t, "deep/about", pagePathKey("./pages/deep/about.ace"))
	require.Equal(t, "deep/about", pagePathKey("./pages-drafts/deep/about.ace"))

	require.Equal(t, "really/deep/about", pagePathKey("./pages/really/deep/about.ace"))
	require.Equal(t, "really/deep/about", pagePathKey("./pages-drafts/really/deep/about.ace"))
}
