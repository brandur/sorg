package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
	"github.com/yosssi/ace"
	"golang.org/x/exp/slices"
	"golang.org/x/xerrors"

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/mace"
	"github.com/brandur/modulir/modules/matom"
	"github.com/brandur/modulir/modules/mfile"
	"github.com/brandur/modulir/modules/mimage"
	"github.com/brandur/modulir/modules/mmarkdown"
	"github.com/brandur/modulir/modules/mmarkdownext"
	"github.com/brandur/modulir/modules/mtemplate"
	"github.com/brandur/modulir/modules/mtemplatemd"
	"github.com/brandur/modulir/modules/mtoc"
	"github.com/brandur/modulir/modules/mtoml"
	"github.com/brandur/sorg/modules/sassets"
	"github.com/brandur/sorg/modules/scommon"
	"github.com/brandur/sorg/modules/snewsletter"
	"github.com/brandur/sorg/modules/squantified"
	"github.com/brandur/sorg/modules/stemplate"
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
	articles     []*Article
	dependencies = &DependencyRegistry{sources: make(map[string][]string)}
	fragments    []*Fragment
	nanoglyphs   []*snewsletter.Issue
	passages     []*snewsletter.Issue
	pages        = make(map[string]*Page)
	photos       []*Photo
	photosOther  []*Photo
	sequences    []*SequenceEntry
	tweets       []*squantified.Tweet
)

// Time zone to show articles / fragments / etc. publishing times in.
var localLocation = mustLocation("America/Los_Angeles")

// List of common build dependencies, a change in any of which will trigger a
// rebuild on everything: partial views, JavaScripts, and stylesheets. Even
// though some of those changes will false positives, these sources are
// pervasive enough, and changes infrequent enough, that it's worth the
// tradeoff. This variable is a global because so many render functions access
// it.
var universalSources []string

var validate = validator.New()

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
	mmarkdownext.FuncMap = scommon.TextTemplateFuncMap
	stemplate.LocalLocation = localLocation
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

