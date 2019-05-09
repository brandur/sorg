package spassages

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/myaml"
	"github.com/brandur/sorg/modules/smarkdown"
)

// Passage represents a single burst of the Passage & Glass newsletter to be
// rendered.
type Passage struct {
	// Content is the HTML content of the passage. It isn't included as YAML
	// frontmatter, and is rather split out of an passage's Markdown file,
	// rendered, and then added separately.
	Content string `yaml:"-"`

	// ContentRaw is the raw Markdown content of the passage.
	ContentRaw string `yaml:"-"`

	// Draft indicates that the passage is not yet published.
	Draft bool `yaml:"-"`

	// Issue is the issue number of the passage like "001". Notably, it's a
	// number, but zero-padded.
	Issue string `yaml:"-"`

	// PublishedAt is when the passage was published.
	PublishedAt *time.Time `yaml:"published_at"`

	// Slug is a unique identifier for the passage that also helps determine
	// where it's addressable by URL. It's a combination of an issue number
	// (like `001` and a short identifier).
	Slug string `yaml:"-"`

	// Title is the passage's title.
	Title string `yaml:"title"`
}

func (p *Passage) validate(source string) error {
	if p.Title == "" {
		return fmt.Errorf("No title for passage: %v", source)
	}

	if p.PublishedAt == nil {
		return fmt.Errorf("No publish date for passage: %v", source)
	}

	return nil
}

// Render reads a passage file and builds a Passage object from it.
//
// The email parameter specifies whether or not the passage is being rendered
// to be sent it an email (as opposed for rendering on the web) and affects
// things like whether images should use absolute URLs.
func Render(c *modulir.Context, dir, name, absoluteURL string, email bool) (*Passage, error) {
	source := path.Join(dir, name)

	var passage Passage
	data, err := myaml.ParseFileFrontmatter(c, source, &passage)
	if err != nil {
		return nil, err
	}

	passage.ContentRaw = string(data)
	passage.Draft = strings.Contains(filepath.Base(dir), "drafts")

	// TODO: Replace with extractSlug brought into scommon
	passage.Slug = strings.Replace(name, ".md", "", -1)

	slugParts := strings.Split(passage.Slug, "-")
	if len(slugParts) < 2 {
		return nil, fmt.Errorf("Expected passage slug to contain issue number: %v",
			passage.Slug)
	}
	passage.Issue = slugParts[0]

	err = passage.validate(source)
	if err != nil {
		return nil, err
	}

	passage.Content = smarkdown.Render(passage.ContentRaw, &smarkdown.RenderOptions{
		AbsoluteURL:     absoluteURL,
		NoFootnoteLinks: email,
		NoHeaderLinks:   email,
		NoRetina:        true,
	})

	return &passage, nil
}
