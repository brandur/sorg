package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/mace"
	"github.com/brandur/modulir/modules/matom"
	"github.com/brandur/modulir/modules/mfile"
	"github.com/brandur/modulir/modules/mimage"
	"github.com/brandur/modulir/modules/mmarkdown"
	"github.com/brandur/modulir/modules/mmarkdownext"
	"github.com/brandur/modulir/modules/mtemplatemd"
	"github.com/brandur/modulir/modules/mtoc"
	"github.com/brandur/modulir/modules/mtoml"
	"github.com/brandur/sorg/modules/sassets"
	"github.com/brandur/sorg/modules/scommon"
	"github.com/brandur/sorg/modules/snewsletter"
	"github.com/brandur/sorg/modules/squantified"
	"github.com/brandur/sorg/modules/stalks"
	_ "github.com/lib/pq"
	"github.com/yosssi/ace"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Constants
//
//
//
//////////////////////////////////////////////////////////////////////////////

const (
	// Special expression for viewport width that says to use the device width.
	viewportWidthDeviceWidth = "device-width"
)

// A set of tag constants to hopefully help ensure that this set doesn't grow
// very much.
const (
	tagPostgres Tag = "postgres"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Variables
//
//
//
//////////////////////////////////////////////////////////////////////////////

// These are all objects that are persisted between build loops so that if
// necessary we can rebuild jobs that depend on them like index pages without
// reparsing all the source material. In each case we try to only reparse the
// sources if those source files actually changed.
var (
	articles   []*Article
	fragments  []*Fragment
	nanoglyphs []*snewsletter.Issue
	passages   []*snewsletter.Issue
	pages      map[string]*Page = make(map[string]*Page)
	photos     []*Photo
	sequences  = make(map[string]*Sequence)
	talks      []*stalks.Talk
)

// A database connection opened to a Black Swan database if one was configured.
var db *sql.DB

// List of common build dependencies, a change in any of which will trigger a
// rebuild on everything: partial views, JavaScripts, and stylesheets. Even
// though some of those changes will false positives, these sources are
// pervasive enough, and changes infrequent enough, that it's worth the
// tradeoff. This variable is a global because so many render functions access
// it.
var universalSources []string

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Init
//
//
//
//////////////////////////////////////////////////////////////////////////////

func init() {
	mmarkdownext.FuncMap = scommon.TemplateFuncMap

	mimage.MagickBin = conf.MagickBin
	mimage.MozJPEGBin = conf.MozJPEGBin
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Build function
//
//
//
//////////////////////////////////////////////////////////////////////////////

func build(c *modulir.Context) []error {
	//
	// PHASE 0: Setup
	//
	// (No jobs should be enqueued here.)
	//

	c.Log.Debugf("Running build loop")

	// This is where we stored "versioned" assets like compiled JS and CSS.
	// These assets have a release number that we can increment and by
	// extension quickly invalidate.
	versionedAssetsDir := path.Join(c.TargetDir, "assets", Release)

	{
		// Only open the first time
		if conf.BlackSwanDatabaseURL != "" {
			if db == nil {
				var err error
				db, err = sql.Open("postgres", conf.BlackSwanDatabaseURL)
				if err != nil {
					return []error{err}
				}
			}
		} else {
			c.Log.Infof("No database set; will not render database-backed views")
		}
	}

	// A set of source paths that rebuild everything when any one of them
	// changes. These are dependencies that are included in more or less
	// everything: common partial views, JavaScript sources, and stylesheet
	// sources.
	universalSources = nil

	// Generate a set of JavaScript sources to add to univeral sources.
	{
		javaScriptSources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/javascripts",
			&mfile.ReadDirOptions{ShowMeta: true})
		if err != nil {
			return []error{err}
		}
		universalSources = append(universalSources, javaScriptSources...)
	}

	// Generate a list of partial views to add to universal sources.
	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/views",
			&mfile.ReadDirOptions{ShowMeta: true})
		if err != nil {
			return []error{err}
		}

		var partialViews []string
		for _, source := range sources {
			if strings.HasPrefix(filepath.Base(source), "_") {
				partialViews = append(partialViews, source)
			}
		}

		universalSources = append(universalSources, partialViews...)
	}

	// Generate a set of stylesheet sources to add to universal sources.
	{
		stylesheetSources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/stylesheets",
			&mfile.ReadDirOptions{ShowMeta: true})
		if err != nil {
			return []error{err}
		}
		universalSources = append(universalSources, stylesheetSources...)
	}

	//
	// PHASE 1
	//
	// The build is broken into phases because some jobs depend on jobs that
	// ran before them. For example, we need to parse all our article metadata
	// before we can create an article index and render the home page (which
	// contains a short list of articles).
	//
	// After each phase, we call `Wait` on our context which will wait for the
	// worker pool to finish all its current work and restart it to accept new
	// jobs after it has.
	//
	// The general rule is to make sure that work is done as early as it
	// possibly can be. e.g. Jobs with no dependencies should always run in
	// phase 1. Try to make sure that as few phases as necessary.
	//

	//
	// Common directories
	//
	// Create these outside of the job system because jobs below may depend on
	// their existence.
	//

	{
		commonDirs := []string{
			c.TargetDir + "/articles",
			c.TargetDir + "/fragments",
			c.TargetDir + "/nanoglyphs",
			c.TargetDir + "/passages",
			c.TargetDir + "/photos",
			c.TargetDir + "/reading",
			c.TargetDir + "/runs",
			c.TargetDir + "/sequences",
			c.TargetDir + "/twitter",
			scommon.TempDir,
			versionedAssetsDir,
		}
		for _, dir := range commonDirs {
			err := mfile.EnsureDir(c, dir)
			if err != nil {
				return []error{nil}
			}
		}
	}

	//
	// Symlinks
	//

	{
		commonSymlinks := [][2]string{
			{c.SourceDir + "/content/fonts", c.TargetDir + "/assets/fonts"},
			{c.SourceDir + "/content/images", c.TargetDir + "/assets/images"},

			// For backwards compatibility as many emails with this style of path
			// have already gone out.
			{c.SourceDir + "/content/images/passages", c.TargetDir + "/assets/passages"},

			{c.SourceDir + "/content/photographs", c.TargetDir + "/photographs"},
			{c.SourceDir + "/content/stylesheets-modular", versionedAssetsDir + "/stylesheets"},
		}
		for _, link := range commonSymlinks {
			err := mfile.EnsureSymlink(c, link[0], link[1])
			if err != nil {
				return []error{nil}
			}
		}
	}

	//
	// Articles
	//

	var articlesChanged bool
	var articlesMu sync.Mutex

	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/articles", nil)
		if err != nil {
			return []error{err}
		}

		if conf.Drafts {
			drafts, err := mfile.ReadDirCached(c, c.SourceDir+"/content/drafts", nil)
			if err != nil {
				return []error{err}
			}
			sources = append(sources, drafts...)
		}

		for _, s := range sources {
			source := s

			name := fmt.Sprintf("article: %s", filepath.Base(source))
			c.AddJob(name, func() (bool, error) {
				return renderArticle(c, source,
					&articles, &articlesChanged, &articlesMu)
			})
		}
	}

	//
	// Fragments
	//

	var fragmentsChanged bool
	var fragmentsMu sync.Mutex

	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/fragments", nil)
		if err != nil {
			return []error{err}
		}

		if conf.Drafts {
			drafts, err := mfile.ReadDirCached(c, c.SourceDir+"/content/fragments-drafts", nil)
			if err != nil {
				return []error{err}
			}
			sources = append(sources, drafts...)
		}

		for _, s := range sources {
			source := s

			name := fmt.Sprintf("fragment: %s", filepath.Base(source))
			c.AddJob(name, func() (bool, error) {
				return renderFragment(c, source,
					&fragments, &fragmentsChanged, &fragmentsMu)
			})
		}
	}

	//
	// Javascripts
	//

	{
		c.AddJob("javascripts", func() (bool, error) {
			return compileJavascripts(c,
				c.SourceDir+"/content/javascripts",
				versionedAssetsDir+"/app.js")
		})
	}

	//
	// Nanoglyphs
	//

	var nanoglyphsChanged bool
	var nanoglyphsMu sync.Mutex

	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/nanoglyphs", nil)
		if err != nil {
			return []error{err}
		}

		if conf.Drafts {
			drafts, err := mfile.ReadDirCached(c, c.SourceDir+"/content/nanoglyphs-drafts", nil)
			if err != nil {
				return []error{err}
			}
			sources = append(sources, drafts...)
		}

		for _, s := range sources {
			source := s

			name := fmt.Sprintf("nanoglyph: %s", filepath.Base(source))
			c.AddJob(name, func() (bool, error) {
				return renderNanoglyph(c, source,
					&nanoglyphs, &nanoglyphsChanged, &nanoglyphsMu)
			})
		}
	}

	//
	// Passages
	//

	var passagesChanged bool
	var passagesMu sync.Mutex

	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/passages", nil)
		if err != nil {
			return []error{err}
		}

		if conf.Drafts {
			drafts, err := mfile.ReadDirCached(c, c.SourceDir+"/content/passages-drafts", nil)
			if err != nil {
				return []error{err}
			}
			sources = append(sources, drafts...)
		}

		for _, s := range sources {
			source := s

			name := fmt.Sprintf("passage: %s", filepath.Base(source))
			c.AddJob(name, func() (bool, error) {
				return renderPassage(c, source,
					&passages, &passagesChanged, &passagesMu)
			})
		}
	}

	//
	// Photos (read `_meta.toml`)
	//

	var photosChanged bool

	{
		c.AddJob("photos _meta.toml", func() (bool, error) {
			source := c.SourceDir + "/content/photographs/_meta.toml"

			if !c.Changed(source) {
				return false, nil
			}

			var photosWrapper PhotoWrapper
			err := mtoml.ParseFile(c, source, &photosWrapper)
			if err != nil {
				return true, err
			}

			photos = photosWrapper.Photos
			photosChanged = true
			return true, nil
		})
	}

	//
	// Reading
	//

	{
		c.AddJob("reading", func() (bool, error) {
			return c.AllowError(renderReading(c, db)), nil
		})
	}

	//
	// Robots.txt
	//

	{
		c.AddJob("robots.txt", func() (bool, error) {
			return renderRobotsTxt(c)
		})
	}

	//
	// Runs
	//

	{
		c.AddJob("runs", func() (bool, error) {
			return c.AllowError(renderRuns(c, db)), nil
		})
	}

	//
	// Sequences (read `_meta.toml`)
	//

	sequencesChanged := make(map[string]bool)

	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/sequences",
			&mfile.ReadDirOptions{ShowDirs: true})
		if err != nil {
			return []error{err}
		}

		if conf.Drafts {
			drafts, err := mfile.ReadDirCached(c, c.SourceDir+"/content/sequences-drafts",
				&mfile.ReadDirOptions{ShowDirs: true})
			if err != nil {
				return []error{err}
			}
			sources = append(sources, drafts...)
		}

		for _, s := range sources {
			sequencePath := s

			name := fmt.Sprintf("sequence %s _meta.toml", filepath.Base(sequencePath))
			c.AddJob(name, func() (bool, error) {
				source := sequencePath + "/_meta.toml"

				if !c.Changed(source) {
					return false, nil
				}

				slug := path.Base(sequencePath)

				var sequence Sequence
				err = mtoml.ParseFile(c, source, &sequence)
				if err != nil {
					return true, err
				}
				sequence.Slug = slug

				if err := sequence.validate(); err != nil {
					return true, err
				}

				// Do a little post-processing on all the entries found in the
				// sequence.
				for _, entry := range sequence.Entries {
					entry.DescriptionHTML =
						string(mmarkdown.Render(c, []byte(entry.Description)))
					entry.Sequence = &sequence
				}

				sequences[slug] = &sequence
				sequencesChanged[slug] = true
				return true, nil
			})
		}
	}

	//
	// Stylesheets
	//

	{
		c.AddJob("stylesheets", func() (bool, error) {
			return compileStylesheets(c,
				c.SourceDir+"/content/stylesheets",
				versionedAssetsDir+"/app.css")
		})
	}

	//
	// Talks
	//

	var talksChanged bool
	var talksMu sync.Mutex

	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/talks", nil)
		if err != nil {
			return []error{err}
		}

		if conf.Drafts {
			drafts, err := mfile.ReadDirCached(c, c.SourceDir+"/content/talks-drafts", nil)
			if err != nil {
				return []error{err}
			}
			sources = append(sources, drafts...)
		}

		for _, s := range sources {
			source := s

			name := fmt.Sprintf("talk: %s", filepath.Base(source))
			c.AddJob(name, func() (bool, error) {
				return renderTalk(c, source, &talks, &talksChanged, &talksMu)
			})
		}
	}

	//
	// Twitter
	//

	{
		c.AddJob("twitter", func() (bool, error) {
			return c.AllowError(renderTwitter(c, db)), nil
		})
	}

	//
	//
	//
	// PHASE 2
	//
	//
	//

	if errors := c.Wait(); errors != nil {
		c.Log.Errorf("Cancelling next phase due to build errors")
		return errors
	}

	// Various sorts for anything that might need it.
	{
		sortArticles(articles)
		sortFragments(fragments)
		sortNewsletters(nanoglyphs)
		sortNewsletters(passages)
		sortPhotos(photos)
		sortTalks(talks)

		for _, sequence := range sequences {
			sortSequenceEntries(sequence.Entries)
		}
	}

	//
	// Articles
	//

	// Index
	{
		c.AddJob("articles index", func() (bool, error) {
			return renderArticlesIndex(c, articles,
				articlesChanged)
		})
	}

	// Feed (all)
	{
		c.AddJob("articles feed", func() (bool, error) {
			return renderArticlesFeed(c, articles, nil,
				articlesChanged)
		})
	}

	// Feed (Postgres)
	{
		c.AddJob("articles feed (postgres)", func() (bool, error) {
			return renderArticlesFeed(c, articles, tagPointer(tagPostgres),
				articlesChanged)
		})
	}

	//
	// Fragments
	//

	// Index
	{
		c.AddJob("fragments index", func() (bool, error) {
			return renderFragmentsIndex(c, fragments,
				fragmentsChanged)
		})
	}

	// Feed
	{
		c.AddJob("fragments feed", func() (bool, error) {
			return renderFragmentsFeed(c, fragments,
				fragmentsChanged)
		})
	}

	//
	// Home
	//

	{
		c.AddJob("home", func() (bool, error) {
			return renderHome(c, articles, fragments, photos,
				articlesChanged, fragmentsChanged, photosChanged)
		})
	}

	//
	// Nanoglyphs
	//

	// Index
	{
		c.AddJob("nanoglyphs index", func() (bool, error) {
			return renderNanoglyphsIndex(c, nanoglyphs,
				nanoglyphsChanged)
		})
	}

	// Feed
	{
		c.AddJob("nanoglyphs feed", func() (bool, error) {
			return renderNanoglyphsFeed(c, nanoglyphs,
				nanoglyphsChanged)
		})
	}

	//
	// Pages (render each view)
	//

	var pagesMu sync.RWMutex

	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/pages", nil)
		if err != nil {
			return []error{err}
		}

		if conf.Drafts {
			drafts, err := mfile.ReadDirCached(c, c.SourceDir+"/pages-drafts", nil)
			if err != nil {
				return []error{err}
			}
			sources = append(sources, drafts...)
		}

		for _, s := range sources {
			source := s

			name := fmt.Sprintf("page: %s", filepath.Base(source))
			c.AddJob(name, func() (bool, error) {
				return renderPage(c, source, pages, &pagesMu)
			})
		}
	}

	//
	// Passages
	//

	// Index
	{
		c.AddJob("passages index", func() (bool, error) {
			return renderPassagesIndex(c, passages,
				passagesChanged)
		})
	}

	// Feed
	{
		c.AddJob("passages feed", func() (bool, error) {
			return renderPassagesFeed(c, passages,
				passagesChanged)
		})
	}

	//
	// Photos (index / fetch + resize)
	//

	// Photo index
	{
		c.AddJob("photos index", func() (bool, error) {
			return renderPhotoIndex(c, photos,
				photosChanged)
		})
	}

	// Photo fetch + resize
	{
		for _, p := range photos {
			photo := p

			name := fmt.Sprintf("photo: %s", photo.Slug)
			c.AddJob(name, func() (bool, error) {
				return fetchAndResizePhoto(c, c.SourceDir+"/content/photographs", photo)
			})
		}
	}

	//
	// Sequences (index / fetch + resize)
	//

	// Sequence master feed
	{
		c.AddJob("sequences: feed", func() (bool, error) {
			var allSequenceEntries []*SequenceEntry
			for _, sequence := range sequences {
				allSequenceEntries = append(allSequenceEntries, sequence.Entries...)
			}
			sortSequenceEntries(allSequenceEntries)

			var anySequenceChanged bool
			for _, changed := range sequencesChanged {
				if changed {
					anySequenceChanged = true
					break
				}
			}

			return renderSequenceFeed(c, nil, allSequenceEntries, anySequenceChanged)
		})
	}

	// Each sequence
	{
		for _, s := range sequences {
			sequence := s

			{
				err := mfile.EnsureDir(c, c.TargetDir+"/sequences/"+sequence.Slug)
				if err != nil {
					return []error{err}
				}
			}

			{
				err := mfile.EnsureDir(c, c.SourceDir+"/content/photographs/sequences/"+sequence.Slug)
				if err != nil {
					return []error{err}
				}
			}

			// Sequence index
			{
				name := fmt.Sprintf("sequence %s: index", sequence.Slug)
				c.AddJob(name, func() (bool, error) {
					return renderSequence(c, sequence, sequence.Entries,
						sequencesChanged[sequence.Slug])
				})
			}

			// Sequence all page
			{
				name := fmt.Sprintf("sequence %s: all", sequence.Slug)
				c.AddJob(name, func() (bool, error) {
					return renderSequenceAll(c, sequence, sequence.Entries,
						sequencesChanged[sequence.Slug])
				})
			}

			// Sequences all page background image
			if sequence.BackgroundImageURL != "" {
				name := fmt.Sprintf("sequence %s background", sequence.Slug)
				photo := &Photo{
					OriginalImageURL: sequence.BackgroundImageURL,
					Slug:             "background",
				}

				c.AddJob(name, func() (bool, error) {
					return fetchAndResizePhoto(c,
						c.SourceDir+"/content/photographs/sequences/"+sequence.Slug, photo)
				})
			}

			// Sequence feed
			{
				name := fmt.Sprintf("sequence %s: feed", sequence.Slug)
				c.AddJob(name, func() (bool, error) {
					return renderSequenceFeed(c, sequence, sequence.Entries,
						sequencesChanged[sequence.Slug])
				})
			}

			for i, e := range sequence.Entries {
				entryIndex := i
				entry := e

				// Sequence page
				name := fmt.Sprintf("sequence %s: %s", sequence.Slug, entry.Slug)
				c.AddJob(name, func() (bool, error) {
					return renderSequenceEntry(c, sequence, entry, entryIndex,
						sequencesChanged[sequence.Slug])
				})

				// Sequence fetch + resize
				for _, p := range entry.Photos {
					photo := p

					name = fmt.Sprintf("sequence %s entry %s photo: %s",
						sequence.Slug, entry.Slug, photo.Slug)
					c.AddJob(name, func() (bool, error) {
						return fetchAndResizePhoto(c,
							c.SourceDir+"/content/photographs/sequences/"+sequence.Slug, photo)
					})
				}
			}
		}
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Types
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Article represents an article to be rendered.
type Article struct {
	// Attributions are any attributions for content that may be included in
	// the article (like an image in the header for example).
	Attributions string `toml:"attributions,omitempty"`

	// Content is the HTML content of the article. It isn't included as TOML
	// frontmatter, and is rather split out of an article's Markdown file,
	// rendered, and then added separately.
	Content string `toml:"-"`

	// Draft indicates that the article is not yet published.
	Draft bool `toml:"-"`

	// HNLink is an optional link to comments on Hacker News.
	HNLink string `toml:"hn_link,omitempty"`

	// Hook is a leading sentence or two to succinctly introduce the article.
	Hook string `toml:"hook"`

	// HookImageURL is the URL for a hook image for the article (to be shown on
	// the article index) if one was found.
	HookImageURL string `toml:"-"`

	// Image is an optional image that may be included with an article.
	Image string `toml:"image,omitempty"`

	// Location is the geographical location where this article was written.
	Location string `toml:"location,omitempty"`

	// PublishedAt is when the article was published.
	PublishedAt *time.Time `toml:"published_at"`

	// Slug is a unique identifier for the article that also helps determine
	// where it's addressable by URL.
	Slug string `toml:"-"`

	// Tags are the set of tags that the article is tagged with.
	Tags []Tag `toml:"tags,omitempty"`

	// Title is the article's title.
	Title string `toml:"title"`

	// TOC is the HTML rendered table of contents of the article. It isn't
	// included as TOML frontmatter, but rather calculated from the article's
	// content, rendered, and then added separately.
	TOC string `toml:"-"`
}

// publishingInfo produces a brief spiel about publication which is intended to
// go into the left sidebar when an article is shown.
func (a *Article) publishingInfo() string {
	return `<p><strong>Article</strong><br>` + a.Title + `</p>` +
		`<p><strong>Published</strong><br>` + a.PublishedAt.Format("January 2, 2006") + `</p> ` +
		`<p><strong>Location</strong><br>` + a.Location + `</p>` +
		scommon.TwitterInfo
}

// taggedWith returns true if the given tag is in this article's set of tags
// and false otherwise.
func (a *Article) taggedWith(tag Tag) bool {
	for _, t := range a.Tags {
		if t == tag {
			return true
		}
	}

	return false
}

func (a *Article) validate(source string) error {
	if a.Location == "" {
		return fmt.Errorf("No location for article: %v", source)
	}

	if a.Title == "" {
		return fmt.Errorf("No title for article: %v", source)
	}

	if a.PublishedAt == nil {
		return fmt.Errorf("No publish date for article: %v", source)
	}

	return nil
}

// Fragment represents a fragment (that is, a short "stream of consciousness"
// style article) to be rendered.
type Fragment struct {
	// Attributions are any attributions for content that may be included in
	// the article (like an image in the header for example).
	Attributions string `toml:"attributions,omitempty"`

	// Content is the HTML content of the fragment. It isn't included as TOML
	// frontmatter, and is rather split out of an fragment's Markdown file,
	// rendered, and then added separately.
	Content string `toml:"-"`

	// Draft indicates that the fragment is not yet published.
	Draft bool `toml:"-"`

	// HNLink is an optional link to comments on Hacker News.
	HNLink string `toml:"hn_link,omitempty"`

	// Hook is a leading sentence or two to succinctly introduce the fragment.
	Hook string `toml:"hook"`

	// Image is an optional image that may be included with a fragment.
	Image string `toml:"image,omitempty"`

	// Location is the geographical location where this article was written.
	Location string `toml:"location,omitempty"`

	// PublishedAt is when the fragment was published.
	PublishedAt *time.Time `toml:"published_at"`

	// Slug is a unique identifier for the fragment that also helps determine
	// where it's addressable by URL.
	Slug string `toml:"-"`

	// Title is the fragment's title.
	Title string `toml:"title"`
}

// PublishingInfo produces a brief spiel about publication which is intended to
// go into the left sidebar when a fragment is shown.
func (f *Fragment) publishingInfo() string {
	s := `<p><strong>Fragment</strong><br>` + f.Title + `</p>` +
		`<p><strong>Published</strong><br>` + f.PublishedAt.Format("January 2, 2006") + `</p> `

	if f.Location != "" {
		s += `<p><strong>Location</strong><br>` + f.Location + `</p>`
	}

	s += scommon.TwitterInfo
	return s
}

func (f *Fragment) validate(source string) error {
	if f.Title == "" {
		return fmt.Errorf("No title for fragment: %v", source)
	}

	if f.PublishedAt == nil {
		return fmt.Errorf("No publish date for fragment: %v", source)
	}

	return nil
}

// Page is the metadata for a static HTML page generated from an ACE file.
type Page struct {
	// Paths for external dependencies that the page included as it was being
	// rendered, and which should be watched so that we can re-render it when
	// one changes.
	//
	// Set the first time a page is rendered and updated every subsequent
	// render.
	dependencies []string
}

// Photo is a photograph.
type Photo struct {
	// Description is the description of the photograph.
	Description string `toml:"description"`

	// KeepInHomeRotation is a special override for photos I really like that
	// keeps them in the home page's random rotation. The rotation then
	// consists of either a recent photo or one of these explicitly selected
	// old ones.
	KeepInHomeRotation bool `toml:"keep_in_home_rotation"`

	// OriginalImageURL is the location where the original-sized version of the
	// photo can be downloaded from.
	OriginalImageURL string `toml:"original_image_url"`

	// OccurredAt is UTC time when the photo was published.
	OccurredAt *time.Time `toml:"occurred_at"`

	// Portrait is a hint to indicate that the photo is in portrait instead of
	// landscape. This helps the build pick a better stand-in image for lazy
	// loading so that there's less jumping around as photos that get loaded in
	// change size.
	Portrait bool `toml:"portrait"`

	// Slug is a unique identifier for the photo. Originally these were
	// generated from Flickr, but I've since just started reusing them for
	// filenames.
	Slug string `toml:"slug"`

	// Title is the title of the photograph.
	Title string `toml:"title"`
}

// PhotoWrapper is a data structure intended to represent the data structure at
// the top level of photograph data file `content/photographs/_meta.toml`.
type PhotoWrapper struct {
	// Photos is a collection of photos within the top-level wrapper.
	Photos []*Photo `toml:"photographs"`
}

// Sequence is a sequence -- a series of photos that represent some kind of
// journey.
type Sequence struct {
	// BackgroundImageURL is the URL of the image to use as the background on
	// the all page.
	BackgroundImageURL string `toml:"background_image_url"`

	// Description is the description of the photograph.
	Description string `toml:"description"`

	// Entries are the set of entries in the sequence. Each contains a slug,
	// description, and one or more photos.
	Entries []*SequenceEntry `toml:"entries"`

	// Slug is a unique identifier for the sequence. It's interpreted from the
	// sequence's path.
	Slug string `toml:"-"`

	// Title is the title of the sequence.
	Title string `toml:"title"`
}

func (s *Sequence) validate() error {
	entrySlugs := make(map[string]struct{})

	for i, entry := range s.Entries {
		if entry.Slug == "" {
			return fmt.Errorf("No slug set for sequence entry: index %v", i)
		}

		if len(entry.Photos) < 1 {
			return fmt.Errorf("Sequence entry needs at least one photo: %v", entry.Slug)
		}

		if _, ok := entrySlugs[entry.Slug]; ok {
			return fmt.Errorf("Duplicate sequence entry slug: %v", entry.Slug)
		}
		entrySlugs[entry.Slug] = struct{}{}

		photoSlugs := make(map[string]struct{})
		for _, photo := range entry.Photos {
			if !strings.HasPrefix(photo.Slug, entry.Slug) {
				return fmt.Errorf("Photo slug '%v' should share prefix with entry slug '%v'",
					photo.Slug, entry.Slug)
			}

			if _, ok := photoSlugs[photo.Slug]; ok {
				return fmt.Errorf("Duplicate photo slug: %v", photo.Slug)
			}
			photoSlugs[photo.Slug] = struct{}{}
		}
	}

	return nil
}

// SequenceEntry is a single entry in a sequence.
type SequenceEntry struct {
	// Description is the description of the entry.
	Description string `toml:"description"`

	// DescriptionHTML is the description rendered to HTML.
	DescriptionHTML string `toml:"-"`

	// OccurredAt is UTC time when the entry was published.
	OccurredAt *time.Time `toml:"occurred_at"`

	// Photos is a collection of photos within this particular entry. Many
	// sequence entries will only have a single photo, but there are alternate
	// layouts for when one contains a number of different ones.
	Photos []*Photo `toml:"photographs"`

	// Sequence links back to the photo's parent sequence. Only set if the
	// photo is part of a sequence.
	Sequence *Sequence `toml:"-"`

	// Slug is a unique identifier for the entry.
	Slug string `toml:"slug"`

	// Title is the title of the entry.
	Title string `toml:"title"`
}

// FirstPhoto returns the first photograph for a sequence entry. This is
// commonly needed in view templates where accessing slice elements via index
// is made awkward by Go (`index arr 0` rather than `arr[0]`).
func (e *SequenceEntry) FirstPhoto() *Photo {
	return e.Photos[0]
}

// Tag is a symbol assigned to an article to categorize it.
//
// This feature is not meanted to be overused. It's really just for tagging
// a few particular things so that we can generate content-specific feeds for
// certain aggregates (so far just Planet Postgres).
type Tag string

// articleYear holds a collection of articles grouped by year.
type articleYear struct {
	Year     int
	Articles []*Article
}

// fragmentYear holds a collection of fragments grouped by year.
type fragmentYear struct {
	Year      int
	Fragments []*Fragment
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

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private
//
//
//
//////////////////////////////////////////////////////////////////////////////

func compileJavascripts(c *modulir.Context, sourceDir, target string) (bool, error) {
	sources, err := mfile.ReadDirCached(c, sourceDir, nil)
	if err != nil {
		return false, err
	}

	sourcesChanged := c.ChangedAny(sources...)
	if !sourcesChanged {
		return false, nil
	}

	return true, sassets.CompileJavascripts(c, sourceDir, target)
}

func compileStylesheets(c *modulir.Context, sourceDir, target string) (bool, error) {
	sources, err := mfile.ReadDirCached(c, sourceDir, nil)
	if err != nil {
		return false, err
	}

	sourcesChanged := c.ChangedAny(sources...)
	if !sourcesChanged {
		return false, nil
	}

	return true, sassets.CompileStylesheets(c, sourceDir, target)
}

var cropDefault = &mimage.PhotoCropSettings{Portrait: "2:3", Landscape: "3:2"}

var defaultPhotoSizes = []mimage.PhotoSize{
	{Suffix: ".jpg", Width: 333, CropSettings: cropDefault},
	{Suffix: "@2x.jpg", Width: 667, CropSettings: cropDefault},
	{Suffix: "_large.jpg", Width: 1500, CropSettings: cropDefault},
	{Suffix: "_large@2x.jpg", Width: 3000, CropSettings: cropDefault},
}

func fetchAndResizePhoto(c *modulir.Context, targetDir string, photo *Photo) (bool, error) {
	u, err := url.Parse(photo.OriginalImageURL)
	if err != nil {
		return false, fmt.Errorf("bad URL for photo '%s'", photo.Slug)
	}

	return mimage.FetchAndResizeImage(c, u, targetDir, photo.Slug, scommon.TempDir,
		mimage.PhotoGravityCenter, defaultPhotoSizes)
}

// getAceOptions gets a good set of default options for Ace template rendering
// for the project.
func getAceOptions(dynamicReload bool) *ace.Options {
	options := &ace.Options{FuncMap: scommon.TemplateFuncMap}

	if dynamicReload {
		options.DynamicReload = true
	}

	return options
}

// Gets a map of local values for use while rendering a template and includes
// a few "special" values that are globally relevant to all templates.
func getLocals(title string, locals map[string]interface{}) map[string]interface{} {
	defaults := map[string]interface{}{
		"BodyClass":         "",
		"EnableGoatCounter": conf.EnableGoatCounter,
		"FavIcon":           "favicon-152.png",
		"GoogleAnalyticsID": conf.GoogleAnalyticsID,
		"LocalFonts":        conf.LocalFonts,
		"Release":           Release,
		"SorgEnv":           conf.SorgEnv,
		"Title":             title,
		"TitleSuffix":       scommon.TitleSuffix,
		"TwitterCard":       nil,
		"ViewportWidth":     viewportWidthDeviceWidth,
	}

	for k, v := range locals {
		defaults[k] = v
	}

	return defaults
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

func insertOrReplaceArticle(articles *[]*Article, article *Article) {
	for i, a := range *articles {
		if article.Slug == a.Slug {
			(*articles)[i] = article
			return
		}
	}

	*articles = append(*articles, article)
}

func insertOrReplaceFragment(fragments *[]*Fragment, fragment *Fragment) {
	for i, f := range *fragments {
		if fragment.Slug == f.Slug {
			(*fragments)[i] = fragment
			return
		}
	}

	*fragments = append(*fragments, fragment)
}

func insertOrReplaceNewsletter(issues *[]*snewsletter.Issue, issue *snewsletter.Issue) {
	for i, s := range *issues {
		if issue.Slug == s.Slug {
			(*issues)[i] = issue
			return
		}
	}

	*issues = append(*issues, issue)
}

func insertOrReplaceTalk(talks *[]*stalks.Talk, talk *stalks.Talk) {
	for i, t := range *talks {
		if talk.Slug == t.Slug {
			(*talks)[i] = talk
			return
		}
	}

	*talks = append(*talks, talk)
}

// Remove the "./pages" directory and extension, but keep the rest of the
// path.
//
// Looks something like "about", or "nested/about".
func pagePathKey(source string) string {
	pagePath := mfile.MustAbs(source)
	pagePath = strings.TrimPrefix(pagePath, mfile.MustAbs("./pages-drafts")+"/")
	pagePath = strings.TrimPrefix(pagePath, mfile.MustAbs("./pages")+"/")
	pagePath = strings.TrimSuffix(pagePath, path.Ext(pagePath))
	return pagePath
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

func renderArticle(c *modulir.Context, source string, articles *[]*Article, articlesChanged *bool, mu *sync.Mutex) (bool, error) {
	sourceChanged := c.Changed(source)
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/articles/show.ace",
		},
		universalSources...,
	)...)
	if !sourceChanged && !viewsChanged {
		return false, nil
	}

	var article Article
	data, err := mtoml.ParseFileFrontmatter(c, source, &article)
	if err != nil {
		return true, err
	}

	err = article.validate(source)
	if err != nil {
		return true, err
	}

	article.Draft = scommon.IsDraft(source)
	article.Slug = scommon.ExtractSlug(source)

	article.Content, err = mmarkdownext.Render(string(data), nil)
	if err != nil {
		return true, err
	}

	article.TOC, err = mtoc.RenderFromHTML(article.Content)
	if err != nil {
		return true, err
	}

	format, ok := pathAsImage(
		path.Join(c.SourceDir, "content", "images", article.Slug, "hook"),
	)
	if ok {
		article.HookImageURL = "/assets/images/" + article.Slug + "/hook." + format
	}

	card := &twitterCard{
		Title:       article.Title,
		Description: article.Hook,
	}
	format, ok = pathAsImage(
		path.Join(c.SourceDir, "content", "images", article.Slug, "twitter@2x"),
	)
	if ok {
		card.ImageURL = conf.AbsoluteURL + "/assets/images/" + article.Slug + "/twitter@2x." + format
	}

	locals := getLocals(article.Title, map[string]interface{}{
		"Article":        article,
		"PublishingInfo": article.publishingInfo(),
		"TwitterCard":    card,
	})

	// Always use force context because if we made it to here we know that our
	// sources have changed.
	err = mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/articles/show.ace",
		path.Join(c.TargetDir, article.Slug), getAceOptions(viewsChanged), locals)
	if err != nil {
		return true, err
	}

	mu.Lock()
	insertOrReplaceArticle(articles, &article)
	*articlesChanged = true
	mu.Unlock()

	return true, nil
}

