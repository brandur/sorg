package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

// ArticleInfo is meta information about an article from its YAML frontmatter.
type ArticleInfo struct {
	// Attributions are any attributions for content that may be included in
	// the article (like an image in the header for example).
	Attributions string `yaml:"attributions"`

	// Content is the HTML content of the article. It isn't included as YAML
	// frontmatter, and is rather split out of an article's Markdown file,
	// rendered, and then added separately.
	Content string `yaml:"-"`

	// HNLink is an optional link to comments on Hacker News.
	HNLink string `yaml:"hn_link"`

	// Hook is a leading sentence or two to succinctly introduce the article.
	Hook string `yaml:"hook"`

	// Image is an optional image that may be included with an article.
	Image string `yaml:"image"`

	// PublishedAt is when the article was published.
	PublishedAt *time.Time `yaml:"published_at"`

	// Title is the article's title.
	Title string `yaml:"title"`

	// TOC is the HTML rendered table of contents of the article. It isn't
	// included as YAML frontmatter, but rather calculated from the article's
	// content, rendered, and then added separately.
	TOC string `yaml:"-"`
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
	// Content is the HTML content of the fragment. It isn't included as YAML
	// frontmatter, and is rather split out of an fragment's Markdown file,
	// rendered, and then added separately.
	Content string `yaml:"-"`

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

	err = compileArticles()
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

	err = linkImageAssets()
	if err != nil {
		log.Fatal(err)
	}
}

func compileArticles() error {
	articleInfos, err := ioutil.ReadDir(sorg.ArticlesDir)
	if err != nil {
		return err
	}

	for _, articleInfo := range articleInfos {
		inPath := sorg.ArticlesDir + articleInfo.Name()
		log.Debugf("Compiling: %v", inPath)

		outName := strings.Replace(articleInfo.Name(), ".md", "", -1)

		raw, err := ioutil.ReadFile(inPath)
		if err != nil {
			return err
		}

		frontmatter, content, err := splitFrontmatter(string(raw))
		if err != nil {
			return err
		}

		var info ArticleInfo
		err = yaml.Unmarshal([]byte(frontmatter), &info)
		if err != nil {
			return err
		}

		if info.Title == "" {
			return fmt.Errorf("No title for article: %v", inPath)
		}

		if info.PublishedAt == nil {
			return fmt.Errorf("No publish date for article: %v", inPath)
		}

		info.Content = string(renderMarkdown([]byte(content)))

		// TODO: Need a TOC!
		info.TOC = ""

		locals := getLocals(info.Title, map[string]interface{}{
			"Article": info,
		})

		err = renderView(sorg.LayoutsDir+"main", sorg.ViewsDir+"/articles/show",
			sorg.TargetArticlesDir+outName, locals)
		if err != nil {
			return err
		}
	}

	return nil
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

		info.Content = string(renderMarkdown([]byte(content)))

		locals := getLocals(info.Title, map[string]interface{}{
			"Fragment": info,
		})

		err = renderView(sorg.LayoutsDir+"main", sorg.ViewsDir+"/fragments/show",
			sorg.TargetFragmentsDir+outName, locals)
		if err != nil {
			return err
		}
	}

	return nil
}

// Gets a map of local values for use while rendering a template and includes
// a few "special" values that are globally relevant to all templates.
func getLocals(title string, locals map[string]interface{}) map[string]interface{} {
	defaults := map[string]interface{}{
		"BodyClass":         "",
		"GoogleAnalyticsID": conf.GoogleAnalyticsID,
		"Release":           sorg.Release,
		"Title":             title,
		"ViewportWidth":     "device-width",
	}

	for k, v := range locals {
		defaults[k] = v
	}

	return defaults
}

func compileStylesheets() error {
	outFile, err := os.Create(sorg.TargetVersionedAssetsDir + "app.css")
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

func linkImageAssets() error {
	assets, err := ioutil.ReadDir(sorg.ImagesDir)
	if err != nil {
		return err
	}

	for _, asset := range assets {
		log.Debugf("Linking image asset: %v", asset)

		// we use absolute paths for source and destination because not doing
		// so can result in some weird symbolic link inception
		source, err := filepath.Abs(sorg.ImagesDir + asset.Name())
		if err != nil {
			return err
		}

		dest, err := filepath.Abs(sorg.TargetAssetsDir + asset.Name())
		if err != nil {
			return err
		}

		err = os.RemoveAll(dest)
		if err != nil {
			return err
		}

		err = os.Symlink(source, dest)
		if err != nil {
			return err
		}
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

func renderView(layout, view, target string, locals map[string]interface{}) error {
	log.Debugf("Rendering: %v", target)

	funcMap := template.FuncMap{
		"FormatTime": func(t *time.Time) string {
			return t.Format("Jan 2, 2006")
		},
	}

	template, err := ace.Load(layout, view, &ace.Options{FuncMap: funcMap})
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
