package modulir

import (
	"fmt"
	"os"
	"sort"
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
			c.QuickPaths = map[string]struct{}{lastChangedPath: struct{}{}}
		}

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
					c.Log.Errorf("... too many errors (limit reached)")
					break
				}
			}
		}

		sortJobsBySlowest(c.Stats.JobsExecuted)
		for i, job := range c.Stats.JobsExecuted {
			// Having this in the loop ensures we don't print it if zero jobs
			// executed
			if i == 0 {
				c.Log.Infof("Jobs executed (slowest first):")
			}

			c.Log.Infof("    %s (time: %v)", job.Name, job.Duration)

			if i >= 9 {
				c.Log.Infof("... many jobs executed (limit reached)")
				break
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

		lastChangedPath = ""
		c.QuickPaths = nil

		select {
		case <-finish:
			c.Log.Infof("Detected finish signal; stopping")
			return len(errors) < 1

		case lastChangedPath =<-rebuild:
			c.Log.Infof("Detected change on '%s'; rebuilding", lastChangedPath)
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

func sortJobsBySlowest(jobs []*Job) {
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[j].Duration < jobs[i].Duration
	})
}
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
