package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/brandur/sorg"
	"github.com/brandur/sorg/assets"
	"github.com/brandur/sorg/atom"
	"github.com/brandur/sorg/markdown"
	"github.com/brandur/sorg/templatehelpers"
	"github.com/brandur/sorg/toc"
	"github.com/joeshaw/envdecode"
	_ "github.com/lib/pq"
	"github.com/yosssi/ace"
	"github.com/yosssi/gcss"
	"gopkg.in/yaml.v2"
)

var javascripts = []string{
	"jquery-1.7.2.js",
	"retina.js",
	"highcharts.js",
	"highcharts_theme.js",
	"highlight.pack.js",
	"main_sorg.js",
}

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

//
// Types
//
// Type definitions. These are mostly models used to represent the site's
// resources, but in some cases we have sorting and grouping helper types as
// well.
//

// Article represents an article to be rendered.
type Article struct {
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

	// Slug is a unique identifier for the article that also helps determine
	// where it's addressable by URL.
	Slug string `yaml:"-"`

	// Title is the article's title.
	Title string `yaml:"title"`

	// TOC is the HTML rendered table of contents of the article. It isn't
	// included as YAML frontmatter, but rather calculated from the article's
	// content, rendered, and then added separately.
	TOC string `yaml:"-"`
}

type articleByPublishedAt []*Article

func (a articleByPublishedAt) Len() int           { return len(a) }
func (a articleByPublishedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a articleByPublishedAt) Less(i, j int) bool { return a[i].PublishedAt.Before(*a[j].PublishedAt) }

// Conf contains configuration information for the command.
type Conf struct {
	// AtomAuthorName is the name of the author to include in Atom feeds.
	AtomAuthorName string `env:"AUTHOR_NAME,default=Brandur Leach"`

	// AtomAuthorName is the URL of the author to include in Atom feeds.
	AtomAuthorURL string `env:"AUTHOR_URL,default=https://brandur.org"`

	// BlackSwanDatabaseURL is a connection string for a database to connect to
	// in order to extract books, tweets, runs, etc.
	BlackSwanDatabaseURL string `env:"BLACK_SWAN_DATABASE_URL"`

	// Drafts is whether drafts of articles and fragments should be compiled
	// along with their published versions.
	Drafts bool `env:"DRAFTS,default=false"`

	// GoogleAnalyticsID is the account identifier for Google Analytics to use.
	GoogleAnalyticsID string `env:"GOOGLE_ANALYTICS_ID"`

	// SiteURL is the absolute URL where the compiled site will be hosted.
	SiteURL string `env:"SITE_URL,default=https://brandur.org"`

	// Verbose is whether the program will print debug output as it's running.
	Verbose bool `env:"VERBOSE,default=false"`
}

// Fragment represents a fragment (that is, a short "stream of consciousness"
// style article) to be rendered.
type Fragment struct {
	// Content is the HTML content of the fragment. It isn't included as YAML
	// frontmatter, and is rather split out of an fragment's Markdown file,
	// rendered, and then added separately.
	Content string `yaml:"-"`

	// Image is an optional image that may be included with a fragment.
	Image string `yaml:"image"`

	// PublishedAt is when the fragment was published.
	PublishedAt *time.Time `yaml:"published_at"`

	// Slug is a unique identifier for the fragment that also helps determine
	// where it's addressable by URL.
	Slug string `yaml:"-"`

	// Title is the fragment's title.
	Title string `yaml:"title"`
}

type fragmentByPublishedAt []*Fragment

func (a fragmentByPublishedAt) Len() int           { return len(a) }
func (a fragmentByPublishedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a fragmentByPublishedAt) Less(i, j int) bool { return a[i].PublishedAt.Before(*a[j].PublishedAt) }

// Photo is a photography downloaded from Flickr.
type Photo struct {
	// LargeImageURL is the location where the large-sized version of the photo
	// can be downloaded from Flickr.
	LargeImageURL string

	// LargeImageHeight and LargeImageWidth are the height and with of the
	// large-sized version of the photo.
	LargeImageHeight, LargeImageWidth int

	// MediumImageURL is the location where the medium-sized version of the
	// photo can be downloaded from Flickr.
	MediumImageURL string

	// MediumImageHeight and MediumImageWidth are the height and with of the
	// medium-sized version of the photo.
	MediumImageHeight, MediumImageWidth int

	// OccurredAt is UTC time when the photo was published.
	OccurredAt *time.Time

	// Slug is a unique identifier for the photo. It can be used to link it
	// back to the photo on Flickr.
	Slug string
}

