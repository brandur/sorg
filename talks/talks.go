package talks

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"time"

	"github.com/brandur/sorg"
	"github.com/brandur/sorg/markdown"
	"gopkg.in/yaml.v2"
)

// Slide represents a slide within a talk.
type Slide struct {
	// CaptionRaw is a caption for the slide, in rendered HTML.
	Caption string `yaml:"-"`

	// CaptionRaw is a caption for the slide, in Markdown.
	CaptionRaw string `yaml:"caption"`

	// Number is the order number of the slide in string format and padded with
	// leading zeros.
	Number string `yaml:"-"`
}

// Talk represents a single talk.
type Talk struct {
	// Draft indicates that the talk is not yet published.
	Draft bool `yaml:"-"`

	// Event is the event for which the talk was originally given.
	Event string `yaml:"event"`

	// Intro is an introduction for the talk, in HTML.
	Intro string `yaml:"-"`

	// IntroRaw is an introduction for the talk, in Markdown.
	IntroRaw string `yaml:"intro"`

	// Location is the city where the talk was originally given.
	Location string `yaml:"location"`

	// PublishedAt is when the talk was published.
	PublishedAt *time.Time `yaml:"published_at"`

	// Slides is the collection of slides that are part of the talk.
	Slides []*Slide `yaml:"slides"`

	// Slug is a unique identifier for the talk that also helps determine
	// where it's addressable by URL.
	Slug string `yaml:"-"`

	// Subtitle is the talk's subtitle.
	Subtitle string `yaml:"subtitle"`

	// Title is the talk's title.
	Title string `yaml:"title"`
}

// PublishingInfo produces a brief spiel about publication which is intended to
// go into the left sidebar when a talk is shown.
func (t *Talk) PublishingInfo() string {
	return `<p><em><strong>` + t.Title + `</strong> is a talk originally given on <strong>` +
		t.PublishedAt.Format("January 2, 2006") + `</strong> for <strong>` +
		t.Event + `</strong> in <strong>` + t.Location + `</strong>.</em></p>` + sorg.TwitterInfo
}

// Compile reads a talk file and builds a Talk object from it.
func Compile(dir, name string, draft bool) (*Talk, error) {
	inPath := path.Join(dir, name)

	raw, err := ioutil.ReadFile(inPath)
	if err != nil {
		return nil, err
	}

	var talk Talk
	err = yaml.Unmarshal([]byte(raw), &talk)
	if err != nil {
		return nil, err
	}

	talk.Draft = draft
	talk.Intro = renderMarkdown(talk.IntroRaw)
	talk.Slug = strings.Replace(name, ".yaml", "", -1)

	if talk.Event == "" {
		return nil, fmt.Errorf("No event for talk: %v", inPath)
	}

	if talk.Intro == "" {
		return nil, fmt.Errorf("No intro for talk: %v", inPath)
	}

	if talk.Location == "" {
		return nil, fmt.Errorf("No location for talk: %v", inPath)
	}

	if talk.Title == "" {
		return nil, fmt.Errorf("No title for talk: %v", inPath)
	}

	if talk.PublishedAt == nil {
		return nil, fmt.Errorf("No publish date for talk: %v", inPath)
	}

	for i, slide := range talk.Slides {
		slide.Caption = renderMarkdown(slide.CaptionRaw)
		slide.Number = fmt.Sprintf("%03d", i+1)
	}

	return &talk, nil
}

func renderMarkdown(content string) string {
	return markdown.Render(content, &markdown.RenderOptions{
		NoFootnoteLinks: true,
		NoHeaderLinks:   true,
		NoRetina:        true,
	})
}