func renderArticlesIndex(c *modulir.Context, articles []*Article, articlesChanged bool) (bool, error) {
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/articles/index.ace",
		},
		universalSources...,
	)...)
	if !articlesChanged && !viewsChanged {
		return false, nil
	}

	articlesByYear := groupArticlesByYear(articles)

	locals := getLocals("Articles", map[string]interface{}{
		"ArticlesByYear": articlesByYear,
	})

	return true, mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/articles/index.ace",
		c.TargetDir+"/articles/index.html", getAceOptions(viewsChanged), locals)
}

func renderArticlesFeed(c *modulir.Context, articles []*Article, tag *Tag, articlesChanged bool) (bool, error) {
	if !articlesChanged {
		return false, nil
	}

	name := "articles"
	if tag != nil {
		name = fmt.Sprintf("articles-%s", *tag)
	}
	filename := name + ".atom"

	title := "Articles" + scommon.TitleSuffix
	if tag != nil {
		title = fmt.Sprintf("Articles%s (%s)", scommon.TitleSuffix, *tag)
	}

	feed := &matom.Feed{
		Title: title,
		ID:    "tag:" + scommon.AtomTag + ",2013:/" + name,

		Links: []*matom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/" + filename},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	if len(articles) > 0 {
		feed.Updated = *articles[0].PublishedAt
	}

	for i, article := range articles {
		if tag != nil && !article.taggedWith(*tag) {
			continue
		}

		if i >= conf.NumAtomEntries {
			break
		}

		entry := &matom.Entry{
			Title:     article.Title,
			Summary:   article.Hook,
			Content:   &matom.EntryContent{Content: article.Content, Type: "html"},
			Published: *article.PublishedAt,
			Updated:   *article.PublishedAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/" + article.Slug},
			ID:        "tag:" + scommon.AtomTag + "," + article.PublishedAt.Format("2006-01-02") + ":" + article.Slug,

			AuthorName: scommon.AtomAuthorName,
			AuthorURI:  conf.AbsoluteURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	f, err := os.Create(path.Join(conf.TargetDir, filename))
	if err != nil {
		return true, err
	}
	defer f.Close()

	return true, feed.Encode(f, "  ")
}

func renderFragment(c *modulir.Context, source string, fragments *[]*Fragment, fragmentsChanged *bool, mu *sync.Mutex) (bool, error) {
	sourceChanged := c.Changed(source)
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/fragments/show.ace",
		},
		universalSources...,
	)...)
	if !sourceChanged && !viewsChanged {
		return false, nil
	}

	var fragment Fragment
	data, err := mtoml.ParseFileFrontmatter(c, source, &fragment)
	if err != nil {
		return true, err
	}

	err = fragment.validate(source)
	if err != nil {
		return true, err
	}

	fragment.Draft = scommon.IsDraft(source)
	fragment.Slug = scommon.ExtractSlug(source)

	fragment.Content, err = mmarkdownext.Render(string(data), nil)
	if err != nil {
		return true, err
	}

	card := &twitterCard{
		Title:       fragment.Title,
		Description: fragment.Hook,
	}
	format, ok := pathAsImage(
		path.Join(c.SourceDir, "content", "images", "fragments", fragment.Slug, "twitter@2x"),
	)
	if ok {
		card.ImageURL = conf.AbsoluteURL + "/assets/images/fragments/" + fragment.Slug + "/twitter@2x." + format
	}

	locals := getLocals(fragment.Title, map[string]interface{}{
		"Fragment":       fragment,
		"PublishingInfo": fragment.publishingInfo(),
		"TwitterCard":    card,
	})

	err = mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/fragments/show.ace",
		path.Join(c.TargetDir, "fragments", fragment.Slug),
		getAceOptions(viewsChanged), locals)
	if err != nil {
		return true, err
	}

	mu.Lock()
	insertOrReplaceFragment(fragments, &fragment)
	*fragmentsChanged = true
	mu.Unlock()

	return true, nil
}