// Reading is a read book procured from Goodreads.
type Reading struct {
	// Author is the full name of the book's author.
	Author string

	// ISBN is the unique identifier for the book.
	ISBN string

	// NumPages are the number of pages in the book. If unavailable, this
	// number will be zero.
	NumPages int

	// OccurredAt is UTC time when the book was read.
	OccurredAt *time.Time

	// Rating is the rating that I assigned to the read book.
	Rating int

	// Title is the title of the book.
	Title string
}

// readingYear holds a collection of readings grouped by year.
type readingYear struct {
	Year     int
	Readings []*Reading
}

// Run is a run as downloaded from Strava.
type Run struct {
	// Distance is the distance traveled for the run in meters.
	Distance float64

	// ElevationGain is the total gain in elevation in meters.
	ElevationGain float64

	// LocationCity is the closest city to which the run occurred. It may be
	// an empty string if Strava wasn't able to match anything.
	LocationCity string

	// MovingTime is the amount of time that the run took.
	MovingTime time.Duration

	// OccurredAt is the local time in which the run occurred. Note that we
	// don't use UTC here so as to not make runs in other timezones look to
	// have occurred at crazy times.
	OccurredAt *time.Time
}

// Tweet is a post to Twitter.
type Tweet struct {
	// Content is the content of the tweet. It may contain shortened URLs and
	// the like and so require extra rendering.
	Content string

	// OccurredAt is UTC time when the tweet was published.
	OccurredAt *time.Time

	// Slug is a unique identifier for the tweet. It can be used to link it
	// back to the post on Twitter.
	Slug string
}

// tweetYear holds a collection of tweetMonths grouped by year.
type tweetYear struct {
	Year   int
	Months []*tweetMonth
}

// tweetMonth holds a collection of Tweets grouped by year.
type tweetMonth struct {
	Month  time.Month
	Tweets []*Tweet
}

//
// Variables
//

// Left as a global for now for the sake of convenience, but it's not used in
// very many places and can probably be refactored as a local if desired.
var conf Conf

var errBadFrontmatter = fmt.Errorf("Unable to split YAML frontmatter")

// pagesVars contains meta information for static pages that are part of the
// site. This mostly titles, but can also be body classes for custom styling.
//
// This isn't the best system, but was the cheapest way to accomplish what I
// needed for the time being. It could probably use an overhaul to something
// better at some point.
var pagesVars = map[string]map[string]interface{}{
	"about": {
		"Title": "About",
	},
	"accidental": {
		"Title":     "Accidental",
		"BodyClass": "quote",
	},
	"crying": {
		"Title":     "Crying",
		"BodyClass": "quote",
	},
	"favors": {
		"Title":     "Favors",
		"BodyClass": "quote",
	},
	"lies": {
		"Title":     "Lies",
		"BodyClass": "quote",
	},
	"talks": {
		"Title": "Talks",
	},
	"that-sunny-dome": {
		"Title":     "That Sunny Dome",
		"BodyClass": "quote",
	},
}

//
// Main
//

func main() {
	err := envdecode.Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	var db *sql.DB
	if conf.BlackSwanDatabaseURL != "" {
		var err error
		db, err = sql.Open("postgres", conf.BlackSwanDatabaseURL)
		if err != nil {
			log.Fatal(err)
		}
	}

	sorg.InitLog(conf.Verbose)

	err = sorg.CreateTargetDirs()
	if err != nil {
		log.Fatal(err)
	}

	articles, err := compileArticles()
	if err != nil {
		log.Fatal(err)
	}

	fragments, err := compileFragments()
	if err != nil {
		log.Fatal(err)
	}

	err = compileJavascripts(javascripts)
	if err != nil {
		log.Fatal(err)
	}

	err = compilePages()
	if err != nil {
		log.Fatal(err)
	}

	photos, err := compilePhotos(db)
	if err != nil {
		log.Fatal(err)
	}

	err = compileReading(db)
	if err != nil {
		log.Fatal(err)
	}

	err = compileRuns(db)
	if err != nil {
		log.Fatal(err)
	}

	err = compileStylesheets(stylesheets)
	if err != nil {
		log.Fatal(err)
	}

	err = compileTwitter(db)
	if err != nil {
		log.Fatal(err)
	}

	err = compileHome(articles, fragments, photos)
	if err != nil {
		log.Fatal(err)
	}

	err = linkImageAssets()
	if err != nil {
		log.Fatal(err)
	}
}

