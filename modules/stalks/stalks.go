package stalks

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/mtoml"
	"github.com/brandur/sorg/modules/scommon"
	"github.com/brandur/sorg/modules/smarkdown"
)

// Slide represents a slide within a talk.
type Slide struct {
	// Content is content for the slide, in rendered HTML.
	Content string

	// ContentRaw is content for the slide, in Markdown.
	ContentRaw string

	// ImagePath is the path to the image asset for this slide. It's generated
	// from a combination of the talk's slug, slide's number, and whether the
	// slide is detected to be in JPG or PNG.
	ImagePath string

	// PresenterNotesRaw are the presenter notest for the slide, in rendered
	// HTML.
	PresenterNotes string

	// PresenterNotesRaw are the presenter notest for the slide, in Markdown.
	PresenterNotesRaw string

	// Number is the order number of the slide in string format and padded with
	// leading zeros.
	Number string
}

// Talk represents a single talk.
type Talk struct {
	// Draft indicates that the talk is not yet published.
	Draft bool `toml:"-"`

	// Event is the event for which the talk was originally given.
	Event string `toml:"event"`

	// Intro is an introduction for the talk, in HTML.
	Intro string `toml:"-"`

	// IntroRaw is an introduction for the talk, in Markdown.
	IntroRaw string `toml:"-"`

	// Location is the city where the talk was originally given.
	Location string `toml:"location"`

	// PublishedAt is when the talk was published.
	PublishedAt *time.Time `toml:"published_at"`

	// Slides is the collection of slides that are part of the talk.
	Slides []*Slide `toml:"-"`

	// Slug is a unique identifier for the talk that also helps determine
	// where it's addressable by URL.
	Slug string `toml:"-"`

	// Subtitle is the talk's subtitle.
	Subtitle string `toml:"subtitle,omitempty"`

	// Title is the talk's title.
	Title string `toml:"title"`
}

// PublishingInfo produces a brief spiel about publication which is intended to
// go into the left sidebar when a talk is shown.
func (t *Talk) PublishingInfo() string {
	return `<p><strong>Talk</strong><br>` + t.Title + `</p>` +
		`<p><strong>Published</strong><br>` + t.PublishedAt.Format("January 2, 2006") + `</p>` +
		`<p><strong>Location</strong><br>` + t.Location + `</p>` +
		`<p><strong>Event</strong><br>` + t.Event + `</p>` +
		scommon.TwitterInfo
}

func (t *Talk) validate(source string) error {
	if t.Event == "" {
		return fmt.Errorf("No event for talk: %v", source)
	}

	if t.Location == "" {
		return fmt.Errorf("No location for talk: %v", source)
	}

	if t.Title == "" {
		return fmt.Errorf("No title for talk: %v", source)
	}

	if t.PublishedAt == nil {
		return fmt.Errorf("No publish date for talk: %v", source)
	}

	return nil
}

// Render reads a talk file and builds a Talk object from it.
func Render(c *modulir.Context, contentDir, dir, name string) (*Talk, error) {
	source := path.Join(dir, name)

	var talk Talk
	data, err := mtoml.ParseFileFrontmatter(c, source, &talk)
	if err != nil {
		return nil, err
	}

	talk.Draft = strings.Contains(filepath.Base(dir), "drafts")

	// TODO: Replace with extractSlug brought into scommon
	talk.Slug = strings.Replace(name, ".md", "", -1)

	err = talk.validate(source)
	if err != nil {
		return nil, err
	}

	talk.Slides, err = splitAndRenderSlides(contentDir, &talk, string(data))
	if err != nil {
		return nil, err
	}

	// The preseneter notes for the first slide (the title slide) also serve as
	// the intro for the entire talk. For convenience, set that content onto
	// the talk's struct.
	if len(talk.Slides) > 0 {
		talk.Intro = talk.Slides[0].PresenterNotes
		talk.IntroRaw = talk.Slides[0].PresenterNotesRaw

		// We also set those notes to empty so that the slide itself renders
		// without them.
		talk.Slides[0].PresenterNotes = ""
		talk.Slides[0].PresenterNotesRaw = ""
	}

	if talk.Intro == "" {
		return nil, fmt.Errorf("No intro for talk: %v (provide one as the presenter notes of the first slide)", source)
	}

	return &talk, nil
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

func renderMarkdown(content string) string {
	return smarkdown.Render(content, &smarkdown.RenderOptions{
		NoFootnoteLinks: true,
		NoHeaderLinks:   true,
		NoRetina:        true,
	})
}

func splitAndRenderSlides(contentDir string, talk *Talk, content string) ([]*Slide, error) {
	talksAssetPath := "/assets/images/talks"
	talksImageDir := path.Join(contentDir, "images", "talks")

	rawSlides := strings.Split(content, "---\n")
	slides := make([]*Slide, len(rawSlides))

	for i, rawSlide := range rawSlides {
		slide := &Slide{}
		slides[i] = slide

		parts := strings.Split(rawSlide, "???\n")
		rawContent := parts[0]

		var rawPresenterNotes string
		if len(parts) > 1 {
			rawPresenterNotes = parts[1]
		}

		slide.ContentRaw = strings.TrimSpace(rawContent)
		slide.Content = renderMarkdown(slide.ContentRaw)

		slide.PresenterNotesRaw = strings.TrimSpace(rawPresenterNotes)
		slide.PresenterNotes = renderMarkdown(slide.PresenterNotesRaw)

		slide.Number = fmt.Sprintf("%03d", i+1)

		// Try PNG then fall back to JPG. If neither exists, error.
		pngName := fmt.Sprintf("%s.%s.png", talk.Slug, slide.Number)
		jpgName := fmt.Sprintf("%s.%s.jpg", talk.Slug, slide.Number)

		if fileExists(path.Join(talksImageDir, talk.Slug, pngName)) {
			slide.ImagePath = fmt.Sprintf("%s/%s/%s", talksAssetPath, talk.Slug, pngName)
		} else if fileExists(path.Join(talksImageDir, talk.Slug, jpgName)) {
			slide.ImagePath = fmt.Sprintf("%s/%s/%s", talksAssetPath, talk.Slug, jpgName)
		} else {
			return nil, fmt.Errorf("Couldn't find any image asset for slide %s / %s at %s",
				pngName, jpgName, path.Join(talksImageDir, talk.Slug))
		}
	}

	return slides, nil
}
