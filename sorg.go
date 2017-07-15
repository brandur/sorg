package sorg

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	// AbsoluteURL is the site's absolute URL. It's usually preferable that
	// it's not used, but it is when generating emails.
	AbsoluteURL = "https://brandur.org"

	// ContentDir is the location of the site's content (articles, fragments,
	// assets, etc.).
	ContentDir = "./content"

	// LayoutsDir is the source directory for view layouts.
	LayoutsDir = "./layouts"

	// MainLayout is the site's main layout.
	MainLayout = LayoutsDir + "/main"

	// PagesDir is the source directory for one-off page content.
	PagesDir = "./pages"

	// PassageLayout is the layout for a Passages & Glass issue (an email
	// newsletter).
	PassageLayout = LayoutsDir + "/passages"

	// ViewsDir is the source directory for views.
	ViewsDir = "./views"
)

var errBadFrontmatter = fmt.Errorf("Unable to split YAML frontmatter")

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
	"passages",
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
		dir = path.Join(targetDir, dir)
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

// SplitFrontmatter takes content that contains a combination and YAML metadata
// and content, and splits it into its components.
func SplitFrontmatter(content string) (string, string, error) {
	parts := regexp.MustCompile("(?m)^---").Split(content, 3)

	if len(parts) > 1 && parts[0] != "" {
		return "", "", errBadFrontmatter
	} else if len(parts) == 2 {
		return "", strings.TrimSpace(parts[1]), nil
	} else if len(parts) == 3 {
		return strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2]), nil
	}

	return "", strings.TrimSpace(parts[0]), nil
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
