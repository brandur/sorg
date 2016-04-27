package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/brandur/sorg"
	"github.com/joeshaw/envdecode"
	"github.com/yosssi/gcss"
	"gopkg.in/yaml.v2"
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

// FragmentInfo is meta information about a fragment from its YAML frontmatter.
type FragmentInfo struct {
	// Image is an optional image that may be included with a fragment.
	Image string `yaml:"image"`

	// PublishedAt is when the fragment was published.
	PublishedAt time.Time `yaml:"published_at"`

	// Title is the fragment's title.
	Title string `yaml:"title"`
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

		raw, err := ioutil.ReadFile(inPath)
		if err != nil {
			return err
		}

		frontmatter, content, err := splitFrontmatter(string(raw))
		if err != nil {
			return err
		}

		var info FragmentInfo
		err = yaml.Unmarshal([]byte(frontmatter), &info)
		if err != nil {
			return err
		}
		log.Infof("Info = %+v", info)

		outFile, err := os.Create(sorg.TargetFragmentsDir + outName)
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = outFile.WriteString(content)
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
				return fmt.Errorf("Error compiling %v: %v", inPath, err)
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

var errBadFrontmatter = fmt.Errorf("Unable to split YAML frontmatter")

func splitFrontmatter(content string) (string, string, error) {
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
