package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	// Unfortunately, this package is somewhat difficult to test because
	// there's no way to shut down a server's ListenAndServe. There's not that
	// much code in here so it's not too bad.
}
