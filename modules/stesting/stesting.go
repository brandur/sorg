package stesting

import (
	"os"
	"path"
	"runtime"
)

func init() {
	// No longer relevant as I migrated to Modulir, but we may need to bring
	// something like this back in, so commented out for now.
	/*
		// Initializes logrus logging if tests are run with `go test -v`. Note that
		// command line flags need to be parsed before testing.Verbose() becomes
		// available.
		flag.Parse()
		if testing.Verbose() {
			sorg.InitLog(true)
		}
	*/

	// Move up into the project's root so that we in the right place relative
	// to content/view/layout/etc. directories. The invocation to runtime
	// returns *this* file (`testing.go`), and we can then trace up to the
	// project's root directory no matter what package is being tested (tests
	// have their CWD set to the project's path).
	_, filename, _, _ := runtime.Caller(0)
	path.Join(path.Dir(filename), "..")
	err := os.Chdir("../../")
	if err != nil {
		panic(err)
	}
}
