package main

import (
	"fmt"
	"os"
	// "testing"

	"github.com/joeshaw/envdecode"
	// assert "github.com/stretchr/testify/require"
)

func init() {
	if err := envdecode.Decode(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding conf from env: %v", err)
		os.Exit(1)
	}
}
