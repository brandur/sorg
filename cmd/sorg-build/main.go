package main

import (
	"bufio"
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
	"github.com/russross/blackfriday"
	"github.com/yosssi/ace"
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
	// GoogleAnalyticsID is the account identifier for Google Analytics to use.
	GoogleAnalyticsID string `env:"GOOGLE_ANALYTICS_ID"`

	// Verbose is whether the program will print debug output as it's running.
	Verbose bool `env:"VERBOSE,default=false"`
}

// FragmentInfo is meta information about a fragment from its YAML frontmatter.
type FragmentInfo struct {
	// Image is an optional image that may be included with a fragment.
	Image string `yaml:"image"`

	// PublishedAt is when the fragment was published.
	PublishedAt *time.Time `yaml:"published_at"`

	// Title is the fragment's title.
	Title string `yaml:"title"`
}

var conf Conf

func main() {
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

		outName := strings.Replace(fragmentInfo.Name(), ".md", "", -1)

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

		if info.Title == "" {
			return fmt.Errorf("No title for fragment: %v", inPath)
		}

		if info.PublishedAt == nil {
			return fmt.Errorf("No publish date for fragment: %v", inPath)
		}

		locals := map[string]string{
			"Content":     string(renderMarkdown([]byte(content))),
			"Image":       info.Image,
			"PublishedAt": info.PublishedAt.Format("Jan 2, 2006"),
			"Title":       info.Title,

			"BodyClass":         "",
			"GoogleAnalyticsID": conf.GoogleAnalyticsID,
			"Release":           sorg.Release,
			"ViewportWidth":     "device-width",
		}

		err = renderView(sorg.LayoutsDir+"main", sorg.ViewsDir+"/fragments/show",
			sorg.TargetFragmentsDir+outName, locals)
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

		outFile.WriteString("/* " + stylesheet + " */\n\n")

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

		outFile.WriteString("\n\n")
	}

	return nil
}

func renderMarkdown(source []byte) []byte {
	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_DASHES
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_USE_XHTML

	extensions := 0
	extensions |= blackfriday.EXTENSION_AUTO_HEADER_IDS
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_HEADER_IDS
	extensions |= blackfriday.EXTENSION_LAX_HTML_BLOCKS
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")
	return blackfriday.Markdown(source, renderer, extensions)
}

func renderView(layout, view, target string, locals map[string]string) error {
	log.Debugf("Rendering: %v", target)

	template, err := ace.Load(layout, view, nil)
	if err != nil {
		return err
	}

	file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	err = template.Execute(writer, locals)
	if err != nil {
		return err
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