func renderFragmentsFeed(c *modulir.Context, fragments []*Fragment,
	fragmentsChanged bool) (bool, error) {
	if !fragmentsChanged {
		return false, nil
	}

	feed := &matom.Feed{
		Title: "Fragments" + scommon.TitleSuffix,
		ID:    "tag:" + scommon.AtomTag + ",2013:/fragments",

		Links: []*matom.Link{
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

		entry := &matom.Entry{
			Title:     fragment.Title,
			Summary:   fragment.Hook,
			Content:   &matom.EntryContent{Content: fragment.Content, Type: "html"},
			Published: *fragment.PublishedAt,
			Updated:   *fragment.PublishedAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/fragments/" + fragment.Slug},
			ID:        "tag:" + scommon.AtomTag + "," + fragment.PublishedAt.Format("2006-01-02") + ":fragments/" + fragment.Slug,

			AuthorName: scommon.AtomAuthorName,
			AuthorURI:  conf.AbsoluteURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	f, err := os.Create(conf.TargetDir + "/fragments.atom")
	if err != nil {
		return true, err
	}
	defer f.Close()

	return true, feed.Encode(f, "  ")
}

func renderFragmentsIndex(c *modulir.Context, fragments []*Fragment,
	fragmentsChanged bool) (bool, error) {
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/fragments/index.ace",
		},
		universalSources...,
	)...)
	if !fragmentsChanged && !viewsChanged {
		return false, nil
	}

	fragmentsByYear := groupFragmentsByYear(fragments)

	locals := getLocals("Fragments", map[string]interface{}{
		"FragmentsByYear": fragmentsByYear,
	})

	return true, mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/fragments/index.ace",
		c.TargetDir+"/fragments/index.html", getAceOptions(viewsChanged), locals)
}

