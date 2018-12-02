package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/brandur/sorg"
	"github.com/brandur/sorg/assets"
	"github.com/brandur/sorg/atom"
	"github.com/brandur/sorg/downloader"
	"github.com/brandur/sorg/markdown"
	"github.com/brandur/sorg/passages"
	"github.com/brandur/sorg/pool"
	"github.com/brandur/sorg/resizer"
	"github.com/brandur/sorg/talks"
	"github.com/brandur/sorg/templatehelpers"
	"github.com/brandur/sorg/toc"
	"github.com/joeshaw/envdecode"
	_ "github.com/lib/pq"
	"github.com/yosssi/ace"
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

	// HookImageURL is the URL for a hook image for the article (to be shown on
	// the article index) if one was found.
	HookImageURL string `yaml:"-"`

	// Image is an optional image that may be included with an article.
	Image string `yaml:"image"`

	// Location is the geographical location where this article was written.
	Location string `yaml:"location"`

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

// PublishingInfo produces a brief spiel about publication which is intended to
// go into the left sidebar when an article is shown.
func (a *Article) PublishingInfo() string {
	return `<p><strong>Article</strong><br>` + a.Title + `</p>` +
		`<p><strong>Published</strong><br>` + a.PublishedAt.Format("January 2, 2006") + `</p> ` +
		`<p><strong>Location</strong><br>` + a.Location + `</p>` +
		sorg.TwitterInfo
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

// Conf contains configuration information for the command. It's extracted from
// environment variables.
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
	// Attributions are any attributions for content that may be included in
	// the article (like an image in the header for example).
	Attributions string `yaml:"attributions"`

	// Content is the HTML content of the fragment. It isn't included as YAML
	// frontmatter, and is rather split out of an fragment's Markdown file,
	// rendered, and then added separately.
	Content string `yaml:"-"`

	// Draft indicates that the fragment is not yet published.
	Draft bool `yaml:"-"`

	// HNLink is an optional link to comments on Hacker News.
	HNLink string `yaml:"hn_link"`

	// Hook is a leading sentence or two to succinctly introduce the fragment.
	Hook string `yaml:"hook"`

	// Image is an optional image that may be included with a fragment.
	Image string `yaml:"image"`

	// Location is the geographical location where this article was written.
	Location string `yaml:"location"`

	// PublishedAt is when the fragment was published.
	PublishedAt *time.Time `yaml:"published_at"`

	// Slug is a unique identifier for the fragment that also helps determine
	// where it's addressable by URL.
	Slug string `yaml:"-"`

	// Title is the fragment's title.
	Title string `yaml:"title"`
}

// PublishingInfo produces a brief spiel about publication which is intended to
// go into the left sidebar when a fragment is shown.
func (f *Fragment) PublishingInfo() string {
	s := `<p><strong>Fragment</strong><br>` + f.Title + `</p>` +
		`<p><strong>Published</strong><br>` + f.PublishedAt.Format("January 2, 2006") + `</p> `

	if f.Location != "" {
		s += `<p><strong>Location</strong><br>` + f.Location + `</p>`
	}

	s += sorg.TwitterInfo
	return s
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

// Page is the metadata for a static HTML page generated from an ACE file.
// Currently the layouting system of ACE doesn't allow us to pass metadata up
// very well, so we have this instead.
type Page struct {
	// BodyClass is the CSS class that will be assigned to the body tag when
	// the page is rendered.
	BodyClass string `yaml:"body_class"`

	// Title is the HTML title that will be assigned to the page when it's
	// rendered.
	Title string `yaml:"title"`
}

type passageByPublishedAt []*passages.Passage

func (p passageByPublishedAt) Len() int           { return len(p) }
func (p passageByPublishedAt) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p passageByPublishedAt) Less(i, j int) bool { return p[i].PublishedAt.Before(*p[j].PublishedAt) }

// Photo is a photograph.
type Photo struct {
	// Description is the description of the photograph.
	Description string `yaml:"description"`

	// OriginalImageURL is the location where the original-sized version of the
	// photo can be downloaded from.
	OriginalImageURL string `yaml:"original_image_url"`

	// OccurredAt is UTC time when the photo was published.
	OccurredAt *time.Time `yaml:"occurred_at"`

	// Slug is a unique identifier for the photo. Originally these were
	// generated from Flickr, but I've since just started reusing them for
	// filenames.
	Slug string `yaml:"slug"`

	// Title is the title of the photograph.
	Title string `yaml:"title"`
}

