package markdown

import (
	"fmt"
	"regexp"

	"github.com/russross/blackfriday"
)

var renderFuncs []func(string) string = []func(string) string{
	// pre-transformations
	transformCodeWithLanguagePrefix,

	// main Markdown rendering
	renderMarkdown,

	// post-transformations
	transformFigures,
}

func Render(source string) string {
	for _, f := range renderFuncs {
		source = f(source)
	}
	return source
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

var codeRE *regexp.Regexp = regexp.MustCompile(`<code class="(\w+)">`)

func transformCodeWithLanguagePrefix(source string) string {
	return codeRE.ReplaceAllString(source, `<code class="language-$1">`)
}

const figureHTML = `
<figure>
  <p><img src="%s"></p>
  <figcaption>%s</figcaption>
</figure>
`

var figureRE *regexp.Regexp = regexp.MustCompile(`!fig src="(.*)" caption="(.*)"`)

func transformFigures(source string) string {
	return figureRE.ReplaceAllStringFunc(source, func(figure string) string {
		matches := figureRE.FindStringSubmatch(figure)
		return fmt.Sprintf(figureHTML, matches[1], matches[2])
	})
}