func renderNanoglyph(c *modulir.Context, source string, issues *[]*snewsletter.Issue, nanoglyphsChanged *bool, mu *sync.Mutex) (bool, error) {
	sourceChanged := c.Changed(source)
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.NanoglyphsLayout,
			scommon.ViewsDir + "/nanoglyphs/show.ace",
		},
		universalSources...,
	)...)
	if !sourceChanged && !viewsChanged {
		return false, nil
	}

	issue, err := snewsletter.Render(c, filepath.Dir(source), filepath.Base(source),
		"", false)
	if err != nil {
		return true, err
	}

	format, ok := pathAsImage(
		path.Join(c.SourceDir, "content", "images", "nanoglyphs", issue.Slug, "hook"),
	)
	if ok {
		issue.HookImageURL = "/assets/images/nanoglyphs/" + issue.Slug + "/hook." + format
	}

	locals := getLocals(issue.Title, map[string]interface{}{
		"BodyClass": "web-only", // For web-specific CSS rules
		"FavIcon":   "nanoglyph-152.png",
		"InEmail":   false,
		"Issue":     issue,
		"URLPrefix": "", // Relative prefix for the web version
	})

	err = mace.RenderFile(c, scommon.NanoglyphsLayout, scommon.ViewsDir+"/nanoglyphs/show.ace",
		c.TargetDir+"/nanoglyphs/"+issue.Slug, getAceOptions(viewsChanged), locals)
	if err != nil {
		return true, err
	}

	mu.Lock()
	insertOrReplaceNewsletter(issues, issue)
	*nanoglyphsChanged = true
	mu.Unlock()

	return true, nil
}

