package markdown

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/brandur/sorg"
	"github.com/brandur/sorg/templatehelpers"
	"github.com/russross/blackfriday"
)

var preTransformationFuncs = []func(string, *RenderOptions) string{
	transformFigures,
	transformHeaders,
}

var postTransformationFuncs = []func(string, *RenderOptions) string{
	transformCodeWithLanguagePrefix,
	transformSections,
	transformFootnotes,
	transformImagesToAbsoluteURLs,
	transformImagesToRetina,
}

// RenderOptions describes a rendering operation to be customized.
type RenderOptions struct {
	// AbsoluteURLs replaces the sources of any images that pointed to relative
	// URLs with absolute URLs.
	AbsoluteURLs bool

	// NoFootnoteLinks disables linking to and from footnotes.
	NoFootnoteLinks bool

	// NoHeaderLinks disables automatic permalinks on headers.
	NoHeaderLinks bool

	// NoRetina disables the Retina.JS rendering attributes.
	NoRetina bool
}

// render is Markdown rendering function from Blackfriday that's been
// pre-composed into our the stack of our transformation functions.
var render func(string, *RenderOptions) string

// init runs on package initialization.
func init() {
	render = ComposeRenderStack(func(source []byte) []byte {
		return blackfriday.Run(source)
	})
}

// ComposeRenderStack takes a Markdown render function and composes it into a
// stack of functions along with sorg's "middleware" that performs various
// duties like adding header links and footnotes..
func ComposeRenderStack(renderF func([]byte) []byte) func(string, *RenderOptions) string {
	return func(source string, options *RenderOptions) string {
		for _, f := range preTransformationFuncs {
			source = f(source, options)
		}

		source = string(renderF([]byte(source)))

		for _, f := range postTransformationFuncs {
			source = f(source, options)
		}

		return source
	}
}

// Render a Markdown string to HTML while applying all custom project-specific
// filters including footnotes and stable header links.
func Render(source string, options *RenderOptions) string {
	return render(source, options)
}

// Look for any whitespace between HTML tags.
var whitespaceRE = regexp.MustCompile(`>\s+<`)

// Simply collapses certain HTML snippets by removing newlines and whitespace
// between tags. This is mainline used to make HTML snippets readable as
// constants, but then to make them fit a little more nicely into the rendered
// markup.
func collapseHTML(html string) string {
	html = strings.Replace(html, "\n", "", -1)
	html = whitespaceRE.ReplaceAllString(html, "><")
	return html
}

func renderMarkdown(source string, options *RenderOptions) string {
	return string(blackfriday.Run([]byte(source)))
}

var codeRE = regexp.MustCompile(`<code class="(\w+)">`)

func transformCodeWithLanguagePrefix(source string, options *RenderOptions) string {
	return codeRE.ReplaceAllString(source, `<code class="language-$1">`)
}

const openSectionHTML = `<section class="%s">`
const closeSectionHTML = `</section>`

var openSectionRE = regexp.MustCompile(`(<p>)?!section class=("|&ldquo;)(.*)("|&rdquo;)(</p>)?`)
var closeSectionRE = regexp.MustCompile(`(<p>)?!/section(</p>)?`)

func transformSections(source string, options *RenderOptions) string {
	out := source

	out = openSectionRE.ReplaceAllStringFunc(out, func(div string) string {
		matches := openSectionRE.FindStringSubmatch(div)
		class := matches[3]
		return fmt.Sprintf(openSectionHTML, class)
	})
	out = closeSectionRE.ReplaceAllString(out, closeSectionHTML)

	return out
}

const figureHTML = `
<figure>
  <p><a href="%s"><img src="%s" class="overflowing"></a></p>
  <figcaption>%s</figcaption>
</figure>
`

var figureRE = regexp.MustCompile(`!fig src="(.*)" caption="(.*)"`)

func transformFigures(source string, options *RenderOptions) string {
	return figureRE.ReplaceAllStringFunc(source, func(figure string) string {
		matches := figureRE.FindStringSubmatch(figure)
		src := matches[1]

		link := src
		extension := filepath.Ext(link)
		if extension != "" && extension != ".svg" {
			link = link[0:len(src)-len(extension)] + "@2x" + extension
		}

		// This is a really ugly hack in that it relies on the regex above
		// being greedy about quotes, but meh, I'll make it better when there's
		// a good reason to.
		caption := strings.Replace(matches[2], `\"`, `"`, -1)

		return fmt.Sprintf(figureHTML, link, src, caption)
	})
}

// A layer that we wrap the entire footer section in for styling purposes.
const footerWrapper = `
<div id="footnotes">
  %s
</div>
`

// HTML for a footnote within the document.
const footnoteAnchorHTML = `
<sup id="footnote-%s">
  <a href="#footnote-%s-source">%s</a>
</sup>
`

// Same as footnoteAnchorHTML but without a link(this is used when sending
// emails).
const footnoteAnchorHTMLWithoutLink = `<sup><strong>%s</strong></sup>`

// HTML for a reference to a footnote within the document.
//
// Make sure there's a single space before the <sup> because we're replacing
// one as part of our search.
const footnoteReferenceHTML = `
 <sup id="footnote-%s-source">
  <a href="#footnote-%s">%s</a>
</sup>
`

