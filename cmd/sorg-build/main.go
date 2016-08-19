package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
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
	"github.com/brandur/sorg/pool"
	"github.com/brandur/sorg/templatehelpers"
	"github.com/brandur/sorg/toc"
	"github.com/joeshaw/envdecode"
	_ "github.com/lib/pq"
	"github.com/yosssi/ace"
	"github.com/yosssi/gcss"
	"gopkg.in/yaml.v2"
)

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

	// Draft indicates that the article is not yet published.
	Draft bool `yaml:"-"`

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

// articleYear holds a collection of articles grouped by year.
type articleYear struct {
	Year     int
	Articles []*Article
}

// Conf contains configuration information for the command.
type Conf struct {
	// AtomAuthorName is the name of the author to include in Atom feeds.
	AtomAuthorName string `env:"AUTHOR_NAME,default=Brandur Leach"`

	// AtomAuthorName is the URL of the author to include in Atom feeds.
	AtomAuthorURL string `env:"AUTHOR_URL,default=https://brandur.org"`

	// BlackSwanDatabaseURL is a connection string for a database to connect to
	// in order to extract books, tweets, runs, etc.
	BlackSwanDatabaseURL string `env:"BLACK_SWAN_DATABASE_URL"`

	// Concurrency is the number of build Goroutines that will be used to
	// perform build work items.
	Concurrency int `env:"CONCURRENCY,default=30"`

	// Drafts is whether drafts of articles and fragments should be compiled
	// along with their published versions.
	//
	// Activating drafts also prompts the creation of a robots.txt to make sure
	// that drafts aren't inadvertently accessed by web crawlers.
	Drafts bool `env:"DRAFTS,default=false"`

	// ContentOnly tells the build step that it should build using only files
	// in the content directory. This means that information imported from a
	// Black Swan database (reading, tweets, etc.) will be skipped. This is
	// a speed optimization for use while watching for file changes.
	ContentOnly bool `env:"CONTENT_ONLY,default=false"`

	// GoogleAnalyticsID is the account identifier for Google Analytics to use.
	GoogleAnalyticsID string `env:"GOOGLE_ANALYTICS_ID"`

	// LocalFonts starts using locally downloaded versions of Google Fonts.
	// This is not ideal for real deployment because you won't be able to
	// leverage Google's CDN and the caching that goes with it, and may not get
	// the font format for requesting browsers, but good for airplane rides
	// where you otherwise wouldn't have the fonts.
	LocalFonts bool `env:"LOCAL_FONTS,default=false"`

	// NumAtomEntries is the number of entries to put in Atom feeds.
	NumAtomEntries int `env:"NUM_ATOM_ENTRIES,default=20"`

	// SiteURL is the absolute URL where the compiled site will be hosted.
	SiteURL string `env:"SITE_URL,default=https://brandur.org"`

	// TargetDir is the target location where the site will be built to.
	TargetDir string `env:"TARGET_DIR,default=./public"`

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

	// Draft indicates that the fragment is not yet published.
	Draft bool `yaml:"-"`

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

// fragmentYear holds a collection of fragments grouped by year.
type fragmentYear struct {
	Year      int
	Fragments []*Fragment
}

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

const javascriptsDir = sorg.ContentDir + "/javascripts"

var javascripts = []string{
	javascriptsDir + "/jquery-1.7.2.js",
	javascriptsDir + "/retina.js",
	javascriptsDir + "/highcharts.js",
	javascriptsDir + "/highcharts_theme.js",
	javascriptsDir + "/highlight.pack.js",
	javascriptsDir + "/main.js",
}

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

const stylesheetsDir = sorg.ContentDir + "/stylesheets"

var stylesheets = []string{
	stylesheetsDir + "/_reset.sass",
	stylesheetsDir + "/main.sass",
	stylesheetsDir + "/about.sass",
	stylesheetsDir + "/articles.sass",
	stylesheetsDir + "/fragments.sass",
	stylesheetsDir + "/index.sass",
	stylesheetsDir + "/photos.sass",
	stylesheetsDir + "/quotes.sass",
	stylesheetsDir + "/reading.sass",
	stylesheetsDir + "/runs.sass",
	stylesheetsDir + "/signature.sass",
	stylesheetsDir + "/solarized-light.css",
	stylesheetsDir + "/tenets.sass",
	stylesheetsDir + "/twitter.sass",
}

//
// Main
//

func main() {
	start := time.Now()
	defer func() {
		log.Infof("Built site in %v.", time.Now().Sub(start))
	}()

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

	// This is where we stored "versioned" assets like compiled JS and CSS.
	// These assets have a release number that we can increment and by
	// extension quickly invalidate.
	versionedAssetsDir := path.Join(conf.TargetDir, "assets", sorg.Release)

	err = sorg.CreateOutputDirs(conf.TargetDir)
	if err != nil {
		log.Fatal(err)
	}

	var tasks []*pool.Task

	//
	// Build step 0: dependency-free
	//

	tasks = nil

	var articles []*Article
	articleChan := accumulateArticles(&articles)

	var fragments []*Fragment
	fragmentChan := accumulateFragments(&fragments)

	articleTasks, err := tasksForArticles(articleChan)
	if err != nil {
		log.Fatal(err)
	}
	tasks = append(tasks, articleTasks...)

	fragmentTasks, err := tasksForFragments(fragmentChan)
	if err != nil {
		log.Fatal(err)
	}
	tasks = append(tasks, fragmentTasks...)

	tasks = append(tasks, pool.NewTask(func() error {
		return compileJavascripts(javascripts,
			path.Join(versionedAssetsDir, "app.js"))
	}))

	pageTasks, err := tasksForPages()
	if err != nil {
		log.Fatal(err)
	}
	tasks = append(tasks, pageTasks...)

	var photos []*Photo
	tasks = append(tasks, pool.NewTask(func() error {
		var err error
		photos, err = compilePhotos(db)
		return err
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return compileReading(db)
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return compileRuns(db)
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return compileRobots(path.Join(conf.TargetDir, "robots.txt"))
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return compileStylesheets(stylesheets,
			path.Join(versionedAssetsDir, "app.css"))
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return compileTwitter(db)
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return linkImageAssets()
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return linkFontAssets()
	}))

	p := pool.NewPool(tasks, conf.Concurrency)
	err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Free up any Goroutines still waiting.
	close(articleChan)
	close(fragmentChan)

	//
	// Build step 1: any tasks dependent on the results of step 0.
	//
	// This includes build output like index pages and RSS feeds.
	//

	tasks = nil

	sort.Sort(sort.Reverse(articleByPublishedAt(articles)))
	sort.Sort(sort.Reverse(fragmentByPublishedAt(fragments)))

	tasks = append(tasks, pool.NewTask(func() error {
		return compileArticlesFeed(articles)
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return compileArticlesIndex(articles)
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return compileFragmentsFeed(fragments)
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return compileFragmentsIndex(fragments)
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return compileHome(articles, fragments, photos)
	}))

	p = pool.NewPool(tasks, conf.Concurrency)
	err = p.Run()
	if err != nil {
		log.Fatal(err)
	}
}

//
// Compilation functions
//
// These functions perform the heavy-lifting in compiling the site's resources.
// They are normally run concurrently.
//

func compileArticle(dir, name string, draft bool) (*Article, error) {
	inPath := dir + "/" + name

	raw, err := ioutil.ReadFile(inPath)
	if err != nil {
		return nil, err
	}

	frontmatter, content, err := splitFrontmatter(string(raw))
	if err != nil {
		return nil, err
	}

	var article Article
	err = yaml.Unmarshal([]byte(frontmatter), &article)
	if err != nil {
		return nil, err
	}

	article.Draft = draft
	article.Slug = strings.Replace(name, ".md", "", -1)

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
		conf.TargetDir+"/"+article.Slug, locals)
	if err != nil {
		return nil, err
	}

	return &article, nil
}