func renderNanoglyphsFeed(c *modulir.Context, issues []*snewsletter.Issue, nanoglyphsChanged bool) (bool, error) {
	if !nanoglyphsChanged {
		return false, nil
	}

	name := "nanoglyphs"
	filename := name + ".atom"
	title := "Nanoglyph" + scommon.TitleSuffix

	feed := &matom.Feed{
		Title: title,
		ID:    "tag:" + scommon.AtomTag + ",2013:/" + name,

		Links: []*matom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/" + filename},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	if len(articles) > 0 {
		feed.Updated = *issues[0].PublishedAt
	}

	for i, issue := range issues {
		if i >= conf.NumAtomEntries {
			break
		}

		content := issue.Content
		if issue.ImageURL != "" {
			content = fmt.Sprintf(`<p><img src="%s" alt="%s" /></p>`, issue.ImageURL, issue.ImageAlt) +
				content
		}

		entry := &matom.Entry{
			Title:     fmt.Sprintf("Nanoglyph %s — %s", issue.Number, issue.Title),
			Content:   &matom.EntryContent{Content: content, Type: "html"},
			Published: *issue.PublishedAt,
			Updated:   *issue.PublishedAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/nanoglyphs/" + issue.Slug},
			ID:        "tag:" + scommon.AtomTag + "," + issue.PublishedAt.Format("2006-01-02") + ":" + issue.Slug,

			AuthorName: scommon.AtomAuthorName,
			AuthorURI:  conf.AbsoluteURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	f, err := os.Create(path.Join(conf.TargetDir, filename))
	if err != nil {
		return true, err
	}
	defer f.Close()

	return true, feed.Encode(f, "  ")
}

func renderNanoglyphsIndex(c *modulir.Context, issues []*snewsletter.Issue,
	nanoglyphsChanged bool) (bool, error) {
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.NanoglyphsLayout,
			scommon.ViewsDir + "/nanoglyphs/index.ace",
		},
		universalSources...,
	)...)
	if !nanoglyphsChanged && !viewsChanged {
		return false, nil
	}

	locals := getLocals("Nanoglyph", map[string]interface{}{
		"BodyClass": "web-only", // For web-specific CSS rules
		"FavIcon":   "nanoglyph-152.png",
		"Issues":    issues,
		"URLPrefix": "", // Relative prefix for the web version
	})

	return true, mace.RenderFile(c, scommon.NanoglyphsLayout, scommon.ViewsDir+"/nanoglyphs/index.ace",
		c.TargetDir+"/nanoglyphs/index.html", getAceOptions(viewsChanged), locals)
}

