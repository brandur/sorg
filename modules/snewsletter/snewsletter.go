package snewsletter

import (
	"html/template"
	"path"
	"strings"
	"time"

	"golang.org/x/xerrors"

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/mmarkdownext"
	"github.com/brandur/modulir/modules/mtoml"
	"github.com/brandur/sorg/modules/scommon"
)

// Possible orientations for a newsletter's main image.
const (
	ImageOrientationLandscape = "landscape"
	ImageOrientationPortrait  = "portrait"
)

// Issue represents a single burst of the Nanoglyph or Passages & Glass
// newsletters to be rendered.
type Issue struct {
	// Content is the HTML content of the issue. It isn't included as TOML
	// frontmatter, and is rather split out of an issue's Markdown file,
	// rendered, and then added separately.
	Content template.HTML `toml:"-"`

	// ContentRaw is the raw Markdown content of the issue.
	ContentRaw string `toml:"-"`

	// Draft indicates that the issue is not yet published.
	Draft bool `toml:"-"`

	// HookImageURL is the URL for a hook image for the issue (to be shown on
	// the newsletter index) if one was found. Should generally be the same
	// image as ImageURL.
	HookImageURL string `toml:"-"`

	// ImageAlt is the alternate description text for the issue's main image.
	//
	// Currently not used by Passages.
	ImageAlt string `toml:"image_alt"`

	// ImageOrientation is the orientation of the main image (either landscape
	// or portrait).
	//
	// Currently not used by Passages.
	ImageOrientation string `toml:"image_orientation"`

	// ImageURL is the source URL for the issue's main image.
	//
	// Currently not used by Passages.
	ImageURL string `toml:"image_url"`

	// Number is the number of the issue like "001". Notably, it's a number,
	// but zero-padded.
	Number string `toml:"-"`

	// PublishedAt is when the issue was published.
	PublishedAt time.Time `toml:"published_at"`

	// Slug is a unique identifier for the issue that also helps determine
	// where it's addressable by URL. It's a combination of an issue number
	// (like `001` and a short identifier).
	Slug string `toml:"-"`

	// Title is the issue's title.
	Title string `toml:"title"`
}

func (p *Issue) validate(source string) error {
	// Unfortunately, no default values with TOML.
	if p.ImageOrientation == "" {
		p.ImageOrientation = ImageOrientationLandscape
	}
	if p.ImageOrientation != ImageOrientationLandscape && p.ImageOrientation != ImageOrientationPortrait {
		return xerrors.Errorf("unsupported image orientation for issue: %v (must be '%v' or '%v')",
			source,
			ImageOrientationLandscape,
			ImageOrientationPortrait)
	}

	if p.Title == "" {
		return xerrors.Errorf("no title for issue: %v", source)
	}

	if p.PublishedAt.IsZero() {
		return xerrors.Errorf("no publish date for issue: %v", source)
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
	issue.Draft = scommon.IsDraft(name)
	issue.Slug = scommon.ExtractSlug(name)

	slugParts := strings.Split(issue.Slug, "-")
	if len(slugParts) < 2 {
		return nil, xerrors.Errorf("expected slug to contain issue number: %v",
			issue.Slug)
	}
	issue.Number = slugParts[0]

	err = issue.validate(source)
	if err != nil {
		return nil, err
	}

	content, err := mmarkdownext.Render(issue.ContentRaw, &mmarkdownext.RenderOptions{
		AbsoluteURL:     absoluteURL,
		NoFootnoteLinks: email,
		NoFollow:        true,
		NoHeaderLinks:   email,
		NoRetina:        true,

		// Pass a special template var so that we can optionally render signup
		// forms right inside the body of a newsletter issue.
		TemplateData: map[string]interface{}{
			"InEmail": email,
		},
	})
	if err != nil {
		return nil, err
	}

	issue.Content = template.HTML(content)

	return &issue, nil
}
