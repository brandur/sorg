package snewsletter

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/mtoml"
	"github.com/brandur/sorg/modules/smarkdown"
)

// Issue represents a single burst of the Nanoglyph or Passages & Glass
// newsletters to be rendered.
type Issue struct {
	// Content is the HTML content of the issue. It isn't included as TOML
	// frontmatter, and is rather split out of an issue's Markdown file,
	// rendered, and then added separately.
	Content string `toml:"-"`

	// ContentRaw is the raw Markdown content of the issue.
	ContentRaw string `toml:"-"`

	// Draft indicates that the issue is not yet published.
	Draft bool `toml:"-"`

	// Number is the number of the issue like "001". Notably, it's a number,
	// but zero-padded.
	Number string `toml:"-"`

	// PublishedAt is when the issue was published.
	PublishedAt *time.Time `toml:"published_at"`

	// Slug is a unique identifier for the issue that also helps determine
	// where it's addressable by URL. It's a combination of an issue number
	// (like `001` and a short identifier).
	Slug string `toml:"-"`

	// Title is the issue's title.
	Title string `toml:"title"`
}

func (p *Issue) validate(source string) error {
	if p.Title == "" {
		return fmt.Errorf("No title for issue: %v", source)
	}

	if p.PublishedAt == nil {
		return fmt.Errorf("No publish date for issue: %v", source)
	}

	return nil
}

// Render reads a newsletter file and builds an Issue object from it.
//
// The email parameter specifies whether or not the issue is being rendered
// to be sent it an email (as opposed for rendering on the web) and affects
// things like whether images should use absolute URLs.
func Render(c *modulir.Context, dir, name, absoluteURL string, email bool) (*Issue, error) {
	source := path.Join(dir, name)

	var issue Issue
	data, err := mtoml.ParseFileFrontmatter(c, source, &issue)
	if err != nil {
		return nil, err
	}

	issue.ContentRaw = string(data)
	issue.Draft = strings.Contains(filepath.Base(dir), "drafts")

	// TODO: Replace with extractSlug brought into scommon
	issue.Slug = strings.Replace(name, ".md", "", -1)

	slugParts := strings.Split(issue.Slug, "-")
	if len(slugParts) < 2 {
		return nil, fmt.Errorf("Expected slug to contain issue number: %v",
			issue.Slug)
	}
	issue.Number = slugParts[0]

	err = issue.validate(source)
	if err != nil {
		return nil, err
	}

	issue.Content = smarkdown.Render(issue.ContentRaw, &smarkdown.RenderOptions{
		AbsoluteURL:     absoluteURL,
		NoFootnoteLinks: email,
		NoHeaderLinks:   email,
		NoRetina:        true,
	})

	return &issue, nil
}