//
// Compilation functions
//
// These functions are the main entry points for compiling the site's
// resources.
//

func compileArticles() ([]*Article, error) {
	articles, err := compileArticlesDir(sorg.ContentDir + "/articles")
	if err != nil {
		return nil, err
	}

	if conf.Drafts {
		drafts, err := compileArticlesDir(sorg.ContentDir + "/drafts")
		if err != nil {
			return nil, err
		}

		articles = append(articles, drafts...)
	}

	sort.Sort(sort.Reverse(articleByPublishedAt(articles)))

	locals := getLocals("Articles", map[string]interface{}{
		"Articles": articles,
	})

	err = renderView(sorg.MainLayout, sorg.ViewsDir+"/articles/index",
		sorg.TargetDir+"/articles/index.html", locals)
	if err != nil {
		return nil, err
	}

	err = compileArticlesFeed(articles)
	if err != nil {
		return nil, err
	}

	return articles, nil
}

func compileFragments() ([]*Fragment, error) {
	fragments, err := compileFragmentsDir(sorg.ContentDir + "/fragments")
	if err != nil {
		return nil, err
	}

	if conf.Drafts {
		drafts, err := compileFragmentsDir(sorg.ContentDir + "/fragments-drafts")
		if err != nil {
			return nil, err
		}

		fragments = append(fragments, drafts...)
	}

	sort.Sort(sort.Reverse(fragmentByPublishedAt(fragments)))

	locals := getLocals("Fragments", map[string]interface{}{
		"Fragments": fragments,
	})

	err = renderView(sorg.MainLayout, sorg.ViewsDir+"/fragments/index",
		sorg.TargetDir+"/fragments/index.html", locals)
	if err != nil {
		return nil, err
	}

	err = compileFragmentsFeed(fragments)
	if err != nil {
		return nil, err
	}

	return fragments, nil
}

func compileHome(articles []*Article, fragments []*Fragment, photos []*Photo) error {
	if len(articles) > 5 {
		articles = articles[0:5]
	}

	if len(fragments) > 5 {
		fragments = fragments[0:5]
	}

	if len(photos) > 27 {
		photos = photos[0:27]
	}

	locals := getLocals("brandur.org", map[string]interface{}{
		"Articles":      articles,
		"Fragments":     fragments,
		"Photos":        photos,
		"ViewportWidth": 600,
	})

	err := renderView(sorg.MainLayout, sorg.ViewsDir+"/index",
		sorg.TargetDir+"/index.html", locals)
	if err != nil {
		return err
	}

	return nil
}

