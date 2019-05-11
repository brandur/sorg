package modulir

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Args are the set of arguments accepted by NewContext.
type Args struct {
	Concurrency int
	Log         LoggerInterface
	Pool        *Pool
	Port        int
	SourceDir   string
	TargetDir   string
	Watcher     *fsnotify.Watcher
}

// Context contains useful state that can be used by a user-provided build
// function.
type Context struct {
	// Concurrency is the number of concurrent workers to run during the build
	// step.
	Concurrency int

	// FirstRun indicates whether this is the first run of the build loop.
	FirstRun bool

	// Forced causes the Changed function to always return true regardless of
	// the path it's invoked on, thereby prompting all jobs that use it to
	// execute.
	//
	// Make sure to unset it after your build run is finished.
	Forced bool

	// Jobs is a channel over which jobs to be done are transmitted.
	Jobs chan *Job

	// Log is a logger that can be used to print information.
	Log LoggerInterface

	// Pool is the job pool used to build the static site.
	Pool *Pool

	// Port specifies the port on which to serve content from TargetDir over
	// HTTP.
	Port int

	// QuickPaths are a set of paths for which Changed will return true when
	// the context is in "quick rebuild mode". During this time all the normal
	// file system checks that Changed makes will be bypassed to enable a
	// faster build loop.
	//
	// Make sure that all paths added here are normalized with filepath.Clean.
	//
	// Make sure to unset it after your build run is finished.
	QuickPaths map[string]struct{}

	// SourceDir is the directory containing source files.
	SourceDir string

	// Stats tracks various statistics about the build process.
	//
	// Statistics are reset between build loops, but are cumulative between
	// build phases within a loop (i.e. calls to Wait).
	Stats *Stats

	// TargetDir is the directory where the site will be built to.
	TargetDir string

	// Watcher is a file system watcher that picks up changes to source files
	// and restarts the build loop.
	Watcher *fsnotify.Watcher

	// fileModTimeCache remembers the last modified times of files.
	fileModTimeCache *FileModTimeCache
}

// NewContext initializes and returns a new Context.
func NewContext(args *Args) *Context {
	c := &Context{
		Concurrency: args.Concurrency,
		FirstRun:    true,
		Log:         args.Log,
		Pool:        args.Pool,
		Port:        args.Port,
		SourceDir:   args.SourceDir,
		Stats:       &Stats{},
		TargetDir:   args.TargetDir,
		Watcher:     args.Watcher,

		fileModTimeCache: NewFileModTimeCache(args.Log),
	}

	if args.Pool != nil {
		c.Jobs = args.Pool.Jobs
	}

	return c
}

// AddJob is a shortcut for adding a new job to the Jobs channel.
func (c *Context) AddJob(name string, f func() (bool, error)) {
	c.Jobs <- NewJob(name, f)
}

// AllowError is a helper that's useful for when an error coming back from a
// job should be logged, but shouldn't fail the build.
func (c *Context) AllowError(executed bool, err error) bool {
	if err != nil {
		c.Log.Errorf("Error allowed: %v", err)
	}
	return executed
}

// Changed returns whether the target path's modified time has changed since
// the last time it was checked. It also saves the last modified time for
// future checks.
//
// This function is very hot in that it gets checked many times, and probably
// many times for every single job in a build loop. It needs to be optimized
// fairly carefully for both speed and lack of contention when running
// concurrently with other jobs.
func (c *Context) Changed(path string) bool {
	// Always return immediately if the context has been forced.
	if c.Forced {
		return true
	}

	// Make sure we're always operating against a normalized path.
	//
	// Note that fsnotify sends us cleaned paths which are what gets added to
	// QuickPaths below, so cleaning here ensures that we're always comparing
	// against the right thing.
	path = filepath.Clean(path)

	// Short circuit quickly if the context is in "quick rebuild mode".
	if c.QuickPaths != nil {
		_, ok := c.QuickPaths[path]
		return ok
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			c.Log.Errorf("Path passed to Changed doesn't exist: %s", path)
		}
		return true
	}

	changed, ok := c.fileModTimeCache.isFileUpdated(fileInfo, path)

	// If we got ok back, then we know the file was in the cache and also
	// therefore would've been already watched. Return as early as possible.
	if ok {
		return changed
	}

	if c.Watcher != nil {
		err := c.addWatched(fileInfo, path)
		if err != nil {
			c.Log.Errorf("Error watching source: %v", err)
		}
	}

	return true
}

// ChangedAny is the same as Changed except it returns true if any of the given
// paths have changed.
func (c *Context) ChangedAny(paths ...string) bool {
	// We have to run through every element in paths even if we detect changed
	// early so that each is correctly added to the file mod time cache and
	// watched.
	changed := false

	for _, path := range paths {
		// Make sure that c.Changed appears first or there seems to be a danger
		// of Go compiling it out.
		changed = c.Changed(path) || changed
	}

	return changed
}

