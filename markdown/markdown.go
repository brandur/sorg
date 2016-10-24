package markdown

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/russross/blackfriday"
)

var renderFuncs = []func(string) string{
	// pre-transformations
	transformFigures,
	transformHeaders,

	// main Markdown rendering
	renderMarkdown,

	// post-transformations
	transformCodeWithLanguagePrefix,
	transformDivs,
	transformFootnotes,
	transformImagesToRetina,
}

// Render a Markdown string to HTML while applying all custom project-specific
// filters including footnotes and stable header links.
func Render(source string) string {
	for _, f := range renderFuncs {
		source = f(source)
	}
	return source
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

func renderMarkdown(source string) string {
	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_DASHES
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_USE_XHTML

	extensions := 0
	extensions |= blackfriday.EXTENSION_AUTO_HEADER_IDS
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_HEADER_IDS
	extensions |= blackfriday.EXTENSION_LAX_HTML_BLOCKS
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")
	return string(blackfriday.Markdown([]byte(source), renderer, extensions))
}

var codeRE = regexp.MustCompile(`<code class="(\w+)">`)

func transformCodeWithLanguagePrefix(source string) string {
	return codeRE.ReplaceAllString(source, `<code class="language-$1">`)
}

const openDivHTML = `<div class="%s">`
const closeDivHTML = `</div>`

var openDivRE = regexp.MustCompile(`(<p>)?!div class=("|&ldquo;)(.*)("|&rdquo;)(</p>)?`)
var closeDivRE = regexp.MustCompile(`(<p>)?!/div(</p>)?`)

func transformDivs(source string) string {
	out := source

	out = openDivRE.ReplaceAllStringFunc(out, func(div string) string {
		matches := openDivRE.FindStringSubmatch(div)
		class := matches[3]
		return fmt.Sprintf(openDivHTML, class)
	})
	out = closeDivRE.ReplaceAllString(out, closeDivHTML)

	return out
}

const figureHTML = `
<figure>
  <p><a href="%s"><img src="%s"></a></p>
  <figcaption>%s</figcaption>
</figure>
`

var figureRE = regexp.MustCompile(`!fig src="(.*)" caption="(.*)"`)

func transformFigures(source string) string {
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

// HTML for a reference to a footnote within the document.
const footnoteReferenceHTML = `
<sup id="footnote-%s-source">
  <a href="#footnote-%s">%s</a>
</sup>
`

// Look for the section the section at the bottom of the page that looks like
// <p>[1] (the paragraph tag is there because Markdown will have already
// wrapped it by this point).
var footerRE = regexp.MustCompile(`(?ms:^<p>\[\d+\].*)`)

// Look for a single footnote within the footer.
var footnoteRE = regexp.MustCompile(`\[(\d+)\](\s+.*)`)

// Note that this must be a post-transform filter. If it wasn't, our Markdown
// renderer would not render the Markdown inside the footnotes layer because it
// would already be wrapped in HTML.
func transformFootnotes(source string) string {
	footer := footerRE.FindString(source)

	if footer != "" {
		// remove the footer for now
		source = strings.Replace(source, footer, "", 1)

		footer = footnoteRE.ReplaceAllStringFunc(footer, func(footnote string) string {
			// first create a footnote with an anchor that links can target
			matches := footnoteRE.FindStringSubmatch(footnote)
			number := matches[1]
			anchor := fmt.Sprintf(footnoteAnchorHTML, number, number, number) + matches[2]

			// then replace all references in the body to this footnote
			referenceRE := regexp.MustCompile(fmt.Sprintf(`\[%s\]`, number))
			reference := fmt.Sprintf(footnoteReferenceHTML, number, number, number)
			source = referenceRE.ReplaceAllString(source, collapseHTML(reference))

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

// Matches one of the following:
//
//   # header
//   # header (#header-id)
//
// For now, only match ## or more so as to remove code comments from
// matches. We need a better way of doing that though.
var headerRE = regexp.MustCompile(`(?m:^(#{2,})\s+(.*?)(\s+\(#(.*)\))?$)`)

func transformHeaders(source string) string {
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
		return collapseHTML(fmt.Sprintf(headerHTML, level, newID, newID, title, level))
	})

	return source
}

var imageRE = regexp.MustCompile(`<img src="(.+)"`)

func transformImagesToRetina(source string) string {
	// The basic idea here is that we give every image a `retina-rjs` tag so
	// that Retina.JS will replace it with a retina version *except* if the
	// image is an SVG. These are resolution agnostic and don't need replacing.
	return imageRE.ReplaceAllStringFunc(source, func(img string) string {
		matches := imageRE.FindStringSubmatch(img)
		if filepath.Ext(matches[1]) == ".svg" {
			return fmt.Sprintf(`<img src="%s"`, matches[1])
		}
		return fmt.Sprintf(`<img data-rjs="2" src="%s"`, matches[1])
	})
}