//nolint:maintidx
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

	// A set of source paths that rebuild everything when any one of them
	// changes. These are dependencies that are included in more or less
	// everything: common partial views, JavaScript sources, and stylesheet
	// sources.
	universalSources = nil

	// Generate a set of JavaScript sources to add to universal sources.
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

	ctx := context.Background()

	ctx, downloadedImageContainer := mtemplate.DownloadedImageContext(ctx)

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
			{c.SourceDir + "/content/javascripts-modular", versionedAssetsDir + "/javascripts"},

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
	// Pages (render each view)
	//

	var pagesMu sync.RWMutex

	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/pages", &mfile.ReadDirOptions{RecurseDirs: true})
		if err != nil {
			return []error{err}
		}

		if conf.Drafts {
			drafts, err := mfile.ReadDirCached(c, c.SourceDir+"/pages-drafts", &mfile.ReadDirOptions{RecurseDirs: true})
			if err != nil {
				return []error{err}
			}
			sources = append(sources, drafts...)
		}

		for _, s := range sources {
			source := s

			name := fmt.Sprintf("page: %s", filepath.Base(source))
			c.AddJob(name, func() (bool, error) {
				return renderPage(ctx, c, source, pages, &pagesMu)
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
	// Photos (other) (read `_other_meta.toml`)
	//

	{
		c.AddJob("photos (other) _meta.toml", func() (bool, error) {
			source := c.SourceDir + "/content/photographs/_other_meta.toml"

			if !c.Changed(source) {
				return false, nil
			}

			var photosWrapper PhotoWrapper
			err := mtoml.ParseFile(c, source, &photosWrapper)
			if err != nil {
				return true, err
			}

			if err := photosWrapper.validate(); err != nil {
				return true, err
			}

			photosOther = photosWrapper.Photos
			return true, nil
		})
	}

	//
	// Reading
	//

	{
		c.AddJob("reading", func() (bool, error) {
			return renderReading(c)
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
			return renderRuns(c)
		})
	}

	//
	// Sequences (read `_meta.toml`)
	//

	var sequenceChanged bool

	{
		c.AddJob("sequences", func() (bool, error) {
			source := c.SourceDir + "/content/sequences/_meta.toml"

			if !c.Changed(source) {
				return false, nil
			}

			sequenceChanged = true

			var sequenceWrapper SequenceWrapper
			err := mtoml.ParseFile(c, source, &sequenceWrapper)
			if err != nil {
				return true, err
			}

			if err := sequenceWrapper.validate(); err != nil {
				return true, err
			}

			// Do a little post-processing on all the entries found in the
			// sequence.
			for _, entry := range sequenceWrapper.Entries {
				entry.DescriptionHTML = template.HTML(string(mmarkdown.Render(c, []byte(entry.Description))))
			}

			sequences = sequenceWrapper.Entries

			return true, nil
		})
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
	// Twitter (read `data/twitter.toml`)
	//

	var tweetsChanged bool

	{
		c.AddJob("twitter data/twitter.toml", func() (bool, error) {
			source := scommon.DataDir + "/twitter.toml"

			if !c.Changed(source) {
				return false, nil
			}

			var err error
			tweets, err = squantified.ReadTwitterData(c, source)
			if err != nil {
				return true, err
			}

			tweetsChanged = true
			return true, nil
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
		slices.SortFunc(articles, func(a, b *Article) bool { return a.PublishedAt.Before(b.PublishedAt) })
		slices.SortFunc(fragments, func(a, b *Fragment) bool { return a.PublishedAt.Before(b.PublishedAt) })
		slices.SortFunc(nanoglyphs, func(a, b *snewsletter.Issue) bool { return a.PublishedAt.Before(b.PublishedAt) })
		slices.SortFunc(passages, func(a, b *snewsletter.Issue) bool { return a.PublishedAt.Before(b.PublishedAt) })
		slices.SortFunc(photos, func(a, b *Photo) bool { return b.OccurredAt.Before(a.OccurredAt) })
		slices.SortFunc(sequences, func(a, b *SequenceEntry) bool { return b.OccurredAt.Before(a.OccurredAt) })
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

	// Photo fetch + resize (other)
	{
		for _, p := range photosOther {
			photo := p

			name := fmt.Sprintf("photo fetch: %s", photo.Slug)
			c.AddJob(name, func() (bool, error) {
				return fetchAndResizePhotoOther(c, c.SourceDir+"/content/photographs", photo)
			})
		}
	}

	// From `DownloadedImage` template tags.
	{
		for i := range downloadedImageContainer.Images {
			imageInfo := downloadedImageContainer.Images[i]

			c.AddJob("downloaded image: "+imageInfo.Slug, func() (bool, error) {
				return fetchAndResizeDownloadedImage(c, c.SourceDir+"/content/photographs", imageInfo)
			})
		}
	}

	//
	// Sequences (index / fetch + resize)
	//

	// Sequences index
	{
		c.AddJob("sequence: index", func() (bool, error) {
			return renderSequenceIndex(ctx, c, sequences, sequenceChanged)
		})
	}

	// Sequences feed
	{
		c.AddJob("sequence: feed", func() (bool, error) {
			return renderSequenceFeed(ctx, c, sequences, sequenceChanged)
		})
	}

	// Each entry
	{
		for _, e := range sequences {
			entry := e

			// Sequence page
			name := fmt.Sprintf("sequence: %s", entry.Slug)
			c.AddJob(name, func() (bool, error) {
				return renderSequenceEntry(ctx, c, entry, sequenceChanged)
			})

			// Sequence fetch + resize
			for _, p := range entry.Photos {
				photo := p

				name = fmt.Sprintf("sequence entry %s photo: %s", entry.Slug, photo.Slug)
				c.AddJob(name, func() (bool, error) {
					return fetchAndResizePhoto(c,
						c.SourceDir+"/content/photographs/sequences", photo)
				})
			}
		}
	}

	//
	// Twitter indexes
	//

	{
		c.AddJob("twitter (no replies)", func() (bool, error) {
			return renderTwitter(c, tweets, tweetsChanged, false)
		})

		c.AddJob("twitter (with replies)", func() (bool, error) {
			return renderTwitter(c, tweets, tweetsChanged, true)
		})
	}

	// Twitter photo fetch + resize
	{
		for _, t := range tweets {
			tweet := t

			if tweet.Entities == nil {
				continue
			}

			for _, m := range tweet.Entities.Medias {
				media := m

				if media.Type != "photo" {
					continue
				}

				name := fmt.Sprintf("twitter photo: %v", media.ID)
				c.AddJob(name, func() (bool, error) {
					return fetchAndResizePhotoTwitter(c, c.SourceDir+"/content/photographs/twitter",
						tweet, media)
				})
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
	Location string `toml:"location,omitempty" validate:"required"`

	// PublishedAt is when the article was published.
	PublishedAt time.Time `toml:"published_at" validate:"required"`

	// Slug is a unique identifier for the article that also helps determine
	// where it's addressable by URL.
	Slug string `toml:"-"`

	// Tags are the set of tags that the article is tagged with.
	Tags []Tag `toml:"tags,omitempty"`

	// Title is the article's title.
	Title string `toml:"title" validate:"required"`

	// TOC is the HTML rendered table of contents of the article. It isn't
	// included as TOML frontmatter, but rather calculated from the article's
	// content, rendered, and then added separately.
	TOC string `toml:"-"`
}

// publishingInfo produces a brief spiel about publication which is intended to
// go into the left sidebar when an article is shown.
func (a *Article) publishingInfo() map[string]string {
	info := make(map[string]string)

	info["Article"] = a.Title
	info["Published"] = a.PublishedAt.In(localLocation).Format("January 2, 2006")
	info["Location"] = a.Location

	return info
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
	if err := validate.Struct(a); err != nil {
		return xerrors.Errorf("error validating article %q: %+v", source, err)
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
	PublishedAt time.Time `toml:"published_at" validate:"required"`

	// Slug is a unique identifier for the fragment that also helps determine
	// where it's addressable by URL.
	Slug string `toml:"-"`

	// Title is the fragment's title.
	Title string `toml:"title" validate:"required"`
}

// PublishingInfo produces a brief spiel about publication which is intended to
// go into the left sidebar when a fragment is shown.
func (f *Fragment) publishingInfo() map[string]string {
	info := make(map[string]string)

	info["Fragment"] = f.Title
	info["Published"] = f.PublishedAt.In(localLocation).Format("January 2, 2006")

	if f.Location != "" {
		info["Location"] = f.Location
	}

	return info
}

func (f *Fragment) validate(source string) error {
	if err := validate.Struct(f); err != nil {
		return xerrors.Errorf("error validating fragment %q: %+v", source, err)
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
	// CropGravity is the gravity to use with ImageMagick when doing a square
	// crop. Should be one of: northwest, north, northeast, west, center, east,
	// southwest, south, southeast. Defaults to north gravity (because most
	// shots are portraits with my head near the top).
	CropGravity string `toml:"crop_gravity" default:"north"`

	// CropWidth is the width to crop the photo to.
	//
	// This should be the non-retina target width. A second file will be
	// created with the `@2x` suffix with twice this number.
	//
	// This is a required property for photos that are not part of the main
	// photographs sequence. It's ignored for photos that *are* part of the
	// main photographs sequence.
	CropWidth int `toml:"crop_width"`

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
	OccurredAt time.Time `toml:"occurred_at"`

	// Portrait is a hint to indicate that the photo is in portrait instead of
	// landscape. This helps the build pick a better stand-in image for lazy
	// loading so that there's less jumping around as photos that get loaded in
	// change size.
	Portrait bool `toml:"portrait"`

	// Slug is a unique identifier for the photo. Originally these were
	// generated from Flickr, but I've since just started reusing them for
	// filenames.
	Slug string `toml:"slug" validate:"required"`

	// Title is the title of the photograph.
	Title string `toml:"title"`
}

// PhotoWrapper is a data structure intended to represent the data structure at
// the top level of photograph data file `content/photographs/_meta.toml`.
type PhotoWrapper struct {
	// Photos is a collection of photos within the top-level wrapper.
	Photos []*Photo `toml:"photographs" validate:"required,dive"`
}

func (w *PhotoWrapper) validate() error {
	if err := validate.Struct(w); err != nil {
		return xerrors.Errorf("error validating photos: %+v", err)
	}
	return nil
}

// SequenceWrapper is a sequence -- a series of photos that represent some kind of
// journey.
type SequenceWrapper struct {
	// Entries are the set of entries in the sequence. Each contains a slug,
	// description, and one or more photos.
	Entries []*SequenceEntry `toml:"entries" validate:"required,dive"`
}

func (w *SequenceWrapper) validate() error {
	if err := validate.Struct(w); err != nil {
		return xerrors.Errorf("error validating sequences: %+v", err)
	}

	entrySlugs := make(map[string]struct{})

	for i, entry := range w.Entries {
		if entry.Slug == "" {
			return xerrors.Errorf("no slug set for sequence entry: index %v", i)
		}

		if _, ok := entrySlugs[entry.Slug]; ok {
			return xerrors.Errorf("duplicate sequence entry slug: %v", entry.Slug)
		}
		entrySlugs[entry.Slug] = struct{}{}

		photoSlugs := make(map[string]struct{})
		for _, photo := range entry.Photos {
			if !strings.HasPrefix(photo.Slug, entry.Slug) {
				return xerrors.Errorf("photo slug '%v' should share prefix with entry slug '%v'",
					photo.Slug, entry.Slug)
			}

			if _, ok := photoSlugs[photo.Slug]; ok {
				return xerrors.Errorf("duplicate photo slug: %v", photo.Slug)
			}
			photoSlugs[photo.Slug] = struct{}{}
		}
	}

	return nil
}

// SequenceEntry is a single entry in a sequence.
type SequenceEntry struct {
	// Description is the description of the entry.
	Description string `toml:"description" validate:"required"`

	// DescriptionHTML is the description rendered to HTML.
	DescriptionHTML template.HTML `toml:"-" validate:"-"`

	// OccurredAt is UTC time when the entry was published.
	OccurredAt time.Time `toml:"occurred_at" validate:"required"`

	// Photos is a collection of photos within this particular entry. Many
	// sequence entries will only have a single photo, but there are alternate
	// layouts for when one contains a number of different ones.
	Photos []*Photo `toml:"photographs" validate:"required,dive"`

	// Slug is a unique identifier for the entry.
	Slug string `toml:"slug" validate:"required"`

	// Title is the title of the entry.
	Title string `toml:"title" validate:"required"`
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
	{Suffix: "", Width: 333, CropSettings: cropDefault},
	{Suffix: "@2x", Width: 667, CropSettings: cropDefault},
	{Suffix: "_large", Width: 1500, CropSettings: cropDefault},
	{Suffix: "_large@2x", Width: 3000, CropSettings: cropDefault},
}

func fetchAndResizePhoto(c *modulir.Context, targetDir string, photo *Photo) (bool, error) {
	u, err := url.Parse(photo.OriginalImageURL)
	if err != nil {
		return false, xerrors.Errorf("bad URL for photo '%s': %w", photo.Slug, err)
	}

	return mimage.FetchAndResizeImage(c, u, targetDir, photo.Slug,
		mimage.PhotoGravityCenter, defaultPhotoSizes)
}

func fetchAndResizeDownloadedImage(c *modulir.Context,
	targetDir string, imageInfo *mtemplate.DownloadedImageInfo,
) (bool, error) {
	base := filepath.Base(imageInfo.Slug)
	dir := targetDir + filepath.Dir(imageInfo.Slug)

	return mimage.FetchAndResizeImage(c, imageInfo.URL, dir, base, mimage.PhotoGravityCenter,
		[]mimage.PhotoSize{
			{Suffix: "", Width: imageInfo.Width, CropSettings: cropDefault},
			{Suffix: "@2x", Width: imageInfo.Width * 2, CropSettings: cropDefault},
		})
}

func fetchAndResizePhotoOther(c *modulir.Context, targetDir string, photo *Photo) (bool, error) {
	if photo.CropWidth == 0 {
		return false, xerrors.Errorf("need `crop_width` specified for photo '%s'", photo.Slug)
	}

	u, err := url.Parse(photo.OriginalImageURL)
	if err != nil {
		return false, xerrors.Errorf("bad URL for photo '%s'", photo.Slug)
	}

	return mimage.FetchAndResizeImage(c, u, targetDir, photo.Slug,
		mimage.PhotoGravity(photo.CropGravity),
		[]mimage.PhotoSize{
			{Suffix: "", Width: photo.CropWidth, CropSettings: nil},
			{Suffix: "@2x", Width: photo.CropWidth * 2, CropSettings: nil},
		})
}

var twitterPhotoSizes = []mimage.PhotoSize{
	{Suffix: "", Width: 550},
	{Suffix: "@2x", Width: 1100},
}

func fetchAndResizePhotoTwitter(c *modulir.Context, targetDir string,
	tweet *squantified.Tweet, media *squantified.TweetEntitiesMedia,
) (bool, error) {
	u, err := url.Parse(media.URL)
	if err != nil {
		return false, xerrors.Errorf("bad URL for Twitter photo '%v': %w", media.ID, err)
	}

	slug := fmt.Sprintf("%v-%v", tweet.ID, media.ID)

	return mimage.FetchAndResizeImage(c, u, targetDir, slug,
		mimage.PhotoGravityCenter, twitterPhotoSizes)
}

// getAceOptions gets a good set of default options for Ace template rendering
// for the project.
func getAceOptions(dynamicReload bool) *ace.Options {
	options := &ace.Options{FuncMap: scommon.HTMLTemplateFuncMap}

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
		"GoogleAnalyticsID": conf.GoogleAnalyticsID,
		"LocalFonts":        conf.LocalFonts,
		"Release":           Release,
		"SorgEnv":           conf.SorgEnv,
		"Title":             title,
		"TitleSuffix":       scommon.TitleSuffix,
		"TwitterCard":       nil,
		"TwitterInfo":       scommon.TwitterInfo,
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

func mustLocation(locationName string) *time.Location {
	locatio, err := time.LoadLocation(locationName)
	if err != nil {
		panic(err)
	}
	return locatio
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
	pagePath = strings.TrimSuffix(pagePath, path.Ext(pagePath)) // again, for `.tmpl.html`
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

func renderArticle(c *modulir.Context, source string,
	articles *[]*Article, articlesChanged *bool, mu *sync.Mutex,
) (bool, error) {
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

	if article.Hook != "" {
		hook, err := mmarkdownext.Render(article.Hook, nil)
		if err != nil {
			return true, err
		}

		article.Hook = mtemplate.CollapseParagraphs(hook)
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

func renderArticlesFeed(_ *modulir.Context, articles []*Article, tag *Tag, articlesChanged bool) (bool, error) {
	if !articlesChanged {
		return false, nil
	}

	name := "articles"
	if tag != nil {
		name = fmt.Sprintf("articles-%s", *tag)
	}
	atomPath := name + ".atom"

	title := "Articles" + scommon.TitleSuffix
	if tag != nil {
		title = fmt.Sprintf("Articles%s (%s)", scommon.TitleSuffix, *tag)
	}

	feed := &matom.Feed{
		Title: title,
		ID:    "tag:" + scommon.AtomTag + ",2013:/" + name,

		Links: []*matom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/" + atomPath},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	if len(articles) > 0 {
		feed.Updated = articles[0].PublishedAt
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
			Published: article.PublishedAt,
			Updated:   article.PublishedAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/" + article.Slug},
			ID:        "tag:" + scommon.AtomTag + "," + article.PublishedAt.Format("2006-01-02") + ":" + article.Slug,

			AuthorName: scommon.AtomAuthorName,
			AuthorURI:  conf.AbsoluteURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	filename := path.Join(conf.TargetDir, atomPath)
	f, err := os.Create(filename)
	if err != nil {
		return true, xerrors.Errorf("error creating file '%s': %w", filename, err)
	}
	defer f.Close()

	return true, feed.Encode(f, "  ")
}

func renderFragment(c *modulir.Context, source string,
	fragments *[]*Fragment, fragmentsChanged *bool, mu *sync.Mutex,
) (bool, error) {
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

	if fragment.Hook != "" {
		hook, err := mmarkdownext.Render(fragment.Hook, nil)
		if err != nil {
			return true, err
		}

		fragment.Hook = mtemplate.CollapseParagraphs(hook)
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

func renderFragmentsFeed(_ *modulir.Context, fragments []*Fragment,
	fragmentsChanged bool,
) (bool, error) {
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
		feed.Updated = fragments[0].PublishedAt
	}

	for i, fragment := range fragments {
		if i >= conf.NumAtomEntries {
			break
		}

		entry := &matom.Entry{
			Title:     fragment.Title,
			Summary:   fragment.Hook,
			Content:   &matom.EntryContent{Content: fragment.Content, Type: "html"},
			Published: fragment.PublishedAt,
			Updated:   fragment.PublishedAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/fragments/" + fragment.Slug},
			ID: "tag:" + scommon.AtomTag + "," + fragment.PublishedAt.Format("2006-01-02") +
				":fragments/" + fragment.Slug,

			AuthorName: scommon.AtomAuthorName,
			AuthorURI:  conf.AbsoluteURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	filename := conf.TargetDir + "/fragments.atom"
	f, err := os.Create(filename)
	if err != nil {
		return true, xerrors.Errorf("error creating file '%s': %w", filename, err)
	}
	defer f.Close()

	return true, feed.Encode(f, "  ")
}

func renderFragmentsIndex(c *modulir.Context, fragments []*Fragment,
	fragmentsChanged bool,
) (bool, error) {
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

func renderNanoglyph(c *modulir.Context, source string,
	issues *[]*snewsletter.Issue, nanoglyphsChanged *bool, mu *sync.Mutex,
) (bool, error) {
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

func renderNanoglyphsFeed(_ *modulir.Context, issues []*snewsletter.Issue, nanoglyphsChanged bool) (bool, error) {
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
		feed.Updated = issues[0].PublishedAt
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
			Published: issue.PublishedAt,
			Updated:   issue.PublishedAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/nanoglyphs/" + issue.Slug},
			ID:        "tag:" + scommon.AtomTag + "," + issue.PublishedAt.Format("2006-01-02") + ":" + issue.Slug,

			AuthorName: scommon.AtomAuthorName,
			AuthorURI:  conf.AbsoluteURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	filePath := path.Join(conf.TargetDir, filename)
	f, err := os.Create(filePath)
	if err != nil {
		return true, xerrors.Errorf("error creating file '%s': %w", filePath, err)
	}
	defer f.Close()

	return true, feed.Encode(f, "  ")
}

func renderNanoglyphsIndex(c *modulir.Context, issues []*snewsletter.Issue,
	nanoglyphsChanged bool,
) (bool, error) {
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
		"Issues":    issues,
		"URLPrefix": "", // Relative prefix for the web version
	})

	return true, mace.RenderFile(c, scommon.NanoglyphsLayout, scommon.ViewsDir+"/nanoglyphs/index.ace",
		c.TargetDir+"/nanoglyphs/index.html", getAceOptions(viewsChanged), locals)
}

func renderPassage(c *modulir.Context, source string,
	issues *[]*snewsletter.Issue, passagesChanged *bool, mu *sync.Mutex,
) (bool, error) {
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

func renderPassagesFeed(_ *modulir.Context, issues []*snewsletter.Issue, passagesChanged bool) (bool, error) {
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
		feed.Updated = issues[0].PublishedAt
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
			Published: issue.PublishedAt,
			Updated:   issue.PublishedAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/passages/" + issue.Slug},
			ID:        "tag:" + scommon.AtomTag + "," + issue.PublishedAt.Format("2006-01-02") + ":" + issue.Slug,

			AuthorName: scommon.AtomAuthorName,
			AuthorURI:  conf.AbsoluteURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	filePath := path.Join(conf.TargetDir, filename)
	f, err := os.Create(filePath)
	if err != nil {
		return true, xerrors.Errorf("error creating file '%s': %w", filePath, err)
	}
	defer f.Close()

	return true, feed.Encode(f, "  ")
}

func renderPassagesIndex(c *modulir.Context, issues []*snewsletter.Issue,
	passagesChanged bool,
) (bool, error) {
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
	articlesChanged, fragmentsChanged, photosChanged bool,
) (bool, error) {
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

func renderPage(ctx context.Context, c *modulir.Context,
	source string, meta map[string]*Page, mu *sync.RWMutex,
) (bool, error) {
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

	ctx, includeMarkdownContainer := mtemplatemd.Context(ctx)

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

	pageMeta.dependencies = nil

	if strings.HasSuffix(source, ".ace") {
		err := mace.RenderFile(c, scommon.MainLayout, source, target,
			getAceOptions(viewsChanged), locals)
		if err != nil {
			return true, err
		}

		pageMeta.dependencies = append(pageMeta.dependencies, scommon.MainLayout)
	} else {
		dependencies, err := renderGoTemplate(source, target, locals)
		if err != nil {
			return true, err
		}

		pageMeta.dependencies = append(pageMeta.dependencies, dependencies...)
	}

	for path := range includeMarkdownContainer.Dependencies {
		pageMeta.dependencies = append(pageMeta.dependencies, path)
	}

	for _, path := range pageMeta.dependencies {
		// Check changed here so that Modulir will add the file to its watch
		// list.
		c.Changed(path)
	}

	return true, nil
}

func renderReading(c *modulir.Context) (bool, error) {
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.DataDir + "/goodreads.toml",
			scommon.MainLayout,
			scommon.ViewsDir + "/reading/index.ace",
		},
		universalSources...,
	)...)
	if !c.FirstRun && !viewsChanged {
		return false, nil
	}

	return true, squantified.RenderReading(c, viewsChanged, getLocals)
}

func renderPhotoIndex(c *modulir.Context, photos []*Photo,
	photosChanged bool,
) (bool, error) {
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

	filePath := c.TargetDir + "/robots.txt"
	outFile, err := os.Create(filePath)
	if err != nil {
		return true, xerrors.Errorf("error creating file '%s': %w", filePath, err)
	}
	if _, err := outFile.WriteString(content); err != nil {
		return true, xerrors.Errorf("error writing file '%s': %w", filePath, err)
	}
	outFile.Close()

	return true, nil
}

func renderRuns(c *modulir.Context) (bool, error) {
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.DataDir + "/strava.toml",
			scommon.MainLayout,
			scommon.ViewsDir + "/runs/index.ace",
		},
		universalSources...,
	)...)
	if !c.FirstRun && !viewsChanged {
		return false, nil
	}

	return true, squantified.RenderRuns(c, viewsChanged, getLocals)
}

// Renders at Atom feed for a sequence. The entries slice is assumed to be
// pre-sorted.
//
// The one non-obvious mechanism worth mentioning is that it can also render a
// general Atom feed for all sequence entries mixed together by specifying
// `sequence` as `nil` and entries as a master list of all sequence entries
// combined.
func renderSequenceFeed(ctx context.Context, c *modulir.Context,
	entries []*SequenceEntry, sequencesChanged bool,
) (bool, error) {
	source := scommon.ViewsDir + "/sequences/_entry_atom.tmpl.html"

	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !sequencesChanged && !viewsChanged {
		return false, nil
	}

	feed := &matom.Feed{
		Title: "Sequences - brandur.org",
		ID:    "tag:" + scommon.AtomTag + ",2019:sequences",

		Links: []*matom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/sequences.atom"},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	if len(entries) > 0 {
		feed.Updated = entries[0].OccurredAt
	}

	for i, entry := range entries {
		if i >= conf.NumAtomEntries {
			break
		}

		locals := getLocals("", map[string]interface{}{
			"Entry": entry,
		})

		var contentBuf bytes.Buffer
		err := dependencies.renderGoTemplateWriter(ctx, source, &contentBuf, locals)
		if err != nil {
			return true, err
		}

		entry := &matom.Entry{
			Title:     entry.Slug + " — " + entry.Title,
			Content:   &matom.EntryContent{Content: contentBuf.String(), Type: "html"},
			Published: entry.OccurredAt,
			Updated:   entry.OccurredAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/sequences/" + entry.Slug},
			ID: "tag:" + scommon.AtomTag + "," + entry.OccurredAt.Format("2006-01-02") +
				":sequences:" + entry.Slug,

			AuthorName: scommon.AtomAuthorName,
			AuthorURI:  conf.AbsoluteURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	filePath := path.Join(conf.TargetDir, "sequences.atom")
	f, err := os.Create(filePath)
	if err != nil {
		return true, xerrors.Errorf("error creating file '%s': %w", filePath, err)
	}
	defer f.Close()

	return true, feed.Encode(f, "  ")
}

func renderSequenceEntry(ctx context.Context, c *modulir.Context, entry *SequenceEntry, sequencesChanged bool,
) (bool, error) {
	source := scommon.ViewsDir + "/sequences/entry.tmpl.html"

	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !sequencesChanged && !viewsChanged {
		return false, nil
	}

	title := fmt.Sprintf("%s — %s", entry.Title, entry.Slug)

	locals := getLocals(title, map[string]interface{}{
		"Entry": entry,
	})

	err := dependencies.renderGoTemplate(ctx, source, path.Join(c.TargetDir, "sequences", entry.Slug), locals)
	if err != nil {
		return true, err
	}

	return true, nil
}

func renderSequenceIndex(ctx context.Context, c *modulir.Context, entries []*SequenceEntry,
	sequenceChanged bool,
) (bool, error) {
	source := scommon.ViewsDir + "/sequences/index.tmpl.html"

	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !sequenceChanged && !viewsChanged {
		return false, nil
	}

	locals := getLocals("Sequences", map[string]interface{}{
		"Entries": entries,
	})

	err := dependencies.renderGoTemplate(ctx, source, path.Join(c.TargetDir, "sequences/index.html"), locals)
	if err != nil {
		return true, err
	}

	return true, nil
}

func renderTwitter(c *modulir.Context, tweets []*squantified.Tweet, tweetsChanged, withReplies bool) (bool, error) {
	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			scommon.ViewsDir + "/twitter/index.ace",
		},
		universalSources...,
	)...)
	if !c.FirstRun && !viewsChanged && !tweetsChanged {
		return false, nil
	}

	return true, squantified.RenderTwitter(c, viewsChanged, getLocals, tweets, withReplies)
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

	//nolint:gosec
	return randomPhotos[rand.Intn(len(randomPhotos))]
}

// Gets a pointer to a tag just to work around the fact that you can take the
// address of a constant like `tagPostgres`.
func tagPointer(tag Tag) *Tag {
	return &tag
}

//
// TODO: Extract types/functions below this line to something better, probably
// in Modulir.
//

// DependencyRegistry maps Go template sources to other Go template sources that
// have been included in them as dependencies. It's used to know when to trigger
// a rebuild on a file change.
type DependencyRegistry struct {
	// Maps sources to their dependencies.
	sources   map[string][]string
	sourcesMu sync.RWMutex
}

func (r *DependencyRegistry) getDependencies(source string) []string {
	r.sourcesMu.RLock()
	defer r.sourcesMu.RUnlock()

	return r.sources[source]
}

func (r *DependencyRegistry) renderGoTemplate(ctx context.Context,
	source, target string, locals map[string]interface{},
) error {
	ctx, includeMarkdownContainer := mtemplatemd.Context(ctx)

	locals["Ctx"] = ctx

	dependencies, err := renderGoTemplate(source, target, locals)
	if err != nil {
		return err
	}

	for path := range includeMarkdownContainer.Dependencies {
		dependencies = append(dependencies, path)
	}

	r.sourcesMu.Lock()
	r.sources[source] = dependencies
	r.sourcesMu.Unlock()

	return nil
}

func (r *DependencyRegistry) renderGoTemplateWriter(ctx context.Context,
	source string, writer io.Writer, locals map[string]interface{},
) error {
	ctx, includeMarkdownContainer := mtemplatemd.Context(ctx)

	locals["Ctx"] = ctx

	dependencies, err := renderGoTemplateWriter(source, writer, locals)
	if err != nil {
		return err
	}

	for path := range includeMarkdownContainer.Dependencies {
		dependencies = append(dependencies, path)
	}

	r.sourcesMu.Lock()
	r.sources[source] = dependencies
	r.sourcesMu.Unlock()

	return nil
}

var goFileTemplateRE = regexp.MustCompile(`\{\{\-? ?template "([^"]+\.tmpl.html)"`)

func findGoSubTemplates(templateData string) []string {
	subTemplateMatches := goFileTemplateRE.FindAllStringSubmatch(templateData, -1)

	subTemplateNames := make([]string, len(subTemplateMatches))
	for i, match := range subTemplateMatches {
		subTemplateNames[i] = match[1]
	}

	return subTemplateNames
}

func parseGoTemplate(baseTmpl *template.Template, path string) (*template.Template, []string, error) {
	templateData, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, xerrors.Errorf("error reading template file %q: %w", path, err)
	}

	dependencies := []string{path}

	for _, subTemplatePath := range findGoSubTemplates(string(templateData)) {
		newBaseTmpl, subDependencies, err := parseGoTemplate(baseTmpl, subTemplatePath)
		if err != nil {
			return nil, nil, err
		}

		dependencies = append(dependencies, subDependencies...)
		baseTmpl = newBaseTmpl
	}

	newBaseTmpl, err := baseTmpl.New(path).Funcs(scommon.HTMLTemplateFuncMap).Parse(string(templateData))
	if err != nil {
		return nil, nil, xerrors.Errorf("error reading parsing template %q: %w", path, err)
	}

	return newBaseTmpl, dependencies, nil
}

func renderGoTemplate(path, target string, locals map[string]interface{}) ([]string, error) {
	tmpl, dependencies, err := parseGoTemplate(template.New("base_empty"), path)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(target)
	if err != nil {
		return nil, xerrors.Errorf("error creating target file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	if err := tmpl.Execute(writer, locals); err != nil {
		return nil, xerrors.Errorf("error executing template: %w", err)
	}

	return dependencies, nil
}

func renderGoTemplateWriter(path string, writer io.Writer, locals map[string]interface{}) ([]string, error) {
	tmpl, dependencies, err := parseGoTemplate(template.New("base_empty"), path)
	if err != nil {
		return nil, err
	}

	if err := tmpl.Execute(writer, locals); err != nil {
		return nil, xerrors.Errorf("error executing template: %w", err)
	}

	return dependencies, nil
}