// ResetBuild signals to the Context to do the bookkeeping it needs to do for
// the next build round.
func (c *Context) ResetBuild() {
	c.Stats.Reset()
	c.fileModTimeCache.promote()
}

// Wait waits on the job pool to execute its current round of jobs.
//
// The worker pool is then primed for a new round so that more jobs can be
// enqueued.
//
// Returns nil if the round of jobs executed successfully, and a set of errors
// that occurred otherwise.
func (c *Context) Wait() []error {
	c.Stats.LoopDuration =
		c.Stats.LoopDuration + time.Now().Sub(c.Stats.lastLoopStart)

	defer func() {
		// Reset the last loop start.
		c.Stats.lastLoopStart = time.Now()
	}()

	// Wait for work to finish.
	c.Pool.Wait()

	c.Stats.JobsExecuted = append(c.Stats.JobsExecuted, c.Pool.JobsExecuted...)
	c.Stats.NumJobs += len(c.Pool.JobsAll)
	c.Stats.NumJobsErrored += len(c.Pool.JobsErrored)
	c.Stats.NumJobsExecuted += len(c.Pool.JobsExecuted)

	// Pull errors out before starting a new round below.
	errors := c.Pool.JobErrors()

	// Then start the pool again, which also has the side effect of
	// reinitializing anything that needs to be reinitialized.
	c.Pool.StartRound()

	// This channel is reinitialized, so make sure to pull in the new one.
	c.Jobs = c.Pool.Jobs

	return errors
}

func (c *Context) addWatched(fileInfo os.FileInfo, absolutePath string) error {
	// Watch the parent directory unless the file is a directory itself. This
	// will hopefully mean fewer individual entries in the notifier.
	if !fileInfo.IsDir() {
		absolutePath = filepath.Dir(absolutePath)
	}

	return c.Watcher.Add(absolutePath)
}

// FileModTimeCache tracks the last modified time of files seen so a
// determination can be made as to whether they need to be recompiled.
type FileModTimeCache struct {
	log                 LoggerInterface
	mu                  sync.Mutex
	pathToModTimeMap    map[string]time.Time
	pathToModTimeMapNew map[string]time.Time
}

// NewFileModTimeCache returns a new FileModTimeCache.
func NewFileModTimeCache(log LoggerInterface) *FileModTimeCache {
	return &FileModTimeCache{
		log:                 log,
		pathToModTimeMap:    make(map[string]time.Time),
		pathToModTimeMapNew: make(map[string]time.Time),
	}
}

// changed returns whether the target path's modified time has changed since
// the last time it was checked. It also saves the last modified time for
// future checks. The second return value is whether or not the record was
// already in the cache.
func (c *FileModTimeCache) isFileUpdated(fileInfo os.FileInfo, absolutePath string) (bool, bool) {
	modTime := fileInfo.ModTime()

	lastModTime, ok := c.pathToModTimeMap[absolutePath]

	if ok {
		changed := lastModTime.Before(modTime)
		if !changed {
			return false, ok
		}
	}

	// Store to the new map for eventual promotion.
	c.mu.Lock()
	c.pathToModTimeMapNew[absolutePath] = modTime
	c.mu.Unlock()

	return true, ok
}

// promote takes all the new modification times collected during this round
// (i.e. a build phase) and promotes them into the main map so that they're
// available for the next one.
func (c *FileModTimeCache) promote() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Promote all new values to the current map.
	for path, modTime := range c.pathToModTimeMapNew {
		c.pathToModTimeMap[path] = modTime
	}

	// Clear the new map for the next round.
	c.pathToModTimeMapNew = make(map[string]time.Time)
}

// Stats tracks various statistics about the build process.
type Stats struct {
	// JobsExecuted is a slice of jobs that were executed on the last run.
	JobsExecuted []*Job

	// LoopDuration is the total amount of time spent in the user's build loop
	// enqueuing jobs. Jobs may be running in the background during this time,
	// but all the time spent waiting for jobs to finish is excluded.
	LoopDuration time.Duration

	// NumJobs is the total number of jobs generated for the build loop.
	NumJobs int

	// NumJobsErrored is the number of jobs that errored during the build loop.
	//
	// Note that if any errors were present, the build loop may have cancelled
	// early as it didn't move onto its later phases, which will lead to
	// commensurate fewer jobs.
	NumJobsErrored int

	// NumJobsExecuted is the number of jobs that did some kind of heavier
	// lifting during the build loop. That's those that returned `true` on
	// execution.
	NumJobsExecuted int

	// Start is the start time of the build loop.
	Start time.Time

	// lastLoopStart is when the last user build loop started (i.e. this is set
	// to the current timestamp whenever a call to context.Wait finishes).
	lastLoopStart time.Time
}

// Reset resets statistics.
func (s *Stats) Reset() {
	s.JobsExecuted = nil
	s.LoopDuration = time.Duration(0)
	s.NumJobs = 0
	s.NumJobsErrored = 0
	s.NumJobsExecuted = 0
	s.Start = time.Now()
	s.lastLoopStart = time.Now()
}
