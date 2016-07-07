package markdown

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	assert.Equal(t, "<p><strong>strong</strong></p>\n", Render("**strong**"))
}

func TestRenderMarkdown(t *testing.T) {
	assert.Equal(t, "<p><strong>strong</strong></p>\n", renderMarkdown("**strong**"))
}

func TestTransformCodeWithLanguagePrefix(t *testing.T) {
	assert.Equal(t,
		`<code class="language-ruby">`,
		transformCodeWithLanguagePrefix(`<code class="ruby">`),
	)
}

func TestTransformFigures(t *testing.T) {
	assert.Equal(t, `
<figure>
  <p><img src="fig-src"></p>
  <figcaption>fig-caption</figcaption>
</figure>
`,
		transformFigures(`!fig src="fig-src" caption="fig-caption"`),
	)
}

func TestTransformFootnotes(t *testing.T) {
	assert.Equal(t, `
<p>This is a reference <sup id="footnote-1-source"><a href="#footnote-1">1</a></sup> to a footnote <sup id="footnote-2-source"><a href="#footnote-2">2</a></sup>.</p>


<div id="footnotes">
  <p><sup id="footnote-1"><a href="#footnote-1-source">1</a></sup> Footnote one.</p>

<p><sup id="footnote-2"><a href="#footnote-2-source">2</a></sup> Footnote two.</p>

</div>
`,
		transformFootnotes(`
<p>This is a reference [1] to a footnote [2].</p>

<p>[1] Footnote one.</p>

<p>[2] Footnote two.</p>
`,
		),
	)
}