// Same as footnoteReferenceHTML but without a link (this is used when sending
// emails).
//
// Make sure there's a single space before the <sup> because we're replacing
// one as part of our search.
const footnoteReferenceHTMLWithoutLink = ` <sup><strong>%s</strong></sup>`

// Look for the section the section at the bottom of the page that looks like
// <p>[1] (the paragraph tag is there because Markdown will have already
// wrapped it by this point).
var footerRE = regexp.MustCompile(`(?ms:^<p>\[\d+\].*)`)

// Look for a single footnote within the footer.
var footnoteRE = regexp.MustCompile(`\[(\d+)\](\s+.*)`)

// Note that this must be a post-transform filter. If it wasn't, our Markdown
// renderer would not render the Markdown inside the footnotes layer because it
// would already be wrapped in HTML.
func transformFootnotes(source string, options *RenderOptions) string {
	footer := footerRE.FindString(source)

	if footer != "" {
		// remove the footer for now
		source = strings.Replace(source, footer, "", 1)

		footer = footnoteRE.ReplaceAllStringFunc(footer, func(footnote string) string {
			// first create a footnote with an anchor that links can target
			matches := footnoteRE.FindStringSubmatch(footnote)
			number := matches[1]

			var anchor string
			if options != nil && options.NoFootnoteLinks {
				anchor = fmt.Sprintf(footnoteAnchorHTMLWithoutLink, number) + matches[2]
			} else {
				anchor = fmt.Sprintf(footnoteAnchorHTML, number, number, number) + matches[2]
			}

			// Then replace all references in the body to this footnote.
			//
			// Note the leading space before ` [%s]`. This is a little hacky,
			// but is there to try and ensure that we don't try to replace
			// strings that look like footnote references, but aren't.
			// `KEYS[1]` from `/redis-cluster` is an example of one of these
			// strings that might be a false positive.
			var reference string
			if options != nil && options.NoFootnoteLinks {
				reference = fmt.Sprintf(footnoteReferenceHTMLWithoutLink, number)
			} else {
				reference = fmt.Sprintf(footnoteReferenceHTML, number, number, number)
			}
			source = strings.Replace(source,
				fmt.Sprintf(` [%s]`, number),
				collapseHTML(reference), -1)

			return collapseHTML(anchor)
		})

		// and wrap the whole footer section in a layer for styling
		footer = fmt.Sprintf(footerWrapper, footer)
		source = source + footer
	}

	return source
}

const headerHTML = `
<h%v id="%s">
  <a href="#%s">%s</a>
</h%v>
`

const headerHTMLNoLink = `
<h%v>%s</h%v>
`

// Matches one of the following:
//
//   # header
//   # header (#header-id)
//
// For now, only match ## or more so as to remove code comments from
// matches. We need a better way of doing that though.
var headerRE = regexp.MustCompile(`(?m:^(#{2,})\s+(.*?)(\s+\(#(.*)\))?$)`)

func transformHeaders(source string, options *RenderOptions) string {
	headerNum := 0

	// Tracks previously assigned headers so that we can detect duplicates.
	headers := make(map[string]int)

	source = headerRE.ReplaceAllStringFunc(source, func(header string) string {
		matches := headerRE.FindStringSubmatch(header)

		level := len(matches[1])
		title := matches[2]
		id := matches[4]

		var newID string

		if id == "" {
			// Header with no name, assign a prefixed number.
			newID = fmt.Sprintf("section-%v", headerNum)

		} else {
			occurrence, ok := headers[id]

			if ok {
				// Give duplicate IDs a suffix.
				newID = fmt.Sprintf("%s-%d", id, occurrence)
				headers[id]++

			} else {
				// Otherwise this is the first such ID we've seen.
				newID = id
				headers[id] = 1
			}
		}

		headerNum++

		// Replace the Markdown header with HTML equivalent.
		if options != nil && options.NoHeaderLinks {
			return collapseHTML(fmt.Sprintf(headerHTMLNoLink, level, title, level))
		}

		return collapseHTML(fmt.Sprintf(headerHTML, level, newID, newID, title, level))

	})

	return source
}

var imageRE = regexp.MustCompile(`<img src="([^"]+)"`)

func transformImagesToRetina(source string, options *RenderOptions) string {
	if options != nil && options.NoRetina {
		return source
	}

	// The basic idea here is that we give every image a `srcset` that includes
	// 2x so that browsers will replace it with a retina version *except* if
	// the image is an SVG. These are resolution agnostic and don't need
	// replacing.
	return imageRE.ReplaceAllStringFunc(source, func(img string) string {
		matches := imageRE.FindStringSubmatch(img)
		if filepath.Ext(matches[1]) == ".svg" {
			return fmt.Sprintf(`<img src="%s"`, matches[1])
		}
		return fmt.Sprintf(`<img src="%s" srcset="%s 2x, %s 1x"`,
			matches[1],
			templatehelpers.To2x(matches[1]),
			matches[1],
		)
	})
}

var relativeImageRE = regexp.MustCompile(`<img src="/`)

func transformImagesToAbsoluteURLs(source string, options *RenderOptions) string {
	if options == nil || !options.AbsoluteURLs {
		return source
	}

	return relativeImageRE.ReplaceAllStringFunc(source, func(img string) string {
		return `<img src="` + sorg.AbsoluteURL + `/`
	})
}