func renderPassage(c *modulir.Context, source string, issues *[]*snewsletter.Issue, passagesChanged *bool, mu *sync.Mutex) (bool, error) {
	sourceChanged := c.Changed(source)
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.PassagesLayout,
			scommon.ViewsDir + "/passages/show.ace",
		},
		universalSources...,
	)...)
	if !sourceChanged && !viewsChanged {
		return false, nil
	}

	issue, err := snewsletter.Render(c, filepath.Dir(source), filepath.Base(source),
		"", false)
	if err != nil {
		return true, err
	}

	format, ok := pathAsImage(
		path.Join(c.SourceDir, "content", "images", "passages", issue.Slug, "hook"),
	)
	if ok {
		issue.HookImageURL = "/assets/images/passages/" + issue.Slug + "/hook." + format
	}

	locals := getLocals(issue.Title, map[string]interface{}{
		"BodyClass": "web-only", // For web-specific CSS rules
		"InEmail":   false,
		"Issue":     issue,
		"URLPrefix": "", // Relative prefix for the web version
	})

	err = mace.RenderFile(c, scommon.PassagesLayout, scommon.ViewsDir+"/passages/show.ace",
		c.TargetDir+"/passages/"+issue.Slug, getAceOptions(viewsChanged), locals)
	if err != nil {
		return true, err
	}

	mu.Lock()
	insertOrReplaceNewsletter(issues, issue)
	*passagesChanged = true
	mu.Unlock()

	return true, nil
}

