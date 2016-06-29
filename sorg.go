package sorg

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

const (
	// Release is the asset version of the site. Bump when any assets are
	// updated to blow away any browser caches.
	Release = "1"
)

const (
	// ArticlesDir is the source directory for articles.
	ArticlesDir = "./org/articles/"

	// ArticlesDraftsDir is the source directory for article drafts.
	ArticlesDraftsDir = "./org/drafts/"

	// AssetsDir is the source directory for image, JavaScript, and stylesheet
	// assets.
	AssetsDir = "./org/assets/"

	// FragmentsDir is the source directory for fragments.
	FragmentsDir = "./org/fragments/"

	// ImagesDir is the source for images.
	ImagesDir = AssetsDir + "images/"

	// JavascriptsDir is the source directory for JavaScripts.
	JavascriptsDir = AssetsDir + "javascripts/"

	// LayoutsDir is the source directory for view layouts.
	LayoutsDir = "./layouts/"

	// StylesheetsDir is the source directory for stylesheets.
	StylesheetsDir = AssetsDir + "stylesheets/"

	// TargetArticlesDir is the target directory for articles.
	TargetArticlesDir = TargetDir + "articles/"

	// TargetAssetsDir is the target directory where static assets are placed
	// which should not be versioned by release number. Unversioned assets are
	// those that probably don't need to change between releases like images.
	TargetAssetsDir = TargetDir + "assets/"

	// TargetVersionedAssetsDir is the target directory where static assets are
	// placed which should be versioned by release number. Versioned assets are
	// those that might need to change on release like CSS or JS files.
	TargetVersionedAssetsDir = TargetDir + "assets/" + Release + "/"

	// TargetDir is the target location where the site will be built to.
	TargetDir = "./public/"

	// TargetFragmentsDir is the target directory for fragments.
	TargetFragmentsDir = TargetDir + "fragments/"

	// ViewsDir is the source directory for views.
	ViewsDir = "./views/"
)

// A list of all directories that are in the built static site.
var targetDirs = []string{
	TargetArticlesDir,
	TargetAssetsDir,
	TargetFragmentsDir,
}

// CreateTargetDirs creates TargetDir if it doesn't already exist.
func CreateTargetDirs() error {
	for _, targetDir := range targetDirs {
		err := os.MkdirAll(targetDir, 0755)
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