// PhotoWrapper is a data structure intended to represent the data structure at
// the top level of photograph data file `photographs.yaml`.
type PhotoWrapper struct {
	// Photos is a collection of photos within the top-level wrapper.
	Photos []*Photo `yaml:"photographs"`
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

// twitterCard represents a Twitter "card" (i.e. one of those rich media boxes
// that sometimes appear under tweets official clients) for use in templates.
type twitterCard struct {
	// Description is the title to show in the card.
	Title string

	// Description is the description to show in the card.
	Description string

	// ImageURL is the URL to the image to show in the card. It should be
	// absolute because Twitter will need to be able to fetch it from our
	// servers. Leave blank if there is no image.
	ImageURL string
}

//
// Variables
//

// Left as a global for now for the sake of convenience, but it's not used in
// very many places and can probably be refactored as a local if desired.
var conf Conf

//
// Main
//

func main() {
	start := time.Now()
	defer func() {
		log.Infof("Built site in %v.", time.Now().Sub(start))
	}()

	// We'll call the general random number generator in at least one place
	// below.
	rand.Seed(time.Now().UnixNano())

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

	err = os.MkdirAll(sorg.TempDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	var tasks []*pool.Task

	//
	// Build step 0: dependency-free
	//

	tasks = nil

	// Articles, fragments, and pages are are slightly special cases in that we
	// parallelize the creation of all of them all at once. That is, every
	// article will have a separately entry in our work queue.

	var articles []*Article
	articleChan := accumulateArticles(&articles)

	var fragments []*Fragment
	fragmentChan := accumulateFragments(&fragments)

	var passages []*passages.Passage
	passageChan := accumulatePassages(&passages)

	var talks []*talks.Talk
	talkChan := accumulateTalks(&talks)

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

	pageTasks, err := tasksForPages()
	if err != nil {
		log.Fatal(err)
	}
	tasks = append(tasks, pageTasks...)

	passageTasks, err := tasksForPassages(passageChan)
	if err != nil {
		log.Fatal(err)
	}
	tasks = append(tasks, passageTasks...)

	// Most other types are all one-off pages or other resources and only get a
	// single entry each in the work queue.

	var photos []*Photo
	tasks = append(tasks, pool.NewTask(func() error {
		var err error
		photos, err = compilePhotos(false)
		return err
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return assets.CompileJavascripts(
			path.Join(sorg.ContentDir, "javascripts"),
			path.Join(versionedAssetsDir, "app.js"))
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
		return assets.CompileStylesheets(
			path.Join(sorg.ContentDir, "stylesheets"),
			path.Join(versionedAssetsDir, "app.css"))
	}))

	talkTasks, err := tasksForTalks(talkChan)
	if err != nil {
		log.Fatal(err)
	}
	tasks = append(tasks, talkTasks...)

	tasks = append(tasks, pool.NewTask(func() error {
		return compileTwitter(db)
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return linkImages()
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return linkFonts()
	}))

	if !runTasks(tasks) {
		os.Exit(1)
	}

	// Free up any Goroutines still waiting.
	close(articleChan)
	close(fragmentChan)
	close(passageChan)

	//
	// Build step 1: any tasks dependent on the results of step 0.
	//
	// This includes build output like index pages and RSS feeds.
	//

	tasks = nil

	sort.Sort(sort.Reverse(articleByPublishedAt(articles)))
	sort.Sort(sort.Reverse(fragmentByPublishedAt(fragments)))
	sort.Sort(sort.Reverse(passageByPublishedAt(passages)))

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
		return compilePassagesIndex(passages)
	}))

	tasks = append(tasks, pool.NewTask(func() error {
		return compileHome(articles, fragments, photos)
	}))

	if !runTasks(tasks) {
		os.Exit(1)
	}
}

//
// Compilation functions
//
// These functions perform the heavy-lifting in compiling the site's resources.
// They are normally run concurrently.
//

