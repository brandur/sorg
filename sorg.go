package sorg

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	// Release is the asset version of the site. Bump when any assets are
	// updated to blow away any browser caches.
	Release = "9"
)

const (
	// ContentDir is the location of the site's content (articles, fragments,
	// assets, etc.).
	ContentDir = "./content"

	// LayoutsDir is the source directory for view layouts.
	LayoutsDir = "./layouts"

	// MainLayout is the site's main layout.
	MainLayout = LayoutsDir + "/main"

	// PagesDir is the source directory for one-off page content.
	PagesDir = "./pages"

	// ViewsDir is the source directory for views.
	ViewsDir = "./views"
)

// A list of all directories that are in the built static site.
var outputDirs = []string{
	".",
	"articles",
	"assets",
	"assets/" + Release,
	"assets/photos",
	"fragments",
	"photos",
	"reading",
	"runs",
	"twitter",
}

// CreateOutputDirs creates a target directory for the static site and all
// other necessary directories for the build if they don't already exist.
func CreateOutputDirs(targetDir string) error {
	start := time.Now()
	defer func() {
		log.Debugf("Created target directories in %v.", time.Now().Sub(start))
	}()

	for _, dir := range outputDirs {
		dir = targetDir + "/" + dir
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

// InitLog initializes logging for singularity programs.
func InitLog(verbose bool) {
	log.SetFormatter(&plainFormatter{})

	if verbose {
		log.SetLevel(log.DebugLevel)
	}
}

// plainFormatter is a logrus formatter that displays text in a much more
// simple fashion that's more suitable as CLI output.
type plainFormatter struct {
}

// Format takes a logrus.Entry and returns bytes that are suitable for log
// output.
func (f *plainFormatter) Format(entry *log.Entry) ([]byte, error) {
	bytes := []byte(entry.Message + "\n")

	if entry.Level == log.DebugLevel {
		bytes = append([]byte("DEBUG: "), bytes...)
	}

	return bytes, nil
}
