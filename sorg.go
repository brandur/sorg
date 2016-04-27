package sorg

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

const (
	AssetsDir = "./org/assets/"

	FragmentsDir = "./org/fragments/"

	JavascriptsDir = AssetsDir + "javascripts/"

	StylesheetsDir = AssetsDir + "stylesheets/"

	TargetAssetsDir = TargetDir + "assets/"

	// TargetDir is the location where the site will be built to.
	TargetDir = "./public/"

	TargetFragmentsDir = TargetDir + "fragments/"
)

// A list of all directories that are in the built static site.
var targetDirs = []string{
	TargetAssetsDir,
	TargetFragmentsDir,
}

// CreateTargetDir creates TargetDir if it doesn't already exist.
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