func compileArticlesFeed(articles []*Article) error {
	start := time.Now()
	defer func() {
		log.Debugf("Compiled articles feed in %v.", time.Now().Sub(start))
	}()

	feed := &atom.Feed{
		Title: "Articles - brandur.org",
		ID:    "tag:brandur.org.org,2013:/articles",

		Links: []*atom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/articles.atom"},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	if len(articles) > 0 {
		feed.Updated = *articles[0].PublishedAt
	}

	for i, article := range articles {
		if i >= conf.NumAtomEntries {
			break
		}

		entry := &atom.Entry{
			Title:     article.Title,
			Content:   &atom.EntryContent{Content: article.Content, Type: "html"},
			Published: *article.PublishedAt,
			Updated:   *article.PublishedAt,
			Link:      &atom.Link{Href: conf.SiteURL + "/" + article.Slug},
			ID:        "tag:brandur.org," + article.PublishedAt.Format("2006-01-02") + ":" + article.Slug,

			AuthorName: conf.AtomAuthorName,
			AuthorURI:  conf.AtomAuthorURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	f, err := os.Create(conf.TargetDir + "/articles.atom")
	if err != nil {
		return err
	}
	defer f.Close()

	return feed.Encode(f, "  ")
}

func compileArticlesIndex(articles []*Article) error {
	start := time.Now()
	defer func() {
		log.Debugf("Compiled articles index in %v.", time.Now().Sub(start))
	}()

	articlesByYear := groupArticlesByYear(articles)

	locals := getLocals("Articles", map[string]interface{}{
		"ArticlesByYear": articlesByYear,
	})

	err := renderView(sorg.MainLayout, sorg.ViewsDir+"/articles/index",
		conf.TargetDir+"/articles/index.html", locals)
	if err != nil {
		return err
	}

	return nil
}

func compileFragment(dir, name string, draft bool) (*Fragment, error) {
	inPath := dir + "/" + name

	raw, err := ioutil.ReadFile(inPath)
	if err != nil {
		return nil, err
	}

	frontmatter, content, err := splitFrontmatter(string(raw))
	if err != nil {
		return nil, err
	}

	var fragment Fragment
	err = yaml.Unmarshal([]byte(frontmatter), &fragment)
	if err != nil {
		return nil, err
	}

	fragment.Draft = draft
	fragment.Slug = strings.Replace(name, ".md", "", -1)

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
		conf.TargetDir+"/fragments/"+fragment.Slug, locals)
	if err != nil {
		return nil, err
	}

	return &fragment, nil
}

func compileFragmentsFeed(fragments []*Fragment) error {
	start := time.Now()
	defer func() {
		log.Debugf("Compiled fragments feed in %v.", time.Now().Sub(start))
	}()

	feed := &atom.Feed{
		Title: "Fragments - brandur.org",
		ID:    "tag:brandur.org.org,2013:/fragments",

		Links: []*atom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/fragments.atom"},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	if len(fragments) > 0 {
		feed.Updated = *fragments[0].PublishedAt
	}

	for i, fragment := range fragments {
		if i >= conf.NumAtomEntries {
			break
		}

		entry := &atom.Entry{
			Title:     fragment.Title,
			Content:   &atom.EntryContent{Content: fragment.Content, Type: "html"},
			Published: *fragment.PublishedAt,
			Updated:   *fragment.PublishedAt,
			Link:      &atom.Link{Href: conf.SiteURL + "/fragments/" + fragment.Slug},
			ID:        "tag:brandur.org," + fragment.PublishedAt.Format("2006-01-02") + ":fragments/" + fragment.Slug,

			AuthorName: conf.AtomAuthorName,
			AuthorURI:  conf.AtomAuthorURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	f, err := os.Create(conf.TargetDir + "/fragments.atom")
	if err != nil {
		return err
	}
	defer f.Close()

	return feed.Encode(f, "  ")
}

func compileFragmentsIndex(fragments []*Fragment) error {
	start := time.Now()
	defer func() {
		log.Debugf("Compiled fragments index in %v.", time.Now().Sub(start))
	}()

	fragmentsByYear := groupFragmentsByYear(fragments)

	locals := getLocals("Fragments", map[string]interface{}{
		"FragmentsByYear": fragmentsByYear,
	})

	err := renderView(sorg.MainLayout, sorg.ViewsDir+"/fragments/index",
		conf.TargetDir+"/fragments/index.html", locals)
	if err != nil {
		return err
	}

	return nil
}

func compileHome(articles []*Article, fragments []*Fragment, photos []*Photo) error {
	if conf.ContentOnly {
		return nil
	}

	start := time.Now()
	defer func() {
		log.Debugf("Compiled home in %v.", time.Now().Sub(start))
	}()

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
		conf.TargetDir+"/index.html", locals)
	if err != nil {
		return err
	}

	return nil
}

func compileJavascripts(javascripts []string, outPath string) error {
	start := time.Now()
	defer func() {
		log.Debugf("Compiled script assets in %v.", time.Now().Sub(start))
	}()

	log.Debugf("Building: %v", outPath)

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	for _, javascript := range javascripts {
		log.Debugf("Including: %v", path.Base(javascript))

		inFile, err := os.Open(javascript)
		if err != nil {
			return err
		}

		outFile.WriteString("/* " + path.Base(javascript) + " */\n\n")
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

func compilePage(dir, name string) error {
	// Remove the "./pages" directory, but keep the rest of the path.
	//
	// Looks something like "about".
	pagePath := strings.TrimPrefix(dir, sorg.PagesDir) + name

	// Looks something like "./public/about".
	target := conf.TargetDir + "/" + pagePath

	locals, ok := pagesVars[pagePath]
	if !ok {
		log.Errorf("No page meta information: %v", pagePath)
	}

	locals = getLocals("Page", locals)

	err := os.MkdirAll(conf.TargetDir+"/"+dir, 0755)
	if err != nil {
		return err
	}

	err = renderView(sorg.MainLayout, dir+"/"+name,
		target, locals)
	if err != nil {
		return err
	}

	return nil
}

func compilePhotos(db *sql.DB) ([]*Photo, error) {
	if conf.ContentOnly {
		return nil, nil
	}

	start := time.Now()
	defer func() {
		log.Debugf("Compiled photos in %v.", time.Now().Sub(start))
	}()

	photos, err := getPhotosData(db)
	if err != nil {
		return nil, err
	}

	// Keep a published copy of all the photos that we need.
	var photoAssets []*assets.Asset
	for _, photo := range photos {
		photoAssets = append(photoAssets,
			&assets.Asset{URL: photo.LargeImageURL,
				Target: conf.TargetDir + "/assets/photos/" + photo.Slug + "@2x.jpg"},
			&assets.Asset{URL: photo.MediumImageURL,
				Target: conf.TargetDir + "/assets/photos/" + photo.Slug + ".jpg"},
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
		conf.TargetDir+"/photos/index.html", locals)
	if err != nil {
		return nil, err
	}

	return photos, nil
}

func compileReading(db *sql.DB) error {
	if conf.ContentOnly {
		return nil
	}

	start := time.Now()
	defer func() {
		log.Debugf("Compiled reading in %v.", time.Now().Sub(start))
	}()

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
		conf.TargetDir+"/reading/index.html", locals)
	if err != nil {
		return err
	}

	return nil
}

func compileRobots(outPath string) error {
	if !conf.Drafts {
		return nil
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	outFile.WriteString("User-agent: *\n" +
		"Disallow: /")

	return nil
}

func compileRuns(db *sql.DB) error {
	if conf.ContentOnly {
		return nil
	}

	start := time.Now()
	defer func() {
		log.Debugf("Compiled runs in %v.", time.Now().Sub(start))
	}()

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
		conf.TargetDir+"/runs/index.html", locals)
	if err != nil {
		return err
	}

	return nil
}

func compileTwitter(db *sql.DB) error {
	if conf.ContentOnly {
		return nil
	}

	start := time.Now()
	defer func() {
		log.Debugf("Compiled tweets in %v.", time.Now().Sub(start))
	}()

	tweets, err := getTwitterData(db, false)
	if err != nil {
		return err
	}

	tweetsWithReplies, err := getTwitterData(db, true)
	if err != nil {
		return err
	}

	optionsMatrix := map[string]bool{
		"index.html":   false,
		"with-replies": true,
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
			conf.TargetDir+"/twitter/"+page, locals)
		if err != nil {
			return err
		}
	}

	return nil
}

func compileStylesheets(stylesheets []string, outPath string) error {
	start := time.Now()
	defer func() {
		log.Debugf("Compiled stylesheet assets in %v.", time.Now().Sub(start))
	}()

	log.Debugf("Building: %v", outPath)

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	for _, stylesheet := range stylesheets {
		log.Debugf("Including: %v", path.Base(stylesheet))

		inFile, err := os.Open(stylesheet)
		if err != nil {
			return err
		}

		outFile.WriteString("/* " + path.Base(stylesheet) + " */\n\n")

		if strings.HasSuffix(stylesheet, ".sass") {
			_, err := gcss.Compile(outFile, inFile)
			if err != nil {
				return fmt.Errorf("Error compiling %v: %v", path.Base(stylesheet), err)
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

func linkFontAssets() error {
	start := time.Now()
	defer func() {
		log.Debugf("Linked font assets in %v.", time.Now().Sub(start))
	}()

	source, err := filepath.Abs(sorg.ContentDir + "/fonts")
	if err != nil {
		return err
	}

	dest, err := filepath.Abs(conf.TargetDir + "/assets/fonts/")
	if err != nil {
		return err
	}

	return ensureSymlink(source, dest)
}

func linkImageAssets() error {
	start := time.Now()
	defer func() {
		log.Debugf("Linked image assets in %v.", time.Now().Sub(start))
	}()

	assets, err := ioutil.ReadDir(sorg.ContentDir + "/images")
	if err != nil {
		return err
	}

	for _, asset := range assets {
		// we use absolute paths for source and destination because not doing
		// so can result in some weird symbolic link inception
		source, err := filepath.Abs(sorg.ContentDir + "/images/" + asset.Name())
		if err != nil {
			return err
		}

		dest, err := filepath.Abs(conf.TargetDir + "/assets/" + asset.Name())
		if err != nil {
			return err
		}

		err = ensureSymlink(source, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

//
// Task generation functions
//
// These functions are the main entry points for compiling the site's
// resources.
//

func tasksForArticles(articleChan chan *Article) ([]*pool.Task, error) {
	tasks, err := tasksForArticlesDir(articleChan, sorg.ContentDir+"/articles", false)
	if err != nil {
		return nil, err
	}

	if conf.Drafts {
		draftTasks, err := tasksForArticlesDir(articleChan,
			sorg.ContentDir+"/drafts", true)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, draftTasks...)
	}

	return tasks, nil
}

func tasksForArticlesDir(articleChan chan *Article, dir string, draft bool) ([]*pool.Task, error) {
	articleInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var tasks []*pool.Task
	for _, articleInfo := range articleInfos {
		if isHidden(articleInfo.Name()) {
			continue
		}

		name := articleInfo.Name()
		tasks = append(tasks, pool.NewTask(func() error {
			article, err := compileArticle(dir, name, draft)
			if err != nil {
				return err
			}

			articleChan <- article
			return nil
		}))
	}

	return tasks, nil
}

func tasksForFragments(fragmentChan chan *Fragment) ([]*pool.Task, error) {
	tasks, err := tasksForFragmentsDir(fragmentChan, sorg.ContentDir+"/fragments", false)
	if err != nil {
		return nil, err
	}

	if conf.Drafts {
		draftTasks, err := tasksForFragmentsDir(fragmentChan,
			sorg.ContentDir+"/fragments-drafts", true)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, draftTasks...)
	}

	return tasks, nil
}

func tasksForFragmentsDir(fragmentChan chan *Fragment, dir string, draft bool) ([]*pool.Task, error) {
	fragmentInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var tasks []*pool.Task
	for _, fragmentInfo := range fragmentInfos {
		if isHidden(fragmentInfo.Name()) {
			continue
		}

		name := fragmentInfo.Name()
		tasks = append(tasks, pool.NewTask(func() error {
			fragment, err := compileFragment(dir, name, draft)
			if err != nil {
				return err
			}

			fragmentChan <- fragment
			return nil
		}))
	}

	return tasks, nil
}

func tasksForPages() ([]*pool.Task, error) {
	return tasksForPagesDir(sorg.PagesDir)
}

func tasksForPagesDir(dir string) ([]*pool.Task, error) {
	log.Debugf("Descending into pages directory: %v", dir)

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var tasks []*pool.Task
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			subtasks, err := tasksForPagesDir(dir + fileInfo.Name())
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, subtasks...)
		} else {
			// Subtract 4 for the ".ace" extension.
			name := fileInfo.Name()[0 : len(fileInfo.Name())-4]

			tasks = append(tasks, pool.NewTask(func() error {
				return compilePage(dir, name)
			}))
		}
	}

	return tasks, nil
}

//
// Other functions
//
// Any other functions. Try to keep them alphabetized.
//

func accumulateArticles(articles *[]*Article) chan *Article {
	articleChan := make(chan *Article, 100)
	go func() {
		for article := range articleChan {
			*articles = append(*articles, article)
		}
	}()
	return articleChan
}

func accumulateFragments(fragments *[]*Fragment) chan *Fragment {
	fragmentChan := make(chan *Fragment, 100)
	go func() {
		for fragment := range fragmentChan {
			*fragments = append(*fragments, fragment)
		}
	}()
	return fragmentChan
}

// Gets a map of local values for use while rendering a template and includes
// a few "special" values that are globally relevant to all templates.
func getLocals(title string, locals map[string]interface{}) map[string]interface{} {
	defaults := map[string]interface{}{
		"BodyClass":         "",
		"GoogleAnalyticsID": conf.GoogleAnalyticsID,
		"LocalFonts":        conf.LocalFonts,
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

func groupArticlesByYear(articles []*Article) []*articleYear {
	var year *articleYear
	var years []*articleYear

	for _, article := range articles {
		if year == nil || year.Year != article.PublishedAt.Year() {
			year = &articleYear{article.PublishedAt.Year(), nil}
			years = append(years, year)
		}

		year.Articles = append(year.Articles, article)
	}

	return years
}

func groupFragmentsByYear(fragments []*Fragment) []*fragmentYear {
	var year *fragmentYear
	var years []*fragmentYear

	for _, fragment := range fragments {
		if year == nil || year.Year != fragment.PublishedAt.Year() {
			year = &fragmentYear{fragment.PublishedAt.Year(), nil}
			years = append(years, year)
		}

		year.Fragments = append(year.Fragments, fragment)
	}

	return years
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

func ensureSymlink(source, dest string) error {
	log.Debugf("Checking symbolic link (%v): %v -> %v",
		path.Base(source), source, dest)

	var actual string

	_, err := os.Stat(dest)

	// Note that if a symlink file does exist, but points to a non-existent
	// location, we still get an "does not exist" error back, so we fall down
	// to the general create path so that the symlink file can be removed.
	//
	// The call to RemoveAll does not affect the other path of the symlink file
	// not being present because it doesn't care whether or not the file it's
	// trying remove is actually there.
	if os.IsNotExist(err) {
		log.Debugf("Destination link does not exist. Creating.")
		goto create
	}
	if err != nil {
		return err
	}

	actual, err = os.Readlink(dest)
	if err != nil {
		return err
	}

	if actual == source {
		log.Debugf("Link exists.")
		return nil
	}

	log.Debugf("Destination links to wrong source. Creating.")

create:
	err = os.RemoveAll(dest)
	if err != nil {
		return err
	}

	return os.Symlink(source, dest)
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