func renderPassagesFeed(c *modulir.Context, issues []*snewsletter.Issue, passagesChanged bool) (bool, error) {
	if !passagesChanged {
		return false, nil
	}

	name := "passages"
	filename := name + ".atom"
	title := "Passages & Glass" + scommon.TitleSuffix

	feed := &matom.Feed{
		Title: title,
		ID:    "tag:" + scommon.AtomTag + ",2013:/" + name,

		Links: []*matom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/" + filename},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	if len(articles) > 0 {
		feed.Updated = *issues[0].PublishedAt
	}

	for i, issue := range issues {
		if i >= conf.NumAtomEntries {
			break
		}

		content := issue.Content
		if issue.ImageURL != "" {
			content = fmt.Sprintf(`<p><img src="%s" alt="%s" /></p>`, issue.ImageURL, issue.ImageAlt) +
				content
		}

		entry := &matom.Entry{
			Title:     fmt.Sprintf("Passages & Glass %s — %s", issue.Number, issue.Title),
			Content:   &matom.EntryContent{Content: content, Type: "html"},
			Published: *issue.PublishedAt,
			Updated:   *issue.PublishedAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/passages/" + issue.Slug},
			ID:        "tag:" + scommon.AtomTag + "," + issue.PublishedAt.Format("2006-01-02") + ":" + issue.Slug,

			AuthorName: scommon.AtomAuthorName,
			AuthorURI:  conf.AbsoluteURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	f, err := os.Create(path.Join(conf.TargetDir, filename))
	if err != nil {
		return true, err
	}
	defer f.Close()

	return true, feed.Encode(f, "  ")
}

func renderPassagesIndex(c *modulir.Context, issues []*snewsletter.Issue,
	passagesChanged bool) (bool, error) {
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.PassagesLayout,
			scommon.ViewsDir + "/passages/index.ace",
		},
		universalSources...,
	)...)
	if !passagesChanged && !viewsChanged {
		return false, nil
	}

	locals := getLocals("Passages", map[string]interface{}{
		"BodyClass": "web-only", // For web-specific CSS rules
		"Issues":    issues,
		"URLPrefix": "", // Relative prefix for the web version
	})

	return true, mace.RenderFile(c, scommon.PassagesLayout, scommon.ViewsDir+"/passages/index.ace",
		c.TargetDir+"/passages/index.html", getAceOptions(viewsChanged), locals)
}

func renderHome(c *modulir.Context,
	articles []*Article, fragments []*Fragment, photos []*Photo,
	articlesChanged, fragmentsChanged, photosChanged bool) (bool, error) {

	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/index.ace",
		},
		universalSources...,
	)...)
	if !articlesChanged && !fragmentsChanged && !photosChanged && !viewsChanged {
		return false, nil
	}

	if len(articles) > 3 {
		articles = articles[0:3]
	}

	// Try just one fragment for now to better balance the page's height.
	if len(fragments) > 1 {
		fragments = fragments[0:1]
	}

	// Find a random photo to put on the homepage.
	photo := selectRandomPhoto(photos)

	locals := getLocals("", map[string]interface{}{
		"Articles":  articles,
		"BodyClass": "index",
		"Fragments": fragments,
		"Photo":     photo,
	})

	return true, mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/index.ace",
		c.TargetDir+"/index.html", getAceOptions(viewsChanged), locals)
}

func renderPage(c *modulir.Context, source string, meta map[string]*Page, mu *sync.RWMutex) (bool, error) {
	pagePath := pagePathKey(source)

	// Other dependencies a page might have if it say, included an external
	// Markdown file. These are added the first time a page is rendered (and
	// watched), and updated on every subsequent run.
	var pageDependencies []string

	mu.RLock()
	pageMeta, ok := meta[pagePath]
	if ok {
		pageDependencies = pageMeta.dependencies
	}
	mu.RUnlock()

	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			source,
		},
		append(
			universalSources,
			pageDependencies...,
		)...,
	)...)
	if !viewsChanged {
		return false, nil
	}

	// Looks something like "./public/about".
	target := path.Join(c.TargetDir, pagePath)

	// Put a ".html" on if this page is an index. This will allow our local
	// server to serve it at a directory path, and our upload script is smart
	// enough to do the right thing with it as well.
	if path.Base(pagePath) == "index" {
		target += ".html"
	}

	pageDependenciesMap := map[string]struct{}{}
	ctx := context.WithValue(context.Background(),
		mtemplatemd.IncludeMarkdownDependencyKeys, pageDependenciesMap)

	// Reuse existing metadata for this page, or create metadata if this is the
	// first time we're rendering it.
	if pageMeta == nil {
		pageMeta = &Page{}

		mu.Lock()
		meta[pagePath] = pageMeta
		mu.Unlock()
	}

	// Pages get their titles by using inner templates. That must be triggered
	// by sending an empty string as `Title`.
	locals := getLocals("", map[string]interface{}{
		"Ctx": ctx,
	})

	err := mfile.EnsureDir(c, path.Dir(target))
	if err != nil {
		return true, err
	}

	err = mace.RenderFile(c, scommon.MainLayout, source, target,
		getAceOptions(viewsChanged), locals)
	if err != nil {
		return true, err
	}

	pageMeta.dependencies = nil
	for path := range pageDependenciesMap {
		// Check changed here so that Modulir will add the file to its watch
		// list.
		c.Changed(path)

		pageMeta.dependencies = append(pageMeta.dependencies, path)
	}

	return true, nil
}

func renderReading(c *modulir.Context, db *sql.DB) (bool, error) {
	if db == nil {
		return false, nil
	}

	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/reading/index.ace",
		},
		universalSources...,
	)...)
	if !c.FirstRun && !viewsChanged {
		return false, nil
	}

	return true, squantified.RenderReading(c, db, viewsChanged, getLocals)
}

func renderPhotoIndex(c *modulir.Context, photos []*Photo,
	photosChanged bool) (bool, error) {
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/photos/index.ace",
		},
		universalSources...,
	)...)
	if !photosChanged && !viewsChanged {
		return false, nil
	}

	locals := getLocals("Photos", map[string]interface{}{
		"BodyClass":     "photos",
		"Photos":        photos,
		"ViewportWidth": 600,
	})

	return true, mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/photos/index.ace",
		c.TargetDir+"/photos/index.html", getAceOptions(viewsChanged), locals)
}

func renderRobotsTxt(c *modulir.Context) (bool, error) {
	if !c.FirstRun && !c.Forced {
		return false, nil
	}

	var content string
	if conf.Drafts {
		// Allow Twitterbot so that we can preview card images on dev.
		//
		// Disallow everything else.
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

	outFile, err := os.Create(c.TargetDir + "/robots.txt")
	if err != nil {
		return true, err
	}
	outFile.WriteString(content)
	outFile.Close()

	return true, nil
}

func renderRuns(c *modulir.Context, db *sql.DB) (bool, error) {
	if db == nil {
		return false, nil
	}

	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/runs/index.ace",
		},
		universalSources...,
	)...)
	if !c.FirstRun && !viewsChanged {
		return false, nil
	}

	return true, squantified.RenderRuns(c, db, viewsChanged, getLocals)
}

func renderSequence(c *modulir.Context, sequence *Sequence, entries []*SequenceEntry,
	sequenceChanged bool) (bool, error) {

	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/sequences/index.ace",
		},
		universalSources...,
	)...)
	if !sequenceChanged && !viewsChanged {
		return false, nil
	}

	title := fmt.Sprintf("Sequence: %s", sequence.Title)
	description := string(mmarkdown.Render(c, []byte(sequence.Description)))

	// Most of the time we want entries with the most recent first, but we want
	// them with the oldest first on the index page.
	entriesReversed := make([]*SequenceEntry, len(entries))
	for i, photo := range entries {
		entriesReversed[len(entries)-i-1] = photo
	}

	locals := getLocals(title, map[string]interface{}{
		"Description": description,
		"Entries":     entriesReversed,
		"Sequence":    sequence,
	})

	return true, mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/sequences/index.ace",
		path.Join(c.TargetDir, "sequences", sequence.Slug, "index.html"),
		getAceOptions(viewsChanged), locals)
}

