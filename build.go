package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
	"github.com/yosssi/ace"
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
	atoms        []*Atom
	dependencies = NewDependencyRegistry()
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
			c.TargetDir + "/atoms",
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
			{c.SourceDir + "/content/videos", c.TargetDir + "/videos"},
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

			name := "article: " + filepath.Base(source)
			c.AddJob(name, func() (bool, error) {
				return renderArticle(ctx, c, source,
					&articles, &articlesChanged, &articlesMu)
			})
		}
	}

	//
	// Atoms (read `_meta.toml`)
	//

	var atomsChanged bool

	{
		c.AddJob("atoms", func() (bool, error) {
			// Constrain descriptions to 2217 bytes as specified by Spring '83
			// even though they're also posted off Spring '83.
			const maxBytesLength = 2217

			source := c.SourceDir + "/content/atoms/_meta.toml"

			if !c.Changed(source) {
				return false, nil
			}

			atomsChanged = true

			var atomsWrapper AtomWrapper
			err := mtoml.ParseFile(c, source, &atomsWrapper)
			if err != nil {
				return true, err
			}

			if err := atomsWrapper.validate(); err != nil {
				return true, err
			}

			var replaceEverything bool
			if len(atoms) != len(atomsWrapper.Atoms) {
				replaceEverything = true
			}

			slices.SortFunc(atomsWrapper.Atoms, func(a, b *Atom) int { return b.PublishedAt.Compare(a.PublishedAt) })

			// Do a little post-processing on each atom, but try to skip any
			// that haven't changed.
			for i, atom := range atomsWrapper.Atoms {
				if !replaceEverything {
					lastAtom := atoms[i]
					if lastAtom.Equal(atom) {
						// Although the raw atoms are equal, we still use the
						// current version because it'll have values for any
						// rendered properties like DescriptionHTML.
						lastAtom.changed = false
						atomsWrapper.Atoms[i] = lastAtom
						continue
					}
				}

				atom.DescriptionHTML = template.HTML(string(mmarkdown.Render(c, []byte(atom.Description))))
				atom.Slug = atomSlug(atom.PublishedAt)

				atom.changed = true

				if len([]byte(atom.DescriptionHTML)) > maxBytesLength && !atom.LengthExempted {
					return true, xerrors.Errorf("atom's length is greater than %d bytes (was %d): %q",
						maxBytesLength, len([]byte(atom.DescriptionHTML)), atom.Description[0:100])
				}
			}

			atoms = atomsWrapper.Atoms

			return true, nil
		})
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

			name := "fragment: " + filepath.Base(source)
			c.AddJob(name, func() (bool, error) {
				return renderFragment(ctx, c, source,
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

			name := "nanoglyph: " + filepath.Base(source)
			c.AddJob(name, func() (bool, error) {
				return renderNanoglyph(ctx, c, source,
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

			name := "page: " + filepath.Base(source)
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

			name := "passage: " + filepath.Base(source)
			c.AddJob(name, func() (bool, error) {
				return renderPassage(ctx, c, source,
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
			return renderReading(ctx, c)
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
			return renderRuns(ctx, c)
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

			var replaceEverything bool
			if len(sequences) != len(sequenceWrapper.Entries) {
				replaceEverything = true
			}

			slices.SortFunc(sequenceWrapper.Entries, func(a, b *SequenceEntry) int { return b.PublishedAt.Compare(a.PublishedAt) })

			// Do a little post-processing on all the entries found in the
			// sequence, but try to skip any that haven't changed.
			for i, entry := range sequenceWrapper.Entries {
				if !replaceEverything {
					lastEntry := sequences[i]
					if lastEntry.Equal(entry) {
						// Although the raw atoms are equal, we still use the
						// current version because it'll have values for any
						// rendered properties like DescriptionHTML.
						lastEntry.changed = false
						sequenceWrapper.Entries[i] = lastEntry
						continue
					}
				}

				entry.DescriptionHTML = template.HTML(string(mmarkdown.Render(c, []byte(entry.Description))))

				entry.changed = true
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
	//
	// Some slices are sorted above when they're read in so that they can be
	// compared against a current version.
	{
		slices.SortFunc(articles, func(a, b *Article) int { return b.PublishedAt.Compare(a.PublishedAt) })
		slices.SortFunc(fragments, func(a, b *Fragment) int { return b.PublishedAt.Compare(a.PublishedAt) })
		slices.SortFunc(nanoglyphs, func(a, b *snewsletter.Issue) int { return b.PublishedAt.Compare(a.PublishedAt) })
		slices.SortFunc(passages, func(a, b *snewsletter.Issue) int { return b.PublishedAt.Compare(a.PublishedAt) })
		slices.SortFunc(photos, func(a, b *Photo) int { return b.OccurredAt.Compare(a.OccurredAt) })
	}

	//
	// Articles
	//

	// Index
	{
		c.AddJob("articles index", func() (bool, error) {
			return renderArticlesIndex(ctx, c, articles,
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
	// Atoms (index / fetch + resize)
	//

	// Atoms archive
	{
		c.AddJob("atoms: archive", func() (bool, error) {
			return renderAtomArchive(ctx, c, atoms, atomsChanged)
		})
	}

	// Atoms index
	{
		c.AddJob("atoms: index", func() (bool, error) {
			return renderAtomIndex(ctx, c, atoms, atomsChanged)
		})
	}

	// Atoms feed
	{
		c.AddJob("atoms: feed", func() (bool, error) {
			return renderAtomFeed(ctx, c, atoms, atomsChanged)
		})
	}

	// Each atom
	{
		for i, a := range atoms {
			atom := a

			if !atom.changed {
				continue
			}

			// Atom page
			name := "atom: " + atom.Slug
			c.AddJob(name, func() (bool, error) {
				return renderAtom(ctx, c, atom, i, atomsChanged)
			})

			// Photo fetch + resize
			for i := range atom.Photos {
				photo := atom.Photos[i]

				name = fmt.Sprintf("atom %q photo: %s", atom.Slug, photo.Slug)
				c.AddJob(name, func() (bool, error) {
					return fetchAndResizePhoto(c,
						c.SourceDir+"/content/photographs/atoms/"+atom.Slug, photo)
				})
			}

			// Video fetch
			for _, video := range atom.Videos {
				for _, u := range video.URL {
					videoURL := u
					name = fmt.Sprintf("atom %q video: %s", atom.Slug, filepath.Base(videoURL))
					c.AddJob(name, func() (bool, error) {
						return fetchVideo(ctx, c,
							c.SourceDir+"/content/videos/atoms/"+atom.Slug, videoURL)
					})
				}
			}
		}
	}

	//
	// Fragments
	//

	// Index
	{
		c.AddJob("fragments index", func() (bool, error) {
			return renderFragmentsIndex(ctx, c, fragments,
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
			return renderHome(ctx, c, articles, fragments, nanoglyphs, photos, sequences,
				articlesChanged, fragmentsChanged, nanoglyphsChanged, photosChanged, sequenceChanged)
		})
	}

	//
	// Nanoglyphs
	//

	// Index
	{
		c.AddJob("nanoglyphs index", func() (bool, error) {
			return renderNanoglyphsIndex(ctx, c, nanoglyphs,
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
			return renderPassagesIndex(ctx, c, passages,
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

			name := "photo: " + photo.Slug
			c.AddJob(name, func() (bool, error) {
				return fetchAndResizePhoto(c, c.SourceDir+"/content/photographs", photo)
			})
		}
	}

	// Photo fetch + resize (other)
	{
		for _, p := range photosOther {
			photo := p

			name := "photo fetch: " + photo.Slug
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
		c.AddJob("sequences: index", func() (bool, error) {
			return renderSequencesIndex(ctx, c, sequences, sequenceChanged)
		})
	}

	// Sequences feed
	{
		c.AddJob("sequences: feed", func() (bool, error) {
			return renderSequenceFeed(ctx, c, sequences, sequenceChanged)
		})
	}

	// Each sequences entry
	{
		for _, e := range sequences {
			entry := e

			if !entry.changed {
				continue
			}

			// Sequence page
			name := "sequences: " + entry.Slug
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
			return renderTwitter(ctx, c, tweets, tweetsChanged, false)
		})

		c.AddJob("twitter (with replies)", func() (bool, error) {
			return renderTwitter(ctx, c, tweets, tweetsChanged, true)
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
	Attributions template.HTML `toml:"attributions,omitempty"`

	// Content is the HTML content of the article. It isn't included as TOML
	// frontmatter, and is rather split out of an article's Markdown file,
	// rendered, and then added separately.
	Content template.HTML `toml:"-"`

	// Draft indicates that the article is not yet published.
	Draft bool `toml:"-"`

	// Footnotes are HTML footnotes extracted from content.
	Footnotes template.HTML `toml:"-"`

	// HNLink is an optional link to comments on Hacker News.
	HNLink string `toml:"hn_link,omitempty"`

	// Hook is a leading sentence or two to succinctly introduce the article.
	Hook template.HTML `toml:"hook"`

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
	TOC template.HTML `toml:"-"`
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

type AtomWrapper struct {
	Atoms []*Atom `toml:"atoms" validate:"required,dive"`
}

func (w *AtomWrapper) validate() error {
	if err := validate.Struct(w); err != nil {
		return xerrors.Errorf("error validating sequences: %+v", err)
	}
	return nil
}

// Atom is a single atom entry.
type Atom struct {
	// Description is the description of the entry.
	Description string `toml:"description" validate:"required"`

	// DescriptionHTML is the description rendered to HTML.
	DescriptionHTML template.HTML `toml:"-" validate:"-"`

	// LengthExempted indicates that the atom is allowed to be longer than the
	// standard 2217 rendered character limits.
	LengthExempted bool `toml:"length_exempted" validate:"-"`

	// Photos are any photos associated with the atom.
	Photos []*Photo `toml:"photos" validate:"omitempty,dive"`

	// PublishedAt is UTC time when the atom was published. It also serves to
	// provide a stable permalink.
	PublishedAt time.Time `toml:"published_at" validate:"required"`

	// Slug is a stable URL slug for the atom which is derived from its
	// timestamp.
	Slug string `toml:"-" validate:"-"`

	// Title is a title for the atom, but is optional. Atoms don't need and
	// mostly don't have titles.
	Title *string `toml:"title" validate:"-"`

	// Videos are any videos associated with the atom.
	Videos []*AtomVideo `toml:"videos" validate:"omitempty,dive"`

	// Tracks whether the atom has changed since the last build run so that
	// atoms can be rendered incrementally.
	changed bool `toml:"changed" validate:"-"`
}

func (a *Atom) Equal(other *Atom) bool {
	return a.Description == other.Description &&
		slices.EqualFunc(a.Photos, other.Photos, func(a, b *Photo) bool { return a.Equal(b) }) &&
		a.PublishedAt.Equal(other.PublishedAt) &&
		a.Title == other.Title &&
		slices.EqualFunc(a.Videos, other.Videos, func(a, b *AtomVideo) bool { return a.Equal(b) })
}

type AtomVideo struct {
	URL []string `toml:"url" validate:"required"`
}

func (v *AtomVideo) Equal(other *AtomVideo) bool {
	if len(v.URL) != len(other.URL) {
		return false
	}

	for i := range v.URL {
		if v.URL[i] != other.URL[i] {
			return false
		}
	}

	return true
}

// Fragment represents a fragment (that is, a short "stream of consciousness"
// style article) to be rendered.
type Fragment struct {
	// Attributions are any attributions for content that may be included in
	// the article (like an image in the header for example).
	Attributions template.HTML `toml:"attributions,omitempty"`

	// Content is the HTML content of the fragment. It isn't included as TOML
	// frontmatter, and is rather split out of an fragment's Markdown file,
	// rendered, and then added separately.
	Content template.HTML `toml:"-"`

	// Draft indicates that the fragment is not yet published.
	Draft bool `toml:"-"`

	// Footnotes are HTML footnotes extracted from content.
	Footnotes template.HTML `toml:"-"`

	// HNLink is an optional link to comments on Hacker News.
	HNLink string `toml:"hn_link,omitempty"`

	// Hook is a leading sentence or two to succinctly introduce the fragment.
	Hook template.HTML `toml:"hook"`

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

	info["Fragment"] = html.EscapeString(f.Title)
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
	// southwest, south, southeast.
	CropGravity string `default:"center" toml:"crop_gravity"`

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

	// LinkURL is a URL to have the image link to. This is only respect for some
	// uses of photographs like in atoms.
	LinkURL string `toml:"link_url" validate:"-"`

	// NoCrop disables cropping on this photo (normally photos are cropped to
	// 3:2 or 2:3).
	NoCrop bool `toml:"no_crop"`

	// OriginalImageURL is the location where the original-sized version of the
	// photo can be downloaded from.
	OriginalImageURL string `toml:"original_image_url" validate:"required"`

	// OccurredAt is UTC time when the photo was published.
	OccurredAt time.Time `toml:"occurred_at"`

	// OverrideExt is an extension like `.webp` that should be used for the
	// resized versions of the photo. Mostly useful for when a screenshot or
	// something is saved as a `.png` and it should really have been a `.jpg` or
	// something because the source being displayed was already lossy.
	OverrideExt string `toml:"override_ext" validate:"-"`

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

	// Internal
	originalExt string `toml:"-"`
}

func (p *Photo) Equal(other *Photo) bool {
	return p.CropGravity == other.CropGravity &&
		p.CropWidth == other.CropWidth &&
		p.Description == other.Description &&
		p.KeepInHomeRotation == other.KeepInHomeRotation &&
		p.LinkURL == other.LinkURL &&
		p.NoCrop == other.NoCrop &&
		p.OriginalImageURL == other.OriginalImageURL &&
		p.OccurredAt.Equal(other.OccurredAt) &&
		p.OverrideExt == other.OverrideExt &&
		p.Portrait == other.Portrait &&
		p.Slug == other.Slug &&
		p.Title == other.Title
}

func (p *Photo) OriginalExt() string {
	if p.originalExt != "" {
		return p.originalExt
	}

	p.originalExt = extCanonical(p.OriginalImageURL)
	return p.originalExt
}

func (p *Photo) TargetExt() string {
	if p.OverrideExt != "" {
		return p.OverrideExt
	}

	return extImageTarget(p.OriginalExt())
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

type SequenceWrapper struct {
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

	// Photos is a collection of photos within this particular entry. Many
	// sequence entries will only have a single photo, but there are alternate
	// layouts for when one contains a number of different ones.
	Photos []*Photo `toml:"photographs" validate:"required,dive"`

	// PublishedAt is UTC time when the entry was published.
	PublishedAt time.Time `toml:"published_at" validate:"required"`

	// Slug is a unique identifier for the entry.
	Slug string `toml:"slug" validate:"required"`

	// Title is the title of the entry.
	Title string `toml:"title" validate:"required"`

	// Tracks whether the entry has changed since the last build run so that
	// entries can be rendered incrementally.
	changed bool `toml:"changed" validate:"-"`
}

func (e *SequenceEntry) Equal(other *SequenceEntry) bool {
	return e.Description == other.Description &&
		slices.EqualFunc(e.Photos, other.Photos, func(a, b *Photo) bool { return a.Equal(b) }) &&
		e.PublishedAt.Equal(other.PublishedAt) &&
		e.Slug == other.Slug &&
		e.Title == other.Title
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

// readingYear holds a collection of readings grouped by year.
type readingYear struct {
	Year     int
	Readings []*squantified.Reading
}

// twitterCard represents a Twitter "card" (i.e. one of those rich media boxes
// that sometimes appear under tweets official clients) for use in templates.
type twitterCard struct {
	// Description is the description to show in the card.
	Description string

	// ImageURL is the URL to the image to show in the card. It should be
	// absolute because Twitter will need to be able to fetch it from our
	// servers. Leave blank if there is no image.
	ImageURL string

	// Title is the title to show in the card.
	Title string
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

// Very similar to RFC 4648 base32 except that numbers come first instead of
// last so that sortable values encoded to base32 will sort in the same
// lexicographic (alphabetical) order as the original values. Also, use lower
// case characters instead of upper.
var lexicographicBase32 = "234567abcdefghijklmnopqrstuvwxyz"

var lexicographicBase32Encoding = base32.NewEncoding(lexicographicBase32).
	WithPadding(base32.NoPadding)

// Produces an atom slug from its timestamp, which is the timestamp's unix time
// encoded via base32.
func atomSlug(publishedAt time.Time) string {
	i := big.NewInt(publishedAt.Unix())
	return lexicographicBase32Encoding.EncodeToString(i.Bytes())
}

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

func extCanonical(originalURL string) string {
	u, err := url.Parse(originalURL)
	if err != nil {
		panic(err)
	}

	return strings.ToLower(filepath.Ext(u.Path))
}

// Returns a target extension and format given an input one. Currently only used
// to make HEICs (which aren't web friendly) into more widely supported WebPs,
// but I should experiment with more broad use of WebPs. Other formats like JPGs
// and PNGs get left with their input extension/format.
func extImageTarget(canonicalExt string) string {
	if canonicalExt == ".heic" {
		return ".webp"
	}

	return canonicalExt
}

var cropDefault = &mimage.PhotoCropSettings{Portrait: "2:3", Landscape: "3:2"}

var defaultPhotoSizes = []mimage.PhotoSize{
	{Suffix: "", Width: 333, CropSettings: cropDefault},
	{Suffix: "@2x", Width: 667, CropSettings: cropDefault},
	{Suffix: "_large", Width: 1500, CropSettings: cropDefault},
	{Suffix: "_large@2x", Width: 3000, CropSettings: cropDefault},
}

var defaultPhotoSizesNoCrop = []mimage.PhotoSize{
	{Suffix: "", Width: 333, CropSettings: nil},
	{Suffix: "@2x", Width: 667, CropSettings: nil},
	{Suffix: "_large", Width: 1500, CropSettings: nil},
	{Suffix: "_large@2x", Width: 3000, CropSettings: nil},
}

func fetchAndResizePhoto(c *modulir.Context, targetDir string, photo *Photo) (bool, error) {
	u, err := url.Parse(photo.OriginalImageURL)
	if err != nil {
		return false, xerrors.Errorf("bad URL for photo '%s': %w", photo.Slug, err)
	}

	photoSizes := defaultPhotoSizes
	if photo.NoCrop {
		photoSizes = defaultPhotoSizesNoCrop
	}

	return mimage.FetchAndResizeImage(c, u, targetDir, photo.Slug, photo.TargetExt(),
		mimage.PhotoGravity(photo.CropGravity), photoSizes)
}

func fetchAndResizeDownloadedImage(c *modulir.Context,
	targetDir string, imageInfo *mtemplate.DownloadedImageInfo,
) (bool, error) {
	base := filepath.Base(imageInfo.Slug)
	dir := targetDir + filepath.Dir(imageInfo.Slug)

	return mimage.FetchAndResizeImage(c, imageInfo.URL, dir, base, extImageTarget(imageInfo.OriginalExt()), mimage.PhotoGravityCenter,
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

	return mimage.FetchAndResizeImage(c, u, targetDir, photo.Slug, photo.TargetExt(),
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

	return mimage.FetchAndResizeImage(c, u, targetDir, slug, extCanonical(extImageTarget(media.OriginalExt())),
		mimage.PhotoGravityCenter, twitterPhotoSizes)
}

// TODO: Needs to be refactored to respect markers.
//
// TODO: Needs to be refactored to do a less manual fetch (probably in Modulir).
//
// TODO: May want to eventually support non-manual video cutting.
func fetchVideo(ctx context.Context, c *modulir.Context, targetDir string, videoURL string) (bool, error) {
	u, err := url.Parse(videoURL)
	if err != nil {
		return false, xerrors.Errorf("bad URL for video '%s': %w", videoURL, err)
	}

	target := filepath.Join(targetDir, filepath.Base(u.Path))

	if mfile.Exists(target) {
		return false, nil
	}

	err = mfile.EnsureDir(c, targetDir)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return false, xerrors.Errorf("error creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, xerrors.Errorf("error fetching: %v", u.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, xerrors.Errorf("unexpected status code fetching '%v': %d",
			u.String(), resp.StatusCode)
	}

	f, err := os.Create(target)
	if err != nil {
		return false, xerrors.Errorf("error creating '%v': %w", target, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	// probably not needed
	defer w.Flush()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return false, xerrors.Errorf("error copying to '%v' from HTTP response: %w",
			target, err)
	}

	return true, nil
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

func groupReadingsByYear(readings []*squantified.Reading) []*readingYear {
	var year *readingYear
	var years []*readingYear

	for _, reading := range readings {
		if year == nil || year.Year != reading.ReadAt.Year() {
			year = &readingYear{reading.ReadAt.Year(), nil}
			years = append(years, year)
		}

		year.Readings = append(year.Readings, reading)
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
	location, err := time.LoadLocation(locationName)
	if err != nil {
		panic(err)
	}
	return location
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

func renderArticle(ctx context.Context, c *modulir.Context, source string,
	articles *[]*Article, articlesChanged *bool, mu *sync.Mutex,
) (bool, error) {
	sourceChanged := c.Changed(source)

	sourceTmpl := scommon.ViewsDir + "/articles/show.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
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

	content, err := mmarkdownext.Render(string(data), nil)
	if err != nil {
		return true, err
	}

	content, footnotes, ok := strings.Cut(content, `<div class="footnotes">`)
	if ok {
		footnotes = strings.TrimSuffix(footnotes, "</div>")
	}

	article.Content = template.HTML(content)
	article.Footnotes = template.HTML(footnotes) // may be empty

	toc, err := mtoc.RenderFromHTML(string(article.Content))
	if err != nil {
		return true, err
	}

	article.TOC = template.HTML(toc)

	if article.Hook != "" {
		hook, err := mmarkdownext.Render(string(article.Hook), nil)
		if err != nil {
			return true, err
		}

		article.Hook = template.HTML(mtemplate.CollapseParagraphs(hook))
	}

	format, ok := pathAsImage(
		path.Join(c.SourceDir, "content", "images", article.Slug, "hook"),
	)
	if ok {
		article.HookImageURL = "/assets/images/" + article.Slug + "/hook." + format
	}

	card := &twitterCard{
		Title:       article.Title,
		Description: string(article.Hook),
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

	err = dependencies.renderGoTemplate(ctx, c, sourceTmpl, path.Join(c.TargetDir, article.Slug), locals)
	if err != nil {
		return true, err
	}

	mu.Lock()
	insertOrReplaceArticle(articles, &article)
	*articlesChanged = true
	mu.Unlock()

	return true, nil
}

func renderArticlesIndex(ctx context.Context, c *modulir.Context, articles []*Article, articlesChanged bool) (bool, error) {
	sourceTmpl := scommon.ViewsDir + "/articles/index.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !articlesChanged && !viewsChanged {
		return false, nil
	}

	articlesByYear := groupArticlesByYear(articles)

	locals := getLocals("Articles"+scommon.TitleSuffix, map[string]interface{}{
		"ArticlesByYear": articlesByYear,
	})

	return true, dependencies.renderGoTemplate(ctx, c, sourceTmpl,
		path.Join(c.TargetDir, "articles/index.html"), locals)
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
			Summary:   string(article.Hook),
			Content:   &matom.EntryContent{Content: string(article.Content), Type: "html"},
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

// Number of atoms on the atom index page (the rest are on the archive page
// instead).
const maxAtomsIndex = 15

func renderAtomArchive(ctx context.Context, c *modulir.Context, atoms []*Atom, atomsChanged bool,
) (bool, error) {
	source := scommon.ViewsDir + "/atoms/index.tmpl.html"

	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !atomsChanged && !viewsChanged {
		return false, nil
	}

	locals := getLocals("Atoms archive"+scommon.TitleSuffix, map[string]interface{}{
		"Atoms": atoms,
	})

	err := dependencies.renderGoTemplate(ctx, c, source, path.Join(c.TargetDir, "atoms/archive"), locals)
	if err != nil {
		return true, err
	}

	return true, nil
}

func renderAtom(ctx context.Context, c *modulir.Context, atom *Atom, atomIndex int, atomsChanged bool,
) (bool, error) {
	source := scommon.ViewsDir + "/atoms/show.tmpl.html"

	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !atomsChanged && !viewsChanged {
		return false, nil
	}

	var title string
	if atom.Title == nil {
		title = fmt.Sprintf("Atom <%s>", atom.Slug)
	} else {
		title = *atom.Title
	}

	var card *twitterCard
	if len(atom.Photos) > 0 {
		photo := atom.Photos[0]
		card = &twitterCard{
			Description: fmt.Sprintf("Published %s.", atom.PublishedAt.Format("2006 / Jan 2 / 15:04 PST")),
			ImageURL: fmt.Sprintf("%s/photographs/atoms/%s/%s_large@2x%s",
				conf.AbsoluteURL, atom.Slug, photo.Slug, photo.TargetExt()),
			Title: title,
		}
	}

	locals := getLocals(title+scommon.TitleSuffix, map[string]interface{}{
		"Atom":        atom,
		"AtomIndex":   atomIndex,
		"IndexMax":    maxAtomsIndex,
		"TwitterCard": card,
	})

	err := dependencies.renderGoTemplate(ctx, c, source, path.Join(c.TargetDir, "atoms", atom.Slug), locals)
	if err != nil {
		return true, err
	}

	return true, nil
}

// Renders an Atom feed for atoms. The entries slice is assumed to be
// pre-sorted.
func renderAtomFeed(ctx context.Context, c *modulir.Context, atoms []*Atom, atomsChanged bool,
) (bool, error) {
	source := scommon.ViewsDir + "/atoms/_atom_atom.tmpl.html"

	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !atomsChanged && !viewsChanged {
		return false, nil
	}

	feed := &matom.Feed{
		Title: "Atoms " + scommon.TitleSuffix,
		ID:    "tag:" + scommon.AtomTag + ",2019:atoms",

		Links: []*matom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/atoms.atom"},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	if len(atoms) > 0 {
		feed.Updated = atoms[0].PublishedAt
	}

	for i, atom := range atoms {
		if i >= conf.NumAtomEntries {
			break
		}

		locals := getLocals("", map[string]interface{}{
			"Atom": atom,
		})

		var contentBuf bytes.Buffer
		err := dependencies.renderGoTemplateWriter(ctx, c, source, &contentBuf, locals)
		if err != nil {
			return true, err
		}

		title := atom.PublishedAt.Format("2006 / Jan 2 / 15:04 PST")
		if atom.Title != nil {
			title = *atom.Title
		}

		entry := &matom.Entry{
			Title:     title,
			Content:   &matom.EntryContent{Content: contentBuf.String(), Type: "html"},
			Published: atom.PublishedAt,
			Updated:   atom.PublishedAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/atoms/" + atom.Slug},
			ID: "tag:" + scommon.AtomTag + "," + atom.PublishedAt.Format("2006-01-02") +
				":atoms:" + atom.Slug,

			AuthorName: scommon.AtomAuthorName,
			AuthorURI:  conf.AbsoluteURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	filePath := path.Join(conf.TargetDir, "atoms.atom")
	f, err := os.Create(filePath)
	if err != nil {
		return true, xerrors.Errorf("error creating file '%s': %w", filePath, err)
	}
	defer f.Close()

	return true, feed.Encode(f, "  ")
}

func renderAtomIndex(ctx context.Context, c *modulir.Context, atoms []*Atom, atomsChanged bool,
) (bool, error) {
	source := scommon.ViewsDir + "/atoms/index.tmpl.html"

	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !atomsChanged && !viewsChanged {
		return false, nil
	}

	if len(atoms) > maxAtomsIndex {
		atoms = atoms[0:maxAtomsIndex]
	}

	locals := getLocals("Atoms"+scommon.TitleSuffix, map[string]interface{}{
		"Atoms":    atoms,
		"IndexMax": maxAtomsIndex,
	})

	err := dependencies.renderGoTemplate(ctx, c, source, path.Join(c.TargetDir, "atoms/index.html"), locals)
	if err != nil {
		return true, err
	}

	return true, nil
}

func renderFragment(ctx context.Context, c *modulir.Context, source string,
	fragments *[]*Fragment, fragmentsChanged *bool, mu *sync.Mutex,
) (bool, error) {
	sourceChanged := c.Changed(source)

	sourceTmpl := scommon.ViewsDir + "/fragments/show.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
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

	content, err := mmarkdownext.Render(string(data), nil)
	if err != nil {
		return true, err
	}

	content, footnotes, ok := strings.Cut(content, `<div class="footnotes">`)
	if ok {
		footnotes = strings.TrimSuffix(footnotes, "</div>")
	}

	fragment.Content = template.HTML(content)
	fragment.Footnotes = template.HTML(footnotes) // may be empty

	if fragment.Hook != "" {
		hook, err := mmarkdownext.Render(string(fragment.Hook), nil)
		if err != nil {
			return true, err
		}

		fragment.Hook = template.HTML(mtemplate.CollapseParagraphs(hook))
	}

	card := &twitterCard{
		Title:       fragment.Title,
		Description: string(fragment.Hook),
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

	err = dependencies.renderGoTemplate(ctx, c, sourceTmpl, path.Join(c.TargetDir, "fragments", fragment.Slug), locals)
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
			Summary:   string(fragment.Hook),
			Content:   &matom.EntryContent{Content: string(fragment.Content), Type: "html"},
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

func renderFragmentsIndex(ctx context.Context, c *modulir.Context, fragments []*Fragment,
	fragmentsChanged bool,
) (bool, error) {
	sourceTmpl := scommon.ViewsDir + "/fragments/index.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !fragmentsChanged && !viewsChanged {
		return false, nil
	}

	fragmentsByYear := groupFragmentsByYear(fragments)

	locals := getLocals("Fragments"+scommon.TitleSuffix, map[string]interface{}{
		"FragmentsByYear": fragmentsByYear,
	})

	return true, dependencies.renderGoTemplate(ctx, c, sourceTmpl,
		path.Join(c.TargetDir, "fragments/index.html"), locals)
}

func renderNanoglyph(ctx context.Context, c *modulir.Context, source string,
	issues *[]*snewsletter.Issue, nanoglyphsChanged *bool, mu *sync.Mutex,
) (bool, error) {
	sourceChanged := c.Changed(source)
	sourceTmpl := scommon.ViewsDir + "/nanoglyphs/show.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
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
		"BodyClass": "web_only", // For web-specific CSS rules
		"InEmail":   false,
		"Issue":     issue,
		"URLPrefix": "", // Relative prefix for the web version
	})

	err = dependencies.renderGoTemplate(ctx, c, sourceTmpl, path.Join(c.TargetDir, "nanoglyphs", issue.Slug), locals)
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

	feed := &matom.Feed{
		Title: "Nanoglyph" + scommon.TitleSuffix,
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
			content = template.HTML(fmt.Sprintf(`<p><img src="%s" alt="%s" /></p>`, issue.ImageURL, issue.ImageAlt)) + content
		}

		entry := &matom.Entry{
			Title:     fmt.Sprintf("Nanoglyph %s — %s", issue.Number, issue.Title),
			Content:   &matom.EntryContent{Content: string(content), Type: "html"},
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

func renderNanoglyphsIndex(ctx context.Context, c *modulir.Context, issues []*snewsletter.Issue,
	nanoglyphsChanged bool,
) (bool, error) {
	sourceTmpl := scommon.ViewsDir + "/nanoglyphs/index.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !nanoglyphsChanged && !viewsChanged {
		return false, nil
	}

	locals := getLocals("Nanoglyph"+scommon.TitleSuffix, map[string]interface{}{
		"BodyClass": "web_only", // For web-specific CSS rules
		"Issues":    issues,
		"URLPrefix": "", // Relative prefix for the web version
	})

	return true, dependencies.renderGoTemplate(ctx, c, sourceTmpl,
		path.Join(c.TargetDir, "nanoglyphs/index.html"), locals)
}

func renderPassage(ctx context.Context, c *modulir.Context, source string,
	issues *[]*snewsletter.Issue, passagesChanged *bool, mu *sync.Mutex,
) (bool, error) {
	sourceChanged := c.Changed(source)
	sourceTmpl := scommon.ViewsDir + "/passages/show.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
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
		"BodyClass": "web_only", // For web-specific CSS rules
		"InEmail":   false,
		"Issue":     issue,
		"URLPrefix": "", // Relative prefix for the web version
	})

	err = dependencies.renderGoTemplate(ctx, c, sourceTmpl, path.Join(c.TargetDir, "passages", issue.Slug), locals)
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

	feed := &matom.Feed{
		Title: "Passages & Glass" + scommon.TitleSuffix,
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
			content = template.HTML(fmt.Sprintf(`<p><img src="%s" alt="%s" /></p>`, issue.ImageURL, issue.ImageAlt)) + content
		}

		entry := &matom.Entry{
			Title:     fmt.Sprintf("Passages & Glass %s — %s", issue.Number, issue.Title),
			Content:   &matom.EntryContent{Content: string(content), Type: "html"},
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

func renderPassagesIndex(ctx context.Context, c *modulir.Context, issues []*snewsletter.Issue,
	passagesChanged bool,
) (bool, error) {
	sourceTmpl := scommon.ViewsDir + "/passages/index.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !passagesChanged && !viewsChanged {
		return false, nil
	}

	locals := getLocals("Passages", map[string]interface{}{
		"BodyClass": "web_only", // For web-specific CSS rules
		"Issues":    issues,
		"URLPrefix": "", // Relative prefix for the web version
	})

	return true, dependencies.renderGoTemplate(ctx, c, sourceTmpl,
		path.Join(c.TargetDir, "passages/index.html"), locals)
}

func renderHome(ctx context.Context, c *modulir.Context,
	articles []*Article, fragments []*Fragment, nanoglyphs []*snewsletter.Issue, photos []*Photo, sequences []*SequenceEntry,
	articlesChanged, fragmentsChanged, nanoglyphsChanged, photosChanged, sequencesChanged bool,
) (bool, error) {
	sourceTmpl := scommon.ViewsDir + "/index.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !articlesChanged && !fragmentsChanged && !nanoglyphsChanged && !photosChanged && !sequencesChanged && !viewsChanged {
		return false, nil
	}

	if len(articles) > 3 {
		articles = articles[0:3]
	}

	if len(fragments) > 3 {
		fragments = fragments[0:3]
	}

	if len(nanoglyphs) > 3 {
		nanoglyphs = nanoglyphs[0:3]
	}

	// Find a random photo to put on the homepage.
	photo := selectRandomPhoto(photos)

	if len(sequences) > 3 {
		sequences = sequences[0:3]
	}

	locals := getLocals("", map[string]interface{}{
		"Articles":   articles,
		"BodyClass":  "index",
		"Fragments":  fragments,
		"Nanoglyphs": nanoglyphs,
		"Photo":      photo,
		"Sequences":  sequences,
	})

	return true, dependencies.renderGoTemplate(ctx, c, sourceTmpl,
		path.Join(c.TargetDir, "index.html"), locals)
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
	err := mfile.EnsureDir(c, path.Dir(target))
	if err != nil {
		return true, err
	}

	pageMeta.dependencies = nil

	if strings.HasSuffix(source, ".ace") {
		ctx, includeMarkdownContainer := mtemplatemd.Context(ctx)

		locals := getLocals("", map[string]interface{}{
			"Ctx": ctx,
		})

		err := mace.RenderFile(c, scommon.MainLayout, source, target,
			getAceOptions(viewsChanged), locals)
		if err != nil {
			return true, err
		}

		pageMeta.dependencies = includeMarkdownContainer.Dependencies

		// Make sure dependencies are all watched on the filesystem.
		c.ChangedAny(pageMeta.dependencies...)
	} else {
		locals := getLocals("", nil)

		err := dependencies.renderGoTemplate(ctx, c, source, target, locals)
		if err != nil {
			return true, err
		}

		pageMeta.dependencies = dependencies.getDependencies(source)
	}

	return true, nil
}

func renderReading(ctx context.Context, c *modulir.Context) (bool, error) {
	source := scommon.ViewsDir + "/reading/index.tmpl.html"

	viewsChanged := c.ChangedAny(
		append([]string{
			c.SourceDir + "/content/reading/_meta.toml",
		},
			dependencies.getDependencies(source)...,
		)...)
	if !c.FirstRun && !viewsChanged {
		return false, nil
	}

	readings, err := squantified.GetReadingsData(c, c.SourceDir+"/content/reading/_meta.toml")
	if err != nil {
		return false, err
	}

	readingsByYear := groupReadingsByYear(readings)

	locals := getLocals("Reading"+scommon.TitleSuffix, map[string]interface{}{
		"ReadingsByYear": readingsByYear,
	})

	err = dependencies.renderGoTemplate(ctx, c, source, path.Join(c.TargetDir, "reading/index.html"), locals)
	if err != nil {
		return true, err
	}

	return true, nil
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
		content = `User-agent: Twitterbot
Disallow:
		
User-agent: *
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

func renderRuns(ctx context.Context, c *modulir.Context) (bool, error) {
	source := scommon.ViewsDir + "/runs/index.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !c.FirstRun && !viewsChanged {
		return false, nil
	}

	locals := getLocals("", map[string]interface{}{})

	err := dependencies.renderGoTemplate(ctx, c, source, path.Join(c.TargetDir, "runs", "index.html"), locals)
	if err != nil {
		return true, err
	}

	return true, nil
}

// Renders an Atom feed for sequences. The entries slice is assumed to be
// pre-sorted.
func renderSequenceFeed(ctx context.Context, c *modulir.Context,
	entries []*SequenceEntry, sequencesChanged bool,
) (bool, error) {
	source := scommon.ViewsDir + "/sequences/_entry_atom.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !sequencesChanged && !viewsChanged {
		return false, nil
	}

	feed := &matom.Feed{
		Title: "Sequences" + scommon.TitleSuffix,
		ID:    "tag:" + scommon.AtomTag + ",2019:sequences",

		Links: []*matom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/sequences.atom"},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	if len(entries) > 0 {
		feed.Updated = entries[0].PublishedAt
	}

	for i, entry := range entries {
		if i >= conf.NumAtomEntries {
			break
		}

		locals := getLocals("", map[string]interface{}{
			"Entry": entry,
		})

		var contentBuf bytes.Buffer
		err := dependencies.renderGoTemplateWriter(ctx, c, source, &contentBuf, locals)
		if err != nil {
			return true, err
		}

		entry := &matom.Entry{
			Title:     entry.Slug + " — " + entry.Title,
			Content:   &matom.EntryContent{Content: contentBuf.String(), Type: "html"},
			Published: entry.PublishedAt,
			Updated:   entry.PublishedAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/sequences/" + entry.Slug},
			ID: "tag:" + scommon.AtomTag + "," + entry.PublishedAt.Format("2006-01-02") +
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
	source := scommon.ViewsDir + "/sequences/show.tmpl.html"

	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !sequencesChanged && !viewsChanged {
		return false, nil
	}

	title := fmt.Sprintf("%s — %s", entry.Title, entry.Slug)

	var card *twitterCard
	if len(entry.Photos) > 0 {
		photo := entry.Photos[0]
		card = &twitterCard{
			Description: "",
			ImageURL: fmt.Sprintf("%s/photographs/sequences/%s_large@2x%s",
				conf.AbsoluteURL, photo.Slug, photo.TargetExt()),
			Title: title,
		}
	}

	locals := getLocals(title, map[string]interface{}{
		"Entry":       entry,
		"TwitterCard": card,
	})

	err := dependencies.renderGoTemplate(ctx, c, source, path.Join(c.TargetDir, "sequences", entry.Slug), locals)
	if err != nil {
		return true, err
	}

	return true, nil
}

func renderSequencesIndex(ctx context.Context, c *modulir.Context, entries []*SequenceEntry,
	sequenceChanged bool,
) (bool, error) {
	source := scommon.ViewsDir + "/sequences/index.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !sequenceChanged && !viewsChanged {
		return false, nil
	}

	locals := getLocals("Sequences"+scommon.TitleSuffix, map[string]interface{}{
		"Entries": entries,
	})

	err := dependencies.renderGoTemplate(ctx, c, source, path.Join(c.TargetDir, "sequences/index.html"), locals)
	if err != nil {
		return true, err
	}

	return true, nil
}

func renderTwitter(ctx context.Context, c *modulir.Context, tweets []*squantified.Tweet, tweetsChanged, withReplies bool) (bool, error) {
	source := scommon.ViewsDir + "/twitter/index.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !tweetsChanged && !viewsChanged {
		return false, nil
	}

	tweetsWithoutReplies := make([]*squantified.Tweet, 0, len(tweets))
	for _, tweet := range tweets {
		if tweet.ReplyOrMention {
			continue
		}

		tweetsWithoutReplies = append(tweetsWithoutReplies, tweet)
	}

	target := "index.html"
	ts := tweets
	if withReplies {
		target = "with-replies"
	} else {
		ts = tweetsWithoutReplies
	}

	tweetsByYearAndMonth := squantified.GroupTwitterByYearAndMonth(ts)
	tweetCountsByMonth := squantified.GetTwitterByMonth(ts)

	tweetCountsByMonthData, err := json.Marshal(tweetCountsByMonth)
	if err != nil {
		return false, xerrors.Errorf("error marshaling tweet counts: %w", err)
	}

	locals := getLocals("Twitter", map[string]interface{}{
		"NumTweets":            len(tweetsWithoutReplies),
		"NumTweetsWithReplies": len(tweets),
		"TweetCountsByMonth":   template.HTML(tweetCountsByMonthData), // chart: tweets by month
		"TweetsByYearAndMonth": tweetsByYearAndMonth,
		"WithReplies":          withReplies,
	})

	err = dependencies.renderGoTemplate(ctx, c, source, path.Join(c.TargetDir, "twitter", target), locals)
	if err != nil {
		return true, err
	}

	return true, nil
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