func compileArticle(dir, name string, draft bool) (*Article, error) {
	inPath := path.Join(dir, name)

	raw, err := ioutil.ReadFile(inPath)
	if err != nil {
		return nil, err
	}

	frontmatter, content, err := sorg.SplitFrontmatter(string(raw))
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

	if article.Location == "" {
		return nil, fmt.Errorf("No location for article: %v", inPath)
	}

	if article.Title == "" {
		return nil, fmt.Errorf("No title for article: %v", inPath)
	}

	if article.PublishedAt == nil {
		return nil, fmt.Errorf("No publish date for article: %v", inPath)
	}

	article.Content = markdown.Render(content, nil)

	article.TOC, err = toc.Render(article.Content)
	if err != nil {
		return nil, err
	}

	format, ok := pathAsImage(
		path.Join(sorg.ContentDir, "images", article.Slug, "hook"),
	)
	if ok {
		article.HookImageURL = "/assets/" + article.Slug + "/hook." + format
	}

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	card := &twitterCard{
		Title:       article.Title,
		Description: article.Hook,
	}
	format, ok = pathAsImage(
		path.Join(sorg.ContentDir, "images", article.Slug, "twitter@2x"),
	)
	if ok {
		card.ImageURL = sorg.AbsoluteURL + "/assets/" + article.Slug + "/twitter@2x." + format
	}

	locals := getLocals(article.Title, map[string]interface{}{
		"Article":        article,
		"PublishingInfo": article.PublishingInfo(),
		"TwitterCard":    card,
	})

	err = renderView(sorg.MainLayout, sorg.ViewsDir+"/articles/show",
		path.Join(conf.TargetDir, article.Slug), locals)
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
	inPath := path.Join(dir, name)

	raw, err := ioutil.ReadFile(inPath)
	if err != nil {
		return nil, err
	}

	frontmatter, content, err := sorg.SplitFrontmatter(string(raw))
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

	fragment.Content = markdown.Render(content, nil)

	// A lot of fragments still have unwritten hooks, so only add a card where
	// a fragment has a configured Twitter image for the time being.
	var card *twitterCard
	format, ok := pathAsImage(
		path.Join(sorg.ContentDir, "images", "fragments", fragment.Slug, "twitter@2x"),
	)
	if ok {
		card = &twitterCard{
			ImageURL:    sorg.AbsoluteURL + "/assets/fragments/" + fragment.Slug + "/twitter@2x." + format,
			Title:       fragment.Title,
			Description: fragment.Hook,
		}
	}

	locals := getLocals(fragment.Title, map[string]interface{}{
		"Fragment":       fragment,
		"PublishingInfo": fragment.PublishingInfo(),
		"TwitterCard":    card,
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
	start := time.Now()
	defer func() {
		log.Debugf("Compiled home in %v.", time.Now().Sub(start))
	}()

	if len(articles) > 3 {
		articles = articles[0:3]
	}

	// Try just one fragment for now to better balance the page's height.
	if len(fragments) > 1 {
		fragments = fragments[0:1]
	}

	var photo *Photo
	if len(photos) > 0 {
		numRecent := 10
		if len(photos) < numRecent {
			numRecent = len(photos)
		}

		recentPhotos := photos[0:numRecent]
		photo = recentPhotos[rand.Intn(len(recentPhotos))]
	}

	locals := getLocals("brandur.org", map[string]interface{}{
		"Articles":  articles,
		"BodyClass": "index",
		"Fragments": fragments,
		"Photo":     photo,
	})

	err := renderView(sorg.MainLayout, sorg.ViewsDir+"/index",
		conf.TargetDir+"/index.html", locals)
	if err != nil {
		return err
	}

	return nil
}

func compilePage(pagesMeta map[string]*Page, dir, name string) error {
	// Remove the "./pages" directory, but keep the rest of the path.
	//
	// Looks something like "about".
	pagePath := strings.TrimPrefix(dir, sorg.PagesDir) + name

	// Looks something like "./public/about".
	target := path.Join(conf.TargetDir, pagePath)

	// Put a ".html" on if this page is an index. This will allow our local
	// server to serve it at a directory path, and our upload script is smart
	// enough to do the right thing with it as well.
	if path.Base(pagePath) == "index" {
		target += ".html"
	}

	locals := map[string]interface{}{
		"BodyClass": "",
		"Title":     "Untitled Page",
	}

	pageMeta, ok := pagesMeta[pagePath]
	if ok {
		locals = map[string]interface{}{
			"BodyClass": pageMeta.BodyClass,
			"Title":     pageMeta.Title,
		}
	} else {
		log.Errorf("No page meta information: %v", pagePath)
	}

	locals = getLocals("Page", locals)

	err := os.MkdirAll(path.Join(conf.TargetDir, dir), 0755)
	if err != nil {
		return err
	}

	err = renderView(sorg.MainLayout, path.Join(dir, name),
		target, locals)
	if err != nil {
		return err
	}

	return nil
}

func compilePassagesIndex(passages []*passages.Passage) error {
	start := time.Now()
	defer func() {
		log.Debugf("Compiled passages index in %v.", time.Now().Sub(start))
	}()

	locals := getLocals("Passages", map[string]interface{}{
		"Passages": passages,
	})

	err := renderView(sorg.PassageLayout, sorg.ViewsDir+"/passages/index",
		conf.TargetDir+"/passages/index.html", locals)
	if err != nil {
		return err
	}

	return nil
}

// Compiles photos based on `content/photographs.yaml` by downloading any that
// are missing and doing resizing work.
//
// `skipWork` initializes parallel jobs but no-ops them. This is useful for
// testing where we don't expect these operations to succeed.
func compilePhotos(skipWork bool) ([]*Photo, error) {
	if conf.ContentOnly {
		return nil, nil
	}

	start := time.Now()
	defer func() {
		log.Debugf("Compiled photos in %v.", time.Now().Sub(start))
	}()

	data, err := ioutil.ReadFile(path.Join(sorg.ContentDir, "photographs.yaml"))
	if err != nil {
		return nil, fmt.Errorf("Error reading photographs data file: %v", err)
	}

	var photosWrapper PhotoWrapper
	err = yaml.Unmarshal(data, &photosWrapper)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling photographs data: %v", err)
	}

	photos := photosWrapper.Photos

	// Sort reverse chronologically so newer photos are first.
	sort.Slice(photos, func(i, j int) bool {
		return photos[j].OccurredAt.Before(*photos[i].OccurredAt)
	})

	// Dropbox is the original source for images, but to avoid doing unnecessary
	// downloading, resizing, and uploading work for every build, we put any
	// work we do into a "cache" which is itself put into S3. Subsequent builds
	// can leverage the cache to avoid repeat work.
	//
	// See also the `photos-*` family of commands in `Makefile`.
	//
	// Note that JPGs in this directory are in `.gitignore` and not eligible to
	// be uploaded to the Git repository, but `.marker` files can and should be
	// committed as often as possible.
	cacheDir := path.Join(sorg.ContentDir, "photographs")

	err = linkPhotographs()
	if err != nil {
		return nil, err
	}

	var markers []string
	var photoFiles []*downloader.File
	var resizeJobs []*resizer.ResizeJob

	for _, photo := range photos {
		image1x := photo.Slug + ".jpg"
		image2x := photo.Slug + "@2x.jpg"

		imageLarge1x := photo.Slug + "_large.jpg"
		imageLarge2x := photo.Slug + "_large@2x.jpg"

		imageMarker := photo.Slug + ".marker"
		imageOriginal := photo.Slug + "_original.jpg"

		// We use a "marker" system so that we don't have to copy the entire
		// set of images to and from every build.
		//
		// When the system creates a new set of resized images it also creates
		// a marker file (same name as the images, but which ends in
		// `.marker`). When initializing a new build, only marker files are
		// copied dwon from S3. When determining which images need to be
		// fetched and resize, we skip any that already have a marker file.
		if !fileExists(path.Join(cacheDir, imageMarker)) {
			photoFiles = append(photoFiles,
				&downloader.File{URL: photo.OriginalImageURL,
					Target: path.Join(sorg.TempDir, imageOriginal)},
			)

			resizeJobs = append(resizeJobs,
				&resizer.ResizeJob{
					SourcePath:  path.Join(sorg.TempDir, imageOriginal),
					TargetPath:  path.Join(cacheDir, image1x),
					TargetWidth: 333,
				},
				&resizer.ResizeJob{
					SourcePath:  path.Join(sorg.TempDir, imageOriginal),
					TargetPath:  path.Join(cacheDir, image2x),
					TargetWidth: 667,
				},
				&resizer.ResizeJob{
					SourcePath:  path.Join(sorg.TempDir, imageOriginal),
					TargetPath:  path.Join(cacheDir, imageLarge1x),
					TargetWidth: 1500,
				},
				&resizer.ResizeJob{
					SourcePath:  path.Join(sorg.TempDir, imageOriginal),
					TargetPath:  path.Join(cacheDir, imageLarge2x),
					TargetWidth: 3000,
				},
			)

			markers = append(markers,
				path.Join(cacheDir, imageMarker))
		}
	}

	log.Debugf("Skipping processing %d image(s) with marker(s)",
		len(photos)-len(markers))

	if !skipWork {
		log.Debugf("Fetching %d photo(s)", len(photoFiles))
		err = downloader.Fetch(photoFiles)
		if err != nil {
			return nil, err
		}

		log.Debugf("Running %d resize job(s)", len(resizeJobs))
		err = resizer.Resize(resizeJobs)
		if err != nil {
			return nil, err
		}

		log.Debugf("Creating %d marker(s)", len(markers))
		for _, marker := range markers {
			file, err := os.OpenFile(marker, os.O_RDONLY|os.O_CREATE, 0755)
			if err != nil {
				return nil, err
			}
			file.Close()
		}
	}

	locals := getLocals("Photography", map[string]interface{}{
		"BodyClass":     "photos",
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
	var content string
	if conf.Drafts {
		// Allow Twitterbot so that we can preview card images on dev.
		content = `User-agent: Twitterbot
Disallow:

User-agent: *
Disallow: /
`
	} else {
		// Disallow acccess to photos because the content isn't very
		// interesting for robots and they're bandwidth heavy.
		content = `User-agent: *
Disallow: /photographs/
Disallow: /photos
`
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	outFile.WriteString(content)
	outFile.Close()

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

	locals := getLocals("Running", map[string]interface{}{
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

func compilePassage(dir, name string, draft bool) (*passages.Passage, error) {
	passage, err := passages.Compile(dir, name, draft, false)
	if err != nil {
		return nil, err
	}

	locals := getLocals(passage.Title, map[string]interface{}{
		"InEmail": false,
		"Passage": passage,
	})

	err = renderView(sorg.PassageLayout, sorg.ViewsDir+"/passages/show",
		conf.TargetDir+"/passages/"+passage.Slug, locals)
	if err != nil {
		return nil, err
	}

	return passage, nil
}

func compileTalk(dir, name string, draft bool) (*talks.Talk, error) {
	talk, err := talks.Compile(sorg.ContentDir, dir, name, draft)
	if err != nil {
		return nil, err
	}

	locals := getLocals(talk.Title, map[string]interface{}{
		"BodyClass":      "talk",
		"PublishingInfo": talk.PublishingInfo(),
		"Talk":           talk,
	})

	err = renderView(sorg.MainLayout, sorg.ViewsDir+"/talks/show",
		conf.TargetDir+"/"+talk.Slug, locals)
	if err != nil {
		return nil, err
	}

	return talk, nil
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

func linkFonts() error {
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

func linkImages() error {
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

func linkPhotographs() error {
	start := time.Now()
	defer func() {
		log.Debugf("Linked photographs in %v.", time.Now().Sub(start))
	}()

	source, err := filepath.Abs(sorg.ContentDir + "/photographs/")
	if err != nil {
		return err
	}

	dest, err := filepath.Abs(conf.TargetDir + "/photographs/")
	if err != nil {
		return err
	}

	return ensureSymlink(source, dest)
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
		return nil, fmt.Errorf("Error reading articles dir: %v", err)
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
		return nil, fmt.Errorf("Error reading fragments dir: %v", err)
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
	meta, err := ioutil.ReadFile(path.Join(sorg.PagesDir, "meta.yaml"))
	if err != nil {
		return nil, fmt.Errorf("Error reading pages metadata: %v", err)
	}

	var pagesMeta map[string]*Page
	err = yaml.Unmarshal(meta, &pagesMeta)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling pages metadata: %v", err)
	}

	return tasksForPagesDir(pagesMeta, sorg.PagesDir)
}

func tasksForPagesDir(pagesMeta map[string]*Page, dir string) ([]*pool.Task, error) {
	log.Debugf("Descending into pages directory: %v", dir)

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("Error reading pages dir: %v", err)
	}

	var tasks []*pool.Task
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			subtasks, err := tasksForPagesDir(pagesMeta, dir+fileInfo.Name())
			if err != nil {
				return nil, fmt.Errorf("Error getting pages tasks: %v", err)
			}
			tasks = append(tasks, subtasks...)
		} else {
			if isHidden(fileInfo.Name()) {
				continue
			}

			if filepath.Ext(fileInfo.Name()) != ".ace" {
				continue
			}

			// Subtract 4 for the ".ace" extension.
			name := fileInfo.Name()[0 : len(fileInfo.Name())-4]

			tasks = append(tasks, pool.NewTask(func() error {
				return compilePage(pagesMeta, dir, name)
			}))
		}
	}

	return tasks, nil
}

func tasksForPassages(passageChan chan *passages.Passage) ([]*pool.Task, error) {
	tasks, err := tasksForPassagesDir(passageChan, sorg.ContentDir+"/passages", false)
	if err != nil {
		return nil, fmt.Errorf("Error getting passage tasks: %v", err)
	}

	if conf.Drafts {
		draftTasks, err := tasksForPassagesDir(passageChan,
			sorg.ContentDir+"/passages-drafts", true)
		if err != nil {
			return nil, fmt.Errorf("Error getting passage draft tasks: %v", err)
		}

		tasks = append(tasks, draftTasks...)
	}

	return tasks, nil
}

func tasksForPassagesDir(passageChan chan *passages.Passage, dir string, draft bool) ([]*pool.Task, error) {
	passageInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var tasks []*pool.Task
	for _, passageInfo := range passageInfos {
		if isHidden(passageInfo.Name()) {
			continue
		}

		name := passageInfo.Name()
		tasks = append(tasks, pool.NewTask(func() error {
			passage, err := compilePassage(dir, name, draft)
			if err != nil {
				return err
			}

			passageChan <- passage
			return nil
		}))
	}

	return tasks, nil
}

func tasksForTalks(talkChan chan *talks.Talk) ([]*pool.Task, error) {
	tasks, err := tasksForTalksDir(talkChan, sorg.ContentDir+"/talks", false)
	if err != nil {
		return nil, fmt.Errorf("Error getting talk tasks: %v", err)
	}

	if conf.Drafts {
		draftTasks, err := tasksForTalksDir(talkChan,
			sorg.ContentDir+"/talks-drafts", true)
		if err != nil {
			return nil, fmt.Errorf("Error getting talk draft tasks: %v", err)
		}

		tasks = append(tasks, draftTasks...)
	}

	return tasks, nil
}

func tasksForTalksDir(talkChan chan *talks.Talk, dir string, draft bool) ([]*pool.Task, error) {
	talkInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var tasks []*pool.Task
	for _, talkInfo := range talkInfos {
		if isHidden(talkInfo.Name()) {
			continue
		}

		// I may store dirty drafts and the like in here, so skip any
		// non-Markdown files that we find.
		if !strings.HasSuffix(talkInfo.Name(), ".md") {
			continue
		}

		name := talkInfo.Name()
		tasks = append(tasks, pool.NewTask(func() error {
			talk, err := compileTalk(dir, name, draft)
			if err != nil {
				return err
			}

			talkChan <- talk
			return nil
		}))
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

func accumulatePassages(p *[]*passages.Passage) chan *passages.Passage {
	passageChan := make(chan *passages.Passage, 100)
	go func() {
		for passage := range passageChan {
			*p = append(*p, passage)
		}
	}()
	return passageChan
}

func accumulateTalks(p *[]*talks.Talk) chan *talks.Talk {
	talkChan := make(chan *talks.Talk, 100)
	go func() {
		for talk := range talkChan {
			*p = append(*p, talk)
		}
	}()
	return talkChan
}

// Naturally not provided by the Go language because copying files "has tricky
// edge cases". You just can't make this stuff up.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Error opening copy source: %v", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("Error creating copy target: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("Error copying data: %v", err)
	}

	return nil
}

// Just a shortcut to try and cut down on Go's extreme verbosity.
func fileExists(file string) bool {
	_, err := os.Stat(file)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	panic(err)
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
		"TwitterCard":       nil,
		"ViewportWidth":     "device-width",
	}

	for k, v := range locals {
		defaults[k] = v
	}

	return defaults
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
		return nil, fmt.Errorf("Error selecting readings: %v", err)
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
			return nil, fmt.Errorf("Error scanning readings: %v", err)
		}

		readings = append(readings, &reading)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Error iterating readings: %v", err)
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
		return nil, nil, fmt.Errorf("Error selecting reading count by year: %v", err)
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
			return nil, nil, fmt.Errorf("Error scanning reading count by year: %v", err)
		}

		byYearXYears = append(byYearXYears, year)
		byYearYCounts = append(byYearYCounts, count)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("Error iterating reading count by year: %v", err)
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
		return nil, nil, fmt.Errorf("Error selecting reading pages by year: %v", err)
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
			return nil, nil, fmt.Errorf("Error scanning reading pages by year: %v", err)
		}

		byYearXYears = append(byYearXYears, year)
		byYearYCounts = append(byYearYCounts, count)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("Error iterating reading pages by year: %v", err)
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
		return nil, fmt.Errorf("Error selecting runs: %v", err)
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
			return nil, fmt.Errorf("Error scanning runs: %v", err)
		}

		if locationCity != nil {
			run.LocationCity = *locationCity
		}

		runs = append(runs, &run)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Error iterating runs: %v", err)
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
		return nil, nil, fmt.Errorf("Error selecting runs by year: %v", err)
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
			return nil, nil, fmt.Errorf("Error scanning runs by year: %v", err)
		}

		byYearXYears = append(byYearXYears, year)
		byYearYDistances = append(byYearYDistances, distance)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("Error iterating runs by year: %v", err)
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
			ORDER BY day
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

		SELECT to_char(d.day, 'Mon') AS day,
			d.distance + COALESCE(rd.distance, 0::float)
		FROM days d
			LEFT JOIN runs_days rd ON d.day = rd.day
	`)
	if err != nil {
		return nil, nil, fmt.Errorf("Error selecting last year runs: %v", err)
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
			return nil, nil, fmt.Errorf("Error scanning last year runs: %v", err)
		}

		lastYearXDays = append(lastYearXDays, day)
		lastYearYDistances = append(lastYearYDistances, distance)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("Error iterating last year runs: %v", err)
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
		return nil, nil, fmt.Errorf("Error selecting tweets by month: %v", err)
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
			return nil, nil, fmt.Errorf("Error scanning tweets by month: %v", err)
		}

		tweetCountXMonths = append(tweetCountXMonths, month)
		tweetCountYCounts = append(tweetCountYCounts, count)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("Error iterating tweets by month: %v", err)
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
		return nil, fmt.Errorf("Error selecting tweets: %v", err)
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
			return nil, fmt.Errorf("Error scanning tweets: %v", err)
		}

		tweets = append(tweets, &tweet)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Error iterating tweets: %v", err)
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

// Detects a hidden file, i.e. one that starts with a dot.
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
		return fmt.Errorf("Error checking symlink: %v", err)
	}

	actual, err = os.Readlink(dest)
	if err != nil {
		return fmt.Errorf("Error reading symlink: %v", err)
	}

	if actual == source {
		log.Debugf("Link exists.")
		return nil
	}

	log.Debugf("Destination links to wrong source. Creating.")

create:
	err = os.RemoveAll(dest)
	if err != nil {
		return fmt.Errorf("Error removing symlink: %v", err)
	}

	err = os.Symlink(source, dest)
	if err != nil {
		return fmt.Errorf("Error creating symlink: %v", err)
	}

	return nil
}

// Checks if the path exists as a common image format (.jpg or .png only). If
// so, returns the discovered extension (e.g. "jpg") and boolean true.
// Otherwise returns an empty string and boolean false.
func pathAsImage(extensionlessPath string) (string, bool) {
	// extensions must be lowercased
	formats := []string{"jpg", "png"}

	for _, format := range formats {
		_, err := os.Stat(extensionlessPath + "." + format)
		if err != nil {
			continue
		}

		return format, true
	}

	return "", false
}

func renderView(layout, view, target string, locals map[string]interface{}) error {
	log.Debugf("Rendering: %v", target)

	template, err := ace.Load(layout, view, &ace.Options{FuncMap: templatehelpers.FuncMap})
	if err != nil {
		return err
	}

	file, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("Error creating view file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	err = template.Execute(writer, locals)
	if err != nil {
		return fmt.Errorf("Error rendering view: %v", err)
	}

	return nil
}

// Runs the given tasks in a pool.
//
// After the run, if any errors occurred, it prints the first 10. Returns true
// if all tasks succeeded. If a false is returned, the caller should consider
// exiting with non-zero status.
func runTasks(tasks []*pool.Task) bool {
	p := pool.NewPool(tasks, conf.Concurrency)
	p.Run()

	var numErrors int
	for _, task := range p.Tasks {
		if task.Err != nil {
			log.Error(task.Err)
			numErrors++
		}
		if numErrors >= 10 {
			log.Error("Too many errors.")
			break
		}
	}

	return !p.HasErrors()
}