func compileJavascripts(javascripts []string) error {
	outFile, err := os.Create(sorg.TargetVersionedAssetsDir + "/app.js")
	if err != nil {
		return err
	}
	defer outFile.Close()

	for _, javascript := range javascripts {
		inPath := sorg.ContentDir + "/assets/javascripts/" + javascript
		log.Debugf("Compiling: %v", inPath)

		inFile, err := os.Open(inPath)
		if err != nil {
			return err
		}

		outFile.WriteString("/* " + javascript + " */\n\n")
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

func compilePages() error {
	return compilePagesDir(sorg.PagesDir)
}

func compilePagesDir(dir string) error {
	log.Debugf("Descending into for pages: %v", dir)

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			err := compilePagesDir(dir + fileInfo.Name())
			if err != nil {
				return err
			}
		} else {
			// Subtract 4 for the ".ace" extension.
			name := fileInfo.Name()[0 : len(fileInfo.Name())-4]

			// Remove the "./pages" directory, but keep the rest of the path.
			//
			// Looks something like "about".
			pagePath := strings.TrimPrefix(dir, sorg.PagesDir) + name

			// Looks something like "./public/about".
			target := sorg.TargetDir + "/" + pagePath

			log.Debugf("Compiling page: %v to %v", dir+"/"+fileInfo.Name(), target)

			locals, ok := pagesVars[pagePath]
			if !ok {
				log.Errorf("No page meta information: %v", pagePath)
			}

			locals = getLocals("Page", locals)

			err := os.MkdirAll(sorg.TargetDir+"/"+dir, 0755)
			if err != nil {
				return err
			}

			err = renderView(sorg.MainLayout, dir+"/"+name,
				target, locals)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func compilePhotos(db *sql.DB) ([]*Photo, error) {
	photos, err := getPhotosData(db)
	if err != nil {
		return nil, err
	}

	// Keep a published copy of all the photos that we need.
	var photoAssets []assets.Asset
	for _, photo := range photos {
		photoAssets = append(photoAssets,
			assets.Asset{URL: photo.LargeImageURL,
				Target: sorg.TargetDir + "/assets/photos/" + photo.Slug + "@2x.jpg"},
			assets.Asset{URL: photo.MediumImageURL,
				Target: sorg.TargetDir + "/assets/photos/" + photo.Slug + ".jpg"},
		)
	}

	log.Debugf("Fetching %d photo(s)", len(photoAssets))
	err = assets.Fetch(photoAssets)
	if err != nil {
		return nil, err
	}

	locals := getLocals("Photos", map[string]interface{}{
		"Photos":        photos,
		"ViewportWidth": 600,
	})

	err = renderView(sorg.MainLayout, sorg.ViewsDir+"/photos/index",
		sorg.TargetDir+"/photos/index.html", locals)
	if err != nil {
		return nil, err
	}

	return photos, nil
}

func compileReading(db *sql.DB) error {
	readings, err := getReadingsData(db)
	if err != nil {
		return err
	}

	readingsByYear := groupReadingsByYear(readings)

	readingsByYearXYears, readingsByYearYCounts, err :=
		getReadingsCountByYearData(db)
	if err != nil {
		return err
	}

	pagesByYearXYears, pagesByYearYCounts, err := getReadingsPagesByYearData(db)
	if err != nil {
		return err
	}

	locals := getLocals("Reading", map[string]interface{}{
		"NumReadings":    len(readings),
		"ReadingsByYear": readingsByYear,

		// chart: readings by year
		"ReadingsByYearXYears":  readingsByYearXYears,
		"ReadingsByYearYCounts": readingsByYearYCounts,

		// chart: pages by year
		"PagesByYearXYears":  pagesByYearXYears,
		"PagesByYearYCounts": pagesByYearYCounts,
	})

	err = renderView(sorg.MainLayout, sorg.ViewsDir+"/reading/index",
		sorg.TargetDir+"/reading/index.html", locals)
	if err != nil {
		return err
	}

	return nil
}

func compileRuns(db *sql.DB) error {
	runs, err := getRunsData(db)
	if err != nil {
		return err
	}

	lastYearXDays, lastYearYDistances, err := getRunsLastYearData(db)
	if err != nil {
		return err
	}

	byYearXYears, byYearYDistances, err := getRunsByYearData(db)
	if err != nil {
		return err
	}

	locals := getLocals("Runs", map[string]interface{}{
		"Runs": runs,

		// chart: runs over last year
		"LastYearXDays":      lastYearXDays,
		"LastYearYDistances": lastYearYDistances,

		// chart: run distance by year
		"ByYearXYears":     byYearXYears,
		"ByYearYDistances": byYearYDistances,
	})

	err = renderView(sorg.MainLayout, sorg.ViewsDir+"/runs/index",
		sorg.TargetDir+"/runs/index.html", locals)
	if err != nil {
		return err
	}

	return nil
}

func compileTwitter(db *sql.DB) error {
	tweets, err := getTwitterData(db, false)
	if err != nil {
		return err
	}

	tweetsWithReplies, err := getTwitterData(db, true)
	if err != nil {
		return err
	}

	optionsMatrix := map[string]bool{
		"/index.html":   false,
		"/with-replies": true,
	}

	for page, withReplies := range optionsMatrix {
		ts := tweets
		if withReplies {
			ts = tweetsWithReplies
		}

		tweetsByYearAndMonth := groupTwitterByYearAndMonth(ts)

		tweetCountXMonths, tweetCountYCounts, err :=
			getTwitterByMonth(db, withReplies)
		if err != nil {
			return err
		}

		locals := getLocals("Twitter", map[string]interface{}{
			"NumTweets":            len(tweets),
			"NumTweetsWithReplies": len(tweetsWithReplies),
			"TweetsByYearAndMonth": tweetsByYearAndMonth,
			"WithReplies":          withReplies,

			// chart: tweets by month
			"TweetCountXMonths": tweetCountXMonths,
			"TweetCountYCounts": tweetCountYCounts,
		})

		err = renderView(sorg.MainLayout, sorg.ViewsDir+"/twitter/index",
			sorg.TargetDir+"/twitter/"+page, locals)
		if err != nil {
			return err
		}
	}

	return nil
}

func compileStylesheets(stylesheets []string) error {
	outFile, err := os.Create(sorg.TargetVersionedAssetsDir + "/app.css")
	if err != nil {
		return err
	}
	defer outFile.Close()

	for _, stylesheet := range stylesheets {
		inPath := sorg.ContentDir + "/assets/stylesheets/" + stylesheet
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

//
// Other functions
//
// Any other functions. Try to keep them alphabetized.
//

func compileArticlesDir(dir string) ([]*Article, error) {
	articleInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var articles []*Article

	for _, articleInfo := range articleInfos {
		if isHidden(articleInfo.Name()) {
			continue
		}

		inPath := dir + "/" + articleInfo.Name()
		log.Debugf("Compiling: %v", inPath)

		raw, err := ioutil.ReadFile(inPath)
		if err != nil {
			return nil, err
		}

		frontmatter, content, err := splitFrontmatter(string(raw))
		if err != nil {
			return nil, err
		}

		var article Article
		articles = append(articles, &article)

		err = yaml.Unmarshal([]byte(frontmatter), &article)
		if err != nil {
			return nil, err
		}

		article.Slug = strings.Replace(articleInfo.Name(), ".md", "", -1)

		if article.Title == "" {
			return nil, fmt.Errorf("No title for article: %v", inPath)
		}

		if article.PublishedAt == nil {
			return nil, fmt.Errorf("No publish date for article: %v", inPath)
		}

		article.Content = markdown.Render(content)

		article.TOC, err = toc.Render(article.Content)
		if err != nil {
			return nil, err
		}

		locals := getLocals(article.Title, map[string]interface{}{
			"Article": article,
		})

		err = renderView(sorg.MainLayout, sorg.ViewsDir+"/articles/show",
			sorg.TargetDir+"/"+article.Slug, locals)
		if err != nil {
			return nil, err
		}
	}

	return articles, nil
}

func compileArticlesFeed(articles []*Article) error {
	feed := &atom.Feed{
		Title: "Articles - brandur.org",
		ID:    "tag:brandur.org.org,2013:/articles",

		Links: []*atom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/articles.atom"},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	for _, article := range articles {
		entry := &atom.Entry{
			Title:     article.Title,
			Content:   &atom.EntryContent{Content: article.Content},
			Published: *article.PublishedAt,
			Updated:   *article.PublishedAt,
			Link:      &atom.Link{Href: conf.SiteURL + "/" + article.Slug},
			ID:        "tag:brandur.org," + article.PublishedAt.Format("2006-01-02") + ":" + article.Slug,

			AuthorName: conf.AtomAuthorName,
			AuthorURI:  conf.AtomAuthorURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	f, err := os.Create(sorg.TargetDir + "/articles.atom")
	if err != nil {
		return err
	}
	defer f.Close()

	return feed.Encode(f, "  ")
}

func compileFragmentsDir(dir string) ([]*Fragment, error) {
	fragmentInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var fragments []*Fragment

	for _, fragmentInfo := range fragmentInfos {
		if isHidden(fragmentInfo.Name()) {
			continue
		}

		inPath := dir + "/" + fragmentInfo.Name()
		log.Debugf("Compiling: %v", inPath)

		raw, err := ioutil.ReadFile(inPath)
		if err != nil {
			return nil, err
		}

		frontmatter, content, err := splitFrontmatter(string(raw))
		if err != nil {
			return nil, err
		}

		var fragment Fragment
		fragments = append(fragments, &fragment)

		err = yaml.Unmarshal([]byte(frontmatter), &fragment)
		if err != nil {
			return nil, err
		}

		fragment.Slug = strings.Replace(fragmentInfo.Name(), ".md", "", -1)

		if fragment.Title == "" {
			return nil, fmt.Errorf("No title for fragment: %v", inPath)
		}

		if fragment.PublishedAt == nil {
			return nil, fmt.Errorf("No publish date for fragment: %v", inPath)
		}

		fragment.Content = markdown.Render(content)

		locals := getLocals(fragment.Title, map[string]interface{}{
			"Fragment": fragment,
		})

		err = renderView(sorg.MainLayout, sorg.ViewsDir+"/fragments/show",
			sorg.TargetDir+"/fragments/"+fragment.Slug, locals)
		if err != nil {
			return nil, err
		}
	}

	return fragments, nil
}

func compileFragmentsFeed(fragments []*Fragment) error {
	feed := &atom.Feed{
		Title: "Fragments - brandur.org",
		ID:    "tag:brandur.org.org,2013:/fragments",

		Links: []*atom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/fragments.atom"},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	for _, fragment := range fragments {
		entry := &atom.Entry{
			Title:     fragment.Title,
			Content:   &atom.EntryContent{Content: fragment.Content},
			Published: *fragment.PublishedAt,
			Updated:   *fragment.PublishedAt,
			Link:      &atom.Link{Href: conf.SiteURL + "/fragments/" + fragment.Slug},
			ID:        "tag:brandur.org," + fragment.PublishedAt.Format("2006-01-02") + ":fragments/" + fragment.Slug,

			AuthorName: conf.AtomAuthorName,
			AuthorURI:  conf.AtomAuthorURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	f, err := os.Create(sorg.TargetDir + "/fragments.atom")
	if err != nil {
		return err
	}
	defer f.Close()

	return feed.Encode(f, "  ")
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

func getPhotosData(db *sql.DB) ([]*Photo, error) {
	var photos []*Photo

	if db == nil {
		return photos, nil
	}

	rows, err := db.Query(`
		SELECT
			metadata -> 'large_image',
			(metadata -> 'large_height')::int,
			(metadata -> 'large_width')::int,
			metadata -> 'medium_image',
			(metadata -> 'medium_height')::int,
			(metadata -> 'medium_width')::int,
			(metadata -> 'occurred_at_local')::timestamptz,
			slug
		FROM events
		WHERE type = 'flickr'
			AND (metadata -> 'medium_width')::int = 500
		ORDER BY occurred_at DESC
		LIMIT 30
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var photo Photo

		err = rows.Scan(
			&photo.LargeImageURL,
			&photo.LargeImageHeight,
			&photo.LargeImageWidth,
			&photo.MediumImageURL,
			&photo.MediumImageHeight,
			&photo.MediumImageWidth,
			&photo.OccurredAt,
			&photo.Slug,
		)
		if err != nil {
			return nil, err
		}

		photos = append(photos, &photo)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return photos, nil
}

func getReadingsData(db *sql.DB) ([]*Reading, error) {
	var readings []*Reading

	if db == nil {
		return readings, nil
	}

	rows, err := db.Query(`
		SELECT
			metadata -> 'author',
			metadata -> 'isbn',
			-- not every book has a number of pages
			(COALESCE(NULLIF(metadata -> 'num_pages', ''), '0'))::int,
			occurred_at,
			(metadata -> 'rating')::int,
			metadata -> 'title'
		FROM events
		WHERE type = 'goodreads'
		ORDER BY occurred_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var reading Reading

		err = rows.Scan(
			&reading.Author,
			&reading.ISBN,
			&reading.NumPages,
			&reading.OccurredAt,
			&reading.Rating,
			&reading.Title,
		)
		if err != nil {
			return nil, err
		}

		readings = append(readings, &reading)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return readings, nil
}

func getReadingsCountByYearData(db *sql.DB) ([]string, []int, error) {
	// Give these arrays 0 elements (instead of null) in case no Black Swan
	// data gets loaded but we still need to render the page.
	byYearXYears := []string{}
	byYearYCounts := []int{}

	if db == nil {
		return byYearXYears, byYearYCounts, nil
	}

	rows, err := db.Query(`
		SELECT date_part('year', occurred_at)::text AS year,
			COUNT(*)
		FROM events
		WHERE type = 'goodreads'
		GROUP BY date_part('year', occurred_at)
		ORDER BY date_part('year', occurred_at)
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var year string
		var count int

		err = rows.Scan(
			&year,
			&count,
		)
		if err != nil {
			return nil, nil, err
		}

		byYearXYears = append(byYearXYears, year)
		byYearYCounts = append(byYearYCounts, count)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, err
	}

	return byYearXYears, byYearYCounts, nil
}

func getReadingsPagesByYearData(db *sql.DB) ([]string, []int, error) {
	// Give these arrays 0 elements (instead of null) in case no Black Swan
	// data gets loaded but we still need to render the page.
	byYearXYears := []string{}
	byYearYCounts := []int{}

	if db == nil {
		return byYearXYears, byYearYCounts, nil
	}

	rows, err := db.Query(`
		SELECT date_part('year', occurred_at)::text AS year,
			sum((metadata -> 'num_pages')::int)
		FROM events
		WHERE type = 'goodreads'
			AND metadata -> 'num_pages' <> ''
		GROUP BY date_part('year', occurred_at)
		ORDER BY date_part('year', occurred_at)
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var year string
		var count int

		err = rows.Scan(
			&year,
			&count,
		)
		if err != nil {
			return nil, nil, err
		}

		byYearXYears = append(byYearXYears, year)
		byYearYCounts = append(byYearYCounts, count)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, err
	}

	return byYearXYears, byYearYCounts, nil
}

func getRunsData(db *sql.DB) ([]*Run, error) {
	var runs []*Run

	if db == nil {
		return runs, nil
	}

	rows, err := db.Query(`
		SELECT
			(metadata -> 'distance')::float,
			(metadata -> 'total_elevation_gain')::float,
			(metadata -> 'location_city'),
			-- we multiply by 10e9 here because a Golang time.Duration is
			-- an int64 represented in nanoseconds
			(metadata -> 'moving_time')::bigint * 1000000000,
			(metadata -> 'occurred_at_local')::timestamptz
		FROM events
		WHERE type = 'strava'
			AND metadata -> 'type' = 'Run'
		ORDER BY occurred_at DESC
		LIMIT 30
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var locationCity *string
		var run Run

		err = rows.Scan(
			&run.Distance,
			&run.ElevationGain,
			&locationCity,
			&run.MovingTime,
			&run.OccurredAt,
		)
		if err != nil {
			return nil, err
		}

		if locationCity != nil {
			run.LocationCity = *locationCity
		}

		runs = append(runs, &run)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return runs, nil
}

func getRunsByYearData(db *sql.DB) ([]string, []float64, error) {
	// Give these arrays 0 elements (instead of null) in case no Black Swan
	// data gets loaded but we still need to render the page.
	byYearXYears := []string{}
	byYearYDistances := []float64{}

	if db == nil {
		return byYearXYears, byYearYDistances, nil
	}

	rows, err := db.Query(`
		WITH runs AS (
			SELECT *,
				(metadata -> 'occurred_at_local')::timestamptz AS occurred_at_local,
				-- convert to distance in kilometers
				((metadata -> 'distance')::float / 1000.0) AS distance
			FROM events
			WHERE type = 'strava'
				AND metadata -> 'type' = 'Run'
		)

		SELECT date_part('year', occurred_at_local)::text AS year,
			SUM(distance)
		FROM runs
		GROUP BY year
		ORDER BY year DESC
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var distance float64
		var year string

		err = rows.Scan(
			&year,
			&distance,
		)
		if err != nil {
			return nil, nil, err
		}

		byYearXYears = append(byYearXYears, year)
		byYearYDistances = append(byYearYDistances, distance)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, err
	}

	return byYearXYears, byYearYDistances, nil
}

func getRunsLastYearData(db *sql.DB) ([]string, []float64, error) {
	// Give these arrays 0 elements (instead of null) in case no Black Swan
	// data gets loaded but we still need to render the page.
	lastYearXDays := []string{}
	lastYearYDistances := []float64{}

	if db == nil {
		return lastYearXDays, lastYearYDistances, nil
	}

	rows, err := db.Query(`
		WITH runs AS (
			SELECT *,
				(metadata -> 'occurred_at_local')::timestamptz AS occurred_at_local,
				-- convert to distance in kilometers
				((metadata -> 'distance')::float / 1000.0) AS distance
			FROM events
			WHERE type = 'strava'
				AND metadata -> 'type' = 'Run'
		),

		runs_days AS (
			SELECT date_trunc('day', occurred_at_local) AS day,
				SUM(distance) AS distance
			FROM runs
			WHERE occurred_at_local > NOW() - '180 days'::interval
				GROUP BY day
		),

		-- generates a baseline series of every day in the last 180 days
		-- along with a zeroed distance which we will then add against the
		-- actual runs we extracted
		days AS (
			SELECT i::date AS day,
				0::float AS distance
			FROM generate_series(NOW() - '180 days'::interval,
				NOW(), '1 day'::interval) i
		)

		SELECT to_char(d.day, 'Mon DD') AS day,
			d.distance + COALESCE(rd.distance, 0::float)
		FROM days d
			LEFT JOIN runs_days rd ON d.day = rd.day
		ORDER BY day ASC
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var day string
		var distance float64

		err = rows.Scan(
			&day,
			&distance,
		)
		if err != nil {
			return nil, nil, err
		}

		lastYearXDays = append(lastYearXDays, day)
		lastYearYDistances = append(lastYearYDistances, distance)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, err
	}

	return lastYearXDays, lastYearYDistances, nil
}

func getTwitterByMonth(db *sql.DB, withReplies bool) ([]string, []int, error) {
	// Give these arrays 0 elements (instead of null) in case no Black Swan
	// data gets loaded but we still need to render the page.
	tweetCountXMonths := []string{}
	tweetCountYCounts := []int{}

	if db == nil {
		return tweetCountXMonths, tweetCountYCounts, nil
	}

	rows, err := db.Query(`
		SELECT to_char(date_trunc('month', occurred_at), 'Mon ''YY'),
			COUNT(*)
		FROM events
		WHERE type = 'twitter'
			-- Note that false is always an allowed value here because we
			-- always want all non-reply tweets.
			AND (metadata -> 'reply')::boolean IN (false, $1)
		GROUP BY date_trunc('month', occurred_at)
		ORDER BY date_trunc('month', occurred_at)
	`, withReplies)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var count int
		var month string

		err = rows.Scan(
			&month,
			&count,
		)
		if err != nil {
			return nil, nil, err
		}

		tweetCountXMonths = append(tweetCountXMonths, month)
		tweetCountYCounts = append(tweetCountYCounts, count)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, err
	}

	return tweetCountXMonths, tweetCountYCounts, nil
}

func getTwitterData(db *sql.DB, withReplies bool) ([]*Tweet, error) {
	var tweets []*Tweet

	if db == nil {
		return tweets, nil
	}

	rows, err := db.Query(`
		SELECT
			content,
			occurred_at,
			slug
		FROM events
		WHERE type = 'twitter'
			-- Note that false is always an allowed value here because we
			-- always want all non-reply tweets.
			AND (metadata -> 'reply')::boolean IN (false, $1)
		ORDER BY occurred_at DESC
	`, withReplies)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tweet Tweet

		err = rows.Scan(
			&tweet.Content,
			&tweet.OccurredAt,
			&tweet.Slug,
		)
		if err != nil {
			return nil, err
		}

		tweets = append(tweets, &tweet)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return tweets, nil
}

func groupReadingsByYear(readings []*Reading) []*readingYear {
	var year *readingYear
	var years []*readingYear

	for _, reading := range readings {
		if year == nil || year.Year != reading.OccurredAt.Year() {
			year = &readingYear{reading.OccurredAt.Year(), nil}
			years = append(years, year)
		}

		year.Readings = append(year.Readings, reading)
	}

	return years
}

func groupTwitterByYearAndMonth(tweets []*Tweet) []*tweetYear {
	var month *tweetMonth
	var year *tweetYear
	var years []*tweetYear

	for _, tweet := range tweets {
		if year == nil || year.Year != tweet.OccurredAt.Year() {
			year = &tweetYear{tweet.OccurredAt.Year(), nil}
			years = append(years, year)
			month = nil
		}

		if month == nil || month.Month != tweet.OccurredAt.Month() {
			month = &tweetMonth{tweet.OccurredAt.Month(), nil}
			year.Months = append(year.Months, month)
		}

		month.Tweets = append(month.Tweets, tweet)
	}

	return years
}

func isHidden(file string) bool {
	return strings.HasPrefix(file, ".")
}

func linkImageAssets() error {
	assets, err := ioutil.ReadDir(sorg.ContentDir + "/assets/images")
	if err != nil {
		return err
	}

	for _, asset := range assets {
		log.Debugf("Linking image asset: %v", asset.Name())

		// we use absolute paths for source and destination because not doing
		// so can result in some weird symbolic link inception
		source, err := filepath.Abs(sorg.ContentDir + "/assets/images/" + asset.Name())
		if err != nil {
			return err
		}

		dest, err := filepath.Abs(sorg.TargetDir + "/assets/" + asset.Name())
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

func renderView(layout, view, target string, locals map[string]interface{}) error {
	log.Debugf("Rendering: %v", target)

	template, err := ace.Load(layout, view, &ace.Options{FuncMap: templatehelpers.FuncMap})
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
