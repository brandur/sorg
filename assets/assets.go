package assets

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/yosssi/gcss"
)

// CompileJavascripts compiles a set of JS files into a single large file by
// appending them all to each other. Files are appended in alphabetical order
// so we depend on the fact that there aren't too many interdependencies
// between files. A common requirement can be given an underscore prefix to be
// loaded first.
func CompileJavascripts(inPath, outPath string) error {
	start := time.Now()
	defer func() {
		log.Debugf("Compiled script assets in %v.", time.Now().Sub(start))
	}()

	log.Debugf("Building: %v", outPath)

	javascriptInfos, err := ioutil.ReadDir(inPath)
	if err != nil {
		return err
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	for _, javascriptInfo := range javascriptInfos {
		if isHidden(javascriptInfo.Name()) {
			continue
		}

		log.Debugf("Including: %v", javascriptInfo.Name())

		inFile, err := os.Open(path.Join(inPath, javascriptInfo.Name()))
		if err != nil {
			return err
		}

		outFile.WriteString("/* " + javascriptInfo.Name() + " */\n\n")
		outFile.WriteString("(function() {\n\n")

		_, err = io.Copy(outFile, inFile)
		if err != nil {
			return err
		}

		outFile.WriteString("\n\n")
		outFile.WriteString("}).call(this);\n\n")
	}

	return nil
}

// CompileStylesheets compiles a set of stylesheet files into a single large
// file by appending them all to each other. Files are appended in alphabetical
// order so we depend on the fact that there aren't too many interdependencies
// between files. CSS reset in particular is given an underscore prefix so that
// it gets to load first.
//
// If a file has a ".sass" suffix, we attempt to render it as GCSS. This isn't
// a perfect symmetry, but works well enough for these cases.
func CompileStylesheets(inPath, outPath string) error {
	start := time.Now()
	defer func() {
		log.Debugf("Compiled stylesheet assets in %v.", time.Now().Sub(start))
	}()

	log.Debugf("Building: %v", outPath)

	stylesheetInfos, err := ioutil.ReadDir(inPath)
	if err != nil {
		return err
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	for _, stylesheetInfo := range stylesheetInfos {
		if isHidden(stylesheetInfo.Name()) {
			continue
		}

		log.Debugf("Including: %v", stylesheetInfo.Name())

		inFile, err := os.Open(path.Join(inPath, stylesheetInfo.Name()))
		if err != nil {
			return err
		}

		outFile.WriteString("/* " + stylesheetInfo.Name() + " */\n\n")

		if strings.HasSuffix(stylesheetInfo.Name(), ".sass") {
			_, err := gcss.Compile(outFile, inFile)
			if err != nil {
				return fmt.Errorf("Error compiling %v: %v",
					stylesheetInfo.Name(), err)
			}
		} else {
			_, err := io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}

		outFile.WriteString("\n\n")
	}

	return nil
}

// Detects a hidden file, i.e. one that starts with a dot.
func isHidden(file string) bool {
	return strings.HasPrefix(file, ".")
}