func renderSequenceAll(c *modulir.Context, sequence *Sequence, entries []*SequenceEntry,
	sequenceChanged bool) (bool, error) {

	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/sequences/all.ace",
		},
		universalSources...,
	)...)
	if !sequenceChanged && !viewsChanged {
		return false, nil
	}

	title := fmt.Sprintf("Sequence: %s", sequence.Title)
	description := string(mmarkdown.Render(c, []byte(sequence.Description)))

	// Oldest first
	entriesReversed := make([]*SequenceEntry, len(entries))
	for i, photo := range entries {
		entriesReversed[len(entries)-i-1] = photo
	}

	locals := getLocals(title, map[string]interface{}{
		"Description": description,
		"Entries":     entriesReversed,
		"Sequence":    sequence,
	})

	return true, mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/sequences/all.ace",
		path.Join(c.TargetDir, "sequences", sequence.Slug, "all"),
		getAceOptions(viewsChanged), locals)
}

// Renders at Atom feed for a sequence. The entries slice is assumed to be
// pre-sorted.
//
// The one non-obvious mechanism worth mentioning is that it can also render a
// general Atom feed for all sequence entries mixed together by specifying
// `sequence` as `nil` and entries as a master list of all sequence entries
// combined.
func renderSequenceFeed(c *modulir.Context, sequence *Sequence, entries []*SequenceEntry,
	sequenceChanged bool) (bool, error) {

	if !sequenceChanged {
		return false, nil
	}

	feedIDSuffix := "sequences"
	if sequence != nil {
		feedIDSuffix += ":" + sequence.Slug
	}

	filename := "sequences.atom"
	if sequence != nil {
		filename = path.Join("sequences", sequence.Slug+".atom")
	}

	title := "Sequences - brandur.org"
	if sequence != nil {
		title = fmt.Sprintf("Sequences (%s) - brandur.org", sequence.Slug)
	}

	feed := &matom.Feed{
		Title: title,
		ID:    "tag:" + scommon.AtomTag + ",2019:" + feedIDSuffix,

		Links: []*matom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/" + filename},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	if len(entries) > 0 {
		feed.Updated = *entries[0].OccurredAt
	}

	for i, entry := range entries {
		if i >= conf.NumAtomEntries {
			break
		}

		htmlContent := entry.DescriptionHTML +
			fmt.Sprintf(`<p><img src="/photographs/sequences/%s/%s_large@2x.jpg"></p>`,
				entry.Sequence.Slug, entry.Slug)

		entry := &matom.Entry{
			Title:     entry.Title,
			Content:   &matom.EntryContent{Content: htmlContent, Type: "html"},
			Published: *entry.OccurredAt,
			Updated:   *entry.OccurredAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/sequences/" + entry.Sequence.Slug + "/" + entry.Slug},
			ID:        "tag:" + scommon.AtomTag + "," + entry.OccurredAt.Format("2006-01-02") + ":sequences:" + entry.Sequence.Slug + ":" + entry.Slug,

			AuthorName: scommon.AtomAuthorName,
			AuthorURI:  conf.AbsoluteURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	f, err := os.Create(path.Join(conf.TargetDir, filename))
	if err != nil {
		return true, err
	}
	defer f.Close()

	return true, feed.Encode(f, "  ")
}

func renderSequenceEntry(c *modulir.Context, sequence *Sequence, entry *SequenceEntry, entryIndex int,
	sequenceChanged bool) (bool, error) {

	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/sequences/entry.ace",
		},
		universalSources...,
	)...)
	if !sequenceChanged && !viewsChanged {
		return false, nil
	}

	// A set of previous and next entries for the carousel.
	//
	// Note that the subtraction/addition operations may appear to be
	// "backwards" and that's because they are. This is because by the time the
	// code gets here, the entries list has already been sorted in _reverse_
	// chronological order.
	var entryPrev, entryPrevPrev *SequenceEntry
	var entryNext, entryNextNext *SequenceEntry
	if entryIndex+2 < len(sequence.Entries) {
		entryPrevPrev = sequence.Entries[entryIndex+2]
	}
	if entryIndex+1 < len(sequence.Entries) {
		entryPrev = sequence.Entries[entryIndex+1]
	}
	if entryIndex-1 >= 0 {
		entryNext = sequence.Entries[entryIndex-1]
	}
	if entryIndex-2 >= 0 {
		entryNextNext = sequence.Entries[entryIndex-2]
	}

	title := fmt.Sprintf("%s — %s %s", entry.Title, sequence.Title, entry.Slug)

	locals := getLocals(title, map[string]interface{}{
		"Entry":         entry,
		"EntryNext":     entryNext,
		"EntryNextNext": entryNextNext,
		"EntryPrev":     entryPrev,
		"EntryPrevPrev": entryPrevPrev,
		"Sequence":      sequence,
		"ViewportWidth": viewportWidthDeviceWidth,
	})

	return true, mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/sequences/entry.ace",
		path.Join(c.TargetDir, "sequences", sequence.Slug, entry.Slug),
		getAceOptions(viewsChanged), locals)
}

func renderTalk(c *modulir.Context, source string, talks *[]*stalks.Talk, talksChanged *bool, mu *sync.Mutex) (bool, error) {
	sourceChanged := c.Changed(source)
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/talks/show.ace",
		},
		universalSources...,
	)...)
	if !sourceChanged && !viewsChanged {
		return false, nil
	}

	// TODO: modulir-ize this package
	talk, err := stalks.Render(
		c, c.SourceDir+"/content", filepath.Dir(source), filepath.Base(source))
	if err != nil {
		return true, err
	}

	locals := getLocals(talk.Title, map[string]interface{}{
		"BodyClass":      "talk",
		"PublishingInfo": talk.PublishingInfo(),
		"Talk":           talk,
	})

	err = mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/talks/show.ace",
		path.Join(c.TargetDir, talk.Slug), getAceOptions(viewsChanged), locals)
	if err != nil {
		return true, err
	}

	mu.Lock()
	insertOrReplaceTalk(talks, talk)
	*talksChanged = true
	mu.Unlock()

	return true, nil
}

func renderTwitter(c *modulir.Context, db *sql.DB) (bool, error) {
	if db == nil {
		return false, nil
	}

	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/twitter/index.ace",
		},
		universalSources...,
	)...)
	if !c.FirstRun && !viewsChanged {
		return false, nil
	}

	return true, squantified.RenderTwitter(c, db, viewsChanged, getLocals)
}

func selectRandomPhoto(photos []*Photo) *Photo {
	if len(photos) < 1 {
		return nil
	}

	numRecent := 20
	if len(photos) < numRecent {
		numRecent = len(photos)
	}

	// All recent photos go into the random selection.
	randomPhotos := photos[0:numRecent]

	// Older photos that are good enough that I've explicitly tagged them
	// as such also get considered for the rotation.
	if len(photos) > numRecent {
		olderPhotos := photos[numRecent : len(photos)-1]

		for _, photo := range olderPhotos {
			if photo.KeepInHomeRotation {
				randomPhotos = append(randomPhotos, photo)
			}
		}
	}

	return randomPhotos[rand.Intn(len(randomPhotos))]
}

func sortArticles(articles []*Article) {
	sort.Slice(articles, func(i, j int) bool {
		return articles[j].PublishedAt.Before(*articles[i].PublishedAt)
	})
}

func sortFragments(fragments []*Fragment) {
	sort.Slice(fragments, func(i, j int) bool {
		return fragments[j].PublishedAt.Before(*fragments[i].PublishedAt)
	})
}

func sortNewsletters(issues []*snewsletter.Issue) {
	sort.Slice(issues, func(i, j int) bool {
		return issues[j].PublishedAt.Before(*issues[i].PublishedAt)
	})
}

func sortPhotos(photos []*Photo) {
	sort.Slice(photos, func(i, j int) bool {
		return photos[j].OccurredAt.Before(*photos[i].OccurredAt)
	})
}

func sortSequenceEntries(entries []*SequenceEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[j].OccurredAt.Before(*entries[i].OccurredAt)
	})
}

func sortTalks(talks []*stalks.Talk) {
	sort.Slice(talks, func(i, j int) bool {
		return talks[j].PublishedAt.Before(*talks[i].PublishedAt)
	})
}

// Gets a pointer to a tag just to work around the fact that you can take the
// address of a constant like `tagPostgres`.
func tagPointer(tag Tag) *Tag {
	return &tag
}
