package main

import (
	"io"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/brandur/sorg"
	"github.com/joeshaw/envdecode"
	"github.com/yosssi/gcss"
)

var stylesheets = []string{
	"_reset.sass",
	"main.sass",
	"about.sass",
	"fragments.sass",
	"index.sass",
	"photos.sass",
	"quotes.sass",
	"reading.sass",
	"runs.sass",
	"signature.sass",
	"solarized-light.css",
	"tenets.sass",
	"twitter.sass",
}

// Conf contains configuration information for the command.
type Conf struct {
	// Verbose is whether the program will print debug output as it's running.
	Verbose bool `env:"VERBOSE,default=false"`
}

func main() {
	var conf Conf
	err := envdecode.Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	sorg.InitLog(conf.Verbose)

	err = sorg.CreateTargetDirs()
	if err != nil {
		log.Fatal(err)
	}

	err = compileFragments()
	if err != nil {
		log.Fatal(err)
	}

	err = compileStylesheets()
	if err != nil {
		log.Fatal(err)
	}
}

func compileFragments() error {
	fragmentInfos, err := ioutil.ReadDir(sorg.FragmentsDir)
	if err != nil {
		return err
	}

	for _, fragmentInfo := range fragmentInfos {
		inPath := sorg.FragmentsDir + fragmentInfo.Name()
		log.Debugf("Compiling: %v", inPath)

		outName := strings.Replace(fragmentInfo.Name(), ".md", ".html", -1)

		inFile, err := os.Open(inPath)
		if err != nil {
			return err
		}

		outFile, err := os.Create(sorg.TargetFragmentsDir + outName)
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, inFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func compileStylesheets() error {
	outFile, err := os.Create(sorg.TargetAssetsDir + "app.css")
	if err != nil {
		return err
	}
	defer outFile.Close()

	for _, stylesheet := range stylesheets {
		inPath := sorg.StylesheetsDir + stylesheet
		log.Debugf("Compiling: %v", inPath)

		inFile, err := os.Open(inPath)
		if err != nil {
			return err
		}

		if strings.HasSuffix(stylesheet, ".sass") {
			_, err := gcss.Compile(outFile, inFile)
			if err != nil {
				return err
			}
		} else {
			_, err := io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
