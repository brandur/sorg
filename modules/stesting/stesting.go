package stesting

import (
	"os"
	"path"
	"runtime"

	"github.com/brandur/modulir"
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

	// Here's a similar thing from toc_test.go. Not sure if we still something
	// like it:
	/*
		// We override TestMain so that we can control the logging level of logrus. If
		// we did nothing it would spit out a lot of output even during a normal test
		// run. This way it's only verbose if using `go test -v`.
		func TestMain(m *testing.M) {
			// we depend on flags so we need to call Parse explicitly (it's otherwise
			// done implicitly inside Run)
			flag.Parse()

			log.SetFormatter(new(log.TextFormatter))

			if testing.Verbose() {
				log.SetLevel(log.DebugLevel)
			} else {
				log.SetLevel(log.FatalLevel)
			}

			os.Exit(m.Run())
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

// NewContext is a convenience helper to create a new modulir.Context suitable
// for use in the test suite.
func NewContext() *modulir.Context {
	return modulir.NewContext(&modulir.Args{Log: &modulir.Logger{Level: modulir.LevelInfo}})
}
