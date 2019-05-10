package modulir

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
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

	go func() {
		<-firstRunComplete
		if err := serveTargetDirHTTP(c); err != nil {
			exitWithError(err)
		}
	}()

	build(c, f, finish, firstRunComplete)
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

// Runs an infinite built loop until a signal is received over the `finish`
// channel.
//
// Returns true of the last build was successful and false otherwise.
func build(c *Context, f func(*Context) []error, finish, firstRunComplete chan struct{}) bool {
	rebuild := make(chan struct{})
	rebuildDone := make(chan struct{})

	if c.Watcher != nil {
		go watchChanges(c, c.Watcher, rebuild, rebuildDone)
	}

	c.Pool.Init()
	c.Pool.StartRound()
	c.Jobs = c.Pool.Jobs

	for {
		c.Log.Debugf("Start loop")
		c.ResetBuild()

		errors := f(c)

		otherErrors := c.Wait()
		buildDuration := time.Now().Sub(c.Stats.Start)

		if otherErrors != nil {
			errors = append(errors, otherErrors...)
		}

		if errors != nil {
			for i, err := range errors {
				c.Log.Errorf("Build error: %v", err)

				if i >= 9 {
					c.Log.Errorf("... too many errors (scroll stopping)")
					break
				}
			}
		}

		if !c.FirstRun {
			// We can expect pretty much everything to have ran on the first
			// run, so only print executed jobs on subsequent runs.
			for i, job := range c.Stats.JobsExecuted {
				c.Log.Infof("Executed job: %s (time: %v)", job.Name, job.Duration)

				if i >= 9 {
					c.Log.Infof("... many jobs executed (scroll stopping)")
					break
				}
			}
		}

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

		select {
		case <-finish:
			c.Log.Infof("Detected finish signal; stopping")
			return len(errors) < 1

		case <-rebuild:
			c.Log.Infof("Detected change; rebuilding")
		}
	}
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func initConfigDefaults(config *Config) *Config {
	if config == nil {
		config = &Config{}
	}

	if config.Concurrency <= 0 {
		config.Concurrency = 10
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

func initContext(config *Config, watcher *fsnotify.Watcher) *Context {
	config = initConfigDefaults(config)

	pool := NewPool(config.Log, config.Concurrency)

	c := NewContext(&Args{
		Log:       config.Log,
		Port:      config.Port,
		Pool:      pool,
		SourceDir: config.SourceDir,
		TargetDir: config.TargetDir,
		Watcher:   watcher,
	})

	return c
}

func shouldRebuild(path string, op fsnotify.Op) bool {
	// A special case, but ignore creates on files that look like Vim backups.
	if strings.HasSuffix(path, "~") && op&fsnotify.Create == fsnotify.Create {
		return false
	}

	if op&fsnotify.Chmod == fsnotify.Chmod {
		return false
	}

	return true
}

func watchChanges(c *Context, watcher *fsnotify.Watcher, rebuild, rebuildDone chan struct{}) {
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
			rebuild <- struct{}{}

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
