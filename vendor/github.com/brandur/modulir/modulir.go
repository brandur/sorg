package modulir

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Public
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Config contains configuration.
type Config struct {
	// Concurrency is the number of concurrent workers to run during the build
	// step.
	//
	// Defaults to 10.
	Concurrency int

	// Log specifies a logger to use.
	//
	// Defaults to an instance of Logger running at informational level.
	Log LoggerInterface

	// Port specifies the port on which to serve content from TargetDir over
	// HTTP.
	//
	// Defaults to not running if left unset.
	Port int

	// SourceDir is the directory containing source files.
	//
	// Defaults to ".".
	SourceDir string

	// TargetDir is the directory where the site will be built to.
	//
	// Defaults to "./public".
	TargetDir string
}

// Build is one of the main entry points to the program. Call this to build
// only one time.
func Build(config *Config, f func(*Context) []error) {
	finish := make(chan struct{}, 1)
	firstRunComplete := make(chan struct{}, 1)

	// Signal the build loop to finish immediately
	finish <- struct{}{}

	c := initContext(config, nil)
	success := build(c, f, finish, firstRunComplete)
	if !success {
		os.Exit(1)
	}
}

// BuildLoop is one of the main entry points to the program. Call this to build
// in a perpetual loop.
func BuildLoop(config *Config, f func(*Context) []error) {
	finish := make(chan struct{}, 1)
	firstRunComplete := make(chan struct{}, 1)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		exitWithError(errors.Wrap(err, "Error starting watcher"))
		os.Exit(1)
	}
	defer watcher.Close()

	c := initContext(config, watcher)

	// Serve HTTP
	var server *http.Server
	go func() {
		<-firstRunComplete
		server = startServingTargetDirHTTP(c)
	}()

	// Run the build loop. Loops forever until receiving on finish.
	go build(c, f, finish, firstRunComplete)

	// Listen for signals
	signals := make(chan os.Signal, 1024)
	signal.Notify(signals, unix.SIGUSR2)
	for {
		s := <-signals
		switch s {
		case unix.SIGUSR2:
			shutdownAndExec(c, finish, watcher, server)
		}
	}
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private
//
//
//
//////////////////////////////////////////////////////////////////////////////

const (
	// Maximum number of errors or jobs to print on screen after a build loop.
	maxMessages = 10
)

// Runs an infinite built loop until a signal is received over the `finish`
// channel.
//
// Returns true of the last build was successful and false otherwise.
func build(c *Context, f func(*Context) []error, finish, firstRunComplete chan struct{}) bool {
	rebuild := make(chan string)
	rebuildDone := make(chan struct{})

	if c.Watcher != nil {
		go watchChanges(c, c.Watcher, rebuild, rebuildDone)
	}

	c.Pool.StartRound()
	c.Jobs = c.Pool.Jobs

	// A path that changed on the last loop (as discovered via fsnotify). If
	// set, we go into quick build mode with only this path activated, and
	// unset it afterwards. This saves us doing lots of checks on the
	// filesystem and makes jobs much faster to run.
	var lastChangedPath string

	for {
		c.Log.Debugf("Start loop")
		c.ResetBuild()

		if lastChangedPath != "" {
			c.QuickPaths = map[string]struct{}{lastChangedPath: {}}
		}

		errors := f(c)

		otherErrors := c.Wait()
		buildDuration := time.Now().Sub(c.Stats.Start)

		if otherErrors != nil {
			errors = append(errors, otherErrors...)
		}

		logErrors(c, errors)
		logSlowestJobs(c)

		c.Log.Infof("Built site in %s (%v / %v job(s) did work; %v errored; loop took %v)",
			buildDuration,
			c.Stats.NumJobsExecuted, c.Stats.NumJobs, c.Stats.NumJobsErrored,
			c.Stats.LoopDuration)

		if c.FirstRun {
			firstRunComplete <- struct{}{}
			c.FirstRun = false
		} else {
			rebuildDone <- struct{}{}
		}

		lastChangedPath = ""
		c.QuickPaths = nil

		select {
		case <-finish:
			c.Log.Infof("Detected finish signal; stopping")
			return len(errors) < 1

		case lastChangedPath = <-rebuild:
			c.Log.Infof("Detected change on '%s'; rebuilding", lastChangedPath)
		}
	}
}

// Exits with status 1 after printing the given error to stderr.
func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

