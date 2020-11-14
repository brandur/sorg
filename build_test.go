package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/joeshaw/envdecode"
	assert "github.com/stretchr/testify/require"
)

func init() {
	if err := envdecode.Decode(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding conf from env: %v", err)
		os.Exit(1)
	}
}

func TestPagePathKey(t *testing.T) {
	assert.Equal(t, "about", pagePathKey("./pages/about.ace"))
	assert.Equal(t, "about", pagePathKey("./pages-drafts/about.ace"))

	assert.Equal(t, "deep/about", pagePathKey("./pages/deep/about.ace"))
	assert.Equal(t, "deep/about", pagePathKey("./pages-drafts/deep/about.ace"))

	assert.Equal(t, "really/deep/about", pagePathKey("./pages/really/deep/about.ace"))
	assert.Equal(t, "really/deep/about", pagePathKey("./pages-drafts/really/deep/about.ace"))
}