// Takes a Modulir configuration and initializes it with defaults for any
// properties that weren't expressly filled in.
func initConfigDefaults(config *Config) *Config {
	if config == nil {
		config = &Config{}
	}

	if config.Concurrency <= 0 {
		config.Concurrency = 50
	}

	if config.Log == nil {
		config.Log = &Logger{Level: LevelInfo}
	}

	if config.SourceDir == "" {
		config.SourceDir = "."
	}

	if config.TargetDir == "" {
		config.TargetDir = "./public"
	}

	return config
}

// Initializes a new Modulir context from the given configuration.
func initContext(config *Config, watcher *fsnotify.Watcher) *Context {
	config = initConfigDefaults(config)

	return NewContext(&Args{
		Log:       config.Log,
		Port:      config.Port,
		Pool:      NewPool(config.Log, config.Concurrency),
		SourceDir: config.SourceDir,
		TargetDir: config.TargetDir,
		Watcher:   watcher,
	})
}

func logErrors(c *Context, errors []error) {
	if errors == nil {
		return
	}

	for i, err := range errors {
		c.Log.Errorf("Build error: %v", err)

		if i >= maxMessages-1 {
			c.Log.Errorf("... too many errors (limit reached)")
			break
		}
	}
}

func logSlowestJobs(c *Context) {
	sortJobsBySlowest(c.Stats.JobsExecuted)

	for i, job := range c.Stats.JobsExecuted {
		// Having this in the loop ensures we don't print it if zero jobs
		// executed
		if i == 0 {
			c.Log.Infof("Jobs executed (slowest first):")
		}

		c.Log.Infof("    %s (time: %v)", job.Name, job.Duration)

		if i >= maxMessages-1 {
			c.Log.Infof("... many jobs executed (limit reached)")
			break
		}
	}
}

// Decides whether a rebuild should be triggered given some input event
// properties from fsnotify.
func shouldRebuild(path string, op fsnotify.Op) bool {
	// A special case, but ignore creates on files that look like Vim backups.
	if strings.HasSuffix(path, "~") {
		return false
	}

	if op&fsnotify.Chmod == fsnotify.Chmod {
		return false
	}

	return true
}

func shutdownAndExec(c *Context, finish chan struct{},
	watcher *fsnotify.Watcher, server *http.Server) {

	// Tell the build loop to finish up
	finish <- struct{}{}

	// DANGER: Defers don't seem to get called on the re-exec, so even though
	// we have a defer which closes our watcher, it won't close, leading to
	// file descriptor leaking. Close it manually here instead.
	watcher.Close()

	// A context that will act as a timeout for connections
	// that are still running as we try and shut down the HTTP
	// server.
	timeoutCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	c.Log.Infof("Shutting down HTTP server")
	if err := server.Shutdown(timeoutCtx); err != nil {
		exitWithError(err)
	}

	// Returns an absolute path.
	execPath, err := os.Executable()
	if err != nil {
		exitWithError(err)
	}

	c.Log.Infof("Execing process '%s' with args %+v\n", execPath, os.Args)
	if err := unix.Exec(execPath, os.Args, os.Environ()); err != nil {
		exitWithError(err)
	}
}

// Sorts a slice of jobs with the slowest on top.
func sortJobsBySlowest(jobs []*Job) {
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[j].Duration < jobs[i].Duration
	})
}

func startServingTargetDirHTTP(c *Context) *http.Server {
	c.Log.Infof("Serving '%s' to: http://localhost:%v/", path.Clean(c.TargetDir), c.Port)

	handler := http.FileServer(http.Dir(c.TargetDir))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", c.Port),
		Handler: handler,
	}

	go func() {
		err := server.ListenAndServe()

		// ListenAndServe always returns a non-nil error
		if err != http.ErrServerClosed {
			exitWithError(errors.Wrap(err, "Error starting HTTP server"))
		}
	}()

	return server
}

// Listens for file system changes from fsnotify and pushes relevant ones back
// out over the rebuild channel.
//
// It doesn't start listening to fsnotify again until the main loop has
// signaled rebuildDone, so there is a possibility that in the case of very
// fast consecutive changes the build might not be perfectly up to date.
func watchChanges(c *Context, watcher *fsnotify.Watcher,
	rebuild chan string, rebuildDone chan struct{}) {
OUTER:
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			c.Log.Debugf("Received event from watcher: %+v", event)

			if !shouldRebuild(event.Name, event.Op) {
				continue
			}

			// Start rebuild
			rebuild <- event.Name

			// Wait until rebuild is finished. In the meantime, drain any
			// new events that come in on the watcher's channel.
			for {
				select {
				case <-rebuildDone:
					continue OUTER
				case <-watcher.Events:
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			c.Log.Errorf("Error from watcher:", err)
		}
	}
}
